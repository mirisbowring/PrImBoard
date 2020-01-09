package primboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var api *App
var config = "env.json"

// App struct to maintain database connection and router
type App struct {
	Router *mux.Router
	DB     *mongo.Database
	Config *Config
}

// Config struct that stores every api related settings
type Config struct {
	Domain           string   `json:"domain"`
	Port             int      `json:"port"`
	MongoURL         string   `json:"mongo_url"`
	DBName           string   `json:"database_name"`
	CookiePath       string   `json:"cookie_path"`
	CookieHTTPOnly   bool     `json:"cookie_http_only"`
	CookieSecure     bool     `json:"cookie_secure"`
	CookieTokenTitle string   `json:"cookie_token_title"`
	AllowedOrigins   []string `json:"allowed_origins"`
}

// Run starts the application on the passed address with the inherited router
// WARN: router must be initialized first
func (a *App) Run(addr string) {
	log.Fatal(
		http.ListenAndServe(
			addr,
			handlers.CORS(
				handlers.AllowedHeaders(
					[]string{
						"X-Requested-With",
						"Content-Type",
						"Authorization",
					},
				),
				handlers.AllowedMethods(
					[]string{
						"GET",
						"POST",
						"PUT",
						"HEAD",
						"OPTIONS",
					},
				),
				handlers.AllowedOrigins(
					a.Config.AllowedOrigins,
				),
				handlers.AllowCredentials(),
			)(a.Router)))
}

// Initialize initializes application related content
// - mongodb connection initialization
// - router initialization
func (a *App) Initialize() {
	a.ReadConfig()
	a.Connect()
	a.InitializeRoutes()
	api = a
}

// helpers

// HashPassword hashes the passed passwort using bcrypt
func HashPassword(password string) (hashedPassword string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

// MatchesBcrypt verifies, that a given string is equal to a encrypted string
func MatchesBcrypt(password string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// RespondWithError Creates an error payload and adds the error message to be
// returned
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON parses the passed payload and returns it with the specified
// code to the client
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	//	enableCors(&w)
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ReadConfig reads all neccessary settings from config file
func (a *App) ReadConfig() {
	f, err := os.Open(config)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	//decode file content into go object
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&a.Config)
	if err != nil {
		log.Fatal(err)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "localhost")
	// should ask if option (put delete post)
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// Returns the collection for the specified model on the passed db instance
func modelCollection(model string, db *mongo.Database) *mongo.Collection {
	return db.Collection(model)
}

// GetColCtx returns the collection for the specified model and initializes a
// timeout context with passed duration
func GetColCtx(model string, db *mongo.Database, duration time.Duration) (*mongo.Collection, context.Context) {
	return modelCollection(model, db), DBContext(duration)
}
