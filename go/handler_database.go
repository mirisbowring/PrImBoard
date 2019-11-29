package swagger

import (
	"context"
	"time"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// database namespace
var DBName = "primboard"

/*
 * Initializes a mongodb connection
 */
func (a *App) Connect(){
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
/*
 * Creates a context of specified duration
 */
func DBContext(i time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), i*time.Second)
	return ctx
}
