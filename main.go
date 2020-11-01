package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	gateway "github.com/mirisbowring/PrImBoard/go"
	"github.com/mirisbowring/PrImBoard/helper"
	"github.com/mirisbowring/PrImBoard/helper/models"
	node "github.com/mirisbowring/PrImBoard/node-api"
)

func main() {
	api := flag.Bool("api", true, "Start api gateway")
	node := flag.Bool("node", false, "Start node instance")
	config := flag.String("config", "env.json", "specify config file (default: env.json)")
	logLevel := flag.String("log", "info", `Set Log Level
	debug - useful debugging information
	info - (default) everything noteworthy
	warn - only warnings are displayed - have an eye on them
	error - only errors are shown (system is still stable)
	fatal - only log the fatal message before quitting`)
	help := flag.Bool("h", false, "shows the help page")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	log.SetOutput(os.Stdout)

	switch *logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		break
	case "info":
		log.SetLevel(log.InfoLevel)
		break
	case "warn":
		log.SetLevel(log.WarnLevel)
		break
	case "error":
		log.SetLevel(log.ErrorLevel)
		break
	case "fatal":
		log.SetLevel(log.FatalLevel)
		break
	default:
		log.Fatal("unknown log level specified")
	}

	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	conf := helper.ParseConfig(*config)

	if *node {
		startNode(conf)
	}

	if *api {
		startAPIGateway(conf)
	}
}

func startAPIGateway(config models.Config) {
	log.Info("starting api gateway")

	a := gateway.App{}
	a.Initialize(config.APIGatewayConfig)
	a.Run(":" + strconv.Itoa(a.Config.Port))
}

func startNode(config models.Config) {
	log.Info("starting node")

	a := node.App{}
	a.Initialize(config.NodeConfig)
	a.Run(":" + strconv.Itoa(a.Config.Port))
}
