# Basic DHCP scope
resource "rtx_dhcp_scope" "basic" {
  scope_id = 1
  network  = "192.168.1.0/24"
}

# Full-featured DHCP scope with gateway, DNS, and exclusions
resource "rtx_dhcp_scope" "full" {
  scope_id = 2
  network  = "192.168.2.0/24"
  gateway  = "192.168.2.1"

  dns_servers = ["8.8.8.8", "8.8.4.4"]
  lease_time  = "24h"

  exclude_ranges {
    start = "192.168.2.1"
    end   = "192.168.2.10"
  }

  exclude_ranges {
    start = "192.168.2.250"
    end   = "192.168.2.254"
  }
}
