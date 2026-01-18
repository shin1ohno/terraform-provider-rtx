# Requirements: rtx_dns_server

## Overview
Terraform resource for managing DNS server configuration on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure DNS server/resolver settings
- **Read**: Query current DNS configuration
- **Update**: Modify DNS settings
- **Delete**: Remove DNS configuration

### 2. DNS Server Mode
- Enable/disable DNS server function
- Configure as DNS proxy/forwarder
- Configure as authoritative DNS server

### 3. Upstream DNS Servers
- Primary and secondary DNS server addresses
- Per-interface DNS server settings
- DNS server selection based on domain

### 4. DNS Proxy Settings
- Enable DNS proxy function
- Cache configuration (size, TTL)
- Query forwarding rules

### 5. Static Host Entries
- Define static hostname to IP mappings
- Support multiple IPs per hostname
- TTL configuration

### 6. Domain-based Routing
- Route DNS queries by domain to specific servers
- Support for internal domain resolution
- Split-horizon DNS

### 7. Import Support
- Import existing DNS configuration

## Non-Functional Requirements

### 8. Validation
- Validate DNS server IP addresses
- Validate domain name format
- Validate TTL ranges

## RTX Commands Reference
```
dns server <ip1> [<ip2>]
dns server pp <n>
dns server select <id> <server> <domain>...
dns private address spoof <switch>
dns static <hostname> <ip>
dns service <switch>
```

## Example Usage
```hcl
resource "rtx_dns_server" "main" {
  enabled = true

  upstream_servers = ["8.8.8.8", "8.8.4.4"]

  domain_routing {
    domain  = "internal.example.com"
    servers = ["192.168.1.53"]
  }

  static_entry {
    hostname = "router.local"
    address  = "192.168.1.1"
  }
}
```
