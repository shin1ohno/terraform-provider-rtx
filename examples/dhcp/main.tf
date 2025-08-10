terraform {
  required_providers {
    rtx = {
      source  = "sh1/rtx"
      version = "0.1.0"
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

# ネットワークインターフェース情報を取得
# data "rtx_interfaces" "all" {}

# ルーティングテーブル情報を取得
# data "rtx_routes" "routing_table" {}

# DHCP静的割り当て設定
resource "rtx_dhcp_binding" "printer" {
  scope_id    = 1
  ip_address  = "192.168.1.100"
  mac_address = "00:11:22:33:44:55"
}

resource "rtx_dhcp_binding" "server" {
  scope_id              = 1
  ip_address            = "192.168.1.50"
  mac_address           = "aa:bb:cc:dd:ee:ff"
  use_client_identifier = true
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
