package node

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	_http "github.com/mirisbowring/PrImBoard/helper/http"
)

func (a *App) createFile(file multipart.File, header *multipart.FileHeader, username string, _type string, w http.ResponseWriter) error {
	var path string
	// eval upload type
	switch _type {
	case "original":
		path = filepath.Join(a.Config.BasePath, username, "own")
		break
	case "thumb":
		path = filepath.Join(a.Config.BasePath, username, "own", "thumb")
		break
	default:
		log.WithFields(log.Fields{
			"type": _type,
		}).Error("unknown transmission type for new files")
		_http.RespondWithError(w, http.StatusBadRequest, "unknown transmission type for new files")
		return errors.New("unknown transmission type for new files")
	}

	// create path if not exist
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not create path")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create path")
		return err
	}

	// create file
	filename := filepath.Join(path, header.Filename)
	dst, err := os.Create(filename)
	defer dst.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"path":     path,
			"filename": filename,
			"error":    err.Error(),
		})
		log.Errorf("could not create file")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create file")
		return err
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		log.WithFields(log.Fields{
			"path":     path,
			"filename": filename,
			"error":    err.Error(),
		}).Error("could not write file to filesystem")
		_http.RespondWithError(w, http.StatusInternalServerError, "could not write file to filesystem")
		return err
	}

	log.WithFields(log.Fields{
		"path":     path,
		"filename": filename,
	}).Info("added file to filesystem")

	// success
	return nil
}
