# Terraform Provider Testing Patterns

This document describes the testing patterns and best practices for the Terraform Provider for RTX routers. Following these patterns ensures consistent, reliable, and maintainable tests across all resources.

## Table of Contents

- [Overview](#overview)
- [Test Types](#test-types)
- [Decision Tree: Choosing the Right Test](#decision-tree-choosing-the-right-test)
- [Test Infrastructure](#test-infrastructure)
- [Test Patterns](#test-patterns)
  - [Pattern 1: Perpetual Diff Prevention](#pattern-1-perpetual-diff-prevention)
  - [Pattern 2: Import Testing](#pattern-2-import-testing)
  - [Pattern 3: State Migration Testing](#pattern-3-state-migration-testing)
  - [Pattern 4: Optional+Computed Field Preservation](#pattern-4-optionalcomputed-field-preservation)
  - [Pattern 5: WriteOnly/Sensitive Attribute Testing](#pattern-5-writeonlysensitive-attribute-testing)
- [Test Utilities](#test-utilities)
- [Running Tests](#running-tests)
- [Troubleshooting](#troubleshooting)

## Overview

The RTX provider uses the standard Terraform Plugin SDK v2 testing framework with additional utilities in the `internal/provider/acctest` package. Tests are organized into two main categories:

1. **Unit Tests** - Test isolated components without external dependencies
2. **Acceptance Tests** - Test full resource lifecycle against a real RTX router

## Test Types

### Unit Tests

Unit tests verify isolated functionality without external dependencies:

- Parser logic validation
- Configuration builder output
- State migration upgraders
- Helper function correctness

```go
func TestRandomName(t *testing.T) {
    name1 := acctest.RandomName("test")
    name2 := acctest.RandomName("test")

    assert.NotEqual(t, name1, name2, "RandomName should generate unique names")
    assert.True(t, strings.HasPrefix(name1, "test-"), "should have correct prefix")
}
```

### Acceptance Tests

Acceptance tests verify resource behavior against a real RTX router:

- CRUD lifecycle (Create, Read, Update, Delete)
- Import functionality
- State persistence
- Configuration changes

Acceptance tests require:
- `TF_ACC=1` environment variable
- Valid router credentials (`RTX_HOST`, `RTX_USERNAME`, `RTX_PASSWORD`)

## Decision Tree: Choosing the Right Test

Use this decision tree to determine which test type(s) to implement:

```
Is this testing a resource/data source?
├── No: Use Unit Tests
│   ├── Parser logic → Parser unit tests
│   ├── Helper functions → Function unit tests
│   └── State migration → StateMigrationTestCase
│
└── Yes: Use Acceptance Tests
    │
    ├── New resource? → Implement ALL of the following:
    │   ├── Basic CRUD test (create, read, update, delete)
    │   ├── Perpetual diff prevention test
    │   ├── Import test
    │   └── Any resource-specific edge cases
    │
    ├── Has Optional+Computed fields? → Add preservation test
    │
    ├── Has sensitive fields? → Add WriteOnly attribute test
    │
    ├── Has schema version upgrade? → Add state migration test
    │
    └── Bug fix? → Add regression test with specific scenario
```

### Minimum Test Coverage Requirements

Every resource MUST have:

| Test Type | Purpose | Required |
|-----------|---------|----------|
| Basic CRUD | Verify lifecycle works | Yes |
| No-diff | Catch perpetual diffs | Yes |
| Import | Verify import functionality | Yes |
| Preservation | Test Optional+Computed fields | If applicable |
| Sensitive | Test write-only fields | If applicable |
| Migration | Test schema upgrades | If schema version > 0 |

## Test Infrastructure

### The acctest Package

The `internal/provider/acctest` package provides reusable test utilities:

```go
import "github.com/sh1/terraform-provider-rtx/internal/provider/acctest"
```

#### Provider Factories

```go
resource.Test(t, resource.TestCase{
    PreCheck:          func() { acctest.PreCheck(t) },
    ProviderFactories: acctest.ProviderFactories,
    Steps: []resource.TestStep{
        // test steps
    },
})
```

#### PreCheck Functions

```go
// Basic connectivity check
acctest.PreCheck(t)

// Check with admin password requirement
acctest.PreCheckWithAdminPassword(t)

// Custom environment variable check
acctest.SkipIfEnvNotSet(t, "RTX_CUSTOM_VAR")
```

#### Random Name Generation

```go
// Generate unique names for parallel test safety
name := acctest.RandomName("testuser")  // e.g., "testuser-abc12def"

// Generate with specific length
name := acctest.RandomNameWithLength("user", 4)  // e.g., "user-ab12"

// Generate random values
port := acctest.RandomInt(1000, 9999)
ip := acctest.RandomIP("192.168.1", 10, 254)  // e.g., "192.168.1.47"
```

## Test Patterns

### Pattern 1: Perpetual Diff Prevention

**Purpose:** Verify that re-applying the same configuration produces no changes.

**When to use:** Every resource should have this test to catch perpetual diff bugs.

```go
func TestAccAdminUser_noDiff(t *testing.T) {
    resourceName := "rtx_admin_user.test"
    username := acctest.RandomName("testuser")

    config := fmt.Sprintf(`
resource "rtx_admin_user" "test" {
    username = %q
    password = "testpass123"
}`, username)

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { acctest.PreCheck(t) },
        ProviderFactories: acctest.ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create the resource
            {
                Config: config,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet(resourceName, "id"),
                    resource.TestCheckResourceAttr(resourceName, "username", username),
                ),
            },
            // Step 2: Re-apply same config - MUST produce no changes
            {
                Config:   config,
                PlanOnly: true,
                // Test fails if any changes are planned
            },
        },
    })
}
```

**Key Points:**
- The second step uses `PlanOnly: true`
- If changes are planned, the test fails automatically
- This catches issues where Read() doesn't properly populate state

### Pattern 2: Import Testing

**Purpose:** Verify that existing resources can be imported into Terraform state.

**When to use:** Every resource that supports import.

```go
func TestAccAdminUser_import(t *testing.T) {
    resourceName := "rtx_admin_user.test"
    username := acctest.RandomName("testuser")

    config := fmt.Sprintf(`
resource "rtx_admin_user" "test" {
    username = %q
    password = "testpass123"
}`, username)

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { acctest.PreCheck(t) },
        ProviderFactories: acctest.ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create resource
            {
                Config: config,
            },
            // Step 2: Import and verify
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
                // Ignore fields that cannot be imported (e.g., passwords)
                ImportStateVerifyIgnore: []string{"password"},
            },
        },
    })
}
```

**Key Points:**
- Use `ImportStateVerify: true` to compare imported state with original
- Use `ImportStateVerifyIgnore` for write-only fields like passwords
- The import ID format must match what your `Importer` expects

### Pattern 3: State Migration Testing

**Purpose:** Verify that state upgraders correctly transform old state formats.

**When to use:** When a resource has `SchemaVersion > 0` and state upgraders.

#### Unit Test Approach (Preferred)

```go
func TestAdminUser_StateUpgradeV0(t *testing.T) {
    cases := []acctest.StateMigrationTestCase{
        {
            Name: "basic upgrade",
            OldState: map[string]interface{}{
                "name": "admin",  // Old field name
            },
            ExpectedState: map[string]interface{}{
                "username": "admin",  // New field name
            },
            UpgradeFunc: resourceAdminUserStateUpgradeV0,
        },
        {
            Name: "nil state handling",
            OldState:      nil,
            ExpectedState: nil,
            UpgradeFunc:   resourceAdminUserStateUpgradeV0,
        },
        {
            Name: "missing optional fields",
            OldState: map[string]interface{}{
                "name": "admin",
            },
            ExpectedState: map[string]interface{}{
                "username":      "admin",
                "administrator": false,  // Default value
            },
            UpgradeFunc: resourceAdminUserStateUpgradeV0,
        },
    }

    acctest.RunStateMigrationTests(t, cases)
}
```

#### Cross-Version Acceptance Test (When Needed)

```go
func TestAccAdminUser_crossVersionUpgrade(t *testing.T) {
    config := `
resource "rtx_admin_user" "test" {
    username = "upgradetest"
    password = "testpass"
}`

    steps := acctest.BuildCrossVersionUpgradeSteps(acctest.CrossVersionUpgradeTestConfig{
        ProviderSource:              "registry.terraform.io/sh1/rtx",
        OldVersion:                  "1.0.0",
        OldConfig:                   config,
        NewConfig:                   config,
        ExpectEmptyPlanAfterUpgrade: true,
        NewChecks: []resource.TestCheckFunc{
            resource.TestCheckResourceAttr("rtx_admin_user.test", "username", "upgradetest"),
        },
    })

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { acctest.PreCheck(t) },
        ProviderFactories: acctest.ProviderFactories,
        Steps:             steps,
    })
}
```

### Pattern 4: Optional+Computed Field Preservation

**Purpose:** Verify that Optional+Computed fields retain their values when not specified in configuration updates.

**When to use:** Resources with `Optional: true, Computed: true` schema fields.

**Background:** When a field is both Optional and Computed, removing it from the config should preserve the existing value rather than resetting to default.

```go
func TestAccAdminUser_preserveAdministrator(t *testing.T) {
    resourceName := "rtx_admin_user.test"
    username := acctest.RandomName("testuser")

    // Config with administrator explicitly set
    configWithAdmin := fmt.Sprintf(`
resource "rtx_admin_user" "test" {
    username      = %q
    password      = "testpass123"
    administrator = true
}`, username)

    // Config without administrator (should preserve existing value)
    configWithoutAdmin := fmt.Sprintf(`
resource "rtx_admin_user" "test" {
    username = %q
    password = "testpass123"
}`, username)

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { acctest.PreCheck(t) },
        ProviderFactories: acctest.ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create with administrator=true
            {
                Config: configWithAdmin,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(resourceName, "administrator", "true"),
                ),
            },
            // Step 2: Apply without administrator - should preserve true
            {
                Config: configWithoutAdmin,
                Check: resource.ComposeTestCheckFunc(
                    // Value should be preserved, not reset to false
                    resource.TestCheckResourceAttr(resourceName, "administrator", "true"),
                ),
            },
        },
    })
}
```

**Key Points:**
- First step sets the Optional+Computed field explicitly
- Second step removes the field from config
- Assertion verifies the value is preserved, not reset to default

### Pattern 5: WriteOnly/Sensitive Attribute Testing

**Purpose:** Verify that sensitive fields are handled securely.

**When to use:** Resources with password fields or other sensitive data.

```go
func TestAccAdminUser_passwordHandling(t *testing.T) {
    resourceName := "rtx_admin_user.test"
    username := acctest.RandomName("testuser")

    config := fmt.Sprintf(`
resource "rtx_admin_user" "test" {
    username = %q
    password = "secretpass123"
}`, username)

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { acctest.PreCheck(t) },
        ProviderFactories: acctest.ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: config,
                Check: resource.ComposeTestCheckFunc(
                    // Verify resource was created
                    resource.TestCheckResourceAttrSet(resourceName, "id"),
                    resource.TestCheckResourceAttr(resourceName, "username", username),
                    // Password should either:
                    // - Not be in state (WriteOnly)
                    // - Be marked sensitive (not exposed in plan output)
                ),
            },
        },
    })
}
```

## Test Utilities

### ConfigBuilder

Build HCL configurations programmatically:

```go
config := acctest.NewConfigBuilder("rtx_admin_user", "test").
    SetAttribute("username", "testuser").
    SetAttribute("password", "testpass").
    SetAttribute("administrator", true).
    Build()

// Generates:
// resource "rtx_admin_user" "test" {
//     administrator = true
//     password = "testpass"
//     username = "testuser"
// }
```

#### Complex Configurations

```go
// With nested blocks
config := acctest.NewConfigBuilder("rtx_interface", "test").
    SetAttribute("name", "lan1").
    SetAttribute("ip_address", "192.168.1.1/24").
    AddBlock(
        acctest.NewBlockBuilder("dhcp", "").
            SetAttribute("enabled", true).
            SetAttribute("range_start", "192.168.1.100"),
    ).
    Build()

// Multiple resources
multi := acctest.NewMultiConfigBuilder().
    AddResource(
        acctest.NewConfigBuilder("rtx_admin_user", "admin").
            SetAttribute("username", "admin").
            SetAttribute("administrator", true),
    ).
    AddResource(
        acctest.NewConfigBuilder("rtx_admin_user", "user").
            SetAttribute("username", "user").
            SetAttribute("administrator", false),
    ).
    Build()
```

### State Migration Helpers

```go
// Run multiple migration test cases
acctest.RunStateMigrationTests(t, cases)

// Run a single test case
acctest.RunStateMigrationTest(t, tc)

// Test upgrade chain through multiple versions
acctest.StateUpgradeChain(t, initialState, expectedFinalState,
    upgradeV0toV1,
    upgradeV1toV2,
)

// Validate upgrader configuration
acctest.ValidateStateUpgrader(t, resourceSchema.StateUpgraders[0])
```

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run specific package tests
go test ./internal/provider/acctest/...

# Run with verbose output
go test -v ./internal/provider/...
```

### Acceptance Tests

```bash
# Set required environment variables
export RTX_HOST="192.168.1.1"
export RTX_USERNAME="admin"
export RTX_PASSWORD="your-password"

# Run all acceptance tests
TF_ACC=1 go test ./internal/provider/... -v -timeout 120m

# Run specific test
TF_ACC=1 go test ./internal/provider/... -v -run TestAccAdminUser_basic

# Run with parallel limit (recommended for router tests)
TF_ACC=1 go test ./internal/provider/... -v -parallel 1
```

### Test Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

## Troubleshooting

### Common Test Failures

#### 1. Perpetual Diff Detected

**Symptom:** Test fails on the `PlanOnly: true` step with unexpected changes.

**Common Causes:**
- Read() doesn't populate all schema fields
- Type mismatch between API response and schema (e.g., int vs string)
- Computed fields not being set properly
- Default values not matching router defaults

**Solution:**
```go
// Ensure Read() populates all fields, including computed ones
d.Set("field_name", apiResponse.FieldName)

// Check for type mismatches
d.Set("port", strconv.Itoa(apiResponse.Port))  // If schema expects string

// Set defaults explicitly in Create/Read if needed
if d.Get("optional_field") == nil {
    d.Set("optional_field", defaultValue)
}
```

#### 2. Import State Verification Failed

**Symptom:** Import test fails with attribute mismatch.

**Common Causes:**
- Write-only fields included in verification
- Different format in imported state vs created state
- Missing fields in Read() implementation

**Solution:**
```go
// Ignore write-only fields
ImportStateVerifyIgnore: []string{"password", "secret_key"},

// Normalize field formats in Read()
d.Set("ip_address", normalizeIPAddress(apiResponse.IP))
```

#### 3. Resource Not Found After Create

**Symptom:** Test fails because resource cannot be found immediately after creation.

**Common Causes:**
- Router needs time to apply configuration
- ID not set correctly after create
- Read() using wrong ID format

**Solution:**
```go
// Ensure ID is set after create
d.SetId(fmt.Sprintf("%s:%s", resourceType, resourceName))

// Add retry logic in Read() if needed
err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
    // read implementation
})
```

#### 4. State Migration Test Fails

**Symptom:** State upgrade test shows unexpected diff.

**Common Causes:**
- Old state format doesn't match test case
- Missing field transformations
- Type conversions not handled

**Solution:**
```go
// Log actual vs expected state for debugging
t.Logf("Old state: %+v", tc.OldState)
t.Logf("Expected: %+v", tc.ExpectedState)
t.Logf("Actual: %+v", actualState)

// Ensure all field mappings are correct
func upgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
    if rawState == nil {
        return nil, nil
    }

    // Explicit field mapping
    if v, ok := rawState["old_field"]; ok {
        rawState["new_field"] = v
        delete(rawState, "old_field")
    }

    return rawState, nil
}
```

#### 5. Test Timeout

**Symptom:** Test fails with context deadline exceeded.

**Common Causes:**
- Router is slow to respond
- Network connectivity issues
- Too many parallel tests

**Solution:**
```bash
# Increase timeout
TF_ACC=1 go test ./... -timeout 30m

# Reduce parallelism
TF_ACC=1 go test ./... -parallel 1
```

#### 6. Environment Variable Missing

**Symptom:** Test skipped with "environment variable not set" message.

**Cause:** Required environment variables not configured.

**Solution:**
```bash
# Check required variables
echo $RTX_HOST
echo $RTX_USERNAME
echo $RTX_PASSWORD

# Set all required variables
export RTX_HOST="your-router-ip"
export RTX_USERNAME="admin"
export RTX_PASSWORD="your-password"
```

### Debugging Tips

1. **Enable verbose logging:**
   ```bash
   TF_LOG=DEBUG TF_ACC=1 go test ./... -v -run TestName
   ```

2. **Add debug output in tests:**
   ```go
   t.Logf("Config: %s", config)
   t.Logf("Resource state: %+v", d.State())
   ```

3. **Check router state directly:**
   ```bash
   ssh admin@router "show config"
   ```

4. **Run single test in isolation:**
   ```bash
   TF_ACC=1 go test ./internal/provider/... -v -run ^TestAccAdminUser_basic$ -count=1
   ```

5. **Inspect test state files:**
   Terraform test framework creates temporary state files. Check for leftover resources if tests fail mid-execution.

## Best Practices Summary

1. **Always use `acctest.PreCheck(t)`** to validate prerequisites
2. **Generate unique names with `acctest.RandomName()`** for parallel test safety
3. **Implement all required test types** for each resource
4. **Use `PlanOnly: true`** to verify no perpetual diffs
5. **Ignore write-only fields** in import verification
6. **Test edge cases** like nil state in migrations
7. **Keep tests focused** - one scenario per test function
8. **Clean up resources** - use CheckDestroy if needed
9. **Document test requirements** in comments
10. **Run tests regularly** during development
