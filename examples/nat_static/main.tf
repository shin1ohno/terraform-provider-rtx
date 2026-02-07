# RTX Static NAT Configuration Examples

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.10"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Basic static NAT - one-to-one address mapping
# Maps a single internal IP to a single external IP
resource "rtx_nat_static" "basic" {
  descriptor_id = 1

  entry {
    inside_local   = "192.168.1.100"
    outside_global = "203.0.113.100"
  }
}

# Static NAT with multiple entries
# Useful for servers that need public IP addresses
resource "rtx_nat_static" "multiple_servers" {
  descriptor_id = 2

  # Web server
  entry {
    inside_local   = "192.168.1.10"
    outside_global = "203.0.113.10"
  }

  # Mail server
  entry {
    inside_local   = "192.168.1.11"
    outside_global = "203.0.113.11"
  }

  # Database server
  entry {
    inside_local   = "192.168.1.12"
    outside_global = "203.0.113.12"
  }
}

# Static NAT with port mapping (port-based static NAT)
# Maps specific ports on external IP to internal servers
resource "rtx_nat_static" "port_mapping" {
  descriptor_id = 3

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

  # SMTP to mail server
  entry {
    inside_local        = "192.168.1.21"
    inside_local_port   = 25
    outside_global      = "203.0.113.1"
    outside_global_port = 25
    protocol            = "tcp"
  }

  # DNS to internal DNS server
  entry {
    inside_local        = "192.168.1.22"
    inside_local_port   = 53
    outside_global      = "203.0.113.1"
    outside_global_port = 53
    protocol            = "udp"
  }
}

variable "rtx_host" {
  description = "RTX router hostname or IP address"
  type        = string
}

variable "rtx_username" {
  description = "RTX router username"
  type        = string
}

variable "rtx_password" {
  description = "RTX router password"
  type        = string
  sensitive   = true
}
