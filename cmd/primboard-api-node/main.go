package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/mirisbowring/primboard/helper"
	"github.com/mirisbowring/primboard/internal/models/infrastructure"
	"github.com/mirisbowring/primboard/node"
	log "github.com/sirupsen/logrus"
)

func main() {
	config := flag.String("config", "env.json", "specify config file (default: env.json)")
	env := flag.Bool("env", false, "specify whether to use environment variables or not")
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

	var conf infrastructure.Config
	if *env {
		conf = helper.ParseNodeEnv()
	} else {
		conf = helper.ParseConfig(*config)
	}

	startNode(conf)
}

func startNode(config infrastructure.Config) {
	log.Info("starting node")

	a := node.AppNode{}
	a.Initialize(config.NodeConfig)
	a.Run(":" + strconv.Itoa(a.Config.Port))
}
