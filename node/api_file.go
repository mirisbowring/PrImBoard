package node

import (
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
	log "github.com/sirupsen/logrus"
)

// AddFile writes the transmitted file to the filesystem for the user
func (n *AppNode) AddFile(w http.ResponseWriter, r *http.Request) {
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

	// parse username
	username := r.FormValue("username")
	if username == "" {
		_http.RespondWithError(w, http.StatusBadRequest, "username must be specified")
		return
	}

	// create original
	handler.CreateFile(n.Config.BasePath, file, handlerOrigin, username, "original", w)
	// create thumbnail
	handler.CreateFile(n.Config.BasePath, fileThumb, handlerThumb, username, "thumb", w)

	// respond success
	_http.RespondWithJSON(w, http.StatusCreated, "upload was successfull")
}

// DeleteFile deletes a specific file from filesystem
func (n *AppNode) DeleteFile(w http.ResponseWriter, r *http.Request) {
	filename, status := _http.ParsePathString(w, r, "filename")
	if status > 0 {
		return
	}
	if status = handler.DeleteFile(n.Config.BasePath, filename, w); status > 0 {
		return
	}
	// respond success
	log.Infof("deleted <%s> from filesystem", filename)
	_http.RespondWithJSON(w, http.StatusNoContent, "deleted file")
}
