package primboard

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserGroup holts the users and the information about the group
type UserGroup struct {
	ID      primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title   string             `json:"title,omitempty" bson:"title,omitempty"`
	Creator string             `json:"creator,omitempty" bson:"creator,omitempty"`
	Users   []string           `json:"users,omitempty" bson:"users,omitempty"`
}

//UserGroupProject is a bson representation of a user group
var UserGroupProject = bson.M{
	"_id":     1,
	"title":   1,
	"creator": 1,
	"users":   1,
}

// name of the mongo collection
var ugColName = "usergroup"

// AddUserGroup creates the model in the mongodb
func (ug *UserGroup) AddUserGroup(db *mongo.Database, skipVerify bool) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(ugColName, db, 30)
	if !skipVerify {
		// verify that the fields are valid
		if err := ug.Verify(db); err != nil {
			return nil, err
		}
	}
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

// Save writes changes, made to the instance itself, to the database and
// overrides the instance with the return value from the database
func (ug *UserGroup) Save(db *mongo.Database, skipVerify bool) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	if !skipVerify {
		// verify that the fields are valid
		if err := ug.Verify(db); err != nil {
			return err
		}
	}
	filter := bson.M{"_id": ug.ID}
	update := bson.M{"$set": ug}
	// options to return the update document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// Execute query
	err := col.FindOneAndUpdate(ctx, filter, update, &options).Decode(&ug)
	CloseContext()
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserGroup updates the record with the passed one
func (ug *UserGroup) UpdateUserGroup(db *mongo.Database, uug UserGroup, skipVerify bool) (*mongo.UpdateResult, error) {
	col, ctx := GetColCtx(ugColName, db, 30)
	if !skipVerify {
		// verify that the fields are valid
		if err := uug.Verify(db); err != nil {
			return nil, err
		}
	}
	// ID should not be changes
	if ug.ID != uug.ID {
		uug.ID = ug.ID
	}
	filter := bson.M{"_id": ug.ID}
	update := bson.M{"$set": uug}
	result, err := col.UpdateOne(ctx, filter, update)
	CloseContext()
	return result, err
}

// Verify tries to verify the usergroup object
func (ug *UserGroup) Verify(db *mongo.Database) error {
	// verify title
	if ug.Title == "" {
		return errors.New("the title is not set")
	}
	// verify creator
	u := User{Username: ug.Creator}
	if !u.Exists(db) {
		return errors.New("the creator does not exist")
	}
	// verify that creator is in group
	if _, found := Find(ug.Users, ug.Creator); !found {
		ug.Users = append(ug.Users, ug.Creator)
	}
	// verify that there are no duplicates
	if len(ug.Users) > 1 {
		ug.Users = UniqueStrings(ug.Users)
	}

	return nil
}
