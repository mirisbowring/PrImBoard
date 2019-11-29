package swagger

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/gorilla/mux"
)

// application struct to maintain database connection and router
type App struct {
	Router	*mux.Router
	DB		*mongo.Database
}

/*
 * Starts the application on the passed address with the inherited router
 * WARN: router must be initialized first
 */
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

/*
 * Initializes application related content
 * - mongodb connection initialization
 * - router initialization
 */
func (a *App) Initialize() {
	a.Connect()
	a.InitializeRoutes()
}

// helpers
/*
 * Creates an error payload and adds the error message to be returned
 */
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

/*
 * parses the passed payload and returns it with the specified code to the client
 */
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

/*
 * Returns the collection for the specified model on the passed db instance
 */
func modelCollection(model string, db *mongo.Database) *mongo.Collection {
	return db.Collection(model)
}

/*
 * Returns the collection for the specified model and initializes a timeout context with passed duration
 */
func GetColCtx(model string, db *mongo.Database, duration time.Duration) (*mongo.Collection, context.Context) {
	return modelCollection(model, db), DBContext(duration)
}