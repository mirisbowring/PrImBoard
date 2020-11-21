package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/mirisbowring/primboard/helper"
	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
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
	if m.Sha1, err = helper.GenerateSHA1(file); err != nil {
		return m, err
	}

	// create thumbanil
	rt := createThumbnail(filePath)
	m.ThumbnailSha1 = fmt.Sprintf("%s_thumb", m.Sha1)

	// create request body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// jump back to start of file (changed due to hash calculation)
	file.Seek(0, io.SeekStart)

	// write original file to request
	original, err := writer.CreateFormFile("uploadfile", fmt.Sprintf("%s.%s", m.Sha1, m.Format))
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
	thumb, err := writer.CreateFormFile("uploadthumb", fmt.Sprintf("%s_thumb.%s", m.Sha1, "jpg"))
	if err != nil {
		log.WithFields(log.Fields{
			"type":  "thumbnail",
			"path":  filePath,
			"error": err.Error(),
		}).Error("could not create writer for formfile")
		return m, err
	}
	io.Copy(thumb, rt)

	// write username
	writer.WriteField("username", m.Creator)

	// close writer
	err = writer.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not close writer")
		return m, err
	}

	endpoint := fmt.Sprintf("%s/api/v1/file", node.APIEndpoint)
	contentType := writer.FormDataContentType()

	// executing request
	res, status, msg := _http.SendRequest(client, "POST", endpoint, node.Secret, body, contentType)
	if status > 0 {
		return m, errors.New(msg)
	}

	// preparing log fields
	logfields := log.Fields{
		"file":        m.Sha1,
		"thumb":       m.ThumbnailSha1,
		"node":        node.ID.Hex(),
		"status-code": res.StatusCode,
	}

	// handling response
	if res.StatusCode == 201 {
		log.WithFields(logfields).Info("media created on node")

		// add node to media
		m.NodeIDs = append(m.NodeIDs, node.ID)
		return m, nil
	} else {
		msg := "could not push media to node"
		log.WithFields(logfields).Error(msg)
		return m, errors.New(msg)
	}
}

func createThumbnail(filename string) io.Reader {
	// create file pointer
	r, _ := os.Open(filename)
	defer r.Close()
	// create thumbnail and receive pointer
	rt, _ := helper.Thumbnail(r, 128)
	return rt
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
