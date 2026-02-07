# Basic NAT masquerade for PPPoE connection
resource "rtx_nat_masquerade" "pppoe" {
  descriptor_id = 1
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"
}

# NAT masquerade with port forwarding
resource "rtx_nat_masquerade" "with_port_forwarding" {
  descriptor_id = 2
  outer_address = "pp1"
  inner_network = "192.168.2.0-192.168.2.255"

  # Forward HTTP to internal web server
  static_entry {
    entry_number        = 1
    inside_local        = "192.168.2.10"
    inside_local_port   = 80
    outside_global_port = 80
    protocol            = "tcp"
  }

  # Forward SSH on non-standard external port
  static_entry {
    entry_number        = 2
    inside_local        = "192.168.2.20"
    inside_local_port   = 22
    outside_global_port = 2222
    protocol            = "tcp"
  }
}
