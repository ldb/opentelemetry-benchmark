package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/ldb/openetelemtry-benchmark/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// start listening for connections on port for ctl
// start listening for connections from SUT
// receive config
// dump config into log file
// run benchmark according to config

const instanceName = "localtesting"

func main() {
	cfg := config.WorkerConfig{
		MaxTraceDepth:  5,
		MaxNumberSpans: 1,
		MaxSpanLength:  1 * time.Second,
		MaxCoolDown:    1 * time.Second,
	}

	workerManager := worker.NewManager(instanceName, cfg)
	if err := workerManager.AddWorkers(500); err != nil {
		log.Fatalf("error adding workers: %v", err)
	}

	workerManager.Start()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	receiver := receiver{Name: instanceName, Address: ":7666"}

	if err := receiver.ReceiveTraces(workerManager.FinishTrace); err != nil {
		log.Fatalf("error receiving traces: %v", err)
	}
}
