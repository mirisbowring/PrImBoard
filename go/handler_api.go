package primboard

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
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

// parseUsername parses the username from the route and returns is
// stats 0 -> ok || status 1 -> error
func parseUsername(w http.ResponseWriter, r *http.Request) (User, int) {
	vars := mux.Vars(r)
	user := User{Username: vars["username"]}
	if user.Username == "" {
		RespondWithError(w, http.StatusBadRequest, "User was not specified!")
		return user, 1
	}
	return user, 0
}

// parseToken parses the registration token from the route and returns it
// status 0 -> ok || status 1 -> error
func parseToken(w http.ResponseWriter, r *http.Request) (Invite, int) {
	vars := mux.Vars(r)
	invite := Invite{Token: vars["token"]}
	if invite.Token == "" {
		RespondWithError(w, http.StatusBadRequest, "Registrationtoken not specified!")
		return invite, 1
	}
	return invite, 0
}

// getPermission parses the permissionfilter and returns it
func getPermission(w http.ResponseWriter) bson.M {
	username := w.Header().Get("user")
	session := GetSessionByUsername(username)
	return CreatePermissionFilter(session.Usergroups, username)
}
