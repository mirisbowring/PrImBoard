package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mirisbowring/primboard/helper/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

// Media holds all information about the item
type Media struct {
	ID              primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Sha1            string               `json:"sha1,omitempty" bson:"sha1,omitempty"`
	FileName        string               `json:"filename,omitempty" bson:"filename,omitempty"`
	FileNameThumb   string               `json:"filenameThumb,omitempty" bson:"filenameThumb,omitempty"`
	Title           string               `json:"title,omitempty" bson:"title,omitempty"`
	Description     string               `json:"description,omitempty" bson:"description,omitempty"`
	Comments        []*Comment           `json:"comments,omitempty" bson:"comments,omitempty"`
	Creator         string               `json:"creator,omitempty" bson:"creator,omitempty"`
	Events          []primitive.ObjectID `json:"events,omitempty" bson:"events,omitempty"`
	GroupIDs        []primitive.ObjectID `json:"groupIDs,omitempty" bson:"groupIDs,omitempty"`
	Timestamp       int64                `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	TimestampUpload int64                `json:"timestampUpload,omitempty" bson:"timestampUpload,omitempty"`
	URL             string               `json:"url,omitempty" bson:"url,omitempty"`
	URLThumb        string               `json:"urlThumb,omitempty" bson:"urlThumb,omitempty"`
	Type            string               `json:"type,omitempty" bson:"type,omitempty"`
	Extension       string               `json:"extension,omitempty" bson:"extension,omitempty"`
	ContentType     string               `json:"contentType,omitempty" bson:"contentType,omitempty"`
	Tags            []string             `json:"tags,omitempty" bson:"tags,omitempty"`
	NodeIDs         []primitive.ObjectID `json:"nodeIDs,omitempty" bson:"nodeIDs,omitempty"`
	Users           []User               `json:"users,omitempty"`
	Groups          []UserGroup          `json:"groups,omitempty"`
	Nodes           []Node               `json:"nodes,omitempty"`
}

// MediaEventMap is used to map an array of events to an array of media
type MediaEventMap struct {
	Events   []Event  `json:"events,omitempty"`
	MediaIDs []string `json:"mediaIDs,omitempty"`
}

// MediaGroupMap is used to map an array of groups to an array of media
type MediaGroupMap struct {
	Groups   []UserGroup `json:"groups,omitempty"`
	MediaIDs []string    `json:"mediaIDs,omitempty"`
}

//MediaProject is a bson representation of the $project aggregation for mongodb
var MediaProject = bson.M{
	"_id":             1,
	"sha1":            1,
	"filename":        1,
	"filenameThumb":   1,
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
	"extension":       1,
	"contentType":     1,
	"tags":            1,
	"users":           UserProject,
	"groups":          UserGroupProject,
	"nodes":           NodeProject,
}

// MediaListProject is a bson representaion of the $project aggregation for mongodb
var MediaListProject = bson.M{
	"_id":           1,
	"sha1":          1,
	"filename":      1,
	"filenameThumb": 1,
	"title":         1,
	"creator":       1,
	"url":           1,
	"urlThumb":      1,
	"type":          1,
	"extension":     1,
	"contentType":   1,
	"nodes":         NodeProject,
	"groups":        UserGroupProject,
}

// MediaCollection is the name of the mongo collection
var MediaCollection = "media"

// AddMedia creates the model in the mongodb
func (m *Media) AddMedia(db *mongo.Database) (*mongo.InsertOneResult, error) {
	conn := database.GetColCtx(MediaCollection, db, 30)
	result, err := conn.Col.InsertOne(conn.Ctx, m)
	defer conn.Cancel()
	return result, err
}

// AddTag adds a tag to the mapped tag set (ignores duplicates)
// Overrides the current model instance
func (m *Media) AddTag(db *mongo.Database, t string) error {
	conn := database.GetColCtx(MediaCollection, db, 30)
	filter := bson.M{"_id": m.ID}
	// specify the tag array to be handled as set
	update := bson.M{"$addToSet": bson.M{"tags": t}}
	// options to return the update document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// Execute query
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&m)
	defer conn.Cancel()
	if err != nil {
		return err
	}
	return nil
}

// AddTags adds a tag array to the mapped tag set (ignores duplicates)
// Overrides the current model instance
func (m *Media) AddTags(db *mongo.Database, tags []string) error {
	conn := database.GetColCtx(MediaCollection, db, 30)
	filter := bson.M{"_id": m.ID}
	// specify the tag array to be handled as set
	update := bson.M{"$addToSet": bson.M{"tags": bson.M{"$each": tags}}}
	// options to return the update document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// Execute query
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&m)
	defer conn.Cancel()
	if err != nil {
		return err
	}
	return nil
}

// AddUserGroups adds an array of primitive.ObjectID (of a usergroup) to the
// mapped usergroup set (ignores duplicates) Overrides the current model
// instance
func (m *Media) AddUserGroups(db *mongo.Database, ugIDs []primitive.ObjectID) error {
	conn := database.GetColCtx(MediaCollection, db, 30)
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
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&m)
	defer conn.Cancel()
	if err != nil {
		return err
	}
	return nil
}

// BulkAddTagMedia bulk operates an add Tags to  many media ids
func BulkAddTagMedia(db *mongo.Database, tags []string, ids []primitive.ObjectID, permission bson.M) (*mongo.BulkWriteResult, error) {
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}

	conn := database.GetColCtx(MediaCollection, db, 30)
	opts := options.BulkWrite().SetOrdered(false)
	// create update list
	models := []mongo.WriteModel{}
	for _, id := range ids {
		filter := bson.M{"$and": []bson.M{
			{"_id": id},
			permission}}
		update := bson.M{"$addToSet": bson.M{"tags": bson.M{"$each": tags}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}
	// execute bulk update
	res, err := conn.Col.BulkWrite(conn.Ctx, models, opts)
	defer conn.Cancel()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return res, nil
}

// BulkAddMediaEvent bulk operates an add events to many media ids
func BulkAddMediaEvent(db *mongo.Database, mediaIDs []primitive.ObjectID, eventIDs []primitive.ObjectID, permission bson.M) (*mongo.BulkWriteResult, error) {
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}

	conn := database.GetColCtx(MediaCollection, db, 30)
	opts := options.BulkWrite().SetOrdered(false)
	// create update list
	models := []mongo.WriteModel{}
	for _, id := range mediaIDs {
		filter := bson.M{"$and": []bson.M{
			{"_id": id},
			permission}}
		update := bson.M{"$addToSet": bson.M{"events": bson.M{"$each": eventIDs}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}
	// execute bulk update
	res, err := conn.Col.BulkWrite(conn.Ctx, models, opts)
	defer conn.Cancel()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return res, nil
}

// BulkAddMediaGroup bulk operates an adds groups to many media ids
func BulkAddMediaGroup(db *mongo.Database, mediaIDs []primitive.ObjectID, groupIDs []primitive.ObjectID, permission bson.M) (*mongo.BulkWriteResult, error) {
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}

	conn := database.GetColCtx(MediaCollection, db, 30)
	opts := options.BulkWrite().SetOrdered(false)
	// create update list
	models := []mongo.WriteModel{}
	for _, id := range mediaIDs {
		filter := bson.M{"$and": []bson.M{
			{"_id": id},
			permission}}
		update := bson.M{"$addToSet": bson.M{"groupIDs": bson.M{"$each": groupIDs}}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}
	// execute bulk update
	res, err := conn.Col.BulkWrite(conn.Ctx, models, opts)
	defer conn.Cancel()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return res, nil
}

// BulkDeleteMedia deletes multiple media by ids
//
// 0 -> ok
// 1 -> no permission was specified
// 2 -> no IDs specified to delete
// 3 -> could not execute BulkDeleteMedia
func BulkDeleteMedia(db *mongo.Database, mediaIDs []primitive.ObjectID, permission bson.M) (int, string) {
	if permission == nil {
		return 1, "no permission was specified"
	}

	if mediaIDs == nil || len(mediaIDs) == 0 {
		return 2, "no IDs specified to delete"
	}

	filters := []bson.M{}
	filters = append(filters, bson.M{"_id": bson.M{"$in": mediaIDs}})
	filters = append(filters, permission)

	conn := database.GetColCtx(MediaCollection, db, 30)
	defer conn.Cancel()

	res, err := conn.Col.DeleteMany(conn.Ctx, bson.M{"$and": filters})
	if err != nil {
		log.Error("could not execute BulkDeleteMedia")
		return 3, "could not execute BulkDeleteMedia"
	}

	log.WithFields(log.Fields{"count": res.DeletedCount}).Debug("bulk deleted media")
	return 0, fmt.Sprintf("deleted %d documents", res.DeletedCount)
}

// BulkRemoveMediaGroup bulk operates an pulls groups from many media ids
func BulkRemoveMediaGroup(db *mongo.Database, mediaIDs []primitive.ObjectID, groupIDs []primitive.ObjectID, permission bson.M) (*mongo.BulkWriteResult, error) {
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}

	conn := database.GetColCtx(MediaCollection, db, 30)
	opts := options.BulkWrite().SetOrdered(false)
	// create update list
	models := []mongo.WriteModel{}
	for _, id := range mediaIDs {
		filter := bson.M{"$and": []bson.M{
			{"_id": id},
			permission}}
		update := bson.M{"$pullAll": bson.M{"groupIDs": groupIDs}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}
	// execute bulk update
	res, err := conn.Col.BulkWrite(conn.Ctx, models, opts)
	defer conn.Cancel()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return res, nil
}

// DeleteMedia deletes the model from the mongodb
func (m *Media) DeleteMedia(db *mongo.Database) (*mongo.DeleteResult, error) {
	conn := database.GetColCtx(MediaCollection, db, 30)
	filter := bson.M{"_id": m.ID}
	result, err := conn.Col.DeleteOne(conn.Ctx, filter)
	defer conn.Cancel()
	return result, err
}

// GetMediaPage returns the requested page after a specific id
func GetMediaPage(db *mongo.Database, query MediaQuery, permission bson.M, nodeMap map[primitive.ObjectID]string) ([]Media, error) {
	// verify that query combination is able to be filtered
	if err := query.IsValid(); err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}
	// parse the order of

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
		filters = append(filters, bson.M{"events": query.Event})
	}
	// check if filter have been specified
	if len(query.Filter) > 0 {
		// tags := parseTags(db, query.Filter)
		filter := buildTagFilter(query.Filter)
		filters = append(filters, filter)
	}
	filters = append(filters, permission)

	// create empty bson if no filter specified to prevent npe
	var tmp bson.M
	if len(filters) > 0 {
		tmp = bson.M{"$and": filters}
	} else {
		tmp = bson.M{}
	}

	// also select nodes
	pipeline := []bson.M{
		{"$sort": bson.M{"_id": query.ASC}},
		{"$match": tmp},
		{"$limit": query.Size},
		{"$lookup": bson.M{
			"from":         "usergroup",
			"localField":   "groupIDs",
			"foreignField": "_id",
			"as":           "groups",
		}},
		{"$lookup": bson.M{
			"from":         "node",
			"localField":   "nodeIDs",
			"foreignField": "_id",
			"as":           "nodes",
		}},
		{"$project": MediaListProject},
	}

	var media []Media
	opts := options.Aggregate()
	conn := database.GetColCtx(MediaCollection, db, 30)
	defer conn.Cancel()
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {
		log.Println(err)
		return media, err
	}

	defer cursor.Close(conn.Ctx)
	// iterate over the cursor and create array
	for cursor.Next(conn.Ctx) {
		var m Media
		cursor.Decode(&m)
		if m.Nodes != nil && len(m.Nodes) > 0 {
			for i, node := range m.Nodes {
				m.Nodes[i].UserSession = nodeMap[node.ID]
			}
		}
		media = append(media, m)
	}
	return media, nil

	// cursor.All(conn.Ctx, &media)
	// return media, nil
}

// GetAllMedia selects all Media from the mongodb
func GetAllMedia(db *mongo.Database) ([]Media, error) {
	conn := database.GetColCtx(MediaCollection, db, 30)
	cursor, err := conn.Col.Find(conn.Ctx, bson.M{}) //find all
	if err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer cursor.Close(conn.Ctx)
	// iterate over the cursor and create array
	var ms []Media
	for cursor.Next(conn.Ctx) {
		var m Media
		cursor.Decode(&m)
		ms = append(ms, m)
	}
	// report errors if occured
	if err = cursor.Err(); err != nil {
		defer conn.Cancel()
		return nil, err
	}
	defer conn.Cancel()
	return ms, nil
}

// GetMedia returns the specified entry from the mongodb
func (m *Media) GetMedia(db *mongo.Database, permission bson.M, nodeMap map[primitive.ObjectID]string) error {
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
			"from":         "usergroup",
			"localField":   "groupIDs",
			"foreignField": "_id",
			"as":           "groups",
		}},
		{"$lookup": bson.M{
			"from":         "node",
			"localField":   "nodeIDs",
			"foreignField": "_id",
			"as":           "nodes",
		}},
		{"$project": MediaProject},
	}
	// append permission
	// pipeline = append(pipeline, permission)
	// pipe := []bson.M{
	// 	bson.M{"$project"}
	// }
	opts := options.Aggregate()
	conn := database.GetColCtx(MediaCollection, db, 30)
	defer conn.Cancel()

	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {
		return err
	}
	var found = false
	for cursor.Next(conn.Ctx) {
		err := cursor.Decode(&m)
		if err != nil {
			return err
		}
		if m.Nodes != nil && len(m.Nodes) > 0 {
			for i, node := range m.Nodes {
				m.Nodes[i].UserSession = nodeMap[node.ID]
			}
		}
		found = true
		break
	}
	if !found {
		return errors.New("no results")
	}
	return nil
}

// GetMediaByIDs selects multiple Media Documents for the passed ids.
// verifies the reading permissions
func GetMediaByIDs(db *mongo.Database, ids []primitive.ObjectID, permission bson.M) ([]Media, error) {
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}
	filter := bson.M{"$and": []bson.M{
		{"_id": bson.M{"$in": ids}},
		permission}}

	pipeline := []bson.M{
		{"$match": filter},
		{"$lookup": bson.M{
			"from":         "node",
			"localField":   "nodeIDs",
			"foreignField": "_id",
			"as":           "nodes",
		}},
		{"$project": MediaProject},
	}

	opts := options.Aggregate()
	conn := database.GetColCtx(MediaCollection, db, 30)
	defer conn.Cancel()

	var media []Media
	cursor, err := conn.Col.Aggregate(conn.Ctx, pipeline, opts)
	if err != nil {
		log.Error(err.Error())
		return media, err
	}
	defer cursor.Close(conn.Ctx)

	cursor.All(conn.Ctx, &media)
	return media, nil
}

// Save writes changes, made to the instance itself, to the database and
// overrides the instance with the return value from the database
func (m *Media) Save(db *mongo.Database) error {
	conn := database.GetColCtx(MediaCollection, db, 30)
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
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&m)
	defer conn.Cancel()
	if err != nil {
		return err
	}
	return nil
}

// UpdateMedia updates the record with the passed one
// Does NOT call the checkComments Method
func (m *Media) UpdateMedia(db *mongo.Database, um Media) error {
	conn := database.GetColCtx(MediaCollection, db, 30)
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
	err := conn.Col.FindOneAndUpdate(conn.Ctx, filter, update, &options).Decode(&m)
	defer conn.Cancel()
	if err != nil {
		return err
	}
	return nil
}

// CheckComments verifies that only one new comment has been passed and assignes
// the passed username and the current unix timestamp to the new comment
// this method is NOT called on create or update
func (m *Media) CheckComments(user string) error {
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

// buildTagFilter parses the passed filter string and creates a bson that can be
// appended to a find.
func buildTagFilter(tagFilter string) bson.M {
	filters := []bson.M{}
	// trim whitespaces and get values onlyx
	tags := strings.Fields(strings.TrimSpace(tagFilter))
	// prevent exception
	if len(tags) == 0 {
		return bson.M{}
	}
	// iterate over passed tags and append regex pattern
	for _, tag := range tags {
		filters = append(filters, bson.M{"tags": bson.M{"$regex": tag, "$options": "i"}})
	}

	if len(filters) > 1 {
		// bundle in $and if more than one tag
		return bson.M{"$and": filters}
	} else if len(filters) == 1 {
		// return first entry if only one exists
		return filters[0]
	} else {
		// return empty if anything went wrong
		return bson.M{}
	}
}
