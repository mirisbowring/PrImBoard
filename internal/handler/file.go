package handler

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models/maps"
	log "github.com/sirupsen/logrus"
)

// CreateFileFromMultipart creates the file on the filesystem for the user
//
// 0 -> ok
// 1 -> unknown transmission type
// 2 -> could not create path
// 3 -> could not create file
// 4 -> could not copy uploaded file into created file
func CreateFileFromMultipart(basePath string, file multipart.File, header *multipart.FileHeader, username string, _type string, w http.ResponseWriter) int {
	var path string
	// eval upload type
	switch _type {
	case "original":
		path = filepath.Join(basePath, "user", username, "own")
		break
	case "thumb":
		path = filepath.Join(basePath, "user", username, "own", "thumb")
		break
	default:
		log.WithFields(log.Fields{
			"type": _type,
		}).Error("unknown transmission type for new files")
		_http.RespondWithError(w, http.StatusBadRequest, "unknown transmission type for new files")
		return 1
	}

	// create path if not exist
	if status, msg := createPath(path); status > 0 {
		_http.RespondWithError(w, http.StatusInternalServerError, msg)
		return 2
	}

	filename := filepath.Join(path, header.Filename)
	status := CreateFile(filename, file)
	switch status {
	case 0:
		// success
		return 0
	case 1:
		// creation failed
		_http.RespondWithError(w, http.StatusInternalServerError, "could not create file")
		return 3
	case 2:
		// write failed
		_http.RespondWithError(w, http.StatusInternalServerError, "could not write file to filesystem")
		return 4
	default:
		// unexpected status code
		log.WithFields(log.Fields{
			"status": status,
		}).Error("unexpected status from function")
		_http.RespondWithError(w, http.StatusInternalServerError, "Could not process file")
		return 5
	}
}

// CreateFile writes a reader to the filesystem to as passed absolute filename
//
// 0 -> ok
// 1 -> could not create file
// 2 -> could not write file
func CreateFile(filepath string, reader io.Reader) int {
	// create file
	log.Warn(filepath)
	dst, err := os.Create(filepath)
	defer dst.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"filepath": filepath,
			"error":    err.Error(),
		})
		log.Errorf("could not create file")
		return 1
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, reader); err != nil {
		log.WithFields(log.Fields{
			"filepath": filepath,
			"error":    err.Error(),
		}).Error("could not write file to filesystem")
		return 2
	}

	// creation successfull
	log.WithFields(log.Fields{
		"filepath": filepath,
	}).Info("added file to filesystem")

	return 0
}

// recursively creates a path
//
// 0 -> ok || 1 -> error occured during creation
func createPath(path string) (int, string) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		msg := "could not create path"
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error(msg)
		return 1, msg
	}
	log.WithFields(log.Fields{"path": path}).Debug("created path")
	return 0, ""
}

// CreateThumbnail uses ffmpeg to generate a thumbnail for the given reader
func CreateThumbnail(filepath string) io.Reader {
	// create file pointer
	r, err := os.Open(filepath)
	if err != nil {
		log.WithFields(log.Fields{
			"filepath": filepath,
			"error":    err.Error(),
		}).Error("could not open file to create thumbnail")
		return nil
	}
	defer r.Close()
	// create thumbnail and receive pointer
	rt, _ := helper.Thumbnail(r, 128)
	return rt
}

// DeleteFile deletes the specified file for the specified user. It deletes all
// local shares before.
//
// Does not write response if w == nil
//
// 0 -> ok || 1 -> could not delete shares || 2 -> could not delete files from
// user dir
func DeleteFile(basePath string, username string, filename string, w http.ResponseWriter) int {
	// create map struct
	groupPath := filepath.Join(basePath, "group")
	tmp, _ := GetDirectories(groupPath)
	maps := maps.FilesGroupsMap{
		Filenames: []string{filename},
		Groups:    tmp,
	}
	logFields := log.Fields{"filename": filename}

	// delete shares first
	if failed := DeleteShares(basePath, username, maps, w); len(failed) > 0 {
		msg := "could not delete alls shares for file, skipping deletion"
		log.WithFields(logFields).Error(msg)
		if w != nil {
			_http.RespondWithJSON(w, 902, _http.ErrorJSON{Error: msg, Payload: failed})
		}
		return 1
	}

	// prepare path
	name, status := ParseThumbnailName(filename)
	if status > 0 {
		if w != nil {
			_http.RespondWithError(w, http.StatusBadRequest, "could not parse thumbnailname")
		}
		return 2
	}
	path := filepath.Join(basePath, "user", username, "own", filename)
	pathThumb := filepath.Join(basePath, "user", username, "own", "thumb", name)

	// delete from user
	if status, msg := RemoveFile(path); status > 0 {
		if w != nil {
			_http.RespondWithError(w, http.StatusInternalServerError, msg)
		}
		return 2
	}

	// remove thumbnail
	if status, msg := RemoveFile(pathThumb); status > 0 {
		if w != nil {
			_http.RespondWithError(w, http.StatusInternalServerError, msg)
		}
		return 2
	}

	return 0
}

// DeleteShares deletes the file and all it's hardlinks
//
// returns a list of group/file maps that has failed
func DeleteShares(basePath string, username string, _maps maps.FilesGroupsMap, w http.ResponseWriter) []maps.FilesGroupsMap {
	groupPath := filepath.Join(basePath, "group")
	var failed []maps.FilesGroupsMap

	for _, group := range _maps.Groups {
		path := filepath.Join(groupPath, group)
		// check if group does exist on this node
		if !helper.PathExists(path) {
			failed = append(failed, maps.FilesGroupsMap{Groups: []string{group}})
			continue
		}
		for _, file := range _maps.Filenames {
			fail := maps.FilesGroupsMap{Groups: []string{group}}
			// build path
			path := filepath.Join(groupPath, file)
			name, status := ParseThumbnailName(file)
			if status > 0 {
				fail.Filenames = append(fail.Filenames, file)
				failed = append(failed, fail)
				continue
			}
			pathThumb := filepath.Join(groupPath, "thumb", name)
			// delete file
			if status, _ := RemoveFile(path); status > 0 {
				fail.Filenames = append(fail.Filenames, file)
				failed = append(failed, fail)
				continue
			}
			// delete thumbnail
			if status, _ := RemoveFile(pathThumb); status > 0 {
				fail.Filenames = append(fail.Filenames, file)
				failed = append(failed, fail)
				continue
			}
		}
	}
	return failed
}

// DeleteFiles deletes all content from specified path. It's using RemoveAll (so
// recreating the folder afterwards)
//
// 0 -> path empty or success
// 1 -> could not remove the files/folders
// 2 -> could not recreate the folder
func DeleteFiles(path string) int {
	if path == "" {
		return 0
	}
	// delete all files
	if err := os.RemoveAll(path); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not delete locations")
		return 1
	}
	// recreate parent folder
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not recreate path")
		return 2
	}
	return 0
}

// GetDirectories lists all directories from specified path
//
// returns [](dirs in path), [](files in path)
func GetDirectories(path string) ([]string, []string) {
	var dirs []string
	var files []string

	// verify that path does exist
	if !helper.PathExists(path) {
		return dirs, files
	}

	// read content from directory
	cont, err := ioutil.ReadDir(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not read path")
		return dirs, files
	}

	// filter for dirs only
	for _, f := range cont {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		} else {
			files = append(files, f.Name())
		}
	}

	return dirs, files
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

	for _, loc := range []string{"group", "user"} {
		file, err := os.Create(fmt.Sprintf("/etc/nginx/locations/%s_%s.location", token, loc))
		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"token":    token,
				"error":    err.Error(),
			}).Error("could not create locations file")
			return 1
		}
		writer := bufio.NewWriter(file)
		switch loc {
		case "user":
			_, err = writer.WriteString(fmt.Sprintf("location /node-data/%s/own {\nalias /data/user/%s/own;\n}", token, username))
			if err != nil {
				log.WithFields(log.Fields{
					"type":     loc,
					"username": username,
					"token":    token,
					"error":    err.Error(),
				}).Error("could not write to locations file")
				return 1
			}
			break
		case "group":
			_, err = writer.WriteString(fmt.Sprintf("location /node-data/%s/groups {\nalias /data/group;\n}", token))
			if err != nil {
				log.WithFields(log.Fields{
					"type":  loc,
					"token": token,
					"error": err.Error(),
				}).Error("could not write to locations file")
				return 1
			}
			break
		}

		writer.Flush()
	}
	return 0
}

// ParseFileName parses the filename from hash, username and extension (thumb toggle)
func ParseFileName(hash string, username string, thumbnail bool, extension string) string {
	log.Info(hash)
	log.Info(username)
	log.Info(thumbnail)
	log.Info(extension)
	switch "" {
	case hash:
		return ""
	case username:
		return ""
	case extension:
		return ""
	default:
		if thumbnail {
			return fmt.Sprintf("%s_%s_%s.%s", hash, username, "thumb", extension)
		}
		return fmt.Sprintf("%s_%s.%s", hash, username, extension)
	}
}

// ParseThumbnailName accepts a file like "abcd.xyz" and adds "_thumb" right
// before the dot.
//
// 0 -> ok || 1 -> to much/few parts
func ParseThumbnailName(filename string) (string, int) {
	parts := strings.Split(filename, ".")
	// verify parts
	if len(parts) != 2 {
		log.WithFields(log.Fields{
			"filename": filename,
			"parts":    len(parts),
		}).Error("filename has an unexpected amount of parts")
		return "", 1
	}
	// build name
	return fmt.Sprintf("%s_thumb.%s", parts[0], parts[1]), 0
}

// RemoveFile removes the file and writes an error response if fails
func RemoveFile(file string) (int, string) {
	msg := "could not delete file"
	err := os.Remove(file)
	if err != nil && !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"path":  file,
			"error": err.Error(),
		}).Error(msg)
		return 1, msg
	}
	log.WithFields(log.Fields{"file": file}).Debug("deleted shared files for file")
	return 0, "deleted file"
}

// ShareFiles tries to hardlink the specified files to the specified groups
//
// returns a list of file/group maps, the sharing process has failed for
func ShareFiles(basePath string, username string, _maps maps.FilesGroupsMap) []maps.FilesGroupsMap {
	var failed []maps.FilesGroupsMap
	for _, file := range _maps.Filenames {
		// create neccessary paths
		fpath := filepath.Join(basePath, "user", username, "own", file)

		// parse thumb name
		fileThumb, status := ParseThumbnailName(file)
		if status > 0 {
			failed = append(failed, maps.FilesGroupsMap{Filenames: []string{file}})
			continue
		}
		fpathThumb := filepath.Join(basePath, "user", username, "own", "thumb", fileThumb)

		// check that file to share does exist
		if !helper.PathExists(fpath) {
			failed = append(failed, maps.FilesGroupsMap{Filenames: []string{file}})
			continue
		}

		// check that the thumbnail does exist
		if !helper.PathExists(fpathThumb) {
			failed = append(failed, maps.FilesGroupsMap{Filenames: []string{file}})
			continue
		}

		// iterate over all groups to share with
		for _, group := range _maps.Groups {
			// prepare group path
			gpath := filepath.Join(basePath, "group", group)
			gpathThumb := filepath.Join(basePath, "group", group, "thumb")
			// create fail object
			fail := maps.FilesGroupsMap{Filenames: []string{file}}
			logfields := log.Fields{
				"filename": file,
				"group":    group,
				"path":     gpath,
			}

			// check that group path is available
			if status, _ := createPath(gpathThumb); status > 0 {
				log.WithFields(logfields).Error("could not share file to group")
				// mark group as failed
				fail.Groups = append(fail.Groups, group)
				continue
			}

			// create hardlink to file
			if err := os.Link(fpath, filepath.Join(gpath, file)); err != nil {
				logfields["error"] = err.Error()
				log.WithFields(logfields).Error("could not link file")
				continue
			}

			// create hardlink to Thumbnail
			if err := os.Link(fpathThumb, filepath.Join(gpathThumb, fileThumb)); err != nil {
				logfields["error"] = err.Error()
				log.WithFields(logfields).Error("could not link file")
				continue
			}

			log.WithFields(logfields).Debug("shared file to group")
		}
	}

	return failed
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
