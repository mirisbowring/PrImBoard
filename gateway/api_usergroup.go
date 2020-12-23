package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
)

// AddUserGroup handles the webrequest for usergroup creation
func (g *AppGateway) AddUserGroup(w http.ResponseWriter, r *http.Request) {
	var ug models.UserGroup
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ug); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// verify the usergroup
	if err := ug.Verify(g.DB); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// try to insert model into db skipVerify because already verified
	result, err := ug.AddUserGroup(g.DB, true)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// AddUserToUserGroupByID adds a User to the specified usergroup
func (g *AppGateway) AddUserToUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// decode
	u, status := _http.ParsePathString(w, r, "username")
	if status != 0 {
		return
	}
	// // check if user does Exists
	// if !u.Exists(g.DB) {
	// 	// user does not exist
	// 	_http.RespondWithError(w, http.StatusBadRequest, "Could not add user!")
	// 	return
	// }

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// init usergroup
	ug := models.UserGroup{ID: id}
	if g.GetUserGroupAPI(w, g.DB, &ug) != 0 {
		return
	}

	// verify that user is owner
	if _http.GetUsernameFromHeader(w) != ug.Creator {
		_http.RespondWithError(w, http.StatusUnauthorized, "You do not own this group!")
		return
	}

	// check if user already in group
	if _, found := helper.FindInSlice(ug.Users, u); found {
		_http.RespondWithError(w, http.StatusFound, "User already added to usergroup!")
		return
	}

	//append user and save object to db
	ug.Users = append(ug.Users, u)
	//skipVerify because we manually added a single username and checked uniqueness
	if err := ug.Save(g.DB, true); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// success
	g.GetUserGroupByID(w, r)
}

// AddUsersToUserGroupByID adds a User to the specified usergroup
func (g *AppGateway) AddUsersToUserGroupByID(w http.ResponseWriter, r *http.Request) {
	var u []string
	// decode
	u, status := DecodeStringsRequest(w, r, u)
	if status != 0 {
		return
	}
	// select all existing users from db that matches the given array
	// u, err := models.GetUsers(g.DB, u)
	// if err != nil {
	// 	_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }
	// verify, that any user was selected
	if len(u) == 0 {
		_http.RespondWithError(w, http.StatusBadRequest, "No valid users specified!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// init usergroup
	ug := models.UserGroup{ID: id}
	if g.GetUserGroupAPI(w, g.DB, &ug) != 0 {
		return
	}

	// verify that user is owner
	if _http.GetUsernameFromHeader(w) != ug.Creator {
		_http.RespondWithError(w, http.StatusUnauthorized, "You do not own this group!")
		return
	}

	// check adding all users and make slice unique
	for _, user := range u {
		ug.Users = append(ug.Users, user)
	}
	ug.Users = helper.UniqueStrings(ug.Users)

	//skipVerify because we manually added a single username and checked uniqueness
	if err := ug.Save(g.DB, true); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// success
	g.GetUserGroupByID(w, r)
}

// DeleteUserGroupByID handles the webrequest for usergroup deletion
func (g *AppGateway) DeleteUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create model by passed id
	ug := models.UserGroup{ID: id}
	// try to delete model
	result, err := ug.DeleteUserGroup(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}

//GetUserGroupByID handles the webrequest for receiving usergroup model by id
func (g *AppGateway) GetUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create model by passed id
	ug := models.UserGroup{ID: id}
	if g.GetUserGroupAPI(w, g.DB, &ug) != 0 {
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, ug)
}

// GetUserGroups returns all groups, the current user is assigned to
func (g *AppGateway) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// receive current user
	username := _http.GetUsernameFromHeader(w)
	// read groups from db
	groups, err := models.GetUserGroups(g.DB, username)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// success
	_http.RespondWithJSON(w, http.StatusOK, groups)
}

// GetUserGroupsByName returns available Tags by their name, starting with
func (g *AppGateway) GetUserGroupsByName(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	keyword := vars["name"]
	groups, err := models.GetUserGroupsByKeyword(g.DB, keyword, g.Config.TagPreviewLimit)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_http.RespondWithJSON(w, http.StatusOK, groups)
}

// RemoveUserFromUserGroupByID adds a User to the specified usergroup
func (g *AppGateway) RemoveUserFromUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// parse username from route
	username, status := _http.ParsePathString(w, r, "username")
	if status != 0 {
		return
	}
	// init usergroup
	ug := models.UserGroup{ID: id}
	if g.GetUserGroupAPI(w, g.DB, &ug) != 0 {
		return
	}

	// remove username from slice
	ug.Users = RemoveString(ug.Users, username)

	if err := ug.Save(g.DB, false); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// success
	g.GetUserGroupByID(w, r)
}

// RemoveUsersFromUserGroupByID adds a User to the specified usergroup
func (g *AppGateway) RemoveUsersFromUserGroupByID(w http.ResponseWriter, r *http.Request) {
	var u []string
	// decode
	u, status := DecodeStringsRequest(w, r, u)
	if status != 0 {
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// init usergroup
	ug := models.UserGroup{ID: id}
	if g.GetUserGroupAPI(w, g.DB, &ug) != 0 {
		return
	}

	// remove usernames from slice
	for _, user := range u {
		ug.Users = RemoveString(ug.Users, user)
	}

	if err := ug.Save(g.DB, false); err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// success
	g.GetUserGroupByID(w, r)
}

// UpdateUserGroupByID handles the webrequest for updating the usergroup with
// the passed request body
func (g *AppGateway) UpdateUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// store new model in tmp object
	var uug models.UserGroup
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&uug); err != nil {
		// error occured during encoding
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()

	// verify the usergroup
	if err := uug.Verify(g.DB); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// trying to update model with requested body
	ug := models.UserGroup{ID: id}
	result, err := ug.UpdateUserGroup(g.DB, uug, true)
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}
