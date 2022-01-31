# Generates a local configuration file for the benchmarking orchestrator to use.
resource "local_file" "config" {
  filename = "../benchctl.config"
  content     = <<EOT
target ${google_compute_instance.otel-collector.network_interface.0.access_config.0.nat_ip}
monitoring ${google_compute_instance.monitoring.network_interface.0.access_config.0.nat_ip}
%{ for ip in google_compute_instance.clients.*.network_interface.0.access_config.0.nat_ip ~}
client ${ip}
%{ endfor ~}
EOT
}