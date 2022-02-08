# opentelemetry-benchmark
A benchmarking tool for the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

## About

This project contains all the code necessary to perform sophisticated maximum throughput benchmarks of the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

Some features:
- Automatic provisioning of benchmarking infrastructure in Google Cloud using Terraform
- Ability to run pre-configured plans that describe the benchmark itself
- Easy to use CLI tool called `benchctl` to control and monitor the status of the benchmark
- Includes Prometheus based monitoring of all components, namely the collector, the benchmarking daemon and Prometheus itself
- Local Grafana Dashboard to easily monitor the components without having to analyze the log files

### Directory structure

The project is layed out as follows:

```shell
.
├── analysis   # Python scripts to generate figures
├── benchmark  # Package Benchmark contains the primary benchmarking abstraction
├── cmd        # This directory contains code to build various command line tools, namely `benchd`, `benchctl` and `promdl`. Read more below.
├── command    # Package command contains the implementation of the command protocl for `benchd` and `benchctl`
├── config     # Package config contains the structures for configuring `benchctl` and `benchd`
├── examples   # This directory contains example for local development
├── plans      # This directory contains benchmarking plans. Read more below.
├── results    # This directory contains the result of each benchmarking plan. Read more below.
├── terraform  # This directory contains all the code necessary to spin up the benchmarking environment in Google Cloud.
└── worker     # Package worker implements the main benchmarking component, a highly concurrent load generator.
```

## Installation

The whole installation can be performed using the local *Makefile*. Run `make help` to get an overview of all the commands:

```shell
help                 Lists the available commands.
all                  Compiles binaries, provisions all infrastructure and starts Grafana Dashbaord.
figures              Analyse all results and generates figures. WARNING: This is CPU intensive, but can be parallelised. Consider running on a machine with multiple cores.
clean                Remove build artifacts. This will NOT remove your result files of previous runs.
rebuild              Tear down and bring everything back up.
compile              Compiles the binaries.
provision            Provisions all infrastructure in Google Cloud
dashboard            Deploy a local Grafana instance for easier monitoring
local                Spins up a local development environment.
```

### Quickstart

The fastest way to run this project requires only two steps:
1. Make sure you have all dependencies installed. See *# Prerequisites* for the full list.
2. Run `make all`. This will automatically compile the code, provision the infrastructure and create a local Grafana Dashboard.  
For more information on these steps check out *# Compilation* and *# Provisioning*.

### Prerequisites

In order to make this tool as easy to use as possible, we tried minimizing external dependencies as much as possible.
However, a few things are still required. Namely:

- `terraform` (version 1.0 or higher) for bringing up the testing infrastructure in Google Cloud
- `gcloud`, the Google Cloud CLI to authenticate to Google Cloud
- `go` (version 1.17 or higher) for compiling the code.
- `docker` and `docker-compose` for a Grafana Dashbaord and local development.
- `python` version 3 as well as the various datascience libraries like `numpy`, `matplotlib` etc.

Please make sure you are already authenticated with Google Cloud and create a new project called `opentelemetry-benchmark`.

### Compilation

To compile you need to have the Go Programming Language environment installed. See [here](https://go.dev/doc/install) for installation instructions.

After that, simply run 
```shell
make compile
```

This will compile three binaries in the `./bin` directory: 
- `benchd`, which is the actual benchmarking daemon running in Google Cloud and is crosscompiled for linux/amd64
- `benchctl`, a command line tool to control one or more `benchd` instances at once
- `promdl`, a small tool that downloads the values of some predefined queries from the monitoring instance and saves the results as CSV files

### Provisioning

To provision the infrastructure simply run 
```shell
make provision
```

This runs Terraform and creates two files: `benchctl.config` and `privatekey.pem`.

**`benchctl.config`** contains a list of node names and their public IP addresses. 
This file is read by `benchctl`, the benchmarking controller.

**`privatekey.pem`** is a private SSH key for the `benchmark` user that can be used to SSH into the instances.
This should rarely be necessary, as the process is fully automated, but you never know.
To use it, use:
```shell
ssh -i privatekey.pem benchmark@{INSTANCE_IP}
```

Note that Terraform a simple firewall rule that grants networking access to all instances from the public IP of the current machine.
If you are using a VPN, or are in an environment that regularly rotates its public IP address, you may have trouble accessing the instances.

## Usage

All interactions with the Benchmarking Daemon `benchd` can be made over a simple HTTP based protocol. For convenience,
`benchctl` implements an easy to use client to communicate with `benchd`

### Overview
After compiling `benchctl` (see *# Compilation*), running it without any arguments will give you the following output: 
```shell
  -config string
        config file generated by terraform (default "benchctl.config")
  -plan string
        benchmarking plan to execute

```

`benchctl` requires two input files:
- A *config file*: This is creatd automatically by Terraform during Provisioning of the infrastructure. 
By default its called `benchctl.config`. It contains a list of IP addresses of the created instances.
- A *plan file*: This file contains a Benchmarking Plan, which is a definition of the different parameters for the benchmarking daemon.
A list of plan files can be found in `./plans`. **Important**: *Make sure to provision the Collector with the corresponding configuration before running a plan.
By default, it is provisioned with the `basic-1` configuration*

### Benchmark plans

Each Benchmark is described in a *plan file* that can be found under `./plans`. 
Note that a plan should only be run if the matching OpenTelemetry config (check the prefix) has been deployed.
To apply a different configuration, provide the respectie configuration name to the `sut_config_file` Terraform variable in `terraform/variables.tf`.

### promdl

`promdl` is a small tool that can be used to download relevant system and machine metrics from the instances for the time of a benchmark.  
It takes the following arguments:
```shell
  -end duration
        how much time to go back to end fetching fetch results from (default 30m0s)
  -host string
        Prometheus host URL
  -start duration
        how much time to go back to start fetching fetch results from (default 1h0m0s)
```
It downloads the results of a set of predefined queries between the timestamps of `[now-start, now-end]`.
Note that the granularity of results may vary, based on the size of the timeframe. For best comparability we recommend all benchmarks to have roughly the same duration.

### Running a Benchmark Plan an Generating Results

In order to run a benchmark plan and generate results, simply use `benchctl` like so for example:

```shell
./bin/benchctl -config benchctl.config -plan plans/basic-100.benchctl.yaml
```

This will:
- load the local `benchctl.config` file that was generated by Terraform
- load the local "basic-100" plan file

After applying and starting the benchmark, periodic updates are given.

At the end of the run you will be prompted to download the results file from a given link. You can do that, for example using `wget` or just a web browser.

Without renaming it, save the file under `results/<PLAN_NAME>/<RESULT_FILE>`. In the example above, that would be `results/basic-100/log-benchd-plan-basic-100-<SOME_RANDOM_NUMBERS>`.

To analyse the results, use `make figures`. Read more on that below under *# Generating Figures*.

## Study Design

We design the study around four basic categories of plans: 

- **basic** plans are meant to benchmark the raw performance of the OpenTelemetry collector without any features enabled. 
The workload is very synthetic and the plans are mostly differentiated by the scale of workers and the number of spans per trace and their depth.  
It is well suited for making an initial assessment on the infrastructure (e.g verify filedescriptor limits are high enough).
- **realistic** plans are meant to model more realistic workloads with fewer but longer spans that contain attributes.
- **mutate** plans are made for benchmarking a specific application scenario of the collector: Filtering out a specific "risky" attribute in incoming spans.
- **sampled** plans, similar to *filtered* ones model the behaviour of probabilistic sampling. In this scenario, a portion of all sent traces is expected to be sampled, the rest timing out on receiving.

The former two models are to answer the question of sensible deployment practices for an expected workload. They are of mere exploratorive nature.  
The latter two models iterate on these results and try to answer the question on how different features used in the collector affect its performance.  
We expect the mutating benchmark plans to perform worse (as additionatl computations and mutations take place), while the sampled workloads should perform better in sending (more sent traces), and similarly or slightly worse in receiving (some traces are sampled, so receiving will time out).

## Analysis

For our analysis we look at mainly the throupghput performance of the OpenTelemetry Collector.  
We want to study how the collector would perform when accepting highly distributed traces, for example by opening it up to the internet.

For each of the plans mentioned above we analyze:
- How the number of active clients and the overall send rate of all clients correlates
- What are the sending and receiving latencies given an overall sending rate
- When do errors occur
- How do different configurations compare in terms of throughput

### Generating Figures

Like the rest of this project, the creation of figures is fully automated.
Assuming you have a `results/` folder, with a log file for each plan in the according subdirectroy, simply run
```shell
make figures
```

All preprocessing is done in the Python scripts. In order to speed up computation, a local cache file will be generated under `results/analysis_cache` for heavy processing steps.  
Remove this file (for example by running `make clean`) to reprocess results (for example after making a change in the plan).

Nevertheless, generating the files is very compute intensive. I am not a good Python programmer and don't know all the efficient ways to do things there.
Sorry for the wasted cycles.

After everything was generated you should have 27 local figures, that each contain the name of the plan they were generated with and what they depict.

## Cleaning up

After each run, it is advisable to run `make clean`. This tears down the infrastructure completely and remove all local artifacts (except the already saved benchmarking results).
This guarantees a fresh environment for each plan.
