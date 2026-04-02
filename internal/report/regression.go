package report

import (
	"context"
	"log/slog"
	"sort"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RegressionDetector compares current score against previous reports
type RegressionDetector struct {
	client client.Client
	logger *slog.Logger
}

// NewRegressionDetector creates a RegressionDetector
func NewRegressionDetector(c client.Client, logger *slog.Logger) *RegressionDetector {
	return &RegressionDetector{client: c, logger: logger}
}

// RegressionResult captures the outcome of regression detection
type RegressionResult struct {
	Detected      bool
	PreviousScore int
	CurrentScore  int
	Delta         int
}

// Detect checks if the current score is lower than the most recent previous RestoreReport
// for the same RestoreTest. Returns regression info.
func (d *RegressionDetector) Detect(ctx context.Context, test *restorev1alpha1.RestoreTest, currentScore int) (*RegressionResult, error) {
	// List all reports for this test
	var reportList restorev1alpha1.RestoreReportList
	if err := d.client.List(ctx, &reportList, client.InNamespace(test.Namespace)); err != nil {
		return nil, err
	}

	// Filter reports for this test and sort by creation time (newest first)
	var matching []restorev1alpha1.RestoreReport
	for _, rr := range reportList.Items {
		if rr.Spec.TestRef == test.Name {
			matching = append(matching, rr)
		}
	}

	sort.Slice(matching, func(i, j int) bool {
		return matching[i].CreationTimestamp.After(matching[j].CreationTimestamp.Time)
	})

	// Need at least 1 previous report (the current one hasn't been created yet)
	if len(matching) == 0 {
		return &RegressionResult{
			Detected:     false,
			CurrentScore: currentScore,
		}, nil
	}

	previousScore := matching[0].Status.Score
	delta := currentScore - previousScore

	result := &RegressionResult{
		Detected:      delta < -10, // regression if score dropped by more than 10 points
		PreviousScore: previousScore,
		CurrentScore:  currentScore,
		Delta:         delta,
	}

	if result.Detected {
		d.logger.WarnContext(ctx, "score regression detected",
			"test", test.Name,
			"previous", previousScore,
			"current", currentScore,
			"delta", delta,
		)
	}

	return result, nil
}
