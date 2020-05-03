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

// AddCommentByMediaID appends a comment to the specified media
func (a *App) AddCommentByMediaID(w http.ResponseWriter, r *http.Request) {
	var c Comment
	// decode body into comment model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// parse route
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new comment
	m.Comments = append(m.Comments, &c)
	if err := m.Save(a.DB); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	RespondWithJSON(w, http.StatusOK, m)
}

// AddTagByMediaID appends a tag to the specified media
// creates a new tag if not in the tag document
func (a *App) AddTagByMediaID(w http.ResponseWriter, r *http.Request) {
	var t Tag
	// decode body into tag model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := t.GetIDCreate(a.DB); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch tag id")
	}

	// parse route
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new tag if not present
	if err := m.AddTag(a.DB, t.ID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	RespondWithJSON(w, http.StatusOK, m)
}

// AddTagsByMediaID appends multiple tags to the specified media
// creates a new tag if not in the tag document
func (a *App) AddTagsByMediaID(w http.ResponseWriter, r *http.Request) {
	var tags []Tag
	var IDs []primitive.ObjectID
	// decode body into tag model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tags); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// iterating over all tags and adding them if not exist
	for _, t := range tags {
		// getting or creating the new tag
		t.GetIDCreate(a.DB)
		IDs = append(IDs, t.ID)
	}

	// parse route
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new tag if not present
	if err := m.AddTags(a.DB, IDs); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

// DeleteMediaByID handles the webrequest for Media deletion
func (a *App) DeleteMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
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
	var query MediaQuery

	// check if event query param is present
	tmp, ok := r.URL.Query()["event"]
	if ok && len(tmp[0]) > 1 {
		query.Event, _ = primitive.ObjectIDFromHex(tmp[0])
	}

	// check if filter query param is present
	tmp, ok = r.URL.Query()["filter"]
	if ok && len(tmp[0]) > 1 {
		query.Filter = tmp[0]
	}

	// check if after query param is present
	tmp, ok = r.URL.Query()["after"]
	if ok && len(tmp[0]) > 1 {
		query.After, _ = primitive.ObjectIDFromHex(tmp[0])
	}

	// check if before query param is present
	tmp, ok = r.URL.Query()["before"]
	if ok && len(tmp[0]) > 1 {
		query.Before, _ = primitive.ObjectIDFromHex(tmp[0])
	}

	tmp, ok = r.URL.Query()["dsc"]
	if ok && len(tmp[0]) > 1 {
		if b, _ := strconv.ParseBool(tmp[0]); b {
			query.ASC = 1
		} else {
			query.ASC = -1
		}
	}

	// check if from query param is present
	tmp, ok = r.URL.Query()["from"]
	if ok && len(tmp[0]) > 1 {
		query.From, _ = primitive.ObjectIDFromHex(tmp[0])
	}

	// check if before query param is present
	tmp, ok = r.URL.Query()["until"]
	if ok && len(tmp[0]) > 1 {
		query.Until, _ = primitive.ObjectIDFromHex(tmp[0])
	}

	// check if page size query param is present
	tmp, ok = r.URL.Query()["size"]
	if !ok || len(tmp[0]) < 1 {
		// no page size specified (using default)
		query.Size = a.Config.DefaultMediaPageSize
	} else if i, err := strconv.Atoi(tmp[0]); err != nil {
		// page size is not an int
		query.Size = a.Config.DefaultMediaPageSize
	} else {
		// page size set
		query.Size = i
	}

	// ms, err := GetAllMedia(a.DB)
	ms, err := GetMediaPage(a.DB, query)
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

// UpdateMediaByID handles the webrequest for updating the Media with the passed
// request body
func (a *App) UpdateMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(vars["id"])
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
