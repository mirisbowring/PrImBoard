package primboard

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/PrImBoard/helper/http"
)

// DecodeMediaEventMapRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediaEventMapRequest(w http.ResponseWriter, r *http.Request) (MediaEventMap, int) {
	var mem MediaEventMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mem); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return mem, 1
	}
	defer r.Body.Close()
	return mem, 0
}

// DecodeTagEventMapRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeTagEventMapRequest(w http.ResponseWriter, r *http.Request) (TagEventMap, int) {
	var tem TagEventMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tem); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return tem, 1
	}
	defer r.Body.Close()
	return tem, 0
}
