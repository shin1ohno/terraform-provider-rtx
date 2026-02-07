# Basic VLAN (802.1Q tagging only)
resource "rtx_vlan" "basic" {
  vlan_id   = 10
  interface = "lan1"
  name      = "Guest Network"
}

# VLAN with Layer 3 routing capability
resource "rtx_vlan" "management" {
  vlan_id    = 100
  interface  = "lan1"
  name       = "Management VLAN"
  ip_address = "192.168.100.1"
  ip_mask    = "255.255.255.0"
  shutdown   = false
}
