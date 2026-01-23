# Requirements Document: SFTP-Based Configuration Reading

## Introduction

This feature enables high-performance configuration reading from Yamaha RTX routers using SFTP file transfer instead of sequential SSH CLI commands. The current implementation reads configuration for each Terraform resource by executing individual SSH commands, resulting in approximately 5-minute read times during `terraform refresh`, `terraform import`, and `terraform plan` operations.

SFTP-based bulk reading will dramatically reduce this time by downloading the entire configuration file (`config.txt` or `config0`) in a single SFTP transfer, then parsing it in memory to extract resource states.

### Key Benefits

1. **Performance**: Single SFTP transfer vs. hundreds of individual SSH commands
2. **Completeness**: Access to raw configuration including passwords and secrets that are masked in CLI output
3. **Consistency**: Atomic snapshot of configuration state (no mid-read changes)

### Trade-offs

Users must enable SFTP server on their RTX router, which has minor security implications:
- Opens an additional network service
- Requires admin password for SFTP authentication

## Alignment with Product Vision

This feature directly supports the product vision of providing reliable, production-ready infrastructure automation:

- **Performance**: Addresses the critical performance bottleneck in read operations
- **Reliability**: Provides complete configuration data without CLI masking
- **User Experience**: Dramatically reduces wait times for Terraform operations

## Requirements

### Requirement 1: Provider-Level SFTP Configuration

**User Story:** As a Terraform user, I want to configure SFTP credentials in the provider block, so that the provider can use SFTP for faster configuration reads.

#### Acceptance Criteria

1. WHEN the provider block includes `sftp_enabled = true` AND `admin_password` is provided THEN the provider SHALL use SFTP for bulk configuration reading.
2. IF `sftp_enabled = true` AND `admin_password` is not provided THEN the provider SHALL return a configuration error.
3. WHEN `sftp_enabled = false` or not specified THEN the provider SHALL use the existing SSH CLI method for all operations.
4. WHEN SFTP connection fails AND `sftp_enabled = true` THEN the provider SHALL fall back to SSH CLI method AND emit a warning log indicating the fallback reason.
5. WHEN fallback to SSH occurs THEN the provider SHALL log the specific error that caused SFTP failure (e.g., "SFTP connection refused", "authentication failed", "file not found").
6. WHEN fallback to SSH occurs THEN the provider SHALL continue to attempt SFTP for subsequent operations (not permanently disable SFTP for the session).

### Requirement 2: SFTP Client Implementation

**User Story:** As a Terraform user, I want the provider to download configuration via SFTP, so that read operations are significantly faster.

#### Acceptance Criteria

1. WHEN SFTP is enabled THEN the provider SHALL connect to the RTX router using the configured SSH connection parameters (host, port, username) with the admin password for SFTP authentication.
2. WHEN the SFTP connection is established THEN the provider SHALL download the configuration file from a configurable path (default: `/config.txt` or `/config0`).
3. WHEN the configuration file is downloaded THEN the provider SHALL store it ONLY in memory (never written to disk).
4. WHEN the download completes THEN the provider SHALL close the SFTP connection immediately.
5. IF the configuration file path does not exist THEN the provider SHALL return a clear error message indicating the file path.

### Requirement 3: Configuration Parsing and Resource Extraction

**User Story:** As a Terraform user, I want the downloaded configuration to be parsed into individual resource states, so that `terraform refresh` shows accurate state.

#### Acceptance Criteria

1. WHEN configuration is downloaded THEN the provider SHALL parse it to extract individual command blocks.
2. WHEN parsing configuration THEN the provider SHALL identify resource types by command prefixes (e.g., `ip route`, `nat descriptor`, `dhcp scope`).
3. WHEN parsing configuration THEN the provider SHALL extract unmasked sensitive values (passwords, keys) that would be hidden in CLI output.
4. WHEN multiple resources of the same type exist THEN the provider SHALL correctly distinguish them by their identifiers.

### Requirement 4: Cached Configuration for Batch Reads

**User Story:** As a Terraform user, I want a single SFTP download to serve multiple resource reads within one Terraform operation, so that performance is maximized.

#### Acceptance Criteria

1. WHEN a Terraform operation begins (plan, apply, refresh) THEN the provider SHALL download configuration at most ONCE per operation.
2. WHEN multiple resources request configuration data THEN the provider SHALL serve them from the cached in-memory configuration.
3. WHEN the Terraform operation ends THEN the provider SHALL clear the cached configuration from memory.
4. WHEN a write operation occurs during the Terraform run THEN the provider SHALL invalidate the cache and re-download on next read.

### Requirement 5: Password and Sensitive Value Extraction

**User Story:** As a Terraform user, I want to import resources with their actual passwords, so that the imported state matches the real configuration.

#### Acceptance Criteria

1. WHEN parsing configuration from SFTP download THEN the provider SHALL extract plaintext passwords that appear in the config file.
2. WHEN a password field contains a plaintext value THEN the provider SHALL set it in the resource state.
3. WHEN a password field contains a hashed/encrypted value (e.g., `*encrypted*`) THEN the provider SHALL treat it as unknown.
4. WHEN importing a resource THEN the provider SHALL use the SFTP-extracted password values to populate the state.

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: SFTP client, configuration parser, and cache manager shall be separate modules
- **Modular Design**: SFTP reading shall be an alternative to SSH reading, using the same resource extraction interfaces
- **Dependency Management**: SFTP functionality shall use Go's standard `sftp` library (github.com/pkg/sftp)
- **Clear Interfaces**: Define `ConfigurationReader` interface that both SSH and SFTP implementations satisfy

### Performance

- SFTP connection establishment: < 3 seconds
- Configuration file download (typical 10-50KB): < 2 seconds
- Configuration parsing: < 1 second
- Total read time improvement: From ~5 minutes to < 10 seconds

### Security

- Admin password must be handled with the same security as SSH password (Terraform sensitive variable)
- Configuration data in memory must not be logged or exposed
- SFTP session must use the same encryption as SSH (it runs over SSH subsystem)
- Sensitive values extracted from config shall be marked as sensitive in Terraform state

### Reliability

- Clear error messages when SFTP is not enabled on the router
- Clear error messages for authentication failures
- Graceful handling of partial downloads or corrupted files
- Validation of downloaded configuration before parsing

### Usability

- Optional feature: existing users need not change anything
- Clear documentation on enabling SFTP on RTX routers
- Migration path: can enable/disable without breaking existing configurations

## Appendix A: Sample Configuration File Format

The following is a representative sample of RTX config.txt format with test data:

```
#
# Admin
#
login password test-login-password-123
administrator password test-admin-password-456
login user testuser encrypted TESTENCRYPTEDHASH123456789
user attribute administrator=off connection=off gui-page=dashboard,lan-map,config login-timer=300
user attribute testuser connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600
timezone +09:00
console character ja.utf8
console prompt "[TEST-RTX] "

httpd host any
sshd service on
sshd host lan1
sftpd host lan1

#
# WAN connection
#
description lan2 test-wan
ip lan2 address 198.51.100.1/24
ip lan2 nat descriptor 1000
ip lan2 secure filter in 200020 200099
ip lan2 secure filter out 200099 dynamic 200080 200081

#
# IP configuration
#
ip route default gateway 198.51.100.254
ip route 10.0.0.0/8 gateway 192.0.2.1

#
# LAN configuration
#
ip lan1 address 192.0.2.253/24

#
# Services
#
dhcp service server
dhcp scope 1 192.0.2.100-192.0.2.199/24 gateway 192.0.2.253 expire 12:00
dhcp scope bind 1 192.0.2.100 01:00:11:22:33:44:55
dhcp scope option 1 dns=192.0.2.10

dns host lan1
dns service recursive
dns server select 1 192.0.2.10 edns=on any example.local
dns server select 500000 8.8.8.8 edns=on 8.8.4.4 edns=on any .

#
# Tunnels
#
pp select anonymous
 pp bind tunnel1
 pp auth request mschap-v2
 pp auth username vpnuser test-vpn-password-789
 ppp ipcp ipaddress on
 pp enable anonymous

tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike pre-shared-key 1 text test-ike-psk-secret
  ipsec ike remote address 1 test.example.com
 l2tp tunnel auth on test-l2tp-auth-secret
 tunnel enable 1

#
# Filters
#
ip filter 200020 reject * * udp,tcp 135 *
ip filter 200099 pass * * * * *
ip filter dynamic 200080 * * ftp
ip filter dynamic 200081 * * www

nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 192.0.2.253 tcp 22
```

### Key Observations for Parser Implementation

1. **Password formats**: Plaintext passwords appear directly after keywords (`login password`, `administrator password`, `pp auth username ... <password>`, `ipsec ike pre-shared-key ... text <secret>`, `l2tp tunnel auth on <secret>`)
2. **Hierarchical context**: Some commands are context-dependent (e.g., commands after `tunnel select 1` apply to tunnel 1)
3. **Numbered resources**: Many resources use numeric identifiers (filter numbers, NAT descriptor IDs, tunnel numbers)
4. **Interface references**: Commands reference interfaces by name (`lan1`, `lan2`, `bridge1`, `tunnel1`)
5. **Comment blocks**: Lines starting with `#` are comments and should be ignored
