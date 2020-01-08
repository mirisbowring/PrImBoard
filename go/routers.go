package swagger

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
	a.Router.HandleFunc("/api/v1/event", Authenticate(http.HandlerFunc(a.AddEvent))).Methods("POST")
	a.Router.HandleFunc("/api/v1/event/{id}", Authenticate(http.HandlerFunc(a.DeleteEventByID))).Methods("DELETE")
	a.Router.HandleFunc("/api/v1/event/{id}", Authenticate(http.HandlerFunc(a.GetEventByID))).Methods("GET")
	a.Router.HandleFunc("/api/v1/event/{id}", Authenticate(http.HandlerFunc(a.UpdateEventByID))).Methods("PUT")
	// media
	a.Router.Handle("/api/v1/media", Authenticate(http.HandlerFunc(a.GetMedia))).Methods("GET")
	a.Router.Handle("/api/v1/media", Authenticate(http.HandlerFunc(a.AddMedia))).Methods("POST")
	a.Router.Handle("/api/v1/media/{id}", Authenticate(http.HandlerFunc(a.DeleteMediaByID))).Methods("DELETE")
	a.Router.Handle("/api/v1/media/{id}", Authenticate(http.HandlerFunc(a.GetMediaByID))).Methods("GET")
	a.Router.Handle("/api/v1/media/{id}", Authenticate(http.HandlerFunc(a.UpdateMediaByID))).Methods("PUT")
	a.Router.Handle("/api/v1/mediaByHash/{ipfs_id}", Authenticate(http.HandlerFunc(a.GetMediaByHash))).Methods("GET")
	// tag
	a.Router.Handle("/api/v1/tag", Authenticate(http.HandlerFunc(a.AddTag).Methods("POST")
	a.Router.Handle("/api/v1/tag/{id}", Authenticate(http.HandlerFunc(a.DeleteTagByID).Methods("POST")
	a.Router.Handle("/api/v1/tag/{id}", Authenticate(http.HandlerFunc(a.GetTagByID).Methods("GET")
	a.Router.Handle("/api/v1/tag/{id}", Authenticate(http.HandlerFunc(a.UpdateTagByID).Methods("PUT")
	// user
	a.Router.Handle("/api/v1/user", Authenticate(http.HandlerFunc(a.CreateUser))).Methods("POST")
	a.Router.Handle("/api/v1/user/{username}", Authenticate(http.HandlerFunc(a.DeleteUser))).Methods("DELETE")
	a.Router.Handle("/api/v1/user/{username}", Authenticate(http.HandlerFunc(a.GetUserByUsername))).Methods("GET")
	a.Router.HandleFunc("/api/v1/login", a.LoginUser).Methods("POST")
	a.Router.Handle("/api/v1/logout", Authenticate(http.HandlerFunc(a.LogoutUser))).Methods("POST")
	a.Router.Handle("/api/v1/user/{username}", Authenticate(http.HandlerFunc(a.UpdateUser))).Methods("PUT")
	// usergroup
	a.Router.Handle("/api/v1/usergroup", Authenticate(http.HandlerFunc(a.AddUserGroup))).Methods("POST")
	a.Router.Handle("/api/v1/usergroup/{id}", Authenticate(http.HandlerFunc(a.DeleteUserGroupByID))).Methods("DELETE")
	a.Router.Handle("/api/v1/usergroup/{id}", Authenticate(http.HandlerFunc(a.GetUserGroupByID))).Methods("GET")
	a.Router.Handle("/api/v1/usergroup/{id}", Authenticate(http.HandlerFunc(a.UpdateUserGroupByID))).Methods("PUT")
}

// Index controller
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
