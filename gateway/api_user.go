package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler/infrastructure"
	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/mongo"
)

//AddIPFSNodeByUsername adds an ipfs-node setting to the users settings
func (g *AppGateway) AddIPFSNodeByUsername(w http.ResponseWriter, r *http.Request) {
	// try to parse node
	var ipfs models.IPFSNode
	ipfs, status := DecodeIPFSNodeRequest(w, r, ipfs)
	if status != 0 {
		return
	}

	var u models.User
	// parse request
	vars := mux.Vars(r)
	// create model by passed username
	u = models.User{Username: vars["username"]}
	// try to select model
	if err := u.GetUser(g.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if err := u.AddIPFSNode(ipfs); err != nil {
		// IPFSNode Setting is not valid
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
	}
	// Save changes
	u.Save(g.DB)

	_http.RespondWithJSON(w, http.StatusOK, u)

}

// CreateUser handles the webrequest for user creation
func (g *AppGateway) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// username and password are mandatory
	if u.Username == "" || u.Password == "" {
		_http.RespondWithError(w, http.StatusBadRequest, "Username and Password must be set!")
		return
	}
	// verify invite
	i := models.Invite{Token: u.Token}
	if err := i.Invalidate(g.DB); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// hash password
	u.Password = HashPassword(u.Password)
	// try to insert model into db
	result, err := u.CreateUser(g.DB)
	if err != nil {
		// prevent the Token from expiration
		i.Revalidate(g.DB, g.Config.InviteValidity)
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// creation successful
	_http.RespondWithJSON(w, http.StatusCreated, result)
}

// DeleteUser handles the webrequest for user deletion
func (g *AppGateway) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	// create model by passed username
	u := models.User{Username: vars["username"]}
	u, status := ParseUserSelect(g.DB, w, r, true)
	if status > 0 {
		return
	}
	// try to delete model
	result, err := u.DeleteUser(g.DB)
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// deletion successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}

// GetIPFSNodesByUsername returns all IPFSNodes for the current user
func (g *AppGateway) GetIPFSNodesByUsername(w http.ResponseWriter, r *http.Request) {
	// parse request an retrieve user
	u, status := ParseUserSelect(g.DB, w, r, true)
	if status > 0 {
		// cancel request if error occured
		return
	}

	_http.RespondWithJSON(w, http.StatusOK, u.Settings.IPFSNodes)
}

// GetSettingsByUsername returns the settings for a given user
func (g *AppGateway) GetSettingsByUsername(w http.ResponseWriter, r *http.Request) {
	// parse request an retrieve user
	u, status := ParseUserSelect(g.DB, w, r, true)
	if status > 0 {
		// cancel request if error occured
		return
	}

	_http.RespondWithJSON(w, http.StatusOK, u.Settings)
}

// GetUserByUsername handles the webrequest for receiving user model by username
func (g *AppGateway) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	// parse request an retrieve user
	u, status := ParseUserSelect(g.DB, w, r, false)
	if status > 0 {
		// cancel request if error occured
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, u)
}

// LoginUser Handles the webrequest for logging the user in
func (g *AppGateway) LoginUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	// decode request into model
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	// validate that username is not empty to prevent high db load
	if u.Username == "" {
		_http.RespondWithError(w, http.StatusBadRequest, "Username cannot be empty!")
		return
	}
	// read corresponding user from db
	var us models.User
	us.Username = u.Username
	if err := us.GetUser(g.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusForbidden, "Login Failed!")
		default:
			// any other error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// hash passed password and compare
	if err := MatchesBcrypt(u.Password, us.Password); err != nil {
		// Passwords do not match
		_http.RespondWithError(w, http.StatusForbidden, "Login Failed!")
		return
	}

	// Create new Session
	session := g.NewSession(u, g.DB)

	// select nodes, the user has access to
	nodes, err := models.GetAllNodes(g.DB, g.GetUserPermission(w, false), "auth")
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not retrieve nodes")
		return
	}

	// authenticate user to nodes
	if status, msg := infrastructure.NodeAuthentication(session, nodes, true, g.HTTPClient); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	// append session and create cookie
	g.Sessions = append(g.Sessions, session)
	g.SetSessionCookie(&w, r, session)

	// assuming that the user was logged in
	_http.RespondWithJSON(w, http.StatusOK, "Login was successful!")
}

// LogoutUser handles the webrequest for logging the user out
func (g *AppGateway) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session := g.GetSessionByUsername(_http.GetUsernameFromHeader(w))
	// unauthenticating user from nodes
	if status, msg := infrastructure.NodeAuthentication(session, g.Nodes, false, g.HTTPClient); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	// remove session from session store and delete from response
	g.CloseSession(&w, r)

	_http.RespondWithJSON(w, http.StatusNoContent, "Logout was successfull!")
}

// UpdateUser handles the webrequest for updating the user with the passed request body
func (g *AppGateway) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// parse request
	vars := mux.Vars(r)
	// store new model in tmp object
	var uu models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&uu); err != nil {
		// error occured during encoding
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid Request payload")
		return
	}
	defer r.Body.Close()
	// trying to update model with requested body
	u := models.User{Username: vars["username"]}
	result, err := u.UpdateUser(g.DB, uu)
	if err != nil {
		// Error occured during update
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Update successful
	_http.RespondWithJSON(w, http.StatusOK, result)
}
