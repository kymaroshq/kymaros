package healthcheck

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

// CheckExec runs a command in a pod matching the selector and checks the exit code.
func (r *Runner) CheckExec(ctx context.Context, check restorev1alpha1.ExecCheck, namespace string) Result {
	if r.restConfig == nil {
		return Result{Status: "skip", Message: "exec checks require rest.Config (not configured)"}
	}

	// Find a pod matching the selector
	var podList corev1.PodList
	matchLabels := client.MatchingLabels(check.PodSelector)
	if err := r.client.List(ctx, &podList, client.InNamespace(namespace), matchLabels); err != nil {
		return Result{Status: "fail", Message: fmt.Sprintf("list pods: %v", err)}
	}

	if len(podList.Items) == 0 {
		return Result{Status: "fail", Message: "no pods match selector"}
	}

	// Use the first running pod
	var targetPod *corev1.Pod
	for i := range podList.Items {
		if podList.Items[i].Status.Phase == corev1.PodRunning {
			targetPod = &podList.Items[i]
			break
		}
	}
	if targetPod == nil {
		return Result{Status: "fail", Message: "no running pods match selector"}
	}

	container := check.Container
	if container == "" && len(targetPod.Spec.Containers) > 0 {
		container = targetPod.Spec.Containers[0].Name
	}

	// Execute command
	clientset, err := kubernetes.NewForConfig(r.restConfig)
	if err != nil {
		return Result{Status: "fail", Message: fmt.Sprintf("create clientset: %v", err)}
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(targetPod.Name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   check.Command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(r.restConfig, "POST", req.URL())
	if err != nil {
		return Result{Status: "fail", Message: fmt.Sprintf("create executor: %v", err)}
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		// Non-zero exit code comes as an error
		if check.SuccessExitCode != 0 {
			return Result{Status: "fail", Message: fmt.Sprintf("exec failed: %v (stderr: %s)", err, stderr.String())}
		}
		return Result{Status: "fail", Message: fmt.Sprintf("command failed: %v", err)}
	}

	// Exit code 0 = success by default
	if check.SuccessExitCode == 0 {
		return Result{
			Status:  "pass",
			Message: fmt.Sprintf("command succeeded (stdout: %s)", truncate(stdout.String(), 200)),
		}
	}

	return Result{Status: "fail", Message: fmt.Sprintf("expected exit code %d, got 0", check.SuccessExitCode)}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
