package http

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrorJSON is an return object to pass data as error response
type ErrorJSON struct {
	Error   string      `json:"error"`
	Payload interface{} `json:"payload"`
}

// GenerateHTTPClient parses the passed cert and adds it to the certpool of a
// custom http client. if insecure enabled, it skips cert validation.
func GenerateHTTPClient(caCert string, insecure bool) (*http.Client, *tls.Config) {
	if caCert != "" {
		if rootCAs, status := loadCaCert(caCert); status == 0 {
			config := &tls.Config{
				InsecureSkipVerify: insecure,
				RootCAs:            rootCAs,
			}
			tr := &http.Transport{TLSClientConfig: config}
			return &http.Client{Transport: tr}, config
		}
	}
	if insecure == true {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		tr := &http.Transport{TLSClientConfig: config}
		return &http.Client{Transport: tr}, config
	}
	return &http.Client{}, nil
}

// GetUsernameFromHeader returns the "user" header value
func GetUsernameFromHeader(w http.ResponseWriter) string {
	return w.Header().Get("user")
}

func loadCaCert(certfile string) (*x509.CertPool, int) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	cert, err := ioutil.ReadFile(certfile)
	if err != nil {
		log.WithFields(log.Fields{
			"file": certfile,
		}).Error("could not read certificate")
		return &x509.CertPool{}, 1
	}

	if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
		log.WithFields(log.Fields{
			"file": certfile,
		}).Error("could not append cert to cert pool")
		return &x509.CertPool{}, 1
	}

	return rootCAs, 0
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

// ParsePathID parses the id from the route and returns it
// returns primitive.NilObjectID if an error occured
// sends a respond if an error occured
func ParsePathID(w http.ResponseWriter, r *http.Request, identifier string) primitive.ObjectID {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars[identifier])
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
// stats 0 -> ok
// status 1 -> error
func ParseQueryString(w http.ResponseWriter, r *http.Request, key string, optional bool) (string, int) {
	val := r.URL.Query().Get(key)
	if val == "" && !optional {
		log.WithFields(log.Fields{
			"key": key,
		}).Warn("key was not specififed in query")
		RespondWithError(w, http.StatusBadRequest, "key was not specified in query")
		return val, 1
	}
	return val, 0
}

// ParseQueryBool parses the string value from the route and converts it into
// bool. If Optional true and value not set, it defaults to false
// 0 -> ok
// 1 -> could not parse string
// 2 -> could not convert to bool
func ParseQueryBool(w http.ResponseWriter, r *http.Request, key string, optional bool) (bool, int) {
	tmp, status := ParseQueryString(w, r, key, optional)
	if status > 0 {
		return false, 1
	}
	if tmp == "" && optional {
		return false, 0
	}
	val, err := strconv.ParseBool(tmp)
	if err != nil {
		log.WithFields(log.Fields{
			"val":   tmp,
			"error": err.Error(),
		}).Error("could not parse value to bool")
		RespondWithError(w, http.StatusBadRequest, "key cannot be converted to bool")
		return false, 2
	}
	return val, 0
}

// ReadCookie reads the stoken cookie from the request and returns the value
func ReadCookie(r *http.Request, title string) string {
	cookie, err := r.Cookie(title)
	if err != nil {
		log.WithFields(log.Fields{
			"title": title,
			"error": err.Error(),
		}).Error("could not read cookie")
		return ""
	}
	return cookie.Value
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

	req.Close = true

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
