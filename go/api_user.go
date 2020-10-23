package primboard

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddIPFSNodeByUser adds an ipfs-node setting to the users settings
func (a *App) AddIPFSNodeByUsername(w http.ResponseWriter, r *http.Request) {
	// try to parse node
	var ipfs IPFSNode
	ipfs, status := DecodeIPFSNodeRequest(w, r, ipfs)
	if status != 0 {
		return
	}

	var u User
	// parse request
	vars := mux.Vars(r)
	// create model by passed username
	u = User{Username: vars["username"]}
	// try to select model
	if err := u.GetUser(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if err := u.AddIPFSNode(ipfs); err != nil {
		// IPFSNode Setting is not valid
		RespondWithError(w, http.StatusBadRequest, err.Error())
	}
	// Save changes
	u.Save(a.DB)

	RespondWithJSON(w, http.StatusOK, u)

}

// CreateUser handles the webrequest for user creation
func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// username and password are mandatory
	if u.Username == "" || u.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Username and Password must be set!")
		return
	}
	// verify invite
	i := Invite{Token: u.Token}
	if err := i.Invalidate(a.DB); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// hash password
	u.Password = HashPassword(u.Password)
	// try to insert model into db
	result, err := u.CreateUser(a.DB)
	if err != nil {
		// prevent the Token from expiration
		i.Revalidate(a.DB, a.Config.InviteValidity)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteUser handles the webrequest for user deletion
func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	// create model by passed username
	u := User{Username: vars["username"]}
	u, status := a.parseUserSelect(w, r, true)
	if status > 0 {
		return
	}
	// try to delete model
	result, err := u.DeleteUser(a.DB)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	RespondWithJSON(w, http.StatusOK, result)
}

// GetIPFSNodesByUsername returns all IPFSNodes for the current user
func (a *App) GetIPFSNodesByUsername(w http.ResponseWriter, r *http.Request) {
	// parse request an retrieve user
	u, status := a.parseUserSelect(w, r, true)
	if status > 0 {
		// cancel request if error occured
		return
	}

	RespondWithJSON(w, http.StatusOK, u.Settings.IPFSNodes)
}

func (a *App) GetSettingsByUsername(w http.ResponseWriter, r *http.Request) {
	// parse request an retrieve user
	u, status := a.parseUserSelect(w, r, true)
	if status > 0 {
		// cancel request if error occured
		return
	}

	RespondWithJSON(w, http.StatusOK, u.Settings)
}

// GetUserByUsername handles the webrequest for receiving user model by username
func (a *App) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	// parse request an retrieve user
	u, status := a.parseUserSelect(w, r, false)
	if status > 0 {
		// cancel request if error occured
		return
	}
	// could select user from mongo
	RespondWithJSON(w, http.StatusOK, u)
}

// LoginUser Handles the webrequest for logging the user in
func (a *App) LoginUser(w http.ResponseWriter, r *http.Request) {
	var u User
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// validate that username is not empty to prevent high db load
	if u.Username == "" {
		RespondWithError(w, http.StatusBadRequest, "Username cannot be empty!")
		return
	}
	// read corresponding user from db
	var us User
	us.Username = u.Username
	if err := us.GetUser(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusForbidden, "Login Failed!")
		default:
			// any other error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// hash passed password and compare
	if err := MatchesBcrypt(u.Password, us.Password); err != nil {
		// Passwords do not match
		RespondWithError(w, http.StatusForbidden, "Login Failed!")
		return
	}
	// Create new Session and Cookie
	SetSessionCookie(&w, r, a.NewSession(u))
	// assuming that the user was logged in
	RespondWithJSON(w, http.StatusOK, "Login was successful!")

}

// LogoutUser handles the webrequest for logging the user out
func (a *App) LogoutUser(w http.ResponseWriter, r *http.Request) {
	CloseSession(&w, r)
	RespondWithJSON(w, http.StatusNoContent, "Logout was successfull!")
}

// UpdateUser handles the webrequest for updating the user with the passed request body
func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	// store new model in tmp object
	var uu User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&uu); err != nil {
		// error occured during encoding
		RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	u := User{Username: vars["username"]}
	result, err := u.UpdateUser(a.DB, uu)
	if err != nil {
		// Error occured during update
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	RespondWithJSON(w, http.StatusOK, result)
}
