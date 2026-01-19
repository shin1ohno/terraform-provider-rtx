# WAN interface with DHCP and security filters
resource "rtx_interface" "wan" {
  name        = "lan2"
  description = "WAN connection"

  ip_address {
    dhcp = true
  }

  secure_filter_in  = [200020, 200021, 200099]
  secure_filter_out = [200020, 200021, 200099]
  nat_descriptor    = 1000
}

# LAN interface with static IP
resource "rtx_interface" "lan" {
  name        = "lan1"
  description = "Internal LAN"

  ip_address {
    address = "192.168.1.1/24"
  }

  proxyarp = true
}
