package http

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParsePrimitiveID parses the id from the route and returns it
// returns primitive.NilObjectID if an error occured
// sends a respond if an error occured
func ParsePrimitiveID(w http.ResponseWriter, r *http.Request) primitive.ObjectID {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Could not parse ID from route!")
		return primitive.NilObjectID
	}
	return id
}

// ParsePathString parses the string value from the route and returns it
// stats 0 -> ok || status 1 -> error
func ParsePathString(w http.ResponseWriter, r *http.Request, key string) (string, int) {
	if val := mux.Vars(r)[key]; val == "" {
		log.Warnf("key <%s> was not specified", key)
		RespondWithError(w, http.StatusBadRequest, "key was not specified")
		return val, 1
	} else {
		return val, 0
	}
}

// RespondWithError Creates an error payload and adds the error message to be
// returned
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON parses the passed payload and returns it with the specified
// code to the client
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	//	enableCors(&w)
	response, _ := json.Marshal(payload)
	// delete the temporary user key from header
	w.Header().Del("user")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
