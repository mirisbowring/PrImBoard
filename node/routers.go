package node

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// initializeRoutes initializes all the available webroutes
func (n *AppNode) initializeRoutes() {
	log.Debug("initializing routes")
	n.Router = mux.NewRouter().StrictSlash(true)
	n.Router.MethodNotAllowedHandler = n.methodNotAllowedHandler()
	// index
	n.Router.HandleFunc("/api/v1/", n.index).Methods("GET")
	// files
	// rela
	n.Router.Handle("/api/v1/file", n.authenticate(http.HandlerFunc(n.uploadFile), false)).Methods("POST")
	n.Router.Handle("/api/v1/file/{username}", n.authenticate(http.HandlerFunc(n.addFile), false)).Methods("POST")
	n.Router.Handle("/api/v1/file/{username}/{filename}", n.authenticate(http.HandlerFunc(n.deleteFile), false)).Methods("DELETE")
	n.Router.Handle("/api/v1/file/{identifier}/{filename}", n.authenticate(http.HandlerFunc(n.getFile), false)).Methods("GET").Queries("thumb", "{thumb}", "group", "{group}", "cookieAuth", "{cookieAuth}")
	n.Router.Handle("/api/v1/file/{username}/{filename}/share/{group}", n.authenticate(http.HandlerFunc(n.deleteShareForGroup), false)).Methods("DELETE")
	n.Router.Handle("/api/v1/files/{username}/remove", n.authenticate(http.HandlerFunc(n.deleteFiles), false)).Methods("POST")
	n.Router.Handle("/api/v1/files/{username}/shares", n.authenticate(http.HandlerFunc(n.shareFiles), false)).Methods("POST")
	n.Router.Handle("/api/v1/files/{username}/shares/remove", n.authenticate(http.HandlerFunc(n.deleteShares), false)).Methods("POST")

	n.Router.Handle("/api/v1/session", n.authenticate(http.HandlerFunc(n.generateSessionCookie), false)).Methods("GET")
	n.Router.Handle("/api/v1/user/{username}/authenticate", n.authenticate(http.HandlerFunc(n.authenticateUser), false)).Methods("POST")
	n.Router.Handle("/api/v1/user/{username}/unauthenticate", n.authenticate(http.HandlerFunc(n.unauthenticateUser), false)).Methods("POST")
}

// Index controller
func (n *AppNode) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
