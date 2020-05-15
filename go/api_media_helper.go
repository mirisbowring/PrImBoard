package primboard

import (
	"encoding/json"
	"net/http"
)

// DecodeMediaRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediaRequest(w http.ResponseWriter, r *http.Request, m Media) (Media, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return Media{}, 1
	}
	defer r.Body.Close()
	return m, 0
}

// DecodeMediasRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediasRequest(w http.ResponseWriter, r *http.Request) ([]Media, int) {
	var m []Media
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return m, 0
}
