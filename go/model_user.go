package primboard

import (
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// User contains all information about the user
type User struct {
	Username  string `json:"username" bson:"username"`
	FirstName string `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Password  string `json:"password,omitempty" bson:"password,omitempty"`
	URLImage  string `json:"urlImage,omitempty" bson:"urlImage,omitempty"`
	Token     string `json:"token,omitempty"`
}

//UserProject is a bson representation of the user object
var UserProject = bson.M{
	"username":  1,
	"firstName": 1,
	"lastName":  1,
	"urlImage":  1,
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

// Exists checks if the user can be selected from the database
// Assumes false if any error occurs
func (u *User) Exists(db *mongo.Database) bool {
	if err := u.GetUser(db); err != nil {
		return false
	}
	if u.Username == "" {
		return false
	}
	return true
}

//GetUser returns the specified entry from the mongodb
func (u *User) GetUser(db *mongo.Database) error {
	col, ctx := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	err := col.FindOne(ctx, filter).Decode(&u)
	CloseContext()
	return err
}

// GetUsers returns a sclice of the specified users from the database
func GetUsers(db *mongo.Database, u []User) ([]User, error) {
	col, ctx := GetColCtx(userColName, db, 30)
	var usernames []string
	// read all usernames
	for _, user := range u {
		usernames = append(usernames, user.Username)
	}
	filter := bson.M{"username": bson.M{"$in": usernames}}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		CloseContext()
		return []User{}, err
	}
	cursor.All(ctx, &u)
	CloseContext()
	log.Println(u)
	return u, nil
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
