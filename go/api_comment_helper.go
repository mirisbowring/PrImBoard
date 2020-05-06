package primboard

import (
	"encoding/json"
	"net/http"
)

// DecodeCommentRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeCommentRequest(w http.ResponseWriter, r *http.Request, c Comment) (Comment, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return Comment{}, 1
	}
	defer r.Body.Close()
	return c, 0
}

// DecodeCommentsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeCommentsRequest(w http.ResponseWriter, r *http.Request, c []Comment) ([]Comment, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return c, 0
}
