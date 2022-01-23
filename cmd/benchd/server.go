package main

import (
	"encoding/json"
	"errors"
	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultHost = ":7666"

// cmdServer is the main communication Component for `benchctl`.
// It contains a thin abstraction layer for managing multiple Benchmarks.
type cmdServer struct {
	Host       string
	s          *http.Server
	benchmarks map[string]*Benchmark
	init       sync.Once
}

// Start starts the commandServer after initializing it exactly once.
func (c *cmdServer) Start() error {
	if c.Host == "" {
		c.Host = defaultHost
	}
	c.init.Do(func() {
		c.benchmarks = make(map[string]*Benchmark)
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.Handle("/create/", c.createHandler())
		mux.Handle("/configure/", c.configureHandler())
		mux.Handle("/start/", c.startHandler())
		mux.Handle("/stop/", c.stopHandler())
		mux.Handle("/status/", c.statusHandler())

		c.s = &http.Server{
			Addr:         c.Host,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
			Handler:      mux,
		}
	})

	log.Println("# starting server")

	return c.s.ListenAndServe()
}

// createHandler handles HTTP requests to create a new Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *cmdServer) createHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}
		log.Println(request.URL.Path)
		benchmarkName, err := nameFromPath(request.URL.Path)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		c.benchmarks[benchmarkName] = &Benchmark{Name: benchmarkName}
		log.Println("# created benchmark", benchmarkName)
	}
}

// configureHandler handles configuring an existing Benchmark with a config.BenchConfig.
// The last component of the HTTP Path is used as the Benchmark name.
// It expects a JSON encoded config.BenchConfig as HTTP Body.
func (c *cmdServer) configureHandler() http.HandlerFunc {
	type requestBody = config.BenchConfig
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}
		benchmarkName, err := nameFromPath(request.URL.Path)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		rb := new(requestBody)
		d := json.NewDecoder(request.Body)
		if err := d.Decode(rb); err != nil {
			http.Error(writer, "error decoding body", http.StatusBadRequest)
			return
		}
		b, ok := c.benchmarks[benchmarkName]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		b.Config = rb
		log.Println("# configured benchmark", benchmarkName)
	}
}

// startHandler handles starting an existing, configured Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *cmdServer) startHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}
		benchmarkName, err := nameFromPath(request.URL.Path)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		b, ok := c.benchmarks[benchmarkName]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		if err := b.Start(); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("# started benchmark", benchmarkName)
	}
}

// stopHandler handles stopping an existing, running Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *cmdServer) stopHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}
		benchmarkName, err := nameFromPath(request.URL.Path)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		b, ok := c.benchmarks[benchmarkName]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		if err := b.Stop(); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		delete(c.benchmarks, benchmarkName)
		log.Println("# stopped benchmark", benchmarkName, c.benchmarks)
	}
}

// statusHandler handles getting the status of an existing Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *cmdServer) statusHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}
		benchmarkName, err := nameFromPath(request.URL.Path)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		b, ok := c.benchmarks[benchmarkName]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		status := b.Status()
		writer.Write([]byte(status))
	}
}

func nameFromPath(path string) (string, error) {
	p := strings.Split(path, "/")
	if len(p) < 2 {
		return "", errors.New("invalid path")
	}
	return p[len(p)-1], nil
}
