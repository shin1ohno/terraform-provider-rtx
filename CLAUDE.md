# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for Yamaha RTX series routers - enterprise-grade internet routers widely used in Japan for network infrastructure.

## Development Setup

This project is currently empty and needs initial setup. A typical Terraform provider project structure would include:

- Provider implementation in Go
- Resource and data source definitions
- Acceptance tests
- Documentation generation
- Build and release automation

## Common Development Commands

Once the project is initialized, typical commands would include:

```bash
# Build the provider
go build -o terraform-provider-rtx

# Run tests
go test ./...

# Run acceptance tests (requires RTX router access)
TF_ACC=1 go test ./... -v

# Generate documentation
tfplugindocs

# Install provider locally for testing
make install
```

## Architecture Notes

- Terraform providers are typically structured around resources and data sources
- RTX router integration would require SSH/Telnet access or API connectivity
- Provider should handle RTX router configuration management
- Resources might include network interfaces, routing tables, firewall rules, and VPN configurations

## Yamaha RTX Router Integration

Yamaha RTX routers are enterprise-grade networking devices that support:
- Advanced routing protocols (BGP, OSPF, RIP)
- VPN connectivity (IPsec, L2TP, PPTP)
- Firewall and security features
- Quality of Service (QoS) management
- VLAN configuration
- Dynamic DNS
- Load balancing and failover

The provider should interact with RTX routers via:
- SSH or Telnet connections for configuration management
- Command-line interface (CLI) commands
- Configuration backup and restore
- Real-time status monitoring