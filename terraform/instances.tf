resource "google_compute_instance" "orchestrator" {
  name         = "test-orchestrator"
  machine_type = "e2-medium"
  zone         = "europe-west1-b"

  tags = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-9"
    }
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = <<EOT
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.41.0/otelcol_0.41.0_linux_amd64.deb
sudo dpkg -i otelcol_0.41.0_linux_amd64.deb
EOT
}