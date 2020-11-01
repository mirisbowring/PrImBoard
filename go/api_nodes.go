package primboard

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddNode handles the webrequest for Node creation
func (a *App) AddNode(w http.ResponseWriter, r *http.Request) {
	var e Node
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
	if err := e.VerifyNode(a.DB); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// try to insert model into db
	result, err := e.AddNode(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteNodeByID handles the webrequest for Node deletion
func (a *App) DeleteNodeByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := Node{ID: id}
	if err := e.GetNode(a.DB, getPermission(w)); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not verify node id")
		return
	}
	// verify that current user is the owner
	if e.Creator != getUsernameFromHeader(w) {
		_http.RespondWithError(w, http.StatusForbidden, "you are not allowed to delete this node from the system")
		return
	}
	// try to delete model
	result, err := e.DeleteNode(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}

// GetNodes handles the webrequest for receiving all nodes
func (a *App) GetNodes(w http.ResponseWriter, r *http.Request) {
	es, err := GetAllNodes(a.DB, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, es)
}

// GetNodeByID handles the webrequest for receiving Node model by id
func (a *App) GetNodeByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := Node{ID: id}
	// try to select user
	if err := e.GetNode(a.DB, getPermission(w)); err != nil {
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
func (a *App) UpdateNodeByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// store new model in tmp object
	var ue Node
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ue); err != nil {
		// error occured during encoding
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// verify the correctness of the update
	if err := ue.VerifyNode(a.DB); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// trying to update model with requested body
	e := Node{ID: id}
	_, err := e.UpdateNode(a.DB, ue, getPermission(w))
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err = e.GetNode(a.DB, getPermission(w)); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, e)
}
