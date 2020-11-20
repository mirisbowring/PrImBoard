package gateway

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
)

// DecodeTagRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagRequest(w http.ResponseWriter, r *http.Request, t models.Tag) (models.Tag, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return models.Tag{}, 1
	}
	defer r.Body.Close()
	return t, 0
}

// DecodeTagsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagsRequest(w http.ResponseWriter, r *http.Request, t []models.Tag) ([]models.Tag, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return t, 0
}

// DecodeTagStringRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagStringRequest(w http.ResponseWriter, r *http.Request, t string) (string, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return "", 1
	}
	defer r.Body.Close()
	return t, 0
}

// DecodeTagStringsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagStringsRequest(w http.ResponseWriter, r *http.Request, t []string) ([]string, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return t, 0
}

// DecodeTagMediaMapRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagMediaMapRequest(w http.ResponseWriter, r *http.Request) (models.TagMediaMap, int) {
	var tmm models.TagMediaMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tmm); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return tmm, 1
	}
	defer r.Body.Close()
	return tmm, 0
}
