#!/usr/bin/env bash

wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.41.0/otelcol_0.41.0_linux_amd64.deb
sudo dpkg -i otelcol_0.41.0_linux_amd64.deb
sudo apt-get update
sudo apt-get -y install prometheus-node-exporter
