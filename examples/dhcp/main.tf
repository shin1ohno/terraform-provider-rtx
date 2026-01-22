terraform {
  required_providers {
    rtx = {
      source  = "sh1/rtx"
      version = "0.1.0"
    }
  }
}

provider "rtx" {
  host           = var.rtx_host
  username       = var.rtx_username
  password       = var.rtx_password
  admin_password = var.admin_password

  # テスト環境の場合
  skip_host_key_check = var.skip_host_key_check
}

# RTXルーターのシステム情報を取得
data "rtx_system_info" "router" {}

# ネットワークインターフェース情報を取得
# data "rtx_interfaces" "all" {}

# ルーティングテーブル情報を取得
# data "rtx_routes" "routing_table" {}

# DHCP静的割り当て設定
resource "rtx_dhcp_binding" "pro-0" {
  scope_id             = 1
  ip_address           = "192.168.1.20"
  mac_address          = "00:30:93:11:0e:33"
  use_mac_as_client_id = true
}

resource "rtx_dhcp_binding" "pro-1" {
  scope_id             = 1
  ip_address           = "192.168.1.21"
  mac_address          = "00:3e:e1:c3:54:b4"
  use_mac_as_client_id = true
}

resource "rtx_dhcp_binding" "pro-2" {
  scope_id             = 1
  ip_address           = "192.168.1.22"
  mac_address          = "00:3e:e1:c3:54:b5"
  use_mac_as_client_id = true
}

resource "rtx_dhcp_binding" "pro-3" {
  scope_id             = 1
  ip_address           = "192.168.1.23"
  mac_address          = "b6:1a:27:ea:28:29" # User's actual MAC from router config
  use_mac_as_client_id = true                # Use correct field name
}

resource "rtx_dhcp_binding" "server" {
  scope_id             = 1
  ip_address           = "192.168.1.51" # IP変更テスト
  mac_address          = "aa:bb:cc:dd:ee:ff"
  use_mac_as_client_id = true
}

# 出力
output "router_info" {
  value = {
    model            = data.rtx_system_info.router.model
    firmware_version = data.rtx_system_info.router.firmware_version
    uptime           = data.rtx_system_info.router.uptime
  }
}

# output "lan_interfaces" {
#   value = [
#     for intf in data.rtx_interfaces.all.interfaces : {
#       name    = intf.name
#       ip      = try(intf.ipv4_addresses[0], "N/A")
#       status  = intf.link_up ? "UP" : "DOWN"
#     } if intf.kind == "lan"
#   ]
# }
