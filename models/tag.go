package models

import (
	"log"
	"strings"

	"github.com/mirisbowring/primboard/helper"
	"github.com/mirisbowring/primboard/helper/database"
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

// TagCollection is the name of the mongo collection
var TagCollection = "tag"

// AddTag creates the model in the mongodb
func (t *Tag) AddTag(db *mongo.Database) (*mongo.InsertOneResult, error) {
	conn := database.GetColCtx(TagCollection, db, 30)
	result, err := conn.Col.InsertOne(conn.Ctx, t)
	defer conn.Cancel()
	return result, err
}

// addTags iterates over an array of tags and adds new one to the db
func addTags(db *mongo.Database, tags []Tag) error {
	for i := range tags {
		if !tags[i].ID.IsZero() {
			if err := tags[i].GetTag(db); err == nil {
				// tag is valid and in db
				continue
			}
			// tag is invalid and needs to be stored in the db
			tags[i].ID = primitive.NilObjectID
		}
		tags[i].Name = strings.TrimSpace(tags[i].Name)
		// check if tag exists already
		if err := tags[i].GetTagByName(db); err != nil {
			switch err {
			case mongo.ErrNoDocuments:
				res, _ := tags[i].AddTag(db)
				tags[i].ID = res.InsertedID.(primitive.ObjectID)
				log.Printf("Created new tag <%s>", tags[i].Name)
			default:
				return err
			}
		}
	}
	return nil
}

// DeleteTag deletes the model from the mongodb
func (t *Tag) DeleteTag(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := database.GetColCtx(TagCollection, db, 30)
	filter := bson.M{"_id": t.ID}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
	return result, err
}

// GetIDCreate searches the database for the passed tag and adds the id to the
// current tag. It creates a new tag document if the passed tag was not find in
// the database
func (t *Tag) GetIDCreate(db *mongo.Database) error {
	// try to select tag from db
	if err := t.GetTagByName(db); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// tag not found - adding to db
			// t.Name = strings.ToLower(t.Name)
			t.Name = strings.TrimSpace(t.Name)
			res, e := t.AddTag(db)
			if e != nil {
				return e
			}
			// append the returned id
			t.ID = res.InsertedID.(primitive.ObjectID)
			return nil
		default:
			// any error occured
			return err
		}
	}
	// tag was found and id was set
	return nil
}

// GetTag returns the specified entry from the mongodb
func (t *Tag) GetTag(db *mongo.Database) error {
	conn := database.GetColCtx(TagCollection, db, 30)
	filter := bson.M{"_id": t.ID}
	err := conn.Col.FindOne(conn.Ctx, filter).Decode(&t)
	defer conn.Cancel()
	return err
}

// GetTagByName returns the specified entry from the mongodb
func (t *Tag) GetTagByName(db *mongo.Database) error {
	conn := database.GetColCtx(TagCollection, db, 30)
	filter := bson.M{"name": bson.M{"$regex": t.Name, "$options": "i"}}
	err := conn.Col.FindOne(conn.Ctx, filter).Decode(&t)
	defer conn.Cancel()
	return err
}

// GetTagsByKeyword returns the topmost tags that are starting with the keyword
func GetTagsByKeyword(db *mongo.Database, keyword string, limit int) ([]Tag, error) {
	conn := database.GetColCtx(TagCollection, db, 30)
	// define options (sort, limit, ...)
	options := options.Find()
	options.SetSort(bson.M{"name": 1}).SetLimit(int64(limit))
	// define filter
	filter := bson.M{
		"name": primitive.Regex{Pattern: "^" + keyword, Options: "i"},
	}
	// execute filter query
	var tags []Tag
	cursor, err := conn.Col.Find(conn.Ctx, filter, options)
	if err = cursor.All(conn.Ctx, &tags); err != nil {
		defer conn.Cancel()
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
	defer conn.Cancel()
	return tags, nil

}

// UpdateTag updates the record with the passed one
func (t *Tag) UpdateTag(db *mongo.Database, ut Tag) (*mongo.UpdateResult, error) {
	conn := database.GetColCtx(TagCollection, db, 30)
	filter := bson.M{"_id": t.ID}
	update := bson.M{"$set": ut}
	result, err := conn.Col.UpdateOne(conn.Ctx, filter, update)
	defer conn.Cancel()
	return result, err
}

// VerifyTag creates the tag if not in the db already and returns the name in
// the db
func VerifyTag(db *mongo.Database, tag string) (string, error) {
	// getting or creating the new tag
	tmp := Tag{Name: tag}
	err := tmp.GetIDCreate(db)
	// returning tag name
	return tmp.Name, err
}

// VerifyTags iterates over an array of tags and creates the tags if not in db
// already. Uniquifies the slice and returns the clean slice.
func VerifyTags(db *mongo.Database, tags []string) ([]string, error) {
	var cleanTags []string
	for _, tag := range tags {
		// find or create tag
		t, err := VerifyTag(db, tag)
		if err != nil {
			return nil, err
		}
		cleanTags = append(cleanTags, t)
	}
	// uniquify tags
	cleanTags = helper.UniqueStrings(cleanTags)
	return cleanTags, nil
}
