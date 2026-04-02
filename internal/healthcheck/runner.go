package healthcheck

import (
	"context"
	"log/slog"
	"time"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Result captures the outcome of a single health check
type Result struct {
	Name     string
	Status   string // pass | fail | skip
	Duration string
	Message  string
}

// Runner executes health checks against a sandbox namespace
type Runner struct {
	client     client.Client
	logger     *slog.Logger
	restConfig *rest.Config // optional, needed for exec checks
}

// NewRunner creates a health check Runner
func NewRunner(c client.Client, logger *slog.Logger) *Runner {
	return &Runner{client: c, logger: logger}
}

// NewRunnerWithExec creates a Runner with exec support
func NewRunnerWithExec(c client.Client, logger *slog.Logger, cfg *rest.Config) *Runner {
	return &Runner{client: c, logger: logger, restConfig: cfg}
}

// RunAll executes all checks from a HealthCheckPolicy against the sandbox namespace
func (r *Runner) RunAll(ctx context.Context, checks []restorev1alpha1.HealthCheck, sandboxNS string) []Result {
	results := make([]Result, 0, len(checks))

	for _, check := range checks {
		start := time.Now()
		var res Result

		switch check.Type {
		case "podStatus":
			if check.PodStatus != nil {
				res = r.CheckPodStatus(ctx, *check.PodStatus, sandboxNS)
			} else {
				res = Result{Status: "skip", Message: "podStatus config missing"}
			}
		case "httpGet":
			if check.HTTPGet != nil {
				res = r.CheckHTTPGet(ctx, *check.HTTPGet, sandboxNS)
			} else {
				res = Result{Status: "skip", Message: "httpGet config missing"}
			}
		case "tcpSocket":
			if check.TCPSocket != nil {
				res = r.CheckTCPSocket(ctx, *check.TCPSocket, sandboxNS)
			} else {
				res = Result{Status: "skip", Message: "tcpSocket config missing"}
			}
		case "resourceExists":
			if check.ResourceExists != nil {
				res = r.CheckResourceExists(ctx, *check.ResourceExists, sandboxNS)
			} else {
				res = Result{Status: "skip", Message: "resourceExists config missing"}
			}
		case "exec":
			res = Result{Status: "skip", Message: "exec checks not implemented (Pro tier)"}
		default:
			res = Result{Status: "skip", Message: "unknown check type: " + check.Type}
		}

		res.Name = check.Name
		res.Duration = time.Since(start).Round(time.Millisecond).String()
		results = append(results, res)

		r.logger.InfoContext(ctx, "health check executed",
			"name", check.Name, "type", check.Type,
			"status", res.Status, "duration", res.Duration,
		)
	}

	return results
}

// PassRatio returns the ratio of passed checks (0.0 to 1.0)
func PassRatio(results []Result) float64 {
	if len(results) == 0 {
		return 0
	}
	passed := 0
	for _, r := range results {
		if r.Status == "pass" {
			passed++
		}
	}
	return float64(passed) / float64(len(results))
}
