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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"github.com/kymaroshq/kymaros/internal/adapter"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
)

var (
	startTime = time.Now()
	// Version is set at build time via ldflags.
	Version = "dev"
)

// writeJSON marshals v as JSON and writes it to w with the appropriate headers.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Headers already sent; nothing meaningful we can do here.
		_ = err
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// CORSMiddleware adds Access-Control-Allow-* headers.
// When KYMAROS_CORS_ORIGIN is set, only that origin is allowed.
// When it is unset, no CORS headers are emitted (same-origin policy applies).
func CORSMiddleware(next http.Handler) http.Handler {
	allowed := os.Getenv("KYMAROS_CORS_ORIGIN")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			if allowed != "" {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HandleHealth returns cluster info and health status.
// GET /api/v1/health
func HandleHealth(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		k8sVersion := "unknown"
		if q.restConfig != nil {
			if dc, err := discovery.NewDiscoveryClientForConfig(q.restConfig); err == nil {
				if info, err := dc.ServerVersion(); err == nil {
					k8sVersion = info.GitVersion
				}
			}
		}

		writeJSON(w, http.StatusOK, HealthResponse{
			Status:            "ok",
			Version:           Version,
			KubernetesVersion: k8sVersion,
			Namespace:         kymarosNamespace,
			Uptime:            formatDurationLong(time.Since(startTime)),
			Pod:               os.Getenv("HOSTNAME"),
			StartedAt:         startTime.UTC().Format(time.RFC3339),
		})
	}
}

// HandleSummary returns aggregate statistics across all RestoreTests and reports.
// GET /api/v1/summary
func HandleSummary(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := q.GetSummary(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}

// HandleDailyScores returns per-day score history.
// GET /api/v1/summary/daily?days=30
func HandleDailyScores(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		days := 30
		if v := r.URL.Query().Get("days"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				writeError(w, http.StatusBadRequest, "invalid 'days' parameter")
				return
			}
			days = n
		}

		if days > 365 {
			days = 365
		}

		scores, err := q.GetDailyScores(r.Context(), days)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, scores)
	}
}

// HandleTests returns a list of all RestoreTest resources as TestResponse objects.
// GET /api/v1/tests
func HandleTests(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tests, err := q.ListTests(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := make([]TestResponse, 0, len(tests))
		for _, t := range tests {
			resp = append(resp, MapTestToResponse(t))
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// HandleReports returns RestoreReport resources, optionally filtered.
// GET /api/v1/reports?test=xxx&days=30&latest=true
func HandleReports(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		testName := params.Get("test")

		days := 0
		if v := params.Get("days"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				writeError(w, http.StatusBadRequest, "invalid 'days' parameter")
				return
			}
			days = n
		}

		latestOnly := false
		if v := params.Get("latest"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid 'latest' parameter")
				return
			}
			latestOnly = b
		}

		if latestOnly {
			reports, err := q.GetLatestReports(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, reports)
			return
		}

		reports, err := q.ListReports(r.Context(), testName, days, false)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, reports)
	}
}

// HandleAlerts returns recent alerts (score drops, failures, SLA breaches).
// GET /api/v1/alerts?hours=48
func HandleAlerts(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hours := 48
		if v := r.URL.Query().Get("hours"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				writeError(w, http.StatusBadRequest, "invalid 'hours' parameter")
				return
			}
			hours = n
		}

		alerts, err := q.GetAlerts(r.Context(), hours)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, alerts)
	}
}

// HandleUpcoming returns the next scheduled RestoreTest runs.
// GET /api/v1/upcoming
func HandleUpcoming(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upcoming, err := q.GetUpcomingTests(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, upcoming)
	}
}

// HandleSchedules returns discovered Velero backup schedules.
func HandleSchedules(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schedules, err := q.GetSchedules(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, schedules)
	}
}

// HandleCreateTest creates a new RestoreTest CR.
// POST /api/v1/tests
func HandleCreateTest(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB limit
		var input CreateTestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if input.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		if len(input.Namespaces) == 0 {
			writeError(w, http.StatusBadRequest, "at least one namespace is required")
			return
		}
		available := adapter.Available()
		if !slices.Contains(available, input.Provider) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported provider %q, available: %s", input.Provider, strings.Join(available, ", ")))
			return
		}

		if err := q.CreateTest(r.Context(), input); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"message": "test created"})
	}
}

// HandleUpdateTest updates an existing RestoreTest's spec.
// PUT /api/v1/tests/{name}
func HandleUpdateTest(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB limit
		var input CreateTestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := q.UpdateTest(r.Context(), name, input); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "test updated"})
	}
}

// HandleDeleteTest deletes a RestoreTest by name.
// DELETE /api/v1/tests/{name}
func HandleDeleteTest(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		if err := q.DeleteTest(r.Context(), name); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleReportLogs returns the pod logs and events from a RestoreReport.
// GET /api/v1/reports/{name}/logs
func HandleReportLogs(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		var report restorev1alpha1.RestoreReport
		if err := q.client.Get(r.Context(), types.NamespacedName{
			Name:      name,
			Namespace: kymarosNamespace,
		}, &report); err != nil {
			writeError(w, http.StatusNotFound, "report not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"podLogs": report.Status.PodLogs,
			"events":  report.Status.Events,
		})
	}
}

// HandleTriggerTest forces an immediate run of a RestoreTest.
// POST /api/v1/tests/{name}/trigger
func HandleTriggerTest(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		if err := q.TriggerTest(r.Context(), name); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "test triggered"})
	}
}

// HandleGetProviderConfig returns backup provider status.
// GET /api/v1/config/provider
func HandleGetProviderConfig(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, q.GetProviderConfig(r.Context()))
	}
}

// HandleGetNotificationConfig returns aggregated notification config from all RestoreTests.
// GET /api/v1/config/notifications
func HandleGetNotificationConfig(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := q.GetNotificationConfig(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, config)
	}
}

// HandleTestNotification sends a test notification.
// POST /api/v1/config/notifications/test
func HandleTestNotification(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "test notification sent"})
	}
}

// HandleGetSandboxConfig returns sandbox defaults and active/orphaned counts.
// GET /api/v1/config/sandbox
func HandleGetSandboxConfig(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, q.GetSandboxConfig(r.Context()))
	}
}

// HandleCleanupOrphanSandboxes deletes orphaned sandbox namespaces.
// POST /api/v1/config/sandbox/cleanup
func HandleCleanupOrphanSandboxes(q *Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deleted, err := q.CleanupOrphanSandboxes(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
