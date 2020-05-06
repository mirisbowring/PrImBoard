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
	Domain               string   `json:"domain"`
	Port                 int      `json:"port"`
	MongoURL             string   `json:"mongo_url"`
	DBName               string   `json:"database_name"`
	CookiePath           string   `json:"cookie_path"`
	CookieHTTPOnly       bool     `json:"cookie_http_only"`
	CookieSecure         bool     `json:"cookie_secure"`
	CookieTokenTitle     string   `json:"cookie_token_title"`
	AllowedOrigins       []string `json:"allowed_origins"`
	TagPreviewLimit      int64    `json:"tag_preview_limit"`
	SessionRotation      bool     `json:"session_rotation"`
	DefaultMediaPageSize int      `json:"default_media_page_size"`
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

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (a *App) Authenticate(h http.Handler, logout bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ReadSessionCookie(&w, r)
		s := GetSession(token)
		if s != nil && s.IsValid() {
			if !logout {
				if a.Config.SessionRotation {
					SetSessionCookie(&w, r, s)
				}
				// set temporary user for internal processing
				// (will be deleted in response)
				w.Header().Set("user", s.User.Username)
			}
			h.ServeHTTP(w, r)
		} else {
			CloseSession(&w, r)
			RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
			return
		}
	})
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
	// delete the temporary user key from header
	w.Header().Del("user")
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

// Find iterates over the slice and returns the position of the element if found
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// UniqueStrings removes all duplicates from a string slice and returns the result
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// RemoveString removes a given string from a given slice
func RemoveString(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
