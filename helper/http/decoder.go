package http

import (
	"encoding/json"
	"net/http"

	"github.com/mirisbowring/primboard/models"
	"github.com/mirisbowring/primboard/models/maps"
)

// DecodeFilesGroupsMapRequest decodes the api request into a slice responds
// with decode error if occurs status
//
// 0 => ok || status 1 => error
func DecodeFilesGroupsMapRequest(w http.ResponseWriter, r *http.Request) (maps.FilesGroupsMap, int) {
	var maps maps.FilesGroupsMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&maps); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return maps, 1
	}
	defer r.Body.Close()
	return maps, 0
}

// DecodeStringsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeStringsRequest(w http.ResponseWriter, r *http.Request, t []string) ([]string, int) {
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
func DecodeTagMediaMapRequest(w http.ResponseWriter, r *http.Request) (models.TagMediaMap, int) {
	var tmm models.TagMediaMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tmm); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return tmm, 1
	}
	defer r.Body.Close()
	return tmm, 0
}
