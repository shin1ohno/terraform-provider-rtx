# NTT FLET'S NGN PPPoE connection
resource "rtx_pppoe" "flets" {
  pp_number      = 1
  name           = "NTT FLET'S NGN"
  bind_interface = "lan2"
  username       = "user@example.ne.jp"
  password       = "example!PASS123"
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}

# PP interface IP configuration for the PPPoE connection
resource "rtx_pp_interface" "flets" {
  pp_number      = rtx_pppoe.flets.pp_number
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1000

  depends_on = [rtx_pppoe.flets]
}
