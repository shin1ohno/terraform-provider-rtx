# Design Document: Administrator Attribute Model Support

## Overview

This design addresses the model-specific differences in RTX router `administrator` attribute values. The implementation will support both `on`/`off` (Model Group A) and `1`/`2`/`off` (Model Group B) formats while maintaining backward compatibility.

## Steering Document Alignment

### Technical Standards (tech.md)
- Follow existing parser patterns in `internal/rtx/parsers/`
- Use zerolog for logging model-specific behavior
- Maintain consistent error handling patterns

### Project Structure (structure.md)
- Parser changes in `internal/rtx/parsers/admin.go`
- Provider schema changes in `internal/provider/resource_rtx_admin_user.go`
- Model configuration in provider config or dedicated helper

## Code Reuse Analysis

### Existing Components to Leverage
- **`parsers.UserAttributes`**: Extend with new field for password-required distinction
- **`parsers.BuildUserAttributeCommand`**: Modify to accept model hint
- **`parseUserAttributeString`**: Extend to recognize `1` and `2` values

### Integration Points
- **Provider Configuration**: May add optional `rtx_model` or `administrator_format` setting
- **Resource Schema**: Add optional `administrator_password_required` attribute

## Architecture

### Design Decision: Model Format Handling

**Option A: Provider-level model configuration**
```hcl
provider "rtx" {
  administrator_format = "numeric"  # or "boolean"
}
```

**Option B: Resource-level attribute (Recommended)**
```hcl
resource "rtx_admin_user" "user" {
  username      = "admin"
  administrator = true
  administrator_password_required = true  # Results in "1" for Group B models
}
```

**Rationale for Option B:**
- More flexible (can manage users on different model routers)
- Explicit configuration reduces ambiguity
- No need to know the model in advance
- Backward compatible (new attribute is optional)

### Data Flow

```
Terraform Config                    Parser                         RTX Router
      │                               │                                 │
      │  administrator=true           │                                 │
      │  admin_password_required=true │                                 │
      │                               │                                 │
      ▼                               │                                 │
┌─────────────────┐                   │                                 │
│ Schema Validate │                   │                                 │
└────────┬────────┘                   │                                 │
         │                            │                                 │
         ▼                            │                                 │
┌─────────────────┐                   │                                 │
│ BuildCommand    │──────────────────►│ "user attribute x admin=1"     │
└─────────────────┘                   │         or "admin=on"          │
                                      │                                 │
                                      │                                 ▼
                                      │                          ┌──────────┐
                                      │◄─────────────────────────│ Response │
                                      │  "admin=1" or "admin=on" └──────────┘
                                      │                                 │
                                      ▼                                 │
                                ┌───────────┐                           │
                                │ Parse     │                           │
                                │ Response  │                           │
                                └─────┬─────┘                           │
                                      │                                 │
                                      │ administrator=true              │
                                      │ admin_password_required=true    │
                                      ▼                                 │
                              ┌──────────────┐                          │
                              │ Terraform    │                          │
                              │ State        │                          │
                              └──────────────┘                          │
```

## Components and Interfaces

### Component 1: Enhanced UserAttributes

**Purpose:** Store administrator attribute with password-required distinction

**Current:**
```go
type UserAttributes struct {
    Administrator *bool    `json:"administrator"`
    Connection    []string `json:"connection"`
    GUIPages      []string `json:"gui_pages"`
    LoginTimer    *int     `json:"login_timer"`
}
```

**Proposed:**
```go
type UserAttributes struct {
    Administrator                *bool    `json:"administrator"`
    AdministratorPasswordRequired *bool   `json:"administrator_password_required"` // New field
    Connection                   []string `json:"connection"`
    GUIPages                     []string `json:"gui_pages"`
    LoginTimer                   *int     `json:"login_timer"`
}
```

### Component 2: Enhanced Parser

**Purpose:** Parse all administrator value formats

**File:** `internal/rtx/parsers/admin.go`

**Changes to `parseUserAttributeString`:**
```go
func parseUserAttributeString(attrStr string) UserAttributes {
    // ...existing code...

    if strings.HasPrefix(part, "administrator=") {
        value := strings.TrimPrefix(part, "administrator=")
        switch value {
        case "on", "2":
            isAdmin := true
            attrs.Administrator = &isAdmin
            if value == "2" {
                pwdRequired := false
                attrs.AdministratorPasswordRequired = &pwdRequired
            }
        case "1":
            isAdmin := true
            pwdRequired := true
            attrs.Administrator = &isAdmin
            attrs.AdministratorPasswordRequired = &pwdRequired
        case "off":
            isAdmin := false
            attrs.Administrator = &isAdmin
        }
    }
    // ...
}
```

### Component 3: Enhanced Command Builder

**Purpose:** Generate model-appropriate commands

**File:** `internal/rtx/parsers/admin.go`

**Changes to `BuildUserAttributeCommand`:**
```go
// AdministratorFormat specifies the command format for administrator attribute
type AdministratorFormat string

const (
    AdminFormatBoolean AdministratorFormat = "boolean" // on/off (default)
    AdminFormatNumeric AdministratorFormat = "numeric" // 1/2/off
)

func BuildUserAttributeCommandWithFormat(username string, attrs UserAttributes, format AdministratorFormat) string {
    var parts []string

    if attrs.Administrator != nil {
        if *attrs.Administrator {
            if format == AdminFormatNumeric {
                if attrs.AdministratorPasswordRequired != nil && *attrs.AdministratorPasswordRequired {
                    parts = append(parts, "administrator=1")
                } else {
                    parts = append(parts, "administrator=2")
                }
            } else {
                parts = append(parts, "administrator=on")
            }
        } else {
            parts = append(parts, "administrator=off")
        }
    }
    // ... rest of attributes
}

// Backward compatible wrapper
func BuildUserAttributeCommand(username string, attrs UserAttributes) string {
    return BuildUserAttributeCommandWithFormat(username, attrs, AdminFormatBoolean)
}
```

### Component 4: Terraform Resource Schema Enhancement

**Purpose:** Allow users to specify administrator password requirement

**File:** `internal/provider/resource_rtx_admin_user.go`

**Schema Addition:**
```go
"administrator_password_required": schema.BoolAttribute{
    Optional:    true,
    Computed:    true,
    Description: "Whether administrator password is required for elevation. Only applicable when administrator=true. Affects Model Group B routers (RTX840, RTX1300, RTX3510).",
},
```

## Data Models

### AdminUser Resource Model (Updated)
```go
type AdminUserResourceModel struct {
    ID                           types.String `tfsdk:"id"`
    Username                     types.String `tfsdk:"username"`
    Password                     types.String `tfsdk:"password"`
    Encrypted                    types.Bool   `tfsdk:"encrypted"`
    Administrator                types.Bool   `tfsdk:"administrator"`
    AdministratorPasswordRequired types.Bool   `tfsdk:"administrator_password_required"` // New
    Connection                   types.List   `tfsdk:"connection"`
    GUIPages                     types.List   `tfsdk:"gui_pages"`
    LoginTimer                   types.Int64  `tfsdk:"login_timer"`
}
```

## Error Handling

### Error Scenarios

1. **Invalid administrator value in router response:**
   - **Handling:** Log warning, treat as `off`
   - **User Impact:** May see unexpected state, log provides explanation

2. **administrator_password_required=true but administrator=false:**
   - **Handling:** Ignore administrator_password_required (no effect when admin is off)
   - **User Impact:** No error, attribute simply not used

## Testing Strategy

### Unit Testing

1. **Parser Tests** (`admin_test.go`):
   - Parse `administrator=on` → Administrator=true
   - Parse `administrator=1` → Administrator=true, AdministratorPasswordRequired=true
   - Parse `administrator=2` → Administrator=true, AdministratorPasswordRequired=false
   - Parse `administrator=off` → Administrator=false

2. **Command Builder Tests**:
   - Boolean format: true → "administrator=on"
   - Numeric format: true + pwd_required=true → "administrator=1"
   - Numeric format: true + pwd_required=false → "administrator=2"
   - Both formats: false → "administrator=off"

### Integration Testing

1. **Import Test**: Import user with `administrator=1` or `administrator=2`
2. **Round-trip Test**: Create user, read back, verify no plan diff

### Acceptance Testing

1. Test on Model Group A router (RTX1210/RTX830)
2. Test on Model Group B router (RTX1300/RTX840) if available

## Migration Guide

### For Existing Users

No action required. Existing configurations will continue to work:
- `administrator = true` generates `administrator=on` (same as before)
- `administrator = false` generates `administrator=off` (same as before)

### For Model Group B Users

To take advantage of the password-less elevation feature:
```hcl
resource "rtx_admin_user" "superadmin" {
  username                        = "superadmin"
  password                        = "secret"
  administrator                   = true
  administrator_password_required = false  # Generates "administrator=2"
}
```

## Implementation Order

1. Update `UserAttributes` struct with new field
2. Update `parseUserAttributeString` to handle `1`/`2` values
3. Add `BuildUserAttributeCommandWithFormat` function
4. Update Terraform resource schema
5. Update resource CRUD operations
6. Add tests
7. Update documentation
