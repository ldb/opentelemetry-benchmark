package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ldb/openetelemtry-benchmark/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "otel-benchmark"

var ErrWorkerNotFound = errors.New("worker not found")
var ErrWorkerManagerStopped = errors.New("manager stopped")

// Manager manages a number of workers based on a config.WorkerConfig. New workers can be added during runtime.
// Workers cannot be removed during runtime. Once the manager is stopped, all workers are stopped.
// A stopped Manager can not be reused.
// Managers can be named. Their name reflects in collected worker metrics.
type Manager struct {
	name           string
	nWorkers       int
	mx             sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	config         config.WorkerConfig
	TracerProvider trace.TracerProvider
	workers        map[int]*Worker
	logger         Logger
	stopped        bool
}

// NewManager creates a new Manager based on a config.WorkerConfig.
func NewManager(name string, config config.WorkerConfig) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := new(Manager)
	m.name = name
	m.ctx = ctx
	m.cancel = cancel
	m.config = config
	m.workers = make(map[int]*Worker)

	m.logger = log.New(os.Stdout, "M ", log.Ltime|log.Lmicroseconds|log.LUTC)

	return m
}

// AddWorkers adds n workers to the current pool of workers. Workers can be added at rutime.
func (m *Manager) AddWorkers(n int) error {
	for i := 1; i <= n; i++ {
		w := new(Worker)

		w.managerName = m.name
		w.ID = m.nWorkers + 1
		w.TraceDepth = m.config.TraceDepth
		w.NumberSpans = m.config.NumberSpans
		w.SpanLength = m.config.SpanLength
		w.MaxCoolDown = m.config.MaxCoolDown

		w.Tracer = otel.Tracer(fmt.Sprintf("%s-%d", m.name, i))
		if m.TracerProvider != nil {
			w.Tracer = m.TracerProvider.Tracer(fmt.Sprintf("%s-%d", m.name, i))
		}

		w.Logger = log.New(os.Stdout, "W ", log.Ltime|log.Lmicroseconds|log.LUTC)

		ch := make(chan struct{}, 1)
		w.FinishTrace = ch

		m.mx.Lock()
		m.nWorkers++
		m.workers[w.ID] = w
		m.mx.Unlock()
		activeWorkers.WithLabelValues(m.name).Inc()
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

// Stop stops the manager and all its workers. A stopped manager can not be used again.
func (m *Manager) Stop() {
	m.cancel()
	m.stopped = true
}

// FinishTrace notifies the worker with ID id that a trace was received so that it can stop it's timer.
// It returns ErrWorkerNotFound if the worker has exited already.
// It returns ErrWorkerManagerStopped if the manager itself has stopped.
func (m *Manager) FinishTrace(id int) error {
	if m.stopped {
		return ErrWorkerManagerStopped
	}
	w, ok := m.workers[id]
	if !ok {
		return ErrWorkerNotFound
	}
	w.FinishTrace <- struct{}{}
	return nil
}
