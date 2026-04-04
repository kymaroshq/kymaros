package healthcheck

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// CheckPodStatus verifies pods matching the selector are ready
func (r *Runner) CheckPodStatus(ctx context.Context, check restorev1alpha1.PodStatusCheck, namespace string) Result {
	var podList corev1.PodList
	matchLabels := client.MatchingLabels(check.LabelSelector)

	if err := r.client.List(ctx, &podList, client.InNamespace(namespace), matchLabels); err != nil {
		return Result{Status: "fail", Message: fmt.Sprintf("list pods: %v", err)}
	}

	ready := 0
	for _, pod := range podList.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				ready++
				break
			}
		}
	}

	total := len(podList.Items)
	if ready >= check.MinReady {
		return Result{
			Status:  "pass",
			Message: fmt.Sprintf("%d/%d pods ready (min: %d)", ready, total, check.MinReady),
		}
	}

	return Result{
		Status:  "fail",
		Message: fmt.Sprintf("%d/%d pods ready (min: %d required)", ready, total, check.MinReady),
	}
}
