output "instance_public_ip" {
  description = "Public IP address of the new server"
  value       = oci_core_instance.free_server.public_ip
}