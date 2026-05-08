terraform {
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = ">= 0.80.0"
    }
  }
}

provider "yandex" {
  token     = var.yc_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_zone
}

resource "yandex_vpc_network" "sumdyk_net" {
  name = "sumdyk-network"
}

resource "yandex_vpc_subnet" "sumdyk_subnet" {
  name           = "sumdyk-subnet"
  zone           = var.yc_zone
  network_id     = yandex_vpc_network.sumdyk_net.id
  v4_cidr_blocks = var.subnet_cidr
}

resource "yandex_compute_instance" "sumdyk_node" {
  name        = var.instance_name
  platform_id = var.instance_platform_id

  resources {
    cores  = var.instance_cores
    memory = var.instance_memory
  }

  boot_disk {
    initialize_params {
      image_id = var.os_image_id
      size     = var.boot_disk_size
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.sumdyk_subnet.id
    nat       = true
  }

  metadata = {
    ssh-keys = "ubuntu:${var.ssh_public_key}"
  }
}

output "instance_external_ip" {
  value = yandex_compute_instance.sumdyk_node.network_interface.0.nat_ip_address
}