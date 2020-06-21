package primboard

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBConnection holds information about the collection and the context
type DBConnection struct {
	Col    *mongo.Collection
	Ctx    context.Context
	Cancel context.CancelFunc
}

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

// GetColCtx returns the collection for the specified model and initializes a
// timeout context with passed duration
func GetColCtx(model string, db *mongo.Database, duration time.Duration) DBConnection {
	var conn DBConnection
	// init the specified collection on the passed db instance
	conn.Col = db.Collection(model)
	conn.Ctx, conn.Cancel = context.WithTimeout(context.Background(), duration*time.Second)
	return conn
}
