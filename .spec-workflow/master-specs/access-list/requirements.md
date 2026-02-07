# Master Requirements: Access List Resources

## Overview

This document specifies the requirements for access list (ACL) resources in the Terraform RTX Provider. Access lists provide packet filtering capabilities for IPv4, IPv6, and MAC-layer traffic on Yamaha RTX routers. The provider implements ten resource types covering filter definition, interface binding (apply), and dynamic (stateful) filtering.

## Alignment with Product Vision

Access lists are fundamental security components for network infrastructure. These resources enable infrastructure-as-code management of:
- Network security policies
- Traffic filtering rules
- Protocol-based access control
- Layer 2 (MAC) and Layer 3 (IP) filtering

## Resource Summary

| Resource Name | Type | Import Support | Description |
|---------------|------|----------------|-------------|
| `rtx_access_list_extended` | Collection | Yes | IPv4 extended ACL with named entries |
| `rtx_access_list_extended_ipv6` | Collection | Yes | IPv6 extended ACL with named entries |
| `rtx_access_list_ip` | Singleton | Yes | Individual IPv4 filter rule (native `ip filter`) |
| `rtx_access_list_ip_apply` | Binding | Yes | Apply IPv4 filters to interface (in/out) |
| `rtx_access_list_ip_dynamic` | Collection | Yes | Named collection of dynamic IPv4 filters |
| `rtx_access_list_ipv6` | Singleton | Yes | Individual IPv6 filter rule (native `ipv6 filter`) |
| `rtx_access_list_ipv6_apply` | Binding | Yes | Apply IPv6 filters to interface (in/out) |
| `rtx_access_list_ipv6_dynamic` | Collection | Yes | Named collection of dynamic IPv6 filters |
| `rtx_access_list_mac` | Collection | Yes | MAC address filter with ethernet filter support |
| `rtx_access_list_mac_apply` | Binding | Yes | Apply MAC filters to Ethernet interface (in/out) |

---

## Resource 1: rtx_access_list_extended

### Functional Requirements

#### Core Operations

##### Create
- Creates a new named IPv4 extended access list on the RTX router
- Validates all entries have either `source_any=true` or `source_prefix` specified
- Validates all entries have either `destination_any=true` or `destination_prefix` specified
- Validates `established` flag is only used with TCP protocol

##### Read
- Retrieves access list by name from router configuration
- Returns all entries with their full configuration
- Handles "not found" gracefully by removing from state

##### Update
- Replaces entire access list with new configuration
- Maintains entry ordering based on sequence numbers

##### Delete
- Removes the named access list from router configuration
- Handles "not found" gracefully (idempotent)

### Feature Requirements

#### Requirement 1: Named Access List Management

**User Story:** As a network engineer, I want to manage named access lists, so that I can reference them by name in interface configurations.

##### Acceptance Criteria

1. WHEN creating an access list THEN the system SHALL accept a unique name identifier
2. IF the name already exists THEN the system SHALL replace the existing configuration
3. WHEN deleting an access list THEN the system SHALL remove all associated entries

#### Requirement 2: Protocol-Based Filtering

**User Story:** As a network engineer, I want to filter by protocol, so that I can control traffic at Layer 4.

##### Acceptance Criteria

1. WHEN specifying protocol THEN the system SHALL accept: tcp, udp, icmp, ip, gre, esp, ah, *
2. IF protocol is tcp THEN the system SHALL allow `established` flag
3. IF protocol is tcp or udp THEN the system SHALL allow port specifications

#### Requirement 3: Address Matching

**User Story:** As a network engineer, I want flexible address matching, so that I can create precise rules.

##### Acceptance Criteria

1. WHEN `source_any=true` THEN the system SHALL match any source address
2. WHEN `source_prefix` is specified THEN the system SHALL require `source_prefix_mask`
3. WHEN `destination_any=true` THEN the system SHALL match any destination address

### Schema Specification

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `name` | string | Yes | Yes | Access list name (identifier) |
| `entry` | list | Yes | No | List of ACL entries |
| `entry.sequence` | int | Yes | No | Sequence number (1-65535) |
| `entry.ace_rule_action` | string | Yes | No | Action: permit or deny |
| `entry.ace_rule_protocol` | string | Yes | No | Protocol: tcp, udp, icmp, ip, gre, esp, ah, * |
| `entry.source_any` | bool | No | No | Match any source (default: false) |
| `entry.source_prefix` | string | No | No | Source IP address |
| `entry.source_prefix_mask` | string | No | No | Source wildcard mask |
| `entry.source_port_equal` | string | No | No | Source port equals |
| `entry.source_port_range` | string | No | No | Source port range (e.g., "1024-65535") |
| `entry.destination_any` | bool | No | No | Match any destination (default: false) |
| `entry.destination_prefix` | string | No | No | Destination IP address |
| `entry.destination_prefix_mask` | string | No | No | Destination wildcard mask |
| `entry.destination_port_equal` | string | No | No | Destination port equals |
| `entry.destination_port_range` | string | No | No | Destination port range |
| `entry.established` | bool | No | No | Match established TCP (default: false) |
| `entry.log` | bool | No | No | Enable logging (default: false) |

### Validation Rules

1. Either `source_any=true` OR `source_prefix` must be specified per entry
2. Either `destination_any=true` OR `destination_prefix` must be specified per entry
3. `established` can only be `true` when `ace_rule_protocol` is "tcp"
4. `sequence` must be between 1 and 65535

---

## Resource 2: rtx_access_list_extended_ipv6

### Functional Requirements

Identical to `rtx_access_list_extended` with IPv6-specific adaptations.

#### Core Operations

Same create/read/update/delete semantics as IPv4 extended ACL.

### Feature Requirements

#### Requirement 1: IPv6 Address Support

**User Story:** As a network engineer, I want to filter IPv6 traffic, so that I can secure dual-stack networks.

##### Acceptance Criteria

1. WHEN specifying source/destination THEN the system SHALL accept IPv6 addresses with prefix lengths
2. WHEN specifying protocol THEN the system SHALL accept: tcp, udp, icmpv6, ipv6, ip, *
3. IF prefix length is specified THEN the system SHALL validate range 0-128

### Schema Specification

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `name` | string | Yes | Yes | Access list name (identifier) |
| `entry` | list | Yes | No | List of ACL entries |
| `entry.sequence` | int | Yes | No | Sequence number (1-65535) |
| `entry.ace_rule_action` | string | Yes | No | Action: permit or deny |
| `entry.ace_rule_protocol` | string | Yes | No | Protocol: tcp, udp, icmpv6, ipv6, ip, * |
| `entry.source_any` | bool | No | No | Match any source (default: false) |
| `entry.source_prefix` | string | No | No | Source IPv6 address |
| `entry.source_prefix_length` | int | No | No | Source prefix length (0-128) |
| `entry.source_port_equal` | string | No | No | Source port equals |
| `entry.source_port_range` | string | No | No | Source port range |
| `entry.destination_any` | bool | No | No | Match any destination (default: false) |
| `entry.destination_prefix` | string | No | No | Destination IPv6 address |
| `entry.destination_prefix_length` | int | No | No | Destination prefix length (0-128) |
| `entry.destination_port_equal` | string | No | No | Destination port equals |
| `entry.destination_port_range` | string | No | No | Destination port range |
| `entry.established` | bool | No | No | Match established TCP (default: false) |
| `entry.log` | bool | No | No | Enable logging (default: false) |

### Validation Rules

1. Either `source_any=true` OR `source_prefix` must be specified per entry
2. Either `destination_any=true` OR `destination_prefix` must be specified per entry
3. `established` can only be `true` when `ace_rule_protocol` is "tcp"
4. `source_prefix_length` and `destination_prefix_length` must be 0-128

---

## Resource 3: rtx_access_list_ip

### Functional Requirements

#### Core Operations

##### Create
- Creates an individual IP filter rule using native `ip filter` command
- Uses filter_id as unique identifier (ForceNew)

##### Read
- Retrieves filter configuration by filter number
- Returns complete filter specification

##### Update
- Updates filter in place (same filter_id)

##### Delete
- Removes the IP filter rule
- Uses `no ip filter <id>` command

### Feature Requirements

#### Requirement 1: Native IP Filter Support

**User Story:** As a network engineer, I want to use RTX native IP filters, so that I can leverage familiar command syntax.

##### Acceptance Criteria

1. WHEN creating a filter THEN the system SHALL use the `ip filter` command
2. WHEN specifying action THEN the system SHALL accept: pass, reject, restrict
3. WHEN specifying addresses THEN the system SHALL accept CIDR notation or "*"

### Schema Specification

| Attribute | Type | Required | ForceNew | Computed | Description |
|-----------|------|----------|----------|----------|-------------|
| `sequence` | int | Yes | Yes | No | Filter number (1-65535) |
| `action` | string | Yes | No | No | Action: pass, reject, restrict |
| `source` | string | Yes | No | No | Source IP/CIDR or "*" |
| `destination` | string | Yes | No | No | Destination IP/CIDR or "*" |
| `protocol` | string | No | No | Yes | Protocol: tcp, udp, icmp, ip, gre, esp, ah, tcpfin, tcprst, * |
| `source_port` | string | No | No | Yes | Source port or "*" |
| `dest_port` | string | No | No | Yes | Destination port or "*" |
| `established` | bool | No | No | Yes | Match established TCP connections |

### Validation Rules

1. `filter_id` must be between 1 and 65535
2. `action` must be one of: pass, reject, restrict
3. `established` is only valid for TCP protocol

---

## Resource 4: rtx_access_list_ipv6

### Functional Requirements

#### Core Operations

Same as `rtx_access_list_ip` with IPv6 addressing.

### Feature Requirements

#### Requirement 1: Native IPv6 Filter Support

**User Story:** As a network engineer, I want to use RTX native IPv6 filters, so that I can filter IPv6 traffic with familiar commands.

##### Acceptance Criteria

1. WHEN creating a filter THEN the system SHALL use the `ipv6 filter` command
2. WHEN specifying protocol THEN the system SHALL accept: tcp, udp, icmp6, ip, *, gre, esp, ah

### Schema Specification

| Attribute | Type | Required | ForceNew | Computed | Description |
|-----------|------|----------|----------|----------|-------------|
| `sequence` | int | Yes | Yes | No | Filter number (1-65535) |
| `action` | string | Yes | No | No | Action: pass, reject, restrict |
| `source` | string | Yes | No | No | Source IPv6/prefix or "*" |
| `destination` | string | Yes | No | No | Destination IPv6/prefix or "*" |
| `protocol` | string | Yes | No | No | Protocol: tcp, udp, icmp6, ip, *, gre, esp, ah |
| `source_port` | string | No | No | Yes | Source port or "*" |
| `dest_port` | string | No | No | Yes | Destination port or "*" |

### Validation Rules

1. `filter_id` must be between 1 and 65535
2. `action` must be one of: pass, reject, restrict
3. Protocol "icmp6" is IPv6-specific (not "icmp")

---

## Resource 5: rtx_access_list_mac

### Functional Requirements

#### Core Operations

##### Create
- Creates a MAC address access list with optional ethernet filter mode
- Supports RTX native actions (pass-log, reject-nolog, etc.)

##### Read
- Retrieves MAC ACL by name
- Returns all entries and apply configuration

##### Update
- Updates entries and apply settings

##### Delete
- Removes all filter entries by collected filter numbers

### Feature Requirements

#### Requirement 1: MAC Address Filtering

**User Story:** As a network engineer, I want to filter by MAC address, so that I can implement Layer 2 security.

##### Acceptance Criteria

1. WHEN specifying MAC addresses THEN the system SHALL accept standard MAC format (00:11:22:33:44:55)
2. WHEN specifying wildcard masks THEN the system SHALL allow partial MAC matching
3. WHEN specifying VLAN THEN the system SHALL validate ID range 1-4094

#### Requirement 2: Ethernet Filter Integration

**User Story:** As a network engineer, I want to apply filters to interfaces, so that I can enforce security policies.

##### Acceptance Criteria

1. WHEN apply block is specified THEN the system SHALL configure ethernet filter on interface
2. WHEN direction is "in" THEN the system SHALL apply inbound filtering
3. WHEN direction is "out" THEN the system SHALL apply outbound filtering

#### Requirement 3: Advanced Matching

**User Story:** As a network engineer, I want advanced matching options, so that I can create complex rules.

##### Acceptance Criteria

1. WHEN ether_type is specified THEN the system SHALL filter by Ethernet type (e.g., 0x0800)
2. WHEN dhcp_match is specified THEN the system SHALL use DHCP binding for matching
3. WHEN byte_list is specified THEN the system SHALL match at specified offset

### Schema Specification

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `name` | string | Yes | Yes | Access list name (identifier) |
| `sequence` | int | No | No | Optional RTX filter ID for numeric mode |
| `apply` | list(1) | No | No | Optional interface application |
| `apply.interface` | string | Yes | No | Interface to apply (e.g., lan1) |
| `apply.direction` | string | Yes | No | Direction: in or out |
| `apply.filter_ids` | list(int) | Yes | No | Filter IDs to apply in order |
| `entry` | list | Yes | No | List of MAC ACL entries |
| `entry.sequence` | int | Yes | No | Sequence number (1-99999) |
| `entry.ace_action` | string | Yes | No | Action (see below) |
| `entry.source_any` | bool | No | No | Match any source MAC (default: false) |
| `entry.source_address` | string | No | No | Source MAC address |
| `entry.source_address_mask` | string | No | No | Source MAC wildcard mask |
| `entry.destination_any` | bool | No | No | Match any destination MAC (default: false) |
| `entry.destination_address` | string | No | No | Destination MAC address |
| `entry.destination_address_mask` | string | No | No | Destination MAC wildcard mask |
| `entry.ether_type` | string | No | No | Ethernet type (e.g., 0x0800) |
| `entry.vlan_id` | int | No | No | VLAN ID (1-4094) |
| `entry.log` | bool | No | No | Enable logging (default: false) |
| `entry.filter_id` | int | No | No | Explicit filter number override |
| `entry.dhcp_match` | list(1) | No | No | DHCP-based matching |
| `entry.dhcp_match.type` | string | Yes | No | dhcp-bind or dhcp-not-bind |
| `entry.dhcp_match.scope` | int | No | No | DHCP scope number |
| `entry.offset` | int | No | No | Offset for byte matching |
| `entry.byte_list` | list(string) | No | No | Hex bytes for offset matching |

### Valid ace_action Values

- `permit` - Allow traffic (standard ACL)
- `deny` - Block traffic (standard ACL)
- `pass-log` - Allow with logging (RTX native)
- `pass-nolog` - Allow without logging (RTX native)
- `reject-log` - Block with logging (RTX native)
- `reject-nolog` - Block without logging (RTX native)
- `pass` - Allow (RTX native)
- `reject` - Block (RTX native)

### Validation Rules

1. `sequence` must be between 1 and 99999
2. `vlan_id` must be between 1 and 4094
3. `filter_id` (if specified) must be at least 1
4. `apply.direction` must be "in" or "out"
5. `dhcp_match.type` must be "dhcp-bind" or "dhcp-not-bind"

---

## Resource 6: rtx_access_list_ip_apply

### Functional Requirements

#### Core Operations

##### Create
- Applies IPv4 filters to an interface in a specified direction (in/out)
- Calls `ApplyIPFiltersToInterface`
- Read back to verify consistency

##### Read
- Retrieves applied filters via `GetIPInterfaceFilters`
- Filters state to only include sequences owned by this ACL
- Removes from state if no matching filters found

##### Update
- Replaces applied filter configuration

##### Delete
- Removes applied filters via `RemoveIPFiltersFromInterface`

### Schema Specification

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `access_list` | string | Yes | No | ACL name reference |
| `interface` | string | Yes | Yes | Interface to apply filters on |
| `direction` | string | Yes | Yes | Direction: "in" or "out" |
| `sequences` | list(int) | No | No | Filter sequence numbers to apply |

### Import Specification

- **Import ID Format**: `{access_list}:{interface}:{direction}`
- **Import Command**: `terraform import rtx_access_list_ip_apply.example my-acl:lan1:in`

### Example Usage

```hcl
resource "rtx_access_list_ip_apply" "lan1_in" {
  access_list = rtx_access_list_ip.block_netbios.id
  interface   = "lan1"
  direction   = "in"
  sequences   = [200000, 200001]
}
```

---

## Resource 7: rtx_access_list_ipv6_apply

### Functional Requirements

Identical to `rtx_access_list_ip_apply` with IPv6-specific client methods:
- `ApplyIPv6FiltersToInterface`
- `GetIPv6InterfaceFilters`
- `RemoveIPv6FiltersFromInterface`

### Schema Specification

Same as `rtx_access_list_ip_apply`.

### Example Usage

```hcl
resource "rtx_access_list_ipv6_apply" "lan1_in" {
  access_list = "ipv6-web-acl"
  interface   = "lan1"
  direction   = "in"
  sequences   = [101000, 101001]
}
```

---

## Resource 8: rtx_access_list_mac_apply

### Functional Requirements

#### Core Operations

##### Create
- Applies MAC (Ethernet) filters to a LAN/Bridge interface
- Calls `ApplyMACFiltersToInterface`

##### Read
- Retrieves applied MAC filters via `GetMACInterfaceFilters`

##### Update
- Replaces applied MAC filter configuration

##### Delete
- Removes applied filters via `RemoveMACFiltersFromInterface`

### Schema Specification

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `id` | string | No (Computed) | No | Resource identifier |
| `access_list` | string | Yes | No | MAC ACL name reference |
| `interface` | string | Yes | Yes | Interface (lan/bridge only) |
| `direction` | string | Yes | Yes | Direction: "in" or "out" |
| `sequences` | list(int) | Yes | No | Filter sequence numbers (min 1) |

### Validation Rules

1. `interface` must be `lan*` or `bridge*` (MAC ACLs not supported on pp/tunnel)
2. `sequences` must have at least one element
3. `direction` must be "in" or "out"

### Example Usage

```hcl
resource "rtx_access_list_mac_apply" "lan1_in" {
  access_list = "secure-mac"
  interface   = "lan1"
  direction   = "in"
  sequences   = [101, 102, 103]
}
```

---

## Resource 9: rtx_access_list_ip_dynamic

### Functional Requirements

#### Core Operations

##### Create
- Creates a named collection of dynamic (stateful) IPv4 filter entries
- Supports automatic sequence numbering via `sequence_start` + `sequence_step`
- Checks for sequence conflicts with existing filters on the router

##### Read
- Retrieves dynamic filter collection by name via `GetAccessListIPDynamic`
- Filters router state to only include sequences owned by this ACL

##### Update
- Compares current vs desired entries
- Adds/removes individual dynamic filter entries as needed

##### Delete
- Removes all dynamic filter entries belonging to this collection

### Schema Specification

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `name` | string | Yes | Yes | ACL name (unique identifier) |
| `sequence_start` | int | No | No | Auto-sequence start number (1-65535) |
| `sequence_step` | int | No | No | Auto-sequence increment (default: 10) |
| `entry` | list(object) | Yes | No | List of dynamic filter entries |

#### Entry Object

| Attribute | Type | Required | Computed | Description |
|-----------|------|----------|----------|-------------|
| `sequence` | int | No | Yes | Filter number (auto or manual) |
| `source` | string | Yes | No | Source IP or "*" |
| `destination` | string | Yes | No | Destination IP or "*" |
| `protocol` | string | Yes | No | Protocol (ftp, www, smtp, ssh, tcp, udp, etc.) |
| `syslog` | bool | No | No | Enable syslog (default: false) |
| `timeout` | int | No | No | Timeout in seconds |

### Valid Protocols

`ftp`, `www`, `smtp`, `pop3`, `dns`, `domain`, `telnet`, `ssh`, `tcp`, `udp`, `*`, `tftp`, `submission`, `https`, `imap`, `imaps`, `pop3s`, `smtps`, `ldap`, `ldaps`, `bgp`, `sip`, `ipsec-nat-t`, `ntp`, `snmp`, `rtsp`, `h323`, `pptp`, `l2tp`, `ike`, `esp`

### RTX Commands

```
ip filter dynamic <n> <src> <dst> <protocol> [syslog on] [timeout=<N>]
no ip filter dynamic <n>
```

### Example Usage

```hcl
resource "rtx_access_list_ip_dynamic" "web_filters" {
  name           = "web-dynamic"
  sequence_start = 100
  sequence_step  = 10

  entry {
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = true
  }

  entry {
    source      = "*"
    destination = "*"
    protocol    = "https"
  }

  entry {
    source      = "192.168.1.0/24"
    destination = "*"
    protocol    = "ftp"
    timeout     = 60
  }
}
```

---

## Resource 10: rtx_access_list_ipv6_dynamic

### Functional Requirements

Identical to `rtx_access_list_ip_dynamic` with IPv6-specific differences:
- Uses `ipv6 filter dynamic` commands
- Entry object does **not** have a `timeout` attribute (not supported for IPv6)
- Limited protocol set compared to IPv4

### Schema Specification

Same as `rtx_access_list_ip_dynamic` except:
- Entry does not include `timeout`

### RTX Commands

```
ipv6 filter dynamic <n> <src> <dst> <protocol> [syslog on]
no ipv6 filter dynamic <n>
```

### Example Usage

```hcl
resource "rtx_access_list_ipv6_dynamic" "ipv6_web" {
  name           = "ipv6-web-dynamic"
  sequence_start = 200
  sequence_step  = 10

  entry {
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = true
  }

  entry {
    source      = "*"
    destination = "*"
    protocol    = "ftp"
  }
}
```

---

## Parsing Reliability

- Filter number parsing must correctly handle RTX line wrapping at ~80 characters
- When filter numbers span line boundaries, the parser must reconstruct the original number
- Smart line joining: if a line ends with a digit AND continuation line starts with a digit, join without space (mid-number wrap)
- Round-trip consistency: parse → generate → parse must produce identical results

---

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Each resource file handles one access list type
- **Modular Design**: Builder and flattener functions are isolated and testable
- **Dependency Management**: Resources depend only on client interfaces
- **Clear Interfaces**: Client methods follow consistent Create/Read/Update/Delete pattern

### Performance

- ACL operations should complete within connection timeout
- Large ACLs (100+ entries) should be handled without timeout
- Batch commands used where possible for efficiency

### Security

- Sensitive log content must not be exposed
- Filter rule validation prevents configuration errors
- No credentials stored in state

### Reliability

- Graceful handling of "not found" errors
- Idempotent delete operations
- State recovery on partial failures

### Validation

- Sequence number ranges enforced
- Protocol-specific constraints validated
- Address format validation (IP, IPv6, MAC)

---

## RTX Commands Reference

```
# Extended Access List (IPv4)
ip access-list extended <name>
  <sequence> permit|deny <protocol> <source> <dest> [ports] [established] [log]

# Extended Access List (IPv6)
ipv6 access-list extended <name>
  <sequence> permit|deny <protocol> <source>/<len> <dest>/<len> [ports] [established] [log]

# IP Filter (Native) - filter number 1-65535
ip filter <number> pass|reject|restrict <src> <dest> <proto> [ports] [established]
no ip filter <number>

# IPv6 Filter (Native) - filter number 1-65535
ipv6 filter <number> pass|reject|restrict <src> <dest> <proto> [ports]
no ipv6 filter <number>

# Ethernet Filter (MAC)
ethernet filter <number> pass-log|pass-nolog|reject-log|reject-nolog <src-mac> <dest-mac> [type] [vlan]
no ethernet filter <number>

# Apply IP Filters to Interface (secure filter)
ip <interface> secure filter in|out <filter-numbers...> [dynamic <dynamic-numbers...>]
no ip <interface> secure filter in|out

# Apply IPv6 Filters to Interface
ipv6 <interface> secure filter in|out <filter-numbers...> [dynamic <dynamic-numbers...>]
no ipv6 <interface> secure filter in|out

# Apply MAC Filters to Interface
ethernet <interface> filter in|out <filter-list>
no ethernet <interface> filter in|out

# IP Filter Dynamic
ip filter dynamic <n> <src> <dst> <protocol> [syslog on] [timeout=<N>]
no ip filter dynamic <n>

# IPv6 Filter Dynamic
ipv6 filter dynamic <n> <src> <dst> <protocol> [syslog on]
no ipv6 filter dynamic <n>
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Shows planned ACL changes |
| `terraform apply` | Required | Applies ACL configuration |
| `terraform destroy` | Required | Removes ACL configuration |
| `terraform import` | Required | Imports existing ACL |
| `terraform refresh` | Required | Syncs state with router |
| `terraform state` | Required | State management operations |

### Import Specifications

#### rtx_access_list_extended / rtx_access_list_extended_ipv6
- **Import ID Format**: `<acl-name>`
- **Import Command**: `terraform import rtx_access_list_extended.example my-acl`

#### rtx_access_list_ip / rtx_access_list_ipv6
- **Import ID Format**: `<filter-number>`
- **Import Command**: `terraform import rtx_access_list_ip.example 100`

#### rtx_access_list_ip_apply / rtx_access_list_ipv6_apply
- **Import ID Format**: `<access-list>:<interface>:<direction>`
- **Import Command**: `terraform import rtx_access_list_ip_apply.example my-acl:lan1:in`

#### rtx_access_list_ip_dynamic / rtx_access_list_ipv6_dynamic
- **Import ID Format**: `<acl-name>`
- **Import Command**: `terraform import rtx_access_list_ip_dynamic.example web-dynamic`

#### rtx_access_list_mac
- **Import ID Format**: `<acl-name>`
- **Import Command**: `terraform import rtx_access_list_mac.example my-mac-acl`

#### rtx_access_list_mac_apply
- **Import ID Format**: `<access-list>:<interface>:<direction>`
- **Import Command**: `terraform import rtx_access_list_mac_apply.example my-mac-acl:lan1:in`

---

## Example Usage

### IPv4 Extended ACL

```hcl
resource "rtx_access_list_extended" "web_server" {
  name = "web-server-acl"

  entry {
    sequence          = 10
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_prefix      = "192.168.1.0"
    destination_prefix_mask = "0.0.0.255"
    destination_port_equal  = "443"
  }

  entry {
    sequence          = 20
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_prefix      = "192.168.1.0"
    destination_prefix_mask = "0.0.0.255"
    destination_port_equal  = "80"
  }

  entry {
    sequence          = 100
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    source_any        = true
    destination_any   = true
    log               = true
  }
}
```

### IPv6 Extended ACL

```hcl
resource "rtx_access_list_extended_ipv6" "ipv6_web" {
  name = "ipv6-web-acl"

  entry {
    sequence                  = 10
    ace_rule_action           = "permit"
    ace_rule_protocol         = "tcp"
    source_any                = true
    destination_prefix        = "2001:db8:1::"
    destination_prefix_length = 64
    destination_port_equal    = "443"
  }

  entry {
    sequence          = 20
    ace_rule_action   = "permit"
    ace_rule_protocol = "icmpv6"
    source_any        = true
    destination_any   = true
  }
}
```

### Native IP Filter

```hcl
resource "rtx_access_list_ip" "block_netbios" {
  sequence    = 200000
  action      = "reject"
  source      = "10.0.0.0/8"
  destination = "*"
  protocol    = "tcp"
  source_port = "*"
  dest_port   = "135"
}
```

### Native IPv6 Filter

```hcl
resource "rtx_access_list_ipv6" "allow_icmpv6" {
  sequence    = 101000
  action      = "pass"
  source      = "*"
  destination = "*"
  protocol    = "icmp6"
}
```

### MAC Access List

```hcl
resource "rtx_access_list_mac" "mac_filter" {
  name      = "secure-mac"
  sequence  = 100

  entry {
    sequence       = 10
    ace_action     = "pass-log"
    source_any     = false
    source_address = "00:11:22:33:44:55"
    destination_any = true
    sequence       = 101
  }

  entry {
    sequence        = 20
    ace_action      = "reject-nolog"
    source_any      = true
    destination_any = true
    sequence        = 102
  }

  apply {
    interface  = "lan1"
    direction  = "in"
    filter_ids = [101, 102]
  }
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (hit counters, etc.) are not stored
- Entry order is maintained via sequence numbers
- Filter ID is the unique identifier for native filters

---

## Change History

| Date | Source | Changes |
|------|--------|---------|
| 2025-01-23 | Implementation Analysis | Initial master spec created from implementation |
| 2026-02-07 | Implementation Audit | Full audit: add 5 resources (ip_apply, ipv6_apply, mac_apply, ip_dynamic, ipv6_dynamic); integrate filter spec content; add parsing reliability requirements |
| 2026-02-07 | RTX Reference Sync | IP Filter: action `restrict-nolog` removed (not in RTX reference), filter number range 1-65535 |
