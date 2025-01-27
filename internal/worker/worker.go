package worker

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"qstreams/internal/destinations"
	"qstreams/internal/metrics"
	"qstreams/internal/storage"
)

type dedupeCache struct {
	Hash      string
	LastSent  time.Time
}

var dedupeStore = struct {
	sync.Mutex
	Cache map[string]dedupeCache
}{Cache: make(map[string]dedupeCache)}

func RunStreamWorker(stream *storage.QueryStream, dest destinations.Destination) {
	stream.State = "running"
	storage.SaveStream(stream) // Persist state to disk
	log.Printf("Stream '%s' is now active (state: 'running', StreamID: '%s').", stream.Name, stream.StreamID)

	ticker := time.NewTicker(time.Duration(stream.Interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if stream.State != "running" {
				log.Printf("Stream '%s' (UUID: '%s') is no longer active (state: '%s').", stream.Name, stream.StreamID, stream.State)
				return
			}

			// Simulate query execution
			data := fmt.Sprintf("Results for query: %s", stream.Query)
			payload, _ := json.Marshal(map[string]string{"data": data})

			// Update metrics
			metrics.Cache.Lock()
			metricsData := metrics.Cache.Data[stream.StreamID]
			metricsData.NumberOfQueries++
			metrics.Cache.Data[stream.StreamID] = metricsData
			metrics.Cache.Unlock()

			// Handle deduplication
			deduped := false
			if stream.Dedupe {
				if skip := handleDeduplication(stream, payload); skip {
					deduped = true
				}
			}

			// Update sent or deduped metrics
			metrics.Cache.Lock()
			if deduped {
				metricsData.EventsDeduped++
			} else {
				metricsData.EventsSent++
				if err := dest.Send(payload); err != nil {
					log.Printf("Stream '%s' (UUID: '%s'): Failed to push data to destination. Error: %v", stream.Name, stream.StreamID, err)
				}
			}
			metrics.Cache.Data[stream.StreamID] = metricsData
			metrics.Cache.Unlock()
		}
	}
}

func handleDeduplication(stream *storage.QueryStream, payload []byte) bool {
	// Compute hash of the payload
	hash := fmt.Sprintf("%x", sha256.Sum256(payload))

	dedupeStore.Lock()
	defer dedupeStore.Unlock()

	cache, exists := dedupeStore.Cache[stream.Name]
	now := time.Now()

	if exists {
		// If hash is the same and within dedupe_duration, skip sending
		if cache.Hash == hash && now.Sub(cache.LastSent) <= time.Duration(min(stream.DedupeDuration, 60000))*time.Millisecond {
			log.Printf("Stream '%s': Duplicate data detected. Skipping push.", stream.StreamID)
			return true
		}
	}

	// Update dedupe cache with the new hash and timestamp
	dedupeStore.Cache[stream.Name] = dedupeCache{
		Hash:     hash,
		LastSent: now,
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}