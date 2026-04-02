package sandbox

import (
	"context"
	"fmt"
)

// RewriteDNS rewrites service references in restored resources to point to sandbox namespaces.
// This is a Pro-tier feature and is not implemented in the MVP.
func (m *Manager) RewriteDNS(ctx context.Context, namespace string, mappings map[string]string) error {
	return fmt.Errorf("rewrite DNS: not implemented (Pro tier)")
}
