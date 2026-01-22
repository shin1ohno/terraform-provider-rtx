# Master Requirements: PPP/PPPoE Resources

## Overview

PPPoE (Point-to-Point Protocol over Ethernet) resources enable Infrastructure as Code management of WAN connectivity on Yamaha RTX routers. PPPoE is the primary method for connecting RTX routers to Japanese ISPs (NTT FLET'S, au Hikari, etc.) and enterprise networks. This resource provides declarative configuration of PPPoE sessions including authentication, keepalive settings, and IP address assignment.

## Alignment with Product Vision

This resource directly supports the provider's key features:

- **Interfaces & Connectivity**: PPP/PPPoE is explicitly listed as a core connectivity feature in product.md
- **Cisco-Compatible Syntax**: Resource and attribute naming follows Cisco IOS XE conventions where applicable
- **Japanese Market Focus**: PPPoE is ubiquitous in Japan for ISP connectivity (FLET'S, etc.)
- **IaC for Network Infrastructure**: Enables version-controlled, reproducible WAN configurations

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_pppoe` |
| Type | Collection (keyed by PP number) |
| Import Support | Yes |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation-based (primary source) |

## Functional Requirements

### Core Operations

#### Create

- Configure a new PPPoE connection on a PP interface
- Execute RTX commands to set up PPPoE session parameters
- Bind physical interface (lan2, lan3, etc.) to PP interface
- Configure authentication credentials (username/password)
- Set connection behavior (always-on, disconnect timeout)
- Enable the PP interface after configuration
- Save configuration to persistent memory

#### Read

- Retrieve PPPoE configuration from router via `show config`
- Parse PP interface settings including description, interface binding, authentication
- Parse IP configuration (address, MTU, NAT descriptor)
- Return current enabled/disabled state
- Handle "not found" case by clearing resource from state

#### Update

- Modify existing PPPoE configuration in place
- Select the target PP interface before making changes
- Update individual settings (description, authentication, always-on, etc.)
- Enable/disable PP interface based on `enabled` attribute
- Save configuration after updates

#### Delete

- Disable the PP interface first
- Remove all PPPoE-related configuration commands
- Clear description, authentication, interface binding
- Reset IP configuration to defaults
- Save configuration after deletion

### Feature Requirements

### Requirement 1: Basic PPPoE Connection

**User Story:** As a network administrator, I want to configure PPPoE connections for WAN access, so that my RTX router can connect to my ISP.

#### Acceptance Criteria

1. WHEN a user specifies `pp_number`, `bind_interface`, `username`, and `password` THEN the system SHALL create a functional PPPoE connection
2. IF `always_on` is true (default) THEN the system SHALL configure the connection to automatically reconnect
3. IF `enabled` is true (default) THEN the system SHALL enable the PP interface after configuration

### Requirement 2: Authentication Method Selection

**User Story:** As a network administrator, I want to specify the PPP authentication method, so that I can meet my ISP's authentication requirements.

#### Acceptance Criteria

1. WHEN `auth_method` is set to "chap" (default) THEN the system SHALL configure `pp auth accept chap`
2. WHEN `auth_method` is set to "pap" THEN the system SHALL configure `pp auth accept pap`
3. WHEN `auth_method` is set to "mschap" or "mschap-v2" THEN the system SHALL configure the corresponding authentication method
4. IF an invalid auth_method is provided THEN the system SHALL return a validation error

### Requirement 3: Service Name and AC Name Filtering

**User Story:** As a network administrator, I want to filter PPPoE sessions by service name or AC name, so that I can connect to the correct service when multiple services are available.

#### Acceptance Criteria

1. IF `service_name` is specified THEN the system SHALL configure `pppoe service-name <value>`
2. IF `ac_name` is specified THEN the system SHALL configure `pppoe ac-name <value>`
3. IF neither is specified THEN the system SHALL accept any service/AC offered by the provider

### Requirement 4: Connection Behavior Control

**User Story:** As a network administrator, I want to control connection persistence and timeout behavior, so that I can optimize bandwidth costs or ensure always-on connectivity.

#### Acceptance Criteria

1. WHEN `always_on` is true THEN the system SHALL configure `pp always-on on`
2. WHEN `always_on` is false THEN the system SHALL configure `pp always-on off`
3. WHEN `disconnect_timeout` is greater than 0 THEN the system SHALL configure `pp disconnect time <seconds>`
4. WHEN `disconnect_timeout` is 0 THEN the system SHALL configure `pp disconnect time off`

### Requirement 5: Reconnection Settings

**User Story:** As a network administrator, I want to configure reconnection behavior, so that the connection recovers gracefully from failures.

#### Acceptance Criteria

1. WHEN `reconnect_interval` is specified THEN the system SHALL configure keepalive interval
2. WHEN `reconnect_attempts` is specified THEN the system SHALL configure retry attempts
3. IF `reconnect_interval` is 0 or unset THEN the system SHALL not configure reconnection parameters

### Requirement 6: Import Existing Configuration

**User Story:** As a network administrator, I want to import existing PPPoE configurations into Terraform state, so that I can manage pre-existing connections with IaC.

#### Acceptance Criteria

1. WHEN importing with PP number THEN the system SHALL retrieve and populate all configuration attributes
2. IF the imported configuration has an encrypted password THEN the system SHALL set password to empty string (user must provide in HCL)
3. IF the PP interface does not exist THEN the system SHALL return an import error

### Requirement 7: Sensitive Data Handling

**User Story:** As a security-conscious administrator, I want passwords to be handled securely, so that credentials are not exposed in logs or state files.

#### Acceptance Criteria

1. WHEN password is set in configuration THEN it SHALL be marked as sensitive in the schema
2. WHEN reading configuration THEN the system SHALL NOT attempt to read the password from router (passwords are encrypted)
3. WHEN logging commands THEN the system SHALL NOT log the password value

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Resource file handles Terraform lifecycle; service handles RTX communication; parser handles CLI output parsing
- **Modular Design**: PPPService isolated in `ppp_service.go`; parser in `ppp.go`
- **Dependency Management**: Parser has no internal dependencies; service depends only on parser and executor
- **Clear Interfaces**: `Executor` interface abstracts command execution for testability

### Performance

- SSH connection establishment: < 5 seconds
- PPPoE configuration commands: Executed sequentially (typically 5-15 commands)
- Configuration save: Triggered once after all commands complete
- Read operation: Single `show config` command parsed locally

### Security

- Password marked as `Sensitive: true` in Terraform schema
- Password not read back from router (encrypted in config)
- Password not logged in debug output
- SSH encryption for all router communication

### Reliability

- Context cancellation checked before command execution
- "Not found" errors handled gracefully (resource removed from state)
- Configuration saved after successful create/update/delete
- Delete operation continues even if some commands fail (cleanup best-effort)

### Validation

| Attribute | Constraint | Error Message |
|-----------|------------|---------------|
| `pp_number` | >= 1 | "PP number must be >= 1" |
| `bind_interface` | Required, valid interface name | "either interface or bind_interface must be specified" |
| `auth_method` | One of: pap, chap, mschap, mschap-v2 | "invalid authentication method" |
| `username` | Required when auth configured | N/A (Terraform required attribute) |
| `password` | Required when username set | "password is required when username is specified" |
| `disconnect_timeout` | >= 0 | Terraform validation |
| `reconnect_interval` | >= 0 | "reconnect interval must be >= 0" |
| `reconnect_attempts` | >= 0 | "reconnect attempts must be >= 0" |

## RTX Commands Reference

### Configuration Commands

```
pp select <num>                           # Select PP interface context
description pp <text>                     # Set description/name
pppoe use <interface>                     # Bind physical interface for PPPoE
pp bind <interface>                       # Alternative interface binding
pppoe service-name <name>                 # Filter by service name
pppoe ac-name <name>                      # Filter by AC name
pp auth accept <method>                   # Set authentication method
pp auth myname <username> <password>      # Set credentials
pp always-on on|off                       # Connection persistence
pp disconnect time <seconds>|off          # Idle timeout
pp keepalive interval <sec> retry-interval <count>  # Reconnection settings
pp enable <num>                           # Enable PP interface
pp disable <num>                          # Disable PP interface
```

### Delete Commands

```
pp disable <num>                          # Disable first
pp select <num>                           # Select context
no description                            # Remove description
no pppoe use                              # Remove interface binding
no pp bind                                # Remove alternative binding
no pppoe service-name                     # Remove service filter
no pp auth accept                         # Remove auth method
no pp auth myname                         # Remove credentials
pp always-on off                          # Reset always-on
no ip pp address                          # Remove IP config
no ip pp mtu                              # Remove MTU
no ip pp nat descriptor                   # Remove NAT binding
no ip pp secure filter in                 # Remove inbound filters
no ip pp secure filter out                # Remove outbound filters
```

### Status Commands

```
show config                               # Full configuration (for parsing)
show status pp <num>                      # Connection status
```

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Shows planned PPPoE configuration changes |
| `terraform apply` | Required | Creates/updates PPPoE configuration on router |
| `terraform destroy` | Required | Removes PPPoE configuration from router |
| `terraform import` | Required | Imports existing PP interface into state |
| `terraform refresh` | Required | Reads current PPPoE config from router |
| `terraform state` | Required | Manages PPPoE resource in state file |

### Import Specification

- **Import ID Format**: `<pp_number>` (integer, e.g., "1", "2")
- **Import Command**: `terraform import rtx_pppoe.main 1`
- **Post-Import Requirements**:
  - User must provide `password` in HCL (cannot be read from router)
  - Run `terraform plan` to verify configuration matches

## Example Usage

### Basic ISP Connection

```hcl
resource "rtx_pppoe" "flets" {
  pp_number      = 1
  name           = "NTT FLET'S NGN"
  bind_interface = "lan2"
  username       = "user@example.ne.jp"
  password       = var.isp_password
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}
```

### Backup Connection with Timeout

```hcl
resource "rtx_pppoe" "backup" {
  pp_number          = 2
  name               = "Backup ISP"
  bind_interface     = "lan3"
  username           = "backup@provider.jp"
  password           = var.backup_password
  auth_method        = "pap"
  always_on          = false
  disconnect_timeout = 300
  enabled            = true
}
```

### With Service Name Filter

```hcl
resource "rtx_pppoe" "business" {
  pp_number      = 1
  bind_interface = "lan2"
  username       = "corp@business.ne.jp"
  password       = var.corp_password
  service_name   = "FLET'S HIKARI NEXT"
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}
```

### With Reconnection Settings

```hcl
resource "rtx_pppoe" "resilient" {
  pp_number          = 1
  bind_interface     = "lan2"
  username           = "user@isp.jp"
  password           = var.password
  auth_method        = "chap"
  always_on          = true
  reconnect_interval = 10
  reconnect_attempts = 3
  enabled            = true
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state
- Connection status (connected/disconnected) is NOT stored in state
- Password is stored in state (marked sensitive) but not read from router
- Changes to operational state do not trigger resource updates

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation-based | Initial master spec created from existing implementation |
