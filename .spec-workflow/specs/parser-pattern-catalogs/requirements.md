# Requirements Document: Parser Pattern Catalogs

## Introduction

This feature creates YAML pattern catalogs for parsers that currently lack explicit documentation, and adds test coverage for the IPsec transport mode parser. The goal is to achieve 100% documentation parity across all RTX command parsers, establishing a consistent reference for command formats and validation rules.

## Alignment with Product Vision

This feature supports the product vision by:
- Ensuring consistent documentation across all parsers
- Enabling automated test generation from pattern catalogs
- Improving maintainability through structured command documentation
- Supporting future parser enhancements with clear specifications

## Requirements

### REQ-1: IP Filter Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for IP filter commands, so that all filter command formats are documented and testable.

#### Acceptance Criteria

1. WHEN creating ip_filter.yaml THEN the catalog SHALL document:
   - ip filter command with all parameter variations
   - ip filter set command
   - ip filter dynamic command
   - no-form commands for deletion
2. WHEN documenting parameters THEN each parameter SHALL include:
   - Type (int, string, enum, ipv4, etc.)
   - Valid range or values
   - Default value if applicable
3. WHEN providing examples THEN each command variant SHALL have:
   - At least 2 working examples from RTX documentation
   - Edge case examples for boundary values

### REQ-2: Ethernet Filter Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for Ethernet filter commands, so that all MAC-based filter formats are documented.

#### Acceptance Criteria

1. WHEN creating ethernet_filter.yaml THEN the catalog SHALL document:
   - ethernet filter command with all parameter variations
   - MAC address matching patterns
   - VLAN ID filtering
   - EtherType filtering
2. WHEN documenting MAC patterns THEN the catalog SHALL include:
   - Single MAC address format (xx:xx:xx:xx:xx:xx)
   - MAC address range format
   - Wildcard patterns

### REQ-3: Bridge Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for bridge commands, so that bridge interface configurations are documented.

#### Acceptance Criteria

1. WHEN creating bridge.yaml THEN the catalog SHALL document:
   - bridge member command for interface grouping
   - bridge group command for L2VPN
   - ip bridge address command
   - no-form commands for deletion
2. WHEN documenting bridge members THEN the catalog SHALL include:
   - Valid interface types (lan, vlan, pp, tunnel)
   - Member limit constraints

### REQ-4: Service Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for service commands, so that HTTPD/SSHD/SFTPD configurations are documented.

#### Acceptance Criteria

1. WHEN creating service.yaml THEN the catalog SHALL document:
   - httpd host command
   - httpd listen command
   - sshd host command
   - sshd host key command
   - sftpd host command
2. WHEN documenting host access THEN the catalog SHALL include:
   - IP address formats
   - Network CIDR formats
   - Wildcard (*) patterns

### REQ-5: System Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for system commands, so that core system configurations are documented.

#### Acceptance Criteria

1. WHEN creating system.yaml THEN the catalog SHALL document:
   - timezone command
   - console character command
   - console speed command
   - statistics command
   - packet-buffer command
2. WHEN documenting timezone THEN the catalog SHALL include:
   - UTC offset formats (+09:00, -05:00)
   - Named timezone support

### REQ-6: IPv6 Interface Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for IPv6 interface commands, so that IPv6 address and RTADV configurations are documented.

#### Acceptance Criteria

1. WHEN creating ipv6_interface.yaml THEN the catalog SHALL document:
   - ipv6 address command with all variations
   - rtadv send command
   - dhcp service client command
   - ipv6 mtu command
   - ipv6 secure filter command
2. WHEN documenting IPv6 addresses THEN the catalog SHALL include:
   - Full address format
   - Compressed address format
   - Link-local addresses
   - Prefix delegation references

### REQ-7: IPv6 Prefix Pattern Catalog

**User Story:** As a provider developer, I want a YAML pattern catalog for IPv6 prefix commands, so that prefix delegation and RA configurations are documented.

#### Acceptance Criteria

1. WHEN creating ipv6_prefix.yaml THEN the catalog SHALL document:
   - ipv6 prefix command for static prefixes
   - ipv6 prefix ra-prefix command
   - ipv6 prefix dhcp-prefix command
2. WHEN documenting prefix types THEN the catalog SHALL include:
   - Static prefix assignments
   - RA-derived prefixes
   - DHCPv6-PD prefixes

### REQ-8: IPsec Transport Mode Tests

**User Story:** As a provider developer, I want test coverage for IPsec transport mode parser, so that transport mode configurations can be validated.

#### Acceptance Criteria

1. WHEN testing ipsec_transport.go THEN tests SHALL cover:
   - Transport mode enable/disable
   - SA configuration parsing
   - Key lifetime settings
   - Peer address configuration
2. WHEN creating test file THEN it SHALL follow:
   - Existing test patterns from ipsec_tunnel_test.go
   - Table-driven test structure
   - Pattern catalog integration

## Non-Functional Requirements

### Code Architecture and Modularity
- **Schema Compliance**: All catalogs SHALL follow the schema defined in schema.yaml
- **Consistent Format**: All catalogs SHALL use identical YAML structure
- **Documentation References**: Each pattern SHALL reference RTX documentation chapter/section

### Performance
- Pattern catalog loading SHALL complete within 100ms
- Test generation from catalogs SHALL be deterministic

### Maintainability
- Catalogs SHALL be human-readable without tooling
- Each catalog SHALL include version and last-update metadata
- Changes to catalogs SHALL be tracked in version control

### Coverage Goals
- 100% of existing parsers SHALL have corresponding pattern catalogs
- All documented command variations SHALL be represented in catalogs
