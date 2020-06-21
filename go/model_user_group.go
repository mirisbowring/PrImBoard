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
	if !skipVerify {
		// verify that the fields are valid
		if err := ug.Verify(db); err != nil {
			return nil, err
		}
	}
	conn := GetColCtx(ugColName, db, 30)
	result, err := conn.Col.InsertOne(conn.Ctx, ug)
	defer conn.Cancel()
	return result, err
}

// DeleteUserGroup deletes the model from the mongodb
func (ug *UserGroup) DeleteUserGroup(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := GetColCtx(ugColName, db, 30)
	filter := bson.M{"_id": ug.ID}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
	return result, err
}

// GetUserGroup returns the specified entry from the mongodb
func (ug *UserGroup) GetUserGroup(db *mongo.Database) error {
	conn := GetColCtx(ugColName, db, 30)
	filter := bson.M{"_id": ug.ID}
	err := conn.Col.FindOne(conn.Ctx, filter).Decode(&ug)
	defer conn.Cancel()
	return err
}

// GetUserGroups returns all groups the user is part of
func GetUserGroups(db *mongo.Database, user string) ([]UserGroup, error) {
	conn := GetColCtx(ugColName, db, 30)
	filter := bson.M{"users": user}
	var groups []UserGroup

	cursor, err := conn.Col.Find(conn.Ctx, filter)
	if err != nil {
		defer conn.Cancel()
		return groups, err
	}
	cursor.All(conn.Ctx, &groups)
	defer conn.Cancel()
	return groups, nil
}

// GetUserGroupsByIDs returns a slice of usergroups, that are matching the given id slice
func GetUserGroupsByIDs(db *mongo.Database, ids []primitive.ObjectID) ([]UserGroup, error) {
	conn := GetColCtx(ugColName, db, 30)
	filter := bson.M{"_id": bson.M{"$in": ids}}

	var groups []UserGroup
	cursor, err := conn.Col.Find(conn.Ctx, filter)
	if err != nil {
		defer conn.Cancel()
		return groups, err
	}
	cursor.All(conn.Ctx, &groups)
	defer conn.Cancel()
	return groups, nil
}

// GetUserGroupsByKeyword returns the topmost groups that are starting with the keyword
func GetUserGroupsByKeyword(db *mongo.Database, keyword string, limit int64) ([]UserGroup, error) {
	conn := GetColCtx(ugColName, db, 30)
	// define options (sort, limit, ...)
	options := options.Find()
	options.SetSort(bson.M{"title": 1}).SetLimit(limit)
	// define filter
	filter := bson.M{
		"title": primitive.Regex{Pattern: "^" + keyword, Options: "i"},
	}
	// execute filter query
	var groups []UserGroup
	cursor, err := conn.Col.Find(conn.Ctx, filter, options)
	if err = cursor.All(conn.Ctx, &groups); err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer conn.Cancel()
	return groups, nil
}

// Save writes changes, made to the instance itself, to the database and
// overrides the instance with the return value from the database
func (ug *UserGroup) Save(db *mongo.Database, skipVerify bool) error {
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
	conn := GetColCtx(ugColName, db, 30)
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&ug)
	defer conn.Cancel()
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserGroup updates the record with the passed one
func (ug *UserGroup) UpdateUserGroup(db *mongo.Database, uug UserGroup, skipVerify bool) (*mongo.UpdateResult, error) {
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
	conn := GetColCtx(ugColName, db, 30)
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	defer conn.Cancel()
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
