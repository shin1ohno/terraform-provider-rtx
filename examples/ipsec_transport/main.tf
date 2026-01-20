# IPsec transport mode configuration for RTX router
#
# IPsec transport is used for L2TP over IPsec configurations,
# mapping L2TP traffic (UDP 1701) to IPsec tunnels.

# Map L2TP traffic to IPsec tunnel 101
resource "rtx_ipsec_transport" "l2tp_tunnel1" {
  transport_id = 1
  tunnel_id    = 101
  protocol     = "udp"
  port         = 1701
}

# Map L2TP traffic to IPsec tunnel 3 (another remote site)
resource "rtx_ipsec_transport" "l2tp_tunnel2" {
  transport_id = 3
  tunnel_id    = 3
  protocol     = "udp"
  port         = 1701
}

# Example with explicit dependencies on IPsec tunnel
# resource "rtx_ipsec_transport" "l2tp_with_tunnel" {
#   transport_id = 2
#   tunnel_id    = rtx_ipsec_tunnel.remote_site.id
#   protocol     = "udp"
#   port         = 1701
# }
