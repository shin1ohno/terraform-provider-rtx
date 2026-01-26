# Requirements Document: SSHD Public Key Authentication

## Introduction

This feature enables comprehensive SSH public key authentication support for Yamaha RTX routers:

1. **Provider Authentication**: Allow the Terraform provider itself to connect to RTX routers using SSH public key authentication (in addition to existing password authentication)
2. **Router Configuration**: Manage SSH public key authentication settings on RTX routers via Terraform resources (host key generation, authorized keys, authentication methods)

This allows network administrators to implement a fully key-based SSH infrastructure through Infrastructure as Code.

## Alignment with Product Vision

This feature directly supports the product goal of comprehensive IaC management for RTX routers. SSH public key authentication is a fundamental security feature for enterprise network management, enabling:

- Passwordless, secure remote access to routers
- Centralized key management across multiple devices
- Compliance with security policies requiring key-based authentication
- Integration with existing SSH key infrastructure (e.g., 1Password SSH Agent)

Following the Cisco-compatibility principle, resource naming will align with similar Terraform provider conventions where applicable.

## Requirements

### Requirement 1: Provider SSH Public Key Authentication

**User Story:** As a network administrator, I want to configure the Terraform provider to connect to RTX routers using SSH public key authentication, so that I can avoid storing passwords and integrate with SSH agents (e.g., 1Password SSH Agent).

#### Acceptance Criteria

1. WHEN `private_key` attribute is set in provider configuration THEN the system SHALL authenticate using the provided private key content
2. WHEN `private_key_file` attribute is set THEN the system SHALL read the private key from the specified file path
3. IF `private_key_passphrase` is provided THEN the system SHALL use it to decrypt the private key
4. IF neither `private_key` nor `private_key_file` is set THEN the system SHALL fall back to password authentication (existing behavior)
5. WHEN using environment variables THEN the system SHALL support `RTX_PRIVATE_KEY`, `RTX_PRIVATE_KEY_FILE`, and `RTX_PRIVATE_KEY_PASSPHRASE`
6. IF SSH agent is available and no explicit key is provided THEN the system SHALL attempt SSH agent authentication

#### Provider Schema Addition

```hcl
provider "rtx" {
  host     = "192.168.1.1"
  username = "admin"

  # Option 1: Private key content (from vault, etc.)
  private_key = var.rtx_private_key

  # Option 2: Private key file path
  private_key_file       = "~/.ssh/id_ed25519"
  private_key_passphrase = var.key_passphrase  # Optional

  # Option 3: Password authentication (existing)
  password = var.rtx_password
}
```

### Requirement 2: SSH Host Key Generation

**User Story:** As a network administrator, I want to generate SSH host keys on RTX routers via Terraform, so that I can establish a secure SSH server identity during initial setup.

#### Acceptance Criteria

1. WHEN `rtx_sshd_host_key` resource is created THEN the system SHALL execute `sshd host key generate` command on the router
2. IF host key already exists THEN the system SHALL read the existing key without regenerating (idempotent behavior)
3. WHEN resource is deleted THEN the system SHALL NOT delete the host key (host keys should persist for security continuity)
4. WHEN resource is imported THEN the system SHALL read the existing host key fingerprint

### Requirement 3: SSH Authorized Keys Management

**User Story:** As a network administrator, I want to manage SSH authorized keys for router users via Terraform, so that I can control who can access routers using public key authentication.

#### Acceptance Criteria

1. WHEN `rtx_sshd_authorized_keys` resource is created THEN the system SHALL register all specified public keys for the user via `import sshd authorized-keys` command
2. IF public keys are added or removed in configuration THEN the system SHALL delete all existing keys and re-register the desired keys (RTX limitation: no individual key deletion)
3. WHEN resource is deleted THEN the system SHALL remove all authorized keys for the user via `delete /ssh/authorized_keys/<username>`
4. WHEN user specifies multiple keys THEN the system SHALL register all keys for that user
5. WHEN resource is imported THEN the system SHALL read existing authorized keys via `show sshd authorized-keys <username>`

#### Key Update Behavior (RTX Limitation Handling)

RTX routers do not support individual key deletion. Therefore, when the `keys` list changes, the provider performs:

1. `delete /ssh/authorized_keys/<username>` (remove all existing keys)
2. `import sshd authorized-keys <username>` × N (register each desired key)

**Example:**
```hcl
# Initial state: 3 keys registered
resource "rtx_sshd_authorized_keys" "admin" {
  username = "admin"
  keys = [
    "ssh-ed25519 AAA... key1",
    "ssh-ed25519 BBB... key2",
    "ssh-ed25519 CCC... key3",
  ]
}

# After removing key2 from config:
# 1. Provider detects change (3 keys → 2 keys)
# 2. Provider executes: delete /ssh/authorized_keys/admin
# 3. Provider executes: import sshd authorized-keys admin (for key1)
# 4. Provider executes: import sshd authorized-keys admin (for key3)
```

This ensures the final state matches the Terraform configuration exactly.

### Requirement 4: SSH Authentication Method Configuration

**User Story:** As a network administrator, I want to configure SSH authentication methods via Terraform, so that I can enforce security policies such as "public key only" authentication.

#### Acceptance Criteria

1. WHEN `auth_method` attribute is set on `rtx_sshd` resource THEN the system SHALL execute `sshd auth method <method>` command
2. IF `auth_method` is "publickey" THEN the system SHALL configure `sshd auth method publickey`
3. IF `auth_method` is "password" THEN the system SHALL configure `sshd auth method password`
4. IF `auth_method` is "any" (default) THEN the system SHALL allow both password and public key authentication
5. WHEN `auth_method` is changed from "any" to "publickey" THEN the system SHALL warn if no authorized keys are configured

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Each resource file handles one RTX feature area
- **Modular Design**: Reuse existing SSH client infrastructure and parser patterns
- **Dependency Management**: Leverage existing `rtx_sshd` resource patterns
- **Clear Interfaces**: Service methods in client package, parsing in parsers package

### Performance

- SSH command execution should complete within the configured timeout (default 30s)
- Key import operations may require multiple commands; total operation time proportional to key count

### Security

- Public keys in state file are not sensitive (public by definition)
- Host key fingerprints in state are acceptable; full private keys must never be stored
- Warn users when disabling password authentication without configured authorized keys

### Reliability

- Handle router responses gracefully; RTX CLI output may vary slightly between firmware versions
- Provide clear error messages when key format is invalid
- Support retry logic for transient SSH connection failures

### Usability

- Resource schemas should be intuitive for users familiar with OpenSSH `authorized_keys` format
- Documentation should include examples for common use cases (single key, multiple keys, key rotation)
- Import support allows adopting existing configurations into Terraform management

## RTX Command Reference

### Host Key Commands

```
sshd host key generate              # Generate new host key
show status sshd                    # Show host key fingerprint
```

### Authorized Keys Commands

```
import sshd authorized-keys <user>  # Add public key (interactive prompt)
show sshd authorized-keys <user>    # List registered keys
delete /ssh/authorized_keys/<user>  # Delete all keys for user
```

### Authentication Method Commands

```
sshd auth method password           # Password only
sshd auth method publickey          # Public key only
sshd auth method                    # Both (default)
no sshd auth method                 # Reset to default
```

## Resource Design Overview

### Proposed Resources

| Resource | Description | Singleton |
|----------|-------------|-----------|
| `rtx_sshd` (existing) | SSH service and listener config | Yes |
| `rtx_sshd` + `auth_method` | Extend existing resource | Yes |
| `rtx_sshd_host_key` | Host key generation | Yes |
| `rtx_sshd_authorized_keys` | Authorized keys per user | No (per user) |

### Example Usage

```hcl
# Enable SSH with public key authentication only
resource "rtx_sshd" "main" {
  enabled     = true
  hosts       = ["lan1"]
  auth_method = "publickey"
}

# Ensure host key exists
resource "rtx_sshd_host_key" "main" {}

# Configure authorized keys for admin user
resource "rtx_sshd_authorized_keys" "admin" {
  username = "admin"
  keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5... user1@workstation",
    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB... user2@laptop",
  ]
}
```
