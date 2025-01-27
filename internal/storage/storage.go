package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var streamDirectory = "./streams"

// SaveStream writes a stream's configuration to a file using its StreamID
func SaveStream(stream *QueryStream) error {
	filePath := filepath.Join(streamDirectory, fmt.Sprintf("%s.json", stream.StreamID))
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create stream file: %w", err)
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(stream)
}

// LoadStream reads a stream's configuration from its file by StreamID
func LoadStream(streamID string) (*QueryStream, error) {
	filePath := filepath.Join(streamDirectory, fmt.Sprintf("%s.json", streamID))
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream file: %w", err)
	}
	defer file.Close()

	var stream QueryStream
	if err := json.NewDecoder(file).Decode(&stream); err != nil {
		return nil, fmt.Errorf("failed to decode stream file: %w", err)
	}

	return &stream, nil
}

// ListStreams reads all stream configurations in the state store
func ListStreams() ([]QueryStream, error) {
	files, err := os.ReadDir(streamDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(streamDirectory, 0755); err != nil {
				return nil, fmt.Errorf("failed to create stream directory: %w", err)
			}
			return []QueryStream{}, nil
		}
		return nil, err
	}

	var streams []QueryStream
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(streamDirectory, file.Name())
		stream, err := LoadStream(file.Name()[:len(file.Name())-len(".json")])
		if err != nil {
			log.Printf("Skipping invalid stream file: %s", filePath)
			continue
		}
		streams = append(streams, *stream)
	}

	return streams, nil
}

// GetStreamFilePath returns the file path for a given stream ID
func GetStreamFilePath(streamID string) string {
	return filepath.Join(streamDirectory, fmt.Sprintf("%s.json", streamID))
}

// DeleteStreamFile deletes a stream file by its file path
func DeleteStreamFile(filePath string) error {
	return os.Remove(filePath)
}