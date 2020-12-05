package http

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrorJSON is an return object to pass data as error response
type ErrorJSON struct {
	Error   string      `json:"error"`
	Payload interface{} `json:"payload"`
}

// GetUsernameFromHeader returns the "user" header value
func GetUsernameFromHeader(w http.ResponseWriter) string {
	return w.Header().Get("user")
}

// ParseBody parses the passed readcloser to a string
//
// 0 -> ok || 1 -> could not parse body
func ParseBody(body io.ReadCloser, logfields log.Fields) (string, int) {
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		log.WithFields(logfields).Error("could not parse message body")
		return "", 0
	}
	return string(bytes), 1
}

// ParsePrimitiveID parses the id from the route and returns it
// returns primitive.NilObjectID if an error occured
// sends a respond if an error occured
func ParsePrimitiveID(w http.ResponseWriter, r *http.Request) primitive.ObjectID {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Could not parse ID from route!")
		return primitive.NilObjectID
	}
	return id
}

// ParsePathString parses the string value from the route and returns it
// stats 0 -> ok || status 1 -> error
func ParsePathString(w http.ResponseWriter, r *http.Request, key string) (string, int) {
	if val := mux.Vars(r)[key]; val == "" {
		log.WithFields(log.Fields{
			"key": key,
		}).Warn("key was not specififed")
		RespondWithError(w, http.StatusBadRequest, "key was not specified")
		return val, 1
	} else {
		return val, 0
	}
}

// ParseQueryString parses the string value from the route and returns it
// stats 0 -> ok || status 1 -> error
func ParseQueryString(w http.ResponseWriter, r *http.Request, key string) (string, int) {
	if val := r.URL.Query().Get(key); val == "" {
		log.WithFields(log.Fields{
			"key": key,
		}).Warn("key was not specififed in query")
		RespondWithError(w, http.StatusBadRequest, "key was not specified in query")
		return val, 1
	} else {
		return val, 0
	}
}

// RespondWithError Creates an error payload and adds the error message to be
// returned
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON parses the passed payload and returns it with the specified
// code to the client
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	//	enableCors(&w)
	response, _ := json.Marshal(payload)
	// delete the temporary user key from header
	w.Header().Del("user")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// SendRequest sends a request to an endpoint with specified content and type
func SendRequest(client *http.Client, method string, endpoint string, bearer string, content io.Reader, contentType string) (*http.Response, int, string) {
	logfields := log.Fields{
		"method":   method,
		"endpoint": endpoint,
	}
	req, err := http.NewRequest(method, endpoint, content)
	if err != nil {
		msg := "could not create request"
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error(msg)
		return nil, 1, msg
	}
	// set content type if specified
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// set bearer token
	if bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))
	}

	res, err := client.Do(req)
	// check if error occured during execution
	if err != nil {
		msg := "could not do request"
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error(msg)
		return nil, 1, msg
	}

	// request has been executed
	logfields["status-code"] = res.StatusCode
	log.WithFields(logfields).Info("request executed")
	return res, 0, ""
}
