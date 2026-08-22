# LAN interface with Router Advertisement and DHCPv6 server
resource "rtx_ipv6_interface" "lan" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  rtadv {
    enabled    = true
    prefix_ids = [1]
    o_flag     = true
    m_flag     = false
    lifetime   = 1800
  }

  dhcpv6_service = "server"
  mtu            = 1500
}

# WAN interface with DHCPv6 client for prefix delegation
resource "rtx_ipv6_interface" "wan" {
  interface = "lan2"

  dhcpv6_service = "client"

  rtadv {
    enabled    = false
    prefix_ids = [2]
  }
}

# LAN interface advertising two prefixes at once:
# a delegated global prefix plus a stable ULA prefix.
# The router advertises them in the order given.
resource "rtx_ipv6_interface" "lan_dual_prefix" {
  interface = "lan3"

  rtadv {
    enabled    = true
    prefix_ids = [1, 3]
    o_flag     = true
    m_flag     = false
  }
}
