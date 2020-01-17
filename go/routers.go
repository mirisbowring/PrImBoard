package primboard

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// InitializeRoutes initializes all the available webroutes
func (a *App) InitializeRoutes() {
	a.Router = mux.NewRouter().StrictSlash(true)
	// index
	a.Router.HandleFunc("/api/v1/", a.Index).Methods("GET")
	// event
	a.Router.Handle("/api/v1/event", Authenticate(http.HandlerFunc(a.AddEvent), false)).Methods("POST")
	a.Router.Handle("/api/v1/event/{id}", Authenticate(http.HandlerFunc(a.DeleteEventByID), false)).Methods("DELETE")
	a.Router.Handle("/api/v1/event/{id}", Authenticate(http.HandlerFunc(a.GetEventByID), false)).Methods("GET")
	a.Router.Handle("/api/v1/event/{id}", Authenticate(http.HandlerFunc(a.UpdateEventByID), false)).Methods("PUT")
	a.Router.Handle("/api/v1/events", Authenticate(http.HandlerFunc(a.GetEvents), false)).Methods("GET")
	// media
	a.Router.Handle("/api/v1/media", Authenticate(http.HandlerFunc(a.GetMedia), false)).Methods("GET")
	a.Router.Handle("/api/v1/media", Authenticate(http.HandlerFunc(a.AddMedia), false)).Methods("POST")
	a.Router.Handle("/api/v1/media/{id}", Authenticate(http.HandlerFunc(a.DeleteMediaByID), false)).Methods("DELETE")
	a.Router.Handle("/api/v1/media/{id}", Authenticate(http.HandlerFunc(a.GetMediaByID), false)).Methods("GET")
	a.Router.Handle("/api/v1/media/{ipfs_id}", Authenticate(http.HandlerFunc(a.UpdateMediaByID), false)).Methods("PUT")
	a.Router.Handle("/api/v1/mediaByHash/{ipfs_id}", Authenticate(http.HandlerFunc(a.GetMediaByHash), false)).Methods("GET")
	// tag
	a.Router.Handle("/api/v1/tag", Authenticate(http.HandlerFunc(a.AddTag), false)).Methods("POST")
	a.Router.Handle("/api/v1/tag/{id}", Authenticate(http.HandlerFunc(a.DeleteTagByID), false)).Methods("POST")
	a.Router.Handle("/api/v1/tag/{id}", Authenticate(http.HandlerFunc(a.GetTagByID), false)).Methods("GET")
	a.Router.Handle("/api/v1/tag/{id}", Authenticate(http.HandlerFunc(a.UpdateTagByID), false)).Methods("PUT")
	a.Router.Handle("/api/v1/tags", Authenticate(http.HandlerFunc(a.GetTags), false)).Methods("GET")
	a.Router.Handle("/api/v1/tags/{name}", Authenticate(http.HandlerFunc(a.GetTagsByName), false)).Methods("GET")
	// user
	a.Router.Handle("/api/v1/user", Authenticate(http.HandlerFunc(a.CreateUser), false)).Methods("POST")
	a.Router.Handle("/api/v1/user/{username}", Authenticate(http.HandlerFunc(a.DeleteUser), false)).Methods("DELETE")
	a.Router.Handle("/api/v1/user/{username}", Authenticate(http.HandlerFunc(a.GetUserByUsername), false)).Methods("GET")
	a.Router.HandleFunc("/api/v1/login", a.LoginUser).Methods("POST")
	a.Router.Handle("/api/v1/logout", Authenticate(http.HandlerFunc(a.LogoutUser), true)).Methods("POST")
	a.Router.Handle("/api/v1/user/{username}", Authenticate(http.HandlerFunc(a.UpdateUser), false)).Methods("PUT")
	// usergroup
	a.Router.Handle("/api/v1/usergroup", Authenticate(http.HandlerFunc(a.AddUserGroup), false)).Methods("POST")
	a.Router.Handle("/api/v1/usergroup/{id}", Authenticate(http.HandlerFunc(a.DeleteUserGroupByID), false)).Methods("DELETE")
	a.Router.Handle("/api/v1/usergroup/{id}", Authenticate(http.HandlerFunc(a.GetUserGroupByID), false)).Methods("GET")
	a.Router.Handle("/api/v1/usergroup/{id}", Authenticate(http.HandlerFunc(a.UpdateUserGroupByID), false)).Methods("PUT")
}

// Index controller
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
