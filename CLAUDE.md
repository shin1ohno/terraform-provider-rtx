# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for Yamaha RTX series routers - enterprise-grade internet routers widely used in Japan for network infrastructure.

## Development Setup

Project structure:
- `internal/provider/` - Provider, resources, and data sources
- `internal/client/` - SSH client and RTX router communication
- `internal/rtx/` - RTX command parsers
- `examples/` - Example Terraform configurations
- `docs/` - Generated documentation

## Common Development Commands

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

## Logging

This project uses **Zerolog** (`github.com/rs/zerolog`) for all logging. Do NOT use the standard `log` package.

```go
// Use structured logging
log.Error().Err(err).Str("file", path).Msg("Failed to read file")
log.Info().Str("resource", name).Msg("Resource created")
log.Debug().Interface("config", cfg).Msg("Configuration loaded")
```

- Main logging infrastructure: `internal/logging/logger.go`
- Environment variables: `TF_LOG` (level), `TF_LOG_JSON` (format)

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

## Release Process

### Prerequisites

1. GPG key configured in GitHub Secrets (`GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`)
2. Public key registered at https://registry.terraform.io

### Version Selection

When user requests a release, ask which version bump type using AskUserQuestion:
- **Patch** (X.Y.Z → X.Y.Z+1): Bug fixes, documentation updates
- **Minor** (X.Y.Z → X.Y+1.0): New features, backward-compatible changes
- **Major** (X.Y.Z → X+1.0.0): Breaking changes

Determine new version by reading current version from `Makefile` (VERSION variable).

### Release Steps

```bash
# 1. Update version number in all locations:
#    - Makefile (VERSION variable)
#    - README.md (installation examples, paths)
#    - docs/index.md (version constraint in example)
#    - examples/**/*.tf (version constraints - typically use ~> X.Y)

# 2. Run go generate to format examples and regenerate docs
go generate ./...

# 3. Run lint and tests
make lint
make test

# 4. Commit changes
git add -A && git commit -m "release: bump version to X.Y.Z"

# 5. Ask user to push and create tag (Claude cannot run git push)
#    User runs:
#      git push origin main
#      # Wait for CI to pass
#      git tag vX.Y.Z
#      git push origin vX.Y.Z
```

**Note**: Claude cannot execute `git push`. Ask user to run push commands manually.

GitHub Actions will automatically:
- Build binaries for all platforms
- Sign with GPG
- Create GitHub Release
- Terraform Registry detects and publishes automatically

### Verify Release

```bash
# Check Terraform Registry API
curl -s "https://registry.terraform.io/v1/providers/shin1ohno/rtx/versions"

# Test terraform init
cd examples/provider
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### Provider Source

- Registry: `shin1ohno/rtx`
- All examples and documentation should use this source