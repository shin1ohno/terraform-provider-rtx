# Terraform Schema Design Patterns

This guide provides comprehensive patterns for designing Terraform provider schemas using terraform-plugin-sdk/v2, with migration notes for terraform-plugin-framework.

## Table of Contents

- [Attribute Configurability Decision Tree](#attribute-configurability-decision-tree)
- [Pattern Reference Table](#pattern-reference-table)
- [Pattern Details](#pattern-details)
  - [Required Attributes](#required-attributes)
  - [Optional Attributes](#optional-attributes)
  - [Computed Attributes](#computed-attributes)
  - [Optional+Computed Attributes](#optionalcomputed-attributes)
  - [WriteOnly/Sensitive Attributes](#writeonlysensitive-attributes)
  - [RequiresReplace (ForceNew) Attributes](#requiresreplace-forcenew-attributes)
- [Diff Suppression Patterns](#diff-suppression-patterns)
- [State Normalization Patterns](#state-normalization-patterns)
- [Nested Block Patterns](#nested-block-patterns)
- [State Migration Patterns](#state-migration-patterns)
- [Plugin Framework Migration Notes](#plugin-framework-migration-notes)

## Attribute Configurability Decision Tree

Use this decision tree to determine the correct attribute configuration:

```
Is the value always required from the user?
├── YES → Required: true
│
└── NO → Is the value ever provided by the user?
    ├── NO → Computed: true (read-only)
    │
    └── YES → Can the API/router provide a default?
        ├── YES → Optional: true, Computed: true
        │   └── Is the value immutable after creation?
        │       ├── YES → Add ForceNew: true
        │       └── NO → (no additional flag)
        │
        └── NO → Optional: true
            └── Is the value immutable after creation?
                ├── YES → Add ForceNew: true
                └── NO → (no additional flag)

Additionally:
├── Is it sensitive (password, key, token)?
│   ├── YES → Add Sensitive: true
│   └── NO → (no additional flag)
│
└── Is it write-only (credentials that cannot be read back)?
    ├── YES → Add Sensitive: true (SDK v2) or WriteOnly: true (Framework)
    └── NO → (no additional flag)
```

## Pattern Reference Table

| Use Case | Required | Optional | Computed | ForceNew | Sensitive |
|----------|:--------:|:--------:|:--------:|:--------:|:---------:|
| User must specify | X | | | | |
| User may specify, no default | | X | | | |
| Read-only from API | | | X | | |
| User may override API default | | X | X | | |
| Immutable identifier | X | | | X | |
| Optional immutable field | | X | | X | |
| Password/credential field | | X | | | X |
| Generated ID/timestamp | | | X | | |
| API key (readable) | | X | | | X |
| Generated secret | | | X | | X |

## Pattern Details

### Required Attributes

Use for values that must always be provided by the user.

```go
// From internal/provider/resource_rtx_admin_user.go
"username": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "Username for the admin user (cannot be changed after creation)",
},
```

**When to use:**
- Primary identifiers (username, name, ID)
- Essential configuration that has no sensible default
- Values the router/API cannot auto-generate

**SDK v2 Helper:**
```go
// From internal/provider/schema_helpers.go
ImmutableStringSchema(description string) *schema.Schema
```

### Optional Attributes

Use for values the user may provide but are not required.

```go
"description": {
    Type:        schema.TypeString,
    Optional:    true,
    Description: "Optional description for the user",
},
```

**When to use:**
- Descriptive fields (description, comment)
- Optional configuration with no API-provided default
- Fields where omission means "not configured"

### Computed Attributes

Use for read-only values populated by the provider or API.

```go
"created_at": {
    Type:        schema.TypeString,
    Computed:    true,
    Description: "Timestamp when the resource was created",
},
```

**When to use:**
- Generated identifiers
- Timestamps (created_at, updated_at)
- Values derived from other attributes
- Status or state information

**SDK v2 Helper:**
```go
// From internal/provider/schema_helpers.go
SensitiveComputedStringSchema(description string) *schema.Schema
```

### Optional+Computed Attributes

Use when the user may provide a value, but the API can provide a default if omitted.

```go
// From internal/provider/schema_helpers.go example
"administrator": {
    Type:        schema.TypeBool,
    Optional:    true,
    Computed:    true,
    Description: "Whether user has admin privileges. Defaults to false if not specified.",
},
```

**When to use:**
- Fields with API-provided defaults
- Configuration where the router assigns a value if user doesn't specify
- Settings that can be overridden but have sensible defaults

**SDK v2 Helpers:**
```go
// From internal/provider/schema_helpers.go
OptionalComputedStringSchema(description string) *schema.Schema
OptionalComputedBoolSchema(description string) *schema.Schema
OptionalComputedIntSchema(description string) *schema.Schema
```

**Important:** With Optional+Computed, Terraform preserves the existing state value when the user doesn't specify a value in configuration.

### WriteOnly/Sensitive Attributes

Use for credentials and secrets that should not be stored in state or displayed in output.

```go
// From internal/provider/schema_helpers.go
"password": WriteOnlyStringSchema("Password for the admin user"),

// Which expands to:
"password": {
    Type:        schema.TypeString,
    Optional:    true,
    Sensitive:   true,
    Description: "Password for the admin user (write-only: value is sent to device but not read back)",
},
```

**SDK v2 Limitation:** True write-only is not supported in SDK v2. The value is still stored in state (encrypted by Terraform). To achieve write-only-like behavior:

1. Use `Sensitive: true` to prevent display in plan/apply output
2. Do NOT read this value back from the router in the Read function
3. Preserve the value from state during Read operations

**SDK v2 Helpers:**
```go
// From internal/provider/schema_helpers.go
WriteOnlyStringSchema(description string) *schema.Schema       // Optional write-only
WriteOnlyRequiredStringSchema(description string) *schema.Schema  // Required write-only
SensitiveStringSchema(description string, required bool) *schema.Schema  // Readable sensitive
```

### RequiresReplace (ForceNew) Attributes

Use for immutable fields that cannot be changed after resource creation.

```go
// From internal/provider/resource_rtx_admin_user.go
"username": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,  // SDK v2 way to mark as RequiresReplace
    Description: "Username for the admin user (cannot be changed after creation)",
},
```

**When to use:**
- Primary resource identifiers
- Fields that the router/API cannot modify after creation
- Configuration that requires recreation to change

**SDK v2 Helpers:**
```go
// From internal/provider/schema_helpers.go
ImmutableStringSchema(description string) *schema.Schema        // Required immutable
ImmutableOptionalStringSchema(description string) *schema.Schema  // Optional immutable
```

**Conditional RequiresReplace:** Use `CustomizeDiff` for conditional replacement:

```go
func resourceCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
    if d.HasChange("type") {
        old, new := d.GetChange("type")
        if old.(string) == "local" && new.(string) == "remote" {
            d.ForceNew("type")
        }
    }
    return nil
}
```

## Diff Suppression Patterns

Use `DiffSuppressFunc` to suppress irrelevant diffs when values are semantically equal.

### Available Functions

All functions are defined in `internal/provider/diff_suppress.go`:

| Function | Purpose | Example |
|----------|---------|---------|
| `SuppressCaseDiff` | Case-insensitive comparison | "TCP" = "tcp" |
| `SuppressWhitespaceDiff` | Ignore leading/trailing whitespace | "  value  " = "value" |
| `SuppressJSONDiff` | Semantic JSON comparison | `{"a":1,"b":2}` = `{"b":2,"a":1}` |
| `SuppressEquivalentIPDiff` | IP address equivalence | "::1" = "0:0:0:0:0:0:0:1" |
| `SuppressEquivalentCIDRDiff` | CIDR block equivalence | Normalizes network addresses |
| `SuppressCaseAndWhitespaceDiff` | Combined case and whitespace | "  VALUE  " = "value" |
| `SuppressBooleanStringDiff` | Boolean string comparison | "TRUE" = "true" |

### Usage Example

```go
"protocol": {
    Type:             schema.TypeString,
    Optional:         true,
    DiffSuppressFunc: SuppressCaseDiff,
    Description:      "Protocol name (case-insensitive)",
},

"config_json": {
    Type:             schema.TypeString,
    Optional:         true,
    DiffSuppressFunc: SuppressJSONDiff,
    Description:      "JSON configuration (key order does not matter)",
},

"ip_address": {
    Type:             schema.TypeString,
    Optional:         true,
    StateFunc:        normalizeIPAddress,
    DiffSuppressFunc: SuppressEquivalentIPDiff,
    ValidateFunc:     validation.IsIPAddress,
    Description:      "IP address in any valid format",
},
```

### Best Practices

1. **Prefer StateFunc over DiffSuppressFunc** when possible - normalize values on storage
2. **Combine with validation** - ensure invalid values are rejected before comparison
3. **Safe fallback** - all suppress functions return `false` on error (showing the diff)
4. **Document the behavior** - update description to explain equivalence rules

## State Normalization Patterns

Use `StateFunc` to normalize values before storing in state, preventing unnecessary diffs.

### Available Normalizers

All functions are defined in `internal/provider/state_funcs.go`:

| Function | Purpose | Example |
|----------|---------|---------|
| `normalizeIPAddress` | Canonical IP format | "192.168.001.001" -> "192.168.1.1" |
| `normalizeLowercase` | Lowercase conversion | "HELLO" -> "hello" |
| `normalizeUppercase` | Uppercase conversion | "hello" -> "HELLO" |
| `normalizeJSON` | Consistent JSON formatting | Sorted keys, compact format |
| `normalizeTrimmedString` | Remove whitespace | "  value  " -> "value" |

### Usage Example

```go
"hostname": {
    Type:      schema.TypeString,
    Optional:  true,
    StateFunc: normalizeLowercase,
    Description: "Hostname (stored in lowercase)",
},

"ip_address": {
    Type:             schema.TypeString,
    Optional:         true,
    StateFunc:        normalizeIPAddress,
    DiffSuppressFunc: SuppressEquivalentIPDiff,
    Description:      "IP address (stored in canonical format)",
},
```

### Best Practices

1. **Handle invalid input gracefully** - return input unchanged if normalization fails
2. **Let validation handle errors** - StateFunc should not fail
3. **Ensure idempotency** - normalizing twice should produce the same result
4. **Combine with DiffSuppressFunc** - for extra safety during comparison

## Nested Block Patterns

### List Nested Block (Ordered)

Use when element order matters:

```go
"filter_rules": {
    Type:     schema.TypeList,
    Optional: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "id": {
                Type:     schema.TypeInt,
                Required: true,
            },
            "action": {
                Type:     schema.TypeString,
                Required: true,
            },
        },
    },
},
```

### Set Nested Block (Unordered)

Use when elements are unique and order doesn't matter:

```go
"allowed_hosts": {
    Type:     schema.TypeSet,
    Optional: true,
    Elem: &schema.Schema{
        Type: schema.TypeString,
    },
},
```

### Single Nested Block

Use when exactly one nested object exists:

```go
"timeouts": {
    Type:     schema.TypeList,
    Optional: true,
    MaxItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "create": {Type: schema.TypeString, Optional: true},
            "update": {Type: schema.TypeString, Optional: true},
            "delete": {Type: schema.TypeString, Optional: true},
        },
    },
},
```

## State Migration Patterns

When schema changes require state migration, use `SchemaVersion` and `StateUpgraders`.

### Example State Upgrader

```go
// From internal/provider/state_upgraders/upgraders.go

func resourceAdminUser() *schema.Resource {
    return &schema.Resource{
        SchemaVersion: 1,
        StateUpgraders: []schema.StateUpgrader{
            {
                Version: 0,
                Type:    ResourceAdminUserV0().CoreConfigSchema().ImpliedType(),
                Upgrade: StateUpgradeAdminUserV0,
            },
        },
        Schema: currentSchema,
    }
}

func StateUpgradeAdminUserV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
    if rawState == nil {
        return nil, nil
    }

    // Rename "name" to "username"
    if v, ok := rawState["name"]; ok {
        rawState["username"] = v
        delete(rawState, "name")
    }

    return rawState, nil
}
```

### Best Practices

1. **Increment SchemaVersion** when schema layout changes
2. **Preserve all versions** - each upgrader handles one version increment
3. **Test thoroughly** - unit tests for upgrader logic, acceptance tests for real upgrades
4. **Document breaking changes** - provide migration guides for users

## Plugin Framework Migration Notes

When migrating from terraform-plugin-sdk/v2 to terraform-plugin-framework:

| SDK v2 | Plugin Framework | Notes |
|--------|------------------|-------|
| `ForceNew: true` | `RequiresReplace()` plan modifier | Use in `PlanModifiers` field |
| `DiffSuppressFunc` | Custom type with `SemanticEquals()` | Or use plan modifiers |
| `StateFunc` | Custom type with `ValueFromTerraform()` | Normalization in type |
| `Sensitive: true` | `Sensitive: true` | Same behavior |
| `GetOk()` | `IsNull()` / `IsUnknown()` | Type-safe null handling |
| N/A | `UseStateForUnknown()` plan modifier | Preserve computed values |
| N/A | `WriteOnly: true` | True write-only support |

### Framework Custom Type Example

```go
// SDK v2 DiffSuppressFunc becomes custom type in Framework
type CaseInsensitiveStringType struct {
    basetypes.StringType
}

func (t CaseInsensitiveStringType) ValueType(_ context.Context) attr.Value {
    return CaseInsensitiveStringValue{}
}

type CaseInsensitiveStringValue struct {
    basetypes.StringValue
}

func (v CaseInsensitiveStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
    newValue, ok := newValuable.(CaseInsensitiveStringValue)
    if !ok {
        return false, nil
    }
    return strings.EqualFold(v.ValueString(), newValue.ValueString()), nil
}
```

### Framework Plan Modifier Example

```go
// SDK v2 ForceNew: true becomes RequiresReplace() plan modifier
schema.StringAttribute{
    Required:    true,
    Description: "Username (cannot be changed after creation)",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},

// SDK v2 Optional+Computed becomes UseStateForUnknown
schema.StringAttribute{
    Optional:    true,
    Computed:    true,
    Description: "Description with API default",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

### Framework WriteOnly Example

```go
// True WriteOnly in Framework (not possible in SDK v2)
schema.StringAttribute{
    Optional:    true,
    Sensitive:   true,
    WriteOnly:   true,  // Only in Framework
    Description: "Password (never stored in state)",
},
```

## Summary

1. **Use the decision tree** to choose the correct attribute configuration
2. **Apply helper functions** from `schema_helpers.go` for consistent patterns
3. **Use diff suppression** only when semantic equality differs from syntactic equality
4. **Prefer StateFunc normalization** over DiffSuppressFunc when possible
5. **Plan for migration** - document where Framework features will be needed
6. **Version your schema** and provide StateUpgraders for changes
