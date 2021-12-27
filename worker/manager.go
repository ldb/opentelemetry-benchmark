package worker

import (
	"context"
	"sync"

	"github.com/ldb/openetelemtry-benchmark/config"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "otel-benchmark"

type Manager struct {
	nWorkers int
	sync.Mutex
	ctx            context.Context
	config         config.WorkerConfig
	tracerProvider trace.TracerProvider
	finishMap      map[int]chan struct{} // maps worker ID to their respective FinishTrace channels

}

func New(ctx context.Context, config config.WorkerConfig) *Manager {
	m := new(Manager)
	m.ctx = ctx
	m.config = config
	m.tracerProvider = trace.NewNoopTracerProvider() //TODO: replace with actual provider
	return m
}

// AddWorkers adds n workers to the current pool of workers
func (m *Manager) AddWorkers(n int) error {
	for i := 0; i <= n; i++ {
		w := new(Worker)
		w.ID = m.nWorkers + 1
		w.TraceDepth = m.config.TraceDepth
		w.NumberSpans = m.config.NumberSpans
		w.SpanLength = m.config.SpanLength
		w.MaxCoolDown = m.config.MaxCoolDown

		w.Tracer = m.tracerProvider.Tracer(instrumentationName)

		ch := make(chan struct{}, 1)

		w.FinishTrace = ch

		m.Lock()
		m.nWorkers++
		m.finishMap[w.ID] = ch
		m.Unlock()
	}

	return nil
}

// startAndWatch is a simple wrapper that restarts a worker should it exit for any reason.
func (m *Manager) startAndWatch(ctx context.Context, w *Worker) {
	for {
		err := w.Run(ctx)
		if err != nil && err == context.Canceled {
			// We canceled this worker ourselves, so we should not restart it.
			break
		}
	}
}
