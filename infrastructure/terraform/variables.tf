variable "yc_token" {
  description = "Yandex Cloud OAuth token"
  type        = string
  sensitive   = true
}

variable "yc_cloud_id" {
  description = "Yandex Cloud ID"
  type        = string
}

variable "yc_folder_id" {
  description = "Yandex Cloud Folder ID"
  type        = string
}

variable "yc_zone" {
  description = "Yandex Cloud Zone"
  type        = string
}

variable "ssh_public_key" {
  description = "SSH public key for instance access"
  type        = string
}

variable "instance_name" {
  description = "Name of the compute instance"
  type        = string
}

variable "instance_cores" {
  description = "Number of CPU cores"
  type        = number
}

variable "instance_memory" {
  description = "Memory in GB"
  type        = number
}

variable "boot_disk_size" {
  description = "Boot disk size in GB"
  type        = number
}

variable "os_image_id" {
  description = "Image ID for the operating system"
  type        = string
}

variable "instance_platform_id" {
  description = "Yandex Cloud compute platform ID"
  type        = string
}

variable "subnet_cidr" {
  description = "CIDR block for the subnet"
  type        = list(string)
}
