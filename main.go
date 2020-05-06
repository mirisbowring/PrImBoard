package main

import (
	"log"
	"strconv"

	sw "github.com/mirisbowring/PrImBoard/go"
)

func main() {
	log.Print("Server starting...")

	a := sw.App{}
	a.Initialize()
	a.Run(":" + strconv.Itoa(a.Config.Port))
}
