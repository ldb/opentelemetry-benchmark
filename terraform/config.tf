# Generates a local configuration file for the benchmarking orchestrator to use.
resource "local_file" "config" {
  filename = "../benchctl.config"
  content  = <<EOT
target ${google_compute_instance.otel-collector.network_interface.0.network_ip}
monitoring ${google_compute_instance.monitoring.network_interface.0.access_config.0.nat_ip}
%{for ip in google_compute_instance.clients.*.network_interface.0.access_config.0.nat_ip~}
client ${ip}
%{endfor~}
EOT
}

# Generate the datasource file for a local Grafana instance.
resource "local_file" "grafana-datasource" {
  filename = "./grafana/datasource/datasource.yaml"
  content  = <<EOT
apiVersion: 1
deleteDatasources:
  - name: Prometheus
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    orgId: 1
    url: http://${google_compute_instance.monitoring.network_interface.0.access_config.0.nat_ip}:9090
    basicAuth: false
    isDefault: true
    version: 1
    editable: false
    apiVersion: 1
    uid: prom
EOT
}