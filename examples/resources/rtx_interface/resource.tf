# WAN interface with DHCP and access list bindings
resource "rtx_interface" "wan" {
  name        = "lan2"
  description = "WAN connection"

  ip_address {
    dhcp = true
  }

  # Access list bindings (reference rtx_access_list_ip resources by name)
  access_list_ip_in  = "wan-secure-in"
  access_list_ip_out = "wan-secure-out"
  nat_descriptor     = 1000
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
