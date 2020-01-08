package swagger

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

// User contains all information about the user
type User struct {
	Username  string `json:"username" bson:"username"`
	FirstName string `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Password  string `json:"password,omitempty" bson:"password,omitempty"`
}

// name of the mongo collection
var userColName = "user"

// CreateUser creates the user model in the mongodb
func (u *User) CreateUser(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(userColName, db, 30)
	result, err := col.InsertOne(ctx, u)
	CloseContext()
	return result, err
}

// DeleteUser deletes the model from the mongodb
func (u *User) DeleteUser(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	result, err := col.DeleteOne(ctx, filter)
	CloseContext()
	return result, err
}

//GetUser returns the specified entry from the mongodb
func (u *User) GetUser(db *mongo.Database) error {
	col, ctx := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	err := col.FindOne(ctx, filter).Decode(&u)
	CloseContext()
	return err
}

// UpdateUser updates the record with the passed one
func (u *User) UpdateUser(db *mongo.Database, uu User) (*mongo.UpdateResult, error) {
	col, ctx := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	update := bson.M{"$set": uu}
	result, err := col.UpdateOne(ctx, filter, update)
	CloseContext()
	return result, err
}
