package swagger

import (
	"context"
	"time"
	"log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// global db connection var
	client *mongo.Client
	dbClient *mongo.Database
	DBName = "primboard"
)

type Model interface {
	GetCollection(db *mongo.Database) *mongo.Collection	
}

func Connect(){
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	dbClient = client.Database(DBName)
}

func FindOne(model Model, filter *bson.M) error {
	col := model.GetCollection(dbClient)
	ctx := dbContext(5)

	err := col.FindOne(ctx, filter).Decode(&model)
	return err
}

func InsertOne(model Model) (*mongo.InsertOneResult, error) {
	col := model.GetCollection(dbClient)
	ctx := dbContext(5)

	result, err := col.InsertOne(ctx, model)
	return result, err
}

func Save(model Model) (*mongo.InsertOneResult, error) {
	col := model.GetCollection(dbClient)
	ctx := dbContext(5)

	result, err := col.InsertOne(ctx, model)
	return result, err
}

// helpers
func dbContext(i time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), i*time.Second)
	defer cancel()
	return ctx
}
