# LAN interface with Router Advertisement and DHCPv6 server
resource "rtx_ipv6_interface" "lan" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  rtadv {
    enabled   = true
    prefix_id = 1
    o_flag    = true
    m_flag    = false
    lifetime  = 1800
  }

  dhcpv6_service = "server"
  mtu            = 1500
}

# WAN interface with DHCPv6 client for prefix delegation
resource "rtx_ipv6_interface" "wan" {
  interface = "lan2"

  dhcpv6_service = "client"

  rtadv {
    enabled   = false
    prefix_id = 2
  }
}
