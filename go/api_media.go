package primboard

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddMedia handles the webrequest for Media creation
func (a *App) AddMedia(w http.ResponseWriter, r *http.Request) {
	var m Media
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// url and type are mandatory
	if m.URL == "" || m.Type == "" {
		RespondWithError(w, http.StatusBadRequest, "URL and type cannot be empty")
		return
	}
	// setting creation timestamp
	m.TimestampUpload = int64(time.Now().Unix())
	// try to insert model into db
	result, err := m.AddMedia(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteMediaByID handles the webrequest for Media deletion
func (a *App) DeleteMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// create model by passed id
	m := Media{ID: id}
	// try to delete model
	result, err := m.DeleteMedia(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	RespondWithJSON(w, http.StatusOK, result)
}

// GetMedia handles the webrequest for receiving all media
func (a *App) GetMedia(w http.ResponseWriter, r *http.Request) {
	ms, err := GetAllMedia(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, ms)
}

// GetMediaByID handles the webrequest for receiving Media model by id
func (a *App) GetMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["_id"])
	// create model by passed id
	m := Media{ID: id}
	// try to select media
	if err := m.GetMedia(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "Media not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select media from mongo
	RespondWithJSON(w, http.StatusOK, m)
}

// GetMediaByHash Handles the webrequest for receiving Media model by ipfs hash
// and mongo id
func (a *App) GetMediaByHash(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	parts := strings.Split(vars["ipfs_id"], "_")
	id, _ := primitive.ObjectIDFromHex(parts[1])
	//create model by passed hash
	m := Media{ID: id}
	// try to select media
	if err := m.GetMedia(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "Media not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if m.Sha1 != parts[0] {
		RespondWithError(w, http.StatusForbidden, "Invalid ipfs/id combination!")
		return
	}
	// could select media from mongo
	RespondWithJSON(w, http.StatusOK, m)
}

// UpdateMediaByID handles the webrequest for updating the Media with the passed
// request body
func (a *App) UpdateMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	parts := strings.Split(vars["ipfs_id"], "_")
	id, _ := primitive.ObjectIDFromHex(parts[1])
	// store new model in tmp object
	var um Media
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&um); err != nil {
		// error occured during encoding
		RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	m := Media{ID: id}
	err := m.UpdateMedia(a.DB, um)
	if err != nil {
		// Error occured during update
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	RespondWithJSON(w, http.StatusOK, m)
}
