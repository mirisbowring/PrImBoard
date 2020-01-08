package swagger

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

// UserGroup holts the users and the information about the group
type UserGroup struct {
	ID                primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title             string             `json:"title,omitempty" bson:"title,omitempty"`
	Creator           string             `json:"creator,omitempty" bson:"creator,omitempty"`
	TimestampCreation int64              `json:"timestamp_creation,omitempty" bson:"timestamp_creation,omitempty"`
	Users             []string           `json:"users,omitempty" bson:"users,omitempty"`
}

// name of the mongo collection
var ugColName = "usergroup"

// AddUserGroup creates the model in the mongodb
func (ug *UserGroup) AddUserGroup(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(ugColName, db, 30)
	result, err := col.InsertOne(ctx, ug)
	CloseContext()
	return result, err
}

// DeleteUserGroup deletes the model from the mongodb
func (ug *UserGroup) DeleteUserGroup(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(ugColName, db, 30)
	filter := bson.M{"_id": ug.ID}
	result, err := col.DeleteOne(ctx, filter)
	CloseContext()
	return result, err
}

// GetUserGroup returns the specified entry from the mongodb
func (ug *UserGroup) GetUserGroup(db *mongo.Database) error {
	col, ctx := GetColCtx(ugColName, db, 30)
	filter := bson.M{"_id": ug.ID}
	err := col.FindOne(ctx, filter).Decode(&ug)
	CloseContext()
	return err
}

// UpdateUserGroup updates the record with the passed one
func (ug *UserGroup) UpdateUserGroup(db *mongo.Database, uug UserGroup) (*mongo.UpdateResult, error) {
	col, ctx := GetColCtx(ugColName, db, 30)
	filter := bson.M{"_id": ug.ID}
	update := bson.M{"$set": uug}
	result, err := col.UpdateOne(ctx, filter, update)
	CloseContext()
	return result, err
}
