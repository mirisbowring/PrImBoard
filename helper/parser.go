package helper

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/mirisbowring/primboard/internal/models/infrastructure"
	log "github.com/sirupsen/logrus"
)

// ParseConfig parses the config file into the config object
func ParseConfig(config string) infrastructure.Config {
	var tmp infrastructure.Config
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

// ParseGatewayEnv parses the Gateway configuration from the environment
func ParseGatewayEnv() infrastructure.Config {
	var tmp infrastructure.Config
	var err error
	tmp.APIGatewayConfig.AllowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ";")
	tmp.APIGatewayConfig.Certificates = os.Getenv("CERTIFICATES")
	tmp.APIGatewayConfig.CaCert = os.Getenv("CA_CERT")
	tmp.APIGatewayConfig.CookieHTTPOnly, err = strconv.ParseBool(os.Getenv("COOKIE_HTTP_ONLY"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "COOKIE_HTTP_ONLY",
			"value": os.Getenv("COOKIE_HTTP_ONLY"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.CookieDomain = os.Getenv("COOKIE_DOMAIN")
	tmp.APIGatewayConfig.CookiePath = os.Getenv("COOKIE_PATH")
	tmp.APIGatewayConfig.CookieSameSite, err = strconv.Atoi(os.Getenv("COOKIE_SAME_SITE"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "COOKIE_SAME_SITE",
			"value": os.Getenv("COOKIE_SAME_SITE"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.CookieSecure, err = strconv.ParseBool(os.Getenv("COOKIE_SECURE"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "COOKIE_SECURE",
			"value": os.Getenv("COOKIE_SECURE"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.CookieTokenTitle = os.Getenv("COOKIE_TOKEN_TITLE")
	tmp.APIGatewayConfig.DBName = os.Getenv("DATABASE_NAME")
	tmp.APIGatewayConfig.DefaultMediaPageSize, err = strconv.Atoi(os.Getenv("DEFAULT_MEDIA_PAGE_SIZE"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "DEFAULT_MEDIA_PAGE_SIZE",
			"value": os.Getenv("DEFAULT_MEDIA_PAGE_SIZE"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.Domain = os.Getenv("DOMAIN")
	tmp.APIGatewayConfig.HTTP, err = strconv.ParseBool(os.Getenv("HTTP"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "HTTP",
			"value": os.Getenv("HTTP"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.InviteValidity, err = strconv.Atoi(os.Getenv("INVITE_VALIDITY"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "INVITE_VALIDITY",
			"value": os.Getenv("INVITE_VALIDITY"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.MongoURL = os.Getenv("MONGO_URL")
	tmp.APIGatewayConfig.Port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "PORT",
			"value": os.Getenv("PORT"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.SessionRotation, err = strconv.ParseBool(os.Getenv("SESSION_ROTATION"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "SESSION_ROTATION",
			"value": os.Getenv("SESSION_ROTATION"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.TagPreviewLimit, err = strconv.Atoi(os.Getenv("TAG_PREVIEW_LIMIT"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "TAG_PREVIEW_LIMIT",
			"value": os.Getenv("TAG_PREVIEW_LIMIT"),
			"error": err.Error(),
		}).Error("could not parse env")
	}
	tmp.APIGatewayConfig.TLSInsecure, err = strconv.ParseBool(os.Getenv("TLS_INSECURE"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "TLS_INSECURE",
			"value": os.Getenv("TLS_INSECURE"),
			"error": err.Error(),
		}).Error("could not parse env")
	}

	return tmp
}

// ParseNodeEnv parses the config file into the config object
func ParseNodeEnv() infrastructure.Config {
	var tmp infrastructure.Config
	var err error

	tmp.NodeConfig.AllowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ";")
	tmp.NodeConfig.BasePath = os.Getenv("BASEPATH")
	tmp.NodeConfig.TargetPath = os.Getenv("TARGETPATH")
	tmp.NodeConfig.GatewayURL = os.Getenv("GATEWAY_URL")
	tmp.NodeConfig.NodeAuth = &infrastructure.NodeAuth{}
	tmp.NodeConfig.NodeAuth.ID = os.Getenv("NODE_AUTH_ID")
	tmp.NodeConfig.NodeAuth.Secret = os.Getenv("NODE_AUTH_SECRET")
	tmp.NodeConfig.Port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.WithFields(log.Fields{
			"env":   "PORT",
			"value": os.Getenv("PORT"),
			"error": err.Error(),
		}).Error("could not parse env")
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
