package swagger

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// Connect initializes a mongodb connection
func (a *App) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(a.Config.MongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	a.DB = client.Database(a.Config.DBName)
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
