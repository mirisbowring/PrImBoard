package swagger

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// AddTag handles the webrequest for Tag creation
func (a *App) AddTag(w http.ResponseWriter, r *http.Request) {
	var t Tag
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// name is mandatory
	if t.Name == "" {
		RespondWithError(w, http.StatusBadRequest, "Tagname cannot be empty")
		return
	}
	// try to insert model into db
	result, err := t.AddTag(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteTagByID handles the webrequest for Tag deletion
func (a *App) DeleteTagByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// create model by passed id
	t := Tag{ID: id}
	// try to delete model
	result, err := t.DeleteTag(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	RespondWithJSON(w, http.StatusOK, result)
}

// GetTagByID handles the webrequest for receiving Tag model by id
func (a *App) GetTagByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// create model by passed id
	t := Tag{ID: id}
	// try to select user
	if err := t.GetTag(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "Tag not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select user from mongo
	RespondWithJSON(w, http.StatusOK, t)
}

// UpdateTagByID handles the webrequest for updating the Tag with the passed request body
func (a *App) UpdateTagByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// store new model in tmp object
	var ut Tag
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ut); err != nil {
		// error occured during encoding
		RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	t := Tag{ID: id}
	result, err := t.UpdateTag(a.DB, ut)
	if err != nil {
		// Error occured during update
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	RespondWithJSON(w, http.StatusOK, result)
}
