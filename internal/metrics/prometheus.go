package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	TestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kymaros_tests_total",
			Help: "Total restore tests executed",
		},
		[]string{"test", "result"},
	)

	Score = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kymaros_score",
			Help: "Latest restore test score (0-100)",
		},
		[]string{"test"},
	)

	RTOSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kymaros_rto_seconds",
			Help: "Measured RTO in seconds",
		},
		[]string{"test"},
	)

	TestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kymaros_test_duration_seconds",
			Help:    "Total test duration",
			Buckets: prometheus.ExponentialBuckets(60, 2, 8),
		},
		[]string{"test"},
	)

	BackupAge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kymaros_backup_age_seconds",
			Help: "Age of the last tested backup",
		},
		[]string{"test"},
	)
)

func init() {
	metrics.Registry.MustRegister(
		TestsTotal,
		Score,
		RTOSeconds,
		TestDuration,
		BackupAge,
	)
}
