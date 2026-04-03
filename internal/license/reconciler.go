package license

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// LicenseReconciler watches the kymaros-license Secret and refreshes the
// LicenseManager whenever it changes. It also requeues hourly to detect
// license expiration even when the Secret is untouched.
type LicenseReconciler struct {
	client.Client
	Manager *LicenseManager
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *LicenseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Name != secretName || req.Namespace != secretNamespace {
		return ctrl.Result{}, nil
	}

	r.Manager.Refresh(ctx)

	// Re-check hourly to catch expiration transitions.
	return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
}

func (r *LicenseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetName() == secretName && obj.GetNamespace() == secretNamespace
		})).
		Named("license").
		Complete(r)
}
