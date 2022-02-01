# opentelemetry-benchmark
A benchmarking tool for the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

## About

This project contains all the code necessary to perform sophisticated maximum throughput benchmarks of the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

Some features:
- Automatic provisioning of testing infrastructure in Google Cloud using Terraform
- Ability to run pre-configured plans that describe the benchmark itself
- Easy to use CLI tool called `benchctl` to control and monitor the status of the benchmark
- Includes Prometheus based monitoring of all components, namely the collector, the benchmarking daemon and Prometheus itself
- Local Grafana Dashboard to easily monitor the components without having to analyze the log files

## Installation

The whole installation can be performed using the local *Makefile*. Run `make help` to get an overview of all the commands:

```shell
help                 Lists the available commands.
all                  Compiles binaries, provisions all infrastructure and starts Grafana Dashbaord.
compile              Compiles the binaries.
provision            Provisions all infrastructure in Google Cloud
dashboard            Deploy a local Grafana instance for easier monitoring
local                Spins up a local development environment.
clean                Remove build artifacts.
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

### Compilation

To compile you need to have the Go Programming Language environment installed. See [here](https://go.dev/doc/install) for installation instructions.

After that, simply run 
```shell
make compile
```

This will compile two binaries: 
- `benchd`, which is the actual benchmarking daemon running in Google Cloud
- `benchctl`, a command line tool to control one or more `benchd` instances at once

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

