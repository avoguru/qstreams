package worker

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	// Set the stream state to "running"
	stream.State = "running"
	storage.SaveStream(stream) // Persist state to disk
	log.Printf("Stream '%s' is now active (state: 'running', StreamID: '%s').", stream.Name, stream.StreamID)

	ticker := time.NewTicker(time.Duration(stream.Pinot.QueryInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Stop if the stream is no longer in the "running" state
			if stream.State != "running" {
				log.Printf("Stream '%s' (StreamID: '%s') is no longer active (state: '%s').", stream.Name, stream.StreamID, stream.State)
				return
			}

			// Prepare the Pinot query payload
			queryPayload := map[string]string{"sql": stream.Pinot.Query}
			payload, _ := json.Marshal(queryPayload)

			// Create the HTTP request to Pinot
			req, err := http.NewRequest("POST", stream.Pinot.BrokerURL, bytes.NewBuffer(payload))
			if err != nil {
				log.Printf("Stream '%s' (StreamID: '%s'): Failed to create Pinot query request. Error: %v", stream.Name, stream.StreamID, err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			// Add Pinot authentication headers
			for key, value := range stream.Pinot.Authentication {
				req.Header.Set(key, value)
			}

			// Execute the query
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Stream '%s' (StreamID: '%s'): Failed to query Pinot. Error: %v", stream.Name, stream.StreamID, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Stream '%s' (StreamID: '%s'): Pinot query failed with status %d.", stream.Name, stream.StreamID, resp.StatusCode)
				continue
			}

			// Handle deduplication
			deduped := false
			if stream.Dedupe.Enabled {
				if skip := handleDeduplication(stream, payload); skip {
					deduped = true
				}
			}

			// Update metrics and push results to the destination
			metrics.Cache.Lock()
			metricsData := metrics.Cache.Data[stream.StreamID]
			metricsData.NumberOfQueries++

			if deduped {
				metricsData.EventsDeduped++
			} else {
				metricsData.EventsSent++
				if err := sendToDestination(dest, payload, stream.Destination.Authentication); err != nil {
					log.Printf("Stream '%s' (StreamID: '%s'): Failed to push data to destination. Error: %v", stream.Name, stream.StreamID, err)
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
		if cache.Hash == hash && now.Sub(cache.LastSent) <= time.Duration(min(stream.Dedupe.Duration, 60000))*time.Millisecond {
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

func sendToDestination(dest destinations.Destination, payload []byte, authHeaders map[string]string) error {
	// Create the HTTP request to the destination
	req, err := http.NewRequest("POST", dest.GetURL(), bytes.NewBuffer(payload)) 
	if err != nil {
		return fmt.Errorf("failed to create destination request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers for the destination
	for key, value := range authHeaders {
		req.Header.Set(key, value)
	}

	// Execute the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to destination: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("destination returned status %d", resp.StatusCode)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}