package storage

type QueryStream struct {
	StreamID    string           `json:"stream_id"`
	Name        string           `json:"name"`
	Pinot       PinotConfig      `json:"pinot"`
	Destination DestinationConfig `json:"destination"`
	Dedupe      DedupeConfig     `json:"dedupe"`
	State       string           `json:"state"` // Add this field to track stream state
}

type PinotConfig struct {
	Query          string            `json:"query"`
	BrokerURL      string            `json:"broker_url"`
	QueryInterval  int               `json:"query_interval"`
	Authentication map[string]string `json:"authentication"`
}

type DestinationConfig struct {
	Type           string            `json:"type"`
	URL            string            `json:"url"`
	Authentication map[string]string `json:"authentication"`
}

type DedupeConfig struct {
	Enabled  bool `json:"enabled"`
	Duration int  `json:"duration"`
}