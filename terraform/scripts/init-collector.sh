#!/usr/bin/env bash

sudo apt-get update
sudo apt-get install -y wget prometheus-node-exporter
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.41.0/otelcol_0.41.0_linux_amd64.deb
sudo dpkg -i otelcol_0.41.0_linux_amd64.deb

sudo cat <<EOF > /etc/otelcol/config.yaml
${config_file}
EOF
sudo sysctl -w fs.file-max=900000000
sudo mkdir -p /lib/systemd/system/otelcol.service.d/
sudo cat <<EOF > /lib/systemd/system/otelcol.service.d/override.conf
[Service]
LimitNOFILE=900000000
EOF
sudo systemctl daemon-reload
sudo systemctl restart otelcol
