package primboard

import (
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
