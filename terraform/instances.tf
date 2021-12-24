resource "tls_private_key" "ssh_key" {
  algorithm   = "RSA"
}

resource "local_file" "ssh_key" {
  filename = "../privatekey.pem"
  sensitive_content     = tls_private_key.ssh_key.private_key_pem
  file_permission = "400"
}

resource "google_compute_instance" "otel-collector" {
  name         = "otel-collector"
  machine_type = var.sut_machine_type
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
      // Ephemeral public IP for SSH.
    }
  }

  metadata = {
    ssh-keys = "benchmark:${replace(tls_private_key.ssh_key.public_key_openssh, "\n", "")} benchmark"
  }

  metadata_startup_script = <<EOT
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.41.0/otelcol_0.41.0_linux_amd64.deb
sudo dpkg -i otelcol_0.41.0_linux_amd64.deb
EOT
}

resource "google_compute_instance" "clients" {
  count = var.number_clients

  name         = "benchmarking-client-${count.index}"
  machine_type = var.client_machine_type
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
      // Ephemeral public IP for SSH.
    }
  }

  metadata = {
    ssh-keys = "benchmark:${replace(tls_private_key.ssh_key.public_key_openssh, "\n", "")} benchmark"
  }


  metadata_startup_script = <<EOT
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.41.0/otelcol_0.41.0_linux_amd64.deb
sudo dpkg -i otelcol_0.41.0_linux_amd64.deb
EOT
}

