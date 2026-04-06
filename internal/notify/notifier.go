package notify

import "context"

// Notification contains the data to send
type Notification struct {
	TestName  string
	Score     int
	Result    string
	ReportRef string
	Message   string
}

// Notifier is the interface for sending notifications
type Notifier interface {
	Send(ctx context.Context, n Notification) error
}
