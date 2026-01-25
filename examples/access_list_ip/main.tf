# Static IP filter examples for RTX router

# Block private network ranges from external interface
resource "rtx_access_list_ip" "block_rfc1918_10" {
  sequence   = 200000
  action      = "reject"
  source      = "10.0.0.0/8"
  destination = "*"
  protocol    = "*"
}

resource "rtx_access_list_ip" "block_rfc1918_172" {
  sequence   = 200001
  action      = "reject"
  source      = "172.16.0.0/12"
  destination = "*"
  protocol    = "*"
}

resource "rtx_access_list_ip" "block_rfc1918_192" {
  sequence   = 200002
  action      = "reject"
  source      = "192.168.0.0/16"
  destination = "*"
  protocol    = "*"
}

# Block NetBIOS ports
resource "rtx_access_list_ip" "block_netbios" {
  sequence   = 200020
  action      = "reject"
  source      = "*"
  destination = "*"
  protocol    = "udp"
  source_port = "135"
  dest_port   = "*"
}

# Allow IPsec traffic
resource "rtx_access_list_ip" "allow_ipsec_500" {
  sequence   = 200100
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "udp"
  source_port = "*"
  dest_port   = "500"
}

resource "rtx_access_list_ip" "allow_ipsec_4500" {
  sequence   = 200101
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "udp"
  source_port = "*"
  dest_port   = "4500"
}

# Allow established TCP connections
resource "rtx_access_list_ip" "allow_established" {
  sequence   = 200098
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "tcp"
  established = true
}

# Default pass rule
resource "rtx_access_list_ip" "default_pass" {
  sequence   = 200099
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "*"
}
