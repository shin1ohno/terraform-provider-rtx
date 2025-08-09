# Terraform Provider for Yamaha RTX Project Overview

## Project Purpose
This is a Terraform provider for managing Yamaha RTX series routers - enterprise-grade network infrastructure devices widely used in Japan. The provider enables Infrastructure as Code (IaC) management of RTX router configurations.

## Tech Stack
- **Language**: Go 1.23.0
- **Framework**: Terraform Plugin SDK v2 (v2.37.0)
- **Testing**: Go testing package, testify for mocking
- **Development Tools**: 
  - Make for build automation
  - Docker for test environment simulation
  - golangci-lint for linting (not yet configured)

## Architecture
- **Provider Package** (`internal/provider/`): Terraform provider implementation and data sources
- **Client Package** (`internal/client/`): SSH client for RTX router communication
  - Executor pattern for command execution abstraction
  - Retry strategies for resilient connections
  - Parser interface for command output processing
- **RTX Package** (`internal/rtx/`): RTX-specific parsers and logic
  - Parser registry pattern for model-specific parsing
  - Support for multiple RTX models (RTX830, RTX1210, RTX1220)

## Key Features Implemented
1. **Data Sources**:
   - `rtx_system_info`: Router system information (model, firmware, serial number, etc.)
   - `rtx_interfaces`: Network interface information
   - `rtx_routes`: Routing table information

2. **SSH Client**:
   - Secure SSH connections with host key verification options
   - Command execution with retry logic
   - Model-specific prompt detection
   - Context-aware connection management

3. **Testing Infrastructure**:
   - Docker-based RTX simulator for integration testing
   - Comprehensive unit tests with mocking
   - Acceptance test framework

## Project Structure
```
terraform-provider-rtx/
├── main.go                 # Provider entry point
├── Makefile               # Build and development commands
├── go.mod/go.sum          # Go module dependencies
├── internal/
│   ├── provider/          # Terraform provider implementation
│   ├── client/            # SSH client for RTX communication
│   └── rtx/               # RTX-specific logic and parsers
├── test/
│   └── docker/            # Docker test environment
├── examples/              # Usage examples
└── docs/                  # Documentation
```

## Environment Configuration
The provider supports configuration via:
- Terraform provider block
- Environment variables:
  - `RTX_HOST`: Router hostname/IP
  - `RTX_USERNAME`: SSH username
  - `RTX_PASSWORD`: SSH password

## Current Status
- Basic read-only data sources implemented
- TDD approach with high test coverage
- Ready for write operations (resources) implementation
- Next milestone: `rtx_dns_host` resource for DNS management