# Requirements Document: RTX DDNS Resource

## Introduction

This feature implements Terraform resources for Dynamic DNS (DDNS) configurations on Yamaha RTX routers. DDNS enables remote access to networks with dynamic IP addresses by automatically updating DNS records when the WAN IP changes, which is essential for VPN endpoints, remote management, and hosted services.

## Alignment with Product Vision

This feature supports the product vision by:
- Enabling Terraform management of DDNS update configurations
- Supporting remote access scenarios common in enterprise deployments
- Complementing VPN resources (IPsec, L2TP, PPTP) that benefit from stable DNS names
- Allowing complete network infrastructure management as code

## Requirements

### REQ-1: DDNS Service Configuration

**User Story:** As a network administrator, I want to configure DDNS services via Terraform, so that I can manage dynamic DNS updates as code.

#### Acceptance Criteria

1. WHEN creating a DDNS configuration THEN the resource SHALL support:
   - Multiple DDNS providers (built-in Yamaha NetVolante DNS, custom providers)
   - Hostname specification
   - Update interval configuration
   - Interface binding for IP detection
2. WHEN configuring NetVolante DNS THEN the resource SHALL support:
   - netvolante-dns hostname command
   - netvolante-dns go command
   - netvolante-dns auto hostname command
   - Server selection (server 1, server 2)
3. WHEN configuring custom DDNS THEN the resource SHALL support:
   - HTTP/HTTPS update URL
   - Authentication (username/password or API key)
   - Custom update parameters

### REQ-2: DDNS Update Triggers

**User Story:** As a network administrator, I want to configure when DDNS updates occur, so that DNS records stay current without excessive updates.

#### Acceptance Criteria

1. WHEN configuring update triggers THEN the resource SHALL support:
   - IP address change detection
   - Periodic update interval (minutes)
   - Manual update trigger
   - Update on interface up/down
2. WHEN configuring update conditions THEN the resource SHALL support:
   - Minimum interval between updates
   - Update retry on failure
   - Update retry count and delay

### REQ-3: NetVolante DNS Specific Features

**User Story:** As a network administrator, I want to use Yamaha's NetVolante DNS service, so that I can leverage integrated DDNS functionality.

#### Acceptance Criteria

1. WHEN using NetVolante DNS THEN the resource SHALL support:
   - Auto hostname generation (MAC-based)
   - Custom hostname specification
   - Multiple hostname registration
   - IPv4 and IPv6 address registration
2. WHEN managing NetVolante DNS THEN the resource SHALL support:
   - Hostname availability check
   - Current registration status
   - Registration timeout settings

### REQ-4: DDNS Status and Monitoring

**User Story:** As a network administrator, I want to monitor DDNS update status, so that I can ensure DNS records are current.

#### Acceptance Criteria

1. WHEN reading DDNS status THEN the data source SHALL provide:
   - Last update timestamp
   - Last update result (success/failure)
   - Current registered IP address
   - Next scheduled update time
2. WHEN update fails THEN the resource SHALL:
   - Report detailed error information
   - Indicate retry status
   - Provide troubleshooting guidance

### REQ-5: Multi-WAN DDNS Support

**User Story:** As a network administrator, I want to configure DDNS for multiple WAN interfaces, so that I can support failover scenarios.

#### Acceptance Criteria

1. WHEN configuring multi-WAN DDNS THEN the resource SHALL support:
   - Binding DDNS to specific interface (pp, lan)
   - Priority-based interface selection
   - Separate hostnames per interface
2. WHEN WAN failover occurs THEN the resource SHALL support:
   - Automatic DNS update on failover
   - Update delay configuration
   - Fallback interface specification

## Non-Functional Requirements

### Code Architecture and Modularity
- **Parser Module**: Create `internal/rtx/parsers/ddns.go` for command parsing
- **Client Module**: Create `internal/client/ddns_service.go` for API operations
- **Provider Resources**: Create `rtx_ddns` and `rtx_netvolante_dns` resources
- **Pattern Catalog**: Create `internal/rtx/testdata/patterns/ddns.yaml`
- **Data Source**: Create `rtx_ddns_status` data source for status monitoring

### Performance
- DDNS configuration changes SHALL complete within 5 seconds
- Status reads SHALL complete within 3 seconds
- Manual update trigger SHALL complete within 30 seconds

### Security
- API keys and passwords SHALL be marked as sensitive
- Credentials SHALL not appear in logs or plan output
- HTTPS SHALL be preferred for custom DDNS providers

### Reliability
- Resource SHALL handle network connectivity issues gracefully
- Import SHALL work for existing DDNS configurations
- Resource SHALL validate hostname format before applying

### Usability
- Resource documentation SHALL include examples for common DDNS providers
- Error messages SHALL clearly indicate authentication or DNS errors
- Examples SHALL cover NetVolante DNS and popular third-party services
