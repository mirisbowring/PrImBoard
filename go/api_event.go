package swagger

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
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
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
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

// GetEventByID handles the webrequest for receiving Event model by id
func (a *App) GetEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
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

// UpdateEventByID handles the webrequest for updating the Event with the passed request body
func (a *App) UpdateEventByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
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
