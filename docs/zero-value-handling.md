# Zero Value Handling Patterns in Terraform Provider SDK v2

This document explains the differences between `Get`, `GetOk`, and `GetOkExists` methods in terraform-plugin-sdk/v2, and provides guidance on when to use each.

## The Zero Value Problem

In Go, variables have "zero values" by default:
- `string`: `""`
- `int`: `0`
- `bool`: `false`

When working with Terraform schemas, distinguishing between "user explicitly set a zero value" versus "user did not set any value" is crucial for correct behavior.

## Methods Comparison

### `d.Get(key)`

**Returns:** The value for the key (never nil for primitive types)

**Behavior:**
- Returns the merged value from config and state
- For unset optional fields, returns the Go zero value
- Cannot distinguish between "not set" and "explicitly set to zero value"

```go
// Schema
"timeout": {
    Type:     schema.TypeInt,
    Optional: true,
}

// Usage
timeout := d.Get("timeout").(int)  // Returns 0 if not set OR if set to 0
```

**When to use:**
- When you need the final merged value
- When zero value is acceptable as a default
- For Required fields (always have a value)
- For Computed fields during Read operations

### `d.GetOk(key)`

**Returns:** `(value, ok bool)` where `ok` is `true` if the value was set AND is non-zero

**Behavior:**
- Returns `false` for both unset values AND zero values
- Cannot distinguish between "not set" and "set to zero"

```go
// Schema
"description": {
    Type:     schema.TypeString,
    Optional: true,
}

// Usage
if v, ok := d.GetOk("description"); ok {
    // User provided a non-empty description
    config.Description = v.(string)
}
// Note: Empty string "" will have ok=false
```

**When to use:**
- For optional string/list/set fields where empty means "not configured"
- For optional int fields where 0 is not a valid value
- Most common pattern for Optional fields

### `d.GetOkExists(key)` (Deprecated but sometimes necessary)

**Returns:** `(value, exists bool)` where `exists` is `true` if the value was set in config

**Behavior:**
- Returns `true` even if the value is a zero value
- Can distinguish between "not set" and "explicitly set to zero"

```go
// Schema
"enabled": {
    Type:     schema.TypeBool,
    Optional: true,
    Computed: true,  // API provides default
}

// Usage
if v, exists := d.GetOkExists("enabled"); exists {
    // User explicitly set the value (could be true or false)
    config.Enabled = v.(bool)
} else {
    // User did not set - use API default or preserve state
}
```

**When to use:**
- For bool fields where `false` is a meaningful, explicit choice
- For int fields where `0` is a valid, meaningful value
- When you need to distinguish "unset" from "zero value"

> **Note:** `GetOkExists` is deprecated in newer SDK versions but remains necessary for certain use cases until migration to terraform-plugin-framework.

## Examples by Field Type

### String Fields

```go
// Pattern A: Empty string means "not configured" (most common)
"description": {
    Type:     schema.TypeString,
    Optional: true,
},

// In CRUD function
if v, ok := d.GetOk("description"); ok {
    config.Description = v.(string)
}
```

```go
// Pattern B: Preserve empty string from state (Optional+Computed)
"hostname": {
    Type:     schema.TypeString,
    Optional: true,
    Computed: true,  // API can provide default
},

// In Update function
hostname := d.Get("hostname").(string)  // Gets merged value from config or state
config.Hostname = hostname
```

### Integer Fields

```go
// Pattern A: 0 means "not configured" (most common)
"port": {
    Type:     schema.TypeInt,
    Optional: true,
},

// In CRUD function
if v, ok := d.GetOk("port"); ok {
    config.Port = v.(int)
}
```

```go
// Pattern B: 0 is a valid value (less common)
"priority": {
    Type:     schema.TypeInt,
    Optional: true,
    Default:  -1,  // Use sentinel value if 0 is valid
},

// In CRUD function
priority := d.Get("priority").(int)
if priority != -1 {
    config.Priority = priority
}
```

```go
// Pattern C: Using GetOkExists when 0 is meaningful
"retry_count": {
    Type:     schema.TypeInt,
    Optional: true,
    Computed: true,
},

// In CRUD function
if v, exists := d.GetOkExists("retry_count"); exists {
    // User explicitly set a value (including 0 = disable retries)
    config.RetryCount = v.(int)
} else {
    // Use API default
    config.RetryCount = d.Get("retry_count").(int)
}
```

### Boolean Fields

```go
// Pattern A: Optional+Computed bool (API provides default)
"administrator": {
    Type:     schema.TypeBool,
    Optional: true,
    Computed: true,
    Description: "Whether user has admin privileges",
},

// In Create/Update function - explicit false is meaningful
if v, exists := d.GetOkExists("administrator"); exists {
    // User explicitly set true or false
    config.Administrator = v.(bool)
} else {
    // Not set - preserve state or let API decide
    config.Administrator = d.Get("administrator").(bool)
}
```

```go
// Pattern B: Simple bool with default (no GetOkExists needed)
"enabled": {
    Type:     schema.TypeBool,
    Optional: true,
    Default:  true,  // Explicit default handles the zero value
},

// In CRUD function
config.Enabled = d.Get("enabled").(bool)  // Default handles unset case
```

```go
// Pattern C: Using GetOk (explicit false treated as unset)
"debug_mode": {
    Type:     schema.TypeBool,
    Optional: true,
},

// In CRUD function
if v, ok := d.GetOk("debug_mode"); ok {
    // Only true triggers this block
    config.DebugMode = v.(bool)
}
// false and unset both result in ok=false
```

## Real-World Examples from This Provider

### Example 1: Password fields with Optional strings

```go
// From resource_rtx_admin.go
func buildAdminConfigFromResourceData(d *schema.ResourceData) client.AdminConfig {
    return client.AdminConfig{
        LoginPassword: d.Get("login_password").(string),
        AdminPassword: d.Get("admin_password").(string),
    }
}

// Empty string correctly means "no password change"
```

### Example 2: Optional nested blocks with GetOk

```go
// From resource_rtx_syslog.go
if v, ok := d.GetOk("local_address"); ok {
    config.LocalAddress = v.(string)
}
if v, ok := d.GetOk("host"); ok {
    config.Hosts = expandSyslogHosts(v.([]interface{}))
}
```

### Example 3: Conditional fields

```go
// From resource_rtx_dhcp_binding.go
if macAddress, ok := d.GetOk("mac_address"); ok {
    config.MACAddress = macAddress.(string)
} else if clientID, ok := d.GetOk("client_identifier"); ok {
    config.ClientIdentifier = clientID.(string)
}
```

## Decision Tree

```
Is the field Required?
├── YES → Use d.Get() directly
└── NO (Optional)
    ├── Is it Computed (API provides default)?
    │   ├── YES
    │   │   ├── Is zero value meaningful? (bool=false, int=0)
    │   │   │   ├── YES → Use d.GetOkExists() or pointer types
    │   │   │   └── NO → Use d.GetOk() for non-zero, d.Get() for merged
    │   │   └── (preserve state value when not in config)
    │   └── NO
    │       ├── Is zero value meaningful?
    │       │   ├── YES → Use Default schema option or GetOkExists
    │       │   └── NO → Use d.GetOk()
    │       └── (no API default, user must specify or accept zero)
    └── Consider using pointer types in internal structs
        for cleaner nil vs zero value handling
```

## Migration Notes for terraform-plugin-framework

When migrating to terraform-plugin-framework, the zero value handling changes:

| SDK v2 | Plugin Framework |
|--------|------------------|
| `d.Get()` | `data.FieldName.ValueString()` etc. |
| `d.GetOk()` | `!data.FieldName.IsNull()` with zero check |
| `d.GetOkExists()` | `!data.FieldName.IsNull()` |

The framework's null handling is more explicit:
- `IsNull()` - returns true if not set in config
- `IsUnknown()` - returns true if value not yet known (planning)
- `ValueString()`, `ValueInt64()`, etc. - returns the actual value

This eliminates most zero value ambiguity issues by design.

## Best Practices

1. **Prefer Optional+Computed** for fields with API defaults to preserve state values

2. **Use pointer types in internal structs** to represent optional fields:
   ```go
   type Config struct {
       Timeout *int  // nil means not specified
   }
   ```

3. **Document zero value semantics** in schema descriptions:
   ```go
   "timeout": {
       Type:        schema.TypeInt,
       Optional:    true,
       Description: "Timeout in seconds. Set to 0 to disable timeout.",
   }
   ```

4. **Use Default when appropriate** to avoid GetOkExists:
   ```go
   "enabled": {
       Type:     schema.TypeBool,
       Optional: true,
       Default:  true,
   }
   ```

5. **Test zero value behavior** explicitly in acceptance tests
