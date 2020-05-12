package primboard

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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

	// verify the usergroup
	if err := ug.Verify(a.DB); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// try to insert model into db skipVerify because already verified
	result, err := ug.AddUserGroup(a.DB, true)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	RespondWithJSON(w, http.StatusCreated, result)
}

// AddUserToUserGroupByID adds a User to the specified usergroup
func (a *App) AddUserToUserGroupByID(w http.ResponseWriter, r *http.Request) {
	var u User
	// decode
	u, status := DecodeUserRequest(w, r, u)
	if status != 0 {
		return
	}
	// check if user does Exists
	if !u.Exists(a.DB) {
		// user does not exist
		RespondWithError(w, http.StatusBadRequest, "Could not add user!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// init usergroup
	ug := UserGroup{ID: id}
	if ug.GetUserGroupAPI(w, a.DB) != 0 {
		return
	}

	// verify that user is owner
	if w.Header().Get("user") != ug.Creator {
		RespondWithError(w, http.StatusUnauthorized, "You do not own this group!")
		return
	}

	// check if user already in group
	if _, found := Find(ug.Users, u.Username); found {
		RespondWithError(w, http.StatusFound, "User already added to usergroup!")
		return
	}

	//append user and save object to db
	ug.Users = append(ug.Users, u.Username)
	//skipVerify because we manually added a single username and checked uniqueness
	if err := ug.Save(a.DB, true); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// success
	a.GetUserGroupByID(w, r)
}

// AddUsersToUserGroupByID adds a User to the specified usergroup
func (a *App) AddUsersToUserGroupByID(w http.ResponseWriter, r *http.Request) {
	var u []User
	// decode
	u, status := DecodeUsersRequest(w, r, u)
	if status != 0 {
		return
	}

	// check if user does Exists
	if !UsersExist(a.DB, u) {
		// user does not exist
		RespondWithError(w, http.StatusBadRequest, "Could not add users!")
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// init usergroup
	ug := UserGroup{ID: id}
	if ug.GetUserGroupAPI(w, a.DB) != 0 {
		return
	}

	// verify that user is owner
	if w.Header().Get("user") != ug.Creator {
		RespondWithError(w, http.StatusUnauthorized, "You do not own this group!")
		return
	}

	// check adding all users and make slice unique
	for _, user := range u {
		ug.Users = append(ug.Users, user.Username)
	}
	ug.Users = UniqueStrings(ug.Users)

	//skipVerify because we manually added a single username and checked uniqueness
	if err := ug.Save(a.DB, true); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// success
	a.GetUserGroupByID(w, r)
}

// DeleteUserGroupByID handles the webrequest for usergroup deletion
func (a *App) DeleteUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
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
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// create model by passed id
	ug := UserGroup{ID: id}
	if ug.GetUserGroupAPI(w, a.DB) != 0 {
		return
	}
	// could select user from mongo
	RespondWithJSON(w, http.StatusOK, ug)
}

// GetUserGroups returns all groups, the current user is assigned to
func (a *App) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// receive current user
	username := w.Header().Get("user")
	// read groups from db
	groups, err := GetUserGroups(a.DB, username)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// success
	RespondWithJSON(w, http.StatusOK, groups)
}

// GetUserGroupsByName returns available Tags by their name, starting with
func (a *App) GetUserGroupsByName(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	keyword := vars["name"]
	groups, err := GetUserGroupsByKeyword(a.DB, keyword, a.Config.TagPreviewLimit)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, groups)
}

// RemoveUserFromUserGroupByID adds a User to the specified usergroup
func (a *App) RemoveUserFromUserGroupByID(w http.ResponseWriter, r *http.Request) {
	var u User

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// parse username from route
	u, status := parseUsername(w, r)
	if status != 0 {
		return
	}

	// init usergroup
	ug := UserGroup{ID: id}
	if ug.GetUserGroupAPI(w, a.DB) != 0 {
		return
	}

	// remove username from slice
	ug.Users = RemoveString(ug.Users, u.Username)

	if err := ug.Save(a.DB, false); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// success
	a.GetUserGroupByID(w, r)
}

// RemoveUsersFromUserGroupByID adds a User to the specified usergroup
func (a *App) RemoveUsersFromUserGroupByID(w http.ResponseWriter, r *http.Request) {
	var u []User
	// decode
	u, status := DecodeUsersRequest(w, r, u)
	if status != 0 {
		return
	}

	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}

	// init usergroup
	ug := UserGroup{ID: id}
	if ug.GetUserGroupAPI(w, a.DB) != 0 {
		return
	}

	// remove usernames from slice
	for _, user := range u {
		ug.Users = RemoveString(ug.Users, user.Username)
	}

	if err := ug.Save(a.DB, false); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// success
	a.GetUserGroupByID(w, r)
}

// UpdateUserGroupByID handles the webrequest for updating the usergroup with
// the passed request body
func (a *App) UpdateUserGroupByID(w http.ResponseWriter, r *http.Request) {
	// parse ID from route
	id := parseID(w, r)
	if id.IsZero() {
		return
	}
	// store new model in tmp object
	var uug UserGroup
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&uug); err != nil {
		// error occured during encoding
		RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()

	// verify the usergroup
	if err := uug.Verify(a.DB); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// trying to update model with requested body
	ug := UserGroup{ID: id}
	result, err := ug.UpdateUserGroup(a.DB, uug, true)
	if err != nil {
		// Error occured during update
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	RespondWithJSON(w, http.StatusOK, result)
}
