package helper

import (
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mirisbowring/PrImBoard/helper/models"
)

// ParseConfig parses the config file into the config object
func ParseConfig(config string) models.Config {
	var tmp models.Config
	json, f := ReadJSONFile(config)
	defer f.Close()
	if err := json.Decode(&tmp); err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"error":  err.Error(),
		}).Fatal("could not parse config file")
	}
	return tmp
}

// ReadJSONFile tries to open a specified config file
func ReadJSONFile(config string) (*json.Decoder, *os.File) {
	f, err := os.Open(config)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"error":  err.Error(),
		}).Fatal("could not open config file")
	}
	// defer f.Close()
	//decode file content into go object
	return json.NewDecoder(f), f
}
