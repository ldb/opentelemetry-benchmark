package main

import (
	"fmt"
	"github.com/ldb/openetelemtry-benchmark/benchmark"
	"github.com/ldb/openetelemtry-benchmark/command"
	"github.com/ldb/openetelemtry-benchmark/config"
)

type stateMachine struct {
	plan    config.BenchmarkPlan
	current benchmark.Status
	client  command.Client
}

func (s *stateMachine) sync() error {
	state, err := s.client.Status(s.plan.Name)
	if err != nil {
		return fmt.Errorf("error syncing state: %v", err)
	}
	s.current = state
	return nil
}

type inputFunc func(input string) error

func (s *stateMachine) next() (inputFunc, error) {
	switch benchmark.StateFrom(s.current.State) {
	case benchmark.Unknown:
	case benchmark.Uninitialized:
	case benchmark.Configured:
	case benchmark.Running:
	case benchmark.Finished:
	case benchmark.Stopped:

	}

	return nil, nil
}

func (s *stateMachine) configureFunc() inputFunc {
	return func(input string) error {
		return nil
	}
}
