# Requirements: rtx_syslog

## Overview
Terraform resource for managing syslog configuration on Yamaha RTX routers. Syslog is essential for centralized logging, monitoring, and compliance in enterprise environments.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure syslog settings
- **Read**: Query current syslog configuration
- **Update**: Modify syslog parameters
- **Delete**: Remove syslog configuration

### 2. Remote Syslog Server Configuration
- Configure one or more remote syslog servers
- Specify server IP address or hostname
- Configure custom UDP port (default: 514)

### 3. Local Address Configuration
- Specify source IP for syslog packets
- Use specific interface IP as source

### 4. Facility Configuration
- Configure syslog facility (local0-local7, user, daemon, etc.)
- Standard syslog facility support

### 5. Log Level Configuration
- Enable/disable notice level logs
- Enable/disable info level logs
- Enable/disable debug level logs (verbose)

### 6. Import Support
- Import existing syslog configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned syslog configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete syslog settings |
| `terraform destroy` | ✅ Required | Remove syslog configuration |
| `terraform import` | ✅ Required | Import existing syslog settings |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `syslog` (singleton resource)
- **Import Command**: `terraform import rtx_syslog.main syslog`
- **Post-Import**: All attributes must be populated

## Non-Functional Requirements

### 7. Validation
- Validate IP address or hostname format
- Validate port is in range 1-65535
- Validate facility is in allowed list

### 8. Security
- Document that syslog content may contain sensitive information
- Recommend secure transport if available

## RTX Commands Reference
```
syslog host <address>
syslog host <address> <port>
syslog local address <ip_address>
syslog facility <facility>
syslog notice on|off
syslog info on|off
syslog debug on|off
```

## Example Usage
```hcl
resource "rtx_syslog" "main" {
  # Remote syslog servers
  host {
    address = "192.168.1.20"
  }

  host {
    address = "192.168.1.21"
    port    = 1514
  }

  # Source address for syslog packets
  local_address = "192.168.1.253"

  # Syslog facility
  facility = "local0"

  # Log levels
  notice = true
  info   = true
  debug  = true
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
