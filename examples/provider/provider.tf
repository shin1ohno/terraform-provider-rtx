terraform {
  required_providers {
    rtx = {
      source = "sh1/rtx"
    }
  }
}

# Configure the RTX Provider
provider "rtx" {
  host     = "192.168.1.1"  # RTX router IP address
  username = "admin"
  password = var.rtx_password
  port     = 22  # SSH port (default: 22)
  timeout  = 30  # Connection timeout in seconds (default: 30)
}