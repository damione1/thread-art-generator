package main

import (
	"fmt"
	"os"

	api "github.com/Damione1/thread-art-generator/pkg/grpcApi"
)

func startAPIService() {
	// Initialize and start your API service here
	api.RunAPI()
	// ...
}

func startWorkerService() {
	// Initialize and start your worker service here
	fmt.Println("Starting worker service...")
	// ...
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No service specified, exiting...")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "api":
		startAPIService()
	case "worker":
		startWorkerService()
	default:
		fmt.Printf("Unknown service: %s\n", os.Args[1])
		os.Exit(1)
	}
}
