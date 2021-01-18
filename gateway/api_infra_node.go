package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/Nerzal/gocloak/v7"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"

	"github.com/mirisbowring/primboard/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthenticateNode selects the specified node from db and verifies the psk
func (g *AppGateway) authenticateNode(w http.ResponseWriter, r *http.Request) {
	id := w.Header().Get("clientID")
	if id == "" {
		log.Error("clientID not specified")
		_http.RespondWithError(w, http.StatusBadRequest, "clientID not specified")
		return
	}

	// parse ID to ObjectID
	nodeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.WithFields(log.Fields{
			"clientID": id,
			"error":    err.Error(),
		}).Error("could not parse ObjectID from clientID")
		_http.RespondWithError(w, http.StatusBadRequest, "could not parse ObjectID from clientID")
		return
	}

	var node = models.Node{ID: nodeID}

	log.WithFields(log.Fields{
		"node":          node.ID,
		"authenticated": false,
	}).Info("node tries to authenticate")

	// select node from database
	if err := node.GetNode(g.DB, bson.M{"_id": node.ID}, models.NodeProjectInternal); err != nil {
		log.WithFields(log.Fields{
			"node":          node.ID,
			"authenticated": false,
			"error":         err.Error(),
		}).Error("could not select node from database")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not select node")
		return
	}

	// check if node was found -> secret / id map matches entry
	if node.ID.IsZero() {
		log.WithFields(log.Fields{
			"node":          node.ID,
			"authenticated": false,
		}).Warn("node failed to authenticate")
		_http.RespondWithError(w, http.StatusUnauthorized, "could not authenticate node")
		return
	}

	// append node if not in list already
	if n := g.Nodes[node.ID]; n == nil {
		g.Nodes[node.ID] = &node
	}

	log.WithFields(log.Fields{
		"node":          node.ID,
		"authenticated": true,
	}).Info("node authenticated to api")

	// go g.syncUserAuthentication(node)

	// respond with ok
	_http.RespondWithJSON(w, http.StatusOK, "")
}

// createNode creates a new Node object in the database for the current user
func (g *AppGateway) registerNode(w http.ResponseWriter, r *http.Request) {
	node := models.Node{
		Title:   "New Node",
		Creator: _http.GetUsernameFromHeader(w),
	}

	// try to create node and receive id
	if node.ID = node.AddNode(g.DB); node.ID.IsZero() {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create new node in database")
		return
	}

	ctx, cancel := context.WithCancel(g.Ctx)
	defer cancel()
	//prepare new keycloak client for node
	client := gocloak.Client{
		ClientID:                  gocloak.StringP(node.ID.Hex()),
		ClientAuthenticatorType:   gocloak.StringP("client-secret"),
		Enabled:                   gocloak.BoolP(true),
		FrontChannelLogout:        gocloak.BoolP(false),
		Protocol:                  gocloak.StringP("openid-connect"),
		StandardFlowEnabled:       gocloak.BoolP(true),
		ImplicitFlowEnabled:       gocloak.BoolP(false),
		DirectAccessGrantsEnabled: gocloak.BoolP(true),
		ServiceAccountsEnabled:    gocloak.BoolP(true),
		PublicClient:              gocloak.BoolP(false),
		BearerOnly:                gocloak.BoolP(false),
		Attributes: &(map[string]string{
			"backchannel.logout.session.required":      "false",
			"backchannel.logout.revoke.offline.tokens": "false",
			"access.token.signed.response.alg":         "RS256",
			"id.token.signed.response.alg":             "RS256",
		}),
	}

	// refresh keycloaktoken
	g.keycloakRefreshToken()

	// create client on keycloak
	id, err := g.KeycloakClient.CreateClient(ctx, g.KeycloakToken.AccessToken, g.Config.Keycloak.Realm, client)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not create client on keycloak")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create node on server")
		return
	}

	// refresh token to display
	secret, err := g.KeycloakClient.RegenerateClientSecret(ctx, g.KeycloakToken.AccessToken, g.Config.Keycloak.Realm, id)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not create a secret for client")
	}

	// assign secret to node
	node.Secret = *secret.Value
	if status := node.Replace(g.DB); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not save the secret for the node")
		return
	}

	// retrieve the node without secret
	node.GetNode(g.DB, g.GetUserPermissionW(w, false), models.NodeProject)

	_http.RespondWithJSON(w, http.StatusCreated, node)
}

// refreshNodeSecret generates a new keycloak secret for the node
func (g *AppGateway) refreshNodeSecret(w http.ResponseWriter, r *http.Request) {
	// parse id from path
	id := _http.ParsePathID(w, r, "id")
	if id.IsZero() {
		return
	}

	// parse return query
	ret, status := _http.ParseQueryBool(w, r, "return", true)
	if status > 0 {
		return
	}

	// select node secret from database
	node := models.Node{ID: id}
	if err := node.GetNode(g.DB, g.GetUserPermissionW(w, true), models.NodeProjectInternal); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ctx, cancel := context.WithCancel(g.Ctx)
	defer cancel()
	// refresh token to display
	log.Warn(node.KeycloakID)
	secret, err := g.KeycloakClient.RegenerateClientSecret(ctx, g.KeycloakToken.AccessToken, g.Config.Keycloak.Realm, node.KeycloakID)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not create a secret for client")
	}

	// assign new secret to node
	if status := node.UpdateNodeSecret(g.DB, g.GetUserPermissionW(w, true), *secret.Value); status > 1 {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not write new secret to db")
		return
	}

	if ret {
		_http.RespondWithJSON(w, http.StatusOK, *secret.Value)
	} else {
		_http.RespondWithJSON(w, http.StatusOK, "refreshed secret")
	}

	return

}

// retrieveNodeSecret selects a simple node representation with id and secret
func (g *AppGateway) retrieveNodeSecret(w http.ResponseWriter, r *http.Request) {
	// parse id from path
	id := _http.ParsePathID(w, r, "id")
	if id.IsZero() {
		return
	}

	// select node secret from database
	node := models.Node{ID: id}
	if err := node.GetNode(g.DB, g.GetUserPermissionW(w, true), models.NodeProjectSecret); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// OK
	_http.RespondWithJSON(w, http.StatusOK, node)
}

// parseNodeStructure returns information about a file and its shares for the node to build the access tree
func (g *AppGateway) parseNodeStructure(w http.ResponseWriter, r *http.Request) {
	_http.RespondWithJSON(w, http.StatusNotImplemented, "currently not supported")
}

func (g *AppGateway) syncUserAuthentication(node models.Node) {
	// wait for api endpoint to finish
	time.Sleep(3 * time.Second)

	//
	users, status := node.GetUser(g.DB)
	if status > 0 {
		log.Error("could not select authorized users for node")
		return
	}

	for _, user := range users {
		session := g.GetSessionByUser(user)
		// check if session does exist
		if session.Token != "" {
			continue
		}
		status, msg := handler.NodeAuthentication(session, []models.Node{node}, true, g.HTTPClient)
		if status > 0 {
			log.WithFields(log.Fields{
				"username": session.User,
				"node":     node.ID,
				"error":    msg,
			}).Error("could not authenticate user")
			continue
		}
	}
}
