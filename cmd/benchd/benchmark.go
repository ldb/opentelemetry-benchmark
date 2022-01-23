package main

import (
	"errors"
	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/ldb/openetelemtry-benchmark/worker"
	"sync"
	"time"
)

// Benchmark wraps a single config.BenchConfig and a single worker.Manager.
type Benchmark struct {
	Name          string
	Config        *config.BenchConfig
	status        string
	workerManager *worker.Manager
	m             sync.RWMutex
}

func (b *Benchmark) Start() error {
	b.m.Lock()
	defer b.m.Unlock()
	if b.Config == nil {
		return errors.New("not configured")
	}
	if b.status != "" {
		return errors.New("already running or stopped")
	}
	b.workerManager = worker.NewManager(b.Name)
	b.workerManager.Configure(b.Config.WorkerConfig)
	b.workerManager.Start()
	go func() {
		for _, step := range b.Config.Steps {
			b.workerManager.AddWorkers(step.NumberWorkers)
			time.Sleep(step.Duration.Duration)
		}
	}()
	b.status = "running"
	return nil
}

func (b *Benchmark) Stop() error {
	b.m.Lock()
	defer b.m.Unlock()
	if b.status != "running" {
		return errors.New("not running")
	}
	b.workerManager.Stop()
	b.status = "stopped"
	return nil
}

func (b *Benchmark) Status() string {
	b.m.RLock()
	defer b.m.RUnlock()
	if b.status == "" {
		return "uninitialized"
	}
	return b.status
}
