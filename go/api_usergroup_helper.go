package primboard

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

// GetUserGroupAPI handles possible errors during the select and writes Responses
func (ug *UserGroup) GetUserGroupAPI(w http.ResponseWriter, db *mongo.Database) int {
	// try to select user
	if err := ug.GetUserGroup(db); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			// model not found
			RespondWithError(w, http.StatusNotFound, "Usergroup not found")
		default:
			// another error occured
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return 1
	}
	return 0
}

// DecodeUserGroupRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUserGroupRequest(w http.ResponseWriter, r *http.Request, ug UserGroup) (UserGroup, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ug); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return UserGroup{}, 1
	}
	defer r.Body.Close()
	return ug, 0
}

// DecodeUserGroupsRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeUserGroupsRequest(w http.ResponseWriter, r *http.Request, ugs []UserGroup) ([]UserGroup, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ugs); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return ugs, 0
}
