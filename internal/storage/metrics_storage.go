package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"qstreams/internal/models"
)

var metricsDirectory = "./metrics"

// LoadAllMetrics reads all metrics files from the metrics directory and returns them as a map.
func LoadAllMetrics() (map[string]models.StreamMetrics, error) {
	files, err := os.ReadDir(metricsDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(metricsDirectory, 0755); err != nil {
				return nil, err
			}
			return make(map[string]models.StreamMetrics), nil
		}
		return nil, err
	}

	metrics := make(map[string]models.StreamMetrics)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		streamName := file.Name()
		if filepath.Ext(streamName) == ".json" {
			streamName = streamName[:len(streamName)-len(filepath.Ext(streamName))]
			metricFile, err := os.Open(filepath.Join(metricsDirectory, file.Name()))
			if err != nil {
				continue
			}
			defer metricFile.Close()

			var streamMetrics models.StreamMetrics
			if err := json.NewDecoder(metricFile).Decode(&streamMetrics); err != nil {
				continue
			}

			metrics[streamName] = streamMetrics
		}
	}

	return metrics, nil
}

// SaveAllMetrics writes the current metrics to individual files in the metrics directory.
func SaveAllMetrics(metrics map[string]models.StreamMetrics) error {
	for streamName, streamMetrics := range metrics {
		filePath := filepath.Join(metricsDirectory, fmt.Sprintf("%s.json", streamName))
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create metrics file for stream '%s': %w", streamName, err)
		}
		defer file.Close()

		if err := json.NewEncoder(file).Encode(streamMetrics); err != nil {
			return fmt.Errorf("failed to write metrics for stream '%s': %w", streamName, err)
		}
	}
	return nil
}