package sandbox

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// ApplyResourceQuota creates a ResourceQuota in the sandbox namespace
func (m *Manager) ApplyResourceQuota(ctx context.Context, namespace string, config *restorev1alpha1.ResourceQuotaConfig) error {
	hard := corev1.ResourceList{}

	if config.CPU != "" {
		hard[corev1.ResourceRequestsCPU] = resource.MustParse(config.CPU)
		hard[corev1.ResourceLimitsCPU] = resource.MustParse(config.CPU)
	}
	if config.Memory != "" {
		hard[corev1.ResourceRequestsMemory] = resource.MustParse(config.Memory)
		hard[corev1.ResourceLimitsMemory] = resource.MustParse(config.Memory)
	}
	if config.Storage != "" {
		hard[corev1.ResourceRequestsStorage] = resource.MustParse(config.Storage)
	}

	if len(hard) == 0 {
		return nil
	}

	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sandbox-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: hard,
		},
	}

	if err := m.client.Create(ctx, quota); err != nil {
		return fmt.Errorf("create ResourceQuota in %q: %w", namespace, err)
	}
	m.logger.InfoContext(ctx, "ResourceQuota applied", "namespace", namespace)
	return nil
}

// ApplyLimitRange creates a LimitRange with sane defaults in the sandbox namespace
func (m *Manager) ApplyLimitRange(ctx context.Context, namespace string) error {
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sandbox-limits",
			Namespace: namespace,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type: corev1.LimitTypeContainer,
					Default: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("512Mi"),
					},
					DefaultRequest: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
				},
			},
		},
	}

	if err := m.client.Create(ctx, lr); err != nil {
		return fmt.Errorf("create LimitRange in %q: %w", namespace, err)
	}
	m.logger.InfoContext(ctx, "LimitRange applied", "namespace", namespace)
	return nil
}
