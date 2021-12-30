package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	activeWorkers = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "otelbench_manager_active_workers_count",
		Help: "The total number of currently active workers",
	}, []string{"name"})

	tracesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "otelbench_manager_traces_sent_count",
		Help: "The total number of traces generated and sent by workers",
	}, []string{"name"})

	traceRoundtrip = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "otelbench_manager_trace_roundtrip_duration_seconds",
		Help: "The duration of trace full trace roundtrip from being sent to being finished",
	}, []string{"name"})
)