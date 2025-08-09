# Code Style and Conventions for Terraform Provider RTX

## Go Code Conventions

### Package Structure
- All code is organized under module `github.com/sh1/terraform-provider-rtx`
- Internal packages under `internal/` directory:
  - `internal/provider/`: Terraform provider implementation
  - `internal/client/`: RTX client implementation
  - `internal/rtx/`: RTX-specific logic and parsers

### Naming Conventions
- **Files**: Snake case (e.g., `data_source_rtx_system_info.go`)
- **Types**: PascalCase for exported types (e.g., `Client`, `SystemInfo`)
- **Interfaces**: Single method interfaces end with "-er" (e.g., `Parser`, `ConnDialer`)
- **Functions**: 
  - Exported: PascalCase (e.g., `NewClient`)
  - Private: camelCase (e.g., `validateConfig`)
- **Variables**: camelCase for both exported and private
- **Constants**: Not explicitly seen in codebase yet

### Code Organization
1. **File Structure**:
   - No copyright headers
   - Package declaration first
   - Imports grouped (standard library, then external, then internal)
   - Type definitions, then methods, then functions

2. **Interface-First Design**:
   - Define interfaces for major components (Client, Parser, Session, etc.)
   - Implement concrete types that satisfy interfaces
   - Use dependency injection via options pattern

### Error Handling
- Custom error types defined in `errors.go`
- Errors wrapped with context using `fmt.Errorf`
- Validation errors return early
- Resource errors include operation context

### Testing Conventions
1. **Test Files**: `*_test.go` alongside implementation files
2. **Test Naming**: `Test{FunctionName}_{Scenario}` (e.g., `TestRTXSystemInfoDataSourceRead_Success`)
3. **Test Structure**:
   - Table-driven tests for multiple scenarios
   - Mocking with testify/mock
   - Separate unit and acceptance tests
4. **Acceptance Tests**: 
   - Prefix with `TestAcc`
   - Skip if `TF_ACC` not set
   - Use real or Docker-simulated RTX routers

### Terraform Provider Specifics
1. **Data Source Naming**: `dataSourceRTX{ResourceName}` (e.g., `dataSourceRTXSystemInfo`)
2. **Schema Definition**:
   - Required fields marked appropriately
   - Computed fields for read-only data
   - Detailed descriptions for each field
3. **Context Usage**: All provider methods use context for cancellation
4. **Resource IDs**: Use meaningful identifiers (hostname for system info)

## Documentation Style
- **Comments**: 
  - Interface methods have documentation comments
  - Public functions have brief descriptions
  - No inline implementation comments unless complex logic
- **Test Comments**: Describe test scenario and expectations

## Design Patterns Used
1. **Options Pattern**: For configurable client creation
   ```go
   NewClient(config *Config, opts ...Option)
   ```

2. **Strategy Pattern**: For retry logic and parsers
   ```go
   type RetryStrategy interface {
       Next(retry int) (delay time.Duration, giveUp bool)
   }
   ```

3. **Registry Pattern**: For model-specific parsers
   ```go
   func RegisterParser(model string, parser Parser)
   ```

4. **Builder Pattern**: Implied in test helpers and mock setup

## Security Considerations
- SSH host key verification options (never default to insecure)
- Password fields marked as sensitive in Terraform schemas
- No hardcoded credentials in code or tests
- Use environment variables for test credentials

## Performance Considerations
- Context-aware operations for proper cancellation
- Connection reuse where possible
- Efficient parsing with early returns
- Proper resource cleanup in defer statements

## Formatting and Linting
- Use `gofmt -s` for standard Go formatting
- Follow standard Go idioms
- Keep line length reasonable (no hard limit set)
- Group related functionality together

## Import Organization
```go
import (
    // Standard library
    "context"
    "fmt"
    
    // External dependencies
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    
    // Internal packages
    "github.com/sh1/terraform-provider-rtx/internal/client"
)
```