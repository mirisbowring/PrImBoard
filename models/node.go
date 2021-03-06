package models

import (
	"errors"

	"github.com/mirisbowring/primboard/helper"
	"github.com/mirisbowring/primboard/helper/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

// Node holts the users and the information about the group
type Node struct {
	ID           primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Title        string               `json:"title,omitempty" bson:"title,omitempty"`
	Creator      string               `json:"creator,omitempty" bson:"creator,omitempty"`
	GroupIDs     []primitive.ObjectID `json:"groupIDs,omitempty" bson:"groupIDs,omitempty"`
	KeycloakID   string               `json:"keycloakID,omitempty" bson:"keycloakID,omitempty"`
	Type         string               `json:"type,omitempty" bson:"type,omitempty"`
	Secret       string               `json:"secret,omitempty" bson:"secret,omitempty"`
	APIEndpoint  string               `json:"APIEndpoint,omitempty" bson:"APIEndpoint,omitempty"`
	DataEndpoint string               `json:"dataEndpoint,omitempty" bson:"dataEndpoint,omitempty"`
	// UserSession  string               `json:"userSession,omitempty" bson:"-"`
	Groups    []UserGroup `json:"groups,omitempty" bson:"-"`
	Users     []string    `json:"users,omitempty" bson:"-"`
	Usernames []string    `json:"usernames,omitempty" bson:"-"`
}

// NodeProject is a bson representation of the ipfs-node setting object
var NodeProject = bson.M{
	"_id":          1,
	"title":        1,
	"creator":      1,
	"type":         1,
	"groups":       UserGroupProject,
	"APIEndpoint":  1,
	"dataEndpoint": 1,
}

// NodeProjectInternal is a bson representation of the ipfs-node setting object
var NodeProjectInternal = bson.M{
	"_id":          1,
	"title":        1,
	"creator":      1,
	"keycloakID":   1,
	"groupIDs":     1,
	"type":         1,
	"APIEndpoint":  1,
	"dataEndpoint": 1,
	"users":        1,
}

// NodeProjectSecret is bson representation of the node to retrieve the secret
var NodeProjectSecret = bson.M{
	"_id":    1,
	"secret": 1,
}

var NodeProjectUserReduction = bson.M{
	"users": bson.M{
		"$reduce": bson.M{
			"input":        "$users",
			"initialValue": bson.A{"$creator"},
			"in": bson.M{
				"$setUnion": bson.A{
					"$$value",
					"$$this",
				},
			},
		},
	},
}

var NodeProjectAuthentication = bson.M{
	"_id":          1,
	"title":        1,
	"creator":      1,
	"groupIDs":     1,
	"secret":       1,
	"type":         1,
	"groups":       UserGroupProject,
	"APIEndpoint":  1,
	"dataEndpoint": 1,
	"usernames": bson.M{
		"$reduce": bson.M{
			"input":        "$usernames",
			"initialValue": bson.A{"$creator"},
			"in": bson.M{
				"$setUnion": bson.A{
					"$$value",
					"$$this",
				},
			},
		},
	},
}

// NodeCollection is the name of the mongo collection
var NodeCollection = "node"

// AddNode creates the model in the mongodb
func (n *Node) AddNode(db *mongo.Database) primitive.ObjectID {
	conn := database.GetColCtx(NodeCollection, db, 30)
	defer conn.Cancel()
	result, err := conn.Col.InsertOne(conn.Ctx, n)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not add new node to database")
		return primitive.NilObjectID
	}

	return result.InsertedID.(primitive.ObjectID)
}

// AddUserGroups adds an array of primitive.ObjectID (of a usergroup) to the
// mapped usergroup set (ignores duplicates) Overrides the current model
// instance
func (n *Node) AddUserGroups(db *mongo.Database, ugIDs []primitive.ObjectID, permission bson.M) error {
	if permission == nil {
		return errors.New("no permission specified")
	}

	conn := database.GetColCtx(NodeCollection, db, 30)
	defer conn.Cancel()

	filter := bson.M{"$and": []bson.M{
		{"_id": n.ID},
		permission,
	}}
	// specify the usergroup array to be handled as set
	update := bson.M{"$addToSet": bson.M{"groupIDs": bson.M{"$each": ugIDs}}}
	// options to return the update document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	log.Warn(n)
	// Execute query
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&n)
	if err != nil {
		return err
	}
	log.Warn(n)
	return nil
}

// DeleteNode deletes the model from the mongodb
func (n *Node) DeleteNode(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := database.GetColCtx(NodeCollection, db, 30)
	filter := bson.M{"_id": n.ID}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
	return result, err
}

// GetAllNodes selects all Nodes from the mongodb
func GetAllNodes(db *mongo.Database, permission bson.M, mode string) ([]Node, error) {
	// create pipeline
	var pipeline []primitive.M
	var err error
	switch mode {
	case "auth":
		pipeline = []bson.M{
			{"$lookup": bson.M{
				"from":         "usergroup",
				"localField":   "groupIDs",
				"foreignField": "_id",
				"as":           "groups",
			}},
			{"$addFields": bson.M{
				"usernames": "$groups.users",
			}},
			{"$project": NodeProjectAuthentication},
		}
		break
	case "internal":
		pipeline, err = database.CreatePermissionProjectPipeline(permission, primitive.NilObjectID, NodeProjectInternal)
		break
	default:
		pipeline, err = database.CreatePermissionProjectPipeline(permission, primitive.NilObjectID, NodeProject)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"mode":  mode,
			"error": err.Error(),
		}).Error("could not create permission pipeline")
		return nil, err
	}
	// execute query
	opts := options.Aggregate()
	conn := database.GetColCtx(NodeCollection, db, 30)
	defer conn.Cancel()
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts) // find all
	if err != nil {
		log.WithFields(log.Fields{
			"mode":  mode,
			"error": err.Error(),
		}).Error("could not select all nodes")
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
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error occured while using the cursor")
		return nil, err
	}
	return nodes, nil
}

// GetNode returns the specified entry from the mongodb
// project -> estimated result representation
func (n *Node) GetNode(db *mongo.Database, permission bson.M, project primitive.M) error {
	// create pipeline
	var pipeline []primitive.M
	var err error
	// create pipeline
	pipeline, err = database.CreatePermissionProjectPipeline(permission, n.ID, project)

	// error handling
	if err != nil {
		return err
	}
	opts := options.Aggregate()
	conn := database.GetColCtx(NodeCollection, db, 30)
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {
		defer conn.Cancel()
		return err
	}
	var found = false
	for cursor.Next(conn.Ctx) {
		err := cursor.Decode(&n)
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

func (n *Node) GetUser(db *mongo.Database) ([]string, int) {
	pipeline := []bson.M{
		{"$match": bson.M{"_id": n.ID}},
		{"$lookup": bson.M{
			"from":         "usergroup",
			"localField":   "groupIDs",
			"foreignField": "_id",
			"as":           "groups",
		}},
		{"$addFields": bson.M{"users": "$groups.users"}},
		{"$project": NodeProjectUserReduction},
		// {"$lookup": bson.M{
		// 	"from":         "user",
		// 	"localField":   "usernames",
		// 	"foreignField": "username",
		// 	"as":           "users",
		// }},
		{"$project": NodeProject},
	}

	var tmp Node
	opts := options.Aggregate()
	conn := database.GetColCtx(NodeCollection, db, 30)
	defer conn.Cancel()
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"node":  n.ID,
			"error": err.Error(),
		}).Error("could not aggregate pipeline")
		return nil, 1
	}
	defer cursor.Close(conn.Ctx)

	for cursor.Next(conn.Ctx) {
		if err = cursor.Decode(&tmp); err != nil {
			log.WithFields(log.Fields{
				"node":  n.ID,
				"error": err.Error(),
			}).Error("could not decode not into struct")
			return nil, 2
		}
	}

	return tmp.Users, 0
}

// Replace replaces the corresponding document in the database with the current state
// 0 -> ok | 1 -> error during find and update
func (n *Node) Replace(db *mongo.Database) int {
	conn := database.GetColCtx(NodeCollection, db, 30)
	filter := bson.M{"_id": n.ID}
	update := bson.M{"$set": n}
	// options to return the update document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// Execute query
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&n)
	defer conn.Cancel()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not find and update node")
		return 1
	}
	return 0
}

// UpdateNodeSecret updates the secret entry of a document
// 0 -> ok
// 1 -> nothing has been changed
// 2 -> update failed
// 3 -> id is zero
// 4 -> permission could not be verified
func (n *Node) UpdateNodeSecret(db *mongo.Database, permission bson.M, secret string) int {
	// verify that user has permission
	if err := n.GetNode(db, permission, NodeProject); err != nil {
		log.WithFields(log.Fields{
			"method": "updateNodeSecret",
			"error":  err.Error(),
		}).Error("failed to check permission")
		return 4
	}
	update := bson.M{"secret": secret}
	return n.updateNodeField(db, update)
}

// updateNodeField takes an update bson, inserts it into a $set and updates the
// current node !!! CHECK PERMISSIONS FIRST !!!
// 0 -> ok
// 1 -> nothing has been changed
// 2 -> update failed
// 3 -> id is zero
func (n *Node) updateNodeField(db *mongo.Database, update bson.M) int {
	if n.ID.IsZero() {
		log.Warn("trying to update on invalid node")
		return 3
	}
	conn := database.GetColCtx(NodeCollection, db, 30)
	defer conn.Cancel()
	filter := bson.M{"_id": n.ID}
	update = bson.M{"$set": update}
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not update field")
		return 2
	}
	if result.ModifiedCount == 0 {
		log.Debug("no field has been changed")
		return 1
	}
	return 0
}

// UpdateNode updates the record with the passed one
func (n *Node) UpdateNode(db *mongo.Database, ue Node, permission bson.M) (*mongo.UpdateResult, error) {
	// check if user is allowed to select this node
	if err := n.GetNode(db, permission, NodeProject); err != nil {
		return nil, err
	}
	// continue with update
	conn := database.GetColCtx(NodeCollection, db, 30)
	defer conn.Cancel()
	filter := bson.M{"_id": n.ID}
	update := bson.M{"$set": ue}
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	return result, err
}

// VerifyNode verifies all mandatory fields of the specified node
// does not verify ID
func (n *Node) VerifyNode(db *mongo.Database, permission bson.M) error {
	if n.Title == "" {
		return errors.New("node title must be set")
	}
	if _, found := helper.FindInSlice([]string{"ipfs", "web"}, n.Type); !found {
		return errors.New("specified type is not valid")
	}
	if n.Creator == "" {
		return errors.New("creator must be specified")
	}
	// if n.DataEndpoint == "" {
	// 	return errors.New("url is not valid")
	// }
	if len(n.GroupIDs) > 0 {
		// select the specified groups
		groups, err := GetUserGroupsByIDs(db, n.GroupIDs, permission)
		if err != nil {
			return err
		}
		// set all groups that could be found
		var tmp []primitive.ObjectID
		for _, g := range groups {
			tmp = append(tmp, g.ID)
		}
		n.GroupIDs = tmp
	}
	return nil
}
