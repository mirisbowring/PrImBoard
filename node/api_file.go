package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
	"github.com/mirisbowring/primboard/models"
	"github.com/mirisbowring/primboard/models/maps"
	log "github.com/sirupsen/logrus"
)

// addFile writes the transmitted file to the filesystem for the user
func (n *AppNode) addFile(w http.ResponseWriter, r *http.Request) {
	// parse username
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}

	// get file
	file, handlerOrigin, err := r.FormFile("uploadfile")
	if err != nil {
		log.WithFields(log.Fields{
			"formfile": "updloadfile",
			"error":    err.Error(),
		}).Errorf("error during file transmission")
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	// get thumb
	fileThumb, handlerThumb, err := r.FormFile("uploadthumb")
	if err != nil {
		log.WithFields(log.Fields{
			"formfile": "uploadthumb",
			"error":    err.Error(),
		}).Errorf("error during file transmission")
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	// create original
	if status := handler.CreateFileFromMultipart(n.Config.BasePath, file, handlerOrigin, username, "original", w); status > 0 {
		return
	}
	// create thumbnail
	if status := handler.CreateFileFromMultipart(n.Config.BasePath, fileThumb, handlerThumb, username, "thumb", w); status > 0 {
		return
	}

	// respond success
	_http.RespondWithJSON(w, http.StatusCreated, "upload was successfull")
}

// deleteFile deletes a specific file from filesystem
func (n *AppNode) deleteFile(w http.ResponseWriter, r *http.Request) {
	// parse filename from url
	filename, status := _http.ParsePathString(w, r, "filename")
	if status > 0 {
		return
	}
	// parse username from url
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}

	// delete the files and all shares
	if status = handler.DeleteFile(n.Config.BasePath, username, "", filename, w); status > 0 {
		return
	}

	// respond success
	_http.RespondWithJSON(w, http.StatusNoContent, "deleted file")
}

func (n *AppNode) deleteFiles(w http.ResponseWriter, r *http.Request) {
	// parse ursername from url
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}

	// parse files from body
	var files []string
	files, status = _http.DecodeStringsRequest(w, r, files)
	if status < 0 {
		return
	}

	var failed []string
	// iterate over
	for _, file := range files {
		if status = handler.DeleteFile(n.Config.BasePath, username, "", file, nil); status > 0 {
			failed = append(failed, file)
		}
	}

	// return corresponding statuscode if any file failed
	if len(failed) != 0 {
		_http.RespondWithJSON(w, 902, failed)
		return
	}

	_http.RespondWithJSON(w, http.StatusOK, "Deleted all files")
}

// deleteShareForGroup deletes the share of a file for the specific group
func (n *AppNode) deleteShareForGroup(w http.ResponseWriter, r *http.Request) {
	// parse username from url
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}

	// parse filename from url
	filename, status := _http.ParsePathString(w, r, "filename")
	if status > 0 {
		return
	}

	// parse froup from url
	group, status := _http.ParsePathString(w, r, "group")
	if status > 0 {
		return
	}

	// create map for deleting shares
	maps := maps.FilesGroupsMap{
		Filenames: []string{filename},
		Groups:    []string{group},
	}

	// delete file for group
	if failed := handler.DeleteShares(n.Config.BasePath, username, maps, w); len(failed) > 0 {
		_http.RespondWithJSON(w, 901, _http.ErrorJSON{Error: "could not remove share for all files", Payload: maps})
		return
	}
	// if status = handler.DeleteFile(n.Config.BasePath, username, group, filename, w); status > 0 {
	// 	return
	// }

	// respond success
	_http.RespondWithJSON(w, http.StatusNoContent, "deleted share for group")

}

func (n *AppNode) deleteShares(w http.ResponseWriter, r *http.Request) {
	// parse username from url
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}

	// parse maps from request
	maps, status := _http.DecodeFilesGroupsMapRequest(w, r)
	if status > 0 {
		return
	}

	// delete the specified shares
	if failed := handler.DeleteShares(n.Config.BasePath, username, maps, w); len(failed) > 0 {
		_http.RespondWithJSON(w, 901, _http.ErrorJSON{Error: "could not remove share for all files", Payload: maps})
		return
	}

	_http.RespondWithJSON(w, http.StatusOK, "deleted shares successfully")

}

func (n *AppNode) getFile(w http.ResponseWriter, r *http.Request) {
	// parse identifier
	ident, status := _http.ParsePathString(w, r, "identifier")
	if status > 0 {
		return
	}

	// parse filename
	file, status := _http.ParsePathString(w, r, "filename")
	if status > 0 {
		return
	}

	// parse optional group query
	group, status := _http.ParseQueryBool(w, r, "group", true)
	if status > 0 {
		return
	}

	// parse optional thumb query
	thumb, status := _http.ParseQueryBool(w, r, "thumb", true)
	if status > 0 {
		return
	}

	var path string
	if group {
		path = n.getDataPath(ident, pathTypeGroup, thumb)
	} else {
		path = n.getDataPath(ident, pathTypeUser, thumb)
	}
	path = fmt.Sprintf("%s/%s", path, file)
	http.ServeFile(w, r, path)
}

func (n *AppNode) shareFiles(w http.ResponseWriter, r *http.Request) {
	// parse username from url
	username, status := _http.ParsePathString(w, r, "username")
	if status > 0 {
		return
	}

	// parse maps from request
	maps, status := _http.DecodeFilesGroupsMapRequest(w, r)
	if status > 0 {
		return
	}

	// share the files
	if maps := handler.ShareFiles(n.Config.BasePath, username, maps); len(maps) > 0 {
		_http.RespondWithJSON(w, 901, _http.ErrorJSON{Error: "could not share all files", Payload: maps})
		return
	}

	_http.RespondWithJSON(w, http.StatusCreated, "shared files successfully")
}

func (n *AppNode) uploadFile(w http.ResponseWriter, r *http.Request) {
	username := _http.GetUsernameFromHeader(w)

	// grep filemeta
	meta := r.FormValue("filemeta")
	m := models.Media{}
	if err := json.Unmarshal([]byte(meta), &m); err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, "could not unmarshal passed filemeta")
		return
	}

	// receive file
	file, fileHeader, err := r.FormFile("uploadfile")
	if err != nil {
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	// set the username
	m.Creator = username

	// verify dir existance
	path := n.getDataPath(m.Creator, pathTypeUser, false)
	pathThumb := n.getDataPath(m.Creator, pathTypeUser, true)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not create path")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create path")
		return
	}
	if err := os.MkdirAll(pathThumb, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"path":  pathThumb,
			"error": err.Error(),
		}).Error("could not create thumbnail path")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create thumbnail path")
		return
	}

	// Create file
	filepath := fmt.Sprintf("%s%s", path, fileHeader.Filename)
	if status := handler.CreateFile(filepath, file); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, "could create file")
		return
	}

	// create new stream
	file, err = os.Open(filepath)
	if err != nil {
		log.WithFields(log.Fields{
			"filepath": filepath,
			"error":    err.Error(),
		}).Error("could not open file")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not open file to calculate checksum")
		return
	}

	// generate hash
	if m.Sha1 = helper.GenerateSHA1(file); m.Sha1 == "" {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not calculate checksum for file")
		return
	}

	// parse filenames
	m.FileNameThumb = handler.ParseFileName(m.Sha1, m.Creator, true, m.Extension)
	m.FileName = handler.ParseFileName(m.Sha1, m.Creator, false, m.Extension)

	// rename tmp original to naming standard
	if err := os.Rename(filepath, fmt.Sprintf("%s%s", path, m.FileName)); err != nil {
		log.WithFields(log.Fields{
			"old":   filepath,
			"new":   fmt.Sprintf("%s%s", path, m.FileName),
			"error": err.Error(),
		}).Error("could not rename file to match standard")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not finish file")
		return
	}

	// create thumbanil
	rt := handler.CreateThumbnail(fmt.Sprintf("%s%s", path, m.FileName))
	if rt == nil {
		_http.RespondWithError(w, http.StatusInternalServerError, "could not render thumbnail")
		return
	}
	filepath = fmt.Sprintf("%s%s", pathThumb, m.FileNameThumb)
	if status := handler.CreateFile(filepath, rt); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, "could create thumbnail file")
		return
	}

	// parse media to json
	data, err := json.Marshal(m)
	if err != nil {
		log.WithFields(log.Fields{
			"media": m,
			"error": err.Error(),
		}).Error("could not marshal media to json")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not marshal media to json")
		return
	}

	// create reader from json bytes
	body := bytes.NewReader(data)

	// refresh keycloaktoken in neccessary
	n.keycloakRefreshToken()

	// post media to gateway
	resp, status, msg := _http.SendRequest(n.HTTPClient, http.MethodPost, n.Config.GatewayURL+"/api/v1/media", n.KeycloakToken.AccessToken, body, "application/json")
	if status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	logfields := log.Fields{
		"media":       m,
		"status-code": resp.StatusCode,
	}
	switch resp.StatusCode {
	case http.StatusCreated:
		log.WithFields(logfields).Debug("created media on gateway successfull")
		_http.RespondWithJSON(w, http.StatusOK, "created media")
		return
	default:
		log.WithFields(logfields).Error("unexpected status code")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create media on gateway")
		return
	}

	// file to specified node
	// m, err = addMediaToNode(filename, m, n, g.HTTPClient)
	// if err != nil {
	// 	_http.RespondWithError(w, http.StatusInternalServerError, "could not push media to node")
	// 	return
	// }
	// m, err = addMediaToIpfsNode(filename, m, n)
	// if err != nil {
	// 	_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	// try to insert model into db
	// result, err := m.AddMedia(g.DB)
	// if err != nil {
	// 	_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	// creation successful
	// _http.RespondWithJSON(w, http.StatusCreated, result)
}
