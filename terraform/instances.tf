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

  metadata_startup_script = file("${path.module}/scripts/init-collector.sh")
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
    network = data.google_compute_network.benchmark_vpc.name

    access_config {
      // Ephemeral public IP for SSH.
    }
  }

  metadata = {
    ssh-keys = "benchmark:${replace(tls_private_key.ssh_key.public_key_openssh, "\n", "")} benchmark"
  }


  metadata_startup_script = file("${path.module}/scripts/init-client.sh")
}

resource "google_compute_instance" "monitoring" {
  name         = "monitoring"
  machine_type = var.monitoring_machine_type
  zone         = "europe-west1-b"

  tags = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-9"
    }
  }

  network_interface {
    network = data.google_compute_network.benchmark_vpc.name

    access_config {
      // Ephemeral public IP for SSH.
    }
  }

  metadata = {
    ssh-keys = "benchmark:${replace(tls_private_key.ssh_key.public_key_openssh, "\n", "")} benchmark"
  }

  metadata_startup_script = templatefile("${path.module}/scripts/init-monitoring.sh.tmpl", {
    client-addresses = google_compute_instance.clients[*].network_interface.0.network_ip,
    collector-address = google_compute_instance.otel-collector.network_interface.0.network_ip,
  })
}

