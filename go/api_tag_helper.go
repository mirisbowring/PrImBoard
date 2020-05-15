package primboard

import (
	"encoding/json"
	"net/http"
)

// DecodeTagRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagRequest(w http.ResponseWriter, r *http.Request, t Tag) (Tag, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return Tag{}, 1
	}
	defer r.Body.Close()
	return t, 0
}

// DecodeTagsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagsRequest(w http.ResponseWriter, r *http.Request, t []Tag) ([]Tag, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return t, 0
}

// DecodeTagMediaMapRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagMediaMapRequest(w http.ResponseWriter, r *http.Request) (TagMediaMap, int) {
	var tmm TagMediaMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tmm); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return tmm, 1
	}
	defer r.Body.Close()
	return tmm, 0
}
