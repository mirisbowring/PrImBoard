package node

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/internal/models/infrastructure"
	log "github.com/sirupsen/logrus"
)

// AppNode struct to maintain router
type AppNode struct {
	Router             *mux.Router
	Config             *infrastructure.NodeConfig
	Ctx                context.Context
	Sessions           []*iModels.Session
	HTTPClient         *http.Client
	KeycloakClient     gocloak.GoCloak
	KeycloakToken      *gocloak.JWT
	KeycloakTokenCache map[string]*gocloak.RetrospecTokenResult
}

type pathType string

const (
	pathTypeGroup pathType = "group"
	pathTypeUser  pathType = "user"
)

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
	n.KeycloakTokenCache = make(map[string]*gocloak.RetrospecTokenResult)
	n.Config = &config
	n.Ctx = context.Background()
	httpClient, tlsConfig := _http.GenerateHTTPClient(n.Config.CaCert, n.Config.TLSInsecure)
	n.HTTPClient = httpClient
	n.KeycloakClient = handler.CreateKeycloakClient(tlsConfig, n.Config.Keycloak.URL)
	n.authenticateToKeycloak(0, 10)

	n.initializeRoutes()

	// remove obsolete locations (could exist if crashed)
	// handler.DeleteFiles("/etc/nginx/locations")
	//
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

func (n *AppNode) methodNotAllowedHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn(r.RequestURI)
		fmt.Fprintf(w, "Method not allowed")
	})
}

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (n *AppNode) authenticate(h http.Handler, useCookie bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn("accessd")
		start := time.Now()
		if cookieAuth, status := _http.ParseQueryBool(w, r, "cookieAuth", true); status == 0 && cookieAuth {
			if token := _http.ReadCookie(r, "keycloak-jwt"); token != "" {
				if n.keycloakTokenActive(token) {
					jwt, claims, err := n.KeycloakClient.DecodeAccessToken(context.Background(), token, n.Config.Keycloak.Realm, "")
					if err == nil && jwt.Valid {
						claim := *claims
						username := claim["preferred_username"].(string)
						w.Header().Set("user", username)
						h.ServeHTTP(w, r)
					} else {
						_http.RespondWithError(w, http.StatusUnauthorized, "Your cookie is invalid")
					}
				}
			} else {
				_http.RespondWithError(w, http.StatusUnauthorized, "Your cookie is invalid")
			}
		} else {
			bearer := r.Header.Get("Authorization")
			bearer = strings.Replace(bearer, "Bearer ", "", 1)
			if n.keycloakTokenActive(bearer) {
				jwt, claims, err := n.KeycloakClient.DecodeAccessToken(context.Background(), bearer, n.Config.Keycloak.Realm, "")
				if err == nil && jwt.Valid {
					claim := *claims
					username := claim["preferred_username"].(string)
					w.Header().Set("user", username)
					h.ServeHTTP(w, r)
				} else {
					_http.RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
				}
			} else {
				_http.RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
			}
		}

		log.WithFields(log.Fields{
			"method":   r.Method,
			"uri":      r.RequestURI,
			"source":   r.RemoteAddr,
			"duration": time.Since(start),
		}).Info("handle request")
	})
}

// authenticateToGateway authenticates the node against the central gateway
//
// 0 -> ok
// 1 -> could not marshal id to json
// 2 -> could not send request
// 3 -> unauthorized (token invalid?)
// 4 -> unexpected status code
func (n *AppNode) authenticateToGateway() int {
	log.Info("authenticating to gateway")

	api := fmt.Sprintf("%s/api/v2/infrastructure/node/authenticate", n.Config.GatewayURL)

	resp, status, _ := _http.SendRequest(n.HTTPClient, http.MethodPost, api, n.KeycloakToken.AccessToken, nil, "application/json")
	if status > 0 {
		return 2
	}

	defer resp.Body.Close()

	logFields := log.Fields{
		"endpoint":    api,
		"status-code": resp.StatusCode,
	}

	switch resp.StatusCode {
	case http.StatusOK:
		log.WithFields(logFields).Info("authentication to gateway successful")
		return 0
	case http.StatusUnauthorized:
		log.WithFields(logFields).Error("could not authenticate to server")
		return 3
	default:
		log.WithFields(logFields).Error("unexpected status code")
		return 4
	}
}

// getDataPath concats the basepath with the type and the identifier. Ends with
// '/'
func (n *AppNode) getDataPath(identifier string, t pathType, thumb bool) string {
	switch t {
	case pathTypeUser:
		if thumb {
			return fmt.Sprintf("%s/%s/%s/own/thumb/", n.Config.BasePath, t, identifier)
		}
		return fmt.Sprintf("%s/%s/%s/own/", n.Config.BasePath, t, identifier)
	case pathTypeGroup:
		if thumb {
			return fmt.Sprintf("%s/%s/%s/thumb/", n.Config.BasePath, t, identifier)
		}
		return fmt.Sprintf("%s/%s/%s/", n.Config.BasePath, t, identifier)
	default:
		log.WithFields(log.Fields{
			"type": t,
		}).Error("unknown pathType specified")
		return ""
	}
}

// logs the client into the keycloak api and retrieves token
func (n *AppNode) authenticateToKeycloak(try int, max int) {
	ctx, cancel := context.WithCancel(n.Ctx)
	defer cancel()
	var err error
	n.KeycloakToken, err = n.KeycloakClient.LoginClient(ctx, n.Config.Keycloak.ClientID, n.Config.Keycloak.Secret, n.Config.Keycloak.Realm)
	if err != nil {
		log.WithFields(log.Fields{
			"clientid": n.Config.Keycloak.ClientID,
			"realm":    n.Config.Keycloak.Realm,
			"error":    err.Error(),
		}).Error("could not authenticate to keycloak api")
		// retry (possibly, keycloak not up)
		if try < max {
			time.Sleep(time.Second * 5)
			n.authenticateToKeycloak(try+1, max)
		}
	}
	n.keycloakTokenActive(n.KeycloakToken.AccessToken)
}

func (n *AppNode) keycloakTokenActive(token string) bool {
	// check if cached tokens expiry date is in future
	if val, ok := n.KeycloakTokenCache[token]; ok && int64(*val.Exp) >= time.Now().Unix() {
		log.Debug("found active token in token cache")
		return true
	} else if ok && int64(*val.Exp) < time.Now().Unix() {
		// cached token expired
		delete(n.KeycloakTokenCache, token)
	}
	// verify token against keycloak
	ctx, cancel := context.WithCancel(n.Ctx)
	defer cancel()
	rptResult, err := n.KeycloakClient.RetrospectToken(ctx, token, n.Config.Keycloak.ClientID, n.Config.Keycloak.Secret, n.Config.Keycloak.Realm)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not retrospect token")
		return false
	}

	// check if token active
	if !*rptResult.Active {
		log.Debug("token is not active according to keycloak api")
		return false
	}

	// cache token
	n.KeycloakTokenCache[token] = rptResult
	return *rptResult.Active
}

func (n *AppNode) keycloakRefreshToken() {
	// no need to refresh if vaid
	if n.keycloakTokenActive(n.KeycloakToken.AccessToken) {
		log.Debug("token still valid - skipping refresh")
		return
	}

	// verify token against keycloak
	var err error
	ctx, cancel := context.WithCancel(n.Ctx)
	defer cancel()
	n.KeycloakToken, err = n.KeycloakClient.RefreshToken(ctx, n.KeycloakToken.RefreshToken, n.Config.Keycloak.ClientID, n.Config.Keycloak.Secret, n.Config.Keycloak.Realm)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not refresh token")
		// possibly, the refresh token is expired -> try to reauthenticate
		n.authenticateToKeycloak(0, 1)
		return
	}

	// map token to cache
	n.keycloakTokenActive(n.KeycloakToken.AccessToken)
}
