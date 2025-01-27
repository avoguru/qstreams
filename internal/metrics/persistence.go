package metrics

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
	for uuid, streamMetrics := range metrics {
		filePath := filepath.Join(metricsDirectory, fmt.Sprintf("%s.json", uuid))
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create metrics file for stream '%s': %w", uuid, err)
		}
		defer file.Close()

		if err := json.NewEncoder(file).Encode(streamMetrics); err != nil {
			return fmt.Errorf("failed to write metrics for stream '%s': %w", uuid, err)
		}
	}
	return nil
}

// DeleteMetricsFile deletes the metrics file for a given stream ID
func DeleteMetricsFile(streamID string) error {
	filePath := filepath.Join(metricsDirectory, fmt.Sprintf("%s.json", streamID))
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, no error
		}
		return fmt.Errorf("failed to delete metrics file for stream '%s': %w", streamID, err)
	}
	return nil
}


// DeleteMetricsForStream deletes metrics for a specific stream
func DeleteMetricsForStream(streamID string) {
	Cache.Lock()
	defer Cache.Unlock()

	// Remove from in-memory cache
	delete(Cache.Data, streamID)

	// Delete metrics file from disk
	filePath := filepath.Join(metricsDirectory, fmt.Sprintf("%s.json", streamID))
	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Failed to delete metrics file for stream '%s': %v\n", streamID, err)
		}
	}
}