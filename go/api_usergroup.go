package swagger

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

// AddUserGroup handles the webrequest for usergroup creation
func (a *App) AddUserGroup(w http.ResponseWriter, r *http.Request) {
	var ug UserGroup
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ug); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// title is mandatory
	if ug.Title == "" {
		RespondWithError(w, http.StatusBadRequest, "Title cannot be empty")
		return
	}
	// setting creation timestamp
	ug.TimestampCreation = int64(time.Now().Unix())
	// try to insert model into db
	result, err := ug.AddUserGroup(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteUserGroupByID handles the webrequest for usergroup deletion
func (a *App) DeleteUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// create model by passed id
	ug := UserGroup{ID: id}
	// try to delete model
	result, err := ug.DeleteUserGroup(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	RespondWithJSON(w, http.StatusOK, result)
}

//GetUserGroupByID handles the webrequest for receiving usergroup model by id
func (a *App) GetUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// create model by passed id
	ug := UserGroup{ID: id}
	// try to select user
	if err := ug.GetUserGroup(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "Usergroup not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select user from mongo
	RespondWithJSON(w, http.StatusOK, ug)
}

// UpdateUserGroupByID handles the webrequest for updating the usergroup with
// the passed request body
func (a *App) UpdateUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// store new model in tmp object
	var uug UserGroup
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&uug); err != nil {
		// error occured during encoding
		RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	ug := UserGroup{ID: id}
	result, err := ug.UpdateUserGroup(a.DB, uug)
	if err != nil {
		// Error occured during update
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	RespondWithJSON(w, http.StatusOK, result)
}
