terraform {
  required_version = ">= 1.5.0"

  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "~> 0.202"
    }
  }
}

provider "yandex" {
  endpoint = "api.yandexcloud.kz:443"

  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_zone
}
