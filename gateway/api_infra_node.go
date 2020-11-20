package gateway

import (
	"io/ioutil"
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// AuthenticateNode selects the specified node from db and verifies the psk
func (g *AppGateway) AuthenticateNode(w http.ResponseWriter, r *http.Request) {
	var node = models.Node{ID: _http.ParsePrimitiveID(w, r)}

	log.WithFields(log.Fields{
		"node":          node.ID,
		"authenticated": false,
	}).Info("node tries to authenticate")

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	psk := string(body)

	// select node from database
	if err := node.GetNode(g.DB, bson.M{"secret": psk}); err != nil {
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

	// add secret to node since it will not be selected from database
	node.Secret = psk

	// append node if not in list already
	if n := g.GetNode(node.ID); n == nil {
		g.Nodes = append(g.Nodes, node)
	}

	log.WithFields(log.Fields{
		"node":          node.ID,
		"authenticated": true,
	}).Info("node authenticated to api")

	// respond with ok
	_http.RespondWithJSON(w, http.StatusOK, "")
}
