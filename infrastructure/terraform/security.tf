resource "yandex_vpc_security_group" "main_sg" {
  name       = "main-security-group"
  network_id = yandex_vpc_network.main.id

  dynamic "ingress" {
    for_each = var.allowed_ports

    content {
      protocol       = "TCP"
      port           = ingress.value
      v4_cidr_blocks = ["0.0.0.0/0"]
    }
  }

  egress {
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
