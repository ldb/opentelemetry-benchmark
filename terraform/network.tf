data "google_compute_network" "benchmark_vpc" {
  name = "default"
}

# We make sure that only we can actually send requests to the benchmarking client.
# Traffic is not encrypted, but hey, at least its something :)
data "http" "own_ip" {
  url = "http://ipv4.icanhazip.com"
}

locals {
  own_ip = "${chomp(data.http.own_ip.body)}/32"
}

resource "google_compute_firewall" "default" {
  name    = "benchmarking-firewall"
  network = data.google_compute_network.benchmark_vpc.name

  allow {
    protocol = "tcp"
    ports    = ["7666", "9090"]
  }

  source_ranges = [local.own_ip]
}

