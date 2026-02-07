# DNS server with forwarding and static host entries
resource "rtx_dns_server" "main" {
  domain_lookup = true
  domain_name   = "example.com"
  name_servers  = ["8.8.8.8", "1.1.1.1"]

  # Domain-based DNS server selection
  server_select {
    priority      = 1
    query_pattern = "internal.example.com"
    server {
      address = "192.168.1.1"
    }
  }

  server_select {
    priority      = 10
    query_pattern = "."
    record_type   = "any"
    server {
      address = "10.0.0.53"
      edns    = true
    }
  }

  # Static DNS host entries
  hosts {
    name    = "router"
    address = "192.168.1.1"
  }

  hosts {
    name    = "nas"
    address = "192.168.1.10"
  }

  service_on            = true
  private_address_spoof = true
}
