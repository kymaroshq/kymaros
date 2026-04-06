package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		input    ScoreInput
		expected int
	}{
		{
			name: "perfect score",
			input: ScoreInput{
				RestoreSucceeded:      true,
				CompletenessRatio:     1.0,
				PodsReadyRatio:        1.0,
				HealthChecksPassRatio: 1.0,
				DepsCoverageRatio:     1.0,
				RTOWithinSLA:          true,
			},
			expected: 100,
		},
		{
			name: "total failure",
			input: ScoreInput{
				RestoreSucceeded:      false,
				CompletenessRatio:     0,
				PodsReadyRatio:        0,
				HealthChecksPassRatio: 0,
				DepsCoverageRatio:     0,
				RTOWithinSLA:          false,
			},
			expected: 0,
		},
		{
			name: "restore succeeded but nothing else",
			input: ScoreInput{
				RestoreSucceeded:      true,
				CompletenessRatio:     0,
				PodsReadyRatio:        0,
				HealthChecksPassRatio: 0,
				DepsCoverageRatio:     0,
				RTOWithinSLA:          false,
			},
			expected: 25,
		},
		{
			name: "restore + completeness + pods",
			input: ScoreInput{
				RestoreSucceeded:      true,
				CompletenessRatio:     1.0,
				PodsReadyRatio:        1.0,
				HealthChecksPassRatio: 0,
				DepsCoverageRatio:     0,
				RTOWithinSLA:          false,
			},
			expected: 65,
		},
		{
			name: "partial scores",
			input: ScoreInput{
				RestoreSucceeded:      true,
				CompletenessRatio:     0.5,
				PodsReadyRatio:        0.75,
				HealthChecksPassRatio: 0.5,
				DepsCoverageRatio:     0.5,
				RTOWithinSLA:          true,
			},
			expected: 70, // 25 + int(0.5*20)=10 + int(0.75*20)=15 + int(0.5*20)=10 + int(0.5*10)=5 + 5 = 70
		},
		{
			name: "half completeness only",
			input: ScoreInput{
				RestoreSucceeded:      false,
				CompletenessRatio:     0.5,
				PodsReadyRatio:        0,
				HealthChecksPassRatio: 0,
				DepsCoverageRatio:     0,
				RTOWithinSLA:          false,
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateScore(&tt.input)
			assert.Equal(t, tt.expected, score)
		})
	}
}
