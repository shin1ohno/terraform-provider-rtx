# Map L2TP traffic (UDP 1701) to IPsec tunnel
resource "rtx_ipsec_transport" "l2tp" {
  transport_id = 1
  tunnel_id    = 101
  protocol     = "udp"
  port         = 1701
}
