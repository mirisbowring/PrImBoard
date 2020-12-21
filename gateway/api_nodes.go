package gateway

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (g *AppGateway) addGroupsToNode(w http.ResponseWriter, r *http.Request) {
	// parse query
	query, status := _http.ParseQueryString(w, r, "groups", false)
	if status > 0 {
		return
	}

	id, status := _http.ParsePathString(w, r, "id")
	if status > 0 {
		return
	}

	nodeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, "invalid node id specified")
		return
	}

	// parse ids from string slice
	groupIDs, err := ParseIDs(strings.Split(query, ","))
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, "could not decode groups query")
		return
	}

	// verify that any valid id was passed
	if len(groupIDs) == 0 {
		_http.RespondWithError(w, http.StatusBadRequest, "no valid group ids have been passed")
	}

	// select valid groups from database
	groups, err := models.GetUserGroupsByIDs(g.DB, groupIDs, g.GetUserPermissionW(w, false))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not select matching groups from database")
		return
	}

	// verify that there is any valid grup to share the node with
	if len(groups) == 0 {
		_http.RespondWithError(w, http.StatusForbidden, "you are not allowed to access one of the specified groups")
		return
	}

	node := models.Node{ID: nodeID}
	if err = node.AddUserGroups(g.DB, groupIDs, g.GetUserPermissionW(w, true)); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not add usergroups to node")
		return
	}

	_http.RespondWithJSON(w, http.StatusOK, node)

}

// AddNode handles the webrequest for Node creation
func (g *AppGateway) AddNode(w http.ResponseWriter, r *http.Request) {
	var e models.Node
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// settings creator
	e.Creator = w.Header().Get("user")
	// check mandatory fields
	if err := e.VerifyNode(g.DB, g.GetUserPermissionW(w, false)); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// try to insert model into db
	result, err := e.AddNode(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteNodeByID handles the webrequest for Node deletion
func (g *AppGateway) DeleteNodeByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := models.Node{ID: id}
	if err := e.GetNode(g.DB, g.GetUserPermissionW(w, true), false); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not verify node id")
		return
	}
	// verify that current user is the owner
	if e.Creator != _http.GetUsernameFromHeader(w) {
		_http.RespondWithError(w, http.StatusForbidden, "you are not allowed to delete this node from the system")
		return
	}
	// try to delete model
	result, err := e.DeleteNode(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}

// GetNodes handles the webrequest for receiving all nodes
func (g *AppGateway) GetNodes(w http.ResponseWriter, r *http.Request) {
	es, err := models.GetAllNodes(g.DB, g.GetUserPermissionW(w, false), "")
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, es)
}

// GetNodeByID handles the webrequest for receiving Node model by id
func (g *AppGateway) GetNodeByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := models.Node{ID: id}
	// try to select user
	if err := e.GetNode(g.DB, g.GetUserPermissionW(w, false), false); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "Node not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, e)
}

// UpdateNodeByID handles the webrequest for updating the Node with the passed request body
func (g *AppGateway) UpdateNodeByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// store new model in tmp object
	var ue models.Node
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ue); err != nil {
		// error occured during encoding
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// verify the correctness of the update
	if err := ue.VerifyNode(g.DB, g.GetUserPermissionW(w, false)); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// trying to update model with requested body
	e := models.Node{ID: id}
	_, err := e.UpdateNode(g.DB, ue, g.GetUserPermissionW(w, true))
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err = e.GetNode(g.DB, g.GetUserPermissionW(w, false), false); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, e)
}
