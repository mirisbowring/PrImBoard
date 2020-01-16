package primboard

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

// Event holds comments, media and the information about the the event
type Event struct {
	ID                primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
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

// name of the mongo collection
var eventColName = "event"

// AddEvent creates the model in the mongodb
func (e *Event) AddEvent(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	result, err := col.InsertOne(ctx, e)
	CloseContext()
	return result, err
}

// DeleteEvent deletes the model from the mongodb
func (e *Event) DeleteEvent(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	filter := bson.M{"_id": e.ID}
	result, err := col.DeleteOne(ctx, filter)
	CloseContext()
	return result, err
}

// GetAllEvents selects all Events from the mongodb
func GetAllEvents(db *mongo.Database) ([]Event, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	cursor, err := col.Find(ctx, bson.M{}) // find all
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	// iterate over the cursor and create array
	var es []Event
	for cursor.Next(ctx) {
		var e Event
		cursor.Decode(&e)
		es = append(es, e)
	}
	// report errors if occured
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	CloseContext()
	return es, nil
}

// GetEvent returns the specified entry from the mongodb
func (e *Event) GetEvent(db *mongo.Database) error {
	col, ctx := GetColCtx(eventColName, db, 30)
	filter := bson.M{"_id": e.ID}
	err := col.FindOne(ctx, filter).Decode(&e)
	CloseContext()
	return err
}

// UpdateEvent updates the record with the passed one
func (e *Event) UpdateEvent(db *mongo.Database, ue Event) (*mongo.UpdateResult, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	filter := bson.M{"_id": e.ID}
	update := bson.M{"$set": ue}
	result, err := col.UpdateOne(ctx, filter, update)
	CloseContext()
	return result, err
}
