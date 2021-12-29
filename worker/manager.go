package worker

import (
	"context"
	"log"
	"os"
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
	workers        map[int]*Worker
	logger         Logger
}

func New(ctx context.Context, config config.WorkerConfig) *Manager {
	if ctx == nil {
		ctx = context.Background()
	}

	m := new(Manager)
	m.ctx = ctx
	m.config = config
	m.tracerProvider = trace.NewNoopTracerProvider() //TODO: replace with actual provider

	m.logger = log.New(os.Stdout, "M", log.Ltime|log.Lmicroseconds|log.LUTC)

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

		w.Logger = log.New(os.Stdout, "W", log.Ltime|log.Lmicroseconds|log.LUTC)

		ch := make(chan struct{}, 1)
		w.FinishTrace = ch

		m.Lock()
		m.nWorkers++
		m.workers[w.ID] = w
		m.Unlock()
	}

	return nil
}

// startAndWatch is a simple wrapper that restarts a worker should it exit for any reason other than being canceled.
func (m *Manager) startAndWatch(ctx context.Context, w *Worker) {
	for {
		err := w.Run(ctx)
		if err != nil && err == context.Canceled {
			// We canceled this worker ourselves, so we should not restart it.
			break
		}
	}
}

// Start runs all added workers concurrently.
// In case of any failure a worker is restarted indefinitely until formally canceled via its context.
func (m *Manager) Start() {
	for _, w := range m.workers {
		go m.startAndWatch(m.ctx, w)
	}
}

// Stop stops all workers and destroys them. Stopped workers cannot be restarted. They have to be added again.
func (m *Manager) Stop() {

}

// FinishTrace notifies the worker with ID id that a trace was received so that it can stop it's timer.
func (m *Manager) FinishTrace(id int) {

}

