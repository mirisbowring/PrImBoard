package primboard

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"go.mongodb.org/mongo-driver/mongo"
)

// DecodeIPFSNodeRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeIPFSNodeRequest(w http.ResponseWriter, r *http.Request, ipfs IPFSNode) (IPFSNode, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ipfs); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return IPFSNode{}, 1
	}
	defer r.Body.Close()
	return ipfs, 0
}

// DecodeUserRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUserRequest(w http.ResponseWriter, r *http.Request, u User) (User, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return User{}, 1
	}
	defer r.Body.Close()
	return u, 0
}

// DecodeUsersRequest decodes the api request into the passed user slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUsersRequest(w http.ResponseWriter, r *http.Request, u []User) ([]User, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return u, 0
}

// parseUserSelect tries to retrieves the requested user from the database.
// writes error responses into ResponseWriter
// status 0 => ok || status 1 => error
func (a *App) parseUserSelect(w http.ResponseWriter, r *http.Request, verifyUser bool) (User, int) {
	// parseUsername from url
	u, status := parseUsername(w, r)
	if status > 0 {
		// cancel request
		return User{}, status
	}
	// verify that actual user matched requested user path
	if verifyUser && u.Username != w.Header().Get("user") {
		_http.RespondWithError(w, http.StatusUnauthorized, "You are not allowed to request information for that user.")
		return User{}, 1
	}
	// try to select model
	if err := u.GetUser(a.DB); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return User{}, 1
	}
	return u, 0
}
