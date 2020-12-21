package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
	"github.com/mirisbowring/primboard/models"
	hModel "github.com/mirisbowring/primboard/models/helper"
	"github.com/mirisbowring/primboard/models/maps"
)

// authCookie stores the temporal cookie object
var authCookie *http.Cookie

// DecodeMediaRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediaRequest(w http.ResponseWriter, r *http.Request, m models.Media) (models.Media, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return models.Media{}, 1
	}
	defer r.Body.Close()
	return m, 0
}

// DecodeMediasRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediasRequest(w http.ResponseWriter, r *http.Request) ([]models.Media, int) {
	var m []models.Media
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return m, 0
}

// DecodeMediaGroupMapRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediaGroupMapRequest(w http.ResponseWriter, r *http.Request) (models.MediaGroupMap, int) {
	var mgm models.MediaGroupMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mgm); err != nil {
		// an decode error occured
		_http.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return mgm, 1
	}
	defer r.Body.Close()
	return mgm, 0
}

// // addMediaToIpfsNode uploads the given file to the specified ipfs node.
// // The passed media model will be completed with path and hashes.
// func addMediaToIpfsNode(file string, media Media, node Node) (Media, error) {
// 	// new ipfs shell
// 	sh := ipfs.NewShell(node.Address + ":" + strconv.Itoa(node.IPFSAPIPort))
// // create file pointer
// r, _ := os.Open(file)
// // create thumbnail and receive pointer
// rt, _ := h.Thumbnail(r, 128)
// 	// add the thumbnail to the ipfs
// 	thumbCid, err := sh.Add(rt)
// 	if err != nil {
// 		log.Println(err)
// 		return Media{}, errors.New("could not upload thumbnail to ipfs node")
// 	}
// 	r.Close()

// 	//recreate file pointer (add is manipulating it)
// 	r, _ = os.Open(file)
// 	// add the file to ipfs
// 	// do not use the recursive AddDir because we need to add all the files to the mongo
// 	cid, err := sh.Add(r)
// 	if err != nil {
// 		log.Println(err)
// 		return Media{}, errors.New("could not upload file to ipfs node")
// 	}
// 	r.Close()

// 	// if successfull, create a media object with the returned ipfs url
// 	// var m Media
// 	// if (src.Meta != thumbnailer.Meta{} && src.Meta.Title != "") {
// 	// 	m.Title = src.Meta.Title
// 	// }
// 	media.Sha1 = cid
// 	media.URL = node.IPFSGateway + cid
// 	media.URLThumb = node.IPFSGateway + thumbCid
// 	// // eval mime to generic type
// 	// if src.HasVideo {
// 	// 	m.Type = "video"
// 	// } else if src.HasAudio {
// 	// 	m.Type = "audio"
// 	// } else {
// 	// 	m.Type = "image"
// 	// }
// 	// m.Format = src.Extension
// 	// encode the object to json
// 	// b := new(bytes.Buffer)
// 	// json.NewEncoder(b).Encode(m)
// 	// // post the object to the api
// 	// post("http://"+PrimboardHost+"/api/v1/media", "application/json", b)
// 	return media, nil
// }

// post creates a http client and posts the data with the auth cookie to the api
func post(url string, contentType string, body io.Reader) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", contentType)
	// set cookie
	req.AddCookie(authCookie)
	res, _ := client.Do(req)
	if res.StatusCode != 201 {
		log.Fatal(res.StatusCode)
	} else {
		readSessionCookie(res)
	}
}

func addMediaToNode(filePath string, m models.Media, node models.Node, client *http.Client) (models.Media, error) {
	// read File from filesystem
	file, err := helper.ReadFile(filePath)
	if err != nil {
		return m, err
	}
	defer file.Close()

	// generate hash
	if m.Sha1 = helper.GenerateSHA1(file); m.Sha1 == "" {
		return m, errors.New("could not generate hash for file")
	}

	// create thumbanil
	rt := handler.CreateThumbnail(filePath)
	m.FileNameThumb = handler.ParseFileName(m.Sha1, m.Creator, true, m.Extension)
	m.FileName = handler.ParseFileName(m.Sha1, m.Creator, false, m.Extension)

	// create request body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// jump back to start of file (changed due to hash calculation)
	file.Seek(0, io.SeekStart)

	// write original file to request
	original, err := writer.CreateFormFile("uploadfile", m.FileName)
	if err != nil {
		log.WithFields(log.Fields{
			"type":  "original",
			"path":  filePath,
			"error": err.Error(),
		}).Error("could not create writer for formfile")
		return m, err
	}
	io.Copy(original, file)

	// write thumbnail to request
	thumb, err := writer.CreateFormFile("uploadthumb", m.FileNameThumb)
	if err != nil {
		log.WithFields(log.Fields{
			"type":  "thumbnail",
			"path":  filePath,
			"error": err.Error(),
		}).Error("could not create writer for formfile")
		return m, err
	}
	io.Copy(thumb, rt)

	// close writer
	err = writer.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not close writer")
		return m, err
	}

	endpoint := fmt.Sprintf("%s/api/v1/file/%s", node.APIEndpoint, m.Creator)
	contentType := writer.FormDataContentType()

	// executing request
	res, status, msg := _http.SendRequest(client, "POST", endpoint, node.Secret, body, contentType)
	if status > 0 {
		return m, errors.New(msg)
	}

	// preparing log fields
	logfields := log.Fields{
		"hash":        m.Sha1,
		"file":        m.FileName,
		"thumb":       m.FileNameThumb,
		"node":        node.ID.Hex(),
		"status-code": res.StatusCode,
	}

	switch res.StatusCode {
	case http.StatusCreated:
		log.WithFields(logfields).Info("media created on node")
		handler.RemoveFile(filePath)

		// add node to media
		m.NodeIDs = append(m.NodeIDs, node.ID)
		return m, nil
	default:
		msg := "could not push media to node"
		log.WithFields(logfields).Error(msg)
		return m, errors.New(msg)
	}
}

func (g *AppGateway) removeMediasFromNode(medias []models.Media) map[string][]string {
	username := medias[0].Creator
	nodes := make(map[string]models.Node)
	requests := make(map[string][]string)
	failed := make(map[string][]string)
	// iterate over medias to map them to respective nodes
	for _, med := range medias {
		// if file is not on any node - skip
		if len(med.Nodes) == 0 {
			continue
		}
		// iterate over all nodes for that file
		for _, node := range med.Nodes {
			id := node.ID.Hex()
			// check if a key for that node does already exist
			if val, ok := requests[id]; ok {
				// add file to node for sharing on node
				val = append(val, med.FileName)
				requests[id] = val
			} else {
				// add node to node map
				nodes[id] = *g.Nodes[node.ID]
				// create new key for node with groups to share with
				requests[id] = []string{med.FileName}
			}
		}
	}

	// iterate over nodes to send bulk delete
	for id, node := range nodes {
		endpoint := fmt.Sprintf("%s/api/v1/files/%s/remove", node.APIEndpoint, username)
		contentType := "application/json"

		val, _ := requests[id]

		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(val)

		// prepare log fields
		logfields := log.Fields{
			"node":        id,
			"endpoint":    endpoint,
			"contentType": contentType,
		}

		res, status, msg := _http.SendRequest(g.HTTPClient, "POST", endpoint, node.Secret, body, contentType)
		if status > 0 {
			logfields["error"] = msg
			log.WithFields(logfields).Error("could not send request")
			continue
		}

		// append status code to logs
		logfields["status-code"] = res.StatusCode

		// check response status code
		switch res.StatusCode {
		case http.StatusOK:
			log.WithFields(logfields).Info("deleted file on node successfully")
			break
		case http.StatusBadRequest:
			bytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.WithFields(logfields).Error("could not decode response body")
				break
			}
			logfields["response"] = string(bytes)
			log.WithFields(logfields).Error("malformed request")
			break
		case 902:
			var err _http.ErrorJSON
			if e := json.NewDecoder(res.Body).Decode(&err); e != nil {
				logfields["error"] = e.Error()
				log.WithFields(logfields).Error("cannot decode response body")
				break
			}
			fail, ok := err.Payload.([]string)
			if !ok {
				log.WithFields(logfields).Error("cannot parse error payload to []string")
				break
			} else {
				failed[id] = fail
			}
			break
		default:
			log.WithFields(logfields).Error("unexpected status code")
		}
	}

	return failed

}

// prepareGroupMedia parses the sharing from body and chooses all related groups
// and medias, the user is able to access
//
// 0 -> ok
// 1 -> could not decode request
// 2 -> body does not contain sharable informations (empty)
// 3 -> could not parse mediaIDs
// 4 -> could not read usergroups from database
// 5 -> no valid groups found
// 6 -> could not read media from database
// 7 -> no valid media found
func (g *AppGateway) prepareGroupMedia(w http.ResponseWriter, r *http.Request) (*hModel.GroupMediaHelper, int) {
	_helper := hModel.GroupMediaHelper{}
	var err error

	mgm, status := DecodeMediaGroupMapRequest(w, r)
	if status != 0 {
		return nil, 1
	}

	// check that there is anything, that could be shared
	if len(mgm.Groups) == 0 || len(mgm.MediaIDs) == 0 {
		_http.RespondWithJSON(w, http.StatusBadRequest, "nothing specified to share")
		return nil, 2
	}

	// parsing ids
	_helper.MediaIDs, err = ParseIDs(mgm.MediaIDs)
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return nil, 3
	}

	// extract group ids from groups
	for _, group := range mgm.Groups {
		_helper.GroupIDs = append(_helper.GroupIDs, group.ID)
	}

	// select all groups from list, the user has access to
	_helper.Groups, err = models.GetUserGroupsByIDs(g.DB, _helper.GroupIDs, g.GetUserPermissionW(w, false))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not select matching groups from database")
		return nil, 4
	}

	// verify that there is any valid group to share the media with
	if _helper.Groups == nil || len(_helper.Groups) == 0 {
		_http.RespondWithJSON(w, http.StatusBadRequest, "no group to add to the media")
		return nil, 5
	}

	// select all medias from list, the user is Owner of
	_helper.Medias, err = models.GetMediaByIDs(g.DB, _helper.MediaIDs, g.GetUserPermissionW(w, true))
	if err != nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not select matching medias from database")
		return nil, 6
	}

	// verify that there is any valid media to share
	if _helper.Medias == nil || len(_helper.Medias) == 0 {
		_http.RespondWithError(w, http.StatusBadRequest, "no group to add to the media")
		return nil, 7
	}

	log.Info(_helper)

	_helper.GroupIDs = nil
	_helper.MediaIDs = nil
	for _, med := range _helper.Medias {
		_helper.MediaIDs = append(_helper.MediaIDs, med.ID)
	}
	for _, group := range _helper.Groups {
		_helper.GroupIDs = append(_helper.GroupIDs, group.ID)
	}

	log.Info(_helper)

	return &_helper, 0
}

// readSessionCookie reads the auth cookie from the response
func readSessionCookie(r *http.Response) {
	cookies := r.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "stoken" {
			authCookie = cookie
			return
		}
	}
	log.Fatal("Could not read authentication token!")
}

// shareMediaToGroup creates a requeast to map all passed media to/from passed
// groups
//
// action must be "remove" || "add" (to add or remove the share for that file)
func (g *AppGateway) shareMediaToGroup(medias []models.Media, groups []models.UserGroup, action string) map[string][]maps.FilesGroupsMap {
	username := medias[0].Creator
	var groupIDs []string

	// prevent exception
	if groups != nil {
		// iterating over all groups
		for _, group := range groups {
			groupIDs = append(groupIDs, group.ID.Hex())
		}
	}

	// requests holds all files / groups that are available on a specific node
	nodes := make(map[string]models.Node)
	requests := make(map[string]maps.FilesGroupsMap)
	failed := make(map[string][]maps.FilesGroupsMap)

	// iterate over all files to share
	for _, med := range medias {
		// if file is not on any node - skip
		if len(med.Nodes) == 0 {
			continue
		}
		if groups == nil {
			groupIDs = UnParseIDs(med.GroupIDs)
		}
		// iterate over all nodes for that file
		for _, node := range med.Nodes {
			id := node.ID.Hex()
			// check if a key for that node does already exist
			if val, ok := requests[id]; ok {
				// add file to node for sharing on node
				val.Filenames = append(val.Filenames, med.FileName)
				requests[id] = val
			} else {
				log.Info(g.Nodes)
				// add node to node map
				nodes[id] = *(g.Nodes)[node.ID]
				// create new key for node with groups to share with
				requests[id] = maps.FilesGroupsMap{
					Filenames: []string{med.FileName},
					Groups:    groupIDs,
				}
			}
		}
	}

	// iterate over find nodes and post request
	for id, node := range nodes {
		endpoint := fmt.Sprintf("%s/api/v1/files/%s/share", node.APIEndpoint, username)
		contentType := "application/json"

		val, _ := requests[id]

		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(val)

		// prepare log fields
		logfields := log.Fields{
			"node":        id,
			"endpoint":    endpoint,
			"contentType": contentType,
			"action":      action,
		}

		// refresh keycloaktoken in neccessary
		g.keycloakRefreshToken()

		var res *http.Response
		var status int
		var msg string
		switch action {
		case "add":
			res, status, msg = _http.SendRequest(g.HTTPClient, "POST", endpoint, g.KeycloakToken.AccessToken, body, contentType)
			break
		case "remove":
			res, status, msg = _http.SendRequest(g.HTTPClient, "DELETE", endpoint, g.KeycloakToken.AccessToken, body, contentType)
			break
		default:
			log.WithFields(logfields).Error("unknown action specified")
			return nil
		}

		if status > 0 {
			logfields["error"] = msg
			log.WithFields(logfields).Error("could not send request")
			continue
		}

		// apend status code to logs
		logfields["status-code"] = res.StatusCode

		// check response status code
		switch res.StatusCode {
		case http.StatusOK:
			log.WithFields(logfields).Info("removed share for files successfully")
			break
		case http.StatusCreated:
			log.WithFields(logfields).Info("shared files successfully")
			break
		case http.StatusUnauthorized:
			log.WithFields(logfields).Error("could not authorize to node")
			break
		case http.StatusBadRequest:
			bytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.WithFields(logfields).Error("could not decode response body")
				break
			}
			logfields["response"] = string(bytes)
			log.WithFields(logfields).Error("malformed request")
			break
		case 901:
			var err _http.ErrorJSON
			json.NewDecoder(res.Body).Decode(&err)
			fail, ok := err.Payload.([]maps.FilesGroupsMap)
			if !ok {
				log.WithFields(logfields).Error("cannot parse payload to FilesGroupsMap")
				break
			} else {
				failed[id] = fail
			}
			break
		default:
			log.WithFields(logfields).Error("unexpected status code")
		}
	}

	return failed

}

// walkDir recursively iterates a given folder and adds all files to a slice
// (ignoring subdirs)
func walkDir(dir string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		//ignore dirs
		mode, _ := os.Stat(path)
		if !mode.Mode().IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal("An error occured while fetching file informations!")
	}

	return files
}
