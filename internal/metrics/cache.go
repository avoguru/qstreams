package metrics

import (
	"log"
	"sync"
	"time"

	"qstreams/internal/models"
	"qstreams/internal/storage"
)

var Cache = struct {
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

	Cache.Lock()
	defer Cache.Unlock()
	Cache.Data = metrics
	log.Printf("Loaded metrics for %d stream(s) from disk.", len(metrics))
	return nil
}

// SaveMetricsFlush periodically flushes the in-memory metrics to disk.
func SaveMetricsFlush(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		Cache.Lock()
		data := make(map[string]models.StreamMetrics)
		for k, v := range Cache.Data {
			data[k] = v
		}
		Cache.Unlock()

		if err := storage.SaveAllMetrics(data); err != nil {
			log.Printf("Failed to flush metrics to disk: %v", err)
		} else {
			log.Println("Metrics successfully flushed to disk.")
		}
	}
}