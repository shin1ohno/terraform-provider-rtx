# One-to-one static NAT mapping
resource "rtx_nat_static" "basic" {
  descriptor_id = 1

  entry {
    inside_local   = "192.168.1.100"
    outside_global = "203.0.113.100"
  }
}

# Static NAT with port-based mapping
resource "rtx_nat_static" "port_mapping" {
  descriptor_id = 2

  # HTTP to web server
  entry {
    inside_local        = "192.168.1.20"
    inside_local_port   = 80
    outside_global      = "203.0.113.1"
    outside_global_port = 80
    protocol            = "tcp"
  }

  # HTTPS to web server
  entry {
    inside_local        = "192.168.1.20"
    inside_local_port   = 443
    outside_global      = "203.0.113.1"
    outside_global_port = 443
    protocol            = "tcp"
  }
}
