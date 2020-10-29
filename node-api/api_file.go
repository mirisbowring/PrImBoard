package node

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/mirisbowring/PrImBoard/helper"
	_http "github.com/mirisbowring/PrImBoard/helper/http"
)

// addFile writes the transmitted file to the filesystem for the user
func (a *App) addFile(w http.ResponseWriter, r *http.Request) {
	// grep node
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("error during file transmission")
		_http.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()
	// get username from header
	username := w.Header().Get("user")
	// create users dir if not exist
	path := filepath.Join(a.Config.BasePath, username, "own")
	_ = os.MkdirAll(path, os.ModePerm)
	// create file
	filename := filepath.Join(path, handler.Filename)
	dst, err := os.Create(filename)
	defer dst.Close()
	if err != nil {
		log.Errorf("could not create <%s>", filename)
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create file")
		return
	}
	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		log.Errorf("could not write <%s> to filesystem", filename)
		_http.RespondWithError(w, http.StatusInternalServerError, "could not write file to filesystem")
		return
	}
	// respond success
	log.Infof("added <%s> to filesystem", filename)
	_http.RespondWithJSON(w, http.StatusCreated, "upload was successfull")
}

// deleteFile deletes a specific file from filesystem
func (a *App) deleteFile(w http.ResponseWriter, r *http.Request) {
	filename, status := _http.ParsePathString(w, r, "filename")
	if status > 0 {
		return
	}
	// check if directory exist
	username := w.Header().Get("user")
	path := filepath.Join(a.Config.BasePath, username)
	if !helper.PathExists(path) {
		log.Errorf("user does not has a path on the filesystem")
		_http.RespondWithError(w, http.StatusBadRequest, "file does not exist")
		return
	}
	// deleting file
	filename = filepath.Join(path, filename)
	if err := os.Remove(filepath.Join(filename)); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("could not delete file <%s>", filename)
		_http.RespondWithError(w, http.StatusBadRequest, "could not delete file")
		return
	}
	// respond success
	log.Infof("deleted <%s> from filesystem", filename)
	_http.RespondWithJSON(w, http.StatusNoContent, "deleted file")
}
