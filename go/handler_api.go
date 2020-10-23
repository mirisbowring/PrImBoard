package primboard

import (
	"errors"
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
	username := getUsernameFromHeader(w)
	session := GetSessionByUsername(username)
	return createPermissionFilter(session.Usergroups, username)
}

// CreatePermissionFilter creates a filter bson that matches the owner and it's groups
func createPermissionFilter(groups []primitive.ObjectID, user string) bson.M {
	filters := []bson.M{}
	// username must be passed
	if user == "" {
		return bson.M{}
	}
	filters = append(filters, bson.M{"creator": user})
	// add groups if passed
	if groups != nil && len(groups) > 0 {
		filters = append(filters, bson.M{"groupIDs": bson.M{"$in": groups}})
	}

	return bson.M{"$or": filters}
}

func createMatcherProjectPipeline(matcher bson.M, project bson.M) []primitive.M {
	// create pipeline
	pipeline := []bson.M{
		{"$match": matcher},
		{"$project": project},
	}
	return pipeline
}

func createPermissionMatcher(permission bson.M, id primitive.ObjectID) (bson.M, error) {
	// verify that permission bson was specified
	if permission == nil {
		return nil, errors.New("no permissions specified")
	}
	// create matcher
	var matcher bson.M
	if id == primitive.NilObjectID {
		matcher = bson.M{"$and": []bson.M{
			permission,
		}}
	} else {
		matcher = bson.M{"$and": []bson.M{
			{"_id": id},
			permission,
		}}
	}
	return matcher, nil
}

func createPermissionProjectPipeline(permission bson.M, id primitive.ObjectID, project bson.M) ([]primitive.M, error) {
	matcher, err := createPermissionMatcher(permission, id)
	if err != nil {
		return nil, err
	}
	return createMatcherProjectPipeline(matcher, project), nil
}

func getUsernameFromHeader(w http.ResponseWriter) string {
	return w.Header().Get("user")
}
