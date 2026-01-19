# Static IPv6 prefix
# Configures a fixed IPv6 prefix for use in interface addressing
resource "rtx_ipv6_prefix" "static_prefix" {
  prefix_id     = 1
  prefix        = "2001:db8:1234::"
  prefix_length = 64
  source        = "static"
}

# RA-derived prefix
# Receives IPv6 prefix via Router Advertisement from upstream router
resource "rtx_ipv6_prefix" "ra_prefix" {
  prefix_id     = 2
  prefix_length = 64
  source        = "ra"
  interface     = "lan2"
}

# DHCPv6-PD prefix
# Receives IPv6 prefix via DHCPv6 Prefix Delegation from ISP
resource "rtx_ipv6_prefix" "dhcpv6_pd_prefix" {
  prefix_id     = 3
  prefix_length = 48
  source        = "dhcpv6-pd"
  interface     = "pp1"
}
