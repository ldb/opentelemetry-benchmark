#!/usr/bin/env bash

sudo apt-get update
sudo apt-get -y install prometheus-node-exporter

sudo cat << EOF > /etc/systemd/system/benchd.service
[Unit]
Description=benchd Benchmarking Client
After=network.target

[Service]
ExecStart=/usr/local/bin/benchd
KillMode=mixed
Restart=on-failure
Type=simple

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
