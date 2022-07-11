package main

import (
	"context"
	"github.com/AyratB/go_diploma/internal/server"
	"github.com/AyratB/go_diploma/internal/utils"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configs := utils.GetConfigs()

	err := server.Run(configs, ctx)
	if err != nil {
		log.Fatal(err)
	}
}
