package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// CheckHTTPGet performs an HTTP GET against a service in the sandbox
func (r *Runner) CheckHTTPGet(ctx context.Context, check restorev1alpha1.HTTPGetCheck, namespace string) Result {
	// Resolve service ClusterIP
	var svc corev1.Service
	if err := r.client.Get(ctx, types.NamespacedName{Name: check.Service, Namespace: namespace}, &svc); err != nil {
		return Result{Status: "fail", Message: fmt.Sprintf("get service %q: %v", check.Service, err)}
	}

	clusterIP := svc.Spec.ClusterIP
	if clusterIP == "" || clusterIP == "None" {
		return Result{Status: "fail", Message: fmt.Sprintf("service %q has no ClusterIP", check.Service)}
	}

	url := fmt.Sprintf("http://%s:%d%s", clusterIP, check.Port, check.Path)

	timeout := 10 * time.Second
	if check.Timeout.Duration > 0 {
		timeout = check.Timeout.Duration
	}
	retries := check.Retries
	if retries <= 0 {
		retries = 1
	}

	httpClient := &http.Client{Timeout: timeout}

	var lastErr error
	for i := 0; i < retries; i++ {
		resp, err := httpClient.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second)
			continue
		}
		_ = resp.Body.Close()

		if check.ExpectedStatus > 0 && resp.StatusCode != check.ExpectedStatus {
			lastErr = fmt.Errorf("expected status %d, got %d", check.ExpectedStatus, resp.StatusCode)
			time.Sleep(time.Second)
			continue
		}

		return Result{
			Status:  "pass",
			Message: fmt.Sprintf("GET %s → %d", url, resp.StatusCode),
		}
	}

	return Result{
		Status:  "fail",
		Message: fmt.Sprintf("GET %s failed after %d retries: %v", url, retries, lastErr),
	}
}
