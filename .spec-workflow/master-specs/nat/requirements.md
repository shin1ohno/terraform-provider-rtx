# Master Requirements: NAT Resources

## Overview

This document defines the requirements for NAT (Network Address Translation) resources in the Terraform RTX Provider. NAT enables address translation between internal (private) and external (public) networks, a fundamental feature for enterprise network infrastructure using Yamaha RTX routers.

The NAT functionality is split into two distinct resources:
- **rtx_nat_static**: One-to-one static address mapping (Static NAT)
- **rtx_nat_masquerade**: Many-to-one address translation with port address translation (PAT/NAPT)

## Alignment with Product Vision

NAT resources directly support the core product vision of managing RTX router infrastructure as code:

- **Infrastructure as Code**: Enables declarative NAT configuration management
- **Network Security**: Provides controlled exposure of internal services
- **Enterprise Features**: Supports both basic and advanced NAT scenarios (port forwarding, protocol-based NAT)
- **Consistency**: Ensures NAT configurations can be version-controlled and replicated

## Resource Summary

### rtx_nat_static

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_nat_static` |
| Type | Collection (identified by descriptor_id) |
| Import Support | Yes |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation-derived |

### rtx_nat_masquerade

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_nat_masquerade` |
| Type | Collection (identified by descriptor_id) |
| Import Support | Yes |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation-derived |

---

## rtx_nat_static Requirements

### Core Operations

#### Create

- Creates a new static NAT descriptor with the specified ID
- Supports one-to-one IP address mapping (full NAT)
- Supports port-based static NAT for TCP and UDP protocols
- Executes commands via batch operation for atomic configuration
- Automatically saves configuration to persistent memory after creation

#### Read

- Retrieves NAT static configuration by descriptor ID
- Parses router output to reconstruct complete entry list
- Returns structured data including all mapping entries
- Handles "not found" condition gracefully by clearing resource from state

#### Update

- Compares current entries with desired entries
- Removes entries that are no longer needed
- Adds new entries that don't exist
- Preserves unchanged entries (differential update)
- Uses batch operations for efficiency

#### Delete

- Removes the entire NAT descriptor and all associated entries
- Idempotent - does not error if descriptor already removed
- Automatically saves configuration after deletion

### Feature Requirements

### Requirement 1: Descriptor ID Management

**User Story:** As a network administrator, I want to specify a unique NAT descriptor ID, so that I can organize and reference multiple NAT configurations.

#### Acceptance Criteria

1. WHEN a descriptor_id is provided THEN the system SHALL validate it is between 1 and 65535
2. IF descriptor_id is outside valid range THEN the system SHALL reject with validation error
3. WHEN descriptor_id is changed THEN the resource SHALL be recreated (ForceNew)

### Requirement 2: One-to-One Static NAT Mapping

**User Story:** As a network administrator, I want to map external IP addresses to internal IP addresses, so that internal servers are accessible from the external network.

#### Acceptance Criteria

1. WHEN an entry has inside_local and outside_global IPs THEN the system SHALL create a 1:1 NAT mapping
2. IF inside_local is not a valid IPv4 address THEN the system SHALL reject with validation error
3. IF outside_global is not a valid IPv4 address THEN the system SHALL reject with validation error

### Requirement 3: Port-Based Static NAT

**User Story:** As a network administrator, I want to map specific ports between external and internal addresses, so that I can forward specific services while preserving IP addresses.

#### Acceptance Criteria

1. WHEN protocol is specified THEN both inside_local_port and outside_global_port SHALL be required
2. WHEN ports are specified THEN protocol (tcp or udp) SHALL be required
3. IF protocol is specified without ports THEN the system SHALL reject with validation error
4. IF ports are specified without protocol THEN the system SHALL reject with validation error
5. WHEN port is specified THEN it SHALL be validated between 1 and 65535

### Requirement 4: Multiple Entry Support

**User Story:** As a network administrator, I want to define multiple NAT mappings within a single descriptor, so that related configurations are grouped together.

#### Acceptance Criteria

1. WHEN multiple entries are defined THEN the system SHALL create all mappings under the same descriptor
2. WHEN an entry is removed from configuration THEN the system SHALL delete only that mapping
3. WHEN an entry is added to configuration THEN the system SHALL add only the new mapping
4. WHEN entries are unchanged THEN the system SHALL not issue commands for them

---

## rtx_nat_masquerade Requirements

### Core Operations

#### Create

- Creates a NAT masquerade descriptor with outer address and inner network
- Supports port forwarding via static entries
- Supports protocol-only entries (ESP, AH, GRE, ICMP) without ports
- Executes all commands in batch for atomic configuration
- Automatically saves configuration after creation

#### Read

- Retrieves masquerade configuration by descriptor ID
- Parses outer address, inner network, and static entries
- Returns complete configuration including all port forwarding rules
- Handles "not found" by clearing resource from state

#### Update

- Updates outer address if changed
- Updates inner network if changed
- Removes static entries no longer in configuration
- Adds new static entries
- Updates existing entries with new values

#### Delete

- Removes the entire NAT masquerade descriptor
- Idempotent - ignores "not found" errors
- Saves configuration after deletion

### Feature Requirements

### Requirement 1: Outer Address Configuration

**User Story:** As a network administrator, I want to specify the external address source for NAT, so that I can use dynamically assigned addresses or specific IPs.

#### Acceptance Criteria

1. WHEN outer_address is "ipcp" THEN the system SHALL use PPPoE-assigned address
2. WHEN outer_address is an interface name (pp1, lan2) THEN the system SHALL use that interface's address
3. WHEN outer_address is an IP address THEN the system SHALL use that specific address
4. IF outer_address is empty THEN the system SHALL reject with validation error

### Requirement 2: Inner Network Configuration

**User Story:** As a network administrator, I want to define which internal network uses NAT, so that I can control which hosts are translated.

#### Acceptance Criteria

1. WHEN inner_network is specified THEN the system SHALL use the IP range format (start_ip-end_ip)
2. IF inner_network format is invalid THEN the system SHALL reject with validation error
3. WHEN inner_network is a valid range THEN both start and end IPs SHALL be validated

### Requirement 3: Static Port Forwarding

**User Story:** As a network administrator, I want to forward specific external ports to internal servers, so that services are accessible from outside.

#### Acceptance Criteria

1. WHEN static_entry is defined with tcp/udp protocol THEN both ports SHALL be required
2. WHEN static_entry uses protocol tcp THEN it SHALL create TCP port forwarding
3. WHEN static_entry uses protocol udp THEN it SHALL create UDP port forwarding
4. WHEN entry_number is specified THEN it SHALL uniquely identify the static entry

### Requirement 4: Protocol-Only Forwarding

**User Story:** As a network administrator, I want to forward entire protocols (like ESP for VPN) to internal servers, so that protocol-specific traffic reaches the correct destination.

#### Acceptance Criteria

1. WHEN protocol is esp, ah, gre, or icmp THEN ports SHALL NOT be specified
2. IF protocol is esp/ah/gre/icmp with ports specified THEN the system SHALL reject
3. WHEN protocol-only entry is created THEN the system SHALL forward all traffic for that protocol

---

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Parser, service, and resource files have distinct purposes
- **Modular Design**: Parser functions are reusable across different NAT types
- **Dependency Management**: Services depend only on Executor interface
- **Clear Interfaces**: Client interface defines standard CRUD operations

### Performance

- **Batch Operations**: All create/update/delete operations use RunBatch for efficiency
- **Differential Updates**: Update operations only change what's necessary
- **Connection Reuse**: Operations share SSH connection via executor
- **Timeout Handling**: Context-based cancellation for long-running operations

### Security

- **Input Validation**: All IP addresses and ports are validated before execution
- **No Credential Exposure**: NAT configuration does not involve sensitive data
- **Command Injection Prevention**: Values are validated before building commands

### Reliability

- **Atomic Configuration**: Batch operations ensure consistent state
- **Configuration Persistence**: Auto-save after all modifications
- **Graceful Degradation**: "Not found" errors during delete are handled silently
- **Context Cancellation**: Operations respect context cancellation

### Validation

| Field | Validation | Error Message |
|-------|------------|---------------|
| descriptor_id | 1-65535 | "descriptor ID must be between 1 and 65535" |
| inside_local | Valid IPv4 | "invalid inside_local IP address" |
| outside_global | Valid IPv4 | "invalid outside_global IP address" |
| inside_local_port | 1-65535 | "port must be between 1 and 65535" |
| outside_global_port | 1-65535 | "port must be between 1 and 65535" |
| protocol (static) | tcp, udp | "protocol must be 'tcp' or 'udp'" |
| protocol (masquerade) | tcp, udp, esp, ah, gre, icmp | "protocol must be 'tcp', 'udp', 'esp', 'ah', 'gre', 'icmp', or empty" |
| outer_address | Non-empty | "outer address cannot be empty" |
| inner_network | IP range format | "must be in format 'start_ip-end_ip'" |

---

## RTX Commands Reference

### NAT Static Commands

```
# Set NAT descriptor type to static
nat descriptor type <id> static

# Create 1:1 static NAT mapping
nat descriptor static <id> <outside_ip>=<inside_ip>

# Create port-based static NAT mapping
nat descriptor static <id> <outside_ip>:<port>=<inside_ip>:<port> <protocol>

# Delete static NAT descriptor
no nat descriptor type <id>

# Delete specific 1:1 mapping
no nat descriptor static <id> <outside_ip>=<inside_ip>

# Delete specific port mapping
no nat descriptor static <id> <outside_ip>:<port>=<inside_ip>:<port> <protocol>

# Show NAT descriptor configuration
show config | grep "nat descriptor.*<id>"
show config | grep "nat descriptor"
```

### NAT Masquerade Commands

```
# Set NAT descriptor type to masquerade
nat descriptor type <id> masquerade

# Configure outer (external) address
nat descriptor address outer <id> <address>

# Configure inner (internal) network
nat descriptor address inner <id> <range>

# Create static port forwarding with ports
nat descriptor masquerade static <id> <entry_num> <outer_addr>:<port>=<inner_addr>:<port> [protocol]

# Create protocol-only static entry
nat descriptor masquerade static <id> <entry_num> <inner_addr> <protocol>

# Delete masquerade descriptor
no nat descriptor type <id>

# Delete static entry
no nat descriptor masquerade static <id> <entry_num>

# Show NAT descriptor configuration
show config | grep "nat descriptor" | grep -E "( <id> | <id>$)"
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Shows planned NAT configuration changes |
| `terraform apply` | Required | Creates/updates NAT configurations |
| `terraform destroy` | Required | Removes NAT configurations |
| `terraform import` | Required | Imports existing NAT descriptors |
| `terraform refresh` | Required | Reads current NAT state from router |
| `terraform state` | Required | Manages NAT resource state |

### Import Specification

#### rtx_nat_static
- **Import ID Format**: `<descriptor_id>` (integer)
- **Import Command**: `terraform import rtx_nat_static.example 1`
- **Post-Import**: All entries are imported automatically

#### rtx_nat_masquerade
- **Import ID Format**: `<descriptor_id>` (integer)
- **Import Command**: `terraform import rtx_nat_masquerade.example 1`
- **Post-Import**: Outer address, inner network, and static entries are imported

---

## Example Usage

### rtx_nat_static - Basic 1:1 NAT

```hcl
resource "rtx_nat_static" "server_nat" {
  descriptor_id = 100

  entry {
    inside_local   = "192.168.1.10"
    outside_global = "203.0.113.10"
  }

  entry {
    inside_local   = "192.168.1.11"
    outside_global = "203.0.113.11"
  }
}
```

### rtx_nat_static - Port-Based NAT

```hcl
resource "rtx_nat_static" "web_server" {
  descriptor_id = 101

  entry {
    inside_local        = "192.168.1.100"
    inside_local_port   = 8080
    outside_global      = "203.0.113.1"
    outside_global_port = 80
    protocol            = "tcp"
  }

  entry {
    inside_local        = "192.168.1.100"
    inside_local_port   = 8443
    outside_global      = "203.0.113.1"
    outside_global_port = 443
    protocol            = "tcp"
  }
}
```

### rtx_nat_masquerade - Basic Configuration

```hcl
resource "rtx_nat_masquerade" "lan_nat" {
  descriptor_id = 1
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"
}
```

### rtx_nat_masquerade - With Port Forwarding

```hcl
resource "rtx_nat_masquerade" "lan_nat_with_forwarding" {
  descriptor_id = 2
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  static_entry {
    entry_number        = 1
    inside_local        = "192.168.1.100"
    inside_local_port   = 8080
    outside_global      = "ipcp"
    outside_global_port = 80
    protocol            = "tcp"
  }

  static_entry {
    entry_number        = 2
    inside_local        = "192.168.1.100"
    inside_local_port   = 8443
    outside_global      = "ipcp"
    outside_global_port = 443
    protocol            = "tcp"
  }
}
```

### rtx_nat_masquerade - VPN Protocol Forwarding

```hcl
resource "rtx_nat_masquerade" "vpn_nat" {
  descriptor_id = 1000
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  # Forward ESP protocol for IPsec VPN
  static_entry {
    entry_number = 1
    inside_local = "192.168.1.253"
    protocol     = "esp"
  }

  # Forward AH protocol for IPsec VPN
  static_entry {
    entry_number = 2
    inside_local = "192.168.1.253"
    protocol     = "ah"
  }
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (active sessions, counters) are NOT stored
- Resource ID is the descriptor_id converted to string
- State is refreshed on every plan to detect drift
- "Not found" during read clears resource from state

---

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation Analysis | Initial master spec created from implementation code |
