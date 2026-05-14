output "external_ip" {
  value = yandex_compute_instance.node.network_interface[0].nat_ip_address
}

output "internal_ip" {
  value = yandex_compute_instance.node.network_interface[0].ip_address
}

output "ssh_command" {
  value = "ssh ubuntu@${yandex_compute_instance.node.network_interface[0].nat_ip_address}"
}
