# Master Requirements: DNS Resources

## Overview

This document defines the requirements for DNS-related resources in the Terraform provider for Yamaha RTX routers. These resources manage DNS server configuration, NetVolante DNS (Yamaha's free DDNS service), and custom DDNS provider integration. DNS functionality is critical for network operations, enabling hostname resolution for internal and external resources.

## Alignment with Product Vision

These DNS resources support the core product vision of providing Infrastructure as Code for Yamaha RTX routers:

1. **Network Management**: DNS configuration is fundamental to network operations
2. **Dynamic DNS**: DDNS resources enable remote access to routers with dynamic IP addresses
3. **Enterprise Features**: NetVolante DNS provides zero-cost dynamic DNS for Yamaha users
4. **Multi-provider Support**: Custom DDNS allows integration with any DDNS provider

## Resource Summary

| Resource | Type | Import Support | Description |
|----------|------|----------------|-------------|
| `rtx_dns_server` | singleton | Yes (ID: "dns") | Router-wide DNS server configuration |
| `rtx_netvolante_dns` | collection | Yes (ID: interface) | NetVolante DNS (Yamaha DDNS) per interface |
| `rtx_ddns` | collection | Yes (ID: server_id 1-4) | Custom DDNS provider configuration |

---

## Resource 1: rtx_dns_server

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_dns_server` |
| Type | singleton |
| Import Support | Yes |
| Import ID Format | `dns` |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

### Functional Requirements

#### Create

The Create operation initializes DNS server configuration on the RTX router:

1. WHEN a new `rtx_dns_server` resource is created THEN the provider SHALL execute DNS configuration commands
2. IF `domain_lookup` is false THEN the provider SHALL execute `no dns domain lookup`
3. IF `domain_name` is specified THEN the provider SHALL execute `dns domain <name>`
4. IF `name_servers` is specified THEN the provider SHALL execute `dns server <ip1> [<ip2>] [<ip3>]`
5. The resource ID SHALL always be set to `"dns"` (singleton)

#### Read

The Read operation retrieves current DNS configuration:

1. WHEN reading THEN the provider SHALL execute `show config | grep dns`
2. The provider SHALL parse domain_lookup, domain_name, name_servers, server_select, hosts, service_on, and private_address_spoof
3. IF configuration cannot be parsed THEN the provider SHALL return an error

#### Update

The Update operation modifies existing DNS configuration:

1. WHEN updating THEN the provider SHALL compare current vs desired state
2. IF domain_name changes THEN the provider SHALL delete old domain with `no dns domain` before setting new
3. IF name_servers changes THEN the provider SHALL delete old servers with `no dns server` before setting new
4. IF server_select entries are removed THEN the provider SHALL delete with `no dns server select <id>`
5. IF static hosts are removed THEN the provider SHALL delete with `no dns static <hostname>`

#### Delete

The Delete operation resets DNS configuration to defaults:

1. WHEN deleting THEN the provider SHALL remove all server_select entries
2. THEN the provider SHALL remove all static host entries
3. THEN the provider SHALL execute reset commands: `no dns server`, `no dns domain`, `dns service off`, `dns private address spoof off`
4. After successful deletion THEN the resource ID SHALL be cleared

### Feature Requirements

#### Requirement 1: DNS Server Configuration

**User Story:** As a network administrator, I want to configure upstream DNS servers so that my network can resolve external hostnames.

##### Acceptance Criteria

1. WHEN `name_servers` is specified with 1-3 IP addresses THEN the provider SHALL configure DNS servers
2. IF more than 3 name servers are specified THEN the provider SHALL return a validation error
3. IF an invalid IP address is specified THEN the provider SHALL return a validation error

#### Requirement 2: Domain-based DNS Server Selection

**User Story:** As a network administrator, I want to route DNS queries to specific servers based on domain patterns so that I can use different DNS servers for internal vs external domains.

##### Acceptance Criteria

1. WHEN `server_select` entries are specified THEN the provider SHALL configure domain-based DNS routing
2. EACH server_select entry MUST have an `id` (positive integer), 1-2 `server` blocks, and `query_pattern`
3. EACH `server` block within a `server_select` entry MUST have an `address` (IPv4 or IPv6) and MAY have `edns` (boolean, defaults to false)
4. IF a `server` block has `edns = true` THEN the provider SHALL append `edns=on` after that server's address
5. IF `record_type` is specified (a, aaaa, ptr, mx, ns, cname, any) THEN the provider SHALL include it in the command
6. IF `original_sender` is specified THEN the provider SHALL restrict queries by source IP/CIDR
7. IF `restrict_pp` is specified THEN the provider SHALL restrict queries to a specific PP session

##### Per-Server EDNS Specification

The RTX router supports per-server EDNS configuration in the command syntax:
```
dns server select 500000 . a 1.1.1.1 edns=on 1.0.0.1 edns=on
```

This allows different EDNS settings for each DNS server within a single server_select entry:
- Server 1 may have EDNS enabled while Server 2 has it disabled
- The Terraform schema must support this granularity via nested `server` blocks

#### Requirement 3: Static DNS Hosts

**User Story:** As a network administrator, I want to define static hostname-to-IP mappings so that local hosts can be resolved without external DNS queries.

##### Acceptance Criteria

1. WHEN `hosts` entries are specified THEN the provider SHALL configure static DNS entries
2. EACH host entry MUST have a `name` (hostname) and `address` (IP address)
3. IF host address is invalid THEN the provider SHALL return a validation error

#### Requirement 4: DNS Service Control

**User Story:** As a network administrator, I want to enable or disable the router's DNS relay service so that I can control whether the router responds to DNS queries from clients.

##### Acceptance Criteria

1. WHEN `service_on` is true THEN the provider SHALL execute `dns service recursive`
2. WHEN `service_on` is false THEN the provider SHALL execute `dns service off`

#### Requirement 5: Private Address Spoofing

**User Story:** As a network administrator, I want to control DNS private address spoofing to prevent DNS rebinding attacks.

##### Acceptance Criteria

1. WHEN `private_address_spoof` is true THEN the provider SHALL execute `dns private address spoof on`
2. WHEN `private_address_spoof` is false THEN the provider SHALL execute `dns private address spoof off`

### Non-Functional Requirements

#### Validation

| Attribute | Constraint |
|-----------|------------|
| name_servers | Maximum 3 entries, each must be valid IPv4 address |
| server_select.id | Positive integer (1-65535) |
| server_select.server | 1-2 blocks required (MinItems: 1, MaxItems: 2) |
| server_select.server.address | Must be valid IPv4 or IPv6 address |
| server_select.server.edns | Boolean, defaults to false |
| server_select.record_type | One of: a, aaaa, ptr, mx, ns, cname, any |
| hosts.address | Must be valid IP address |

#### Performance

- DNS configuration operations should complete within the standard command timeout
- Batch commands should be used where possible to minimize round trips

#### Reliability

- The provider SHALL save configuration after changes (`save` command)
- The provider SHALL handle line wrapping in router output (RTX wraps at ~80 chars)

### RTX Commands Reference

```
# DNS Server Commands
dns server <ip1> [<ip2>] [<ip3>]
no dns server

# Domain Configuration
dns domain <name>
no dns domain
dns domain lookup on|off
no dns domain lookup

# Domain-based Server Selection
dns server select <id> <server1> [edns=on] [<server2> [edns=on]] [type] <query-pattern> [original-sender] [restrict pp n]
no dns server select <id>

# Static Host Entries
dns static <hostname> <ip>
no dns static <hostname>

# DNS Service
dns service recursive|off

# Private Address Spoof
dns private address spoof on|off

# Show Commands
show config | grep dns
```

### Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Compare current vs desired DNS configuration |
| `terraform apply` | Required | Configure DNS settings on router |
| `terraform destroy` | Required | Reset DNS configuration to defaults |
| `terraform import` | Required | Import existing DNS configuration |
| `terraform refresh` | Required | Read current DNS configuration |

### Import Specification

- **Import ID Format**: `dns` (fixed string for singleton)
- **Import Command**: `terraform import rtx_dns_server.main dns`
- **Post-Import**: State will contain all current DNS configuration

### Example Usage

```hcl
resource "rtx_dns_server" "main" {
  domain_lookup = true
  domain_name   = "example.com"

  name_servers = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]

  server_select {
    id            = 1
    query_pattern = "internal.example.com"

    server {
      address = "192.168.1.10"
    }
  }

  server_select {
    id              = 2
    record_type     = "any"
    query_pattern   = "."
    original_sender = "192.168.1.0/24"
    restrict_pp     = 1

    server {
      address = "10.0.0.1"
      edns    = true
    }
    server {
      address = "10.0.0.2"
      edns    = true
    }
  }

  # Example with mixed EDNS settings
  server_select {
    id            = 500000
    query_pattern = "."
    record_type   = "a"

    server {
      address = "1.1.1.1"
      edns    = true
    }
    server {
      address = "1.0.0.1"
      edns    = false
    }
  }

  hosts {
    name    = "router"
    address = "192.168.1.1"
  }

  hosts {
    name    = "nas"
    address = "192.168.1.10"
  }

  service_on             = true
  private_address_spoof  = false
}
```

### Terraform Schema

| Attribute | Type | Required | Default | ForceNew | Description |
|-----------|------|----------|---------|----------|-------------|
| domain_lookup | bool | No | true (computed) | No | Enable DNS domain lookup |
| domain_name | string | No | - | No | Default domain name for DNS queries |
| name_servers | list(string) | No | - | No | List of DNS server IPs (max 3) |
| server_select | list(object) | No | - | No | Domain-based DNS server selection |
| server_select.id | int | Yes | - | No | Selector ID (positive integer) |
| server_select.server | list(object) | Yes | - | No | DNS servers with per-server EDNS (MinItems: 1, MaxItems: 2) |
| server_select.server.address | string | Yes | - | No | DNS server IP address (IPv4 or IPv6) |
| server_select.server.edns | bool | No | false | No | Enable EDNS for this server |
| server_select.record_type | string | No | "a" (computed) | No | DNS record type |
| server_select.query_pattern | string | Yes | - | No | Domain pattern to match |
| server_select.original_sender | string | No | - | No | Source IP/CIDR restriction |
| server_select.restrict_pp | int | No | 0 (computed) | No | PP session restriction |
| hosts | list(object) | No | - | No | Static DNS host entries |
| hosts.name | string | Yes | - | No | Hostname |
| hosts.address | string | Yes | - | No | IP address |
| service_on | bool | No | false (computed) | No | Enable DNS service |
| private_address_spoof | bool | No | false (computed) | No | Enable private address spoofing |

---

## Resource 2: rtx_netvolante_dns

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_netvolante_dns` |
| Type | collection (per interface) |
| Import Support | Yes |
| Import ID Format | interface name (e.g., "pp 1") |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

### Functional Requirements

#### Create

1. WHEN a new `rtx_netvolante_dns` resource is created THEN the provider SHALL configure NetVolante DNS
2. The provider SHALL execute commands for hostname, server, timeout, IPv6, auto_hostname, and use
3. IF `use` is true THEN the provider SHALL trigger registration with `netvolante-dns go <interface>`
4. The resource ID SHALL be the interface name

#### Read

1. WHEN reading THEN the provider SHALL execute `show config | grep netvolante-dns`
2. The provider SHALL parse configurations for the specific interface
3. IF configuration not found for interface THEN the resource SHALL be marked for recreation

#### Update

1. WHEN updating THEN the provider SHALL delete existing configuration first
2. THEN the provider SHALL configure new settings
3. This delete-and-recreate approach ensures clean state

#### Delete

1. WHEN deleting THEN the provider SHALL execute `no netvolante-dns hostname host <interface>`
2. The provider SHALL save configuration after deletion

### Feature Requirements

#### Requirement 1: NetVolante DNS Registration

**User Story:** As a network administrator with a dynamic IP, I want to register my router's IP with Yamaha's free NetVolante DNS so that I can access my network remotely using a hostname.

##### Acceptance Criteria

1. WHEN `hostname` is specified with a valid *.netvolante.jp hostname THEN the provider SHALL register it
2. WHEN `interface` is specified THEN the provider SHALL use that interface's IP for registration
3. IF interface is changed THEN the resource SHALL be recreated (ForceNew)

#### Requirement 2: Server Selection

**User Story:** As a network administrator, I want to choose between NetVolante DNS servers for redundancy.

##### Acceptance Criteria

1. WHEN `server` is 1 or 2 THEN the provider SHALL use that NetVolante DNS server
2. IF server is outside 1-2 range THEN the provider SHALL return a validation error

#### Requirement 3: IPv6 Support

**User Story:** As a network administrator, I want to register both IPv4 and IPv6 addresses with NetVolante DNS.

##### Acceptance Criteria

1. WHEN `ipv6_enabled` is true THEN the provider SHALL execute `netvolante-dns use ipv6 <interface> on`

### Non-Functional Requirements

#### Validation

| Attribute | Constraint |
|-----------|------------|
| interface | Required, ForceNew (e.g., "pp 1", "lan1") |
| hostname | Required, must be valid hostname format |
| server | 1-2 |
| timeout | 1-3600 seconds |

### RTX Commands Reference

```
# NetVolante DNS Commands
netvolante-dns hostname host <interface> <hostname>
no netvolante-dns hostname host <interface>
netvolante-dns server <1|2>
netvolante-dns timeout <seconds>
netvolante-dns use ipv6 <interface> on|off
netvolante-dns auto hostname <interface> on|off
netvolante-dns use <interface> on|off
netvolante-dns go <interface>

# Show Commands
show config | grep netvolante-dns
show status netvolante-dns
```

### Import Specification

- **Import ID Format**: interface name (e.g., "pp 1", "lan1")
- **Import Command**: `terraform import rtx_netvolante_dns.main "pp 1"`

### Example Usage

```hcl
resource "rtx_netvolante_dns" "main" {
  interface    = "pp 1"
  hostname     = "myrouter.aa0.netvolante.jp"
  server       = 1
  timeout      = 60
  ipv6_enabled = false
}

resource "rtx_netvolante_dns" "lan" {
  interface     = "lan2"
  hostname      = "home.bb1.netvolante.jp"
  server        = 2
  timeout       = 120
  ipv6_enabled  = true
  auto_hostname = true
}
```

### Terraform Schema

| Attribute | Type | Required | Default | ForceNew | Description |
|-----------|------|----------|---------|----------|-------------|
| interface | string | Yes | - | Yes | Interface for DDNS updates (pp 1, lan1, etc.) |
| hostname | string | Yes | - | No | NetVolante DNS hostname (*.netvolante.jp) |
| server | int | No | 1 | No | NetVolante DNS server (1 or 2) |
| timeout | int | No | 60 | No | Update timeout in seconds (1-3600) |
| ipv6_enabled | bool | No | false | No | Enable IPv6 address registration |
| auto_hostname | bool | No | computed | No | Enable automatic hostname generation |

---

## Resource 3: rtx_ddns

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_ddns` |
| Type | collection (4 server slots) |
| Import Support | Yes |
| Import ID Format | server_id (1-4) |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

### Functional Requirements

#### Create

1. WHEN a new `rtx_ddns` resource is created THEN the provider SHALL configure the DDNS server slot
2. The provider SHALL execute commands for URL, hostname, and optionally user credentials
3. The provider SHALL trigger update with `ddns server go <id>`
4. The resource ID SHALL be the server_id (converted to string)

#### Read

1. WHEN reading THEN the provider SHALL execute `show config | grep "ddns server"`
2. The provider SHALL parse configuration for the specific server ID
3. Password SHALL NOT be read back from router (security)

#### Update

1. WHEN updating THEN the provider SHALL delete existing configuration first
2. THEN the provider SHALL configure new settings

#### Delete

1. WHEN deleting THEN the provider SHALL execute:
   - `no ddns server user <id>`
   - `no ddns server hostname <id>`
   - `no ddns server url <id>`

### Feature Requirements

#### Requirement 1: Custom DDNS Provider Configuration

**User Story:** As a network administrator, I want to configure third-party DDNS providers (No-IP, DynDNS, etc.) so that I can use my preferred DDNS service.

##### Acceptance Criteria

1. WHEN `url` is specified with http:// or https:// THEN the provider SHALL configure the update URL
2. WHEN `hostname` is specified THEN the provider SHALL configure the hostname to update
3. WHEN `username` and `password` are specified THEN the provider SHALL configure credentials
4. IF server_id is outside 1-4 range THEN the provider SHALL return a validation error

### Non-Functional Requirements

#### Validation

| Attribute | Constraint |
|-----------|------------|
| server_id | 1-4, ForceNew |
| url | Required, must start with http:// or https:// |
| hostname | Required, valid hostname format |
| username | Optional |
| password | Optional, Sensitive |

#### Security

- Password is marked as Sensitive and will not appear in plan output
- Password is not read back from router configuration

### RTX Commands Reference

```
# Custom DDNS Commands
ddns server url <id> <url>
no ddns server url <id>
ddns server hostname <id> <hostname>
no ddns server hostname <id>
ddns server user <id> <username> <password>
no ddns server user <id>
ddns server go <id>

# Show Commands
show config | grep "ddns server"
show status ddns
```

### Import Specification

- **Import ID Format**: server_id (1-4)
- **Import Command**: `terraform import rtx_ddns.noip 1`

### Example Usage

```hcl
resource "rtx_ddns" "noip" {
  server_id = 1
  url       = "https://dynupdate.no-ip.com/nic/update"
  hostname  = "myhost.no-ip.org"
  username  = "noip_user"
  password  = var.noip_password  # Use variable for sensitive data
}

resource "rtx_ddns" "dyndns" {
  server_id = 2
  url       = "https://members.dyndns.org/nic/update"
  hostname  = "myhost.dyndns.org"
  username  = "dyndns_user"
  password  = var.dyndns_password
}
```

### Terraform Schema

| Attribute | Type | Required | Default | ForceNew | Sensitive | Description |
|-----------|------|----------|---------|----------|-----------|-------------|
| server_id | int | Yes | - | Yes | No | DDNS server ID (1-4) |
| url | string | Yes | - | No | No | DDNS update URL |
| hostname | string | Yes | - | No | No | DDNS hostname to update |
| username | string | No | - | No | No | DDNS account username |
| password | string | No | - | No | Yes | DDNS account password |

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (current IP, last update time) must not be stored in state
- Password values in `rtx_ddns` are stored in state but marked sensitive
- State comparison uses configuration values only, not runtime status

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation code | Initial master spec creation |
| 2026-01-23 | dns-server-select-per-server-edns | Updated server_select schema to use nested `server` blocks with per-server EDNS support |
