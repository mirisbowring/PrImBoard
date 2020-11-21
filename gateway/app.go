package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mirisbowring/primboard/helper/database"
	_http "github.com/mirisbowring/primboard/helper/http"
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/internal/models/infrastructure"
	"github.com/mirisbowring/primboard/models"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// AppGateway struct to maintain database connection and router
type AppGateway struct {
	Router     *mux.Router
	DB         *mongo.Database
	Config     *infrastructure.APIGatewayConfig
	Nodes      []models.Node // stores all authenticated nodes
	Sessions   []*iModels.Session
	HTTPClient *http.Client
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
	g.Config = &config
	// load ca cert if specified
	g.HTTPClient = generateHTTPClient(g.Config.CaCert, g.Config.TLSInsecure)
	g.Connect()
	g.initializeRoutes()
}

func generateHTTPClient(caCert string, insecure bool) *http.Client {
	if caCert != "" {
		if rootCAs, status := loadCaCert(caCert); status == 0 {
			config := &tls.Config{
				InsecureSkipVerify: insecure,
				RootCAs:            rootCAs,
			}
			tr := &http.Transport{TLSClientConfig: config}
			return &http.Client{Transport: tr}
		}
	}
	if insecure == true {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		tr := &http.Transport{TLSClientConfig: config}
		return &http.Client{Transport: tr}
	}
	return &http.Client{}
}

func loadCaCert(certfile string) (*x509.CertPool, int) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	cert, err := ioutil.ReadFile(certfile)
	if err != nil {
		log.WithFields(log.Fields{
			"file": certfile,
		}).Error("could not read certificate")
		return &x509.CertPool{}, 1
	}

	if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
		log.WithFields(log.Fields{
			"file": certfile,
		}).Error("could not append cert to cert pool")
		return &x509.CertPool{}, 1
	}

	return rootCAs, 0
}

// Authenticate is a middleware to pre-authenticate routes via the session token
// if logout is true, no new session token is beeing generated
func (g *AppGateway) Authenticate(h http.Handler, logout bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		token := g.ReadSessionCookie(&w, r)
		s := g.GetSession(token)
		if s != nil && s.IsValid() {
			// set temporary user for internal processing
			// (will be deleted in response)
			w.Header().Set("user", s.User.Username)
			if !logout {
				if g.Config.SessionRotation {
					g.SetSessionCookie(&w, r, s)
				}
			}
			h.ServeHTTP(w, r)
		} else {
			g.CloseSession(&w, r)
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

// GetNode returns the session object for the passed token
func (g *AppGateway) GetNode(id primitive.ObjectID) *models.Node {
	if g.Nodes == nil || len(g.Nodes) == 0 {
		return nil
	}
	for _, n := range g.Nodes {
		if n.ID == id {
			return &n
		}
	}
	return nil
}

// GetUserPermission parses the permissionfilter and returns it
func (g *AppGateway) GetUserPermission(w http.ResponseWriter) bson.M {
	username := _http.GetUsernameFromHeader(w)
	session := g.GetSessionByUsername(username)
	return database.CreatePermissionFilter(session.Usergroups, username)
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
