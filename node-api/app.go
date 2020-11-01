package node

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mirisbowring/PrImBoard/helper/models"
	m "github.com/mirisbowring/PrImBoard/helper/models"
)

// App struct to maintain router
type App struct {
	Router *mux.Router
	Config *m.NodeConfig
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
// - router initialization
func (a *App) Initialize(config models.NodeConfig) {
	log.Info("Starting Initialization")
	a.Config = &config
	a.InitializeRoutes()
	a.authenticateToGateway()
}

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (a *App) Authenticate(h http.Handler, logout bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// authenticateToGateway authenticates the node against the central gateway
func (a *App) authenticateToGateway() {
	log.Info("authenticating to gateway")
	client := &http.Client{}
	// encode setting to json
	api := fmt.Sprintf("%s/api/v2/infrastructure/node/%s/authenticate", a.Config.GatewayURL, a.Config.NodeAuth.ID)
	req, _ := http.NewRequest("POST", api, strings.NewReader(a.Config.NodeAuth.Secret))
	req.Header.Set("Content-Type", "application/json")
	// send request
	res, _ := client.Do(req)
	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status-code": res.StatusCode,
		}).Fatal("could not authenticate to gateway")
	}
	log.Info("authentication to gateway successfull")
}
