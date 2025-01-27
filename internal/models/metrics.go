package models

type StreamMetrics struct {
	StreamID       string `json:"stream_id"`
	EventsSent     int    `json:"events_sent"`
	EventsDeduped  int    `json:"events_deduped"`
	NumberOfQueries int    `json:"number_of_queries"`
}