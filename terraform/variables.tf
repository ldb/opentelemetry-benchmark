variable "sut_machine_type" {
  type = string
  description = "Instance Type for System under Test."
  default = "e2-medium"
}

variable "client_machine_type" {
  type = string
  description = "Instance Type for benchmarking clients."
  default = "e2-medium"
}

variable "number_clients" {
  type = number
  description = "Number of benchmarking clients to launch."
  default = 1
}