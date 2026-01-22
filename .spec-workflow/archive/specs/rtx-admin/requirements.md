# Requirements: rtx_admin

## Overview
Terraform resources for managing administrative access settings on Yamaha RTX routers. This includes login passwords, administrator passwords, and user account management.

**Security Note**: This resource manages sensitive credentials. Special care must be taken to handle passwords securely.

## Covered Resources

This specification covers two Terraform resources:

- **`rtx_admin`**: Login and administrator password settings
- **`rtx_admin_user`**: User account creation and attributes

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure administrative access settings
- **Read**: Query current admin configuration (passwords cannot be read back)
- **Update**: Modify passwords and user accounts
- **Delete**: Reset to default (remove custom passwords/users)

### 2. Login Password Management
- Configure console login password
- Support plaintext and encrypted password formats

### 3. Administrator Password Management
- Configure privileged mode (administrator) password
- Support plaintext and encrypted password formats

### 4. User Account Management
- Create named user accounts
- Configure user passwords
- Configure user attributes

### 5. User Attributes
- Administrator privilege flag
- Allowed connection methods (serial, telnet, remote, ssh, sftp, http)
- Allowed GUI pages (dashboard, lan-map, config)
- Login timer (session timeout)

### 6. Import Support
- Import existing admin configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned admin configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete admin settings |
| `terraform destroy` | ✅ Required | Remove custom admin configuration |
| `terraform import` | ✅ Required | Import existing admin settings into state |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `admin` (singleton resource)
- **Import Command**: `terraform import rtx_admin.main admin`
- **Post-Import**: Passwords cannot be imported (use `ignore_changes` lifecycle)

## Non-Functional Requirements

### 7. Validation
- Validate username format
- Validate connection types are valid
- Validate GUI pages are valid
- Validate login timer is positive integer

### 8. Security
- Mark all passwords as sensitive
- Never log or display passwords
- Support state file encryption recommendations

## RTX Commands Reference
```
login password <password>
administrator password <password>
login user <username> <password>
login user <username> encrypted <encrypted_password>
user attribute <username> administrator=on|off connection=<types> gui-page=<pages> login-timer=<seconds>
no login user <username>
no user attribute <username>
```

## Example Usage
```hcl
# Admin passwords only (singleton)
resource "rtx_admin" "main" {
  login_password = var.login_password
  admin_password = var.admin_password
}

# User accounts (separate resource)
resource "rtx_admin_user" "operator" {
  username = "operator"
  password = var.operator_password

  attributes {
    administrator = false
    connection    = ["ssh", "http"]
    gui_pages     = ["dashboard", "lan-map"]
    login_timer   = 300
  }
}

resource "rtx_admin_user" "admin_user" {
  username = "admin_user"
  password = var.admin_user_password

  attributes {
    administrator = true
    connection    = ["serial", "telnet", "remote", "ssh", "sftp", "http"]
    gui_pages     = ["dashboard", "lan-map", "config"]
    login_timer   = 3600
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
