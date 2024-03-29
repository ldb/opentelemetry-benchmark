package config

import (
	"encoding/json"
	"errors"
	"time"
)

// BenchmarkPlan describes a full benchmark that should be executed by `benchctl`.
// Notice how at the moment we only support a single client per plan, and thus, only a single benchmarking client.
type BenchmarkPlan struct {
	Name               string      `json:"name" yaml:"name"`
	ClientAddress      string      `json:"clientAddress" yaml:"clientAddress"`
	BenchConfig        BenchConfig `json:"benchConfig" yaml:"benchConfig"`
	Duration           Duration    `json:"duration" yaml:"duration"`
	MonitoringEndpoint string      `json:"monitoringEndpoint,omitempty" yaml:"monitoringEndpoint,omitempty"`
}

// BenchConfig describes the configuration for a Benchmark plan.
// It can be used in two ways: FixedRate and Step.
// In FixedRate mode, the benchmark will create new Workers at a constant rate until it is stopped.
// In Step mode, a sequence of scaling steps is executed.
// FixedRate mode can be used to quickly find a breaking point for the system under test, which can later be closely observed in Step mode.
type BenchConfig struct {
	WorkerConfig WorkerConfig    `json:"workerConfig" yaml:"workerConfig"`
	FixedRate    FixedRate       `json:"fixedRate" yaml:"fixedRate"`
	Steps        []BenchmarkStep `json:"steps" yaml:"steps"`
}

type WorkerConfig struct {
	Target          string   `json:"target" yaml:"target"`
	ReceiverAddress string   `json:"receiverAddress" yaml:"receiverAddress"`
	MaxTraceDepth   int      `json:"maxTraceDepth" yaml:"maxTraceDepth"`   // How deeply the generate spans should be nested.
	MaxNumberSpans  int      `json:"maxNumberSpans" yaml:"maxNumberSpans"` // How many simultanous spans to generate per trace. // Not used yet.
	MaxSpanLength   Duration `json:"maxSpanLength" yaml:"maxSpanLength"`
	MaxCoolDown     Duration `json:"maxCoolDown" yaml:"maxCoolDown"` // Maximum random cooldown between requests.
	SendTimeout     Duration `json:"sendTimeout" yaml:"sendTimeout"`
	ReceiveTimeout  Duration `json:"receiveTimeout" yaml:"receiveTimeout"`
	// RiskyAttribute is a special attribute that is added to a random span in the trace for the filter based benchmarks.
	RiskyAttributeProbability int `json:"riskyAttributeProbability" yaml:"riskyAttributeProbability"`
	MaxExtraAttributes        int `json:"maxExtraAttributes" yaml:"maxExtraAttributes"` // The maximum number of extra attributes to add to each span.
}

// FixedRate represents scaling at a fixed rate of NumberWorkers per Duration.
type FixedRate struct {
	NumberWorkers int      `json:"numberWorkers" yaml:"numberWorkers"`
	Duration      Duration `json:"duration" yaml:"duration"`
}

// BenchmarkStep represents a single scaling step, that creates NumberWorkers and takes Duration to complete.
type BenchmarkStep struct {
	Duration      Duration `json:"duration" yaml:"duration"`
	NumberWorkers int      `json:"numberWorkers" yaml:"numberWorkers"`
}

// Duration wraps time.Duration for it to implement json.(Un)Marshaler
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration{time.Duration(value)}
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration{tmp}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
