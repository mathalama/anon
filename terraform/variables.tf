variable "tenancy_ocid" {
  description = "OCI tenancy OCID"
  type        = string
  sensitive   = true
}

variable "user_ocid" {
  description = "OCI user OCID"
  type        = string
  sensitive   = true
}

variable "fingerprint" {
  description = "Fingerprint for the OCI API key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "OCI region"
  type        = string
}

variable "compartment_id" {
  description = "OCI compartment OCID"
  type        = string
  sensitive   = true
}

variable "subnet_id" {
  description = "OCI subnet OCID"
  type        = string
  sensitive   = true
}

variable "private_key_path" {
  description = "Path to the OCI API private key"
  type        = string
  sensitive   = true
}

variable "ssh_public_key" {
  description = "SSH public key for instance access"
  type        = string
  sensitive   = true
}

variable "open_ports" {
  description = "List of TCP ports to expose publicly"
  type        = list(number)
}

variable "instance_shape" {
  description = "OCI instance shape"
  type        = string
}

variable "instance_display_name" {
  description = "Display name for the OCI instance"
  type        = string
}

variable "instance_ocpus" {
  description = "Number of OCPUs for the instance (for Flexible shapes)"
  type        = number
  default     = 1
}

variable "instance_memory_gbs" {
  description = "Amount of memory in GBs for the instance (for Flexible shapes)"
  type        = number
  default     = 6
}

variable "os_name" {
  description = "Operating system name"
  type        = string
}

variable "os_version" {
  description = "Operating system version"
  type        = string
}

variable "ingress_protocol" {
  description = "The protocol for ingress security rules (e.g., '6' for TCP)"
  default     = "6"
}

variable "ingress_source" {
  description = "The source CIDR for ingress security rules"
  type        = string
}

variable "assign_public_ip" {
  description = "Whether to assign a public IP to the instance"
  type        = bool
  default     = true
}

variable "image_source_type" {
  description = "The source type for the instance image"
  type        = string
  default     = "image"
}

variable "image_sort_by" {
  description = "The criteria to sort images by"
  type        = string
  default     = "TIMECREATED"
}

variable "image_sort_order" {
  description = "The order to sort images"
  type        = string
  default     = "DESC"
}
