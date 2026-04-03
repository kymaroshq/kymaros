package report

import (
	"context"
	"fmt"
	"io"
	"strings"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultMaxLogLines = 100
	MaxLogBytes        = 64 * 1024 // 64KB max per container
	MaxTotalLogBytes   = 512 * 1024 // 512KB max for all logs in a report
)

// CollectPodLogs collects logs from all pods in the given namespace.
func CollectPodLogs(ctx context.Context, k8sClient client.Client, clientset kubernetes.Interface, namespace string, cfg *restorev1alpha1.LogCollectionSpec) ([]restorev1alpha1.PodLog, error) {
	maxLines := int64(DefaultMaxLogLines)
	includeInit := true
	includePrevious := true

	if cfg != nil {
		if cfg.MaxLines > 0 {
			maxLines = int64(cfg.MaxLines)
		}
		includeInit = cfg.IncludeInit
		includePrevious = cfg.IncludePrevious
	}

	podList := &corev1.PodList{}
	if err := k8sClient.List(ctx, podList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("listing pods in %s: %w", namespace, err)
	}

	var podLogs []restorev1alpha1.PodLog

	for _, pod := range podList.Items {
		podLog := restorev1alpha1.PodLog{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			Phase:     string(pod.Status.Phase),
		}

		// Init containers
		if includeInit {
			for _, initContainer := range pod.Spec.InitContainers {
				log, truncated, totalLines := getContainerLog(ctx, clientset, namespace, pod.Name, initContainer.Name, maxLines, false)
				podLog.Containers = append(podLog.Containers, restorev1alpha1.ContainerLog{
					Name:       initContainer.Name,
					Type:       "init",
					Log:        log,
					Truncated:  truncated,
					TotalLines: totalLines,
				})
			}
		}

		// Regular containers
		for _, container := range pod.Spec.Containers {
			log, truncated, totalLines := getContainerLog(ctx, clientset, namespace, pod.Name, container.Name, maxLines, false)
			podLog.Containers = append(podLog.Containers, restorev1alpha1.ContainerLog{
				Name:       container.Name,
				Type:       "container",
				Log:        log,
				Truncated:  truncated,
				TotalLines: totalLines,
			})

			// Previous logs (if the container has restarted)
			if includePrevious && hasRestartedContainer(pod, container.Name) {
				prevLog, _, _ := getContainerLog(ctx, clientset, namespace, pod.Name, container.Name, maxLines, true)
				if prevLog != "" {
					podLog.Containers = append(podLog.Containers, restorev1alpha1.ContainerLog{
						Name: container.Name,
						Type: "previous",
						Log:  prevLog,
					})
				}
			}
		}

		podLogs = append(podLogs, podLog)
	}

	return podLogs, nil
}

func getContainerLog(ctx context.Context, clientset kubernetes.Interface, namespace, podName, containerName string, maxLines int64, previous bool) (string, bool, int) {
	limitBytes := int64(MaxLogBytes)
	opts := &corev1.PodLogOptions{
		Container:  containerName,
		TailLines:  &maxLines,
		Previous:   previous,
		LimitBytes: &limitBytes,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Sprintf("[error retrieving logs: %v]", err), false, 0
	}
	defer stream.Close()

	bytes, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Sprintf("[error reading logs: %v]", err), false, 0
	}

	log := string(bytes)
	lines := strings.Count(log, "\n")
	truncated := lines >= int(maxLines) || len(bytes) >= int(limitBytes)

	return log, truncated, lines
}

func hasRestartedContainer(pod corev1.Pod, containerName string) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName && cs.RestartCount > 0 {
			return true
		}
	}
	return false
}

// CollectEvents collects Kubernetes events from the given namespace.
func CollectEvents(ctx context.Context, k8sClient client.Client, namespace string) ([]restorev1alpha1.EventLog, error) {
	eventList := &corev1.EventList{}
	if err := k8sClient.List(ctx, eventList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("listing events in %s: %w", namespace, err)
	}

	var events []restorev1alpha1.EventLog
	for _, event := range eventList.Items {
		ts := ""
		if !event.LastTimestamp.IsZero() {
			ts = event.LastTimestamp.Format("2006-01-02T15:04:05Z")
		} else if !event.EventTime.IsZero() {
			ts = event.EventTime.Format("2006-01-02T15:04:05Z")
		}
		events = append(events, restorev1alpha1.EventLog{
			Type:           event.Type,
			Reason:         event.Reason,
			Message:        event.Message,
			InvolvedObject: fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			LastTimestamp:   ts,
			Count:          int(event.Count),
		})
	}

	return events, nil
}

// TruncateLogs truncates pod logs to fit within maxBytes total.
func TruncateLogs(podLogs []restorev1alpha1.PodLog, maxBytes int) []restorev1alpha1.PodLog {
	totalSize := 0
	for _, pl := range podLogs {
		for _, cl := range pl.Containers {
			totalSize += len(cl.Log)
		}
	}
	if totalSize <= maxBytes {
		return podLogs
	}

	// Truncate longest logs first until we fit
	for totalSize > maxBytes {
		var longestPod, longestContainer int
		longestLen := 0
		for i, pl := range podLogs {
			for j, cl := range pl.Containers {
				if len(cl.Log) > longestLen {
					longestLen = len(cl.Log)
					longestPod = i
					longestContainer = j
				}
			}
		}
		if longestLen == 0 {
			break
		}
		// Cut in half
		newLen := longestLen / 2
		if newLen < 256 {
			newLen = 0
		}
		cl := &podLogs[longestPod].Containers[longestContainer]
		totalSize -= len(cl.Log)
		if newLen > 0 {
			cl.Log = cl.Log[len(cl.Log)-newLen:]
		} else {
			cl.Log = "[truncated — log too large]"
		}
		cl.Truncated = true
		totalSize += len(cl.Log)
	}
	return podLogs
}
