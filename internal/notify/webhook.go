package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier sends notifications to a generic webhook endpoint
type WebhookNotifier struct {
	url        string
	httpClient *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		url:        url,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends a notification via webhook POST
func (w *WebhookNotifier) Send(ctx context.Context, n Notification) error {
	body, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook notification: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
