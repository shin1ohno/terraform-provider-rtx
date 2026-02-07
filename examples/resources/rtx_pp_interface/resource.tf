# PP interface with dynamic IP from ISP (PPPoE)
resource "rtx_pp_interface" "wan" {
  pp_number      = 1
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1000
}

# PP interface with static IP assignment
resource "rtx_pp_interface" "static" {
  pp_number  = 2
  ip_address = "203.0.113.1/30"
  mtu        = 1500
}
