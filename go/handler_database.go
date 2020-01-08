package swagger

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// DBName is the database namespace
var DBName = "primboard"

// Connect initializes a mongodb connection
func (a *App) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	a.DB = client.Database(DBName)
}

// helpers

var cancel context.CancelFunc

// DBContext creates a context of specified duration
func DBContext(i time.Duration) context.Context {
	var ctx context.Context
	ctx, cancel = context.WithTimeout(context.Background(), i*time.Second)
	return ctx
}

// CloseContext defers the cancel function
func CloseContext() {
	defer cancel()
}
