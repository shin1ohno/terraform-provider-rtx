# Master Requirements: Filter Resources

## Overview

This specification covers the filter-related Terraform resources for managing packet filtering on Yamaha RTX routers. The filter resources provide Layer 2 (Ethernet) and Layer 3 (IP/IPv6) traffic filtering capabilities, including both static and stateful (dynamic) filters, as well as interface bindings for applying filters.

## Alignment with Product Vision

The filter resources support the product goal of providing comprehensive network security management through Terraform:
- **Infrastructure as Code**: Define firewall rules declaratively
- **Stateful Inspection**: Enable dynamic filters for application-aware filtering
- **Multi-Layer Security**: Support both L2 (MAC/Ethernet) and L3 (IP/IPv6) filtering
- **Centralized Policy Management**: Apply filters to interfaces consistently

## Resource Summary

| Resource Name | Type | Import Support | ID Format | Description |
|---------------|------|----------------|-----------|-------------|
| `rtx_ethernet_filter` | collection | Yes | `{number}` | Layer 2 (Ethernet/MAC) filter rules |
| `rtx_interface_acl` | collection | Yes | `{interface}` | Interface ACL bindings (IP/IPv6) |
| `rtx_interface_mac_acl` | collection | Yes | `{interface}` | Interface MAC ACL bindings |
| `rtx_ip_filter_dynamic` | collection | Yes | `{filter_id}` | IPv4 stateful/dynamic filters |
| `rtx_ipv6_filter_dynamic` | singleton | Yes | `ipv6_filter_dynamic` | IPv6 stateful/dynamic filters |

---

## Resource: rtx_ethernet_filter

### Functional Requirements

#### Core Operations

**Create**: Creates a new Ethernet (Layer 2) filter rule on the RTX router. The filter can match traffic based on MAC addresses, Ethernet types, VLAN IDs, or DHCP binding status.

**Read**: Retrieves an existing Ethernet filter configuration by filter number.

**Update**: Modifies an existing Ethernet filter by re-applying the configuration with the same number.

**Delete**: Removes an Ethernet filter by filter number.

### Schema Attributes

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `number` | Int | Yes | Yes | Filter number | 1-512 |
| `action` | String | Yes | No | Action to take | `pass-log`, `pass-nolog`, `reject-log`, `reject-nolog`, `pass`, `reject` |
| `source_mac` | String | No | No | Source MAC address | MAC format or `*` for any |
| `destination_mac` | String | No | No | Destination MAC address | MAC format or `*` for any |
| `ether_type` | String | No | No | Ethernet type | Hex format (e.g., `0x0800`) |
| `vlan_id` | Int | No | No | VLAN ID to match | 1-4094 |
| `dhcp_type` | String | No | No | DHCP filter type | `dhcp-bind`, `dhcp-not-bind` |
| `dhcp_scope` | Int | No | No | DHCP scope number | >= 1 |
| `offset` | Int | No | No | Byte offset for byte-match filtering | - |
| `byte_list` | List(String) | No | No | Byte patterns for byte-match filtering | - |

### Attribute Conflicts

- `source_mac`, `destination_mac`, `ether_type`, `vlan_id`, `offset`, `byte_list` conflict with `dhcp_type`

### Acceptance Criteria

1. WHEN a user creates an Ethernet filter with MAC addresses THEN the system SHALL create the filter with the specified source and destination MACs
2. WHEN a user creates a DHCP-based filter THEN the system SHALL reject any MAC address parameters
3. WHEN a user specifies a filter number outside 1-512 THEN the system SHALL return a validation error
4. WHEN a user imports an Ethernet filter THEN the system SHALL read the filter configuration and populate all attributes

---

## Resource: rtx_interface_acl

### Functional Requirements

#### Core Operations

**Create**: Binds access control lists to an interface for inbound and/or outbound traffic filtering at Layer 3 (IPv4/IPv6).

**Read**: Retrieves the current ACL bindings for an interface.

**Update**: Modifies ACL bindings by removing existing configuration and re-applying new bindings.

**Delete**: Removes all ACL bindings from an interface.

### Schema Attributes

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `interface` | String | Yes | Yes | Interface name | Pattern: `^(lan[0-9]+|pp[0-9]+|tunnel[0-9]+|bridge[0-9]+|vlan[0-9]+)$` |
| `ip_access_group_in` | String | No | No | Inbound IPv4 access list name | - |
| `ip_access_group_out` | String | No | No | Outbound IPv4 access list name | - |
| `ipv6_access_group_in` | String | No | No | Inbound IPv6 access list name | - |
| `ipv6_access_group_out` | String | No | No | Outbound IPv6 access list name | - |
| `dynamic_filters_in` | List(Int) | No | No | Inbound dynamic filter numbers | - |
| `dynamic_filters_out` | List(Int) | No | No | Outbound dynamic filter numbers | - |
| `ipv6_dynamic_filters_in` | List(Int) | No | No | Inbound IPv6 dynamic filter numbers | - |
| `ipv6_dynamic_filters_out` | List(Int) | No | No | Outbound IPv6 dynamic filter numbers | - |

### Acceptance Criteria

1. WHEN a user creates an interface ACL binding THEN the system SHALL apply filters in the specified directions
2. WHEN a user specifies an invalid interface name THEN the system SHALL return a validation error
3. WHEN an interface ACL is not found during read THEN the system SHALL remove the resource from state
4. WHEN a user imports an interface ACL THEN the system SHALL read all filter bindings for that interface

---

## Resource: rtx_interface_mac_acl

### Functional Requirements

#### Core Operations

**Create**: Binds MAC access control lists to an interface for Layer 2 filtering.

**Read**: Retrieves the current MAC ACL bindings for an interface.

**Update**: Modifies MAC ACL bindings by removing existing configuration and re-applying new bindings.

**Delete**: Removes all MAC ACL bindings from an interface.

### Schema Attributes

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `interface` | String | Yes | Yes | Interface name | Pattern: `^(lan[0-9]+|bridge[0-9]+|vlan[0-9]+)$` |
| `mac_access_group_in` | String | No | No | Inbound MAC access list name | - |
| `mac_access_group_out` | String | No | No | Outbound MAC access list name | - |

### Acceptance Criteria

1. WHEN a user creates a MAC ACL binding THEN the system SHALL apply Ethernet filters to the interface
2. WHEN a user specifies an interface that doesn't support MAC ACL (like `pp1`) THEN the system SHALL return a validation error
3. WHEN a MAC ACL binding is not found during read THEN the system SHALL remove the resource from state

---

## Resource: rtx_ip_filter_dynamic

### Functional Requirements

#### Core Operations

**Create**: Creates a new IPv4 dynamic (stateful) filter for application-aware packet inspection. Supports two forms:
- **Form 1 (Protocol-based)**: Specify a protocol for automatic stateful inspection
- **Form 2 (Filter-reference)**: Reference static IP filters for custom rules

**Read**: Retrieves a dynamic filter configuration by filter ID.

**Update**: Modifies an existing dynamic filter by re-creating it with the same number.

**Delete**: Removes a dynamic filter by filter ID.

### Schema Attributes

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `filter_id` | Int | Yes | Yes | Filter number (unique identifier) | >= 1 |
| `source` | String | Yes | No | Source address or `*` for any | IP address, CIDR, or `*` |
| `destination` | String | Yes | No | Destination address or `*` for any | IP address, CIDR, or `*` |
| `protocol` | String | No | No | Protocol for stateful inspection (Form 1) | See valid protocols below |
| `filter_list` | List(Int) | No | No | Static filter numbers to reference (Form 2) | - |
| `in_filter_list` | List(Int) | No | No | Inbound filter numbers (Form 2) | - |
| `out_filter_list` | List(Int) | No | No | Outbound filter numbers (Form 2) | - |
| `syslog` | Bool | No | No | Enable syslog logging | Default: `false` |
| `timeout` | Int | No | No | Timeout value in seconds | >= 1 |

### Valid Protocols (Form 1)

`ftp`, `www`, `smtp`, `pop3`, `dns`, `domain`, `telnet`, `ssh`, `tcp`, `udp`, `*`, `tftp`, `submission`, `https`, `imap`, `imaps`, `pop3s`, `smtps`, `ldap`, `ldaps`, `bgp`, `sip`, `ipsec-nat-t`, `ntp`, `snmp`, `rtsp`, `h323`, `pptp`, `l2tp`, `ike`, `esp`

### Attribute Conflicts

- `protocol` conflicts with `filter_list`, `in_filter_list`, `out_filter_list`

### Acceptance Criteria

1. WHEN a user creates a Form 1 dynamic filter THEN the system SHALL create stateful inspection for the specified protocol
2. WHEN a user creates a Form 2 dynamic filter THEN the system SHALL reference the specified static filters
3. WHEN neither `protocol` nor `filter_list` is specified THEN the system SHALL return an error
4. WHEN both `protocol` and `filter_list` are specified THEN the system SHALL return a conflict error
5. WHEN a user imports a dynamic filter THEN the system SHALL correctly identify Form 1 vs Form 2

---

## Resource: rtx_ipv6_filter_dynamic

### Functional Requirements

#### Core Operations

**Create**: Creates IPv6 dynamic (stateful) filters for the router. This is a singleton-like resource that manages all IPv6 dynamic filter entries.

**Read**: Retrieves all IPv6 dynamic filter configurations.

**Update**: Modifies IPv6 dynamic filter configuration by updating all entries.

**Delete**: Removes all IPv6 dynamic filter entries.

### Schema Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `entry` | List(Object) | Yes | List of dynamic filter entries |

#### Entry Object Attributes

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `number` | Int | Yes | Filter number | >= 1 |
| `source` | String | Yes | Source address or `*` | IPv6 address or `*` |
| `destination` | String | Yes | Destination address or `*` | IPv6 address or `*` |
| `protocol` | String | Yes | Protocol | `ftp`, `www`, `smtp`, `submission`, `pop3`, `dns`, `domain`, `telnet`, `ssh`, `tcp`, `udp`, `*` |
| `syslog` | Bool | No | Enable syslog | Default: `false` |

### Acceptance Criteria

1. WHEN a user creates IPv6 dynamic filters THEN the system SHALL create all specified entries
2. WHEN a user updates IPv6 dynamic filters THEN the system SHALL update all entries to match the new configuration
3. WHEN a user deletes IPv6 dynamic filters THEN the system SHALL remove all entries

---

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Each resource file handles one Terraform resource
- **Modular Design**: Service layer (`*_service.go`) separates from provider layer (`resource_*.go`)
- **Dependency Management**: Resources depend on client interfaces defined in `interfaces.go`
- **Clear Interfaces**: Parser layer provides clean command building and output parsing

### Performance

- Configuration changes are saved automatically after each operation
- Filter lookups use `grep` for efficient pattern matching
- List operations parse all relevant filter lines in a single command

### Parsing Reliability

- Filter number parsing must correctly handle RTX line wrapping at ~80 characters
- When filter numbers span line boundaries (e.g., `20010` at line end, `0` at next line), the parser must reconstruct the original number (`200100`)
- Smart line joining: if a line ends with a digit AND continuation line starts with a digit, join without space (mid-number wrap)
- Round-trip consistency: parse → generate → parse must produce identical results for all filter number configurations

### Security

- No sensitive data in filter rules (no passwords, no credentials)
- Filter numbers provide isolation between different filter rules
- Configuration is persisted to router non-volatile memory

### Validation

- Filter numbers validated against RTX router limits (1-512 for Ethernet, 1-65535 for IP)
- Interface names validated against RTX naming patterns
- Protocol names validated against allowed values
- MAC address formats validated (colon-separated or wildcard)
- EtherType validated as hex format

---

## RTX Commands Reference

### Ethernet Filter Commands

```
ethernet filter <n> <action> <src_mac> <dst_mac> [<eth_type>] [vlan <vlan_id>]
ethernet filter <n> <action> dhcp-bind|dhcp-not-bind [<scope>]
ethernet filter <n> <action> <src_mac> [<dst_mac>] offset=<N> <byte1> <byte2> ...
no ethernet filter <n>
ethernet <interface> filter <direction> <filter_numbers...>
no ethernet <interface> filter <direction>
show config | grep "ethernet filter"
```

### IP Filter Dynamic Commands

```
ip filter dynamic <n> <src> <dst> <protocol> [syslog on] [timeout=<N>]
ip filter dynamic <n> <src> <dst> filter <list> [in <list>] [out <list>] [syslog on] [timeout=<N>]
no ip filter dynamic <n>
show config | grep "ip filter"
```

### IPv6 Filter Dynamic Commands

```
ipv6 filter dynamic <n> <src> <dst> <protocol> [syslog on]
no ipv6 filter dynamic <n>
show config | grep "ipv6 filter"
```

### Interface Secure Filter Commands

```
ip <interface> secure filter <direction> <filter_numbers...> [dynamic <dynamic_numbers...>]
no ip <interface> secure filter <direction>
ipv6 <interface> secure filter <direction> <filter_numbers...> [dynamic <dynamic_numbers...>]
no ipv6 <interface> secure filter <direction>
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Shows planned filter changes |
| `terraform apply` | Required | Applies filter configuration to router |
| `terraform destroy` | Required | Removes filter configuration |
| `terraform import` | Required | Imports existing filter configuration |
| `terraform refresh` | Required | Refreshes state from router |
| `terraform state` | Required | State management operations |

### Import Specifications

| Resource | Import ID Format | Example |
|----------|------------------|---------|
| `rtx_ethernet_filter` | `{filter_number}` | `terraform import rtx_ethernet_filter.myfilter 100` |
| `rtx_interface_acl` | `{interface_name}` | `terraform import rtx_interface_acl.lan1 lan1` |
| `rtx_interface_mac_acl` | `{interface_name}` | `terraform import rtx_interface_mac_acl.lan1 lan1` |
| `rtx_ip_filter_dynamic` | `{filter_id}` | `terraform import rtx_ip_filter_dynamic.http 100` |
| `rtx_ipv6_filter_dynamic` | `ipv6_filter_dynamic` | `terraform import rtx_ipv6_filter_dynamic.main ipv6_filter_dynamic` |

---

## Example Usage

### Ethernet Filter (MAC-based)

```hcl
resource "rtx_ethernet_filter" "allow_known_mac" {
  number          = 100
  action          = "pass-log"
  source_mac      = "00:11:22:33:44:55"
  destination_mac = "*"
}

resource "rtx_ethernet_filter" "block_arp" {
  number          = 200
  action          = "reject-nolog"
  source_mac      = "*"
  destination_mac = "*"
  ether_type      = "0x0806"
}

resource "rtx_ethernet_filter" "vlan_filter" {
  number          = 300
  action          = "pass-nolog"
  source_mac      = "*"
  destination_mac = "*"
  vlan_id         = 100
}
```

### Ethernet Filter (DHCP-based)

```hcl
resource "rtx_ethernet_filter" "dhcp_bound" {
  number     = 400
  action     = "pass-log"
  dhcp_type  = "dhcp-bind"
  dhcp_scope = 1
}
```

### Interface ACL

```hcl
resource "rtx_interface_acl" "lan1_acl" {
  interface           = "lan1"
  ip_access_group_in  = "acl-inbound"
  ip_access_group_out = "acl-outbound"
  dynamic_filters_in  = [100, 101, 102]
  dynamic_filters_out = [200]
}
```

### Interface MAC ACL

```hcl
resource "rtx_interface_mac_acl" "lan1_mac" {
  interface            = "lan1"
  mac_access_group_in  = "mac-acl-in"
  mac_access_group_out = "mac-acl-out"
}
```

### IP Filter Dynamic (Form 1 - Protocol)

```hcl
resource "rtx_ip_filter_dynamic" "http" {
  filter_id   = 100
  source      = "*"
  destination = "*"
  protocol    = "www"
  syslog      = true
}

resource "rtx_ip_filter_dynamic" "ftp" {
  filter_id   = 101
  source      = "192.168.1.0/24"
  destination = "*"
  protocol    = "ftp"
  timeout     = 60
}
```

### IP Filter Dynamic (Form 2 - Filter Reference)

```hcl
resource "rtx_ip_filter_dynamic" "custom" {
  filter_id       = 200
  source          = "*"
  destination     = "*"
  filter_list     = [1000, 1001]
  in_filter_list  = [2000, 2001]
  out_filter_list = [3000]
  syslog          = true
}
```

### IPv6 Filter Dynamic

```hcl
resource "rtx_ipv6_filter_dynamic" "main" {
  entry {
    number      = 100
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = true
  }

  entry {
    number      = 101
    source      = "*"
    destination = "*"
    protocol    = "ftp"
  }
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (e.g., active sessions, hit counters) are not stored
- Filter numbers serve as unique identifiers for collection resources
- Interface names serve as unique identifiers for interface binding resources

---

## Change History

| Date | Source | Changes |
|------|--------|---------|
| 2025-01-23 | Implementation Code | Initial master spec created from implementation analysis |
| 2026-01-23 | filter-number-parsing-fix | Added parsing reliability requirements for line wrapping handling |
| 2026-01-23 | terraform-plan-differences-fix | Ethernet filter parser accepts `*:*:*:*:*:*` MAC wildcard format; regex patterns use `.*$` for line wrapping |
