# Requirements Document: Parser Edge Case Tests

## Introduction

This feature enhances RTX command parser test coverage by adding edge case tests for parsers with high pattern counts but proportionally fewer test functions. The goal is to ensure all parameter combinations and boundary values are explicitly tested, improving parser reliability and catching potential regressions.

## Alignment with Product Vision

This feature supports the product vision by:
- Ensuring parser reliability for accurate import/export of RTX configurations
- Reducing configuration errors through comprehensive validation
- Improving confidence in handling complex VPN, routing, and NAT configurations
- Following established pattern-driven testing methodology

## Requirements

### REQ-1: IPsec Tunnel Parser Edge Case Tests

**User Story:** As a provider developer, I want comprehensive tests for IPsec tunnel parser, so that all encryption algorithm and IKE configuration combinations are validated.

#### Acceptance Criteria

1. WHEN testing encryption algorithms THEN the parser SHALL correctly handle:
   - DES-CBC, 3DES-CBC, AES-128-CBC, AES-256-CBC
   - AES-GCM-128, AES-GCM-256
   - All valid algorithm combinations
2. WHEN testing IKE versions THEN the parser SHALL correctly handle:
   - IKEv1 configurations
   - IKEv2 configurations
   - Version-specific options (main mode, aggressive mode)
3. WHEN testing NAT traversal THEN the parser SHALL correctly handle:
   - NAT-T enabled/disabled
   - Keep-alive intervals
   - Port configurations
4. WHEN testing authentication THEN the parser SHALL correctly handle:
   - Pre-shared key with special characters
   - Certificate-based authentication
   - Extended authentication (XAUTH)

**Test Count Target:** ~20 additional tests

### REQ-2: OSPF Parser Edge Case Tests

**User Story:** As a provider developer, I want comprehensive tests for OSPF parser, so that all area configurations and routing options are validated.

#### Acceptance Criteria

1. WHEN testing area types THEN the parser SHALL correctly handle:
   - Normal areas (area 0, area 1-255)
   - Stub areas with no-summary option
   - NSSA areas with default-information-originate
   - Totally stubby areas
2. WHEN testing authentication THEN the parser SHALL correctly handle:
   - No authentication
   - Simple password authentication
   - MD5 authentication with key-id
3. WHEN testing redistribution THEN the parser SHALL correctly handle:
   - Redistribution from static routes
   - Redistribution from connected networks
   - Redistribution with metric and metric-type
   - Route filtering with prefix-lists
4. WHEN testing interface settings THEN the parser SHALL correctly handle:
   - Cost configuration per interface
   - Priority settings
   - Hello and dead intervals
   - Network type (broadcast, point-to-point)

**Test Count Target:** ~15 additional tests

### REQ-3: NAT Parser Edge Case Tests

**User Story:** As a provider developer, I want comprehensive tests for NAT parser, so that all NAT types and port mapping configurations are validated.

#### Acceptance Criteria

1. WHEN testing port range mappings THEN the parser SHALL correctly handle:
   - Single port mapping (tcp/udp)
   - Port range mapping (e.g., 1000-2000)
   - Multiple protocol mappings on same address
2. WHEN testing protocol variations THEN the parser SHALL correctly handle:
   - TCP-only mappings
   - UDP-only mappings
   - ICMP mappings
   - Protocol-agnostic mappings (any)
3. WHEN testing dynamic NAT THEN the parser SHALL correctly handle:
   - Dynamic pool configurations
   - Address pool ranges
   - Overload (PAT) configurations
4. WHEN testing static NAT edge cases THEN the parser SHALL correctly handle:
   - 1:1 mappings with specific ports
   - Inside/outside address inversions
   - NAT hairpinning configurations

**Test Count Target:** ~15 additional tests

### REQ-4: BGP Parser Edge Case Tests

**User Story:** As a provider developer, I want comprehensive tests for BGP parser, so that all neighbor and routing policy configurations are validated.

#### Acceptance Criteria

1. WHEN testing route-map options THEN the parser SHALL correctly handle:
   - Route-map in/out per neighbor
   - Multiple route-map clauses
   - Match and set actions
2. WHEN testing 4-byte ASN THEN the parser SHALL correctly handle:
   - ASN in asdot notation (e.g., 1.65535)
   - ASN in asplain notation (e.g., 65536)
   - Private ASN ranges
3. WHEN testing community attributes THEN the parser SHALL correctly handle:
   - Standard communities (AA:NN format)
   - Well-known communities (no-export, no-advertise)
   - Extended communities
4. WHEN testing neighbor configurations THEN the parser SHALL correctly handle:
   - eBGP multihop with TTL security
   - Update source specification
   - Password with special characters
   - Timers (keepalive, holdtime)

**Test Count Target:** ~10 additional tests

## Non-Functional Requirements

### Code Architecture and Modularity
- **Table-Driven Tests**: All new tests SHALL use Go's table-driven test pattern
- **Pattern Alignment**: Tests SHALL reference YAML pattern catalog entries
- **Test Independence**: Each test case SHALL be independent and not rely on other test state
- **Clear Naming**: Test names SHALL describe the specific edge case being tested

### Performance
- Individual test cases SHALL complete within 10ms
- Full test suite execution SHALL complete within 60 seconds

### Maintainability
- Test cases SHALL include comments referencing the RTX documentation source
- Edge cases SHALL be documented in the corresponding YAML pattern catalog
- Test failures SHALL produce clear, actionable error messages

### Coverage Goals
- Target: 90% edge case coverage for documented parameter combinations
- All encryption/authentication algorithms explicitly tested
- All boundary values for numeric parameters tested
