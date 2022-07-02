package main

import (
	"github.com/AyratB/go_diploma/internal/server"
	"github.com/AyratB/go_diploma/internal/utils"
	"log"
)

func main() {
	resourcesCloser, err := server.Run(utils.GetConfigs())
	defer func() {
		if resourcesCloser != nil {
			resourcesCloser()
		}
	}()

	if err != nil {
		resourcesCloser()
		log.Fatal(err)
	}
}
