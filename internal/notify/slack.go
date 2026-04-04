package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier sends notifications to Slack via webhook
type SlackNotifier struct {
	webhookURL string
	channel    string
	httpClient *http.Client
}

// NewSlackNotifier creates a SlackNotifier
func NewSlackNotifier(webhookURL, channel string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		channel:    channel,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type slackPayload struct {
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text"`
}

// Send sends a notification to Slack
func (s *SlackNotifier) Send(ctx context.Context, n Notification) error {
	var emoji string
	switch n.Result {
	case "fail":
		emoji = "x"
	case "partial":
		emoji = "warning"
	default:
		emoji = "white_check_mark"
	}

	text := fmt.Sprintf(":%s: *Kymaros* — `%s` score: *%d/100* (%s)\n%s",
		emoji, n.TestName, n.Score, n.Result, n.Message)

	payload := slackPayload{
		Channel: s.channel,
		Text:    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send slack notification: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}
