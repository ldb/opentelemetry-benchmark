resource "google_compute_network" "benchmark_vpc" {
  name = "benchmark-vpc"
}

resource "google_compute_firewall" "default" {
  name    = "benchmarking-firewall"
  network = google_compute_network.benchmark_vpc.name

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_tags = ["ssh"]
}

