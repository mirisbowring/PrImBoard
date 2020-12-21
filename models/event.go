package models

import (
	"errors"
	"time"

	"github.com/mirisbowring/primboard/helper/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

// Event holds comments, media and the information about the the event
type Event struct {
	ID                primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Title             string               `json:"title,omitempty" bson:"title,omitempty"`
	Description       string               `json:"description,omitempty" bson:"description,omitempty"`
	Comments          []*Comment           `json:"comments,omitempty" bson:"comments,omitempty"`
	Creator           string               `json:"creator,omitempty" bson:"creator,omitempty"`
	Groups            []primitive.ObjectID `json:"groups,omitempty" bson:"groups,omitempty"`
	TimestampCreation int64                `json:"timestampCreation,omitempty" bson:"timestampCreation,omitempty"`
	TimestampStart    int64                `json:"timestampStart,omitempty" bson:"timestampStart,omitempty"`
	TimestampEnd      int64                `json:"timestampEnd,omitempty" bson:"timestampEnd,omitempty"`
	URL               string               `json:"url,omitempty" bson:"url,omitempty"`
	URLThumb          string               `json:"urlThumb,omitempty" bson:"urlThumb,omitempty"`
}

// EventProject is a bson representation of the event object
var EventProject = bson.M{
	"id":                1,
	"title":             1,
	"description":       1,
	"comments":          1,
	"creator":           1,
	"groups":            1,
	"timestampCreation": 1,
	"timestampStart":    1,
	"timestampEnd":      1,
	"url":               1,
	"urlThumb":          1,
}

// name of the mongo collection
var eventColName = "event"

// AddEvent creates the model in the mongodb
func (e *Event) AddEvent(db *mongo.Database) (*mongo.InsertOneResult, error) {
	conn := database.GetColCtx(eventColName, db, 30)
	result, err := conn.Col.InsertOne(conn.Ctx, e)
	defer conn.Cancel()
	return result, err
}

// BulkAddTagEvent bulk operates a tag slice to  many media ids
func BulkAddTagEvent(db *mongo.Database, tags []string, ids []primitive.ObjectID, permission bson.M) (*mongo.BulkWriteResult, error) {
	// create update list
	models := []mongo.WriteModel{}
	for _, id := range ids {
		filter := bson.M{"$and": []bson.M{
			{"_id": id},
			permission}}
		update := bson.M{"$addToSet": bson.M{"tags": bson.M{"$each": tags}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}
	// execute bulk update
	conn := database.GetColCtx(eventColName, db, 30)
	opts := options.BulkWrite().SetOrdered(false)
	res, err := conn.Col.BulkWrite(conn.Ctx, models, opts)
	if err != nil {
		defer conn.Cancel()
		log.Println(err)
		return nil, err
	}
	defer conn.Cancel()
	return res, nil
}

// DeleteEvent deletes the model from the mongodb
func (e *Event) DeleteEvent(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := database.GetColCtx(eventColName, db, 30)
	filter := bson.M{"_id": e.ID}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
	return result, err
}

// GetAllEvents selects all Events from the mongodb
func GetAllEvents(db *mongo.Database) ([]Event, error) {
	conn := database.GetColCtx(eventColName, db, 30)
	cursor, err := conn.Col.Find(conn.Ctx, bson.M{}) // find all
	if err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer cursor.Close(conn.Ctx)
	// iterate over the cursor and create array
	var es []Event
	for cursor.Next(conn.Ctx) {
		var e Event
		cursor.Decode(&e)
		es = append(es, e)
	}
	// report errors if occured
	if err = cursor.Err(); err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer conn.Cancel()
	return es, nil
}

// GetEvent returns the specified entry from the mongodb
func (e *Event) GetEvent(db *mongo.Database, permission bson.M) error {
	// create pipeline
	pipeline, err := database.CreatePermissionProjectPipeline(permission, e.ID, EventProject)
	if err != nil {
		return err
	}
	opts := options.Aggregate()
	conn := database.GetColCtx(eventColName, db, 30)
	defer conn.Cancel()
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {

		return err
	}
	defer cursor.Close(conn.Ctx)

	var found = false
	for cursor.Next(conn.Ctx) {
		err := cursor.Decode(&e)
		if err != nil {
			return err
		}
		found = true
		break
	}

	if !found {
		return mongo.ErrNoDocuments
	}
	return nil
}

// GetEventCreate selects the passed event from database -> creates if not exist
func (e *Event) GetEventCreate(db *mongo.Database, permission bson.M, creator string) error {
	// read event if ID was specified
	if e.ID.Hex() != "" && !e.ID.IsZero() {
		if err := e.GetEvent(db, permission); err != nil {
			log.Error(err)
			switch err {
			case mongo.ErrNoDocuments:
				// event does not exist and should be created
				e.ID = primitive.NilObjectID
			default:
				return err
			}
		} else {
			return nil
		}
	}
	// create event instead
	e.Creator = creator
	e.TimestampCreation = int64(time.Now().Unix())
	if err := e.VerifyEvent(db, permission); err != nil {
		return err
	}
	// insert into db
	res, err := e.AddEvent(db)
	if err != nil {
		return err
	}
	e.ID = res.InsertedID.(primitive.ObjectID)
	if err = e.GetEvent(db, permission); err != nil {
		return err
	}
	return nil
}

// GetEventsByIDs selects multiple Media Documents for the passed ids.
// verifies the reading permissions
func GetEventsByIDs(db *mongo.Database, ids []primitive.ObjectID, permission bson.M) ([]Event, error) {
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}
	filter := bson.M{"$and": []bson.M{
		{"_id": bson.M{"$in": ids}},
		permission}}

	conn := database.GetColCtx(eventColName, db, 30)
	defer conn.Cancel()
	var events []Event
	cursor, err := conn.Col.Find(conn.Ctx, filter)
	if err != nil {
		log.Println(err)
		return events, err
	}
	defer cursor.Close(conn.Ctx)

	cursor.All(conn.Ctx, &events)
	return events, nil
}

// GetEventsByKeyword returns the topmost events that are starting with the keyword
func GetEventsByKeyword(db *mongo.Database, permission bson.M, keyword string, limit int) ([]Event, error) {
	conn := database.GetColCtx(eventColName, db, 30)
	// define options (sort, limit, ...)
	options := options.Find()
	options.SetSort(bson.M{"title": 1}).SetLimit(int64(limit))
	// define filter
	filter := bson.M{"$and": []bson.M{
		{"title": primitive.Regex{Pattern: "^" + keyword, Options: "i"}},
		permission}}
	// execute filter query
	var events []Event
	cursor, err := conn.Col.Find(conn.Ctx, filter, options)
	if err = cursor.All(conn.Ctx, &events); err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer conn.Cancel()
	return events, nil

}

// UpdateEvent updates the record with the passed one
func (e *Event) UpdateEvent(db *mongo.Database, ue Event, permission bson.M) (*mongo.UpdateResult, error) {
	// check if user is allowed to select this node
	if err := e.GetEvent(db, permission); err != nil {
		return nil, err
	}
	conn := database.GetColCtx(eventColName, db, 30)
	filter := bson.M{"_id": e.ID}
	update := bson.M{"$set": ue}
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	defer conn.Cancel()
	return result, err
}

// VerifyEvent verifies all mandatory fields of the specified event
// does not verify ID
func (e *Event) VerifyEvent(db *mongo.Database, permission bson.M) error {
	if e.Title == "" {
		return errors.New("event title must be set")
	}
	if e.Creator == "" {
		return errors.New("creator must be specified")
	}
	if e.TimestampCreation == 0 {
		return errors.New("creation timestamp was not set")
	}
	if len(e.Groups) > 0 {
		groups, err := GetUserGroupsByIDs(db, e.Groups, permission)
		if err != nil {
			return err
		}
		var tmp []primitive.ObjectID
		for _, g := range groups {
			tmp = append(tmp, g.ID)
		}
		e.Groups = tmp
	}
	return nil
}
