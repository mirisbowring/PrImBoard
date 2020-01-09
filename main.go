package main

import (
	sw "github.com/mirisbowring/PrImBoard/go"
	"log"
	"strconv"
)

func main() {
	log.Print("Server starting...")

	a := sw.App{}
	a.Initialize()
	a.Run(":" + strconv.Itoa(a.Config.Port))
}
