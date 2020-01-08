package swagger

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

// Tag has a name and an ID for the reference
type Tag struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name,omitempty" bson:"name,omitempty"`
}

// name of the mongo collection
var tColName = "tag"

// AddTag creates the model in the mongodb
func (t *Tag) AddTag(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	result, err := col.InsertOne(ctx, t)
	CloseContext()
	return result, err
}

// DeleteTag deletes the model from the mongodb
func (t *Tag) DeleteTag(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"_id": t.ID}
	result, err := col.DeleteOne(ctx, filter)
	CloseContext()
	return result, err
}

// GetTag returns the specified entry from the mongodb
func (t *Tag) GetTag(db *mongo.Database) error {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"_id": t.ID}
	err := col.FindOne(ctx, filter).Decode(&t)
	CloseContext()
	return err
}

// UpdateTag updates the record with the passed one
func (t *Tag) UpdateTag(db *mongo.Database, ut Tag) (*mongo.UpdateResult, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"_id": t.ID}
	update := bson.M{"$set": ut}
	result, err := col.UpdateOne(ctx, filter, update)
	CloseContext()
	return result, err
}
