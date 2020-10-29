package helper

import (
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
)

// ReadJSONConfig tries to open a specified config file
func ReadJSONConfig(config string) *json.Decoder {
	f, err := os.Open(config)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("could not open config file: %s", config)
	}
	defer f.Close()
	//decode file content into go object
	return json.NewDecoder(f)
}
