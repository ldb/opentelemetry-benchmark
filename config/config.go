package config

import "time"

// BenchConfig describes the configuration for a Benchmark run.
type BenchConfig struct {
	StartTime    time.Time       `json:"startTime" yaml:"startTime"`
	WorkerConfig WorkerConfig    `json:"workerConfig" yaml:"workerConfig"`
	Steps        []BenchmarkStep `json:"steps" yaml:"steps"`
}

type WorkerConfig struct {
	TraceDepth  int           `json:"traceDepth" yaml:"traceDepth"`   // How deeply the generate spans should be nested.
	NumberSpans int           `json:"numberSpans" yaml:"numberSpans"` // How many simultanous spans to generate per trace.
	SpanLength  time.Duration `json:"spanLength" yaml:"spanLength"`
	MaxCoolDown time.Duration `json:"maxCoolDown" yaml:"maxCoolDown"` // Maximum random cooldown between requests.
}

type BenchmarkStep struct {
	StartTime     time.Time `json:"startTime" yaml:"startTime"`
	NumberWorkers int       `json:"numberWorkers" yaml:"numberWorkers"`
}
