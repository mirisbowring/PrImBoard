package node

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mirisbowring/PrImBoard/helper"
)

// App struct to maintain router
type App struct {
	Router *mux.Router
	Config *Config
}

// Config struct that stores every api related settings
type Config struct {
	BasePath string `json:"basePath"`
}

// Run starts the application on the passed address with the inherited router
// WARN: router must be initialized first
func (a *App) Run(addr string) {

}

// Initialize initializes application related content
// - router initialization
func (a *App) Initialize(config string) {
	log.Info("Starting Initialization")
	a.readConfig(config)
	a.InitializeRoutes()
}

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (a *App) Authenticate(h http.Handler, logout bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// ReadConfig reads all neccessary settings from config file
func (a *App) readConfig(config string) {
	if err := helper.ReadJSONConfig(config).Decode(&a.Config); err != nil {
		log.Println("could not parse config file: " + config)
		log.Fatal(err)
	}
}
