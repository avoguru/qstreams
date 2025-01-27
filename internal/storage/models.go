package storage

type QueryStream struct {
	StreamID         string `json:"stream_id"`
	Name             string `json:"name"`
	Query            string `json:"query"`
	BrokerURL        string `json:"broker_url"`
	DestinationType  string `json:"destination_type"`
	DestinationConfig string `json:"destination_config"`
	Interval         int    `json:"interval"`
	State            string `json:"state"`
	Dedupe           bool   `json:"dedupe"`
	DedupeDuration   int    `json:"dedupe_duration"`
}