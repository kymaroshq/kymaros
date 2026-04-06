package healthcheck

import (
	"context"
	"fmt"
	"net"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// CheckTCPSocket verifies TCP connectivity to a service
func (r *Runner) CheckTCPSocket(ctx context.Context, check restorev1alpha1.TCPSocketCheck, namespace string) Result {
	// Resolve service ClusterIP
	var svc corev1.Service
	if err := r.client.Get(ctx, types.NamespacedName{Name: check.Service, Namespace: namespace}, &svc); err != nil {
		return Result{Status: "fail", Message: fmt.Sprintf("get service %q: %v", check.Service, err)}
	}

	clusterIP := svc.Spec.ClusterIP
	if clusterIP == "" || clusterIP == "None" {
		return Result{Status: "fail", Message: fmt.Sprintf("service %q has no ClusterIP", check.Service)}
	}

	addr := net.JoinHostPort(clusterIP, fmt.Sprintf("%d", check.Port))

	timeout := 10 * time.Second
	if check.Timeout.Duration > 0 {
		timeout = check.Timeout.Duration
	}

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return Result{
			Status:  "fail",
			Message: fmt.Sprintf("TCP connect to %s: %v", addr, err),
		}
	}
	_ = conn.Close()

	return Result{
		Status:  "pass",
		Message: fmt.Sprintf("TCP connect to %s succeeded", addr),
	}
}
