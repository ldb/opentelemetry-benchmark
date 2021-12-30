package main

import (
	"log"
	"time"

	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/ldb/openetelemtry-benchmark/worker"
)

// start listening for connections on port for ctl
// start listening for connections from SUT
// start OTEL receiver
// receive config
// dump config into log file
// initialize thread pool
// initialize n workers
// run benchmark according to config
// record timing for each sent / received request

func main() {

	cfg := config.WorkerConfig{
		TraceDepth:  1,
		NumberSpans: 1,
		SpanLength:  1 * time.Second,
		MaxCoolDown: 1 * time.Second,
	}

	workerManager := worker.NewManager("test", cfg)
	if err := workerManager.AddWorkers(100); err != nil {
		log.Fatalf("error adding workers: %v", err)
	}

	workerManager.Start()

	time.Sleep(2 * time.Second)
	for i := 1; i <= 100; i++ {
		if err := workerManager.FinishTrace(i); err != nil {
			log.Fatalf("error finishing trace for worker %d", i)
		}

	}

	workerManager.Stop()

	time.Sleep(3 * time.Second)
	if err := workerManager.FinishTrace(1); err != nil {
		log.Fatalf("error finishing trace for worker %v", err)

	}
}
