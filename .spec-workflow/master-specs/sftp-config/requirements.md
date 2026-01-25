# Master Specification: SFTP-Based Configuration Reading

## Introduction

This specification defines the SFTP-based configuration reading feature for the Yamaha RTX Terraform provider. When enabled, the provider downloads the router's complete configuration file via SFTP, caches it in memory, and parses it to serve resource read operations. This replaces hundreds of individual SSH CLI commands with a single file transfer.

### Key Benefits

1. **Performance**: Single SFTP transfer vs. hundreds of individual SSH commands (~5 minutes reduced to <10 seconds)
2. **Completeness**: Access to raw configuration including passwords and secrets masked in CLI output
3. **Consistency**: Atomic snapshot of configuration state (no mid-read changes)

### Prerequisites

- RTX router must have SFTP server enabled (`sftpd host` command)
- User must have SFTP connection permission
- Admin password required for configuration file access

## Alignment with Product Vision

This feature directly supports the product vision of providing reliable, production-ready infrastructure automation:

- **Performance**: Addresses the critical performance bottleneck in read operations
- **Reliability**: Provides complete configuration data without CLI masking
- **User Experience**: Dramatically reduces wait times for Terraform operations

## Requirements

### REQ-1: Provider-Level SFTP Configuration

**User Story:** As a Terraform user, I want to configure SFTP settings in the provider block, so that the provider can use SFTP for faster configuration reads.

#### Acceptance Criteria

| ID | Criterion |
|----|-----------|
| 1.1 | WHEN `use_sftp = true` AND `admin_password` is provided THEN provider SHALL use SFTP for bulk configuration reading |
| 1.2 | IF `use_sftp = true` AND `admin_password` is not provided THEN provider SHALL return a configuration error |
| 1.3 | WHEN `use_sftp = false` or not specified THEN provider SHALL use existing SSH CLI method |
| 1.4 | WHEN SFTP connection fails THEN provider SHALL fall back to SSH CLI method AND emit warning log |
| 1.5 | WHEN fallback to SSH occurs THEN provider SHALL log specific error (e.g., "SFTP connection refused", "authentication failed") |
| 1.6 | WHEN fallback to SSH occurs THEN provider SHALL continue to attempt SFTP for subsequent operations |

#### Implementation Status

- **Provider Schema**: `internal/provider/provider.go` - `use_sftp` (bool) and `sftp_config_path` (string) fields
- **Environment Variables**: `RTX_USE_SFTP`, `RTX_SFTP_CONFIG_PATH`
- **Client Config**: `internal/client/interfaces.go` - `SFTPEnabled` and `SFTPConfigPath` fields

### REQ-2: SFTP Client Implementation

**User Story:** As a Terraform user, I want the provider to download configuration via SFTP, so that read operations are significantly faster.

#### Acceptance Criteria

| ID | Criterion |
|----|-----------|
| 2.1 | WHEN SFTP enabled THEN provider SHALL connect using configured SSH parameters with admin password |
| 2.2 | WHEN connection established THEN provider SHALL download config from configurable path (default auto-detected) |
| 2.3 | WHEN config downloaded THEN provider SHALL store ONLY in memory (never written to disk) |
| 2.4 | WHEN download completes THEN provider SHALL close SFTP connection immediately |
| 2.5 | IF config file path does not exist THEN provider SHALL return clear error with file path |

#### Implementation Status

- **SFTP Client**: `internal/client/sftp_client.go`
  - `SFTPClient` interface with `Download(ctx, path)` and `Close()` methods
  - `NewSFTPClient(config)` constructor
  - Uses `github.com/pkg/sftp` library over SSH connection

- **Config Path Resolution**: `internal/client/config_path_resolver.go`
  - `ConfigPathResolver` with `Resolve(ctx)` method
  - Parses `show environment` output to detect startup config number
  - Returns SFTP path `/system/config{N}`
  - Falls back to `/system/config0` if detection fails

### REQ-3: Configuration Parsing and Resource Extraction

**User Story:** As a Terraform user, I want the downloaded configuration to be parsed into individual resource states.

#### Acceptance Criteria

| ID | Criterion |
|----|-----------|
| 3.1 | WHEN config downloaded THEN provider SHALL parse it to extract individual command blocks |
| 3.2 | WHEN parsing THEN provider SHALL identify resource types by command prefixes (e.g., `ip route`, `nat descriptor`) |
| 3.3 | WHEN parsing THEN provider SHALL extract unmasked sensitive values |
| 3.4 | WHEN multiple resources of same type exist THEN provider SHALL distinguish by identifiers |

#### Implementation Status

- **Config Parser**: `internal/rtx/parsers/config_file.go`
  - `ConfigFileParser` struct with context tracking
  - `ParsedConfig` struct containing parsed commands and contexts
  - Context-aware parsing for `tunnel select`, `pp select`, `ipsec tunnel` blocks
  - Extract methods: `ExtractStaticRoutes()`, `ExtractDHCPScopes()`, `ExtractNATMasquerade()`, `ExtractIPFilters()`, `ExtractPasswords()`

### REQ-4: Cached Configuration for Batch Reads

**User Story:** As a Terraform user, I want a single SFTP download to serve multiple resource reads.

#### Acceptance Criteria

| ID | Criterion |
|----|-----------|
| 4.1 | WHEN Terraform operation begins THEN provider SHALL download config at most ONCE |
| 4.2 | WHEN multiple resources request data THEN provider SHALL serve from cached config |
| 4.3 | WHEN operation ends THEN provider SHALL clear cached config from memory |
| 4.4 | WHEN write operation occurs THEN provider SHALL invalidate cache and re-download on next read |

#### Implementation Status

- **Config Cache**: `internal/client/config_cache.go`
  - `ConfigCache` struct with thread-safe access (sync.RWMutex)
  - Methods: `Get()`, `Set()`, `SetWithTTL()`, `Invalidate()`, `MarkDirty()`, `ClearDirty()`, `IsValid()`, `IsDirty()`
  - Default TTL: configurable via `DefaultCacheTTL`

- **Client Integration**: `internal/client/client.go`
  - `GetCachedConfig(ctx)` method on rtxClient
  - `SFTPEnabled()` method to check if SFTP is enabled
  - Automatic fallback to SSH with warning log on SFTP failure

### REQ-5: Password and Sensitive Value Extraction

**User Story:** As a Terraform user, I want to import resources with their actual passwords.

#### Acceptance Criteria

| ID | Criterion |
|----|-----------|
| 5.1 | WHEN parsing THEN provider SHALL extract plaintext passwords from config |
| 5.2 | WHEN password is plaintext THEN provider SHALL set it in resource state |
| 5.3 | WHEN password is hashed/encrypted THEN provider SHALL treat as unknown |
| 5.4 | WHEN importing THEN provider SHALL use SFTP-extracted passwords to populate state |

#### Implementation Status

- **Password Extraction**: `internal/rtx/parsers/config_file.go`
  - `ExtractedPasswords` struct with fields for all password types
  - Patterns recognized:
    - `login password <plaintext>`
    - `administrator password <plaintext>`
    - `pp auth username <user> <password>`
    - `ipsec ike pre-shared-key <n> text <secret>`
    - `l2tp tunnel auth on <secret>`
    - `login user <name> encrypted <hash>` (marked as encrypted)

## Non-Functional Requirements

### Performance Requirements

| Metric | Target |
|--------|--------|
| SFTP connection establishment | < 3 seconds |
| Configuration download (10-50KB) | < 2 seconds |
| Configuration parsing | < 1 second |
| Total read time improvement | From ~5 minutes to < 10 seconds |

### Security Requirements

- Admin password handled with same security as SSH password (Terraform sensitive variable)
- Configuration data in memory not logged or exposed
- SFTP session uses same encryption as SSH (runs over SSH subsystem)
- Sensitive values extracted from config marked as sensitive in Terraform state

### Reliability Requirements

- Clear error messages when SFTP not enabled on router
- Clear error messages for authentication failures
- Graceful handling of partial downloads or corrupted files
- Validation of downloaded configuration before parsing

### Usability Requirements

- Optional feature: existing users need not change anything
- Clear documentation on enabling SFTP on RTX routers
- Migration path: can enable/disable without breaking existing configurations

## Resource Integration

The following resources support SFTP cache reading:

| Resource | Implementation | Extract Method |
|----------|---------------|----------------|
| `rtx_static_route` | `resource_rtx_static_route.go` | `ExtractStaticRoutes()` |
| `rtx_dhcp_scope` | `resource_rtx_dhcp_scope.go` | `ExtractDHCPScopes()` |
| `rtx_nat_masquerade` | `resource_rtx_nat_masquerade.go` | `ExtractNATMasquerade()` |

Each resource's Read function:
1. Checks if SFTP is enabled via `client.SFTPEnabled()`
2. Calls `client.GetCachedConfig(ctx)` to get parsed config
3. Extracts resource state from parsed config
4. Falls back to SSH if not found in cache or on error

## Appendix A: RTX SFTP Configuration

To enable SFTP on RTX router:

```
sshd service on
sshd host lan1
sftpd host lan1
```

Configuration files are stored at:
- `/system/config0` through `/system/config4`
- Current startup config identified via `show environment` output

## Appendix B: Sample Configuration Format

```
# Admin
login password <plaintext>
administrator password <plaintext>

# Static Routes
ip route default gateway pp 1
ip route 10.0.0.0/8 gateway 192.168.1.1

# DHCP
dhcp scope 1 192.168.1.100-192.168.1.200/24

# NAT
nat descriptor type 1 masquerade
nat descriptor address outer 1 primary

# Tunnels (context-aware)
tunnel select 1
 tunnel encapsulation l2tpv3
 ipsec tunnel 101
  ipsec ike pre-shared-key 1 text <secret>
 tunnel enable 1
```
