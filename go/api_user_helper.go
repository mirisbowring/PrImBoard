package primboard

import (
	"encoding/json"
	"net/http"
)

// DecodeUserRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUserRequest(w http.ResponseWriter, r *http.Request, u User) (User, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
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
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return u, 0
}