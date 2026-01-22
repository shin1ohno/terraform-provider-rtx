# Requirements Document: Filter Enhancements

## Introduction

This specification defines three enhancements to the RTX Terraform provider's filtering capabilities:

1. **Ethernet Filter Interface Application** - Apply ethernet filters to interfaces
2. **IPv6 Dynamic Filter Protocol Extensions** - Add `submission` protocol support
3. **Restrict Action Support** - Enable `restrict` and `restrict-log` actions with `tcpfin`/`tcprst` protocols

These enhancements address gaps identified during import reconciliation between Terraform configurations and actual router settings, enabling full declarative management of RTX router filter configurations.

## Alignment with Product Vision

These features directly support the product objectives:

- **Complete IaC Management**: Currently, ethernet filter application to interfaces requires manual configuration, breaking the IaC paradigm
- **Security & Filtering Coverage**: The provider claims support for "IP Filters: Layer 3/4 packet filtering and firewall rules" but lacks `restrict` actions used for stateful TCP connection tracking
- **Import Support**: Users cannot import existing router configurations that use these features

## Requirements

### Requirement 1: Ethernet Filter Interface Application

**User Story:** As a network administrator, I want to apply ethernet filters to interfaces via Terraform, so that I can manage Layer 2 filtering rules declaratively without manual CLI intervention.

#### Background

RTX routers support applying ethernet filters to interfaces:
```
ethernet lan1 filter in 1 100
ethernet lan1 filter out 2 100
```

Currently, `rtx_ethernet_filter` resources define the filters but cannot apply them to interfaces.

#### Acceptance Criteria

1. WHEN a user defines ethernet filter application on an interface THEN the provider SHALL generate the appropriate `ethernet <interface> filter in/out` commands
2. WHEN multiple filter numbers are specified THEN the provider SHALL apply them in the specified order (first-match semantics)
3. WHEN the resource is destroyed THEN the provider SHALL remove the filter application using `no ethernet <interface> filter in/out`
4. WHEN importing an existing interface THEN the provider SHALL read and populate the ethernet filter application settings
5. IF an invalid interface name is provided THEN the provider SHALL return a validation error

### Requirement 2: IPv6 Dynamic Filter Protocol Extensions

**User Story:** As a network administrator, I want to use `submission` protocol in IPv6 dynamic filters, so that I can configure stateful inspection for email submission traffic (port 587).

#### Background

RTX routers support the `submission` protocol in IPv6 dynamic filters:
```
ipv6 filter dynamic 101085 * * submission syslog=off
```

The current `rtx_ipv6_filter_dynamic` resource schema restricts protocols to: `ftp`, `www`, `smtp`, `pop3`, `dns`, `domain`, `telnet`, `ssh`, `tcp`, `udp`, `*`.

#### Acceptance Criteria

1. WHEN a user specifies `submission` as protocol THEN the provider SHALL accept it as valid
2. WHEN generating CLI commands THEN the provider SHALL output `submission` as the protocol keyword
3. WHEN importing existing filters THEN the provider SHALL correctly parse `submission` protocol entries
4. IF an unsupported protocol is specified THEN the provider SHALL return a clear validation error listing valid options

### Requirement 3: Restrict Action Support

**User Story:** As a network administrator, I want to use `restrict` and `restrict-log` actions with TCP flag protocols (`tcpfin`, `tcprst`) in IP filters, so that I can implement stateful connection tracking and prevent DoS attacks.

#### Background

RTX routers support `restrict` actions that track TCP connection state:
```
ip filter 200026 restrict * * tcpfin * www,21,nntp
ip filter 200027 restrict * * tcprst * www,21,nntp
ip filter 500000 restrict * * * * *
```

The `restrict` action allows outbound TCP connections while blocking unsolicited inbound connections. The `tcpfin` and `tcprst` protocols filter TCP FIN/RST packets for connection teardown tracking.

#### Acceptance Criteria

1. WHEN a user specifies `restrict` or `restrict-log` as action THEN the provider SHALL accept it as valid
2. WHEN a user specifies `tcpfin` or `tcprst` as protocol THEN the provider SHALL accept it as valid
3. WHEN generating CLI commands THEN the provider SHALL output the exact action and protocol keywords
4. WHEN importing existing filters THEN the provider SHALL correctly parse restrict actions and TCP flag protocols
5. IF combining restrict action with incompatible protocols THEN the provider SHALL provide guidance via documentation

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each enhancement should be isolated to its relevant resource file
- **Modular Design**: Reuse existing validation helpers where possible
- **Dependency Management**: Minimize changes to shared parsing infrastructure
- **Clear Interfaces**: Extend existing schemas without breaking changes

### Performance
- No additional SSH commands required; changes are schema/validation only
- Parser changes should not impact performance

### Security
- Filter application follows router security model
- `restrict` action enhances security by enabling stateful filtering

### Reliability
- All changes must include unit tests for parsers
- Acceptance tests should cover CRUD operations
- Import functionality must be tested

### Usability
- Clear documentation with examples for each new feature
- Validation error messages should be actionable
- Maintain consistency with existing resource patterns

### Backward Compatibility
- Existing configurations must continue to work
- Schema changes must be additive, not breaking
