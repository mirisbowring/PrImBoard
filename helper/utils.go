package helper

import (
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
