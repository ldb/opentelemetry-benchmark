# opentelemetry-benchmark
A benchmarking tool for the OpenTelemetry Collector.

## About

## Installation

The fastest way to run this project requires only two steps:
1. Make sure you have all dependencies installed. See *# Prerequisites* for the full list.
2. Run `make all`. This will automatically compile the code, provision the infrastructure and generate config files.  
For more information on these steps check out *# Compilation* *# Provisioning* and *# Generating Configs*.

### Prerequisites

In order to make this tool as easy to use as possible, we tried minimizing external dependencies as much as possible.
However, a few things are still required. Namely:

- `terraform` (version 1.0 or higher) for bringing up the testing infrastructure in Google Cloud
- `gcloud`, the Google Cloud CLI to authenticate to Google Cloud
- `go` (version 1.17 or higher) for compiling the code.
- (optionally )`docker` and `docker-compose` for local development. Can also be used to compile the code if the `go` dependency can not be met (read below in *# Compilation*).

### Compilation

To compile you need to have the Go Programming Language environment installed. See [here](https://go.dev/doc/install) for installation instructions.

After that, simply run 
```shell
make compile
```

This will compile two binaries: 
- `benchd`, which is the actual benchmarking client running in Google Cloud
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

### Generating Configs

In order to generate configuration files for the benchmarks, run:
```shell
make generate
```

This will generate a set of benchmarking plans that can be played using `benchctl` and applied using `make deploy`.

## Usage
