package primboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddEvent handles the webrequest for Event creation
func (a *App) AddEvent(w http.ResponseWriter, r *http.Request) {
	var e Event
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// title is mandatory
	if e.Title == "" {
		RespondWithError(w, http.StatusBadRequest, "Title cannot be empty")
		return
	}
	// setting creation timestamp
	e.TimestampCreation = int64(time.Now().Unix())
	// try to insert model into db
	result, err := e.AddEvent(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteEventByID handles the webrequest for Event deletion
func (a *App) DeleteEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := Event{ID: id}
	// try to delete model
	result, err := e.DeleteEvent(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	RespondWithJSON(w, http.StatusOK, result)
}

// GetEvents handles the webrequest for receiving all events
func (a *App) GetEvents(w http.ResponseWriter, r *http.Request) {
	es, err := GetAllEvents(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, es)
}

// GetEventByID handles the webrequest for receiving Event model by id
func (a *App) GetEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	e := Event{ID: id}
	// try to select user
	if err := e.GetEvent(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "Event not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select user from mongo
	RespondWithJSON(w, http.StatusOK, e)
}

// GetEventsByName returns available Events by their name, starting with
func (a *App) GetEventsByName(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	keyword := vars["title"]
	events, err := GetEventsByKeyword(a.DB, getPermission(w), keyword, a.Config.TagPreviewLimit)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, events)
}

// MapTagsToEvents maps a slice of Tags to a slice of events
func (a *App) MapTagsToEvents(w http.ResponseWriter, r *http.Request) {
	tem, status := DecodeTagEventMapRequest(w, r)
	if status != 0 {
		return
	}
	// iterating over all tags and adding them if not exist
	for _, t := range tem.Tags {
		// getting or creating the new tag
		tmp := Tag{Name: t}
		tmp.GetIDCreate(a.DB)
	}

	var IDs []primitive.ObjectID
	// iterating over all events and add them if not exist
	for _, e := range tem.Events {
		if err := e.GetEventCreate(a.DB); err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		IDs = append(IDs, e.ID)
	}

	// execute bulk update
	_, err := BulkAddTagEvent(a.DB, tem.Tags, IDs, getPermission(w))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Could not bulk update documents!")
		return
	}

	media, err := GetEventsByIDs(a.DB, IDs, getPermission(w))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, media)
	return
}

// UpdateEventByID handles the webrequest for updating the Event with the passed request body
func (a *App) UpdateEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// store new model in tmp object
	var ue Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ue); err != nil {
		// error occured during encoding
		RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	e := Event{ID: id}
	result, err := e.UpdateEvent(a.DB, ue)
	if err != nil {
		// Error occured during update
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	RespondWithJSON(w, http.StatusOK, result)
}
