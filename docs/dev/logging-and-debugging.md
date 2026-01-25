# Logging and Debugging

This document describes the logging infrastructure and debugging techniques for the terraform-provider-rtx.

## Environment Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `TF_LOG` | `debug`, `info`, `warn`, `error` | Log level (default: `warn`) |
| `TF_LOG_JSON` | `1` | Enable JSON output format |

### Log Levels

```bash
# Debug - verbose output including SSH commands and responses
TF_LOG=debug terraform plan

# Info - informational messages
TF_LOG=info terraform apply

# Warn - warnings only (default)
TF_LOG=warn terraform apply

# Error - errors only
TF_LOG=error terraform apply
```

### Output Format

**Console format** (default for local development):
```
15:04:05 DBG Executing command cmd="show config"
15:04:05 INF Resource created resource=rtx_interface id=lan1
```

**JSON format** (auto-enabled in CI, or with `TF_LOG_JSON=1`):
```json
{"level":"debug","time":"15:04:05","cmd":"show config","message":"Executing command"}
{"level":"info","time":"15:04:05","resource":"rtx_interface","id":"lan1","message":"Resource created"}
```

JSON format is automatically enabled when:
- `TF_LOG_JSON=1` is set
- `CI` environment variable is set
- `GITHUB_ACTIONS` environment variable is set
- `JENKINS_URL` environment variable is set

## Logging Infrastructure

### Package Structure

```
internal/logging/
├── logger.go      # Logger initialization and configuration
├── context.go     # Context-based logger propagation
└── sanitizer.go   # Sensitive data redaction
```

### Logger Initialization

The logger is initialized based on `TF_LOG` and `TF_LOG_JSON`:

```go
import "github.com/sh1/terraform-provider-rtx/internal/logging"

// Get logger from context (recommended)
logger := logging.FromContext(ctx)
logger.Info().Str("resource", name).Msg("Resource created")

// Get global logger
logger := logging.Global()
```

### Context-Based Logging

Loggers are propagated through context for request tracing:

```go
// Attach logger to context
ctx = logging.WithContext(ctx, logger)

// Attach resource info to context
ctx = logging.WithResource(ctx, "rtx_interface", "lan1")

// Retrieve in downstream functions
logger := logging.FromContext(ctx)
resourceInfo := logging.ResourceFromContext(ctx)
```

### Structured Logging

Use zerolog's fluent API for structured logs:

```go
// Good - structured fields
log.Info().
    Str("resource", "rtx_interface").
    Str("id", "lan1").
    Msg("Resource created")

// Good - error with context
log.Error().
    Err(err).
    Str("cmd", command).
    Msg("Command execution failed")

// Avoid - unstructured messages
log.Info().Msgf("Resource %s created with id %s", resource, id)
```

## Sensitive Data Protection

### Automatic Redaction

The sanitizer automatically redacts sensitive data from logs:

**Protected patterns:**
- `password`
- `pre-shared-key`
- `secret`
- `community` (SNMP)
- `token`
- `key`
- `credential`

**Protected field names:**
- `password`
- `pre_shared_key`
- `secret`
- `community`
- `token`
- `api_key`
- `credential`

### Manual Sanitization

Use sanitization functions when logging potentially sensitive data:

```go
import "github.com/sh1/terraform-provider-rtx/internal/logging"

// Sanitize a string
safe := logging.SanitizeString(maybeSecretCommand)
log.Debug().Str("cmd", safe).Msg("Executing")

// Sanitize a map
safeMap := logging.SanitizeMap(config)
log.Debug().Interface("config", safeMap).Msg("Configuration")

// Check if field is sensitive
if logging.IsSensitiveField("password") {
    // Don't log the value
}
```

### Sanitized Logger

Create a logger with the sanitizing hook:

```go
logger := logging.NewSanitizedLogger()
```

## Debugging Techniques

### Enable Debug Logging

```bash
# Full debug output
TF_LOG=debug terraform plan 2>&1 | tee debug.log

# Filter provider logs only
TF_LOG=debug terraform plan 2>&1 | grep "rtx"
```

### SSH Command Tracing

Debug logs include SSH commands sent to the router:

```
15:04:05 DBG Executing command cmd="show config"
15:04:05 DBG Command output output="ip lan1 address 192.168.1.1/24..."
```

### State Debugging

```bash
# Show current state
terraform state show rtx_interface.lan1

# List all resources
terraform state list

# Refresh state from router
terraform refresh
```

### Plan Debugging

```bash
# Detailed plan output
terraform plan -out=plan.tfplan
terraform show -json plan.tfplan | jq .
```

### Import Debugging

```bash
# Debug import process
TF_LOG=debug terraform import rtx_interface.lan1 lan1

# Verify imported state
terraform state show rtx_interface.lan1
```

## Common Issues

### "Permission denied" errors

Check SSH credentials and ensure the user has administrator privileges:

```bash
TF_LOG=debug terraform plan 2>&1 | grep -i "permission\|denied\|auth"
```

### State drift

Enable debug logging to see what the provider reads from the router vs. what's in state:

```bash
TF_LOG=debug terraform plan 2>&1 | grep -E "(Read|current|desired)"
```

### Connection timeouts

Increase timeout and check connectivity:

```hcl
provider "rtx" {
  timeout = 60  # Increase from default 30
}
```

```bash
# Test SSH connectivity
ssh -v admin@192.168.1.1
```

### Parallel operation issues

Reduce parallelism if experiencing race conditions:

```hcl
provider "rtx" {
  max_parallelism = 1  # Serialize all operations
}
```

Or use environment variable:

```bash
RTX_MAX_PARALLELISM=1 terraform apply
```

## Best Practices

1. **Always use structured logging** - Use `.Str()`, `.Int()`, `.Err()` instead of `Msgf()`

2. **Include context** - Add resource type, ID, and operation to log entries

3. **Never log secrets** - Use `SanitizeString()` for any user-provided data

4. **Use appropriate levels**:
   - `Debug`: SSH commands, detailed flow
   - `Info`: Resource CRUD operations
   - `Warn`: Recoverable issues
   - `Error`: Failures requiring attention

5. **Propagate logger through context** - Don't create new loggers in every function
