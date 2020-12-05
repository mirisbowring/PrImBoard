package node

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
	"github.com/mirisbowring/primboard/internal/handler/session"
	"github.com/mirisbowring/primboard/internal/models"
	log "github.com/sirupsen/logrus"
)

// AuthenticateUser gets called after a user has been authenticated to the gateway
// creates temporary session symlink to be able to access the files
func (n *AppNode) authenticateUser(w http.ResponseWriter, r *http.Request) {
	var token string
	// get username
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}
	// decode request into string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&token); err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"error":    err.Error(),
		}).Error("could not decode session from request")
		_http.RespondWithError(w, http.StatusBadRequest, "could not decode session from request")
		return
	}
	defer r.Body.Close()
	// verify that session is not empty
	if token == "" {
		log.WithFields(log.Fields{
			"username": username,
		}).Error("passed invalid session (cannot be empty)")
		_http.RespondWithError(w, http.StatusBadRequest, "session cannot be empty")
		return
	}
	// save session
	if status, msg := n.AddSession(username, token); status > 0 {
		_http.RespondWithError(w, http.StatusBadRequest, msg)
		return
	}
	// link user path
	handler.LinkUser(n.Config.BasePath, n.Config.TargetPath, username, token)
	_http.RespondWithJSON(w, http.StatusOK, "authentication successfull")
}

func (n *AppNode) unauthenticateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}
	// get session for user
	s := session.GetSessionByUsername(n.Sessions, username)
	if s == nil || (s == &models.Session{}) || s.Token == "" {
		log.WithFields(log.Fields{
			"username": username,
		}).Error("cannot unlink user - no session found")
		_http.RespondWithJSON(w, http.StatusUnauthorized, "no session found for user")
		return
	}
	// unlink the user
	if status, msg := handler.UnlinkUser(n.Config.TargetPath, s.Token); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	// remove session
	n.Sessions = session.RemoveSession(n.Sessions, s.Token)
	// everything went well
	_http.RespondWithJSON(w, http.StatusOK, "unauthenticated user")
}
