# Technology Stack

## Project Type

Terraform Provider plugin for Yamaha RTX series routers. This is a Go-based infrastructure automation tool that integrates with HashiCorp Terraform to enable Infrastructure as Code (IaC) management of network devices.

## Core Technologies

### Primary Language(s)
- **Language**: Go 1.23
- **Toolchain**: go1.23.12
- **Package Management**: Go modules (go.mod/go.sum)

### Key Dependencies/Libraries
- **terraform-plugin-sdk/v2 v2.37.0**: HashiCorp's SDK for building Terraform providers
- **terraform-plugin-go v0.27.0**: Low-level Terraform plugin protocol implementation
- **terraform-plugin-log v0.9.0**: Structured logging for Terraform providers
- **golang.org/x/crypto**: SSH client implementation for RTX router communication
- **hcl/v2 v2.23.0**: HashiCorp Configuration Language parser
- **go-cty v1.16.2**: Type system for Terraform values
- **stretchr/testify v1.10.0**: Testing utilities and assertions

### Application Architecture
- **Plugin Architecture**: Implements Terraform's gRPC-based plugin protocol
- **Resource-Oriented Design**: Each RTX feature maps to Terraform resources/data sources
- **Parser Registry Pattern**: Centralized parser registration for RTX CLI output parsing
- **Stateless Communication**: Each operation establishes fresh SSH connection to router

### Data Storage
- **Primary Storage**: Terraform state file (managed by Terraform core)
- **Router Storage**: RTX router's persistent configuration memory
- **Data Format**: HCL for configuration, JSON for state, plain text for RTX CLI

### External Integrations
- **Protocols**: SSH for RTX router CLI access
- **Authentication**: SSH password or public key authentication
- **Target Device**: Yamaha RTX series routers (RTX810, RTX830, RTX1200, RTX1210, RTX1220, etc.)

## Development Environment

### Build & Development Tools
- **Build System**: Make + Go build
- **Package Management**: Go modules
- **Local Installation**: `make install` deploys to ~/.terraform.d/plugins/

### Code Quality Tools
- **Static Analysis**: golangci-lint
- **Formatting**: gofmt, terraform fmt
- **Testing Framework**: Go testing package + testify
- **Documentation**: terraform-plugin-docs (tfplugindocs)

### Version Control & Collaboration
- **VCS**: Git
- **Branching Strategy**: Feature branches merged to main
- **Code Review Process**: Pull request based

## Deployment & Distribution

- **Target Platform(s)**: Cross-platform (Linux, macOS, Windows)
- **Distribution Method**:
  - Local build via `make install`
  - Future: Terraform Registry publication
- **Installation Requirements**:
  - Terraform >= 1.0
  - Network access to RTX router via SSH
- **Update Mechanism**: Terraform provider version constraints

## Technical Requirements & Constraints

### Performance Requirements
- SSH connection establishment: < 5 seconds
- Command execution timeout: Configurable, default 30 seconds
- Concurrent operations: Sequential per router (SSH session limitation)

### Compatibility Requirements
- **Platform Support**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- **Terraform Versions**: >= 1.0.0
- **RTX Firmware**: Tested with Rev.15.x and later
- **Go Version**: >= 1.23

### Security & Compliance
- **Security Requirements**:
  - SSH encryption for all router communication
  - Credential handling via Terraform's secure variable mechanism
  - No credential storage in state file
- **Threat Model**:
  - Protect router access credentials
  - Validate all user inputs before sending to router
  - Prevent command injection in SSH sessions

### Scalability & Reliability
- **Expected Load**: Single router per provider instance (multi-router via aliases)
- **Availability Requirements**: Graceful handling of router unreachability
- **Error Recovery**: Automatic retry with exponential backoff for transient failures

## Technical Decisions & Rationale

### Decision Log
1. **terraform-plugin-sdk/v2 over plugin-framework**: SDK v2 chosen for mature ecosystem and extensive documentation; migration to framework possible in future
2. **SSH over Telnet/API**: SSH provides encryption, authentication, and is universally supported on RTX routers; no REST API available on RTX
3. **Parser Registry Pattern**: Enables modular parsing of different RTX CLI outputs; each feature can register its own parsers
4. **Stateless SSH Sessions**: Each CRUD operation opens new SSH connection for reliability; connection pooling considered but adds complexity

## Known Limitations

- **Sequential Operations**: SSH sessions are not multiplexed; concurrent operations on same router may conflict
- **CLI Parsing Fragility**: RTX CLI output format may vary between firmware versions; parsers need version-specific handling
- **No Transaction Support**: RTX router doesn't support atomic multi-command transactions; partial failures possible
- **Limited Dry-Run**: `terraform plan` cannot fully validate against router state without actual SSH connection
