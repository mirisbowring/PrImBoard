package primboard

import (
	"errors"

	"github.com/mirisbowring/PrImBoard/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Node holts the users and the information about the group
type Node struct {
	ID      primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Title   string               `json:"title,omitempty" bson:"title,omitempty"`
	Creator string               `json:"creator,omitempty" bson:"creator,omitempty"`
	Type    string               `json:"type,omitempty" bson:"type,omitempty"`
	Secret  string               `json:"secret,omitempty" bson:"secret,omitempty"`
	Groups  []primitive.ObjectID `json:"groups,omitempty" bson:"groups,omitempty"`
	URL     string               `json:"url,omitempty" bson:"url,omitempty"`
}

// NodeProject is a bson representation of the ipfs-node setting object
var NodeProject = bson.M{
	"id":      1,
	"title":   1,
	"creator": 1,
	"type":    1,
	"groups":  1,
	"url":     1,
}

// name of the mongo collection
var nodeColName = "node"

// AddNode creates the model in the mongodb
func (e *Node) AddNode(db *mongo.Database) (*mongo.InsertOneResult, error) {
	conn := GetColCtx(nodeColName, db, 30)
	result, err := conn.Col.InsertOne(conn.Ctx, e)
	defer conn.Cancel()
	return result, err
}

// DeleteNode deletes the model from the mongodb
func (e *Node) DeleteNode(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := GetColCtx(nodeColName, db, 30)
	filter := bson.M{"_id": e.ID}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
	return result, err
}

// GetAllNodes selects all Nodes from the mongodb
func GetAllNodes(db *mongo.Database, permission bson.M) ([]Node, error) {
	// create pipeline
	pipeline, err := createPermissionProjectPipeline(permission, primitive.NilObjectID, NodeProject)
	if err != nil {
		return nil, err
	}
	// execute query
	opts := options.Aggregate()
	conn := GetColCtx(nodeColName, db, 30)
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts) // find all
	if err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer cursor.Close(conn.Ctx)
	// iterate over the cursor and create array
	var nodes []Node
	for cursor.Next(conn.Ctx) {
		var n Node
		cursor.Decode(&n)
		nodes = append(nodes, n)
	}
	// report errors if occured
	if err = cursor.Err(); err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer conn.Cancel()
	return nodes, nil
}

// GetNode returns the specified entry from the mongodb
func (e *Node) GetNode(db *mongo.Database, permission bson.M) error {
	// create pipeline
	pipeline, err := createPermissionProjectPipeline(permission, e.ID, NodeProject)
	if err != nil {
		return err
	}
	opts := options.Aggregate()
	conn := GetColCtx(nodeColName, db, 30)
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {
		defer conn.Cancel()
		return err
	}
	var found = false
	for cursor.Next(conn.Ctx) {
		err := cursor.Decode(&e)
		if err != nil {
			defer conn.Cancel()
			return err
		}
		found = true
		break
	}
	defer conn.Cancel()
	if !found {
		return errors.New("no results")
	}
	return nil
}

// UpdateNode updates the record with the passed one
func (e *Node) UpdateNode(db *mongo.Database, ue Node, permission bson.M) (*mongo.UpdateResult, error) {
	// check if user is allowed to select this node
	if err := e.GetNode(db, permission); err != nil {
		return nil, err
	}
	// continue with update
	conn := GetColCtx(nodeColName, db, 30)
	filter := bson.M{"_id": e.ID}
	update := bson.M{"$set": ue}
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	defer conn.Cancel()
	return result, err
}

// VerifyNode verifies all mandatory fields of the specified node
// does not verify ID
func (e *Node) VerifyNode(db *mongo.Database) error {
	if e.Title == "" {
		return errors.New("node title must be set")
	}
	if _, found := helper.FindInSlice([]string{"ipfs", "web"}, e.Type); !found {
		return errors.New("specified type is not valid")
	}
	if e.Creator == "" {
		return errors.New("creator must be specified")
	}
	if e.URL == "" {
		return errors.New("url is not valid")
	}
	if len(e.Groups) > 0 {
		// select the specified groups
		groups, err := GetUserGroupsByIDs(db, e.Groups)
		if err != nil {
			return err
		}
		// set all groups that could be found
		var tmp []primitive.ObjectID
		for _, g := range groups {
			tmp = append(tmp, g.ID)
		}
		e.Groups = tmp
	}
	return nil
}
