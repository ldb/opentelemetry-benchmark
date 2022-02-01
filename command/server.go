package command

import (
	"encoding/json"
	"errors"
	"github.com/ldb/openetelemtry-benchmark/benchmark"
	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const defaultHost = ":7666"

// Server is the main communication component for `benchctl`.
type Server struct {
	Host       string
	s          *http.Server
	benchmarks map[string]*benchmark.Benchmark
	init       sync.Once
}

// Start starts the commandServer after initializing it exactly once.
func (c *Server) Start() error {
	if c.Host == "" {
		c.Host = defaultHost
	}
	c.init.Do(func() {
		c.benchmarks = make(map[string]*benchmark.Benchmark)
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.Handle("/logs/", http.StripPrefix("/logs/", Gzip(http.FileServer(http.Dir(os.TempDir())))))
		mux.Handle("/create/", c.createHandler())
		mux.Handle("/configure/", c.configureHandler())
		mux.Handle("/start/", c.startHandler())
		mux.Handle("/stop/", c.stopHandler())
		mux.Handle("/status/", c.statusHandler())
		mux.Handle("/destroy/", c.destroyHandler())

		c.s = &http.Server{
			Addr:    c.Host,
			Handler: Log(mux),
		}
	})
	log.Println("starting server")
	return c.s.ListenAndServe()
}

// createHandler handles HTTP requests to create a new Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *Server) createHandler() http.HandlerFunc {
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
		b := &benchmark.Benchmark{Name: benchmarkName}
		c.benchmarks[benchmarkName] = b
		status := b.Status()
		e := json.NewEncoder(writer)
		if err := e.Encode(status); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("created benchmark", benchmarkName)
	}
}

// configureHandler handles configuring an existing Benchmark with a config.BenchConfig.
// The last component of the HTTP Path is used as the Benchmark name.
// It expects a JSON encoded config.BenchConfig as HTTP Body.
// If a Benchmark with the provided name does not exist, it is transparently created.
func (c *Server) configureHandler() http.HandlerFunc {
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
			b = &benchmark.Benchmark{Name: benchmarkName}
			c.benchmarks[benchmarkName] = b
		}
		b.Configure(rb)
		status := b.Status()
		e := json.NewEncoder(writer)
		if err := e.Encode(status); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("configured benchmark", benchmarkName)
	}
}

// startHandler handles starting an existing, configured Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *Server) startHandler() http.HandlerFunc {
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
		status := b.Status()
		e := json.NewEncoder(writer)
		if err := e.Encode(status); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("started benchmark", benchmarkName)
	}
}

// stopHandler handles stopping an existing, running Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *Server) stopHandler() http.HandlerFunc {
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
		status := b.Status()
		e := json.NewEncoder(writer)
		if err := e.Encode(status); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("stopped benchmark", benchmarkName)
	}
}

// statusHandler handles getting the status of an existing Benchmark.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *Server) statusHandler() http.HandlerFunc {
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
		e := json.NewEncoder(writer)
		if err := e.Encode(status); err != nil {
			log.Printf("error cre")
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// destroyHandler handles destroying an existing Benchmark.
// If it is not stopped, it is stopped automatically.
// The last component of the HTTP Path is used as the Benchmark name.
func (c *Server) destroyHandler() http.HandlerFunc {
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
		if err := b.Destroy(); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		delete(c.benchmarks, benchmarkName)
		log.Println("destroyed benchmark", benchmarkName, c.benchmarks)
	}
}

func nameFromPath(path string) (string, error) {
	p := strings.Split(path, "/")
	if len(p) < 2 {
		return "", errors.New("invalid path")
	}
	return p[len(p)-1], nil
}
