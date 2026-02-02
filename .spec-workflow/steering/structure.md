# Project Structure

## Directory Organization

```
terraform-provider-rtx/
├── internal/                    # Private implementation packages
│   ├── client/                  # RTX router SSH client and services
│   │   ├── client.go            # Main RTX client implementation
│   │   ├── ssh_dialer.go        # SSH connection management with key auth
│   │   ├── ssh_session_pool.go  # SSH session pooling for connection reuse
│   │   ├── rtx_session.go       # RTX terminal session handling
│   │   ├── executor.go          # Command execution interface
│   │   ├── *_service.go         # Feature-specific service implementations
│   │   ├── errors.go            # Custom error types
│   │   └── interfaces.go        # Shared interfaces
│   ├── provider/                # Terraform provider implementation (Plugin Framework)
│   │   ├── provider_framework.go # Main provider definition
│   │   ├── resources/           # Resource implementations (modular structure)
│   │   │   └── {name}/          # One directory per resource
│   │   │       ├── resource.go  # Resource CRUD, schema, validators
│   │   │       └── model.go     # Data model with ToClient/FromClient
│   │   ├── datasources/         # Data source implementations
│   │   ├── fwhelpers/           # Framework helper functions
│   │   ├── validation/          # Custom schema validators
│   │   ├── validators/          # Attribute validators
│   │   ├── planmodifiers/       # Custom plan modifiers
│   │   └── acctest/             # Acceptance test helpers
│   ├── logging/                 # Structured logging utilities
│   │   └── logger.go            # Zerolog-based logging with context
│   └── rtx/                     # RTX-specific utilities
│       ├── parsers/             # CLI output parsers
│       │   ├── service.go       # Parser implementations
│       │   └── *_test.go        # Parser tests
│       └── testdata/            # Test fixtures for parsers
├── examples/                    # Terraform configuration examples
│   └── {resource}/              # One directory per resource
├── docs/                        # Generated documentation (tfplugindocs)
│   ├── resources/               # Resource documentation
│   └── data-sources/            # Data source documentation
├── main.go                      # Provider entry point
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── go.sum                       # Go dependency checksums
```

## Naming Conventions

### Files
- **Resources**: `internal/provider/resources/{name}/resource.go` + `model.go` (e.g., `resources/vlan/`)
- **Data Sources**: `internal/provider/datasources/{name}.go`
- **Services**: `internal/client/{feature}_service.go` (e.g., `dhcp_service.go`, `config_service.go`)
- **Parsers**: Functions in `internal/rtx/parsers/service.go` (e.g., `ParseVLAN()`)
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

### Resource Implementation Pattern (Plugin Framework)
```go
// internal/provider/resources/{name}/resource.go

type <Feature>Resource struct {
    client *client.Client
}

var (
    _ resource.Resource                = &<Feature>Resource{}
    _ resource.ResourceWithImportState = &<Feature>Resource{}
)

func New<Feature>Resource() resource.Resource {
    return &<Feature>Resource{}
}

func (r *<Feature>Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_<feature>"
}

func (r *<Feature>Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "...",
        Attributes: map[string]schema.Attribute{
            // Attribute definitions with validators
        },
    }
}

func (r *<Feature>Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    // Extract client from fwhelpers.ProviderData
}

func (r *<Feature>Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Get plan data into model
    // 2. Convert via model.ToClient()
    // 3. Call client service method
    // 4. Update model via model.FromClient()
    // 5. Set state
}
```

### Model Pattern (Plugin Framework)
```go
// internal/provider/resources/{name}/model.go

type <Feature>Model struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    // ... other fields
}

func (m *<Feature>Model) ToClient() *client.<Feature>Config {
    // Convert Terraform model to client struct
}

func (m *<Feature>Model) FromClient(data *client.<Feature>) {
    // Update Terraform model from client struct
}

func (m *<Feature>Model) ID() string {
    return m.ID.ValueString()
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
// internal/rtx/parsers/service.go

func Parse<Feature>(output string) ([]<Feature>, error) {
    // 1. Split output into lines
    // 2. Apply regex patterns
    // 3. Build structured result
    // 4. Return with validation
}
```

## Module Boundaries

### Public API (Terraform Interface)
- `provider_framework.go`: Provider configuration schema and resource/datasource registration
- `resources/{name}/resource.go`: Resource schemas and CRUD operations
- `datasources/{name}.go`: Data source schemas and read operations

### Internal Implementation
- `client/`: SSH connection, command execution, RTX communication
- `rtx/parsers/`: CLI output parsing (isolated, testable)
- `fwhelpers/`: Plugin Framework helper utilities
- `validation/`, `validators/`: Custom schema validators
- `planmodifiers/`: Custom plan modifiers

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
1. Create `internal/provider/resources/{name}/` directory
2. Create `model.go` with data model and ToClient/FromClient conversion methods
3. Create `resource.go` implementing the resource.Resource interface
4. Add service method in `internal/client/{feature}_service.go` if needed
5. Add parser function in `internal/rtx/parsers/service.go` if needed
6. Register resource in `provider_framework.go` Resources() method
7. Add example in `examples/{resource}/`
8. Run `go generate ./...` to generate documentation
9. Add tests for parser, service, and resource

### New Data Source Checklist
1. Create `internal/provider/datasources/{name}.go`
2. Add service method in `internal/client/`
3. Create parser function in `internal/rtx/parsers/service.go` if needed
4. Register data source in `provider_framework.go` DataSources() method
5. Add tests

## Documentation Standards

- All exported functions must have GoDoc comments
- Resource schemas should include `Description` for each field
- Complex parsing logic should have inline comments
- Examples directory should have working Terraform configurations
- README in examples subdirectories explaining usage
