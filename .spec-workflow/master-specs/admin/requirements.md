# Master Requirements: Admin Resources

## Overview

Admin resources manage authentication and user access control on Yamaha RTX routers. This includes router-level password configuration (`rtx_admin`) and individual user account management (`rtx_admin_user`). These resources are critical for securing router access and implementing role-based access control.

## Alignment with Product Vision

These resources support the Terraform provider's goal of enabling infrastructure-as-code management for RTX routers by:

- Allowing declarative management of router authentication settings
- Supporting multi-user environments with granular access control
- Enabling secure, auditable changes to authentication configuration
- Providing import capability for existing router configurations

## Resources Covered

### 1. rtx_admin (Singleton)

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_admin` |
| Type | Singleton |
| Import Support | Yes |
| Resource ID | Fixed: `admin` |

### 2. rtx_admin_user

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_admin_user` |
| Type | Collection |
| Import Support | Yes |
| Resource ID | Username |

---

## Functional Requirements: rtx_admin

### Core Operations

#### Create
- Sets login password and/or administrator password on the router
- Uses fixed resource ID `admin` (singleton pattern)
- Saves configuration to persistent memory after setting passwords
- Passwords are stored in Terraform state (cannot be read from router)

#### Read
- Passwords cannot be read back from the router for security reasons
- Read operation verifies the resource exists by checking ID
- State values remain as originally configured

#### Update
- Updates login password and/or administrator password
- Same implementation as Create (password set commands are idempotent)

#### Delete
- Removes both login password and administrator password
- Executes `no login password` and `no administrator password` commands
- Ignores "not found" errors during deletion

### Feature Requirements

#### REQ-ADMIN-1: Login Password Management

**User Story:** As a network administrator, I want to set a login password for the router so that unauthorized users cannot access the CLI.

**Acceptance Criteria:**
1. WHEN `login_password` is specified THEN the system SHALL execute `login password <password>`
2. WHEN `login_password` is empty THEN the system SHALL NOT modify the login password
3. IF deletion is requested THEN the system SHALL execute `no login password`

#### REQ-ADMIN-2: Administrator Password Management

**User Story:** As a network administrator, I want to set an administrator password so that configuration changes require elevated privileges.

**Acceptance Criteria:**
1. WHEN `admin_password` is specified THEN the system SHALL execute `administrator password <password>`
2. WHEN `admin_password` is empty THEN the system SHALL NOT modify the administrator password
3. IF deletion is requested THEN the system SHALL execute `no administrator password`

#### REQ-ADMIN-3: Configuration Persistence

**User Story:** As a network administrator, I want password changes to persist across router reboots.

**Acceptance Criteria:**
1. WHEN passwords are configured THEN the system SHALL save the configuration
2. IF save fails THEN the system SHALL return an error indicating partial success

#### REQ-ADMIN-4: Import Support

**User Story:** As a network administrator, I want to import existing admin configuration so I can manage it with Terraform.

**Acceptance Criteria:**
1. WHEN import is requested THEN the system SHALL set the resource ID to `admin`
2. WHEN import completes THEN the user SHALL be notified that passwords must be set in configuration
3. IF passwords are not provided after import THEN plan/apply SHALL update the state correctly

---

## Functional Requirements: rtx_admin_user

### Core Operations

#### Create
- Creates a new user account with username, password, and optional attributes
- Validates username format (alphanumeric and underscores, must start with letter)
- Validates password is provided (required for create)
- Sets user attributes if specified (administrator, connection methods, GUI pages, login timer)
- Saves configuration to persistent memory

#### Read
- Retrieves user configuration from router using `show config | grep`
- Parses `login user` and `user attribute` lines
- Updates Terraform state with current values (except password)
- Password remains in state as originally configured (cannot be read from router)
- If user not found, removes from Terraform state

#### Update
- Updates user password (overwrites existing user entry)
- Updates user attributes
- Same validation as Create

#### Delete
- Removes user attributes first (`no user attribute <username>`)
- Removes user account (`no login user <username>`)
- Ignores "not found" errors (idempotent deletion)

### Feature Requirements

#### REQ-USER-1: Username Validation

**User Story:** As a network administrator, I want username validation so that only valid usernames are accepted.

**Acceptance Criteria:**
1. WHEN username is empty THEN the system SHALL return a validation error
2. WHEN username starts with a number THEN the system SHALL return a validation error
3. WHEN username contains special characters (except underscore) THEN the system SHALL return a validation error
4. WHEN username contains only alphanumeric characters and underscores THEN the system SHALL accept it

Valid username pattern: `^[a-zA-Z][a-zA-Z0-9_]*$`

#### REQ-USER-2: Password Management

**User Story:** As a network administrator, I want to manage user passwords including encrypted passwords.

**Acceptance Criteria:**
1. WHEN password is provided as plaintext THEN the system SHALL execute `login user <username> <password>`
2. WHEN password is provided with `encrypted=true` THEN the system SHALL execute `login user <username> encrypted <password>`
3. WHEN reading from router with encrypted password THEN the system SHALL set `encrypted=true`
4. IF password is empty during create THEN the system SHALL return a validation error
5. IF password is empty during import THEN the system SHALL allow it (password provided later)

#### REQ-USER-3: Administrator Privilege

**User Story:** As a network administrator, I want to grant or deny administrator privileges to users.

**Acceptance Criteria:**
1. WHEN `administrator=true` THEN the system SHALL include `administrator=on` in user attribute command
2. WHEN `administrator=false` or not specified THEN the system SHALL include `administrator=off`
3. WHEN reading from router THEN the system SHALL correctly parse `administrator=on/off`

#### REQ-USER-4: Connection Methods

**User Story:** As a network administrator, I want to control which connection methods each user can use.

**Acceptance Criteria:**
1. WHEN `connection_methods` is specified THEN the system SHALL include `connection=<methods>` in command
2. IF invalid connection method is specified THEN the system SHALL return a validation error
3. WHEN reading from router THEN the system SHALL correctly parse comma-separated connection types

Valid connection methods: `serial`, `telnet`, `remote`, `ssh`, `sftp`, `http`

#### REQ-USER-5: GUI Page Access

**User Story:** As a network administrator, I want to control which GUI pages each user can access.

**Acceptance Criteria:**
1. WHEN `gui_pages` is specified THEN the system SHALL include `gui-page=<pages>` in command
2. IF invalid GUI page is specified THEN the system SHALL return a validation error
3. WHEN reading from router THEN the system SHALL correctly parse comma-separated GUI pages
4. WHEN `gui-page=none` is returned THEN the system SHALL set `gui_pages` to empty

Valid GUI pages: `dashboard`, `lan-map`, `config`

#### REQ-USER-6: Login Timer

**User Story:** As a network administrator, I want to set session timeout for each user.

**Acceptance Criteria:**
1. WHEN `login_timer` is specified with value > 0 THEN the system SHALL include `login-timer=<seconds>`
2. WHEN `login_timer=0` THEN the session SHALL have infinite timeout (no `login-timer` in command)
3. WHEN reading from router THEN the system SHALL correctly parse `login-timer=<seconds>`
4. IF `login_timer` is negative THEN the system SHALL return a validation error

#### REQ-USER-7: Username Immutability

**User Story:** As a network administrator, I want username changes to force resource recreation.

**Acceptance Criteria:**
1. WHEN username is changed THEN Terraform SHALL force resource recreation (`ForceNew=true`)
2. IF username is changed THEN old user SHALL be deleted and new user created

#### REQ-USER-8: Import Support

**User Story:** As a network administrator, I want to import existing users into Terraform state.

**Acceptance Criteria:**
1. WHEN import is requested with username THEN the system SHALL retrieve user configuration
2. WHEN import completes THEN all readable attributes SHALL be populated in state
3. WHEN import completes THEN user SHALL be notified that password must be set in configuration
4. IF user does not exist during import THEN the system SHALL return an error

---

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Separate files for resource definition, service layer, and parser
- **Modular Design**: Parser functions isolated from service logic
- **Dependency Management**: Service depends only on Executor interface
- **Clear Interfaces**: Well-defined AdminConfig, AdminUser, and AdminUserAttributes types

### Performance

- Password commands are non-blocking and complete quickly
- User listing uses grep filtering at router level for efficiency
- Configuration is saved once per operation, not per attribute

### Security

- **Sensitive Attributes**: All password fields marked as `Sensitive: true` in schema
- **State Storage**: Passwords stored in Terraform state (encrypted by backend)
- **Logging**: Password values sanitized in debug logs using `SanitizeCommandForLog()`
- **No Plaintext Output**: Router does not display passwords in `show config` output
- **Encrypted Passwords**: Support for pre-hashed passwords to avoid plaintext transmission

### Reliability

- **Idempotent Operations**: Password set commands are idempotent
- **Error Handling**: "not found" errors ignored during deletion
- **Context Support**: All operations respect context cancellation
- **Configuration Persistence**: Changes saved to flash memory after successful operations

### Validation

| Field | Type | Constraints |
|-------|------|-------------|
| `login_password` | string | Optional, sensitive |
| `admin_password` | string | Optional, sensitive |
| `username` | string | Required, ForceNew, pattern: `^[a-zA-Z][a-zA-Z0-9_]*$` |
| `password` | string | Optional (required for create), sensitive |
| `encrypted` | bool | Optional, computed |
| `administrator` | bool | Optional, computed |
| `connection_methods` | set(string) | Optional, values: serial/telnet/remote/ssh/sftp/http |
| `gui_pages` | set(string) | Optional, values: dashboard/lan-map/config |
| `login_timer` | int | Optional, computed, >= 0 |

---

## RTX Commands Reference

### rtx_admin Commands

```
# Set login password
login password <password>

# Remove login password
no login password

# Set administrator password
administrator password <password>

# Remove administrator password
no administrator password

# Save configuration
save
```

### rtx_admin_user Commands

```
# Create/update user with plaintext password
login user <username> <password>

# Create/update user with encrypted password
login user <username> encrypted <password>

# Delete user
no login user <username>

# Set user attributes
user attribute <username> administrator=on|off [connection=<types>] [gui-page=<pages>] [login-timer=<seconds>]

# Delete user attributes
no user attribute <username>

# Show user configuration (used for read)
# Note: RTX routers do not support grep \| OR operator
# Use separate grep commands instead
show config | grep "login user <username>"
show config | grep "user attribute <username>"

# Show all users (used for list)
show config | grep "login user"
show config | grep "user attribute"
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Detects drift from router state |
| `terraform apply` | Required | Creates/updates admin configuration |
| `terraform destroy` | Required | Removes passwords and users |
| `terraform import` | Required | Imports existing configuration |
| `terraform refresh` | Required | Updates state from router |
| `terraform state` | Required | Manages state file |

### Import Specification

#### rtx_admin
- **Import ID Format**: `admin` (fixed, any value accepted)
- **Import Command**: `terraform import rtx_admin.main admin`
- **Post-Import**: Passwords must be provided in configuration

#### rtx_admin_user
- **Import ID Format**: `<username>`
- **Import Command**: `terraform import rtx_admin_user.admin admin`
- **Post-Import**: Password must be provided in configuration; all other attributes populated from router

---

## Example Usage

### rtx_admin

```hcl
# Singleton resource for router-level passwords
resource "rtx_admin" "main" {
  login_password = var.login_password
  admin_password = var.admin_password
}
```

### rtx_admin_user

```hcl
# Basic user with password
resource "rtx_admin_user" "admin" {
  username = "admin"
  password = var.admin_user_password
}

# Administrator with full access
resource "rtx_admin_user" "netadmin" {
  username      = "netadmin"
  password      = var.netadmin_password
  administrator = true
  connection_methods = ["ssh", "telnet", "http"]
  gui_pages          = ["dashboard", "lan-map", "config"]
  login_timer        = 3600  # 1 hour timeout
}

# Limited user with HTTP-only access
resource "rtx_admin_user" "guest" {
  username      = "guest"
  password      = var.guest_password
  administrator = false
  connection_methods = ["http"]
  gui_pages          = ["dashboard"]
  login_timer        = 300  # 5 minute timeout
}

# User with encrypted password (for import scenarios)
resource "rtx_admin_user" "imported" {
  username  = "imported_user"
  password  = "$1$encryptedpasswordhash"
  encrypted = true
  administrator = false
  connection_methods = ["ssh", "sftp"]
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Passwords are stored in state (cannot be read from router)
- Operational/runtime status must not be stored in state
- `encrypted` flag indicates whether password is pre-hashed
- User attributes (administrator, connection_methods, gui_pages, login_timer) are readable from router

### State Drift Detection

| Attribute | Detectable | Notes |
|-----------|------------|-------|
| `login_password` | No | Cannot read from router |
| `admin_password` | No | Cannot read from router |
| `username` | Yes | Parsed from `login user` output |
| `password` | No | Cannot read from router |
| `encrypted` | Yes | Parsed from `login user encrypted` |
| `administrator` | Yes | Parsed from `user attribute` |
| `connection_methods` | Yes | Parsed from `user attribute` |
| `gui_pages` | Yes | Parsed from `user attribute` |
| `login_timer` | Yes | Parsed from `user attribute` |

---

## Change History

| Date | Source | Changes |
|------|--------|---------|
| 2026-01-23 | Implementation Analysis | Initial master spec created from implementation code |
| 2026-01-23 | terraform-plan-differences-fix | RTX grep compatibility (no OR operator); attribute-only updates for imported users without passwords |
