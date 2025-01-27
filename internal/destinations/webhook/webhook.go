package webhook

import (
	"bytes"
	"fmt"
	"net/http"
)

type Webhook struct {
	Endpoint string
}

func NewWebhook(endpoint string) *Webhook {
	return &Webhook{Endpoint: endpoint}
}

func (w *Webhook) Send(data []byte) error {
	resp, err := http.Post(w.Endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send data to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook responded with status: %d", resp.StatusCode)
	}
	return nil
}

func (w *Webhook) Validate() error {
	if w.Endpoint == "" {
		return fmt.Errorf("webhook endpoint cannot be empty")
	}
	return nil
}