# Requirements Document: RTX PPP/PPPoE Resource

## Introduction

This feature implements Terraform resources for PPP (Point-to-Point Protocol) and PPPoE (PPP over Ethernet) configurations on Yamaha RTX routers. PPP/PPPoE is essential for WAN connectivity, particularly for ISP connections in Japan where PPPoE is the dominant connection method for fiber (FLET'S) and DSL services.

## Alignment with Product Vision

This feature supports the product vision by:
- Enabling Terraform management of WAN connectivity configurations
- Supporting the most common ISP connection method in Japan
- Completing the network connectivity stack (LAN, VLAN, PPPoE for WAN)
- Allowing Infrastructure-as-Code for complete router configuration

## Requirements

### REQ-1: PPPoE Client Configuration

**User Story:** As a network administrator, I want to configure PPPoE client settings via Terraform, so that I can manage ISP connections as code.

#### Acceptance Criteria

1. WHEN creating a PPPoE connection THEN the resource SHALL support:
   - PP interface selection (pp 1, pp 2, etc.)
   - Username and password authentication
   - Service name specification (optional)
   - AC name specification (optional)
2. WHEN configuring authentication THEN the resource SHALL support:
   - PAP authentication
   - CHAP authentication
   - MS-CHAP authentication
   - MS-CHAPv2 authentication
3. WHEN managing connection state THEN the resource SHALL support:
   - Auto-connect on boot
   - Manual connect/disconnect
   - Connection timeout settings
   - LCP echo configuration

### REQ-2: PP Interface IP Configuration

**User Story:** As a network administrator, I want to configure IP settings on PP interfaces, so that I can manage addressing for WAN connections.

#### Acceptance Criteria

1. WHEN configuring IP address THEN the resource SHALL support:
   - Dynamic IP (IPCP negotiation)
   - Static IP assignment
   - IP address with netmask
2. WHEN configuring DNS THEN the resource SHALL support:
   - Dynamic DNS from ISP (IPCP)
   - Static DNS server specification
3. WHEN configuring routing THEN the resource SHALL support:
   - Default route via PP interface
   - Route metric specification
   - Multiple default routes for failover

### REQ-3: PPP Keepalive and LCP

**User Story:** As a network administrator, I want to configure PPP keepalive settings, so that I can detect and recover from connection failures.

#### Acceptance Criteria

1. WHEN configuring LCP echo THEN the resource SHALL support:
   - Echo interval (seconds)
   - Echo failure count before disconnect
   - LCP restart timer
2. WHEN configuring reconnection THEN the resource SHALL support:
   - Auto-reconnect enable/disable
   - Reconnect delay
   - Maximum reconnect attempts

### REQ-4: PPPoE Session Management

**User Story:** As a network administrator, I want to manage multiple PPPoE sessions, so that I can configure multi-WAN setups.

#### Acceptance Criteria

1. WHEN managing multiple sessions THEN the resource SHALL support:
   - Multiple PP interfaces (pp 1-30)
   - Binding PP to LAN interface (pp bind lan2)
   - Session priority for load balancing
2. WHEN importing existing configurations THEN the resource SHALL:
   - Read current PPPoE configuration from router
   - Map to Terraform state correctly
   - Handle encrypted passwords appropriately

### REQ-5: PPP Compression and MTU

**User Story:** As a network administrator, I want to configure PPP compression and MTU settings, so that I can optimize connection performance.

#### Acceptance Criteria

1. WHEN configuring compression THEN the resource SHALL support:
   - Van Jacobson header compression
   - MPPC compression (if available)
   - Compression enable/disable
2. WHEN configuring MTU/MRU THEN the resource SHALL support:
   - MTU setting (typically 1454 for PPPoE)
   - MRU setting
   - TCP MSS adjustment

## Non-Functional Requirements

### Code Architecture and Modularity
- **Parser Module**: Create `internal/rtx/parsers/ppp.go` for command parsing
- **Client Module**: Create `internal/client/ppp_service.go` for API operations
- **Provider Resources**: Create separate resources for PPPoE and PP interface configuration
- **Pattern Catalog**: Create `internal/rtx/testdata/patterns/ppp.yaml`

### Performance
- PPPoE configuration changes SHALL complete within 10 seconds
- Connection state reads SHALL complete within 5 seconds
- Bulk operations on multiple PP interfaces SHALL be parallelizable

### Security
- Passwords SHALL be marked as sensitive in Terraform state
- Credentials SHALL not appear in plan output by default
- Encrypted password format from router SHALL be preserved during import

### Reliability
- Resource SHALL handle connection state changes gracefully
- Import SHALL work for existing PPPoE configurations
- Resource SHALL validate configuration before applying to router

### Usability
- Resource documentation SHALL include common ISP configuration examples
- Error messages SHALL clearly indicate authentication or connection failures
- Examples SHALL cover NTT FLET'S and common Japanese ISP patterns
