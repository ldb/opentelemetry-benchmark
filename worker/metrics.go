package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	activeWorkers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "benchd_manager_active_workers_count",
		Help: "The total number of currently active workers",
	}, []string{"name"})

	tracesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "benchd_manager_traces_sent_count",
		Help: "The total number of traces generated and sent by all workers",
	}, []string{"name"})

	tracesReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "benchd_manager_traces_received_count",
		Help: "The total number of traces received by all workers",
	}, []string{"name"})

	workerErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "benchd_manager_worker_error_count",
		Help: "The total number of errors that occurred in all workers",
	}, []string{"name", "kind"})

	traceRoundtrip = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "benchd_worker_trace_roundtrip_duration_seconds",
		Help:       "The duration of trace full trace roundtrip from being sent to being finished",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
	}, []string{"name"})
)
