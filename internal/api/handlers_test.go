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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"github.com/kymaroshq/kymaros/internal/license"
)

// --- Test helpers ---

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = restorev1alpha1.AddToScheme(s)
	return s
}

func newTestQueries(objects ...runtime.Object) *Queries {
	scheme := newTestScheme()
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objects...).
		WithStatusSubresource(&restorev1alpha1.RestoreTest{}, &restorev1alpha1.RestoreReport{}).
		Build()
	// LicenseManager defaults to Community tier without calling Refresh.
	return NewQueries(c, license.NewManager(c))
}

// makeRestoreTest creates a RestoreTest in kymaros-system namespace.
func makeRestoreTest(name string) *restorev1alpha1.RestoreTest {
	return &restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kymarosNamespace,
		},
		Spec: restorev1alpha1.RestoreTestSpec{
			BackupSource: restorev1alpha1.BackupSource{
				Provider:   "velero",
				BackupName: "test-backup",
				Namespaces: []restorev1alpha1.NamespaceMapping{{Name: "production"}},
			},
			Schedule: restorev1alpha1.ScheduleConfig{Cron: "0 3 * * *"},
			Sandbox: restorev1alpha1.SandboxConfig{
				NamespacePrefix:  "kymaros-test",
				NetworkIsolation: "strict",
			},
		},
	}
}

// makeRestoreReport creates a RestoreReport in kymaros-system namespace.
func makeRestoreReport(name, testRef string, score int, result string) *restorev1alpha1.RestoreReport {
	return &restorev1alpha1.RestoreReport{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         kymarosNamespace,
			CreationTimestamp: metav1.NewTime(time.Now().UTC()),
		},
		Spec: restorev1alpha1.RestoreReportSpec{
			TestRef: testRef,
		},
		Status: restorev1alpha1.RestoreReportStatus{
			Score:  score,
			Result: result,
		},
	}
}

// doRequest executes a handler against a recorder and returns the recorder.
func doRequest(handler http.HandlerFunc, method, target string, body []byte) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// --- HandleHealth ---

func TestHandleHealth(t *testing.T) {
	w := doRequest(HandleHealth(), http.MethodGet, "/api/v1/health", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}

// --- HandleTests ---

func TestHandleTestsEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleTests(q), http.MethodGet, "/api/v1/tests", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []TestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body)
}

func TestHandleTestsWithData(t *testing.T) {
	test1 := makeRestoreTest("test-alpha")
	test2 := makeRestoreTest("test-beta")
	q := newTestQueries(test1, test2)

	w := doRequest(HandleTests(q), http.MethodGet, "/api/v1/tests", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []TestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 2)

	// Verify fields are mapped correctly.
	names := []string{body[0].Name, body[1].Name}
	assert.ElementsMatch(t, []string{"test-alpha", "test-beta"}, names)
}

func TestHandleTestsFieldMapping(t *testing.T) {
	test := makeRestoreTest("my-test")
	q := newTestQueries(test)

	w := doRequest(HandleTests(q), http.MethodGet, "/api/v1/tests", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var body []TestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)

	resp := body[0]
	assert.Equal(t, "my-test", resp.Name)
	assert.Equal(t, kymarosNamespace, resp.Namespace)
	assert.Equal(t, "velero", resp.Provider)
	assert.Equal(t, "0 3 * * *", resp.Schedule)
	assert.Equal(t, []string{"production"}, resp.SourceNamespaces)
}

// --- HandleReports ---

func TestHandleReportsEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleReports(q), http.MethodGet, "/api/v1/reports", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []restorev1alpha1.RestoreReport
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body)
}

func TestHandleReportsWithData(t *testing.T) {
	report := makeRestoreReport("report-1", "test-alpha", 85, restorev1alpha1.ResultPass)
	q := newTestQueries(report)

	w := doRequest(HandleReports(q), http.MethodGet, "/api/v1/reports", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []restorev1alpha1.RestoreReport
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 1)
	assert.Equal(t, "report-1", body[0].Name)
}

func TestHandleReportsBadDaysParam(t *testing.T) {
	q := newTestQueries()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?days=abc", nil)
	w := httptest.NewRecorder()
	HandleReports(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleReportsBadLatestParam(t *testing.T) {
	q := newTestQueries()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?latest=notbool", nil)
	w := httptest.NewRecorder()
	HandleReports(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleReportsLatestOnly(t *testing.T) {
	// Two reports for the same test — only the latest should be returned.
	older := makeRestoreReport("report-old", "test-alpha", 70, restorev1alpha1.ResultPass)
	older.CreationTimestamp = metav1.NewTime(time.Now().UTC().Add(-2 * time.Hour))
	newer := makeRestoreReport("report-new", "test-alpha", 90, restorev1alpha1.ResultPass)
	q := newTestQueries(older, newer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports?latest=true", nil)
	w := httptest.NewRecorder()
	HandleReports(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []restorev1alpha1.RestoreReport
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 1, "only one report per test expected when latest=true")
}

// --- HandleSummary ---

func TestHandleSummaryEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleSummary(q), http.MethodGet, "/api/v1/summary", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body SummaryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.TotalTests)
	assert.Equal(t, 0.0, body.AverageScore)
}

func TestHandleSummaryWithTests(t *testing.T) {
	test1 := makeRestoreTest("test-alpha")
	test2 := makeRestoreTest("test-beta")
	report1 := makeRestoreReport("report-1", "test-alpha", 80, restorev1alpha1.ResultPass)
	report2 := makeRestoreReport("report-2", "test-beta", 60, restorev1alpha1.ResultFail)
	q := newTestQueries(test1, test2, report1, report2)

	w := doRequest(HandleSummary(q), http.MethodGet, "/api/v1/summary", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body SummaryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 2, body.TotalTests)
	assert.Equal(t, 70.0, body.AverageScore) // (80 + 60) / 2
}

func TestHandleSummaryNamespacesCovered(t *testing.T) {
	test := makeRestoreTest("test-one")
	// The test spec already includes namespace "production".
	q := newTestQueries(test)

	w := doRequest(HandleSummary(q), http.MethodGet, "/api/v1/summary", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var body SummaryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 1, body.NamespacesCovered)
}

// --- HandleDailyScores ---

func TestHandleDailyScoresEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleDailyScores(q), http.MethodGet, "/api/v1/summary/daily", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []DailySummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body)
}

func TestHandleDailyScoresBadParam(t *testing.T) {
	q := newTestQueries()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary/daily?days=bad", nil)
	w := httptest.NewRecorder()
	HandleDailyScores(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- HandleAlerts ---

func TestHandleAlertsEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleAlerts(q), http.MethodGet, "/api/v1/alerts", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []Alert
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body)
}

func TestHandleAlertsBadParam(t *testing.T) {
	q := newTestQueries()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?hours=xyz", nil)
	w := httptest.NewRecorder()
	HandleAlerts(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleAlertsDetectsFailure(t *testing.T) {
	report := makeRestoreReport("rpt-fail", "test-alpha", 0, restorev1alpha1.ResultFail)
	// Recent enough to be picked up by default 48h window.
	report.CreationTimestamp = metav1.NewTime(time.Now().UTC().Add(-1 * time.Hour))
	q := newTestQueries(report)

	w := doRequest(HandleAlerts(q), http.MethodGet, "/api/v1/alerts", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var body []Alert
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, "test-alpha", body[0].TestName)
	assert.Equal(t, restorev1alpha1.ResultFail, body[0].Result)
}

// --- HandleCompliance ---

func TestHandleComplianceEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleCompliance(q), http.MethodGet, "/api/v1/compliance", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body ComplianceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "soc2", body.Framework) // default framework
	assert.Equal(t, "no-data", body.Status) // no reports → no-data
}

func TestHandleComplianceBadParam(t *testing.T) {
	q := newTestQueries()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance?period=bad", nil)
	w := httptest.NewRecorder()
	HandleCompliance(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleComplianceWithFramework(t *testing.T) {
	q := newTestQueries()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance?framework=iso27001", nil)
	w := httptest.NewRecorder()
	HandleCompliance(q).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body ComplianceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "iso27001", body.Framework)
}

// --- HandleUpcoming ---

func TestHandleUpcomingEmpty(t *testing.T) {
	q := newTestQueries()
	w := doRequest(HandleUpcoming(q), http.MethodGet, "/api/v1/upcoming", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []UpcomingTest
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body)
}

func TestHandleUpcomingWithScheduledTest(t *testing.T) {
	test := makeRestoreTest("scheduled-test")
	future := metav1.NewTime(time.Now().UTC().Add(2 * time.Hour))
	test.Status.NextRunAt = &future

	scheme := newTestScheme()
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(test).
		WithStatusSubresource(&restorev1alpha1.RestoreTest{}).
		Build()
	// Persist the status with NextRunAt via status update.
	require.NoError(t, c.Status().Update(context.Background(), test))
	q := NewQueries(c, license.NewManager(c))

	w := doRequest(HandleUpcoming(q), http.MethodGet, "/api/v1/upcoming", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var body []UpcomingTest
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 1)
	assert.Equal(t, "scheduled-test", body[0].Name)
}

// --- HandleCreateTest ---

func TestHandleCreateTestSuccess(t *testing.T) {
	q := newTestQueries()

	input := CreateTestInput{
		Name:       "new-test",
		Provider:   "velero",
		BackupName: "latest",
		Namespaces: []string{"staging"},
		Cron:       "0 2 * * *",
	}
	body, err := json.Marshal(input)
	require.NoError(t, err)

	w := doRequest(HandleCreateTest(q), http.MethodPost, "/api/v1/tests", body)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "test created", resp["message"])
}

func TestHandleCreateTestMissingName(t *testing.T) {
	q := newTestQueries()

	input := CreateTestInput{Provider: "velero", Namespaces: []string{"staging"}}
	body, err := json.Marshal(input)
	require.NoError(t, err)

	w := doRequest(HandleCreateTest(q), http.MethodPost, "/api/v1/tests", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateTestMissingNamespaces(t *testing.T) {
	q := newTestQueries()

	input := CreateTestInput{Name: "test", Provider: "velero"}
	body, err := json.Marshal(input)
	require.NoError(t, err)

	w := doRequest(HandleCreateTest(q), http.MethodPost, "/api/v1/tests", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateTestInvalidProvider(t *testing.T) {
	q := newTestQueries()

	input := CreateTestInput{Name: "test", Provider: "kasten", Namespaces: []string{"staging"}}
	body, err := json.Marshal(input)
	require.NoError(t, err)

	w := doRequest(HandleCreateTest(q), http.MethodPost, "/api/v1/tests", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateTestInvalidJSON(t *testing.T) {
	q := newTestQueries()

	w := doRequest(HandleCreateTest(q), http.MethodPost, "/api/v1/tests", []byte("not json"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- HandleUpdateTest ---

func TestHandleUpdateTestSuccess(t *testing.T) {
	test := makeRestoreTest("existing-test")
	q := newTestQueries(test)

	input := CreateTestInput{
		Provider:   "velero",
		BackupName: "new-backup",
		Namespaces: []string{"production"},
		Cron:       "0 4 * * *",
	}
	body, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/tests/existing-test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("name", "existing-test")
	w := httptest.NewRecorder()
	HandleUpdateTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "test updated", resp["message"])
}

func TestHandleUpdateTestNotFound(t *testing.T) {
	q := newTestQueries() // no objects seeded

	input := CreateTestInput{Provider: "velero", Namespaces: []string{"ns"}}
	body, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/tests/ghost", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("name", "ghost")
	w := httptest.NewRecorder()
	HandleUpdateTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- HandleDeleteTest ---

func TestHandleDeleteTestSuccess(t *testing.T) {
	test := makeRestoreTest("to-delete")
	q := newTestQueries(test)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tests/to-delete", nil)
	req.SetPathValue("name", "to-delete")
	w := httptest.NewRecorder()
	HandleDeleteTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandleDeleteTestMissingName(t *testing.T) {
	q := newTestQueries()

	// PathValue("name") returns "" when not set.
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tests/", nil)
	w := httptest.NewRecorder()
	HandleDeleteTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- HandleTriggerTest ---

func TestHandleTriggerTestSuccess(t *testing.T) {
	test := makeRestoreTest("trigger-me")
	lastRun := metav1.NewTime(time.Now().UTC().Add(-1 * time.Hour))
	test.Status.LastRunAt = &lastRun
	test.Status.Phase = restorev1alpha1.PhaseIdle

	scheme := newTestScheme()
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(test).
		WithStatusSubresource(&restorev1alpha1.RestoreTest{}).
		Build()
	require.NoError(t, c.Status().Update(context.Background(), test))
	q := NewQueries(c, license.NewManager(c))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests/trigger-me/trigger", nil)
	req.SetPathValue("name", "trigger-me")
	w := httptest.NewRecorder()
	HandleTriggerTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "test triggered", resp["message"])
}

func TestHandleTriggerTestNotFound(t *testing.T) {
	q := newTestQueries()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests/ghost/trigger", nil)
	req.SetPathValue("name", "ghost")
	w := httptest.NewRecorder()
	HandleTriggerTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleTriggerTestMissingName(t *testing.T) {
	q := newTestQueries()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests//trigger", nil)
	w := httptest.NewRecorder()
	HandleTriggerTest(q).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- CORSMiddleware ---

func TestCORSMiddlewarePreflightReturns204(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set("Origin", "https://dashboard.kymaros.io")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://dashboard.kymaros.io", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddlewarePassesThrough(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "https://dashboard.kymaros.io")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://dashboard.kymaros.io", w.Header().Get("Access-Control-Allow-Origin"))
}
