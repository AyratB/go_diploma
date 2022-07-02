package main

import (
	"github.com/AyratB/go_diploma/internal/server"
	"github.com/AyratB/go_diploma/internal/utils"
	"log"
)

func main() {

	configs := utils.GetConfigs()

	resourcesCloser, err := server.Run(configs)
	defer func() {
		if resourcesCloser != nil {
			resourcesCloser()
		}
	}()

	if err != nil {
		log.Fatal(err)
	}
}
