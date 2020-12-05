package gateway

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddEvent handles the webrequest for Event creation
func (g *AppGateway) AddEvent(w http.ResponseWriter, r *http.Request) {
	var e models.Event
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// title is mandatory
	if e.Title == "" {
		_http.RespondWithError(w, http.StatusBadRequest, "Title cannot be empty")
		return
	}
	// settings creator
	e.Creator = w.Header().Get("user")
	// setting creation timestamp
	e.TimestampCreation = int64(time.Now().Unix())
	// try to insert model into db
	result, err := e.AddEvent(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteEventByID handles the webrequest for Event deletion
func (g *AppGateway) DeleteEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := models.Event{ID: id}
	// try to delete model
	result, err := e.DeleteEvent(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}

// GetEvents handles the webrequest for receiving all events
func (g *AppGateway) GetEvents(w http.ResponseWriter, r *http.Request) {
	es, err := models.GetAllEvents(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, es)
}

// GetEventByID handles the webrequest for receiving Event model by id
func (g *AppGateway) GetEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := models.Event{ID: id}
	// try to select user
	if err := e.GetEvent(g.DB, g.GetUserPermission(w, false)); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "Event not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, e)
}

// GetEventsByName returns available Events by their name, starting with
func (g *AppGateway) GetEventsByName(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	keyword := vars["title"]
	events, err := models.GetEventsByKeyword(g.DB, g.GetUserPermission(w, false), keyword, g.Config.TagPreviewLimit)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, events)
}

// MapTagsToEvents maps a slice of Tags to a slice of events
func (g *AppGateway) MapTagsToEvents(w http.ResponseWriter, r *http.Request) {
	tem, status := DecodeTagEventMapRequest(w, r)
	if status != 0 {
		return
	}
	// iterating over all tags and adding them if not exist
	for _, t := range tem.Tags {
		// getting or creating the new tag
		tmp := models.Tag{Name: t}
		tmp.GetIDCreate(g.DB)
	}

	var IDs []primitive.ObjectID
	username := w.Header().Get("user")
	// iterating over all events and add them if not exist
	for _, e := range tem.Events {
		if err := e.GetEventCreate(g.DB, g.GetUserPermission(w, false), username); err != nil {
			_http.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		IDs = append(IDs, e.ID)
	}

	// execute bulk update
	_, err := models.BulkAddTagEvent(g.DB, tem.Tags, IDs, g.GetUserPermission(w, false))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not bulk update documents!")
		return
	}

	media, err := models.GetEventsByIDs(g.DB, IDs, g.GetUserPermission(w, false))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, media)
	return
}

// UpdateEventByID handles the webrequest for updating the Event with the passed request body
func (g *AppGateway) UpdateEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// store new model in tmp object
	var ue models.Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ue); err != nil {
		// error occured during encoding
		_http.RespondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	defer r.Body.Close()
	// verify that no other object will be overwritten
	if ue.ID != id {
		_http.RespondWithError(w, http.StatusBadRequest, "id's do not match")
		return
	}
	// trying to update model with requested body
	e := models.Event{ID: id}
	_, err := e.UpdateEvent(g.DB, ue, g.GetUserPermission(w, true))
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// trying to select updated event
	if err = e.GetEvent(g.DB, g.GetUserPermission(w, false)); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, e)
}
