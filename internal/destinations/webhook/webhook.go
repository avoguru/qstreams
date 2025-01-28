package webhook

import (
	"bytes"
	"fmt"
	"net/http"
)

type Webhook struct {
	URL string
}

func NewWebhook(url string) *Webhook {
	return &Webhook{URL: url}
}

func (w *Webhook) Send(data []byte) error {
	resp, err := http.Post(w.URL, "application/json", bytes.NewBuffer(data))
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
	if w.URL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	return nil
}

func (w *Webhook) GetURL() string {
	return w.URL
}