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


type App struct {
	Router	*mux.Router
	DB		*mongo.Database
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) Initialize() {
	a.Connect()
	a.InitializeRoutes()
}

// helpers
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	log.Print(code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func modelCollection(col string, db *mongo.Database) *mongo.Collection {
	return db.Collection(col)
}

func GetColCtx(col string, db *mongo.Database, duration time.Duration) (*mongo.Collection, context.Context) {
	return modelCollection(col, db), DBContext(duration)
}