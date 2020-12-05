package gateway

import (
	"encoding/json"
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetUserGroupAPI handles possible errors during the select and writes Responses
func (g *AppGateway) GetUserGroupAPI(w http.ResponseWriter, db *mongo.Database, ug *models.UserGroup) int {
	// try to select user
	if err := ug.GetUserGroup(db, g.GetUserPermission(w, false)); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			_http.RespondWithError(w, http.StatusNotFound, "Usergroup not found")
		default:
			// another error occured
			_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return 1
	}
	return 0
}

// DecodeUserGroupRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUserGroupRequest(w http.ResponseWriter, r *http.Request, ug models.UserGroup) (models.UserGroup, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ug); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return models.UserGroup{}, 1
	}
	defer r.Body.Close()
	return ug, 0
}

// DecodeUserGroupsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUserGroupsRequest(w http.ResponseWriter, r *http.Request, ugs []models.UserGroup) ([]models.UserGroup, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ugs); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return ugs, 0
}
