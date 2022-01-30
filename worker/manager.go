package worker

import (
	"context"
	"errors"
	"github.com/ldb/openetelemtry-benchmark/config"
	"io"
	"log"
	"sync"
	"time"
)

var ErrWorkerManagerStopped = errors.New("manager statusStopped")

// Manager manages a number of workers based on a config.WorkerConfig. New workers can be added during runtime.
// Workers cannot be removed during runtime. Once the manager is statusStopped, all workers are statusStopped.
// A statusStopped Manager can not be reused.
// Managers can be named. Their name reflects in collected worker metrics.
type Manager struct {
	name     string
	nWorkers int
	ctx      context.Context
	cancel   context.CancelFunc
	config   config.WorkerConfig
	// workers tracks active Workers.
	workers []*Worker
	// newWorkers is a list of newly added Workers that are not yet statusInitialized.
	newWorkers           []*Worker
	receiver             *receiver
	receiverShutdownFunc func(ctx context.Context) error
	logger               Logger
	stopped              bool
	// Errors that have occured thus far, not including Workers being shut down.
	errors    int
	mu        sync.RWMutex
	logWriter io.Writer
}

// NewManager creates a new Manager based on a config.WorkerConfig.
func NewManager(name string, writer io.Writer) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := new(Manager)
	m.name = name
	m.ctx = ctx
	m.cancel = cancel
	m.workers = make([]*Worker, 0)
	m.newWorkers = make([]*Worker, 0)
	m.logWriter = writer

	m.logger = log.New(writer, "M "+name+" ", log.Ltime|log.Lmicroseconds|log.LUTC)

	return m
}

func (m *Manager) Configure(config config.WorkerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
	m.receiver = &receiver{Host: m.config.ReceiverAddress}
}

// AddWorkers adds n workers to the current pool of workers. Workers can be added at runtime.
func (m *Manager) AddWorkers(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < n; i++ {
		w := m.newWorker(m.nWorkers)

		m.nWorkers++
		m.newWorkers = append(m.newWorkers, w)
		activeWorkers.WithLabelValues(m.name).Inc()
	}
	m.logger.Println("AddWorkers", n, m.nWorkers)
	// We add all workers before starting them to make sure they are all properly initialized.
	for _, w := range m.newWorkers {
		go m.startAndWatch(m.ctx, w)
		m.workers = append(m.workers, w)
	}
	// After all m.newWorkers are added to m.workers, we reset m.newWorkers.
	m.newWorkers = make([]*Worker, 0)
}

func (m *Manager) newWorker(id int) *Worker {
	w := new(Worker)
	w.managerName = m.name
	w.ID = id
	w.Config = m.config
	w.Logger = log.New(m.logWriter, "W "+m.name+" ", log.Ltime|log.Lmicroseconds|log.LUTC)
	ch := make(chan struct{}, 1)
	w.FinishTrace = ch
	w.initTracer(w.Config.Target)
	return w
}

// startAndWatch is a simple wrapper that restarts a worker should it exit for any reason other than being canceled.
func (m *Manager) startAndWatch(ctx context.Context, w *Worker) {
	for {
		err := w.Run(ctx)
		if err != nil {
			if err == context.Canceled {
				// We canceled this worker ourselves, so we should not restart it.
				break
			}
			m.mu.Lock()
			m.errors += 1
			m.mu.Unlock()
			if err == context.DeadlineExceeded {
				// We don't need to log timeouts, the worker already does this.
				continue
			}
			m.logger.Printf("W %d err: %v", w.ID, err)
		}

	}
}

// Start runs all added workers concurrently.
// In case of any failure a worker is restarted indefinitely until formally canceled via its context.
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return
	}
	go func() {
		shutdown, listenAndServe := m.receiver.ReceiveTraces(m.finishTrace)
		m.receiverShutdownFunc = shutdown
		if err := listenAndServe(); err != nil {
			m.logger.Printf("error receiving traces: %v", err)
		}
	}()
}

// Stop stops the manager and all its workers. A statusStopped manager cannot be reused.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.receiverShutdownFunc(ctx)
	time.Sleep(2 * time.Second) // Wait a short time so that all workers finish writing.
	m.nWorkers = 0
	m.stopped = true
}

// finishTrace notifies the worker with ID id that a trace was received so that it can stop it's timer.
// It returns ErrWorkerManagerStopped if the manager itself has statusStopped.
func (m *Manager) finishTrace(id int) error {
	if m.stopped {
		return ErrWorkerManagerStopped
	}
	m.workers[id].FinishTrace <- struct{}{}
	return nil
}

func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Status{
		ActiveWorkers: m.nWorkers,
		Errors:        m.errors,
	}
}
