package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mirisbowring/primboard/helper/database"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/internal/models/infrastructure"
	"github.com/mirisbowring/primboard/models"
	log "github.com/sirupsen/logrus"

	"github.com/Nerzal/gocloak/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// AppGateway struct to maintain database connection and router
type AppGateway struct {
	Router             *mux.Router
	DB                 *mongo.Database
	Config             *infrastructure.APIGatewayConfig
	Ctx                context.Context
	Nodes              map[primitive.ObjectID]*models.Node // stores all authenticated nodes
	Sessions           []*iModels.Session
	HTTPClient         *http.Client
	KeycloakClient     gocloak.GoCloak
	KeycloakToken      *gocloak.JWT
	KeycloakTokenCache map[string]*gocloak.RetrospecTokenResult
}

// Run starts the application on the passed address with the inherited router
// WARN: router must be initialized first
func (g *AppGateway) Run(addr string) {
	if g.Config.HTTP {
		log.Error(
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
						g.Config.AllowedOrigins,
					),
					handlers.AllowCredentials(),
				)(g.Router)))
	} else {
		log.Error(
			http.ListenAndServeTLS(
				addr,
				fmt.Sprintf("%s/server.crt", g.Config.Certificates),
				fmt.Sprintf("%s/server.key", g.Config.Certificates),
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
						g.Config.AllowedOrigins,
					),
					handlers.AllowCredentials(),
				)(g.Router)))
	}

}

// Connect initializes a mongodb connection
func (g *AppGateway) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(g.Config.MongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	g.DB = client.Database(g.Config.DBName)
}

// Initialize initializes application related content
// - mongodb connection initialization
// - router initialization
func (g *AppGateway) Initialize(config infrastructure.APIGatewayConfig) {
	log.Info("Starting Initialization")
	g.KeycloakTokenCache = make(map[string]*gocloak.RetrospecTokenResult)
	g.Nodes = make(map[primitive.ObjectID]*models.Node)
	g.Config = &config
	g.Ctx = context.Background()
	// load ca cert if specified
	httpClient, tlsConfig := _http.GenerateHTTPClient(g.Config.CaCert, g.Config.TLSInsecure)
	g.HTTPClient = httpClient
	g.KeycloakClient = handler.CreateKeycloakClient(tlsConfig, g.Config.Keycloak.URL)
	g.authenticateToKeycloak(0, 10)
	g.Connect()
	g.initializeRoutes()
}

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (g *AppGateway) Authenticate(h http.Handler, logout bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		bearer := r.Header.Get("Authorization")
		bearer = strings.Replace(bearer, "Bearer ", "", 1)
		if g.keycloakTokenActive(bearer) {
			ctx, cancel := context.WithCancel(g.Ctx)
			defer cancel()
			jwt, claims, err := g.KeycloakClient.DecodeAccessToken(ctx, bearer, g.Config.Keycloak.Realm, "")
			if err == nil && jwt.Valid {
				if val, ok := (*claims)["clientId"]; ok {
					w.Header().Set("clientID", val.(string))
				} else {
					username := ""
					if tmp, ok := (*claims)["preferred_username"]; ok {
						username = tmp.(string)
					}
					// generate session
					s := g.GetSession(bearer)
					if s == nil {
						s = g.prepareUsersession(username, bearer)
						g.Sessions = append(g.Sessions, s)
					}
					w.Header().Set("user", username)
				}
				h.ServeHTTP(w, r)
			} else {
				g.RemoveSessionByToken(bearer)
				_http.RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
				return
			}
		} else {
			g.RemoveSessionByToken(bearer)
			_http.RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
			return
		}
		log.WithFields(log.Fields{
			"method":   r.Method,
			"uri":      r.RequestURI,
			"source":   r.RemoteAddr,
			"duration": time.Since(start),
		}).Info("handle request")
	})
}

// logs the client into the keycloak api and retrieves token
func (g *AppGateway) authenticateToKeycloak(try int, max int) {
	ctx, cancel := context.WithCancel(g.Ctx)
	defer cancel()
	var err error
	g.KeycloakToken, err = g.KeycloakClient.LoginClient(ctx, g.Config.Keycloak.ClientID, g.Config.Keycloak.Secret, g.Config.Keycloak.Realm)
	if err != nil {
		log.WithFields(log.Fields{
			"clientid": g.Config.Keycloak.ClientID,
			"realm":    g.Config.Keycloak.Realm,
			"error":    err.Error(),
		}).Error("could not authenticate to keycloak api")
		// retry (possibly, keycloak not up)
		if try < max {
			time.Sleep(time.Second * 5)
			g.authenticateToKeycloak(try+1, max)
		}
	}
	g.keycloakTokenActive(g.KeycloakToken.AccessToken)
}

func (g *AppGateway) keycloakTokenActive(token string) bool {
	// check if cached tokens expiry date is in future
	if val, ok := g.KeycloakTokenCache[token]; ok && int64(*val.Exp) >= time.Now().Unix() {
		log.Debug("found active token in token cache")
		return true
	} else if ok && int64(*val.Exp) < time.Now().Unix() {
		// cached token expired
		delete(g.KeycloakTokenCache, token)
	}
	// verify token against keycloak
	ctx, cancel := context.WithCancel(g.Ctx)
	defer cancel()
	rptResult, err := g.KeycloakClient.RetrospectToken(ctx, token, g.Config.Keycloak.ClientID, g.Config.Keycloak.Secret, g.Config.Keycloak.Realm)
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
	g.KeycloakTokenCache[token] = rptResult
	return *rptResult.Active
}

func (g *AppGateway) keycloakRefreshToken() {
	// no need to refresh if vaid
	if g.keycloakTokenActive(g.KeycloakToken.AccessToken) {
		log.Debug("token still valid - skipping refresh")
		return
	}

	// verify token against keycloak
	var err error
	ctx, cancel := context.WithCancel(g.Ctx)
	defer cancel()
	g.KeycloakToken, err = g.KeycloakClient.RefreshToken(ctx, g.KeycloakToken.RefreshToken, g.Config.Keycloak.ClientID, g.Config.Keycloak.Secret, g.Config.Keycloak.Realm)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not refresh token")
		// possibly, the refresh token is expired -> try to reauthenticate
		g.authenticateToKeycloak(0, 1)
		return
	}

	// map token to cache
	g.keycloakTokenActive(g.KeycloakToken.AccessToken)
}

// prepareUsersession selects the user and prepares a local session
func (g *AppGateway) prepareUsersession(username string, token string) *iModels.Session {
	// verify username
	if username == "" {
		log.WithFields(log.Fields{
			"username": username,
		}).Error("username empty - cannot prepare user")
	}

	// // read corresponding user from db
	// var u models.User
	// u.Username = username
	// if err := u.GetUser(g.DB); err != nil {
	// 	switch err {
	// 	case mongo.ErrNoDocuments:
	// 		// model not found
	// 		log.WithFields(log.Fields{
	// 			"username": username,
	// 			"error":    "no user found for username",
	// 		}).Error("could not select user from database")
	// 		break
	// 	default:
	// 		log.WithFields(log.Fields{
	// 			"username": username,
	// 			"error":    err.Error(),
	// 		}).Error("could not select user from database")
	// 	}
	// 	return &iModels.Session{}
	// }

	// Create new Session
	session := g.NewSession(username, g.DB, token)

	// // select nodes, the user has access to
	// nodes, err := models.GetAllNodes(g.DB, g.GetUserPermission(u.Username, false), "auth")
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"username": u.Username,
	// 		"error":    err.Error(),
	// 	}).Error("could not retrieve nodes")
	// 	return &iModels.Session{}
	// }

	// // // authenticate user to nodes
	// if status, msg := handler.NodeAuthentication(session, nodes, true, g.HTTPClient); status > 0 {
	// 	log.WithFields(log.Fields{
	// 		"username": username,
	// 		"error":    msg,
	// 	}).Error("could not get nodes for user")
	// 	return &iModels.Session{}
	// }

	return session
}

// GetUserPermission parses the permissionfilter and returns it
func (g *AppGateway) GetUserPermission(username string, ownerOnly bool) bson.M {
	if ownerOnly {
		return database.CreatePermissionFilter(nil, username)
	}
	session := g.GetSessionByUsername(username)
	return database.CreatePermissionFilter(session.Usergroups, username)
}

// GetUserPermissionW parses the permissionfilter and returns it
func (g *AppGateway) GetUserPermissionW(w http.ResponseWriter, ownerOnly bool) bson.M {
	username := _http.GetUsernameFromHeader(w)
	return g.GetUserPermission(username, ownerOnly)
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

// UnParseIDs parses a slice of hex ids into primitive.ObjectIDs
func UnParseIDs(ids []primitive.ObjectID) []string {
	var IDs []string
	for _, i := range ids {
		id := i.Hex()
		IDs = append(IDs, id)
	}
	return IDs
}

// getNodeTokenMap parses the nodetoken map for the currentuser
//
// creates http error response if fails
func (g *AppGateway) getNodeTokenMap(w http.ResponseWriter) (map[primitive.ObjectID]string, int) {
	nodeMap := g.GetSessionByUsername(_http.GetUsernameFromHeader(w)).NodeTokenMap
	if nodeMap == nil || len(nodeMap) == 0 {
		_http.RespondWithError(w, http.StatusNotFound, "no node available")
		return nil, 1
	}
	return nodeMap, 0
}

// func (g *AppGateway) keycloakTokenActive(token string) bool {
// 	// check if cached tokens expiry date is in future
// 	if val, ok := g.KeycloakTokenCache[token]; ok && int64(*val.Exp) >= time.Now().Unix() {
// 		log.Warn("cached token")
// 		return true
// 	} else if ok && int64(*val.Exp) < time.Now().Unix() {
// 		// cached token expired
// 		delete(g.KeycloakTokenCache, token)
// 	}
// 	// verify token against keycloak
// 	ctx, cancel := context.WithCancel(g.Ctx)
// 	defer cancel()
// 	rptResult, err := g.KeycloakClient.RetrospectToken(ctx, token, g.Config.Keycloak.ClientID, g.Config.Keycloak.Secret, g.Config.Keycloak.Realm)
// 	if err != nil {
// 		log.WithFields(log.Fields{
// 			"error": err.Error(),
// 		}).Error("could not retrospect token")
// 		return false
// 	}

// 	// check if token active
// 	if !*rptResult.Active {
// 		return false
// 	}

// 	// cache token
// 	g.KeycloakTokenCache[token] = rptResult
// 	return *rptResult.Active
// }
