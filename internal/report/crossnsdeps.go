/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package report

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// crossNSRef represents a detected cross-namespace dependency.
type crossNSRef struct {
	// FromNamespace is the source namespace where the reference was found.
	FromNamespace string
	// ToNamespace is the target source namespace being referenced.
	ToNamespace string
	// Source describes where the reference was found (e.g. "Deployment/api env DATABASE_URL").
	Source string
}

// CrossNSDepsChecker detects cross-namespace service dependencies in sandbox
// namespaces and checks whether the referenced namespaces are part of the
// restore group.
type CrossNSDepsChecker struct {
	client client.Client
	logger *slog.Logger
}

// NewCrossNSDepsChecker creates a CrossNSDepsChecker.
func NewCrossNSDepsChecker(c client.Client, logger *slog.Logger) *CrossNSDepsChecker {
	return &CrossNSDepsChecker{client: c, logger: logger}
}

// Check scans sandbox namespaces for cross-namespace service references.
// sourceToSandbox maps source namespace name → sandbox namespace name.
// Returns the coverage ratio (0.0-1.0) and a LevelResult.
//
// A single-namespace test always returns ratio 1.0 (no cross-NS deps to verify).
// When no cross-NS references are detected, the ratio is also 1.0.
func (c *CrossNSDepsChecker) Check(
	ctx context.Context,
	sourceToSandbox map[string]string,
) (float64, *restorev1alpha1.LevelResult, error) {
	sourceNames := make(map[string]bool, len(sourceToSandbox))
	for src := range sourceToSandbox {
		sourceNames[src] = true
	}

	// Single namespace — nothing to check
	if len(sourceToSandbox) <= 1 {
		return 1.0, &restorev1alpha1.LevelResult{
			Status: "pass",
			Detail: "single namespace, no cross-namespace dependencies",
		}, nil
	}

	// Build a regex matching any of the source namespace names in DNS patterns.
	// Matches: <service>.<namespace>.svc or <namespace>.svc.cluster.local
	var refs []crossNSRef
	for sourceNS, sandboxNS := range sourceToSandbox {
		found := c.scanNamespace(ctx, sandboxNS, sourceNS)
		refs = append(refs, found...)
	}

	// No cross-NS refs found at all
	if len(refs) == 0 {
		return 1.0, &restorev1alpha1.LevelResult{
			Status: "pass",
			Detail: "no cross-namespace references detected",
		}, nil
	}

	// Check which refs are satisfied (target namespace is in the restore group)
	var satisfied, total int
	var testedList, notTestedList []string
	for _, ref := range refs {
		total++
		desc := fmt.Sprintf("%s → %s (%s)", ref.FromNamespace, ref.ToNamespace, ref.Source)
		if sourceNames[ref.ToNamespace] {
			satisfied++
			testedList = append(testedList, desc)
		} else {
			notTestedList = append(notTestedList, desc)
		}
	}

	ratio := float64(satisfied) / float64(total)
	status := "pass"
	if ratio < 1.0 {
		status = "partial"
	}
	if ratio == 0 {
		status = "fail"
	}

	c.logger.InfoContext(ctx, "cross-namespace deps check",
		"satisfied", satisfied, "total", total, "ratio", fmt.Sprintf("%.2f", ratio),
	)

	return ratio, &restorev1alpha1.LevelResult{
		Status:    status,
		Detail:    fmt.Sprintf("%d/%d cross-namespace dependencies covered", satisfied, total),
		Tested:    testedList,
		NotTested: notTestedList,
	}, nil
}

// scanNamespace scans a single sandbox namespace for references to other namespaces.
func (c *CrossNSDepsChecker) scanNamespace(
	ctx context.Context,
	sandboxNS, sourceNS string,
) []crossNSRef {
	var refs []crossNSRef

	// Scan Deployments
	var deploys appsv1.DeploymentList
	if err := c.client.List(ctx, &deploys, client.InNamespace(sandboxNS)); err == nil {
		for _, d := range deploys.Items {
			found := c.scanPodSpec(d.Spec.Template.Spec, sourceNS,
				fmt.Sprintf("Deployment/%s", d.Name))
			refs = append(refs, found...)
		}
	}

	// Scan StatefulSets
	var sts appsv1.StatefulSetList
	if err := c.client.List(ctx, &sts, client.InNamespace(sandboxNS)); err == nil {
		for _, s := range sts.Items {
			found := c.scanPodSpec(s.Spec.Template.Spec, sourceNS,
				fmt.Sprintf("StatefulSet/%s", s.Name))
			refs = append(refs, found...)
		}
	}

	// Scan ExternalName Services
	var services corev1.ServiceList
	if err := c.client.List(ctx, &services, client.InNamespace(sandboxNS)); err == nil {
		for _, svc := range services.Items {
			if svc.Spec.Type != corev1.ServiceTypeExternalName {
				continue
			}
			for _, target := range extractNamespacesFromDNS(svc.Spec.ExternalName, sourceNS) {
				refs = append(refs, crossNSRef{
					FromNamespace: sourceNS,
					ToNamespace:   target,
					Source:        fmt.Sprintf("Service/%s ExternalName", svc.Name),
				})
			}
		}
	}

	return refs
}

// scanPodSpec scans env vars in all containers for cross-namespace DNS references.
func (c *CrossNSDepsChecker) scanPodSpec(
	spec corev1.PodSpec,
	sourceNS string,
	owner string,
) []crossNSRef {
	var refs []crossNSRef
	seen := make(map[string]bool) // deduplicate per target namespace

	for _, container := range append(spec.Containers, spec.InitContainers...) {
		for _, env := range container.Env {
			for _, target := range extractNamespacesFromDNS(env.Value, sourceNS) {
				if !seen[target] {
					seen[target] = true
					refs = append(refs, crossNSRef{
						FromNamespace: sourceNS,
						ToNamespace:   target,
						Source:        fmt.Sprintf("%s env %s", owner, env.Name),
					})
				}
			}
		}
	}

	return refs
}

// dnsPattern matches Kubernetes internal DNS: <service>.<namespace>.svc
var dnsPattern = regexp.MustCompile(`[a-z0-9-]+\.([a-z0-9-]+)\.svc(?:\.cluster\.local)?`)

// extractNamespaceFromDNS extracts namespace references from Kubernetes internal DNS
// patterns in the given value. Returns all unique namespaces found that differ from
// excludeNS (the pod's own namespace).
func extractNamespacesFromDNS(value, excludeNS string) []string {
	var result []string
	seen := make(map[string]bool)
	matches := dnsPattern.FindAllStringSubmatch(strings.ToLower(value), -1)
	for _, m := range matches {
		ns := m[1]
		if ns != excludeNS && !seen[ns] {
			seen[ns] = true
			result = append(result, ns)
		}
	}
	return result
}
