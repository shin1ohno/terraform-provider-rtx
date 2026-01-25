# Static IPv6 filter examples for RTX router

# Allow ICMPv6 (required for IPv6 to function properly)
resource "rtx_access_list_ipv6" "allow_icmp6" {
  sequence    = 101000
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "icmp6"
  source_port = "*"
  dest_port   = "*"
}

# Allow DHCPv6 client (UDP 546)
resource "rtx_access_list_ipv6" "allow_dhcpv6_client" {
  sequence    = 101001
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "udp"
  source_port = "*"
  dest_port   = "546"
}

# Allow DHCPv6 server (UDP 547)
resource "rtx_access_list_ipv6" "allow_dhcpv6_server" {
  sequence    = 101002
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "udp"
  source_port = "*"
  dest_port   = "547"
}

# Block incoming SSH except from specific prefix
resource "rtx_access_list_ipv6" "allow_ssh_from_trusted" {
  sequence    = 101010
  action      = "pass"
  source      = "2001:db8::/32"
  destination = "*"
  protocol    = "tcp"
  source_port = "*"
  dest_port   = "22"
}

# Default pass rule
resource "rtx_access_list_ipv6" "default_pass" {
  sequence    = 101099
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "*"
  source_port = "*"
  dest_port   = "*"
}
