package notify

import "fmt"

// NotifierFactory creates a Notifier.
type NotifierFactory func() Notifier

// notifierRegistry holds all registered notifier factories.
var notifierRegistry = map[string]NotifierFactory{}

// RegisterNotifier adds a new notifier factory (used by external packages via init()).
func RegisterNotifier(name string, factory NotifierFactory) {
	notifierRegistry[name] = factory
}

// GetNotifier returns the notifier for the given name.
func GetNotifier(name string) (Notifier, error) {
	factory, ok := notifierRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown notifier %q", name)
	}
	return factory(), nil
}

func init() {
	RegisterNotifier("slack", func() Notifier { return &SlackNotifier{} })
	RegisterNotifier("webhook", func() Notifier { return &WebhookNotifier{} })
}
