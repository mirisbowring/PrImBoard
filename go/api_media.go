package primboard

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	// set the username
	m.Creator = w.Header().Get("user")
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
	var after primitive.ObjectID
	var size int
	// check if after query param is present
	tmp, ok := r.URL.Query()["after"]
	if !ok || len(tmp[0]) < 1 {
		// no after specified (selecting from top)
		// after = primitive.NewObjectID()
	} else {
		after, _ = primitive.ObjectIDFromHex(tmp[0])
	}
	// check if page size query param is present
	tmp, ok = r.URL.Query()["size"]
	if !ok || len(tmp[0]) < 1 {
		// no page size specified (using default)
		size = a.Config.DefaultMediaPageSize
	} else if i, err := strconv.Atoi(tmp[0]); err != nil {
		// page size is not an int
		size = a.Config.DefaultMediaPageSize
	} else {
		// page size set
		size = i
	}

	// ms, err := GetAllMedia(a.DB)
	ms, err := GetMediaPage(a.DB, after, int64(size))
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
	id, _ := primitive.ObjectIDFromHex(vars["id"])
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

// UpdateMediaByHash handles the webrequest for updating the Media with the passed
// request body
func (a *App) UpdateMediaByHash(w http.ResponseWriter, r *http.Request) {
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
	if err := um.checkComments(w.Header().Get("user")); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
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
