package gateway

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
)

// DecodeIPFSNodeRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeIPFSNodeRequest(w http.ResponseWriter, r *http.Request, ipfs models.IPFSNode) (models.IPFSNode, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ipfs); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return models.IPFSNode{}, 1
	}
	defer r.Body.Close()
	return ipfs, 0
}

// DecodeUserRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
// func DecodeUserRequest(w http.ResponseWriter, r *http.Request, u models.User) (models.User, int) {
// 	decoder := json.NewDecoder(r.Body)
// 	if err := decoder.Decode(&u); err != nil {
// 		// an decode error occured
// 		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
// 		return models.User{}, 1
// 	}
// 	defer r.Body.Close()
// 	return u, 0
// }

// DecodeStringsRequest decodes the api request into the passed user slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeStringsRequest(w http.ResponseWriter, r *http.Request, u []string) ([]string, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return u, 0
}

// ParseUserSelect tries to retrieves the requested user from the database.
// writes error responses into ResponseWriter
// status 0 => ok || status 1 => error
// func ParseUserSelect(db *mongo.Database, w http.ResponseWriter, r *http.Request, verifyUser bool) (models.User, int) {
// 	username, status := _http.ParsePathString(w, r, "username")
// 	if status > 0 {
// 		return models.User{}, status
// 	}
// 	u := models.User{Username: username}
// 	// verify that actual user matched requested user path
// 	if verifyUser && u.Username != _http.GetUsernameFromHeader(w) {
// 		_http.RespondWithError(w, http.StatusUnauthorized, "You are not allowed to request information for that user.")
// 		return models.User{}, 1
// 	}
// 	// try to select model
// 	if err := u.GetUser(db); err != nil {
// 		switch err {
// 		case mongo.ErrNoDocuments:
// 			// model not found
// 			_http.RespondWithError(w, http.StatusNotFound, "User not found")
// 		default:
// 			// another error occured
// 			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
// 		}
// 		return models.User{}, 1
// 	}
// 	return u, 0
// }
