package node

import (
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/internal/handler"
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
	if status := handler.CreateFile(n.Config.BasePath, file, handlerOrigin, username, "original", w); status > 0 {
		return
	}
	// create thumbnail
	if status := handler.CreateFile(n.Config.BasePath, fileThumb, handlerThumb, username, "thumb", w); status > 0 {
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
	if status = handler.DeleteFile(n.Config.BasePath, username, filename, w); status > 0 {
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
		if status = handler.DeleteFile(n.Config.BasePath, username, file, nil); status > 0 {
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
