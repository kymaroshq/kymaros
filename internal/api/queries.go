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
	"context"
	"fmt"
	"sort"
	"time"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"github.com/kymaroshq/kymaros/internal/adapter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const kymarosNamespace = "kymaros-system"

// Queries provides access to Kymaros CRDs for the REST API.
type Queries struct {
	client client.Client
}

// NewQueries creates a new Queries instance backed by the given controller-runtime client.
func NewQueries(c client.Client) *Queries {
	return &Queries{client: c}
}

// ListTests returns all RestoreTests in the kymaros-system namespace.
func (q *Queries) ListTests(ctx context.Context) ([]restorev1alpha1.RestoreTest, error) {
	var list restorev1alpha1.RestoreTestList
	if err := q.client.List(ctx, &list, client.InNamespace(kymarosNamespace)); err != nil {
		return nil, fmt.Errorf("listing RestoreTests: %w", err)
	}
	return list.Items, nil
}

// ListReports returns RestoreReports in the kymaros-system namespace, optionally
// filtered by test name and limited to the last N days. When latestOnly is true
// only the most recent report per test is returned. Results are sorted by
// creation timestamp descending.
func (q *Queries) ListReports(ctx context.Context, testName string, days int, latestOnly bool) ([]restorev1alpha1.RestoreReport, error) {
	var list restorev1alpha1.RestoreReportList
	if err := q.client.List(ctx, &list, client.InNamespace(kymarosNamespace)); err != nil {
		return nil, fmt.Errorf("listing RestoreReports: %w", err)
	}

	cutoff := time.Time{}
	if days > 0 {
		cutoff = time.Now().UTC().AddDate(0, 0, -days)
	}

	filtered := make([]restorev1alpha1.RestoreReport, 0, len(list.Items))
	for i := range list.Items {
		r := &list.Items[i]
		if testName != "" && r.Spec.TestRef != testName {
			continue
		}
		if !cutoff.IsZero() && r.CreationTimestamp.Time.Before(cutoff) {
			continue
		}
		filtered = append(filtered, *r)
	}

	sortReportsDesc(filtered)

	if latestOnly {
		latest := make(map[string]restorev1alpha1.RestoreReport)
		for i := range filtered {
			r := &filtered[i]
			if _, ok := latest[r.Spec.TestRef]; !ok {
				latest[r.Spec.TestRef] = *r
			}
		}
		result := make([]restorev1alpha1.RestoreReport, 0, len(latest))
		for _, r := range latest {
			result = append(result, r)
		}
		sortReportsDesc(result)
		return result, nil
	}

	return filtered, nil
}

// GetLatestReports returns the most recent RestoreReport for each RestoreTest.
func (q *Queries) GetLatestReports(ctx context.Context) ([]restorev1alpha1.RestoreReport, error) {
	return q.ListReports(ctx, "", 0, true)
}

// GetSummary computes the aggregate summary from the latest reports and tests.
func (q *Queries) GetSummary(ctx context.Context) (SummaryResponse, error) {
	tests, err := q.ListTests(ctx)
	if err != nil {
		return SummaryResponse{}, err
	}

	latestReports, err := q.GetLatestReports(ctx)
	if err != nil {
		return SummaryResponse{}, err
	}

	// Recent reports from the last 24 hours for "last night" stats.
	recentReports, err := q.ListReports(ctx, "", 1, false)
	if err != nil {
		return SummaryResponse{}, err
	}

	// Average score from latest reports.
	var totalScore float64
	for i := range latestReports {
		totalScore += float64(latestReports[i].Status.Score)
	}
	avgScore := 0.0
	if len(latestReports) > 0 {
		avgScore = totalScore / float64(len(latestReports))
	}

	// Count failures and partials in recent reports.
	var failures, partials int
	for i := range recentReports {
		switch recentReports[i].Status.Result {
		case restorev1alpha1.ResultFail:
			failures++
		case restorev1alpha1.ResultPartial:
			partials++
		}
	}

	// Average RTO from latest reports.
	var totalRTO time.Duration
	var rtoCount int
	for i := range latestReports {
		measured := latestReports[i].Status.RTO.Measured.Duration
		if measured > 0 {
			totalRTO += measured
			rtoCount++
		}
	}
	avgRTO := ""
	if rtoCount > 0 {
		avgRTO = formatDuration(totalRTO / time.Duration(rtoCount))
	}

	// Count distinct source namespaces across all tests.
	nsSet := make(map[string]struct{})
	for i := range tests {
		for _, ns := range tests[i].Spec.BackupSource.Namespaces {
			nsSet[ns.Name] = struct{}{}
		}
	}

	return SummaryResponse{
		AverageScore:      avgScore,
		TotalTests:        len(tests),
		TestsLastNight:    len(recentReports),
		TotalFailures:     failures,
		TotalPartial:      partials,
		AverageRTO:        avgRTO,
		NamespacesCovered: len(nsSet),
	}, nil
}

// GetDailyScores groups reports by date and computes the average score per day.
func (q *Queries) GetDailyScores(ctx context.Context, days int) ([]DailySummary, error) {
	reports, err := q.ListReports(ctx, "", days, false)
	if err != nil {
		return nil, err
	}

	type dayAccum struct {
		totalScore float64
		tests      int
		failures   int
	}
	byDate := make(map[string]*dayAccum)

	for i := range reports {
		r := &reports[i]
		date := r.CreationTimestamp.Time.UTC().Format("2006-01-02")
		acc, ok := byDate[date]
		if !ok {
			acc = &dayAccum{}
			byDate[date] = acc
		}
		acc.totalScore += float64(r.Status.Score)
		acc.tests++
		if r.Status.Result == restorev1alpha1.ResultFail {
			acc.failures++
		}
	}

	result := make([]DailySummary, 0, len(byDate))
	for date, acc := range byDate {
		avg := 0.0
		if acc.tests > 0 {
			avg = acc.totalScore / float64(acc.tests)
		}
		result = append(result, DailySummary{
			Date:     date,
			Score:    avg,
			Tests:    acc.tests,
			Failures: acc.failures,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// GetAlerts returns alerts for reports within the given time window that either
// failed or show a score regression (current score more than 10 points below
// the previous score for the same test).
func (q *Queries) GetAlerts(ctx context.Context, hours int) ([]Alert, error) {
	// Fetch enough history to detect regressions: the window itself plus extra
	// context so we can find the "previous" report for comparison.
	lookbackDays := (hours / 24) + 2
	allReports, err := q.ListReports(ctx, "", lookbackDays, false)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	// Group reports by test, already sorted desc by creation time.
	byTest := make(map[string][]restorev1alpha1.RestoreReport)
	for i := range allReports {
		ref := allReports[i].Spec.TestRef
		byTest[ref] = append(byTest[ref], allReports[i])
	}

	var alerts []Alert
	for testName, testReports := range byTest {
		for i, r := range testReports {
			if r.CreationTimestamp.Time.Before(cutoff) {
				continue
			}

			var prevScore int
			if i+1 < len(testReports) {
				prevScore = testReports[i+1].Status.Score
			}

			isFail := r.Status.Result == restorev1alpha1.ResultFail
			isRegression := prevScore > 0 && r.Status.Score < prevScore-10

			if !isFail && !isRegression {
				continue
			}

			msg := ""
			if isFail {
				msg = fmt.Sprintf("Test %s failed with score %d", testName, r.Status.Score)
			}
			if isRegression {
				if msg != "" {
					msg += "; "
				}
				msg += fmt.Sprintf("score regression from %d to %d", prevScore, r.Status.Score)
			}

			alerts = append(alerts, Alert{
				Timestamp: r.CreationTimestamp.Time.UTC().Format(time.RFC3339),
				TestName:  testName,
				Namespace: r.Namespace,
				Score:     r.Status.Score,
				PrevScore: prevScore,
				Result:    r.Status.Result,
				Message:   msg,
			})
		}
	}

	// Sort alerts by timestamp descending.
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Timestamp > alerts[j].Timestamp
	})

	return alerts, nil
}

// GetUpcomingTests returns tests whose NextRunAt is in the future, sorted by
// next run time ascending.
func (q *Queries) GetUpcomingTests(ctx context.Context) ([]UpcomingTest, error) {
	tests, err := q.ListTests(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var upcoming []UpcomingTest
	for i := range tests {
		t := &tests[i]
		if t.Status.NextRunAt == nil || t.Status.NextRunAt.Time.Before(now) {
			continue
		}
		upcoming = append(upcoming, UpcomingTest{
			Name:      t.Name,
			Namespace: t.Namespace,
			NextRunAt: t.Status.NextRunAt.Time.UTC().Format(time.RFC3339),
			LastScore: t.Status.LastScore,
		})
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].NextRunAt < upcoming[j].NextRunAt
	})

	return upcoming, nil
}

// MapTestToResponse converts a RestoreTest CRD into a TestResponse suitable for
// the REST API.
func MapTestToResponse(test restorev1alpha1.RestoreTest) TestResponse {
	sourceNS := make([]string, 0, len(test.Spec.BackupSource.Namespaces))
	for _, ns := range test.Spec.BackupSource.Namespaces {
		sourceNS = append(sourceNS, ns.Name)
	}

	lastRunAt := ""
	if test.Status.LastRunAt != nil {
		lastRunAt = test.Status.LastRunAt.Time.UTC().Format(time.RFC3339)
	}

	nextRunAt := ""
	if test.Status.NextRunAt != nil {
		nextRunAt = test.Status.NextRunAt.Time.UTC().Format(time.RFC3339)
	}

	rtoTarget := ""
	if test.Spec.SLA != nil && test.Spec.SLA.MaxRTO.Duration > 0 {
		rtoTarget = formatDuration(test.Spec.SLA.MaxRTO.Duration)
	}

	return TestResponse{
		Name:             test.Name,
		Namespace:        test.Namespace,
		Provider:         test.Spec.BackupSource.Provider,
		Schedule:         test.Spec.Schedule.Cron,
		Phase:            test.Status.Phase,
		LastScore:        test.Status.LastScore,
		LastResult:       test.Status.LastResult,
		LastRunAt:        lastRunAt,
		NextRunAt:        nextRunAt,
		SourceNamespaces: sourceNS,
		RTOTarget:        rtoTarget,
	}
}

// sortReportsDesc sorts reports by creation timestamp descending (newest first).
func sortReportsDesc(reports []restorev1alpha1.RestoreReport) {
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].CreationTimestamp.After(reports[j].CreationTimestamp.Time)
	})
}

// parseDuration parses a Go/K8s duration string like "15m0s" into a
// time.Duration. Returns zero on parse failure.
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// CreateTest creates a new RestoreTest CR in the kymaros-system namespace.
func (q *Queries) CreateTest(ctx context.Context, input CreateTestInput) error {
	test := buildRestoreTest(input)
	return q.client.Create(ctx, &test)
}

// UpdateTest updates an existing RestoreTest's spec.
func (q *Queries) UpdateTest(ctx context.Context, name string, input CreateTestInput) error {
	var test restorev1alpha1.RestoreTest
	key := types.NamespacedName{Name: name, Namespace: kymarosNamespace}
	if err := q.client.Get(ctx, key, &test); err != nil {
		return fmt.Errorf("getting RestoreTest %q: %w", name, err)
	}

	test.Spec = buildRestoreTestSpec(input)
	return q.client.Update(ctx, &test)
}

// DeleteTest deletes a RestoreTest by name from the kymaros-system namespace.
func (q *Queries) DeleteTest(ctx context.Context, name string) error {
	test := restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kymarosNamespace,
		},
	}
	return q.client.Delete(ctx, &test)
}

// TriggerTest forces an immediate run by clearing LastRunAt and setting Phase to Idle.
func (q *Queries) TriggerTest(ctx context.Context, name string) error {
	var test restorev1alpha1.RestoreTest
	key := types.NamespacedName{Name: name, Namespace: kymarosNamespace}
	if err := q.client.Get(ctx, key, &test); err != nil {
		return fmt.Errorf("getting RestoreTest %q: %w", name, err)
	}

	test.Status.LastRunAt = nil
	test.Status.Phase = restorev1alpha1.PhaseIdle
	return q.client.Status().Update(ctx, &test)
}

// buildRestoreTest creates a full RestoreTest object from input.
func buildRestoreTest(input CreateTestInput) restorev1alpha1.RestoreTest {
	return restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: kymarosNamespace,
		},
		Spec: buildRestoreTestSpec(input),
	}
}

// buildRestoreTestSpec converts CreateTestInput into a RestoreTestSpec.
func buildRestoreTestSpec(input CreateTestInput) restorev1alpha1.RestoreTestSpec {
	namespaces := make([]restorev1alpha1.NamespaceMapping, 0, len(input.Namespaces))
	for _, ns := range input.Namespaces {
		namespaces = append(namespaces, restorev1alpha1.NamespaceMapping{Name: ns})
	}

	spec := restorev1alpha1.RestoreTestSpec{
		BackupSource: restorev1alpha1.BackupSource{
			Provider:   input.Provider,
			BackupName: input.BackupName,
			Namespaces: namespaces,
		},
		Schedule: restorev1alpha1.ScheduleConfig{
			Cron:     input.Cron,
			Timezone: input.Timezone,
		},
		Sandbox: restorev1alpha1.SandboxConfig{
			NamespacePrefix:  input.SandboxPrefix,
			TTL:              metav1.Duration{Duration: parseDuration(input.TTL)},
			NetworkIsolation: input.NetworkIsolation,
		},
	}

	if input.QuotaCPU != "" || input.QuotaMemory != "" {
		spec.Sandbox.ResourceQuota = &restorev1alpha1.ResourceQuotaConfig{
			CPU:    input.QuotaCPU,
			Memory: input.QuotaMemory,
		}
	}

	if input.HealthCheckRef != "" || input.HealthCheckTimeout != "" {
		spec.HealthChecks = restorev1alpha1.HealthCheckRef{
			PolicyRef: input.HealthCheckRef,
			Timeout:   metav1.Duration{Duration: parseDuration(input.HealthCheckTimeout)},
		}
	}

	if input.MaxRTO != "" {
		spec.SLA = &restorev1alpha1.SLAConfig{
			MaxRTO:        metav1.Duration{Duration: parseDuration(input.MaxRTO)},
			AlertOnExceed: input.AlertOnExceed,
		}
	}

	return spec
}

// GetSchedules discovers Velero backup schedules.
func (q *Queries) GetSchedules(ctx context.Context) ([]adapter.ScheduleInfo, error) {
	a := adapter.NewVeleroAdapter(q.client)
	return a.ListSchedules(ctx)
}

// formatDuration formats a duration into a human-readable string like "15m 30s".
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case hours > 0:
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}
