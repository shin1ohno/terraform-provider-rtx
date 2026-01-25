# Project Structure

## Directory Organization

```
terraform-provider-rtx/
├── internal/                    # Private implementation packages
│   ├── client/                  # RTX router SSH client and services
│   │   ├── client.go            # Main RTX client implementation
│   │   ├── ssh_dialer.go        # SSH connection management
│   │   ├── ssh_session_pool.go  # SSH session pooling for connection reuse
│   │   ├── rtx_session.go       # RTX terminal session handling
│   │   ├── executor.go          # Command execution interface
│   │   ├── *_service.go         # Feature-specific service implementations
│   │   ├── errors.go            # Custom error types
│   │   └── interfaces.go        # Shared interfaces
│   ├── provider/                # Terraform provider implementation
│   │   ├── provider.go          # Provider configuration and setup
│   │   ├── resource_rtx_*.go    # Resource implementations
│   │   └── data_source_rtx_*.go # Data source implementations
│   ├── logging/                 # Structured logging utilities
│   │   └── logger.go            # Zerolog-based logging with context
│   └── rtx/                     # RTX-specific utilities
│       ├── parsers/             # CLI output parsers
│       │   ├── registry.go      # Parser registry pattern
│       │   └── *.go             # Feature-specific parsers
│       └── testdata/            # Test fixtures for parsers
├── examples/                    # Terraform configuration examples
│   ├── provider/                # Provider configuration example
│   ├── dhcp/                    # DHCP binding example
│   └── dhcp_scope/              # DHCP scope example
├── docs/                        # Documentation
│   └── RTX-commands/            # RTX CLI command reference (Japanese)
├── test/                        # Integration test infrastructure
│   └── docker/                  # RTX simulator for testing
├── main.go                      # Provider entry point
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── go.sum                       # Go dependency checksums
```

## Naming Conventions

### Files
- **Resources**: `resource_rtx_<feature>.go` (e.g., `resource_rtx_dhcp_binding.go`)
- **Data Sources**: `data_source_rtx_<feature>.go` (e.g., `data_source_rtx_interfaces.go`)
- **Services**: `<feature>_service.go` (e.g., `dhcp_service.go`, `config_service.go`)
- **Parsers**: `<feature>.go` in parsers package (e.g., `dhcp_bindings.go`)
- **Tests**: `<filename>_test.go` (standard Go convention)

### Code
- **Types/Structs**: `PascalCase` (e.g., `DhcpBinding`, `RTXClient`)
- **Functions/Methods**: `PascalCase` for exported, `camelCase` for private
- **Constants**: `PascalCase` for exported, `camelCase` for private
- **Variables**: `camelCase`
- **Package Names**: `lowercase` single word (e.g., `client`, `provider`, `parsers`)

## Import Patterns

### Import Order
```go
import (
    // 1. Standard library
    "context"
    "fmt"

    // 2. External dependencies
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

    // 3. Internal packages
    "github.com/sh1/terraform-provider-rtx/internal/client"
)
```

### Package Dependencies
```
main.go
    └── internal/provider
            ├── internal/client
            │       └── internal/rtx/parsers
            └── internal/rtx/parsers
```

- `provider` depends on `client` and `parsers`
- `client` depends on `parsers`
- `parsers` has no internal dependencies (leaf package)

## Code Structure Patterns

### Resource Implementation Pattern
```go
// resource_rtx_<feature>.go

func resourceRtx<Feature>() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceRtx<Feature>Create,
        ReadContext:   resourceRtx<Feature>Read,
        UpdateContext: resourceRtx<Feature>Update,
        DeleteContext: resourceRtx<Feature>Delete,
        Importer: &schema.ResourceImporter{
            StateContext: resourceRtx<Feature>Import,
        },
        Schema: map[string]*schema.Schema{
            // Field definitions
        },
    }
}

func resourceRtx<Feature>Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*client.RTXClient)
    // 1. Extract values from ResourceData
    // 2. Call client service method
    // 3. Set resource ID
    // 4. Call Read to populate state
}
```

### Service Implementation Pattern
```go
// internal/client/<feature>_service.go

// Service methods on RTXClient
func (c *RTXClient) Get<Feature>(ctx context.Context, id string) (*<Feature>, error) {
    // 1. Build RTX CLI command
    // 2. Execute via SSH session
    // 3. Parse output using parser
    // 4. Return structured result
}

func (c *RTXClient) Create<Feature>(ctx context.Context, config *<Feature>Config) error {
    // 1. Build RTX CLI commands
    // 2. Execute commands sequentially
    // 3. Save configuration
    // 4. Handle errors
}
```

### Parser Implementation Pattern
```go
// internal/rtx/parsers/<feature>.go

func init() {
    // Register parser in global registry
    RegisterParser("<feature>", Parse<Feature>)
}

func Parse<Feature>(output string) (interface{}, error) {
    // 1. Split output into lines
    // 2. Apply regex patterns
    // 3. Build structured result
    // 4. Return with validation
}
```

## Module Boundaries

### Public API (Terraform Interface)
- `provider.go`: Provider configuration schema
- `resource_rtx_*.go`: Resource schemas and CRUD operations
- `data_source_rtx_*.go`: Data source schemas and read operations

### Internal Implementation
- `client/`: SSH connection, command execution, RTX communication
- `parsers/`: CLI output parsing (isolated, testable)

### Dependency Rules
1. Provider package may import client and parsers
2. Client package may import parsers
3. Parsers package must not import other internal packages
4. No circular dependencies allowed

## Code Size Guidelines

- **File Size**: Prefer < 500 lines; split if exceeding 800 lines
- **Function Size**: Prefer < 50 lines; extract helpers if larger
- **Cyclomatic Complexity**: Keep functions under 10; refactor complex logic
- **Nesting Depth**: Maximum 4 levels; flatten with early returns

## Adding New Features

### New Resource Checklist
1. Create `internal/provider/resource_rtx_<feature>.go`
2. Create `internal/client/<feature>_service.go`
3. Create `internal/rtx/parsers/<feature>.go`
4. Register resource in `provider.go` ResourcesMap
5. Add example in `examples/<feature>/`
6. Add tests for parser, service, and resource

### New Data Source Checklist
1. Create `internal/provider/data_source_rtx_<feature>.go`
2. Add service method in `internal/client/`
3. Create parser if needed in `internal/rtx/parsers/`
4. Register data source in `provider.go` DataSourcesMap
5. Add tests

## Documentation Standards

- All exported functions must have GoDoc comments
- Resource schemas should include `Description` for each field
- Complex parsing logic should have inline comments
- Examples directory should have working Terraform configurations
- README in examples subdirectories explaining usage
