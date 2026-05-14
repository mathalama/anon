resource "yandex_compute_instance" "node" {
  name        = var.instance_name
  platform_id = var.instance_platform_id

  resources {
    cores         = var.instance_cores
    memory        = var.instance_memory
    core_fraction = 100
  }

  scheduling_policy {
    preemptible = false
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
      size     = var.boot_disk_size
      type     = "network-ssd"
    }
  }

  network_interface {
    subnet_id          = yandex_vpc_subnet.main.id
    nat                = true
    security_group_ids = [yandex_vpc_security_group.main_sg.id]
  }

  metadata = {
    serial-port-enable = 1
    ssh-keys           = "ubuntu:${file(pathexpand(var.ssh_public_key_path))}"
  }
}
