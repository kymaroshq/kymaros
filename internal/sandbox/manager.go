package sandbox

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

const (
	LabelManagedBy     = "kymaros.io/managed-by"
	LabelTestName      = "kymaros.io/test"
	LabelTestNamespace = "kymaros.io/test-namespace"
	LabelGroup         = "kymaros.io/group"
	ManagedByValue     = "kymaros"
)

// Manager handles sandbox namespace lifecycle
type Manager struct {
	client client.Client
	logger *slog.Logger
}

// NewManager creates a sandbox Manager
func NewManager(c client.Client, logger *slog.Logger) *Manager {
	return &Manager{client: c, logger: logger}
}

// Create creates an isolated sandbox namespace with NetworkPolicy, ResourceQuota, and LimitRange.
// Returns the created namespace name.
func (m *Manager) Create(ctx context.Context, config restorev1alpha1.SandboxConfig, testName, testNamespace string) (string, error) {
	prefix := config.NamespacePrefix
	if prefix == "" {
		prefix = "rp-test"
	}

	suffix, err := randomSuffix(6)
	if err != nil {
		return "", fmt.Errorf("generate namespace suffix: %w", err)
	}
	nsName := fmt.Sprintf("%s-%s-%s", prefix, testName, suffix)

	// Truncate to 63 chars (K8s namespace name limit)
	if len(nsName) > 63 {
		nsName = nsName[:63]
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
			Labels: map[string]string{
				LabelManagedBy:     ManagedByValue,
				LabelTestName:      testName,
				LabelTestNamespace: testNamespace,
				LabelGroup:         testName,
			},
		},
	}

	if err := m.client.Create(ctx, ns); err != nil {
		return "", fmt.Errorf("create namespace %q: %w", nsName, err)
	}
	m.logger.InfoContext(ctx, "sandbox namespace created", "namespace", nsName, "test", testName)

	// Apply network isolation
	isolation := config.NetworkIsolation
	if isolation == "" {
		isolation = "strict"
	}
	switch isolation {
	case "strict":
		if err := m.ApplyDenyAll(ctx, nsName); err != nil {
			return nsName, fmt.Errorf("apply deny-all: %w", err)
		}
	case "group":
		if err := m.ApplyGroupAllow(ctx, nsName, testName); err != nil {
			return nsName, fmt.Errorf("apply group-allow: %w", err)
		}
	}

	// Apply ResourceQuota if configured
	if config.ResourceQuota != nil {
		if err := m.ApplyResourceQuota(ctx, nsName, config.ResourceQuota); err != nil {
			return nsName, fmt.Errorf("apply resource quota: %w", err)
		}
	}

	// Always apply LimitRange with sane defaults
	if err := m.ApplyLimitRange(ctx, nsName); err != nil {
		return nsName, fmt.Errorf("apply limit range: %w", err)
	}

	return nsName, nil
}

// Cleanup deletes the sandbox namespace and all its resources
func (m *Manager) Cleanup(ctx context.Context, namespace string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if err := m.client.Delete(ctx, ns); err != nil {
		return fmt.Errorf("delete namespace %q: %w", namespace, err)
	}
	m.logger.InfoContext(ctx, "sandbox namespace deleted", "namespace", namespace)
	return nil
}

// IsExpired checks if a sandbox namespace has exceeded its TTL
func (m *Manager) IsExpired(ns *corev1.Namespace, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	return time.Since(ns.CreationTimestamp.Time) > ttl
}

func randomSuffix(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}
