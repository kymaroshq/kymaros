package adapter

import (
	"fmt"
	"sort"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AdapterFactory creates a BackupAdapter for the given client.
type AdapterFactory func(c client.Client) BackupAdapter

// registry holds all registered adapter factories.
var registry = map[string]AdapterFactory{}

// Register adds a new adapter factory (used by external packages via init()).
func Register(name string, factory AdapterFactory) {
	registry[name] = factory
}

// NewBackupAdapter creates the appropriate adapter for the given provider.
func NewBackupAdapter(provider string, c client.Client) (BackupAdapter, error) {
	factory, ok := registry[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported backup provider %q, available: %v", provider, Available())
	}
	return factory(c), nil
}

// Available returns the sorted list of registered adapter names.
func Available() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func init() {
	Register("velero", func(c client.Client) BackupAdapter {
		return NewVeleroAdapter(c)
	})
}
