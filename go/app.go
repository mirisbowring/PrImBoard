package primboard

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"github.com/mirisbowring/PrImBoard/helper/models"
	m "github.com/mirisbowring/PrImBoard/helper/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var api *App
var config = "env.json"

// App struct to maintain database connection and router
type App struct {
	Router *mux.Router
	DB     *mongo.Database
	Config *m.APIGatewayConfig
	Nodes  []Node // stores all authenticated nodes
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
						"DELETE",
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
func (a *App) Initialize(config models.APIGatewayConfig) {
	log.Info("Starting Initialization")
	a.Config = &config
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
			_http.RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
			return
		}
	})
}

// helpers

// GetSession returns the session object for the passed token
func (a *App) getNode(id primitive.ObjectID) *Node {
	for _, n := range a.Nodes {
		if n.ID == id {
			return &n
		}
	}
	return nil
}

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

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "localhost")
	// should ask if option (put delete post)
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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

// ParseIDs parses a slice of hex ids into primitive.ObjectIDs
func ParseIDs(ids []string) ([]primitive.ObjectID, error) {
	var IDs []primitive.ObjectID
	for _, i := range ids {
		id, err := primitive.ObjectIDFromHex(i)
		if err != nil {
			e := fmt.Sprintf("could not parse %s to primitive.ObjectID.", i)
			return nil, errors.New(e)
		}
		IDs = append(IDs, id)
	}
	return IDs, nil
}
