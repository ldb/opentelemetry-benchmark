variable "sut_machine_type" {
  type        = string
  description = "Instance Type for System under Test."
  default     = "e2-standard-2"
}

variable "sut_config_file" {
  type        = string
  description = "The config file template to provision the system under test with."
  default     = "../plans/basic-1.otel.yaml.tmpl"
}

variable "monitoring_machine_type" {
  type        = string
  description = "Instance Type for monitoring instance."
  default     = "e2-medium"
}

variable "client_machine_type" {
  type        = string
  description = "Instance Type for benchmarking clients."
  default     = "e2-standard-4"
}

variable "number_clients" {
  type        = number
  description = "Number of benchmarking clients to launch. At the moment, only one is supported in code."
  default     = 1
}