package primboard

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddTag handles the webrequest for Tag creation
func (a *App) AddTag(w http.ResponseWriter, r *http.Request) {
	var t string
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	tmp := Tag{Name: t}
	// name is mandatory
	if tmp.Name == "" {
		_http.RespondWithError(w, http.StatusBadRequest, "Tagname cannot be empty")
		return
	}
	// try to insert model into db
	result, err := tmp.AddTag(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteTagByID handles the webrequest for Tag deletion
func (a *App) DeleteTagByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	t := Tag{ID: id}
	// try to delete model
	result, err := t.DeleteTag(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}

// GetTagByID handles the webrequest for receiving Tag model by id
func (a *App) GetTagByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create model by passed id
	t := Tag{ID: id}
	// try to select user
	if err := t.GetTag(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "Tag not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, t)
}

// GetTags returns all Tags available
func (a *App) GetTags(w http.ResponseWriter, r *http.Request) {
	// var t Tag

}

// GetTagsByName returns available Tags by their name, starting with
func (a *App) GetTagsByName(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	keyword := vars["name"]
	tags, err := GetTagsByKeyword(a.DB, keyword, a.Config.TagPreviewLimit)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// clean to string slice
	var tagnames []string
	for _, tag := range tags {
		tagnames = append(tagnames, tag.Name)
	}
	_http.RespondWithJSON(w, http.StatusOK, tagnames)
}

// UpdateTagByID handles the webrequest for updating the Tag with the passed request body
func (a *App) UpdateTagByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// store new model in tmp object
	var ut Tag
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ut); err != nil {
		// error occured during encoding
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	t := Tag{ID: id}
	result, err := t.UpdateTag(a.DB, ut)
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}
