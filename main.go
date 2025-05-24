package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/steezeburger/storage-shower/internal/scan"
	"github.com/steezeburger/storage-shower/internal/server"
)

//go:embed web
var webFS embed.FS

var debugMode = false

// Debug flag to control verbose logging

func main() {
	// Parse command line flags
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flag.Parse()

	// Set debug mode for scan package
	scan.DebugMode = debugMode

	// Debug mode enables verbose logging
	if debugMode {
		log.Printf("Debug mode enabled")
	}

	// Start server with embedded web files
	port := server.StartServer(webFS)

	// Log server information
	log.Printf("Server started on port %d", port)
	if debugMode {
		log.Printf("Debug mode enabled")
	}

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalChan
	log.Printf("Received termination signal, shutting down...")
}
