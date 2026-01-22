# Requirements Document: Security and Filter Resources

## Introduction

This specification addresses the implementation of security, filter, and VPN-related resource types that are currently missing from the Terraform provider. These resources are essential for complete router configuration management and include:

- **IP/IPv6 Filters**: Layer 3/4 packet filtering and firewall rules
- **Ethernet Filters**: Layer 2 MAC address filtering
- **L2TP Service**: L2TP/L2TPv3 protocol service configuration
- **IPsec Transport**: IPsec transport mode mappings for L2TP over IPsec

Each resource type will support full CRUD (Create, Read, Update, Delete) operations and Terraform import functionality.

## Alignment with Product Vision

From `product.md`:
- "Ethernet Filters: Layer 2 packet filtering"
- "IP Filters: Layer 3/4 packet filtering and firewall rules"
- "Enable complete IaC management of Yamaha RTX router configurations"

Without filter resources, users cannot manage a significant portion of their router's security configuration via Terraform.

## Current State Analysis

Based on actual router configuration (`config.txt`), the following filter types are in use:

### IP Filters (Layer 3/4)
```
ip filter 200000 reject 10.0.0.0/8 * * * *
ip filter 200020 reject * * udp,tcp 135 *
ip filter 200099 pass * * * * *
ip filter 200100 pass * * udp * 500
ip filter dynamic 200080 * * ftp syslog=off
```

### IPv6 Filters
```
ipv6 filter 101000 pass * * icmp6 * *
ipv6 filter 101099 pass * * * * *
ipv6 filter dynamic 101080 * * ftp syslog=off
```

### Ethernet Filters (Layer 2)
```
ethernet filter 1 reject-nolog bc:5c:17:05:59:3a *:*:*:*:*:*
ethernet filter 100 pass *:*:*:*:*:* *:*:*:*:*:*
ethernet lan1 filter in 1 100
ethernet lan1 filter out 2 100
```

### L2TP Service
```
l2tp service on l2tpv3 l2tp
```

### IPsec Transport
```
ipsec transport 1 101 udp 1701
ipsec transport 3 3 udp 1701
```

## Requirements

### REQ-1: IP Filter Resource

**User Story:** As a network administrator, I want to define IP packet filters in Terraform, so that firewall rules can be version-controlled and consistently deployed.

#### Acceptance Criteria

1. WHEN `ip filter N action src dst proto sport dport` is defined THEN a `rtx_access_list_ip` resource SHALL represent it
2. WHEN filter action is `reject`, `pass`, or `restrict` THEN it SHALL be captured as an attribute
3. WHEN source/destination are network ranges THEN they SHALL be parsed correctly
4. WHEN protocol is `udp,tcp` (combined) THEN it SHALL be represented as a list
5. WHEN port is a named service (e.g., `netbios_ns-netbios_ssn`) THEN it SHALL be preserved

#### Resource Schema (Draft)
```hcl
resource "rtx_access_list_ip" "block_netbios" {
  filter_id   = 200020
  action      = "reject"
  source      = "*"
  destination = "*"
  protocol    = ["udp", "tcp"]
  source_port = 135
  dest_port   = "*"
}
```

### REQ-2: IP Filter Dynamic Resource

**User Story:** As a network administrator, I want to define dynamic (stateful) IP filters, so that connection tracking rules are managed via Terraform.

#### Acceptance Criteria

1. WHEN `ip filter dynamic N src dst service options` is defined THEN a `rtx_access_list_ip_dynamic` resource SHALL represent it
2. WHEN service is a named service (ftp, domain, www, etc.) THEN it SHALL be captured
3. WHEN options like `syslog=off` are present THEN they SHALL be captured as attributes

#### Resource Schema (Draft)
```hcl
resource "rtx_access_list_ip_dynamic" "ftp" {
  filter_id = 200080
  source    = "*"
  destination = "*"
  service   = "ftp"
  syslog    = false
}
```

### REQ-3: IPv6 Filter Resource

**User Story:** As a network administrator, I want to define IPv6 packet filters, so that IPv6 security policies can be managed via Terraform.

#### Acceptance Criteria

1. WHEN `ipv6 filter N action src dst proto sport dport` is defined THEN a `rtx_access_list_ipv6` resource SHALL represent it
2. WHEN protocol is `icmp6` THEN it SHALL be correctly identified
3. WHEN filter is dynamic THEN `rtx_access_list_ipv6_dynamic` resource SHALL be used

#### Resource Schema (Draft)
```hcl
resource "rtx_access_list_ipv6" "allow_icmp6" {
  filter_id   = 101000
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "icmp6"
  source_port = "*"
  dest_port   = "*"
}
```

### REQ-4: Ethernet Filter Resource (EXISTING)

**Status:** Already implemented as `rtx_access_list_mac`

**User Story:** As a network administrator, I want to define Layer 2 MAC address filters, so that Ethernet-level access control can be managed via Terraform.

**Note:** The existing `rtx_access_list_mac` resource generates RTX `ethernet filter` commands internally. No new resource is needed.

### REQ-5: Interface Filter Binding

**User Story:** As a network administrator, I want to associate filters with interfaces, so that filter application points are managed via Terraform.

#### Acceptance Criteria

1. WHEN `ethernet lanN filter in/out filter_ids...` is defined THEN interface SHALL reference filter resources
2. WHEN `ip lanN secure filter in/out filter_ids...` is defined THEN interface SHALL reference filter resources
3. WHEN multiple filters are applied THEN order SHALL be preserved

**Note:** This may be handled as attributes on existing `rtx_interface` resource rather than separate resource.

### REQ-6: L2TP Service Resource

**User Story:** As a network administrator, I want to configure the global L2TP service settings, so that L2TP/L2TPv3 enablement is managed via Terraform.

#### Acceptance Criteria

1. WHEN `l2tp service on` is configured THEN service SHALL be enabled
2. WHEN service supports multiple protocols (l2tpv3, l2tp) THEN they SHALL be listed as attributes
3. IF service is `off` THEN resource SHALL reflect disabled state

#### Resource Schema (Draft)
```hcl
resource "rtx_l2tp_service" "main" {
  enabled   = true
  protocols = ["l2tpv3", "l2tp"]
}
```

### REQ-7: IPsec Transport Resource

**User Story:** As a network administrator, I want to define IPsec transport mode mappings, so that L2TP over IPsec configurations are complete.

#### Acceptance Criteria

1. WHEN `ipsec transport N tunnel_id proto port` is defined THEN a `rtx_ipsec_transport` resource SHALL represent it
2. WHEN protocol is `udp` and port is `1701` (L2TP) THEN they SHALL be captured
3. WHEN transport maps to a tunnel THEN tunnel_id SHALL reference the IPsec tunnel

#### Resource Schema (Draft)
```hcl
resource "rtx_ipsec_transport" "l2tp_tunnel1" {
  transport_id = 1
  tunnel_id    = 101
  protocol     = "udp"
  port         = 1701
}
```

## Non-Functional Requirements

### Code Architecture and Modularity

- **Consistent Naming**: Follow Cisco IOS XE naming conventions where applicable
- **Parser Registry**: Register parsers in the existing parser registry pattern
- **Resource Consistency**: Follow existing resource implementation patterns

### Performance

- Filter parsing should handle configurations with 100+ filter rules within 5 seconds

### Reliability

- Invalid filter syntax should produce clear error messages
- Circular references between filters and interfaces should be detected

### Usability

- All resources SHALL support full CRUD operations (Create, Read, Update, Delete)
- All resources SHALL support `terraform import` for existing configurations
- Documentation should include common filter pattern examples

## Implementation Priority

1. **P0 - Core Filters**: `rtx_access_list_ip`, `rtx_access_list_ipv6` (most commonly used)
2. **P1 - Dynamic Filters**: Already exist as `rtx_ip_filter_dynamic`, `rtx_ipv6_filter_dynamic` (rename consideration)
3. **P1 - Layer 2**: Already exists as `rtx_access_list_mac`
4. **P2 - Supporting**: `rtx_l2tp_service`, `rtx_ipsec_transport`

## Dependencies

- Parser infrastructure established in `import-fidelity-fix` spec provides patterns for command parsing
- Interface filter binding may require updates to existing `rtx_interface` resource
- L2TP and IPsec transport resources complement existing `rtx_ipsec_tunnel` and VPN resources

## Estimated Scope

- **New Resources**: 4 resource types (`rtx_access_list_ip`, `rtx_access_list_ipv6`, `rtx_l2tp_service`, `rtx_ipsec_transport`)
- **Existing Resources**: Layer 2 filter covered by `rtx_access_list_mac`
- **New Parsers**: 2 parser implementations (L2TP service, IPsec transport)
- **Files Changed**: ~10-12 new files
- **Complexity**: Medium (leverages existing parser patterns)
