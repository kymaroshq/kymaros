package report

import (
	"context"
	"fmt"
	"log/slog"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Generator creates RestoreReport CRs from test results
type Generator struct {
	client client.Client
	logger *slog.Logger
}

// NewGenerator creates a report Generator
func NewGenerator(c client.Client, logger *slog.Logger) *Generator {
	return &Generator{client: c, logger: logger}
}

// Generate creates a RestoreReport CR with the test results
func (g *Generator) Generate(ctx context.Context, test *restorev1alpha1.RestoreTest, score int, result string) (*restorev1alpha1.RestoreReport, error) {
	return nil, fmt.Errorf("generate report: %w", errNotImplemented)
}

var errNotImplemented = fmt.Errorf("not implemented")
