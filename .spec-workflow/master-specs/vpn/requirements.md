# Master Requirements: VPN Resources

## Overview

This document defines the requirements for VPN resources in the Terraform RTX Provider. VPN resources enable secure remote access and site-to-site connectivity through Yamaha RTX routers using IPsec, L2TP, and PPTP protocols.

## Alignment with Product Vision

VPN resources support enterprise-grade network infrastructure management by:
- Enabling Infrastructure-as-Code for VPN tunnel configurations
- Supporting multiple VPN protocols for diverse connectivity requirements
- Providing secure remote access for distributed workforces
- Facilitating site-to-site connectivity for branch office integration

## Resources Summary

| Resource | Type | Description |
|----------|------|-------------|
| `rtx_ipsec_tunnel` | Collection | IPsec VPN tunnel with IKEv2 |
| `rtx_ipsec_transport` | Collection | IPsec transport mode for L2TP/IPsec |
| `rtx_l2tp` | Collection | L2TP/L2TPv3 tunnel configuration |
| `rtx_l2tp_service` | Singleton | L2TP service enable/disable |
| `rtx_pptp` | Singleton | PPTP VPN server (legacy) |

---

## Resource 1: rtx_ipsec_tunnel

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_ipsec_tunnel` |
| Type | Collection |
| Import Support | Yes |
| Last Updated | 2026-01-23 |

### Functional Requirements

#### Create
- Create a new IPsec tunnel with specified ID
- Configure IKE Phase 1 (IKEv2) proposal with encryption, integrity, and DH group
- Configure IPsec Phase 2 (ESP/AH) transform settings
- Set local and remote endpoint addresses
- Configure pre-shared key authentication
- Optionally configure Dead Peer Detection (DPD)
- Optionally specify local and remote networks for tunnel mode

#### Read
- Retrieve IPsec tunnel configuration by tunnel ID
- Parse IKE and IPsec settings from router output
- Return all configurable attributes
- Note: Pre-shared key cannot be read back from router for security

#### Update
- Modify existing tunnel settings in-place
- Support updating encryption algorithms, integrity algorithms, DH groups
- Support updating endpoint addresses
- Support updating pre-shared key (if provided)
- Support updating DPD settings

#### Delete
- Remove IPsec tunnel configuration
- Clean up associated SA policies
- Remove tunnel select entry

### Requirement 1: IPsec Tunnel Creation

**User Story:** As a network administrator, I want to create IPsec tunnels, so that I can establish secure site-to-site VPN connections.

#### Acceptance Criteria

1. WHEN a valid IPsec tunnel configuration is provided THEN the system SHALL create the tunnel with specified settings
2. WHEN tunnel_id is not unique THEN the system SHALL return an error
3. WHEN encryption algorithm is specified THEN the system SHALL configure IKE Phase 1 accordingly
4. WHEN DPD is enabled THEN the system SHALL configure keepalive with specified interval and retry count

### Requirement 2: Pre-Shared Key Security

**User Story:** As a security administrator, I want pre-shared keys to be handled securely, so that credentials are not exposed.

#### Acceptance Criteria

1. WHEN pre_shared_key is provided THEN the system SHALL mark it as sensitive in Terraform state
2. WHEN reading tunnel configuration THEN the system SHALL NOT return the pre-shared key value
3. WHEN plan output is displayed THEN the system SHALL mask the pre-shared key value

### Attributes

| Attribute | Type | Required | ForceNew | Sensitive | Constraints | Description |
|-----------|------|----------|----------|-----------|-------------|-------------|
| tunnel_id | int | Yes | Yes | No | 1-65535 | Unique tunnel identifier |
| name | string | No | No | No | - | Tunnel description |
| local_address | string | No | No | No | Valid IPv4 | Local endpoint IP address |
| remote_address | string | No | No | No | Valid IPv4/hostname | Remote endpoint address |
| pre_shared_key | string | No | No | Yes | - | IKE pre-shared key |
| local_network | string | No | No | No | CIDR notation | Local network for tunnel mode |
| remote_network | string | No | No | No | CIDR notation | Remote network for tunnel mode |
| ikev2_proposal | block | No | No | No | MaxItems: 1 | IKE Phase 1 settings |
| ipsec_transform | block | No | No | No | MaxItems: 1 | IPsec Phase 2 settings |
| dpd_enabled | bool | No | No | No | - | Enable Dead Peer Detection |
| dpd_interval | int | No | No | No | >= 1 | DPD interval in seconds |
| dpd_retry | int | No | No | No | >= 0 | DPD retry count |
| enabled | bool | No | No | No | - | Enable the tunnel |
| tunnel_interface | string | No | No | No | - | Computed tunnel interface name (e.g., "tunnel1") |

#### ikev2_proposal Block

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| encryption_aes256 | bool | No | Use AES-256 encryption |
| encryption_aes128 | bool | No | Use AES-128 encryption |
| encryption_3des | bool | No | Use 3DES encryption |
| integrity_sha256 | bool | No | Use SHA-256 integrity |
| integrity_sha1 | bool | No | Use SHA-1 integrity |
| integrity_md5 | bool | No | Use MD5 integrity |
| group_fourteen | bool | No | DH Group 14 (2048-bit) |
| group_five | bool | No | DH Group 5 (1536-bit) |
| group_two | bool | No | DH Group 2 (1024-bit) |
| lifetime_seconds | int | No | IKE SA lifetime (>= 60) |

#### ipsec_transform Block

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| protocol | string | No | Protocol: "esp" or "ah" |
| encryption_aes256 | bool | No | Use AES-256 encryption |
| encryption_aes128 | bool | No | Use AES-128 encryption |
| encryption_3des | bool | No | Use 3DES encryption |
| integrity_sha256 | bool | No | Use SHA-256-HMAC |
| integrity_sha1 | bool | No | Use SHA-1-HMAC |
| integrity_md5 | bool | No | Use MD5-HMAC |
| pfs_group_fourteen | bool | No | PFS with DH Group 14 |
| pfs_group_five | bool | No | PFS with DH Group 5 |
| pfs_group_two | bool | No | PFS with DH Group 2 |
| lifetime_seconds | int | No | IPsec SA lifetime (>= 60) |

---

## Resource 2: rtx_ipsec_transport

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_ipsec_transport` |
| Type | Collection |
| Import Support | Yes |
| Last Updated | 2026-01-23 |

### Functional Requirements

#### Create
- Create IPsec transport mode configuration
- Associate with existing IPsec tunnel
- Configure protocol and port for transport mode

#### Read
- Retrieve IPsec transport configuration by transport ID
- Parse transport settings from router output

#### Update
- Modify transport settings (protocol, port, tunnel association)

#### Delete
- Remove IPsec transport configuration

### Requirement 1: L2TP/IPsec Transport Configuration

**User Story:** As a network administrator, I want to configure IPsec transport mode, so that I can use L2TP over IPsec for remote access VPN.

#### Acceptance Criteria

1. WHEN transport_id and tunnel_id are provided THEN the system SHALL create the transport mode configuration
2. WHEN protocol is "udp" and port is 1701 THEN the system SHALL configure standard L2TP/IPsec
3. WHEN transport_id is not unique THEN the system SHALL return an error

### Attributes

| Attribute | Type | Required | ForceNew | Constraints | Description |
|-----------|------|----------|----------|-------------|-------------|
| transport_id | int | Yes | Yes | 1-65535 | Unique transport identifier |
| tunnel_id | int | Yes | No | 1-65535 | Associated IPsec tunnel ID |
| protocol | string | No | No | "udp", "tcp" | Transport protocol |
| port | int | Yes | No | 1-65535 | Port number (typically 1701) |

---

## Resource 3: rtx_l2tp

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_l2tp` |
| Type | Collection |
| Import Support | Yes |
| Last Updated | 2026-01-23 |

### Functional Requirements

#### Create
- Create L2TPv2 (LNS) or L2TPv3 (L2VPN) tunnel
- Configure tunnel encapsulation type
- Set tunnel endpoints (source/destination)
- Configure authentication for L2TPv2
- Configure IP pool for L2TPv2 LNS
- Configure L2TPv3-specific settings (router IDs, session ID)

#### Read
- Retrieve L2TP tunnel configuration by tunnel ID
- Parse both L2TPv2 and L2TPv3 configurations
- Return version-specific attributes

#### Update
- Modify tunnel settings based on version
- Update endpoint addresses
- Update authentication/encryption settings

#### Delete
- Remove L2TP tunnel configuration
- Clean up associated PP bindings for LNS mode

### Requirement 1: L2TPv2 LNS (Remote Access VPN)

**User Story:** As a network administrator, I want to configure L2TPv2 LNS, so that remote users can connect via L2TP/IPsec.

#### Acceptance Criteria

1. WHEN version is "l2tp" and mode is "lns" THEN the system SHALL configure L2TPv2 LNS
2. WHEN authentication is configured THEN the system SHALL set the authentication method and credentials
3. WHEN ip_pool is configured THEN the system SHALL allocate IPs from the specified range

### Requirement 2: L2TPv3 L2VPN (Site-to-Site)

**User Story:** As a network administrator, I want to configure L2TPv3 L2VPN, so that I can extend Layer 2 networks across sites.

#### Acceptance Criteria

1. WHEN version is "l2tpv3" and mode is "l2vpn" THEN the system SHALL configure L2TPv3 tunnel
2. WHEN l2tpv3_config is provided THEN the system SHALL configure router IDs and session settings
3. WHEN tunnel_destination is an FQDN THEN the system SHALL configure dynamic DNS resolution

### Attributes

| Attribute | Type | Required | ForceNew | Sensitive | Constraints | Description |
|-----------|------|----------|----------|-----------|-------------|-------------|
| tunnel_id | int | Yes | Yes | No | 1-65535 | Unique tunnel identifier |
| name | string | No | No | No | - | Tunnel description |
| version | string | Yes | Yes | No | "l2tp", "l2tpv3" | L2TP version |
| mode | string | Yes | Yes | No | "lns", "l2vpn" | Operating mode |
| shutdown | bool | No | No | No | - | Admin shutdown state |
| tunnel_source | string | No | No | No | - | Source IP/interface |
| tunnel_destination | string | No | No | No | - | Destination IP/FQDN |
| tunnel_dest_type | string | No | No | No | "ip", "fqdn" | Destination type |
| authentication | block | No | No | No | MaxItems: 1 | L2TPv2 authentication |
| ip_pool | block | No | No | No | MaxItems: 1 | L2TPv2 IP pool |
| ipsec_profile | block | No | No | No | MaxItems: 1 | IPsec encryption |
| l2tpv3_config | block | No | No | No | MaxItems: 1 | L2TPv3 settings |
| keepalive_enabled | bool | No | No | No | - | Enable keepalive |
| keepalive_interval | int | No | No | No | >= 1 | Keepalive interval |
| keepalive_retry | int | No | No | No | >= 1 | Keepalive retry count |
| disconnect_time | int | No | No | No | >= 0 | Idle disconnect time |
| always_on | bool | No | No | No | - | Always-on mode |
| enabled | bool | No | No | No | - | Enable the tunnel |
| tunnel_interface | string | No | No | No | - | Computed tunnel interface name (e.g., "tunnel1") |

#### authentication Block

| Attribute | Type | Required | Sensitive | Description |
|-----------|------|----------|-----------|-------------|
| method | string | Yes | No | Auth method: pap, chap, mschap, mschap-v2 |
| username | string | No | No | Username for authentication |
| password | string | No | Yes | Password for authentication |

#### ip_pool Block

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| start | string | Yes | Start IP address |
| end | string | Yes | End IP address |

#### ipsec_profile Block

| Attribute | Type | Required | Sensitive | Description |
|-----------|------|----------|-----------|-------------|
| enabled | bool | No | No | Enable IPsec encryption |
| pre_shared_key | string | No | Yes | IPsec pre-shared key |
| tunnel_id | int | No | No | Associated IPsec tunnel ID |

#### l2tpv3_config Block

| Attribute | Type | Required | Sensitive | Constraints | Description |
|-----------|------|----------|-----------|-------------|-------------|
| local_router_id | string | No | No | Valid IPv4 | Local router ID |
| remote_router_id | string | No | No | Valid IPv4 | Remote router ID |
| remote_end_id | string | No | No | - | Remote end ID (hostname) |
| session_id | int | No | No | - | Session ID |
| cookie_size | int | No | No | 0, 4, 8 | Cookie size in bytes |
| bridge_interface | string | No | No | - | Bridge interface for L2VPN |
| tunnel_auth_enabled | bool | No | No | - | Enable tunnel authentication |
| tunnel_auth_password | string | No | Yes | - | Tunnel auth password |

---

## Resource 4: rtx_l2tp_service

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_l2tp_service` |
| Type | Singleton |
| Import Support | Yes |
| Import ID | "default" |
| Last Updated | 2026-01-23 |

### Functional Requirements

#### Create
- Enable L2TP service on the router
- Optionally specify which protocols to enable (l2tp, l2tpv3)

#### Read
- Retrieve L2TP service state
- Return enabled protocols

#### Update
- Toggle service enabled state
- Modify enabled protocols

#### Delete
- Disable L2TP service

### Requirement 1: L2TP Service Management

**User Story:** As a network administrator, I want to manage the L2TP service state, so that L2TP tunnels can function.

#### Acceptance Criteria

1. WHEN enabled is true THEN the system SHALL enable the L2TP service
2. WHEN protocols are specified THEN the system SHALL enable only those protocols
3. WHEN resource is deleted THEN the system SHALL disable the L2TP service

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| enabled | bool | Yes | Enable/disable L2TP service |
| protocols | list(string) | No | Enabled protocols: "l2tp", "l2tpv3" |

---

## Resource 5: rtx_pptp

### Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_pptp` |
| Type | Singleton |
| Import Support | Yes |
| Import ID | "pptp" |
| Last Updated | 2026-01-23 |

### Functional Requirements

#### Create
- Enable PPTP VPN server
- Configure authentication method and credentials
- Configure MPPE encryption settings
- Configure IP pool for PPTP clients

#### Read
- Retrieve PPTP server configuration
- Return authentication and encryption settings

#### Update
- Modify authentication, encryption, or IP pool settings

#### Delete
- Disable PPTP service

### Requirement 1: PPTP Server Configuration

**User Story:** As a network administrator, I want to configure PPTP VPN server, so that legacy clients can connect remotely.

#### Acceptance Criteria

1. WHEN authentication is configured THEN the system SHALL set the method and credentials
2. WHEN encryption is configured THEN the system SHALL set MPPE encryption strength
3. WHEN resource is deleted THEN the system SHALL disable PPTP service

### Security Warning

PPTP is considered insecure due to known vulnerabilities in its authentication and encryption protocols. Consider using L2TP/IPsec or IKEv2 instead for better security.

### Attributes

| Attribute | Type | Required | Sensitive | Constraints | Description |
|-----------|------|----------|-----------|-------------|-------------|
| shutdown | bool | No | No | - | Admin shutdown state |
| listen_address | string | No | No | Valid IPv4 | Listen IP address |
| max_connections | int | No | No | >= 0 | Maximum connections (0=unlimited) |
| authentication | block | Yes | No | MaxItems: 1 | PPTP authentication |
| encryption | block | No | No | MaxItems: 1 | MPPE encryption settings |
| ip_pool | block | No | No | MaxItems: 1 | IP pool for clients |
| disconnect_time | int | No | No | >= 0 | Idle disconnect time |
| keepalive_enabled | bool | No | No | - | Enable keepalive |
| enabled | bool | No | No | - | Enable PPTP service |

#### authentication Block

| Attribute | Type | Required | Sensitive | Constraints | Description |
|-----------|------|----------|-----------|-------------|-------------|
| method | string | Yes | No | pap, chap, mschap, mschap-v2 | Auth method |
| username | string | No | No | - | Username |
| password | string | No | Yes | - | Password |

#### encryption Block

| Attribute | Type | Required | Constraints | Description |
|-----------|------|----------|-------------|-------------|
| mppe_bits | int | No | 40, 56, 128 | MPPE encryption strength |
| required | bool | No | - | Require encryption |

#### ip_pool Block

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| start | string | Yes | Start IP address |
| end | string | Yes | End IP address |

---

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each VPN resource has dedicated service, parser, and resource files
- **Modular Design**: Parser layer is separate from service layer
- **Clear Interfaces**: Client interface defines all VPN methods

### Performance
- Batch command execution for IPsec tunnel creation to minimize SSH round-trips
- Configuration saved after batch operations complete
- Efficient parsing using compiled regular expressions

### Security
- Pre-shared keys marked as sensitive in Terraform schema
- Passwords not read back from router for security
- PPTP resource includes security warning in documentation
- Support for strong encryption algorithms (AES-256, SHA-256)

### Reliability
- Graceful handling of "not found" errors during read operations
- Automatic resource removal from state when tunnel is deleted externally
- Configuration validation before executing commands
- Error detection in command output

### Validation
- Tunnel/Transport IDs must be between 1 and 65535
- IP addresses validated as proper IPv4 format
- CIDR notation validated for network specifications
- Protocol values validated against allowed options
- DPD interval validated within reasonable range

---

## RTX Commands Reference

### IPsec Tunnel Commands

```
# Tunnel selection
tunnel select <n>

# IPsec tunnel association
ipsec tunnel <n>

# IKE settings
ipsec ike local address <n> <ip>
ipsec ike remote address <n> <ip>
ipsec ike pre-shared-key <n> text <key>
ipsec ike encryption <n> aes-cbc-256|aes-cbc|3des-cbc
ipsec ike hash <n> sha256|sha|md5
ipsec ike group <n> modp2048|modp1536|modp1024

# SA policy
ipsec sa policy <policy> <tunnel> esp|ah <enc> <hash>

# DPD (Dead Peer Detection)
ipsec ike keepalive use <n> on dpd <interval> [retry]
ipsec ike keepalive use <n> off

# Delete
no ipsec tunnel <n>
no tunnel select <n>
```

### IPsec Transport Commands

```
# Create transport mode
ipsec transport <transport_id> <tunnel_id> <protocol> <port>

# Delete
no ipsec transport <transport_id>
```

### L2TP Commands

```
# L2TP service
l2tp service on [l2tpv3] [l2tp]
l2tp service off

# L2TPv2 LNS
pp select anonymous
pp bind tunnel<n>
pp auth accept <method>
pp auth myname <user> <pass>
ip pp remote address pool <start>-<end>

# L2TPv3 L2VPN
tunnel select <n>
tunnel encapsulation l2tpv3
tunnel endpoint address <local> <remote>
l2tp local router-id <ip>
l2tp remote router-id <ip>
l2tp always-on on|off
l2tp keepalive use on <interval> <retry>
l2tp tunnel disconnect time <seconds>

# Delete
no tunnel select <n>
```

### PPTP Commands

```
# PPTP service
pptp service on|off
pptp tunnel disconnect time <seconds>
pptp keepalive use on|off

# Authentication
pp auth accept <method>
pp auth myname <user> <pass>

# MPPE encryption
ppp ccp type mppe-128|mppe-56|mppe-40|mppe-any [require]

# IP pool
ip pp remote address pool <start>-<end>
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Preview VPN configuration changes |
| `terraform apply` | Required | Apply VPN configuration to router |
| `terraform destroy` | Required | Remove VPN configurations |
| `terraform import` | Required | Import existing VPN configurations |
| `terraform refresh` | Required | Sync state with router configuration |
| `terraform state` | Required | Manage VPN resources in state |

### Import Specifications

#### rtx_ipsec_tunnel
- **Import ID Format**: `<tunnel_id>` (integer)
- **Import Command**: `terraform import rtx_ipsec_tunnel.example 1`
- **Post-Import**: Pre-shared key must be provided in configuration

#### rtx_ipsec_transport
- **Import ID Format**: `<transport_id>` (integer)
- **Import Command**: `terraform import rtx_ipsec_transport.example 1`

#### rtx_l2tp
- **Import ID Format**: `<tunnel_id>` (integer)
- **Import Command**: `terraform import rtx_l2tp.example 1`
- **Post-Import**: Sensitive fields may need to be provided

#### rtx_l2tp_service
- **Import ID Format**: `default`
- **Import Command**: `terraform import rtx_l2tp_service.main default`

#### rtx_pptp
- **Import ID Format**: `pptp`
- **Import Command**: `terraform import rtx_pptp.main pptp`
- **Post-Import**: Password must be provided in configuration

---

## Example Usage

### IPsec Site-to-Site VPN

```hcl
resource "rtx_ipsec_tunnel" "site_to_site" {
  tunnel_id      = 1
  name           = "Tokyo to Osaka"
  local_address  = "203.0.113.1"
  remote_address = "198.51.100.1"
  pre_shared_key = "supersecretkey"
  local_network  = "192.168.1.0/24"
  remote_network = "192.168.2.0/24"

  ikev2_proposal {
    encryption_aes256 = true
    integrity_sha256  = true
    group_fourteen    = true
    lifetime_seconds  = 28800
  }

  ipsec_transform {
    protocol          = "esp"
    encryption_aes256 = true
    integrity_sha256  = true
    pfs_group_fourteen = true
    lifetime_seconds  = 3600
  }

  dpd_enabled  = true
  dpd_interval = 30
  dpd_retry    = 5
  enabled      = true
}
```

### L2TP/IPsec Remote Access

```hcl
resource "rtx_l2tp_service" "main" {
  enabled   = true
  protocols = ["l2tp"]
}

resource "rtx_ipsec_tunnel" "l2tp_ipsec" {
  tunnel_id      = 2
  local_address  = "203.0.113.1"
  pre_shared_key = "remotevpnkey"

  ikev2_proposal {
    encryption_aes256 = true
    integrity_sha256  = true
    group_fourteen    = true
  }
}

resource "rtx_ipsec_transport" "l2tp" {
  transport_id = 1
  tunnel_id    = rtx_ipsec_tunnel.l2tp_ipsec.tunnel_id
  protocol     = "udp"
  port         = 1701
}

resource "rtx_l2tp" "remote_access" {
  tunnel_id = 2
  version   = "l2tp"
  mode      = "lns"

  authentication {
    method   = "mschap-v2"
    username = "vpnuser"
    password = "vpnpassword"
  }

  ip_pool {
    start = "192.168.100.10"
    end   = "192.168.100.50"
  }

  ipsec_profile {
    enabled   = true
    tunnel_id = rtx_ipsec_tunnel.l2tp_ipsec.tunnel_id
  }

  enabled = true
}
```

### L2TPv3 Site-to-Site L2VPN

```hcl
resource "rtx_l2tp_service" "main" {
  enabled   = true
  protocols = ["l2tpv3"]
}

resource "rtx_l2tp" "l2vpn" {
  tunnel_id          = 3
  name               = "Branch L2VPN"
  version            = "l2tpv3"
  mode               = "l2vpn"
  tunnel_source      = "203.0.113.1"
  tunnel_destination = "198.51.100.1"

  l2tpv3_config {
    local_router_id  = "1.1.1.1"
    remote_router_id = "2.2.2.2"
    session_id       = 1
    cookie_size      = 8
  }

  keepalive_enabled  = true
  keepalive_interval = 30
  keepalive_retry    = 5
  always_on          = true
  enabled            = true
}
```

### PPTP Server (Legacy)

```hcl
resource "rtx_pptp" "main" {
  authentication {
    method   = "mschap-v2"
    username = "pptpuser"
    password = "pptppassword"
  }

  encryption {
    mppe_bits = 128
    required  = true
  }

  ip_pool {
    start = "192.168.200.10"
    end   = "192.168.200.30"
  }

  disconnect_time   = 3600
  keepalive_enabled = true
  enabled           = true
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (connection state, SA state) must not be stored
- Pre-shared keys and passwords stored as sensitive values
- Import operations require post-import configuration for sensitive fields

---

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation | Initial documentation from codebase |
| 2026-01-25 | Implementation Sync | Add computed `tunnel_interface` for rtx_ipsec_tunnel and rtx_l2tp |
