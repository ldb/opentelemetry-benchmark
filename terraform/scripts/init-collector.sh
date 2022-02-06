#!/usr/bin/env bash

sudo apt-get update
sudo apt-get install -y wget prometheus-node-exporter
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.43.0/otelcol-contrib_0.43.0_linux_amd64.deb
sudo dpkg -i otelcol-contrib_0.43.0_linux_amd64.deb

sudo cat <<EOF > /etc/otelcol-contrib/config.yaml
${config_file}
EOF
sudo sysctl -w fs.file-max=900000000
sudo mkdir -p /lib/systemd/system/otelcol-contrib.service.d/
sudo cat <<EOF > /lib/systemd/system/otelcol-contrib.service.d/override.conf
[Service]
LimitNOFILE=900000000
EOF
sudo systemctl daemon-reload
sudo systemctl restart otelcol-contrib
