package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/ldb/openetelemtry-benchmark/config"
	"io/ioutil"
	"os"
)

const (
	defaultCommandPort    = ":7666"
	defaultReceiverPort   = ":2113"
	defaultTargetPort     = ":4317"
	defaultMonitoringPort = ":9090"
)

func createPlan(configFilename, planFilename string) (config.BenchmarkPlan, error) {
	cf, err := os.Open(configFilename)
	if err != nil {
		return config.BenchmarkPlan{}, fmt.Errorf("error opening config file %q: %v", configFilename, err)
	}
	defer cf.Close()
	ctlConfig, err := config.NewFrom(cf)
	if err != nil {
		return config.BenchmarkPlan{}, fmt.Errorf("error parsing config file %q: %v", configFilename, err)
	}
	pf, err := os.Open(planFilename)
	if err != nil {
		return config.BenchmarkPlan{}, fmt.Errorf("error opening plan file %q: %v", planFilename, err)
	}
	defer pf.Close()
	bb, err := ioutil.ReadAll(pf)
	if err != nil {
		return config.BenchmarkPlan{}, fmt.Errorf("error reading plan file %q: %v", planFilename, err)
	}
	plan := config.BenchmarkPlan{}
	if err := yaml.Unmarshal(bb, &plan); err != nil {
		return config.BenchmarkPlan{}, fmt.Errorf("error parsing plan file %q: %v", planFilename, err)
	}
	plan.MonitoringEndpoint = ctlConfig.Monitoring + defaultMonitoringPort
	plan.BenchConfig.WorkerConfig.Target = ctlConfig.Target + defaultTargetPort
	plan.BenchConfig.WorkerConfig.ReceiverAddress = defaultReceiverPort

	// At the moment we only support a single benchmarking client making requests, as we otherwise need some kind of routing on the SUT.
	// So we simply take the first one.
	plan.ClientAddress = "http://" + ctlConfig.Clients[0] + defaultCommandPort
	return plan, nil
}
