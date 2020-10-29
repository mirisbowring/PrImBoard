package primboard

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddMedia handles the webrequest for Media creation
func (a *App) AddMedia(w http.ResponseWriter, r *http.Request) {
	var m Media
	m, status := DecodeMediaRequest(w, r, m)
	if status != 0 {
		return
	}

	// url and type are mandatory
	if m.URL == "" || m.Type == "" {
		_http.RespondWithError(w, http.StatusBadRequest, "URL and type cannot be empty")
		return
	}
	// setting creation timestamp
	m.TimestampUpload = int64(time.Now().Unix())
	// set the username
	m.Creator = w.Header().Get("user")
	// try to insert model into db
	result, err := m.AddMedia(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// AddCommentByMediaID appends a comment to the specified media
func (a *App) AddCommentByMediaID(w http.ResponseWriter, r *http.Request) {
	var c Comment
	c, status := DecodeCommentRequest(w, r, c)
	if status != 0 {
		return
	}

	c.AddMetadata(w.Header().Get("user"))
	// verifiy integrity of comment
	if err := c.IsValid(); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new comment
	m.Comments = append(m.Comments, &c)
	if err := m.Save(a.DB); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

//AddDescriptionByMediaID adds the description to the media
func (a *App) AddDescriptionByMediaID(w http.ResponseWriter, r *http.Request) {
	var m Media
	m, status := DecodeMediaRequest(w, r, m)
	if status != 0 {
		return
	}

	// check if description is valid
	if strings.TrimSpace(m.Description) == "" {
		// description should not be empty
		_http.RespondWithError(w, http.StatusBadRequest, "Title cannot be empty!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	media := Media{ID: id}
	// add the title to the database document
	media.Description = m.Description
	// append the new tag if not present
	if err := m.Save(a.DB); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

// AddTagByMediaID appends a tag to the specified media
// creates a new tag if not in the tag document
func (a *App) AddTagByMediaID(w http.ResponseWriter, r *http.Request) {
	var t string
	t, status := DecodeTagStringRequest(w, r, t)
	if status != 0 {
		return
	}

	t, err := VerifyTag(a.DB, t)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not process tag")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new tag if not present
	if err := m.AddTag(a.DB, t); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

// AddTagsByMediaID appends multiple tags to the specified media
// creates a new tag if not in the tag document
func (a *App) AddTagsByMediaID(w http.ResponseWriter, r *http.Request) {
	var tags []string
	tags, status := DecodeTagStringsRequest(w, r, tags)
	if status != 0 {
		return
	}

	tagnames, err := VerifyTags(a.DB, tags)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not process tags")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new tag if not present
	if err := m.AddTags(a.DB, tagnames); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

// AddUserGroupsByMediaID appends multiple tags to the specified media
// creates a new tag if not in the tag document
func (a *App) AddUserGroupsByMediaID(w http.ResponseWriter, r *http.Request) {
	var groups []UserGroup
	var IDs []primitive.ObjectID
	groups, status := DecodeUserGroupsRequest(w, r, groups)
	if status != 0 {
		return
	}
	// creating id slice
	for _, t := range groups {
		IDs = append(IDs, t.ID)
	}

	groups, err := GetUserGroupsByIDs(a.DB, IDs)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If length does not match, receive all new IDs
	// could be that a requested group does not exist
	if len(groups) != len(IDs) {
		IDs = nil
		for _, group := range groups {
			IDs = append(IDs, group.ID)
		}
	}

	// verify that any valid group was specified
	if IDs == nil {
		_http.RespondWithError(w, http.StatusBadRequest, "No valid groups specified!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	m := Media{ID: id}
	// append the new tag if not present
	if err := m.AddUserGroups(a.DB, IDs); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

//AddTimestampByMediaID adds the creation date to the media
func (a *App) AddTimestampByMediaID(w http.ResponseWriter, r *http.Request) {
	var m Media
	m, status := DecodeMediaRequest(w, r, m)
	if status != 0 {
		return
	}

	// check if timestamp is valid
	if m.Timestamp == 0 {
		_http.RespondWithError(w, http.StatusBadRequest, "Creation date cannot be empty!")
		return
	}

	// verify that the creation date is not in the future
	if time.Unix(m.Timestamp, 0).UTC().After(time.Now().UTC()) {
		_http.RespondWithError(w, http.StatusBadRequest, "Creation date cannot be the future!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	media := Media{ID: id}
	// add the timestamp to the database document
	media.Timestamp = m.Timestamp
	// append the new tag if not present
	if err := m.Save(a.DB); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

//AddTitleByMediaID adds the title to the media
func (a *App) AddTitleByMediaID(w http.ResponseWriter, r *http.Request) {
	var m Media
	m, status := DecodeMediaRequest(w, r, m)
	if status != 0 {
		return
	}

	// check if title is valid
	if strings.TrimSpace(m.Title) == "" {
		// title should not be empty
		_http.RespondWithError(w, http.StatusBadRequest, "Title cannot be empty!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create media model by id to select from db
	media := Media{ID: id}
	// add the title to the database document
	media.Title = m.Title
	// append the new tag if not present
	if err := m.Save(a.DB); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Error during document update")
		return
	}
	// success
	a.GetMediaByID(w, r)
}

// DeleteMediaByID handles the webrequest for Media deletion
func (a *App) DeleteMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create model by passed id
	m := Media{ID: id}
	// try to delete model
	result, err := m.DeleteMedia(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
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

	tmp, ok = r.URL.Query()["asc"]
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
	ms, err := GetMediaPage(a.DB, query, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, ms)
}

// GetMediaByID handles the webrequest for receiving Media model by id
func (a *App) GetMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create model by passed id
	m := Media{ID: id}
	// try to select media
	if err := m.GetMedia(a.DB, getPermission(w)); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "Media not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// could select media from mongo
	_http.RespondWithJSON(w, http.StatusOK, m)
}

// GetMediaByIDs handles the webrequest for receiving Media models by ids
func (a *App) GetMediaByIDs(w http.ResponseWriter, r *http.Request) {
	m, status := DecodeMediasRequest(w, r)
	if status != 0 {
		return
	}
	var IDs []primitive.ObjectID
	for _, id := range m {
		IDs = append(IDs, id.ID)
	}
	media, err := GetMediaByIDs(a.DB, IDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, media)
	return
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
	if err := m.GetMedia(a.DB, getPermission(w)); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "Media not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if m.Sha1 != parts[0] {
		_http.RespondWithError(w, http.StatusForbidden, "Invalid ipfs/id combination!")
		return
	}
	// could select media from mongo
	_http.RespondWithJSON(w, http.StatusOK, m)
}

// MapEventsToMedia maps an media slice to each event entry
func (a *App) MapEventsToMedia(w http.ResponseWriter, r *http.Request) {
	mem, status := DecodeMediaEventMapRequest(w, r)
	if status != 0 {
		return
	}

	var eventIDs []primitive.ObjectID
	// iterating over all events and add them if not exist
	username := w.Header().Get("user")
	for _, e := range mem.Events {
		if err := e.GetEventCreate(a.DB, getPermission(w), username); err != nil {
			_http.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		eventIDs = append(eventIDs, e.ID)
	}
	// parsing ids
	mediaIDs, err := ParseIDs(mem.MediaIDs)
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// execute bulk update
	_, err = BulkAddMediaEvent(a.DB, mediaIDs, eventIDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not bulk update documents!")
		return
	}

	// select updated documents
	media, err := GetMediaByIDs(a.DB, mediaIDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, media)
	return
}

// MapGroupsToMedia maps an media slice to each event entry
func (a *App) MapGroupsToMedia(w http.ResponseWriter, r *http.Request) {
	mgm, status := DecodeMediaGroupMapRequest(w, r)
	if status != 0 {
		return
	}

	var groupIDs []primitive.ObjectID
	// iterating over all groups
	for _, g := range mgm.Groups {
		if err := g.GetUserGroup(a.DB, getPermission(w)); err != nil {
			_http.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		groupIDs = append(groupIDs, g.ID)
	}
	// parsing ids
	mediaIDs, err := ParseIDs(mgm.MediaIDs)
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// execute bulk update
	_, err = BulkAddMediaGroup(a.DB, mediaIDs, groupIDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not bulk update documents!")
		return
	}

	// select updated documents
	media, err := GetMediaByIDs(a.DB, mediaIDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, media)
	return
}

// MapTagsToMedia adds a list of tags to a list of media
func (a *App) MapTagsToMedia(w http.ResponseWriter, r *http.Request) {
	tmm, status := DecodeTagMediaMapRequest(w, r)
	if status != 0 {
		return
	}
	// iterating over all tags and adding them if not exist
	for _, t := range tmm.Tags {
		// getting or creating the new tag
		tmp := Tag{Name: t}
		tmp.GetIDCreate(a.DB)
	}
	// parsing ids
	IDs, err := ParseIDs(tmm.IDs)
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// execute bulk update
	_, err = BulkAddTagMedia(a.DB, tmm.Tags, IDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not bulk update documents!")
		return
	}

	media, err := GetMediaByIDs(a.DB, IDs, getPermission(w))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, media)
	return
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
	um, status := DecodeMediaRequest(w, r, um)
	if status != 0 {
		return
	}
	if err := um.checkComments(w.Header().Get("user")); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	m := Media{ID: id}
	err := m.UpdateMedia(a.DB, um)
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, m)
}

// UpdateMediaByID handles the webrequest for updating the Media with the passed
// request body
func (a *App) UpdateMediaByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// store new model in tmp object
	var um Media
	um, status := DecodeMediaRequest(w, r, um)
	if status != 0 {
		return
	}
	if err := um.checkComments(w.Header().Get("user")); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	m := Media{ID: id}
	err := m.UpdateMedia(a.DB, um)
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, m)
}

// UploadMedia handles the webrequest for uploading a file to the api
func (a *App) UploadMedia(w http.ResponseWriter, r *http.Request) {
	// grep node
	node := r.FormValue("node")
	n := Node{}
	if err := json.Unmarshal([]byte(node), &n); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, "could not unmarshal passed node")
		return
	}

	// verfiy node
	if err := n.GetNode(a.DB, getPermission(w)); err != nil {
		_http.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// grep filemeta
	meta := r.FormValue("filemeta")
	m := Media{}
	if err := json.Unmarshal([]byte(meta), &m); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, "could not unmarshal passed filemeta")
		return
	}

	// verify Tags
	var err error
	m.Tags, err = VerifyTags(a.DB, m.Tags)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not process tags")
		return
	}

	// receive file
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	// setting creation timestamp
	m.TimestampUpload = int64(time.Now().Unix())
	// set the username
	m.Creator = w.Header().Get("user")

	// create user's tmp dir
	_ = os.Mkdir("tmp/"+m.Creator, 0755)
	// Create file
	filename := "tmp/" + m.Creator + "/" + handler.Filename
	dst, err := os.Create(filename)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// file to specified node
	m, err = addMediaToIpfsNode(filename, m, n)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// try to insert model into db
	result, err := m.AddMedia(a.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// remove temporary file
	os.Remove(filename)

	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}
