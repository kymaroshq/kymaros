package healthcheck

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// CheckResourceExists verifies that specified K8s resources exist in the sandbox
func (r *Runner) CheckResourceExists(ctx context.Context, check restorev1alpha1.ResourceExistsCheck, namespace string) Result {
	var missing []string

	for _, ref := range check.Resources {
		exists, err := r.resourceExists(ctx, ref, namespace)
		if err != nil {
			return Result{Status: "fail", Message: fmt.Sprintf("check %s/%s: %v", ref.Kind, ref.Name, err)}
		}
		if !exists {
			missing = append(missing, fmt.Sprintf("%s/%s", ref.Kind, ref.Name))
		}
	}

	if len(missing) > 0 {
		return Result{
			Status:  "fail",
			Message: fmt.Sprintf("missing resources: %s", strings.Join(missing, ", ")),
		}
	}

	return Result{
		Status:  "pass",
		Message: fmt.Sprintf("all %d resources exist", len(check.Resources)),
	}
}

func (r *Runner) resourceExists(ctx context.Context, ref restorev1alpha1.ResourceRef, namespace string) (bool, error) {
	key := types.NamespacedName{Name: ref.Name, Namespace: namespace}

	var obj client.Object
	switch ref.Kind {
	case "Secret":
		obj = &corev1.Secret{}
	case "ConfigMap":
		obj = &corev1.ConfigMap{}
	case "Service":
		obj = &corev1.Service{}
	case "PersistentVolumeClaim":
		obj = &corev1.PersistentVolumeClaim{}
	default:
		return false, fmt.Errorf("unsupported resource kind: %s", ref.Kind)
	}

	if err := r.client.Get(ctx, key, obj); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
