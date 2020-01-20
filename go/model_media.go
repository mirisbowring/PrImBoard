package primboard

import (
	"errors"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Media holds all information about the item
type Media struct {
	ID              primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Sha1            string               `json:"sha1,omitempty" bson:"sha1,omitempty"`
	Title           string               `json:"title,omitempty" bson:"title,omitempty"`
	Description     string               `json:"description,omitempty" bson:"description,omitempty"`
	Comments        []*Comment           `json:"comments,omitempty" bson:"comments,omitempty"`
	Creator         string               `json:"creator,omitempty" bson:"creator,omitempty"`
	Tags            []Tag                `json:"tags,omitempty" bson:"tags,omitempty"`
	Events          []primitive.ObjectID `json:"events,omitempty" bson:"events,omitempty"`
	Groups          []primitive.ObjectID `json:"groups,omitempty" bson:"groups,omitempty"`
	Timestamp       int64                `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	TimestampUpload int64                `json:"timestampUpload,omitempty" bson:"timestampUpload,omitempty"`
	URL             string               `json:"url,omitempty" bson:"url,omitempty"`
	URLThumb        string               `json:"urlThumb,omitempty" bson:"urlThumb,omitempty"`
	Type            string               `json:"type,omitempty" bson:"type,omitempty"`
	Format          string               `json:"format,omitempty" bson:"format,omitempty"`
}

// name of the mongo collection
var mediaColName = "media"

// AddMedia creates the model in the mongodb
func (m *Media) AddMedia(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(mediaColName, db, 30)
	m.checkTags(db)
	result, err := col.InsertOne(ctx, m)
	CloseContext()
	return result, err
}

// DeleteMedia deletes the model from the mongodb
func (m *Media) DeleteMedia(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	result, err := col.DeleteOne(ctx, filter)
	CloseContext()
	return result, err
}

// GetAllMedia selects all Media from the mongodb
func GetAllMedia(db *mongo.Database) ([]Media, error) {
	col, ctx := GetColCtx(mediaColName, db, 30)
	cursor, err := col.Find(ctx, bson.M{}) //find all
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	// iterate over the cursor and create array
	var ms []Media
	for cursor.Next(ctx) {
		var m Media
		cursor.Decode(&m)
		ms = append(ms, m)
	}
	// report errors if occured
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	CloseContext()
	return ms, nil
}

// GetMedia returns the specified entry from the mongodb
func (m *Media) GetMedia(db *mongo.Database) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	err := col.FindOne(ctx, filter).Decode(&m)
	CloseContext()
	return err
}

// UpdateMedia updates the record with the passed one
func (m *Media) UpdateMedia(db *mongo.Database, um Media) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	um.checkTags(db)
	filter := bson.M{"_id": m.ID}
	update := bson.M{"$set": um}
	// options to return the update document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// Execute query
	err := col.FindOneAndUpdate(ctx, filter, update, &options).Decode(&m)
	CloseContext()
	if err != nil {
		return err
	}
	return nil
}

// checkComments verifies that only one new comment has been passed and assignes
// the passed username and the current unix timestamp to the new comment
// this method is NOT called on create or update
func (m *Media) checkComments(user string) error {
	// skip if comments are nil or empty
	if m.Comments == nil || len(m.Comments) == 0 {
		return nil
	}
	// check if more than one new comment added
	var newComments int = 0
	for i := range m.Comments {
		if m.Comments[i].Timestamp != 0 && m.Comments[i].Username != "" {
			continue
		}
		// increment new comment counter
		newComments++
		// throw error if more than one new comment
		if newComments > 1 {
			return errors.New("too many new comments")
		}
		// assign username to new comment
		m.Comments[i].Username = user
		m.Comments[i].Timestamp = int64(time.Now().Unix())
	}
	return nil
}

// checkTags iterates over the tag array of the media and adds new tags to the
// tag collection
func (m *Media) checkTags(db *mongo.Database) error {
	for i := range m.Tags {
		if !m.Tags[i].ID.IsZero() {
			continue
		}
		m.Tags[i].Name = strings.ToLower(m.Tags[i].Name)
		m.Tags[i].Name = strings.TrimSpace(m.Tags[i].Name)
		// check of tag exist already
		if err := m.Tags[i].GetTagByName(db); err != nil {
			switch err {
			case mongo.ErrNoDocuments:
				res, _ := m.Tags[i].AddTag(db)
				m.Tags[i].ID = res.InsertedID.(primitive.ObjectID)
				log.Printf("Created new tag <%s>", m.Tags[i].Name)
			default:
				return err
			}
		}
	}
	return nil
}
