package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ldb/openetelemtry-benchmark/benchmark"
	"github.com/ldb/openetelemtry-benchmark/command"
	"log"
	"os"
	"time"
)

var (
	configFlag = flag.String("config", "", "config file generated by terraform")
	planFlag   = flag.String("plan", "", "benchmarking plan to execute")
)

func main() {
	flag.Parse()
	if *configFlag == "" || *planFlag == "" {
		flag.PrintDefaults()
		return
	}

	plan, err := createPlan(*configFlag, *planFlag)
	if err != nil {
		log.Fatalf("error generating benchmarking plan: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("parsed the following plan:\n%+v\n", plan)
	fmt.Println("do you want to apply this plan? [Y/n]")
	scanner.Scan()
	t := scanner.Text()
	if t != "yes" && t != "y" && t != "" {
		fmt.Println("ok, aborting")
		return
	}
	fmt.Printf("applying plan %q\n", plan.Name)
	client := command.NewClient(plan.ClientAddress)
	status, err := client.CreateBenchmark(plan.Name)
	if err != nil {
		log.Fatalf("error creating benchmark: %v", err)
	}
	status, err = client.ConfigureBenchmark(plan.Name, plan.BenchConfig)
	if err != nil {
		log.Fatalf("error configuring benchmark: %v", err)
	}
	if status.State != benchmark.Configured.String() {
		log.Fatalf("benchmark not configured. current state: %+v", status)
	}
	fmt.Println("plan applied.")
	fmt.Println("do you want to start this plan? [Y/n]")
	scanner.Scan()
	t = scanner.Text()
	if t != "yes" && t != "y" && t != "" {
		fmt.Println("ok, aborting")
		return
	}
	status, err = client.StartBenchmark(plan.Name)
	if err != nil {
		log.Fatalf("error starting benchmark: %v", err)
	}
	if status.State != benchmark.Running.String() {
		log.Fatalf("benchmark not running. current state: %+v", status)
	}
	planDuration := time.After(plan.Duration.Duration)
outer:
	for {
		select {
		case <-planDuration:
			fmt.Println("plan finished.")
			break outer
		case <-time.Tick(time.Second):
			status, err = client.Status(plan.Name)
			if err != nil {
				log.Fatalf("error getting benchmark status: %v", err)
			}
			fmt.Printf("%+v\n", status)
		}
	}
	status, err = client.StopBenchmark(plan.Name)
	if err != nil {
		log.Fatalf("error stopping benchmark: %v", err)
	}
	if status.State != benchmark.Stopped.String() {
		log.Fatalf("benchmark not stopped. current state: %+v", status)
	}
	fmt.Printf("%+v\n", status)
}
