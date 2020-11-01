package primboard

import (
	"io/ioutil"
	"net/http"

	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"go.mongodb.org/mongo-driver/bson"
)

// AuthenticateNode selects the specified node from db and verifies the psk
func (a *App) AuthenticateNode(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	psk := string(body)

	// select node from database
	var node = Node{ID: _http.ParsePrimitiveID(w, r)}
	if err := node.GetNode(a.DB, bson.M{"secret": psk}); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not select node")
		return
	}

	// check if node was found -> secret / id map matches entry
	if node.ID.IsZero() {
		_http.RespondWithError(w, http.StatusUnauthorized, "could not authenticate node")
		return
	}

	// append node if not in list already
	if n := a.getNode(node.ID); n == nil {
		a.Nodes = append(a.Nodes, node)
	}

	// respond with ok
	_http.RespondWithJSON(w, http.StatusOK, "")
}
