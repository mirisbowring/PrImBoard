package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DBConnection holds information about the collection and the context
type DBConnection struct {
	Col    *mongo.Collection
	Ctx    context.Context
	Cancel context.CancelFunc
}

// GetColCtx returns the collection for the specified model and initializes a
// timeout context with passed duration
func GetColCtx(model string, db *mongo.Database, duration time.Duration) DBConnection {
	var conn DBConnection
	// init the specified collection on the passed db instance
	conn.Col = db.Collection(model)
	conn.Ctx, conn.Cancel = context.WithTimeout(context.Background(), duration*time.Second)
	return conn
}

// CreatePermissionProjectPipeline creates a pipeline with permission for a specific user and model project
func CreatePermissionProjectPipeline(permission bson.M, id primitive.ObjectID, project bson.M) ([]primitive.M, error) {
	matcher, err := CreatePermissionMatcher(permission, id)
	if err != nil {
		return nil, err
	}
	return CreateMatcherProjectPipeline(matcher, project), nil
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

	// if len(filters) > 1 {
	return bson.M{"$or": filters}
	// }
	// return filters[0]
}

// CreateMatcherProjectPipeline creates a pipeline with a match and project
func CreateMatcherProjectPipeline(matcher bson.M, project bson.M) []primitive.M {
	// create pipeline
	pipeline := []bson.M{
		{"$match": matcher},
		{"$project": project},
	}
	return pipeline
}

// CreatePermissionMatcher creates a matcher that checks for permissions
func CreatePermissionMatcher(permission bson.M, id primitive.ObjectID) (bson.M, error) {
	// verify that permission bson was specified
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}
	// create matcher
	var matcher bson.M
	if id == primitive.NilObjectID {
		matcher = bson.M{"$and": []bson.M{
			permission,
		}}
	} else {
		matcher = bson.M{"$and": []bson.M{
			{"_id": id},
			permission,
		}}
	}
	return matcher, nil
}
