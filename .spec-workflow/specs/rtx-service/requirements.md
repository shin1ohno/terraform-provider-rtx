# Requirements: rtx_service (rtx_httpd, rtx_sshd, rtx_sftpd)

## Overview
Terraform resources for managing network services on Yamaha RTX routers. This includes HTTPD (web interface), SSHD (SSH access), and SFTPD (SFTP file transfer) services.

**Note**: This specification covers three separate Terraform resources: `rtx_httpd`, `rtx_sshd`, and `rtx_sftpd`.

## Covered Resources

This specification covers three Terraform resources:

- **`rtx_httpd`**: Web interface (HTTPD) configuration
- **`rtx_sshd`**: SSH daemon configuration
- **`rtx_sftpd`**: SFTP daemon configuration

## Functional Requirements

### 1. HTTPD (Web Interface) - `rtx_httpd`

#### 1.1 CRUD Operations
- **Create**: Enable and configure HTTP server
- **Read**: Query HTTPD configuration
- **Update**: Modify HTTPD settings
- **Delete**: Disable HTTP server

#### 1.2 Host Configuration
- Listen on all interfaces (`any`)
- Listen on specific interface

#### 1.3 Proxy Access
- Enable/disable L2MS proxy access
- Yamaha LAN Monitor System integration

### 2. SSHD (SSH Server) - `rtx_sshd`

#### 2.1 CRUD Operations
- **Create**: Enable and configure SSH daemon
- **Read**: Query SSHD configuration
- **Update**: Modify SSHD settings
- **Delete**: Disable SSH daemon

#### 2.2 Service Control
- Enable/disable SSH service

#### 2.3 Host Configuration
- Specify interfaces to listen on
- Multiple interface support

#### 2.4 Host Key Management
- Generate or provide RSA host key
- Key fingerprint storage

### 3. SFTPD (SFTP Server) - `rtx_sftpd`

#### 3.1 CRUD Operations
- **Create**: Enable and configure SFTP daemon
- **Read**: Query SFTPD configuration
- **Update**: Modify SFTPD settings
- **Delete**: Disable SFTP daemon

#### 3.2 Host Configuration
- Specify interfaces to listen on
- Multiple interface support

### 4. Import Support
- Import existing service configurations

## Terraform Command Support

All three resources must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned service configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete service settings |
| `terraform destroy` | ✅ Required | Disable service |
| `terraform import` | ✅ Required | Import existing service configuration |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `httpd`, `sshd`, or `sftpd` (singleton resources)
- **Import Command**: `terraform import rtx_sshd.main sshd`
- **Post-Import**: All attributes must be populated

## Non-Functional Requirements

### 5. Validation
- Validate interface names
- Validate "any" keyword for HTTPD host
- Validate host key format

### 6. Security
- Mark SSH host key as sensitive
- Document security implications of service exposure
- Warn about interface restrictions

## RTX Commands Reference
```
# HTTPD
httpd host any
httpd host <interface>
httpd proxy-access l2ms permit on|off

# SSHD
sshd service on|off
sshd host <interface1> [<interface2> ...]
sshd host key generate <key_length> <key_data>

# SFTPD
sftpd host <interface1> [<interface2> ...]
```

## Example Usage
```hcl
# HTTPD (Web Interface)
resource "rtx_httpd" "main" {
  host         = "any"
  proxy_access = true
}

# SSHD (SSH Server)
resource "rtx_sshd" "main" {
  enabled = true
  hosts   = ["lan2", "bridge1"]
}

# SFTPD (SFTP Server)
resource "rtx_sftpd" "main" {
  hosts = ["bridge1"]
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
