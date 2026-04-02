package report

import (
	"context"
	"fmt"
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// CompletenessChecker compares resources in source vs sandbox namespace
type CompletenessChecker struct {
	client client.Client
	logger *slog.Logger
}

// NewCompletenessChecker creates a CompletenessChecker
func NewCompletenessChecker(c client.Client, logger *slog.Logger) *CompletenessChecker {
	return &CompletenessChecker{client: c, logger: logger}
}

// Check compares resource counts between source and sandbox namespaces.
// Returns the CompletenessStatus, overall ratio (0.0-1.0), and any error.
func (c *CompletenessChecker) Check(ctx context.Context, sourceNS, sandboxNS string) (*restorev1alpha1.CompletenessStatus, float64, error) {
	type resourceCount struct {
		name    string
		source  int
		sandbox int
	}

	counts := []resourceCount{
		{name: "deployments", source: c.countDeployments(ctx, sourceNS), sandbox: c.countDeployments(ctx, sandboxNS)},
		{name: "services", source: c.countServices(ctx, sourceNS), sandbox: c.countServices(ctx, sandboxNS)},
		{name: "secrets", source: c.countSecrets(ctx, sourceNS), sandbox: c.countSecrets(ctx, sandboxNS)},
		{name: "configMaps", source: c.countConfigMaps(ctx, sourceNS), sandbox: c.countConfigMaps(ctx, sandboxNS)},
		{name: "pvcs", source: c.countPVCs(ctx, sourceNS), sandbox: c.countPVCs(ctx, sandboxNS)},
	}

	status := &restorev1alpha1.CompletenessStatus{}
	var totalSource, totalSandbox int

	for _, rc := range counts {
		totalSource += rc.source
		totalSandbox += rc.sandbox
		ratio := fmt.Sprintf("%d/%d", rc.sandbox, rc.source)

		switch rc.name {
		case "deployments":
			status.Deployments = ratio
		case "services":
			status.Services = ratio
		case "secrets":
			status.Secrets = ratio
		case "configMaps":
			status.ConfigMaps = ratio
		case "pvcs":
			status.PVCs = ratio
		}
	}

	var overallRatio float64
	if totalSource > 0 {
		overallRatio = float64(totalSandbox) / float64(totalSource)
		if overallRatio > 1.0 {
			overallRatio = 1.0
		}
	}

	c.logger.InfoContext(ctx, "completeness check",
		"source", sourceNS, "sandbox", sandboxNS,
		"ratio", fmt.Sprintf("%.2f", overallRatio),
		"sourceTotal", totalSource, "sandboxTotal", totalSandbox,
	)

	return status, overallRatio, nil
}

func (c *CompletenessChecker) countDeployments(ctx context.Context, ns string) int {
	var list appsv1.DeploymentList
	if err := c.client.List(ctx, &list, client.InNamespace(ns)); err != nil {
		return 0
	}
	return len(list.Items)
}

func (c *CompletenessChecker) countServices(ctx context.Context, ns string) int {
	var list corev1.ServiceList
	if err := c.client.List(ctx, &list, client.InNamespace(ns)); err != nil {
		return 0
	}
	return len(list.Items)
}

func (c *CompletenessChecker) countSecrets(ctx context.Context, ns string) int {
	var list corev1.SecretList
	if err := c.client.List(ctx, &list, client.InNamespace(ns)); err != nil {
		return 0
	}
	return len(list.Items)
}

func (c *CompletenessChecker) countConfigMaps(ctx context.Context, ns string) int {
	var list corev1.ConfigMapList
	if err := c.client.List(ctx, &list, client.InNamespace(ns)); err != nil {
		return 0
	}
	return len(list.Items)
}

func (c *CompletenessChecker) countPVCs(ctx context.Context, ns string) int {
	var list corev1.PersistentVolumeClaimList
	if err := c.client.List(ctx, &list, client.InNamespace(ns)); err != nil {
		return 0
	}
	return len(list.Items)
}
