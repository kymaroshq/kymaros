package report

// ScoreInput contains the inputs for the score calculation
type ScoreInput struct {
	RestoreSucceeded      bool
	CompletenessRatio     float64
	PodsReadyRatio        float64
	HealthChecksPassRatio float64
	DepsCoverageRatio     float64
	RTOWithinSLA          bool
}

// CalculateScore computes a 0-100 confidence score using 6 weighted levels:
//   - Level 1: Restore integrity (25 points)
//   - Level 2: Completeness (20 points)
//   - Level 3: Pod startup (20 points)
//   - Level 4: Health checks (20 points)
//   - Level 5: Dependency coverage (10 points)
//   - Level 6: RTO compliance (5 points)
func CalculateScore(r *ScoreInput) int {
	score := 0

	// Level 1: Restore integrity (25 points)
	if r.RestoreSucceeded {
		score += 25
	}

	// Level 2: Completeness (20 points)
	score += int(r.CompletenessRatio * 20)

	// Level 3: Pod startup (20 points)
	score += int(r.PodsReadyRatio * 20)

	// Level 4: Health checks (20 points)
	score += int(r.HealthChecksPassRatio * 20)

	// Level 5: Dependency coverage (10 points)
	score += int(r.DepsCoverageRatio * 10)

	// Level 6: RTO compliance (5 points)
	if r.RTOWithinSLA {
		score += 5
	}

	if score > 100 {
		return 100
	}
	return score
}
