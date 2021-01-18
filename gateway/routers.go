package gateway

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// initializeRoutes initializes all the available webroutes
func (g *AppGateway) initializeRoutes() {
	g.Router = mux.NewRouter().StrictSlash(true)
	// index
	g.Router.HandleFunc("/api/v1/", g.index).Methods("GET")
	g.Router.HandleFunc("/api/v2/", g.index).Methods("GET")
	// event
	g.Router.Handle("/api/v1/event", g.Authenticate(http.HandlerFunc(g.AddEvent), false)).Methods("POST")
	g.Router.Handle("/api/v1/event/{id}", g.Authenticate(http.HandlerFunc(g.DeleteEventByID), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/event/{id}", g.Authenticate(http.HandlerFunc(g.GetEventByID), false)).Methods("GET")
	g.Router.Handle("/api/v1/event/{id}", g.Authenticate(http.HandlerFunc(g.UpdateEventByID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/events", g.Authenticate(http.HandlerFunc(g.GetEvents), false)).Methods("GET")
	g.Router.Handle("/api/v1/events/{title}", g.Authenticate(http.HandlerFunc(g.GetEventsByName), false)).Methods("GET")
	g.Router.Handle("/api/v1/events/maptags", g.Authenticate(http.HandlerFunc(g.MapTagsToEvents), false)).Methods("GET")
	// media
	g.Router.Handle("/api/v1/media", g.Authenticate(http.HandlerFunc(g.GetMedia), false)).Methods("GET")
	g.Router.Handle("/api/v1/media", g.Authenticate(http.HandlerFunc(g.AddMedia), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/remove", g.Authenticate(http.HandlerFunc(g.deleteMediaByIDs), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/upload", g.Authenticate(http.HandlerFunc(g.UploadMedia), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/byids", g.Authenticate(http.HandlerFunc(g.GetMediaByIDs), false)).Methods("GET")
	g.Router.Handle("/api/v1/media/maptags", g.Authenticate(http.HandlerFunc(g.MapTagsToMedia), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/mapevents", g.Authenticate(http.HandlerFunc(g.MapEventsToMedia), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/addgroups", g.Authenticate(http.HandlerFunc(g.MapGroupsToMedia), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/removegroups", g.Authenticate(http.HandlerFunc(g.removeGroupsFromMedias), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/{id}/groups/{group}", g.Authenticate(http.HandlerFunc(g.removeGroupFromMedia), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/media/{id}", g.Authenticate(http.HandlerFunc(g.DeleteMediaByID), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/media/{id}/{node}", g.Authenticate(http.HandlerFunc(g.deleteMediaByIDFromNode), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/media/{id}", g.Authenticate(http.HandlerFunc(g.GetMediaByID), false)).Methods("GET")
	// a.Router.Handle("/api/v1/media/{id}", a.Authenticate(http.HandlerFunc(a.UpdateMediaByID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/media/{id}/comment", g.Authenticate(http.HandlerFunc(g.AddCommentByMediaID), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/{id}/description", g.Authenticate(http.HandlerFunc(g.AddDescriptionByMediaID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/media/{id}/tag", g.Authenticate(http.HandlerFunc(g.AddTagByMediaID), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/{id}/tags", g.Authenticate(http.HandlerFunc(g.AddTagsByMediaID), false)).Methods("POST")
	// g.Router.Handle("/api/v1/media/{id}/usergroups", g.Authenticate(http.HandlerFunc(g.AddUserGroupsByMediaID), false)).Methods("POST")
	g.Router.Handle("/api/v1/media/{id}/timestamp", g.Authenticate(http.HandlerFunc(g.AddTimestampByMediaID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/media/{id}/title", g.Authenticate(http.HandlerFunc(g.AddTitleByMediaID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/mediaByHash/{ipfs_id}", g.Authenticate(http.HandlerFunc(g.UpdateMediaByHash), false)).Methods("PUT")
	// tag
	g.Router.Handle("/api/v1/tag", g.Authenticate(http.HandlerFunc(g.AddTag), false)).Methods("POST")
	g.Router.Handle("/api/v1/tag/{id}", g.Authenticate(http.HandlerFunc(g.DeleteTagByID), false)).Methods("POST")
	g.Router.Handle("/api/v1/tag/{id}", g.Authenticate(http.HandlerFunc(g.GetTagByID), false)).Methods("GET")
	g.Router.Handle("/api/v1/tag/{id}", g.Authenticate(http.HandlerFunc(g.UpdateTagByID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/tags", g.Authenticate(http.HandlerFunc(g.GetTags), false)).Methods("GET")
	g.Router.Handle("/api/v1/tags/{name}", g.Authenticate(http.HandlerFunc(g.GetTagsByName), false)).Methods("GET")
	// user
	// g.Router.HandleFunc("/api/v1/user", g.CreateUser).Methods("POST")
	g.Router.Handle("/api/v1/user/invite", g.Authenticate(http.HandlerFunc(g.GenerateInvite), false)).Methods("GET")
	g.Router.Handle("/api/v1/user/node", g.Authenticate(http.HandlerFunc(g.AddNode), false)).Methods("POST")
	g.Router.Handle("/api/v1/user/node/{id}", g.Authenticate(http.HandlerFunc(g.DeleteNodeByID), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/user/node/{id}", g.Authenticate(http.HandlerFunc(g.GetNodeByID), false)).Methods("GET")
	g.Router.Handle("/api/v1/user/node/{id}", g.Authenticate(http.HandlerFunc(g.UpdateNodeByID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/user/node/{id}", g.Authenticate(http.HandlerFunc(g.addGroupsToNode), false)).Methods("POST").Queries("groups", "{groups}")
	g.Router.Handle("/api/v1/user/nodes", g.Authenticate(http.HandlerFunc(g.GetNodes), false)).Methods("GET")
	// g.Router.Handle("/api/v1/user/nodes/removegroups", g.Authenticate(http.HandlerFunc(g.MapGroupsToMedia), false)).Methods("POST")
	// g.Router.Handle("/api/v1/user/{username}", g.Authenticate(http.HandlerFunc(g.DeleteUser), false)).Methods("DELETE")
	// g.Router.Handle("/api/v1/user/{username}", g.Authenticate(http.HandlerFunc(g.GetUserByUsername), false)).Methods("GET")
	// g.Router.Handle("/api/v1/user/{username}/settings", g.Authenticate(http.HandlerFunc(g.GetSettingsByUsername), false)).Methods("GET")
	// // g.Router.HandleFunc("/api/v1/login", g.LoginUser).Methods("POST")
	// g.Router.Handle("/api/v1/logout", g.Authenticate(http.HandlerFunc(g.LogoutUser), true)).Methods("POST")
	// g.Router.Handle("/api/v1/user/{username}", g.Authenticate(http.HandlerFunc(g.UpdateUser), false)).Methods("PUT")
	// usergroup
	g.Router.Handle("/api/v1/usergroup", g.Authenticate(http.HandlerFunc(g.AddUserGroup), false)).Methods("POST")
	g.Router.Handle("/api/v1/usergroups", g.Authenticate(http.HandlerFunc(g.GetUserGroups), false)).Methods("GET")
	g.Router.Handle("/api/v1/usergroups/{name}", g.Authenticate(http.HandlerFunc(g.GetUserGroupsByName), false)).Methods("GET")
	g.Router.Handle("/api/v1/usergroup/{id}", g.Authenticate(http.HandlerFunc(g.DeleteUserGroupByID), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/usergroup/{id}", g.Authenticate(http.HandlerFunc(g.GetUserGroupByID), false)).Methods("GET")
	g.Router.Handle("/api/v1/usergroup/{id}", g.Authenticate(http.HandlerFunc(g.UpdateUserGroupByID), false)).Methods("PUT")
	g.Router.Handle("/api/v1/usergroup/{id}/user/{username}", g.Authenticate(http.HandlerFunc(g.RemoveUserFromUserGroupByID), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/usergroup/{id}/user/{username}", g.Authenticate(http.HandlerFunc(g.AddUserToUserGroupByID), false)).Methods("POST")
	g.Router.Handle("/api/v1/usergroup/{id}/users", g.Authenticate(http.HandlerFunc(g.RemoveUsersFromUserGroupByID), false)).Methods("DELETE")
	g.Router.Handle("/api/v1/usergroup/{id}/users", g.Authenticate(http.HandlerFunc(g.AddUsersToUserGroupByID), false)).Methods("POST")
	// infrastructure
	g.Router.Handle("/api/v2/infrastructure/node/authenticate", g.Authenticate(http.HandlerFunc(g.authenticateNode), false)).Methods("POST")
	g.Router.Handle("/api/v2/infrastructure/node/register", g.Authenticate(http.HandlerFunc(g.registerNode), false)).Methods("GET")
	g.Router.Handle("/api/v2/infrastructure/node/{id}/secret", g.Authenticate(http.HandlerFunc(g.retrieveNodeSecret), false)).Methods("GET")
	g.Router.Handle("/api/v2/infrastructure/node/{id}/secret/refresh", g.Authenticate(http.HandlerFunc(g.refreshNodeSecret), false)).Methods("GET").Queries("return", "{return}")
	g.Router.Handle("/api/v2/infrastructure/node/{id}/structure", g.Authenticate(http.HandlerFunc(g.parseNodeStructure), false)).Methods("GET").Queries("page", "{page}", "size", "{size}")
}

// Index controller
func (g *AppGateway) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
