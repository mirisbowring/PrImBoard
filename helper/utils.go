package helper

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

// PathExists checks if a given path is available on the filesystem
func PathExists(path string) bool {
	if _, err := os.Stat("./conf/app.ini"); err != nil {
		if os.IsNotExist(err) {
			log.Debugf("checked path: %s - does not exist", path)
			return false
		} else {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Errorf("error occured while checking if <%s> exist", path)
			return false
		}
	}
	log.Debugf("checked path: %s - does exist", path)
	return true
}

// FindInSlice iterates over the slice and returns the position of the element if found
func FindInSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// GenerateSHA1 generates a sha1 hash for the specified reader
func GenerateSHA1(reader io.Reader) (string, error) {
	var _hash string
	hash := sha1.New()
	if _, err := io.Copy(hash, reader); err != nil {
		log.WithFields(log.Fields{
			"hash":  "sha1",
			"path":  "thumbnail",
			"error": err.Error(),
		}).Error("could not create hash")
		return _hash, err
	}
	hashInBytes := hash.Sum(nil)[:20]
	_hash = hex.EncodeToString(hashInBytes)
	return _hash, nil
}

// ReadContent reads the content from a reader into a byte array
func ReadContent(reader io.Reader) ([]byte, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("could not read file content from reader")
		return nil, err
	}
	return content, nil

}

// ReadFile reads a given file
func ReadFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not read file")
		return nil, err
	}
	return file, nil
}

// ReadFileAndContent reads a given file and reads its content into a byte array
func ReadFileAndContent(path string) (*os.File, []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not read file")
		return nil, nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		defer file.Close()
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("could not read file content")
		return nil, nil, err
	}
	return file, content, nil
}
