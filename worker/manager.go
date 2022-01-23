package worker

import (
	"context"
	"errors"
	"github.com/ldb/openetelemtry-benchmark/config"
	"log"
	"os"
)

var ErrWorkerManagerStopped = errors.New("manager stopped")

// Manager manages a number of workers based on a config.WorkerConfig. New workers can be added during runtime.
// Workers cannot be removed during runtime. Once the manager is stopped, all workers are stopped.
// A stopped Manager can not be reused.
// Managers can be named. Their name reflects in collected worker metrics.
type Manager struct {
	name     string
	nWorkers int
	ctx      context.Context
	cancel   context.CancelFunc
	config   config.WorkerConfig
	// workers tracks active Workers.
	workers []*Worker
	// newWorkers is a list of newly added Workers that are not yet started.
	newWorkers []*Worker
	receiver   *receiver
	logger     Logger
	stopped    bool
}

// NewManager creates a new Manager based on a config.WorkerConfig.
func NewManager(name string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := new(Manager)
	m.name = name
	m.ctx = ctx
	m.cancel = cancel
	m.workers = make([]*Worker, 0)
	m.newWorkers = make([]*Worker, 0)

	m.logger = log.New(os.Stdout, "M ", log.Ltime|log.Lmicroseconds|log.LUTC)

	return m
}

func (m *Manager) Configure(config config.WorkerConfig) {
	m.config = config
	m.receiver = &receiver{Host: m.config.ReceiverAddress}
}

// AddWorkers adds n workers to the current pool of workers. Workers can be added at runtime.
func (m *Manager) AddWorkers(n int) {
	for i := 0; i < n; i++ {
		w := m.newWorker(m.nWorkers)

		m.nWorkers++
		m.newWorkers = append(m.newWorkers, w)
		activeWorkers.WithLabelValues(m.name).Inc()
	}
	m.logger.Println("AddWorkers", n, m.nWorkers)
}

func (m *Manager) newWorker(id int) *Worker {
	w := new(Worker)

	w.managerName = m.name
	w.ID = id
	w.Config = m.config

	w.initTracer()

	w.Logger = log.New(os.Stdout, "W ", log.Ltime|log.Lmicroseconds|log.LUTC)

	ch := make(chan struct{}, 1)
	w.FinishTrace = ch
	return w
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
	go func() {
		if err := m.receiver.ReceiveTraces(m.finishTrace); err != nil {
			m.logger.Println("error receiving traces: %v", err)
		}
	}()
	for _, w := range m.newWorkers {
		go m.startAndWatch(m.ctx, w)
		m.workers = append(m.workers, w)
	}
	// After all m.newWorkers are added to m.workers, we reset m.newWorkers.
	m.newWorkers = make([]*Worker, 0)
}

// Stop stops the manager and all its workers. A stopped manager cannot be reused.
func (m *Manager) Stop() {
	m.cancel()
	m.stopped = true
}

// finishTrace notifies the worker with ID id that a trace was received so that it can stop it's timer.
// It returns ErrWorkerManagerStopped if the manager itself has stopped.
func (m *Manager) finishTrace(id int) error {
	if m.stopped {
		return ErrWorkerManagerStopped
	}
	m.workers[id].FinishTrace <- struct{}{}
	return nil
}
