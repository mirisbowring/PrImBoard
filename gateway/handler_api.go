package gateway

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// parseID parses the id from the route and returns it
// returns primitive.NilObjectID if an error occured
// sends a respond if an error occured
func parseID(w http.ResponseWriter, r *http.Request) primitive.ObjectID {
	return parseIDCustomKey(w, r, "id")
}

func parseIDCustomKey(w http.ResponseWriter, r *http.Request, key string) primitive.ObjectID {
	vars := mux.Vars(r)
	id := helper.ParsePrimitiveID(vars[key])
	if id.IsZero() {
		_http.RespondWithError(w, http.StatusBadRequest, "Could not parse ID from route!")
	}
	return id
}

// // parseUsername parses the username from the route and returns is
// // stats 0 -> ok || status 1 -> error
// func parseUsername(w http.ResponseWriter, r *http.Request) (User, int) {
// 	vars := mux.Vars(r)
// 	user := User{Username: vars["username"]}
// 	if user.Username == "" {
// 		_http.RespondWithError(w, http.StatusBadRequest, "User was not specified!")
// 		return user, 1
// 	}
// 	return user, 0
// }

// // parseToken parses the registration token from the route and returns it
// // status 0 -> ok || status 1 -> error
// func parseToken(w http.ResponseWriter, r *http.Request) (Invite, int) {
// 	vars := mux.Vars(r)
// 	invite := Invite{Token: vars["token"]}
// 	if invite.Token == "" {
// 		_http.RespondWithError(w, http.StatusBadRequest, "Registrationtoken not specified!")
// 		return invite, 1
// 	}
// 	return invite, 0
// }
