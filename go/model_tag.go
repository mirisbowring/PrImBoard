/*
 * PrImBoard
 *
 * PrImBoard (Private Image Board) can be best described as an image board for all the picures and videos you have taken. You can invite users to the board and share specific images with them or your family members!
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type Tag struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name,omitempty" bson:"name,omitempty"`
}

// name of the mongo collection
var tColName = "tag"

/*
 * Creates the model in the mongodb
 */
func (t *Tag) AddTag(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	result, err := col.InsertOne(ctx, t)
	return result, err
}

/*
 * Deletes the model from the mongodb
 */
 func (t *Tag) DeleteTag(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"_id": t.Id}
	result, err := col.DeleteOne(ctx, filter)
	return result, err
}

/*
 * Returns the specified entry from the mongodb
 */
func (t *Tag) GetTag(db *mongo.Database) error {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"_id": t.Id}
	err := col.FindOne(ctx, filter).Decode(&t)
	return err
}

/*
 * Updates the record with the passed one
 */
func (t *Tag) UpdateTag(db *mongo.Database, ut Tag) (*mongo.UpdateResult, error) {
	col, ctx := GetColCtx(tColName, db, 30)
	filter := bson.M{"_id": t.Id}
	update := bson.M{"$set": ut}
	result, err := col.UpdateOne(ctx, filter, update)
	return result, err
}