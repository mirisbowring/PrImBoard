package node

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/primboard/helper/http"
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/internal/models/infrastructure"
	log "github.com/sirupsen/logrus"
)

// AppNode struct to maintain router
type AppNode struct {
	Router   *mux.Router
	Config   *infrastructure.NodeConfig
	Sessions []*iModels.Session
}

// Run starts the application on the passed address with the inherited router
// WARN: router must be initialized first
func (n *AppNode) Run(addr string) {
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
					n.Config.AllowedOrigins,
				),
				handlers.AllowCredentials(),
			)(n.Router)))
}

// Initialize initializes application related content
// - router initialization
func (n *AppNode) Initialize(config infrastructure.NodeConfig) {
	log.Info("Starting Initialization")
	n.Config = &config
	n.initializeRoutes()
	resp := n.authenticateToGateway()
	// if authentication failed due to gateway down (status == 1), retry every
	// 10 seconds until gateway up
	for resp == 1 {
		log.WithFields(log.Fields{
			"gateway": n.Config.GatewayURL,
		}).Info("retry connection to gateway in 10 sec")
		// wait 10 seconds
		time.Sleep(10 * time.Second)
		resp = n.authenticateToGateway()
	}
}

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (n *AppNode) authenticate(h http.Handler, logout bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fmt.Sprintf("Bearer %s", n.Config.NodeAuth.Secret) != r.Header.Get("Authorization") {
			_http.RespondWithError(w, http.StatusUnauthorized, "authentication failed")
			return
		}
		h.ServeHTTP(w, r)
	})
}

// authenticateToGateway authenticates the node against the central gateway
// returns 0 if ok | 1 if request failed (perhaps gateway is not up)
func (n *AppNode) authenticateToGateway() int {
	log.Info("authenticating to gateway")
	client := &http.Client{}
	// encode setting to json
	api := fmt.Sprintf("%s/api/v2/infrastructure/node/%s/authenticate", n.Config.GatewayURL, n.Config.NodeAuth.ID)
	req, _ := http.NewRequest("POST", api, strings.NewReader(n.Config.NodeAuth.Secret))
	req.Header.Set("Content-Type", "application/json")
	// send request
	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoint": api,
			"error":    err.Error(),
		}).Error("authentication request failed")
		return 1
	}
	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"endpoint":    api,
			"status-code": res.StatusCode,
		}).Fatal("could not authenticate to gateway")
	}
	log.Info("authentication to gateway successfull")
	return 0
}
