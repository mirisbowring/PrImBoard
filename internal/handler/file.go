package handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	log "github.com/sirupsen/logrus"
)

func CreateFile(basePath string, file multipart.File, header *multipart.FileHeader, username string, _type string, w http.ResponseWriter) error {
	var path string
	// eval upload type
	switch _type {
	case "original":
		path = filepath.Join(basePath, username, "own")
		break
	case "thumb":
		path = filepath.Join(basePath, username, "own", "thumb")
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

func DeleteFile(basePath string, filename string, w http.ResponseWriter) int {
	// check if directory exist
	username := _http.GetUsernameFromHeader(w)
	path := filepath.Join(basePath, username)
	if !helper.PathExists(path) {
		log.Errorf("user does not has a path on the filesystem")
		_http.RespondWithError(w, http.StatusBadRequest, "file does not exist")
		return 1
	}
	// deleting file
	filename = filepath.Join(path, filename)
	if err := os.Remove(filepath.Join(filename)); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("could not delete file <%s>", filename)
		_http.RespondWithError(w, http.StatusBadRequest, "could not delete file")
		return 1
	}
	return 0
}

// LinkUser creates a symlink for the user folder from basepath to token link
// in targetpath
//
// 0 -> ok || 1 -> could not create symlink
func LinkUser(basePath string, targetPath string, username string, token string) int {
	// verify that locations path does exist
	if err := os.MkdirAll("/etc/nginx/locations/", os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"token":    token,
			"error":    err.Error(),
		}).Error("could not create locations directory")
		return 1
	}

	file, err := os.Create(fmt.Sprintf("/etc/nginx/locations/%s.location", token))
	if err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"token":    token,
			"error":    err.Error(),
		}).Error("could not create locations file")
		return 1
	}
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(fmt.Sprintf("location /node-data/%s {\nalias /data/%s;\n}", token, username))
	if err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"token":    token,
			"error":    err.Error(),
		}).Error("could not write to locations file")
		return 1
	}
	writer.Flush()
	return 0
	// symlink := filepath.Join(basePath, username)
	// if err := os.Symlink(symlink, filepath.Join(targetPath, token)); err != nil {
	// 	log.WithFields(log.Fields{
	// 		"username":   username,
	// 		"token":      token,
	// 		"basepath":   basePath,
	// 		"targetpath": targetPath,
	// 		"error":      err.Error(),
	// 	}).Error("could not create symlink")
	// 	return 1
	// }
	// return 0
}

// UnlinkUser removes the symlink for the token from targetPath
//
// 0 -> ok || 1 -> could not delete symlink
func UnlinkUser(targetPath string, token string) (int, string) {
	if err := os.Remove(fmt.Sprintf("/etc/nginx/locations/%s.location", token)); err != nil {
		msg := "could not delete symlink"
		log.WithFields(log.Fields{
			"token":      token,
			"targetpath": targetPath,
			"error":      err.Error(),
		}).Error(msg)
		return 1, msg
	}
	return 0, ""
}
