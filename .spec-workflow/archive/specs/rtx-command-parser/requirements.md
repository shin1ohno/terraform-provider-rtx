# Requirements Document: RTX Command Parser Test Coverage

## Introduction

This feature ensures comprehensive test coverage for existing RTX command parsers by systematically extracting all command patterns from the 59-chapter RTX command reference documentation and creating test cases to verify parser correctness. The goal is to identify and fix any parsing gaps in the existing implementation.

## Alignment with Product Vision

This feature supports the product vision by:
- Ensuring parser reliability for accurate import/export of RTX configurations
- Reducing configuration errors through comprehensive validation
- Following the existing Parser Registry Pattern without unnecessary refactoring

## Requirements

### REQ-1: Command Pattern Extraction from Documentation

**User Story:** As a provider developer, I want all RTX command patterns extracted from the documentation, so that I can create comprehensive test cases.

#### Acceptance Criteria

1. WHEN processing each chapter of RTX command documentation THEN the system SHALL extract all `[書式]` (format) sections
2. WHEN extracting command patterns THEN the system SHALL capture:
   - Command name and subcommands
   - Required parameters
   - Optional parameters (in brackets `[]`)
   - Parameter value types and ranges from `[設定値]` sections
   - Default values from `[初期値]` sections
3. WHEN a command has multiple format variations THEN each variation SHALL be extracted as a separate pattern
4. WHEN a `no` form exists THEN it SHALL be extracted alongside the positive form

### REQ-2: Command Pattern Categorization

**User Story:** As a provider developer, I want command patterns organized by category, so that I can prioritize test coverage for implemented resources.

#### Acceptance Criteria

1. WHEN organizing extracted patterns THEN they SHALL be categorized by chapter/feature area:
   - IP Configuration (Chapter 8)
   - Ethernet Filter (Chapter 9)
   - DHCP (Chapter 12)
   - IPsec (Chapter 15)
   - L2TP (Chapter 16)
   - PPTP (Chapter 17)
   - NAT (Chapter 23)
   - DNS (Chapter 24)
   - QoS (Chapter 26)
   - OSPF (Chapter 28)
   - BGP (Chapter 29)
   - IPv6 (Chapter 30)
   - VLAN (Chapter 38)
   - Schedule (Chapter 37)
   - SNMP (Chapter 21)
   - Syslog (Chapter 59)
   - System/Admin (Chapter 4)
2. WHEN a pattern maps to an existing parser THEN it SHALL be tagged with the parser name
3. WHEN a pattern has no corresponding parser THEN it SHALL be flagged for future implementation

### REQ-3: Test Case Generation

**User Story:** As a provider developer, I want test cases generated from extracted patterns, so that I can verify parser correctness.

#### Acceptance Criteria

1. WHEN generating test cases THEN each test SHALL include:
   - Input: Raw RTX command string (from `[設定例]` or constructed from pattern)
   - Expected output: Structured data that the parser should produce
   - Edge cases: Boundary values, optional parameter combinations
2. WHEN a pattern has documented examples (`[設定例]`) THEN those SHALL be used as test inputs
3. WHEN a pattern has value ranges THEN test cases SHALL include:
   - Minimum valid value
   - Maximum valid value
   - Typical/common value
4. WHEN a pattern has optional parameters THEN test cases SHALL cover:
   - All parameters omitted
   - Each optional parameter included individually
   - All optional parameters included

### REQ-4: Existing Parser Verification

**User Story:** As a provider developer, I want to run test cases against existing parsers, so that I can identify parsing gaps.

#### Acceptance Criteria

1. WHEN running tests against existing parsers THEN the system SHALL report:
   - Pass: Parser correctly handles the pattern
   - Fail: Parser produces incorrect output
   - Skip: No parser exists for this pattern
2. WHEN a test fails THEN the report SHALL include:
   - Input command string
   - Expected output
   - Actual output
   - Diff between expected and actual
3. WHEN tests complete THEN a coverage report SHALL show:
   - Total patterns extracted per category
   - Patterns with passing tests
   - Patterns with failing tests
   - Patterns with no parser coverage

### REQ-5: Interface Name Pattern Coverage

**User Story:** As a provider developer, I want test cases for all interface naming conventions, so that interface-scoped commands are correctly parsed.

#### Acceptance Criteria

1. WHEN generating interface-related tests THEN coverage SHALL include:
   - Physical LAN: `lan1`, `lan2`
   - LAN division: `lan1.1`, `lan1.2`
   - Tagged VLAN: `lan1/1`, `lan1/2`
   - PP interfaces: `pp`, `pp1`
   - Tunnel interfaces: `tunnel1`, `tunnel2`
   - Loopback interfaces: `loopback1` through `loopback9`
   - Bridge interfaces: `bridge1`
   - VLAN interfaces: `vlan1`, `vlan2`
   - Special interfaces: `null`
2. WHEN a command accepts multiple interface types THEN test cases SHALL cover each valid type

### REQ-6: Value Type Coverage

**User Story:** As a provider developer, I want test cases for all value types, so that value parsing is thoroughly tested.

#### Acceptance Criteria

1. WHEN generating value-related tests THEN coverage SHALL include:
   - Boolean switches: `on`, `off`
   - IPv4 addresses: `192.168.1.1`, `0.0.0.0`, `255.255.255.255`
   - IPv4 with CIDR: `192.168.1.0/24`, `10.0.0.0/8`
   - Subnet masks: `/24`, `/255.255.255.0`, `0xffffff00`
   - IPv6 addresses: Standard and compressed formats
   - Numeric ranges: Boundary values for each documented range
   - MAC addresses: `00:00:00:00:00:00` format
   - Time values: `HH:MM` format
   - Named parameters: `key=value` syntax
   - Keywords: All documented keyword options

### REQ-7: Gap Analysis Report

**User Story:** As a provider developer, I want a gap analysis report, so that I can prioritize parser improvements.

#### Acceptance Criteria

1. WHEN analysis completes THEN the report SHALL list:
   - Commands with complete test coverage
   - Commands with partial test coverage (some patterns tested)
   - Commands with no test coverage
   - Commands with no parser implementation
2. WHEN prioritizing gaps THEN the report SHALL consider:
   - Commands used by existing Terraform resources (high priority)
   - Commands commonly used in enterprise configurations (medium priority)
   - Commands for deprecated/legacy features (low priority)

### REQ-8: Test Fixture Organization

**User Story:** As a provider developer, I want test fixtures organized systematically, so that tests are maintainable.

#### Acceptance Criteria

1. WHEN organizing test fixtures THEN they SHALL follow the pattern:
   ```
   internal/rtx/testdata/
   ├── commands/
   │   ├── ip/
   │   │   ├── ip_routing.txt
   │   │   ├── ip_address.txt
   │   │   └── ...
   │   ├── ipsec/
   │   ├── vlan/
   │   └── ...
   └── expected/
       ├── ip/
       ├── ipsec/
       └── ...
   ```
2. WHEN adding new test cases THEN they SHALL include comments referencing the documentation source (chapter, section, page)

## Non-Functional Requirements

### Code Architecture and Modularity

- **Leverage Existing Structure**: Use existing `internal/rtx/parsers/` and test file patterns
- **Table-Driven Tests**: Use Go's table-driven test pattern for command variations
- **Test Data Separation**: Keep test inputs and expected outputs in separate files for readability

### Performance

- Test suite execution shall complete within 60 seconds for all patterns
- Individual parser tests shall complete within 100ms

### Maintainability

- Test fixtures shall be human-readable plain text
- Expected outputs shall be documented JSON for easy comparison
- Each test file shall reference its source documentation section

### Coverage Goals

- Phase 1: 100% coverage for commands used by existing Terraform resources
- Phase 2: 80% coverage for all documented command patterns
- Phase 3: Edge case coverage for complex parameter combinations
