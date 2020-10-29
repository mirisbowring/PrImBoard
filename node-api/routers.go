package node

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// InitializeRoutes initializes all the available webroutes
func (a *App) InitializeRoutes() {
	log.Debug("initializing routes")
	a.Router = mux.NewRouter().StrictSlash(true)
	// index
	a.Router.HandleFunc("/api/v1/", a.Index).Methods("GET")
	// files
	a.Router.Handle("/api/v1/file", a.Authenticate(http.HandlerFunc(a.addFile), false)).Methods("POST")
	a.Router.Handle("/api/v1/file/{filename}", a.Authenticate(http.HandlerFunc(a.deleteFile), false)).Methods("DELETE")
	// event
	// a.Router.Handle("/api/v1/event", a.Authenticate(http.HandlerFunc(a.AddEvent), false)).Methods("POST")
	// a.Router.Handle("/api/v1/event/{id}", a.Authenticate(http.HandlerFunc(a.DeleteEventByID), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/event/{id}", a.Authenticate(http.HandlerFunc(a.GetEventByID), false)).Methods("GET")
	// a.Router.Handle("/api/v1/event/{id}", a.Authenticate(http.HandlerFunc(a.UpdateEventByID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/events", a.Authenticate(http.HandlerFunc(a.GetEvents), false)).Methods("GET")
	// a.Router.Handle("/api/v1/events/{title}", a.Authenticate(http.HandlerFunc(a.GetEventsByName), false)).Methods("GET")
	// a.Router.Handle("/api/v1/events/maptags", a.Authenticate(http.HandlerFunc(a.MapTagsToEvents), false)).Methods("GET")
	// // media
	// a.Router.Handle("/api/v1/media", a.Authenticate(http.HandlerFunc(a.GetMedia), false)).Methods("GET")
	// a.Router.Handle("/api/v1/media", a.Authenticate(http.HandlerFunc(a.AddMedia), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/upload", a.Authenticate(http.HandlerFunc(a.UploadMedia), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/byids", a.Authenticate(http.HandlerFunc(a.GetMediaByIDs), false)).Methods("GET")
	// a.Router.Handle("/api/v1/media/maptags", a.Authenticate(http.HandlerFunc(a.MapTagsToMedia), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/mapevents", a.Authenticate(http.HandlerFunc(a.MapEventsToMedia), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/mapgroups", a.Authenticate(http.HandlerFunc(a.MapGroupsToMedia), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/{id}", a.Authenticate(http.HandlerFunc(a.DeleteMediaByID), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/media/{id}", a.Authenticate(http.HandlerFunc(a.GetMediaByID), false)).Methods("GET")
	// // a.Router.Handle("/api/v1/media/{id}", a.Authenticate(http.HandlerFunc(a.UpdateMediaByID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/media/{id}/comment", a.Authenticate(http.HandlerFunc(a.AddCommentByMediaID), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/{id}/description", a.Authenticate(http.HandlerFunc(a.AddDescriptionByMediaID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/media/{id}/tag", a.Authenticate(http.HandlerFunc(a.AddTagByMediaID), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/{id}/tags", a.Authenticate(http.HandlerFunc(a.AddTagsByMediaID), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/{id}/usergroups", a.Authenticate(http.HandlerFunc(a.AddUserGroupsByMediaID), false)).Methods("POST")
	// a.Router.Handle("/api/v1/media/{id}/timestamp", a.Authenticate(http.HandlerFunc(a.AddTimestampByMediaID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/media/{id}/title", a.Authenticate(http.HandlerFunc(a.AddTitleByMediaID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/mediaByHash/{ipfs_id}", a.Authenticate(http.HandlerFunc(a.UpdateMediaByHash), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/mediaByHash/{ipfs_id}", a.Authenticate(http.HandlerFunc(a.GetMediaByHash), false)).Methods("GET")
	// // tag
	// a.Router.Handle("/api/v1/tag", a.Authenticate(http.HandlerFunc(a.AddTag), false)).Methods("POST")
	// a.Router.Handle("/api/v1/tag/{id}", a.Authenticate(http.HandlerFunc(a.DeleteTagByID), false)).Methods("POST")
	// a.Router.Handle("/api/v1/tag/{id}", a.Authenticate(http.HandlerFunc(a.GetTagByID), false)).Methods("GET")
	// a.Router.Handle("/api/v1/tag/{id}", a.Authenticate(http.HandlerFunc(a.UpdateTagByID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/tags", a.Authenticate(http.HandlerFunc(a.GetTags), false)).Methods("GET")
	// a.Router.Handle("/api/v1/tags/{name}", a.Authenticate(http.HandlerFunc(a.GetTagsByName), false)).Methods("GET")
	// // user
	// a.Router.HandleFunc("/api/v1/user", a.CreateUser).Methods("POST")
	// a.Router.Handle("/api/v1/user/invite", a.Authenticate(http.HandlerFunc(a.GenerateInvite), false)).Methods("GET")
	// a.Router.Handle("/api/v1/user/node", a.Authenticate(http.HandlerFunc(a.AddNode), false)).Methods("POST")
	// a.Router.Handle("/api/v1/user/node/{id}", a.Authenticate(http.HandlerFunc(a.DeleteNodeByID), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/user/node/{id}", a.Authenticate(http.HandlerFunc(a.GetNodeByID), false)).Methods("GET")
	// a.Router.Handle("/api/v1/user/node/{id}", a.Authenticate(http.HandlerFunc(a.UpdateNodeByID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/user/nodes", a.Authenticate(http.HandlerFunc(a.GetNodes), false)).Methods("GET")
	// a.Router.Handle("/api/v1/user/{username}", a.Authenticate(http.HandlerFunc(a.DeleteUser), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/user/{username}", a.Authenticate(http.HandlerFunc(a.GetUserByUsername), false)).Methods("GET")
	// a.Router.Handle("/api/v1/user/{username}/settings", a.Authenticate(http.HandlerFunc(a.GetSettingsByUsername), false)).Methods("GET")
	// a.Router.HandleFunc("/api/v1/login", a.LoginUser).Methods("POST")
	// a.Router.Handle("/api/v1/logout", a.Authenticate(http.HandlerFunc(a.LogoutUser), true)).Methods("POST")
	// a.Router.Handle("/api/v1/user/{username}", a.Authenticate(http.HandlerFunc(a.UpdateUser), false)).Methods("PUT")
	// // usergroup
	// a.Router.Handle("/api/v1/usergroup", a.Authenticate(http.HandlerFunc(a.AddUserGroup), false)).Methods("POST")
	// a.Router.Handle("/api/v1/usergroups", a.Authenticate(http.HandlerFunc(a.GetUserGroups), false)).Methods("GET")
	// a.Router.Handle("/api/v1/usergroups/{name}", a.Authenticate(http.HandlerFunc(a.GetUserGroupsByName), false)).Methods("GET")
	// a.Router.Handle("/api/v1/usergroup/{id}", a.Authenticate(http.HandlerFunc(a.DeleteUserGroupByID), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/usergroup/{id}", a.Authenticate(http.HandlerFunc(a.GetUserGroupByID), false)).Methods("GET")
	// a.Router.Handle("/api/v1/usergroup/{id}", a.Authenticate(http.HandlerFunc(a.UpdateUserGroupByID), false)).Methods("PUT")
	// a.Router.Handle("/api/v1/usergroup/{id}/user/{username}", a.Authenticate(http.HandlerFunc(a.RemoveUserFromUserGroupByID), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/usergroup/{id}/user", a.Authenticate(http.HandlerFunc(a.AddUserToUserGroupByID), false)).Methods("POST")
	// a.Router.Handle("/api/v1/usergroup/{id}/users", a.Authenticate(http.HandlerFunc(a.RemoveUsersFromUserGroupByID), false)).Methods("DELETE")
	// a.Router.Handle("/api/v1/usergroup/{id}/users", a.Authenticate(http.HandlerFunc(a.AddUsersToUserGroupByID), false)).Methods("POST")
}

// Index controller
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
