package primboard

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	ipfs "github.com/ipfs/go-ipfs-api"
	h "github.com/mirisbowring/PrImBoard/helper"
)

// authCookie stores the temporal cookie object
var authCookie *http.Cookie

// DecodeMediaRequest decodes the api request into the passed object
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediaRequest(w http.ResponseWriter, r *http.Request, m Media) (Media, int) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return Media{}, 1
	}
	defer r.Body.Close()
	return m, 0
}

// DecodeMediasRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediasRequest(w http.ResponseWriter, r *http.Request) ([]Media, int) {
	var m []Media
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return nil, 1
	}
	defer r.Body.Close()
	return m, 0
}

// DecodeMediaGroupMapRequest decodes the api request into the passed slice
// responds with decode error if occurs
// status 0 => ok || status 1 => error
func DecodeMediaGroupMapRequest(w http.ResponseWriter, r *http.Request) (MediaGroupMap, int) {
	var mgm MediaGroupMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mgm); err != nil {
		// an decode error occured
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return mgm, 1
	}
	defer r.Body.Close()
	return mgm, 0
}

// addMediaToIpfsNode uploads the given file to the specified ipfs node.
// The passed media model will be completed with path and hashes.
func addMediaToIpfsNode(file string, media Media, node Node) (Media, error) {
	// new ipfs shell
	sh := ipfs.NewShell(node.Address + ":" + strconv.Itoa(node.IPFSAPIPort))
	// create file pointer
	r, _ := os.Open(file)
	// create thumbnail and receive pointer
	rt, _ := h.Thumbnail(r, 128)
	// add the thumbnail to the ipfs
	thumbCid, err := sh.Add(rt)
	if err != nil {
		log.Println(err)
		return Media{}, errors.New("could not upload thumbnail to ipfs node")
	}
	r.Close()

	//recreate file pointer (add is manipulating it)
	r, _ = os.Open(file)
	// add the file to ipfs
	// do not use the recursive AddDir because we need to add all the files to the mongo
	cid, err := sh.Add(r)
	if err != nil {
		log.Println(err)
		return Media{}, errors.New("could not upload file to ipfs node")
	}
	r.Close()

	// if successfull, create a media object with the returned ipfs url
	// var m Media
	// if (src.Meta != thumbnailer.Meta{} && src.Meta.Title != "") {
	// 	m.Title = src.Meta.Title
	// }
	media.Sha1 = cid
	media.URL = node.IPFSGateway + cid
	media.URLThumb = node.IPFSGateway + thumbCid
	// // eval mime to generic type
	// if src.HasVideo {
	// 	m.Type = "video"
	// } else if src.HasAudio {
	// 	m.Type = "audio"
	// } else {
	// 	m.Type = "image"
	// }
	// m.Format = src.Extension
	// encode the object to json
	// b := new(bytes.Buffer)
	// json.NewEncoder(b).Encode(m)
	// // post the object to the api
	// post("http://"+PrimboardHost+"/api/v1/media", "application/json", b)
	return media, nil
}

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

func removeMediaFromIpfs(media Media, node Node) error {
	// new ipfs shell
	sh := ipfs.NewShell(node.Address + ":" + strconv.Itoa(node.IPFSAPIPort))
	// unpin from node
	sh.Unpin(media.Sha1)
	// should implement repo gc
	// not available on ipfs-api at moment
	return nil
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
