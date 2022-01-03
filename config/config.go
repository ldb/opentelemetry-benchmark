package config

import "time"

// BenchConfig describes the configuration for a Benchmark run.
type BenchConfig struct {
	StartTime    time.Time       `json:"startTime" yaml:"startTime"`
	WorkerConfig WorkerConfig    `json:"workerConfig" yaml:"workerConfig"`
	Steps        []BenchmarkStep `json:"steps" yaml:"steps"`
}

type WorkerConfig struct {
	MaxTraceDepth  int           `json:"maxTraceDepth" yaml:"maxTraceDepth"`   // How deeply the generate spans should be nested.
	MaxNumberSpans int           `json:"maxNumberSpans" yaml:"maxNumberSpans"` // How many simultanous spans to generate per trace.
	MaxSpanLength  time.Duration `json:"maxSpanLength" yaml:"maxSpanLength"`
	MaxCoolDown    time.Duration `json:"maxCoolDown" yaml:"maxCoolDown"` // Maximum random cooldown between requests.
}

type BenchmarkStep struct {
	StartTime     time.Time `json:"startTime" yaml:"startTime"`
	NumberWorkers int       `json:"numberWorkers" yaml:"numberWorkers"`
}
