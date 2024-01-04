package main

import (
	"fmt"

	"os"

	"github.com/rs/zerolog/log"

	api "github.com/Damione1/thread-art-generator/pkg/grpcApi"
	"github.com/Damione1/thread-art-generator/pkg/util"
)

func startWorkerService() {
	// Initialize and start your worker service here
	fmt.Println("Starting worker service...")
	// ...
}

func main() {

	config, err := util.LoadConfig(".env")
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}

	if len(os.Args) < 2 {
		fmt.Println("No service specified, exiting...")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "api":
		api.RunAPI(config)
		//startAPIService()
	case "worker":
		startWorkerService()
	default:
		fmt.Printf("Unknown service: %s\n", os.Args[1])
		os.Exit(1)
	}
}
