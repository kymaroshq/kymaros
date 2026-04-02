package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlackNotifierSend(t *testing.T) {
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		received = buf
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL, "#ops")
	err := notifier.Send(context.Background(), Notification{
		TestName: "prod-test",
		Score:    85,
		Result:   "partial",
		Message:  "postgres TCP check failed",
	})
	require.NoError(t, err)

	var payload map[string]string
	err = json.Unmarshal(received, &payload)
	require.NoError(t, err)
	assert.Equal(t, "#ops", payload["channel"])
	assert.Contains(t, payload["text"], "prod-test")
	assert.Contains(t, payload["text"], "85/100")
	assert.Contains(t, payload["text"], "warning")
}

func TestSlackNotifierError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL, "#ops")
	err := notifier.Send(context.Background(), Notification{TestName: "test", Score: 0, Result: "fail"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestWebhookNotifierSend(t *testing.T) {
	var received Notification
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier(server.URL)
	err := notifier.Send(context.Background(), Notification{
		TestName:  "my-test",
		Score:     100,
		Result:    "pass",
		ReportRef: "my-test-20260327",
		Message:   "all checks passed",
	})
	require.NoError(t, err)
	assert.Equal(t, "my-test", received.TestName)
	assert.Equal(t, 100, received.Score)
	assert.Equal(t, "pass", received.Result)
}

func TestWebhookNotifierError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier(server.URL)
	err := notifier.Send(context.Background(), Notification{TestName: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 502")
}
