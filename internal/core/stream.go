package core

import (
	"fmt"
	"log"
	"qstreams/internal/destinations"
	"qstreams/internal/destinations/webhook"
	"qstreams/internal/storage"
	"qstreams/internal/worker"

	"github.com/google/uuid"
)

func NewDestination(destType, destConfig string) (destinations.Destination, error) {
	switch destType {
	case "webhook":
		dest := webhook.NewWebhook(destConfig)
		if err := dest.Validate(); err != nil {
			return nil, fmt.Errorf("invalid webhook configuration: %w", err)
		}
		return dest, nil
	default:
		return nil, fmt.Errorf("unsupported destination type: %s", destType)
	}
}

// CreateStream initializes and saves a new stream with a unique UUID
func CreateStream(stream *storage.QueryStream) error {
	// Generate a unique StreamID for the stream
	stream.StreamID = uuid.New().String()

	// Log the creation of the stream
	log.Printf("Stream '%s' created with ID: %s", stream.Name, stream.StreamID)

	// Save the stream to the state store
	if err := storage.SaveStream(stream); err != nil {
		return fmt.Errorf("failed to save stream: %w", err)
	}

	// Create and validate the destination
	dest, err := NewDestination(stream.DestinationType, stream.DestinationConfig)
	if err != nil {
		return fmt.Errorf("failed to create destination for stream '%s': %w", stream.StreamID, err)
	}

	// Start the worker for the stream
	go worker.RunStreamWorker(stream, dest)

	return nil
}

func RestoreStreams() error {
	// Load all streams from the state store
	streams, err := storage.ListStreams()
	if err != nil {
		return fmt.Errorf("failed to list streams from the state store: %w", err)
	}

	log.Printf("Found %d stream(s) in the state store. Beginning restoration process...", len(streams))

	for _, stream := range streams {
		switch stream.State {
		case "submitted", "creating", "running":
			log.Printf("Initializing stream '%s' (state: %s). Transitioning to 'running' state...", stream.StreamID, stream.State)
			stream.State = "running" // Move to running state

			// Create and validate destination
			dest, err := NewDestination(stream.DestinationType, stream.DestinationConfig)
			if err != nil {
				log.Printf("Failed to initialize stream '%s'. Error: %v", stream.StreamID, err)
				continue
			}

			// Start the worker for the stream
			go worker.RunStreamWorker(&stream, dest)

		case "stopped":
			log.Printf("Skipping stream '%s'. Current state: 'stopped'. Stream will remain inactive.", stream.StreamID)

		default:
			log.Printf("Unknown state for stream '%s'. Current state: '%s'. Skipping.", stream.StreamID, stream.State)
		}
	}

	log.Println("Stream restoration process completed.")
	return nil
}

// RestartStreamWorker restarts a worker for a stream
func RestartStreamWorker(stream *storage.QueryStream) {
	// Create and validate the destination
	dest, err := NewDestination(stream.DestinationType, stream.DestinationConfig)
	if err != nil {
		log.Printf("Failed to restart stream '%s': invalid destination configuration. Error: %v", stream.StreamID, err)
		return
	}

	// Start the worker
	go worker.RunStreamWorker(stream, dest)
	log.Printf("Stream '%s' worker restarted.", stream.StreamID)
}