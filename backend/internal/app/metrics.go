package app

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	healthMetricName                     = "health_status"
	panicRecoveryCounterName             = "panic_recovery_total"
	externalServiceErrorCounterName      = "external_service_error_total"
	errorsTotalCounterName               = "errors_total"
	idleViolationQueriesTotalCounterName = "idle_violation_queries_total_counter"

	healthMetricHelp                     = "Health status"
	panicRecoveryCounterHelp             = "Total number of recovered panics"
	externalServiceErrorCounterHelp      = "Total number of external service errors"
	errorsTotalCounterHelp               = "Total number of errors"
	idleViolationQueriesTotalCounterHelp = "The total number of queries that violated the idle"
)

type metrics struct {
	PanicRecoveryCounterMetric        *prometheus.CounterVec
	CallDurationMetric                *prometheus.HistogramVec
	ExternalServiceErrorCounterMetric *prometheus.CounterVec
	ErrorsTotalCounterMetric          *prometheus.CounterVec
	HealthMetric                      prometheus.Gauge
	IdleViolationQueriesCounterMetric *prometheus.CounterVec
}

func initializeMetrics() metrics {
	var m metrics

	m.HealthMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: healthMetricName,
		Help: healthMetricHelp,
	})

	m.PanicRecoveryCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: panicRecoveryCounterName,
		Help: panicRecoveryCounterHelp,
	}, []string{"module"})

	m.ExternalServiceErrorCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: externalServiceErrorCounterName,
		Help: externalServiceErrorCounterHelp,
	}, []string{"service", "error_type"})

	m.ErrorsTotalCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: errorsTotalCounterName,
		Help: errorsTotalCounterHelp,
	}, []string{"module", "error_type"})

	m.IdleViolationQueriesCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: idleViolationQueriesTotalCounterName,
		Help: idleViolationQueriesTotalCounterHelp,
	}, []string{"cluster", "version"})

	return m
}
