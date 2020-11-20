package gateway

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
)

// DecodeCommentRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeCommentRequest(w http.ResponseWriter, r *http.Request, c models.Comment) (models.Comment, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return models.Comment{}, 1
	}
	defer r.Body.Close()
	return c, 0
}

// DecodeCommentsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeCommentsRequest(w http.ResponseWriter, r *http.Request, c []models.Comment) ([]models.Comment, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return c, 0
}
