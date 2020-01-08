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
	a.Router.HandleFunc("/api/v1/event", a.AddEvent).Methods("POST")
	a.Router.HandleFunc("/api/v1/event/{id}", a.DeleteEventByID).Methods("DELETE")
	a.Router.HandleFunc("/api/v1/event/{id}", a.GetEventByID).Methods("GET")
	a.Router.HandleFunc("/api/v1/event/{id}", a.UpdateEventByID).Methods("PUT")
	// media
	a.Router.Handle("/api/v1/media", Authenticate(http.HandlerFunc(a.GetMedia))).Methods("GET")
	// a.Router.HandleFunc("/api/v1/media", a.GetMedia).Methods("GET")
	a.Router.HandleFunc("/api/v1/media", a.AddMedia).Methods("POST")
	a.Router.HandleFunc("/api/v1/media/{id}", a.DeleteMediaByID).Methods("DELETE")
	a.Router.HandleFunc("/api/v1/media/{id}", a.GetMediaByID).Methods("GET")
	a.Router.HandleFunc("/api/v1/media/{id}", a.UpdateMediaByID).Methods("PUT")
	a.Router.HandleFunc("/api/v1/mediaByHash/{ipfs_id}", a.GetMediaByHash).Methods("GET")
	// tag
	a.Router.HandleFunc("/api/v1/tag", a.AddTag).Methods("POST")
	a.Router.HandleFunc("/api/v1/tag/{id}", a.DeleteTagByID).Methods("POST")
	a.Router.HandleFunc("/api/v1/tag/{id}", a.GetTagByID).Methods("GET")
	a.Router.HandleFunc("/api/v1/tag/{id}", a.UpdateTagByID).Methods("PUT")
	// user
	a.Router.HandleFunc("/api/v1/user", a.CreateUser).Methods("POST")
	a.Router.HandleFunc("/api/v1/user/{username}", a.DeleteUser).Methods("DELETE")
	a.Router.HandleFunc("/api/v1/user/{username}", a.GetUserByUsername).Methods("GET")
	a.Router.HandleFunc("/api/v1/login", a.LoginUser).Methods("POST")
	a.Router.HandleFunc("/api/v1/logout", a.LogoutUser).Methods("POST")
	a.Router.HandleFunc("/api/v1/user/{username}", a.UpdateUser).Methods("PUT")
	// usergroup
	a.Router.HandleFunc("/api/v1/usergroup", a.AddUserGroup).Methods("POST")
	a.Router.HandleFunc("/api/v1/usergroup/{id}", a.DeleteUserGroupByID).Methods("DELETE")
	a.Router.HandleFunc("/api/v1/usergroup/{id}", a.GetUserGroupByID).Methods("GET")
	a.Router.HandleFunc("/api/v1/usergroup/{id}", a.UpdateUserGroupByID).Methods("PUT")
}

// Index controller
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
