package primboard

import (
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
	conn := GetColCtx(userColName, db, 30)
	result, err := conn.Col.InsertOne(conn.Ctx, u)
	defer conn.Cancel()
	return result, err
}

// DeleteUser deletes the model from the mongodb
func (u *User) DeleteUser(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
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
	conn := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	err := conn.Col.FindOne(conn.Ctx, filter).Decode(&u)
	defer conn.Cancel()
	return err
}

// GetUsers returns a sclice of the specified users from the database
func GetUsers(db *mongo.Database, u []User) ([]User, error) {
	conn := GetColCtx(userColName, db, 30)
	var usernames []string
	// read all usernames
	for _, user := range u {
		usernames = append(usernames, user.Username)
	}
	filter := bson.M{"username": bson.M{"$in": usernames}}

	cursor, err := conn.Col.Find(conn.Ctx, filter)
	if err != nil {
		defer conn.Cancel()
		return []User{}, err
	}
	cursor.All(conn.Ctx, &u)
	defer conn.Cancel()
	return u, nil
}

// UpdateUser updates the record with the passed one
func (u *User) UpdateUser(db *mongo.Database, uu User) (*mongo.UpdateResult, error) {
	conn := GetColCtx(userColName, db, 30)
	filter := bson.M{"username": u.Username}
	update := bson.M{"$set": uu}
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	defer conn.Cancel()
	return result, err
}
