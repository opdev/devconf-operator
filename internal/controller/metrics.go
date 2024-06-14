package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	upgrades = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "upgrades_total",
			Help: "Number of upgrades processed",
		},
	)
	upgradesFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "upgrades_failures_total",
			Help: "Number of failed upgrades",
		},
	)

)
func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(upgrades, upgradesFailures)
}
