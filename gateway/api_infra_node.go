package gateway

import (
	"net/http"
	"time"

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
	if err := node.GetNode(g.DB, bson.M{"_id": node.ID}, true); err != nil {
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
				"username": session.User.Username,
				"node":     node.ID,
				"error":    msg,
			}).Error("could not authenticate user")
			continue
		}
	}
}
