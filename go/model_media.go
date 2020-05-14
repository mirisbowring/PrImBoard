package primboard

import (
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Media holds all information about the item
type Media struct {
	ID              primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Sha1            string               `json:"sha1,omitempty" bson:"sha1,omitempty"`
	Title           string               `json:"title,omitempty" bson:"title,omitempty"`
	Description     string               `json:"description,omitempty" bson:"description,omitempty"`
	Comments        []*Comment           `json:"comments,omitempty" bson:"comments,omitempty"`
	Creator         string               `json:"creator,omitempty" bson:"creator,omitempty"`
	TagIDs          []primitive.ObjectID `json:"tagIDs,omitempty" bson:"tagIDs,omitempty"`
	Events          []primitive.ObjectID `json:"events,omitempty" bson:"events,omitempty"`
	GroupIDs        []primitive.ObjectID `json:"groupIDs,omitempty" bson:"groupIDs,omitempty"`
	Timestamp       int64                `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	TimestampUpload int64                `json:"timestampUpload,omitempty" bson:"timestampUpload,omitempty"`
	URL             string               `json:"url,omitempty" bson:"url,omitempty"`
	URLThumb        string               `json:"urlThumb,omitempty" bson:"urlThumb,omitempty"`
	Type            string               `json:"type,omitempty" bson:"type,omitempty"`
	Format          string               `json:"format,omitempty" bson:"format,omitempty"`
	Tags            []string             `json:"tags,omitempty"`
	Users           []User               `json:"users,omitempty"`
	Groups          []UserGroup          `json:"groups,omitempty"`
}

//MediaProject is a bson representation of the $project aggregation for mongodb
var MediaProject = bson.M{
	"_id":             1,
	"sha1":            1,
	"title":           1,
	"description":     1,
	"comments":        1,
	"creator":         1,
	"events":          1,
	"timestamp":       1,
	"timestampUpload": 1,
	"url":             1,
	"urlThumb":        1,
	"type":            1,
	"format":          1,
	"tags":            "$tags.name",
	"users":           UserProject,
	"groups":          UserGroupProject,
}

// CreatePermissionFilter creates a filter bson that matches the owner and it's groups
func CreatePermissionFilter(groups []primitive.ObjectID, user string) bson.M {
	filters := []bson.M{}
	// username must be passed
	if user == "" {
		return bson.M{}
	}
	filters = append(filters, bson.M{"creator": user})
	// add groups if passed
	if groups != nil && len(groups) > 0 {
		filters = append(filters, bson.M{"groupIDs": bson.M{"$in": groups}})
	}

	return bson.M{"$or": filters}

}

// name of the mongo collection
var mediaColName = "media"

// AddMedia creates the model in the mongodb
func (m *Media) AddMedia(db *mongo.Database) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(mediaColName, db, 30)
	result, err := col.InsertOne(ctx, m)
	CloseContext()
	return result, err
}

// AddTag adds a tag primitive.ObjectID to the mapped tag set (ignores
// duplicates)
// Overrides the current model instance
func (m *Media) AddTag(db *mongo.Database, t primitive.ObjectID) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	// specify the tag array to be handled as set
	update := bson.M{"$addToSet": bson.M{"tagIDs": t}}
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

// AddTags adds an array of primitive.ObjectID (of a tag) to the mapped tag set
// (ignores duplicates)
// Overrides the current model instance
func (m *Media) AddTags(db *mongo.Database, tIDs []primitive.ObjectID) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	// specify the tag array to be handled as set
	update := bson.M{"$addToSet": bson.M{"tagIDs": bson.M{"$each": tIDs}}}
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

// AddUserGroups adds an array of primitive.ObjectID (of a usergroup) to the
// mapped usergroup set (ignores duplicates) Overrides the current model
// instance
func (m *Media) AddUserGroups(db *mongo.Database, ugIDs []primitive.ObjectID) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	// specify the usergroup array to be handled as set
	update := bson.M{"$addToSet": bson.M{"groupIDs": bson.M{"$each": ugIDs}}}
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

// DeleteMedia deletes the model from the mongodb
func (m *Media) DeleteMedia(db *mongo.Database) (*mongo.DeleteResult, error) {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	result, err := col.DeleteOne(ctx, filter)
	CloseContext()
	return result, err
}

// GetMediaPage returns the requested page after a specific id
func GetMediaPage(db *mongo.Database, query MediaQuery, permission bson.M) ([]Media, error) {
	// verify that query combination is able to be filtered
	if err := query.IsValid(); err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}
	// parse the order of

	opts := options.Find().SetSort(bson.M{"_id": query.ASC}).SetLimit(int64(query.Size))
	filters := []bson.M{}
	// check if an previous page item was passed
	if !query.After.IsZero() {
		filters = append(filters, bson.M{"_id": bson.M{"$lt": query.After}})
	}
	// check if an previous page item was passed
	if !query.Before.IsZero() {
		filters = append(filters, bson.M{"_id": bson.M{"$gt": query.Before}})
	}
	if !query.From.IsZero() {
		filters = append(filters, bson.M{"_id": bson.M{"$lte": query.From}})
	}
	if !query.Until.IsZero() {
		filters = append(filters, bson.M{"_id": bson.M{"$gte": query.Until}})
	}
	// check if event was specified
	if !query.Event.IsZero() {
		filters = append(filters, bson.M{"mo": query.Event})
	}
	// check if filter have been specified
	if len(query.Filter) > 0 {
		tags := parseTags(db, query.Filter)
		log.Println(tags)
		filters = append(filters, bson.M{"tagIDs": bson.M{"$all": tags}})
	}
	filters = append(filters, permission)

	// create empty bson if no filter specified to prevent npe
	var tmp bson.M
	if len(filters) > 0 {
		tmp = bson.M{"$and": filters}
	} else {
		tmp = bson.M{}
	}

	col, ctx := GetColCtx(mediaColName, db, 30)

	// fetch results
	var media []Media
	cursor, err := col.Find(ctx, tmp, opts)
	if err != nil {
		log.Println(err)
		CloseContext()
		return media, err
	}

	cursor.All(ctx, &media)
	CloseContext()
	return media, nil
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
func (m *Media) GetMedia(db *mongo.Database, permission bson.M) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	if permission == nil {
		return errors.New("no permissions specified")
	}
	filter := bson.M{"$and": []bson.M{
		{"_id": m.ID},
		permission,
	}}
	// filter := bson.M{"_id": m.ID}
	pipeline := []bson.M{
		{"$match": filter},
		{"$lookup": bson.M{
			"from":         "user",
			"localField":   "comments.username",
			"foreignField": "username",
			"as":           "users",
		}},
		{"$lookup": bson.M{
			"from":         "tag",
			"localField":   "tagIDs",
			"foreignField": "_id",
			"as":           "tags",
		}},
		{"$lookup": bson.M{
			"from":         "usergroup",
			"localField":   "groupIDs",
			"foreignField": "_id",
			"as":           "groups",
		}},
		{"$project": MediaProject},
	}
	// append permission
	// pipeline = append(pipeline, permission)
	// pipe := []bson.M{
	// 	bson.M{"$project"}
	// }
	opts := options.Aggregate()
	cursor, err := col.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return err
	}
	var found = false
	for cursor.Next(ctx) {
		err := cursor.Decode(&m)
		if err != nil {
			CloseContext()
			return err
		}
		found = true
		break
	}
	CloseContext()
	if !found {
		return errors.New("no results")
	}
	return nil
}

// Save writes changes, made to the instance itself, to the database and
// overrides the instance with the return value from the database
func (m *Media) Save(db *mongo.Database) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
	filter := bson.M{"_id": m.ID}
	update := bson.M{"$set": m}
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

// UpdateMedia updates the record with the passed one
// Does NOT call the checkComments Method
func (m *Media) UpdateMedia(db *mongo.Database, um Media) error {
	col, ctx := GetColCtx(mediaColName, db, 30)
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
