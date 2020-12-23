package models

// import (
// 	"github.com/mirisbowring/primboard/helper/database"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // User contains all information about the user
// type User struct {
// 	Username  string    `json:"username" bson:"username"`
// 	FirstName string    `json:"firstName,omitempty" bson:"firstName,omitempty"`
// 	LastName  string    `json:"lastName,omitempty" bson:"lastName,omitempty"`
// 	Password  string    `json:"password,omitempty" bson:"password,omitempty"`
// 	URLImage  string    `json:"urlImage,omitempty" bson:"urlImage,omitempty"`
// 	Token     string    `json:"token,omitempty"`
// 	Settings  *Settings `json:"settings,omitempty" bson:"settings,omitempty"`
// }

// //UserProject is a bson representation of the user object
// var UserProject = bson.M{
// 	"username":  1,
// 	"firstName": 1,
// 	"lastName":  1,
// 	"urlImage":  1,
// }

// var UserSettingsProject = bson.M{
// 	"settings": SettingsProject,
// }

// // name of the mongo collection
// var UserCollection = "user"

// // CreateUser creates the user model in the mongodb
// func (u *User) CreateUser(db *mongo.Database) (*mongo.InsertOneResult, error) {
// 	conn := database.GetColCtx(UserCollection, db, 30)
// 	result, err := conn.Col.InsertOne(conn.Ctx, u)
// 	defer conn.Cancel()
// 	return result, err
// }

// // DeleteUser deletes the model from the mongodb
// func (u *User) DeleteUser(db *mongo.Database) (*mongo.DeleteResult, error) {
// 	conn := database.GetColCtx(UserCollection, db, 30)
// 	filter := bson.M{"username": u.Username}
// 	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
// 	defer conn.Cancel()
// 	return result, err
// }

// // Exists checks if the user can be selected from the database
// // Assumes false if any error occurs
// func (u *User) Exists(db *mongo.Database) bool {
// 	if err := u.GetUser(db); err != nil {
// 		return false
// 	}
// 	if u.Username == "" {
// 		return false
// 	}
// 	return true
// }

// //GetUser returns the specified entry from the mongodb
// func (u *User) GetUser(db *mongo.Database) error {
// 	conn := database.GetColCtx(UserCollection, db, 30)
// 	filter := bson.M{"username": u.Username}
// 	err := conn.Col.FindOne(conn.Ctx, filter).Decode(&u)
// 	defer conn.Cancel()
// 	return err
// }

// // GetUsers returns a sclice of the specified users from the database
// func GetUsers(db *mongo.Database, u []User) ([]User, error) {
// 	conn := database.GetColCtx(UserCollection, db, 30)
// 	var usernames []string
// 	// read all usernames
// 	for _, user := range u {
// 		usernames = append(usernames, user.Username)
// 	}
// 	filter := bson.M{"username": bson.M{"$in": usernames}}

// 	cursor, err := conn.Col.Find(conn.Ctx, filter)
// 	if err != nil {
// 		defer conn.Cancel()
// 		return []User{}, err
// 	}
// 	cursor.All(conn.Ctx, &u)
// 	defer conn.Cancel()
// 	return u, nil
// }

// // Save writes changes, made to the instance itself, to the database and
// // overrides the instance with the return value from the database
// func (u *User) Save(db *mongo.Database) error {
// 	filter := bson.M{"username": u.Username}
// 	update := bson.M{"$set": u}
// 	// options to return the update document
// 	after := options.After
// 	upsert := true
// 	options := options.FindOneAndUpdateOptions{
// 		ReturnDocument: &after,
// 		Upsert:         &upsert,
// 	}
// 	// Execute query
// 	conn := database.GetColCtx(UserCollection, db, 30)
// 	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&u)
// 	defer conn.Cancel()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // UpdateUser updates the record with the passed one
// func (u *User) UpdateUser(db *mongo.Database, uu User) (*mongo.UpdateResult, error) {
// 	conn := database.GetColCtx(UserCollection, db, 30)
// 	filter := bson.M{"username": u.Username}
// 	update := bson.M{"$set": uu}
// 	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
// 	defer conn.Cancel()
// 	return result, err
// }
