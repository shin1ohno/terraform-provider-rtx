# Master Requirements: Routing Resources

## Overview

The routing resources provide comprehensive management of IP routing configuration on Yamaha RTX routers. This includes static routes for manual path definition, BGP for inter-AS routing and Internet peering, and OSPF for dynamic intra-domain routing. These resources enable network administrators to define, version-control, and automate their routing infrastructure through Terraform.

## Alignment with Product Vision

These routing resources directly support the product's goal of enabling Infrastructure as Code (IaC) management of RTX routers by providing:
- Declarative configuration of complex routing topologies
- Version-controlled routing policy changes
- Automated deployment of routing configurations across multiple sites
- Cisco-compatible naming conventions for familiar user experience

## Resource Summary

### rtx_bgp

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_bgp` |
| Type | singleton |
| Import Support | yes |
| Import ID | `bgp` |
| Last Updated | 2025-01-23 |
| Source Files | `resource_rtx_bgp.go`, `bgp_service.go`, `bgp.go` |

### rtx_ospf

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_ospf` |
| Type | singleton |
| Import Support | yes |
| Import ID | `ospf` |
| Last Updated | 2025-01-23 |
| Source Files | `resource_rtx_ospf.go`, `ospf_service.go`, `ospf.go` |

### rtx_static_route

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_static_route` |
| Type | collection |
| Import Support | yes |
| Import ID | `<prefix>/<mask>` (e.g., `10.0.0.0/255.0.0.0`) |
| Last Updated | 2025-01-23 |
| Source Files | `resource_rtx_static_route.go`, `static_route_service.go`, `static_route.go` |

---

## Functional Requirements

### rtx_bgp Resource

#### Core Operations

##### Create
- Configure BGP autonomous system number on the router
- Set router ID if specified (optional)
- Configure BGP neighbors with their remote AS and connection parameters
- Configure networks to announce via BGP import filters
- Enable route redistribution if specified
- Enable BGP service (`bgp use on`)
- Save configuration to persistent memory

##### Read
- Execute `show config | grep bgp` to retrieve current configuration
- Parse BGP configuration including ASN, router ID, neighbors, networks
- Detect if BGP is enabled or disabled
- If BGP is disabled or not configured, remove resource from state

##### Update
- Compare current configuration with desired state
- Add/remove/modify neighbors as needed
- Update redistribution settings
- Maintain BGP session stability during updates where possible

##### Delete
- Disable BGP service (`bgp use off`)
- Remove BGP from operational state
- Save configuration to persistent memory

#### Feature Requirements

##### Requirement 1: AS Number Management

**User Story:** As a network administrator, I want to configure my BGP autonomous system number, so that my router can participate in BGP routing.

**Acceptance Criteria:**
1. WHEN a valid ASN (1-65535) is provided THEN the system SHALL configure `bgp autonomous-system <asn>`
2. WHEN an invalid ASN is provided THEN the system SHALL reject with a validation error
3. RTX supports 2-byte ASN only (1-65535)

##### Requirement 2: BGP Neighbor Configuration

**User Story:** As a network administrator, I want to configure BGP neighbors with their connection parameters, so that I can establish peering sessions.

**Acceptance Criteria:**
1. WHEN a neighbor is defined with IP and remote_as THEN the system SHALL configure `bgp neighbor <n> <as> <ip>`
2. WHEN hold_time is specified (3-28800) THEN the system SHALL include `hold-time=<seconds>` as inline parameter
3. WHEN local_address is specified THEN the system SHALL include `local-address=<ip>` as inline parameter
4. WHEN passive mode is enabled THEN the system SHALL include `passive=on` as inline parameter
5. WHEN password is specified THEN the system SHALL configure `bgp neighbor pre-shared-key <n> text <password>`
6. All options are inline parameters in the neighbor command: `bgp neighbor <n> <as> <ip> [hold-time=<sec>] [local-address=<ip>] [passive=on]`

##### Requirement 3: Network Announcements

**User Story:** As a network administrator, I want to announce networks via BGP, so that my routes are advertised to peers.

**Acceptance Criteria:**
1. WHEN a network is defined with prefix and mask THEN the system SHALL configure `bgp import filter <n> include <prefix>/<cidr>`
2. WHEN multiple networks are defined THEN each SHALL be configured with unique filter IDs
3. Network mask in Terraform schema is dotted decimal but output as CIDR notation (e.g., /24)

##### Requirement 4: Route Redistribution

**User Story:** As a network administrator, I want to redistribute routes into BGP, so that my internal routes are advertised.

**Acceptance Criteria:**
1. WHEN redistribute_static is true THEN the system SHALL configure `bgp import from static`
2. WHEN redistribute_connected is true THEN the system SHALL configure `bgp import from connected`
3. WHEN redistribution is disabled THEN the system SHALL configure `no bgp import from <type>`

---

### rtx_ospf Resource

#### Core Operations

##### Create
- Configure OSPF router ID (required)
- Configure OSPF areas with their types (normal, stub, nssa)
- Assign interfaces to OSPF areas
- Enable route redistribution if specified
- Enable OSPF service (`ospf use on`)
- Save configuration to persistent memory

##### Read
- Execute `show config | grep ospf` to retrieve current configuration
- Parse OSPF configuration including router ID, areas, networks
- Detect if OSPF is enabled or disabled
- If OSPF is disabled or not configured, remove resource from state

##### Update
- Compare current configuration with desired state
- Add/remove/modify areas as needed
- Update interface area assignments
- Update redistribution settings

##### Delete
- Disable OSPF service (`ospf use off`)
- Remove OSPF from operational state
- Save configuration to persistent memory

#### Feature Requirements

##### Requirement 1: Router ID Configuration

**User Story:** As a network administrator, I want to configure my OSPF router ID, so that my router is uniquely identified in the OSPF domain.

**Acceptance Criteria:**
1. WHEN a valid IPv4 address is provided as router_id THEN the system SHALL configure `ospf router id <ip>`
2. WHEN an invalid router_id is provided THEN the system SHALL reject with a validation error
3. IF router_id is not provided THEN the system SHALL reject with a validation error (required field)

##### Requirement 2: Area Configuration

**User Story:** As a network administrator, I want to configure OSPF areas with different types, so that I can implement hierarchical routing.

**Acceptance Criteria:**
1. WHEN an area with type "normal" is defined THEN the system SHALL configure `ospf area <id>`
2. WHEN an area with type "stub" is defined THEN the system SHALL configure `ospf area <id> stub`
3. WHEN an area with type "nssa" is defined THEN the system SHALL configure `ospf area <id> nssa`
4. WHEN no_summary is true for stub/nssa THEN the system SHALL append `no-summary`
5. WHEN area ID is provided as dotted decimal THEN the system SHALL accept it (e.g., "0.0.0.0")

##### Requirement 3: Interface Area Assignment

**User Story:** As a network administrator, I want to assign interfaces to OSPF areas, so that they participate in OSPF routing.

**Acceptance Criteria:**
1. WHEN a network entry with IP and area is defined THEN the system SHALL configure `ip <interface> ospf area <area>`
2. WHEN interface names like "lan1", "pp1" are used THEN the system SHALL accept them

##### Requirement 4: OSPF Route Redistribution

**User Story:** As a network administrator, I want to redistribute routes into OSPF, so that external routes are advertised.

**Acceptance Criteria:**
1. WHEN redistribute_static is true THEN the system SHALL configure `ospf import from static`
2. WHEN redistribute_connected is true THEN the system SHALL configure `ospf import from connected`
3. WHEN redistribution is disabled THEN the system SHALL configure `no ospf import from <type>`

---

### rtx_static_route Resource

#### Core Operations

##### Create
- Configure static route with destination prefix and mask
- Configure one or more next hops (gateway IP or interface)
- Set optional parameters (distance/weight, filter, permanent/keepalive)
- Save configuration to persistent memory

##### Read
- Execute `show config | grep "ip route <prefix>"` to retrieve route configuration
- Parse static route including all next hops
- If route doesn't exist, remove resource from state

##### Update
- Delete all existing next hops for the route
- Recreate all next hops with new configuration
- Save configuration to persistent memory

##### Delete
- Delete all next hops for the route
- Save configuration to persistent memory

#### Feature Requirements

##### Requirement 1: Basic Static Route

**User Story:** As a network administrator, I want to create static routes, so that traffic reaches destinations not learned dynamically.

**Acceptance Criteria:**
1. WHEN prefix and mask are provided THEN the system SHALL configure `ip route <prefix>/<cidr> gateway <gateway>`
2. WHEN prefix is "0.0.0.0" and mask is "0.0.0.0" THEN the system SHALL use `default` in the command
3. WHEN mask is in dotted decimal THEN the system SHALL convert to CIDR notation for commands

##### Requirement 2: Multiple Next Hops (ECMP/Failover)

**User Story:** As a network administrator, I want to configure multiple next hops for a route, so that I can implement load balancing or failover.

**Acceptance Criteria:**
1. WHEN multiple next_hops are defined THEN the system SHALL create separate route commands for each
2. WHEN distance/weight varies between hops THEN the system SHALL implement failover routing
3. WHEN distances are equal THEN the system SHALL enable load balancing (ECMP)

##### Requirement 3: Interface-Based Routing

**User Story:** As a network administrator, I want to route traffic through interfaces instead of IP gateways, so that I can use PPP or tunnel interfaces.

**Acceptance Criteria:**
1. WHEN interface is "pp 1" THEN the system SHALL configure `ip route <prefix> gateway pp 1`
2. WHEN interface is "tunnel 1" THEN the system SHALL configure `ip route <prefix> gateway tunnel 1`
3. WHEN interface is "dhcp <iface>" THEN the system SHALL configure `ip route <prefix> gateway dhcp <iface>`

##### Requirement 4: Route Options

**User Story:** As a network administrator, I want to configure route options like weight and keepalive, so that I can fine-tune routing behavior.

**Acceptance Criteria:**
1. WHEN distance > 1 THEN the system SHALL append `weight <distance>`
2. WHEN filter > 0 THEN the system SHALL append `filter <number>`
3. WHEN permanent is true THEN the system SHALL append `keepalive`

---

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each file handles one resource type
- **Modular Design**: Service layer separates router communication from Terraform logic
- **Dependency Management**: Parser layer has no internal dependencies
- **Clear Interfaces**: Client interface defines all CRUD methods

### Performance
- SSH connection establishment: < 5 seconds
- Command execution timeout: Configurable, default 30 seconds
- Batch command execution for multi-step operations

### Security
- Passwords (BGP neighbor passwords) marked as Sensitive in schema
- No credentials stored in Terraform state
- SSH encryption for all router communication

### Reliability
- Automatic retry with exponential backoff for transient failures
- Graceful handling of "not found" errors during delete
- Configuration saved to persistent memory after each operation

### Validation
- ASN validated as integer 1-65535 (2-byte ASN only, RTX limitation)
- IPv4 addresses validated with net.ParseIP
- OSPF area types validated: "normal", "stub", "nssa"
- Route distance validated: 1-100
- BGP hold-time: 3-28800
- BGP keepalive: 1-21845
- BGP multihop: 1-255
- OSPF neighbor priority: 0-255
- Subnet masks validated for contiguous 1s followed by 0s

---

## RTX Commands Reference

### BGP Commands

```
# Enable/Disable BGP
bgp use on
bgp use off

# Set AS number (1-65535, 2-byte ASN only)
bgp autonomous-system <asn>

# Set router ID
bgp router id <ip>

# Configure neighbor (inline options format)
bgp neighbor <n> <as> <ip>
bgp neighbor <n> <as> <ip> hold-time=<seconds> local-address=<ip> passive=on

# Configure neighbor password (pre-shared-key)
bgp neighbor pre-shared-key <n> text <password>

# Network announcements (CIDR notation)
bgp import filter <n> include <prefix>/<cidr>

# Redistribution
bgp import from static
bgp import from connected

# Delete commands
no bgp neighbor <n>
no bgp import filter <n>
no bgp import from static
no bgp import from connected

# Show configuration
show config | grep bgp
```

### OSPF Commands

```
# Enable/Disable OSPF
ospf use on
ospf use off

# Set router ID
ospf router id <ip>

# Configure areas
ospf area <id>
ospf area <id> stub
ospf area <id> stub no-summary
ospf area <id> nssa
ospf area <id> nssa no-summary

# Interface to area assignment
ip <interface> ospf area <area>

# Redistribution
ospf import from static
ospf import from connected

# Delete commands
no ospf area <id>
no ip <interface> ospf area
no ospf import from static
no ospf import from connected

# Show configuration
show config | grep ospf
```

### Static Route Commands

```
# Create routes
ip route default gateway <gateway>
ip route <prefix>/<cidr> gateway <gateway>
ip route <prefix>/<cidr> gateway <gateway> weight <n>
ip route <prefix>/<cidr> gateway <gateway> filter <n>
ip route <prefix>/<cidr> gateway <gateway> keepalive
ip route <prefix>/<cidr> gateway pp <n>
ip route <prefix>/<cidr> gateway tunnel <n>
ip route <prefix>/<cidr> gateway dhcp <interface>

# Delete routes
no ip route <prefix>/<cidr>
no ip route <prefix>/<cidr> gateway <gateway>

# Show configuration
show config | grep "ip route"
show config | grep "ip route <prefix>"
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Preview routing configuration changes |
| `terraform apply` | Required | Apply routing configuration to router |
| `terraform destroy` | Required | Remove routing configuration from router |
| `terraform import` | Required | Import existing routing configuration into state |
| `terraform refresh` | Required | Sync state with current router configuration |
| `terraform state` | Required | Manage routing resources in state file |

### Import Specification

#### rtx_bgp
- **Import ID Format**: `bgp` (literal string)
- **Import Command**: `terraform import rtx_bgp.main bgp`
- **Post-Import**: Review sensitive fields (passwords) not imported from router

#### rtx_ospf
- **Import ID Format**: `ospf` (literal string)
- **Import Command**: `terraform import rtx_ospf.main ospf`
- **Post-Import**: Review area and network configurations

#### rtx_static_route
- **Import ID Format**: `<prefix>/<mask>` (e.g., `10.0.0.0/255.0.0.0`, `0.0.0.0/0.0.0.0`)
- **Import Command**: `terraform import rtx_static_route.default "0.0.0.0/0.0.0.0"`
- **Post-Import**: All next hops for the route are imported

---

## Example Usage

### BGP Configuration

```hcl
resource "rtx_bgp" "main" {
  asn       = "65001"
  router_id = "10.0.0.1"

  neighbor {
    index     = 1
    ip        = "10.0.0.2"
    remote_as = "65002"
    hold_time = 90
    keepalive = 30
  }

  neighbor {
    index     = 2
    ip        = "10.0.0.3"
    remote_as = "65003"
    multihop  = 2
    password  = "secret123"
  }

  network {
    prefix = "192.168.1.0"
    mask   = "255.255.255.0"
  }

  network {
    prefix = "10.0.0.0"
    mask   = "255.0.0.0"
  }

  redistribute_static    = true
  redistribute_connected = true
}
```

### OSPF Configuration

```hcl
resource "rtx_ospf" "backbone" {
  router_id = "1.1.1.1"

  area {
    area_id = "0"
    type    = "normal"
  }

  area {
    area_id    = "1"
    type       = "stub"
    no_summary = true
  }

  network {
    ip       = "lan1"
    wildcard = "0.0.0.255"
    area     = "0"
  }

  network {
    ip       = "lan2"
    wildcard = "0.0.0.255"
    area     = "1"
  }

  redistribute_static    = true
  redistribute_connected = false
}
```

### Static Route Configuration

```hcl
# Default route with single gateway
resource "rtx_static_route" "default" {
  prefix = "0.0.0.0"
  mask   = "0.0.0.0"

  next_hop {
    gateway  = "192.168.1.1"
    distance = 1
  }
}

# Route with multiple next hops (failover)
resource "rtx_static_route" "private_networks" {
  prefix = "10.0.0.0"
  mask   = "255.0.0.0"

  next_hop {
    gateway   = "192.168.1.1"
    distance  = 1
    permanent = true
  }

  next_hop {
    gateway  = "192.168.2.1"
    distance = 10
  }
}

# Route via PPP interface
resource "rtx_static_route" "wan_route" {
  prefix = "172.16.0.0"
  mask   = "255.240.0.0"

  next_hop {
    interface = "pp 1"
    distance  = 1
  }
}

# Route via tunnel interface with filter
resource "rtx_static_route" "vpn_route" {
  prefix = "192.168.100.0"
  mask   = "255.255.255.0"

  next_hop {
    interface = "tunnel 1"
    distance  = 1
    filter    = 100
  }
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (neighbor state, route status) must not be stored in state
- BGP and OSPF "enabled" status is tracked through resource existence
- Password fields are marked Sensitive to prevent display in logs/output

---

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2025-01-23 | Initial | Created from implementation code analysis |
| 2026-02-01 | Implementation Audit | Fix attribute names: neighbor.id→index, area.id→area_id |
| 2026-02-07 | Implementation Audit | Full audit against implementation code |
| 2026-02-07 | RTX Reference Sync | BGP commands updated to match RTX reference: neighbor inline format, pre-shared-key, CIDR notation, 2-byte ASN (1-65535), hold-time 3-28800 |
