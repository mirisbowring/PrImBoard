package primboard

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// parseID parses the id from the route and returns it
// returns primitive.NilObjectID if an error occured
// sends a respond if an error occured
func parseID(w http.ResponseWriter, r *http.Request) primitive.ObjectID {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Could not parse ID from route!")
		return primitive.NilObjectID
	}
	return id
}
