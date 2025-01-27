package core

import (
	"log"
	"sync"
	"time"

	"qstreams/internal/metrics"
	"qstreams/internal/models"
	"qstreams/internal/storage"
)

var MetricsCache = struct {
	sync.Mutex
	Data map[string]models.StreamMetrics
}{
	Data: make(map[string]models.StreamMetrics),
}

// LoadMetrics loads all metrics from the metrics store (disk) into the cache.
func LoadMetrics() error {
	metrics, err := storage.LoadAllMetrics()
	if err != nil {
		return err
	}

	MetricsCache.Lock()
	defer MetricsCache.Unlock()
	MetricsCache.Data = metrics
	log.Printf("Loaded metrics for %d stream(s) from state store.", len(metrics))
	return nil
}

// SaveMetricsFlush periodically flushes the in-memory metrics to disk.
func SaveMetricsFlush(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		MetricsCache.Lock()
		data := make(map[string]models.StreamMetrics)
		for k, v := range MetricsCache.Data {
			data[k] = v
		}
		MetricsCache.Unlock()

		if err := storage.SaveAllMetrics(data); err != nil {
			log.Printf("Failed to flush metrics to disk: %v", err)
		} else {
			log.Println("Metrics successfully flushed to disk.")
		}
	}
}

func DeleteMetricsForStream(streamID string) {
	metrics.Cache.Lock()
	defer metrics.Cache.Unlock()

	delete(metrics.Cache.Data, streamID)
	err := metrics.DeleteMetricsFile(streamID)
	if err != nil {
		log.Printf("Failed to delete metrics file for stream '%s': %v", streamID, err)
	}
}