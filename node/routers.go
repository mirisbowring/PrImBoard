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
	// index
	n.Router.HandleFunc("/api/v1/", n.index).Methods("GET")
	// files
	n.Router.Handle("/api/v1/file", n.authenticate(http.HandlerFunc(n.AddFile), false)).Methods("POST")
	n.Router.Handle("/api/v1/file/{filename}", n.authenticate(http.HandlerFunc(n.DeleteFile), false)).Methods("DELETE")

	n.Router.Handle("/api/v1/user/{username}/authenticate", n.authenticate(http.HandlerFunc(n.authenticateUser), false)).Methods("POST")
	n.Router.Handle("/api/v1/user/{username}/unauthenticate", n.authenticate(http.HandlerFunc(n.unauthenticateUser), false)).Methods("POST")
}

// Index controller
func (n *AppNode) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently Not Supported!")
}
