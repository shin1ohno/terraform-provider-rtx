# Basic bridge with LAN interfaces
resource "rtx_bridge" "internal" {
  name    = "bridge1"
  members = ["lan1", "lan2"]
}

# L2VPN bridge combining LAN and L2TPv3 tunnel interfaces
resource "rtx_bridge" "l2vpn" {
  name    = "bridge2"
  members = ["lan3", "tunnel1", "tunnel2"]
}
