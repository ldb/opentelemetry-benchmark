package benchmark

import (
	"context"
	"errors"
	"fmt"
	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/ldb/openetelemtry-benchmark/worker"
	"os"
	"sync"
	"time"
)

// Benchmark wraps a single config.BenchConfig and a single worker.Manager.
type Benchmark struct {
	Name          string
	config        *config.BenchConfig
	status        string
	workerManager *worker.Manager
	m             sync.RWMutex
	currentStep   int
	ctx           context.Context
	cancel        context.CancelFunc
	logFile       *os.File
}

func (b *Benchmark) Start() error {
	b.m.Lock()
	defer b.m.Unlock()
	if b.config == nil {
		b.status = "uninitialized"
		return errors.New("uninitialized")
	}
	if b.status == "running" || b.status == "stopped" {
		return errors.New("already running or stopped")
	}

	f, err := os.CreateTemp("", "benchd-"+b.Name)
	if err != nil {
		return fmt.Errorf("error creating temporary file: %v", err)
	}
	b.logFile = f
	b.workerManager = worker.NewManager(b.Name, f)
	b.workerManager.Configure(b.config.WorkerConfig)
	b.workerManager.Start()
	ctx, cancel := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancel = cancel
	go func(ctx context.Context) {
		// If FixedRate was configured, we run this mode;
		if b.config.FixedRate.NumberWorkers > 0 {
			for {
				// The benchmark was stopped and we should attempt to create new Workers.
				if ctx.Err() != nil {
					return
				}
				b.currentStep += 1
				b.workerManager.AddWorkers(b.config.FixedRate.NumberWorkers)
				time.Sleep(b.config.FixedRate.Duration.Duration)
			}
		}
		// ... otherwise attempt to run Step mode.
		for i, step := range b.config.Steps {
			// The benchmark was stopped and we should attempt to create new Workers.
			if ctx.Err() != nil {
				return
			}
			b.currentStep = i + 1
			b.workerManager.AddWorkers(step.NumberWorkers)
			time.Sleep(step.Duration.Duration)
		}
		b.status = "finished"
		return
	}(ctx)
	b.status = "running"
	return nil
}

func (b *Benchmark) Configure(config *config.BenchConfig) {
	b.m.Lock()
	defer b.m.Unlock()
	b.config = config
	b.status = "configured"
}

func (b *Benchmark) Stop() error {
	b.m.Lock()
	defer b.m.Unlock()
	if b.status != "running" && b.status != "finished" {
		return errors.New("not running")
	}
	b.cancel()
	b.workerManager.Stop()
	if err := b.logFile.Close(); err != nil {
		return fmt.Errorf("error closing log file: %v", err)
	}
	b.status = "stopped"
	return nil
}

func (b *Benchmark) Destroy() error {
	b.m.Lock()
	defer b.m.Unlock()
	if b.status != "stopped" {
		if err := b.Stop(); err != nil {
			return fmt.Errorf("error stopping benchmark: %v", err)
		}
	}
	if err := os.Remove(b.logFile.Name()); err != nil {
		return fmt.Errorf("error deleting logfile %s: %v", b.logFile.Name(), err)
	}

	return nil
}

type Status struct {
	State        string        `json:"state"`
	CurrentStep  int           `json:"currentStep"`
	MaxStep      int           `json:"maxStep"`
	ManagerState worker.Status `json:"managerState"`
	LogFile      string        `json:"logFile"`
}

func (b *Benchmark) Status() Status {
	b.m.Lock()
	defer b.m.Unlock()
	if b.status == "uninitialized" || b.status == "" {
		b.status = "uninitialized"
		return Status{State: b.status}
	}
	s := Status{
		State:        b.status,
		CurrentStep:  b.currentStep,
		MaxStep:      len(b.config.Steps),
		ManagerState: worker.Status{},
		LogFile:      "",
	}
	if b.workerManager != nil {
		s.ManagerState = b.workerManager.Status()
	}
	if b.logFile != nil {
		s.LogFile = b.logFile.Name()
	}
	return s
}
