resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
}

resource "local_file" "ssh_key" {
  filename          = "../privatekey.pem"
  sensitive_content = tls_private_key.ssh_key.private_key_pem
  file_permission   = "400"
}

resource "google_compute_instance" "otel-collector" {
  name         = "otel-collector"
  machine_type = var.sut_machine_type
  zone         = "europe-west1-b"

  tags = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
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

  metadata_startup_script = templatefile("${path.module}/scripts/init-collector.sh", {
    config_file = templatefile("${path.module}/${var.sut_config_file}", {
      client_ip = google_compute_instance.clients.0.network_interface.0.network_ip
    })
  })

  connection {
    type        = "ssh"
    user        = "benchmark"
    host        = self.network_interface.0.access_config.0.nat_ip
    port        = 22
    private_key = tls_private_key.ssh_key.private_key_pem
  }
}

resource "google_compute_instance" "clients" {
  count = var.number_clients

  name         = "benchmarking-client-${count.index}"
  machine_type = var.client_machine_type
  zone         = "europe-west1-b"

  tags = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
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

  # Provision instance with `benchd` binary
  provisioner "file" {
    source      = "${path.module}/../bin/benchd"
    destination = "benchd"
  }
  provisioner "remote-exec" {
    on_failure = continue
    inline = [
      #"sudo base64 --decode benchdb64 > benchd",
      "sudo chmod +x benchd",
      "sudo cp benchd /usr/local/bin/benchd",
      "sleep 1",
      "sudo systemctl start benchd",
    ]
  }

  connection {
    type        = "ssh"
    user        = "benchmark"
    host        = self.network_interface.0.access_config.0.nat_ip
    port        = 22
    private_key = tls_private_key.ssh_key.private_key_pem
  }
}

resource "google_service_account" "prometheus-sd-user" {
  account_id   = "prometheus-sd-user"
  display_name = "Prometheus Service Discovery"
}

resource "google_project_iam_member" "prometheus-sd-user-compute-viewer" {
  role    = "roles/compute.viewer"
  member  = "serviceAccount:${google_service_account.prometheus-sd-user.email}"
  project = "opentelemetry-benchmark"
}

resource "google_compute_instance" "monitoring" {
  name         = "monitoring"
  machine_type = var.monitoring_machine_type
  zone         = "europe-west1-b"

  tags = ["ssh"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
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
    client-addresses  = google_compute_instance.clients[*].network_interface.0.network_ip,
    collector-address = google_compute_instance.otel-collector.network_interface.0.network_ip,
    zone              = "europe-west1-b"
    project           = "opentelemetry-benchmark"
    replacement       = "$${1}:9100"
  })

  service_account {
    # Google recommends custom service accounts that have cloud-platform scope and permissions granted via IAM Roles.
    email  = google_service_account.prometheus-sd-user.email
    scopes = ["cloud-platform", "compute-ro"]
  }
}
