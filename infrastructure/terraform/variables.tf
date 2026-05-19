variable "yc_cloud_id" {
  type = string
}

variable "yc_folder_id" {
  type = string
}

variable "yc_zone" {
  type = string
}

variable "subnet_cidr" {
  type = list(string)
}

variable "instance_name" {
  type = string
}

variable "instance_cores" {
  type = number
}

variable "instance_memory" {
  type = number
}

variable "boot_disk_size" {
  type = number
}

variable "instance_platform_id" {
  type = string
}

variable "allowed_ports" {
  type = list(number)
}

variable "allowed_udp_ports" {
  type    = list(number)
  default = []
}

variable "udp_port_ranges" {
  type = list(object({
    from = number
    to   = number
  }))
  default = []
}

variable "ssh_public_key_path" {
  type = string
}
