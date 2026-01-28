terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.8"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password

  # テスト環境の場合
  skip_host_key_check = var.skip_host_key_check
}

# RTXルーターのシステム情報を取得
data "rtx_system_info" "router" {}

# 出力
output "router_info" {
  value = {
    model            = data.rtx_system_info.router.model
    firmware_version = data.rtx_system_info.router.firmware_version
    uptime           = data.rtx_system_info.router.uptime
  }
}