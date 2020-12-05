package infrastructure

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/models"
	log "github.com/sirupsen/logrus"
)

// NodeAuthentication can be called by the server to authenticate or
// unauthenticate a user from a node.
func NodeAuthentication(session *iModels.Session, nodes []models.Node, authenticate bool, client *http.Client) (int, string) {
	username := session.User.Username

	// verify that nodes have been passed
	if nodes == nil || len(nodes) == 0 {
		// did not iterate over the nodes
		log.WithFields(log.Fields{
			"username":     username,
			"authenticate": authenticate,
		}).Info("no nodes to authenticate to")
	}

	// iterate over the passed nodes
	for _, node := range nodes {
		var endpoint string
		var contentType string
		var contentReader *strings.Reader
		var token string
		var msg string

		// check if user should be authenticated or unauthenticated
		if authenticate {
			endpoint = fmt.Sprintf("%s/api/v1/user/%s/authenticate", node.APIEndpoint, username)
			contentType = "application/json"
			token = helper.GenerateRandomToken(30)
			contentReader = strings.NewReader(fmt.Sprintf("\"%s\"", token))
			msg = "authenticated user to node"
		} else {
			endpoint = fmt.Sprintf("%s/api/v1/user/%s/unauthenticate", node.APIEndpoint, username)
			contentReader = &strings.Reader{}
			msg = "unauthenticated user from node"
		}

		// verify that user should have access to node
		if username != node.Creator && !helper.ObjectIDIntersect(session.Usergroups, node.GroupIDs) || node.Secret == "" {
			continue
		}

		// create endpoint (api call)
		res, status, msg := _http.SendRequest(client, "POST", endpoint, node.Secret, contentReader, contentType)
		if status > 0 {
			return 1, msg
		}

		// collecting neccessary fields for logging
		logfields := log.Fields{
			"authenticate": authenticate,
			"username":     username,
			"node":         node.Title,
			"endpoint":     endpoint,
			"status-code":  res.StatusCode,
		}

		// read body (for logging mainly)
		logfields["message"], _ = _http.ParseBody(res.Body, logfields)

		switch res.StatusCode {
		case http.StatusOK:
			break
		case http.StatusNotFound:
			msg = "node not found"
			log.WithFields(logfields).Error("could not fond node")
			continue
		default:
			msg = "unexpected response"
			log.WithFields(logfields).Error(msg)

			continue
		}

		// un/authentication succeded
		log.WithFields(logfields).Info(msg)

		// add / remove node/token map for user
		if authenticate {
			session.NodeTokenMap[node.ID] = token
		} else {
			delete(session.NodeTokenMap, node.ID)
		}
	}

	return 0, ""
}
