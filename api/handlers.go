package api

import (
	"encoding/json"
	"net/http"
	"qstreams/internal/core"
	"qstreams/internal/metrics"
	"qstreams/internal/models"
	"qstreams/internal/storage"

	"github.com/gorilla/mux"
)

// CreateStreamHandler creates a new stream
func CreateStreamHandler(w http.ResponseWriter, r *http.Request) {
	var stream storage.QueryStream
	err := json.NewDecoder(r.Body).Decode(&stream)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validation for deduplication settings
	if !stream.Dedupe {
		if stream.DedupeDuration > 0 {
			http.Error(w, "dedupe_duration cannot be set if dedupe is false or missing", http.StatusBadRequest)
			return
		}
		stream.DedupeDuration = 0
	} else {
		if stream.DedupeDuration < 1000 {
			stream.DedupeDuration = 1000
		} else if stream.DedupeDuration > 60000 {
			stream.DedupeDuration = 60000
		}
	}

	// Create the stream (generate a StreamID internally)
	err = core.CreateStream(&stream)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created stream's StreamID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Stream created successfully",
		"stream_id": stream.StreamID,
	})
}

// StartStreamHandler starts a stopped stream
func StartStreamHandler(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]
	stream, err := storage.LoadStream(streamID)
	if err != nil {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	if stream.State == "running" {
		http.Error(w, "Stream is already running", http.StatusBadRequest)
		return
	}

	stream.State = "running"
	if err := storage.SaveStream(stream); err != nil {
		http.Error(w, "Failed to start stream", http.StatusInternalServerError)
		return
	}

	// Optionally restart the worker if needed
	go core.RestartStreamWorker(stream)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Stream started successfully",
		"stream_id": stream.StreamID,
	})
}

// StopStreamHandler stops a running stream
func StopStreamHandler(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]
	stream, err := storage.LoadStream(streamID)
	if err != nil {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	if stream.State != "running" {
		http.Error(w, "Stream is not running", http.StatusBadRequest)
		return
	}

	stream.State = "stopped"
	if err := storage.SaveStream(stream); err != nil {
		http.Error(w, "Failed to stop stream", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Stream stopped successfully",
		"stream_id": stream.StreamID,
	})
}

// UpdateStreamHandler updates an existing stream
func UpdateStreamHandler(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]
	stream, err := storage.LoadStream(streamID)
	if err != nil {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	var updatedStream storage.QueryStream
	if err := json.NewDecoder(r.Body).Decode(&updatedStream); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Merge updated fields
	stream.Name = updatedStream.Name
	stream.Query = updatedStream.Query
	stream.BrokerURL = updatedStream.BrokerURL
	stream.DestinationType = updatedStream.DestinationType
	stream.DestinationConfig = updatedStream.DestinationConfig
	stream.Interval = updatedStream.Interval
	stream.Dedupe = updatedStream.Dedupe
	stream.DedupeDuration = updatedStream.DedupeDuration

	// Save the updated stream
	if err := storage.SaveStream(stream); err != nil {
		http.Error(w, "Failed to update stream", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Stream updated successfully",
		"stream_id": stream.StreamID,
	})
}

// DeleteStreamHandler deletes an existing stream
func DeleteStreamHandler(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]
	filePath := storage.GetStreamFilePath(streamID)

	if err := storage.DeleteStreamFile(filePath); err != nil {
		http.Error(w, "Failed to delete stream", http.StatusInternalServerError)
		return
	}

	// Optionally clean up metrics
	metrics.DeleteMetricsForStream(streamID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Stream deleted successfully",
		"stream_id": streamID,
	})
}

// ListStreamsHandler lists all existing streams
func ListStreamsHandler(w http.ResponseWriter, r *http.Request) {
	streams, err := storage.ListStreams()
	if err != nil {
		http.Error(w, "Failed to list streams", http.StatusInternalServerError)
		return
	}

	// Directly encode the list of QueryStream models
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"streams": streams,
	})
}

// MetricsHandler handles the /metrics endpoint to expose metrics for all streams
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics.Cache.Lock()
	defer metrics.Cache.Unlock()

	// Transform the map of metrics to an array of StreamMetrics
	var response struct {
		Streams []models.StreamMetrics `json:"streams"`
	}

	for streamID, metricsData := range metrics.Cache.Data {
		response.Streams = append(response.Streams, models.StreamMetrics{
			StreamID:       streamID,
			EventsSent:     metricsData.EventsSent,
			EventsDeduped:  metricsData.EventsDeduped,
			NumberOfQueries: metricsData.NumberOfQueries,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}