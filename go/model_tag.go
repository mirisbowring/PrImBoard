package primboard

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Tag has a name and an ID for the reference
type Tag struct {
	ID   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
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

// GetTagByName returns the specified entry from the mongodb
func (t *Tag) GetTagByName(db *mongo.Database) error {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"name": t.Name}
	err := col.FindOne(ctx, filter).Decode(&t)
	CloseContext()
	return err
}

// GetTagsByKeyword returns the topmost tags that are starting with the keyword
func GetTagsByKeyword(db *mongo.Database, keyword string, limit int64) ([]Tag, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	// define options (sort, limit, ...)
	options := options.Find()
	options.SetSort(bson.M{"name": 1}).SetLimit(limit)
	// define filter
	filter := bson.M{
		"name": primitive.Regex{Pattern: "^" + keyword},
	}
	// execute filter query
	var tags []Tag
	cursor, err := col.Find(ctx, filter, options)
	if err = cursor.All(ctx, &tags); err != nil {
		return nil, err
	}

	// cursor, err := col.Find(ctx, filter)
	// if err != nil {
	// 	return nil, err
	// }
	// var tags []Tag
	// // iterate over cursor and map tags
	// for cursor.Next(ctx) {
	// 	var t Tag
	// 	cursor.Decode(&t)
	// 	tags = append(tags, t)
	// }
	// // report errors if occured
	// if err = cursor.Err(); err != nil {
	// 	return nil, err
	// }
	CloseContext()
	return tags, nil

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
