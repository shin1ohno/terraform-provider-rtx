# Requirements: rtx_dns_server

## Overview
Terraform resource for managing DNS server configuration on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_system` (DNS settings) / No direct equivalent for DNS server

## Cisco Compatibility

This resource follows general Cisco naming patterns where applicable:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name_servers` | `ip_name_server` | Upstream DNS servers list |
| `domain_name` | `ip_domain_name` | Default domain name |
| `domain_lookup` | `ip_domain_lookup` | Enable DNS lookup |
| `hosts` | `ip_host` | Static host entries |

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

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned DNS configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete DNS server settings |
| `terraform destroy` | ✅ Required | Remove DNS configuration and restore defaults |
| `terraform import` | ✅ Required | Import existing DNS configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual DNS configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `dns` (singleton resource)
- **Import Command**: `terraform import rtx_dns_server.main dns`
- **Post-Import**: All DNS settings must be populated from router

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
# DNS server configuration
resource "rtx_dns_server" "main" {
  domain_lookup = true
  domain_name   = "example.local"

  name_servers = ["8.8.8.8", "8.8.4.4"]

  # Domain-specific DNS routing
  server_select = [
    {
      domain  = "internal.example.com"
      servers = ["192.168.1.53"]
    }
  ]

  # Static host entries
  hosts = [
    {
      name    = "router.local"
      address = "192.168.1.1"
    }
  ]
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
