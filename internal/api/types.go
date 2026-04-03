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

// SummaryResponse for GET /api/v1/summary
type SummaryResponse struct {
	AverageScore      float64 `json:"averageScore"`
	TotalTests        int     `json:"totalTests"`
	TestsLastNight    int     `json:"testsLastNight"`
	TotalFailures     int     `json:"totalFailures"`
	TotalPartial      int     `json:"totalPartial"`
	AverageRTO        string  `json:"averageRTO"`
	NamespacesCovered int     `json:"namespacesCovered"`
}

// DailySummary for GET /api/v1/summary/daily
type DailySummary struct {
	Date     string  `json:"date"`
	Score    float64 `json:"score"`
	Tests    int     `json:"tests"`
	Failures int     `json:"failures"`
}

// Alert for GET /api/v1/alerts
type Alert struct {
	Timestamp string `json:"timestamp"`
	TestName  string `json:"testName"`
	Namespace string `json:"namespace"`
	Score     int    `json:"score"`
	PrevScore int    `json:"prevScore"`
	Result    string `json:"result"`
	Message   string `json:"message"`
}

// ComplianceResponse for GET /api/v1/compliance
type ComplianceResponse struct {
	Framework         string         `json:"framework"`
	Period            string         `json:"period"`
	Status            string         `json:"status"`
	TestsExecuted     int            `json:"testsExecuted"`
	AverageScore      float64        `json:"averageScore"`
	NamespacesCovered string         `json:"namespacesCovered"`
	DaysWithTests     int            `json:"daysWithTests"`
	DaysInPeriod      int            `json:"daysInPeriod"`
	IssuesDetected    int            `json:"issuesDetected"`
	AverageRTO        string         `json:"averageRTO"`
	RTOTarget         string         `json:"rtoTarget"`
	RTOCompliant      bool           `json:"rtoCompliant"`
	DailyData         []DailySummary `json:"dailyData"`
}

// TestResponse wraps a RestoreTest for API output.
type TestResponse struct {
	Name             string   `json:"name"`
	Namespace        string   `json:"namespace"`
	Provider         string   `json:"provider"`
	Schedule         string   `json:"schedule"`
	Phase            string   `json:"phase"`
	LastScore        int      `json:"lastScore"`
	LastResult       string   `json:"lastResult"`
	LastRunAt        string   `json:"lastRunAt"`
	NextRunAt        string   `json:"nextRunAt"`
	SourceNamespaces []string `json:"sourceNamespaces"`
	RTOTarget        string   `json:"rtoTarget,omitempty"`
}

// UpcomingTest for the "next tests" section.
type UpcomingTest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	NextRunAt string `json:"nextRunAt"`
	LastScore int    `json:"lastScore"`
}

// CreateTestInput is the JSON payload for creating or updating a RestoreTest.
type CreateTestInput struct {
	Name               string   `json:"name"`
	Provider           string   `json:"provider"`           // "velero"
	BackupName         string   `json:"backupName"`         // "latest" or specific
	Namespaces         []string `json:"namespaces"`         // source namespace names
	Cron               string   `json:"cron"`               // "0 3 * * *"
	Timezone           string   `json:"timezone"`           // "UTC"
	SandboxPrefix      string   `json:"sandboxPrefix"`      // "kymaros-test"
	TTL                string   `json:"ttl"`                // "30m"
	NetworkIsolation   string   `json:"networkIsolation"`   // "strict"
	QuotaCPU           string   `json:"quotaCpu"`           // "2"
	QuotaMemory        string   `json:"quotaMemory"`        // "4Gi"
	HealthCheckRef     string   `json:"healthCheckRef"`     // policy name
	HealthCheckTimeout string   `json:"healthCheckTimeout"` // "10m"
	MaxRTO             string   `json:"maxRTO"`             // "15m"
	AlertOnExceed      bool     `json:"alertOnExceed"`
}

// LicenseResponse for GET /api/v1/license
type LicenseResponse struct {
	Tier          string      `json:"tier"`
	ExpiresAt     string      `json:"expiresAt,omitempty"`
	TrialEndsAt   string      `json:"trialEndsAt,omitempty"`
	IsExpired     bool        `json:"isExpired"`
	IsTrialing    bool        `json:"isTrialing"`
	TrialDaysLeft *int        `json:"trialDaysLeft,omitempty"`
	Features      any `json:"features"`
}
