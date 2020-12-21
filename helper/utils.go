package helper

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PathExists checks if a given path is available on the filesystem
func PathExists(path string) bool {
	logfields := log.Fields{"path": path}
	defaultMsg := "checked path existance"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			logfields["exist"] = false
			log.WithFields(logfields).Debug(defaultMsg)
			return false
		} else {
			logfields["error"] = err.Error()
			log.WithFields(logfields).Error("error occured during path existance check")
			return false
		}
	}
	logfields["exist"] = true
	log.WithFields(logfields).Debug(defaultMsg)
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
func GenerateSHA1(reader io.Reader) string {
	var _hash string
	hash := sha1.New()
	if _, err := io.Copy(hash, reader); err != nil {
		log.WithFields(log.Fields{
			"hash":  "sha1",
			"error": err.Error(),
		}).Error("could not create hash")
		return _hash
	}
	hashInBytes := hash.Sum(nil)[:20]
	_hash = hex.EncodeToString(hashInBytes)
	return _hash
}

// GenerateRandomToken generates a token of the specified length
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// ObjectIDIntersect checks whether slice a contains any element of slice b vice versa
func ObjectIDIntersect(a []primitive.ObjectID, b []primitive.ObjectID) bool {
	if a == nil || b == nil {
		return false
	}
	for _, id := range a {
		for _, id2 := range b {
			if id == id2 {
				return true
			}
		}
	}
	return false
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

// UniqueStrings removes all duplicates from a string slice and returns the result
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
