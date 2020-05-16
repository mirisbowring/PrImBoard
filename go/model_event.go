package primboard

import (
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// name of the mongo collection
var eventColName = "event"

// AddEvent creates the model in the mongodb
func (e *Event) AddEvent(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	result, err := col.InsertOne(ctx, e)
	CloseContext()
	return result, err
}

// BulkAddTagEvent bulk operates a tag slice to  many media ids
func BulkAddTagEvent(db *mongo.Database, tags []string, ids []primitive.ObjectID, permission bson.M) (*mongo.BulkWriteResult, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	opts := options.BulkWrite().SetOrdered(false)
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
	res, err := col.BulkWrite(ctx, models, opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return res, nil
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

// GetEventCreate selects the passed event from database -> creates if not exist
func (e *Event) GetEventCreate(db *mongo.Database) error {
	// read event if ID was specified
	if e.ID.Hex() != "" {
		if err := e.GetEvent(db); err != nil {
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
	if err := e.VerifyEvent(db); err != nil {
		return err
	}
	// insert into db
	res, err := e.AddEvent(db)
	if err != nil {
		return err
	}
	e.ID = res.InsertedID.(primitive.ObjectID)
	if err = e.GetEvent(db); err != nil {
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

	col, ctx := GetColCtx(eventColName, db, 30)
	var events []Event
	cursor, err := col.Find(ctx, filter)
	if err != nil {
		log.Println(err)
		CloseContext()
		return events, err
	}

	cursor.All(ctx, &events)
	CloseContext()
	return events, nil
}

// GetEventsByKeyword returns the topmost events that are starting with the keyword
func GetEventsByKeyword(db *mongo.Database, permission bson.M, keyword string, limit int64) ([]Event, error) {
	col, ctx := GetColCtx(eventColName, db, 30)
	// define options (sort, limit, ...)
	options := options.Find()
	options.SetSort(bson.M{"title": 1}).SetLimit(limit)
	// define filter
	filter := bson.M{"$and": []bson.M{
		{"title": primitive.Regex{Pattern: "^" + keyword, Options: "i"}},
		permission}}
	// execute filter query
	var events []Event
	cursor, err := col.Find(ctx, filter, options)
	if err = cursor.All(ctx, &events); err != nil {
		CloseContext()
		return nil, err
	}
	CloseContext()
	return events, nil

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

// VerifyEvent verifies all mandatory fields of the specified event
// does not verify ID
func (e *Event) VerifyEvent(db *mongo.Database) error {
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
		groups, err := GetUserGroupsByIDs(db, e.Groups)
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
