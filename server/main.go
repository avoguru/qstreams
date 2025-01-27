package main

import (
	"log"
	"net/http"
	"time"

	"qstreams/api"
	"qstreams/internal/core"
	"qstreams/internal/metrics"
)

func main() {
	log.Println("Starting qstreams Server...")

	// Restore metrics from disk
	if err := metrics.LoadMetrics(); err != nil {
		log.Fatalf("Failed to restore metrics: %v", err)
	}

	// Start periodic metrics flushing
	go metrics.SaveMetricsFlush(30 * time.Second)

	// Restore streams
	if err := core.RestoreStreams(); err != nil {
		log.Fatalf("Failed to restore streams: %v", err)
	}

	// Serve static files from the "console" folder
	http.Handle("/console/", http.StripPrefix("/console/", http.FileServer(http.Dir("./console"))))

	// Initialize API routes
	router := api.InitRoutes()
	http.Handle("/", router)

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}