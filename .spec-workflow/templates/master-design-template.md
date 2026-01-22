# Master Design: {resource-name}

## Overview

{High-level description of the resource implementation and its place in the overall system}

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `{terraform_resource_name}` |
| Service File | `internal/client/{service}_service.go` |
| Parser File | `internal/rtx/parsers/{parser}.go` |
| Resource File | `internal/provider/resource_{resource}.go` |
| Last Updated | {date} |
| Source Specs | {list of contributing specs} |

## Steering Document Alignment

### Technical Standards (tech.md)
{How the design follows documented technical patterns and standards}

### Project Structure (structure.md)
{How the implementation follows project organization conventions}

## Code Reuse Analysis

### Existing Components to Leverage
- **{Component/Utility Name}**: {How it will be used}
- **{Service/Helper Name}**: {How it will be extended}

### Integration Points
- **{Existing System/API}**: {How the feature integrates}
- **{Database/Storage}**: {How data connects to existing schemas}

## Architecture

{Describe the overall architecture and design patterns used}

### Modular Design Principles
- **Single File Responsibility**: Each file handles one specific concern or domain
- **Component Isolation**: Create small, focused components rather than large monolithic files
- **Service Layer Separation**: Separate data access, business logic, and presentation layers
- **Utility Modularity**: Break utilities into focused, single-purpose modules

```mermaid
graph TD
    subgraph Provider Layer
        TFResource[resource_{resource}.go]
    end

    subgraph Client Layer
        Client[client.go - Interface Extension]
        Service[{resource}_service.go]
    end

    subgraph Parser Layer
        Parser[{resource}.go]
        Commands[Command Builders]
        OutputParser[Output Parsers]
    end

    TFResource --> Client
    Client --> Service
    Service --> Parser
    Parser --> Commands
    Parser --> OutputParser
```

## Components and Interfaces

### Component 1: {Service} (`internal/client/{service}_service.go`)

- **Purpose:** {What this component does}
- **Interfaces:**
  ```go
  type {Service}Service struct {
      executor Executor
      client   *rtxClient
  }

  func (s *{Service}Service) Create(ctx context.Context, config {Config}) error
  func (s *{Service}Service) Get(ctx context.Context) (*{Config}, error)
  func (s *{Service}Service) Update(ctx context.Context, config {Config}) error
  func (s *{Service}Service) Delete(ctx context.Context) error
  ```
- **Dependencies:** {What it depends on}
- **Reuses:** {Existing components/utilities it builds upon}

### Component 2: {Parser} (`internal/rtx/parsers/{parser}.go`)

- **Purpose:** {Parses RTX router output and builds commands}
- **Interfaces:**
  ```go
  func Parse{Config}(raw string) (*{Config}, error)
  func Build{Command}(config {Config}) string
  func BuildDelete{Command}() string
  ```
- **Dependencies:** {What it depends on}
- **Reuses:** {Existing components/utilities it builds upon}

### Component 3: Terraform Resource (`internal/provider/resource_{resource}.go`)

- **Purpose:** Terraform resource definition implementing CRUD lifecycle
- **Interfaces:**
  ```go
  func resource{Resource}() *schema.Resource
  func resource{Resource}Create(ctx, d, meta) diag.Diagnostics
  func resource{Resource}Read(ctx, d, meta) diag.Diagnostics
  func resource{Resource}Update(ctx, d, meta) diag.Diagnostics
  func resource{Resource}Delete(ctx, d, meta) diag.Diagnostics
  func resource{Resource}Import(ctx, d, meta) ([]*schema.ResourceData, error)
  ```
- **Dependencies:** {What it depends on}
- **Reuses:** {Existing components/utilities it builds upon}

### Component 4: Client Interface Extension (`internal/client/interfaces.go`)

- **Purpose:** Extend Client interface with resource methods
- **Interfaces:**
  ```go
  // Add to existing Client interface:
  Get{Resource}(ctx context.Context) (*{Config}, error)
  Create{Resource}(ctx context.Context, config {Config}) error
  Update{Resource}(ctx context.Context, config {Config}) error
  Delete{Resource}(ctx context.Context) error
  ```
- **Dependencies:** Existing Client interface
- **Reuses:** Pattern from existing methods

## Data Models

### {Config} (Primary Configuration Model)

```go
// {Config} represents {resource} configuration on an RTX router
type {Config} struct {
    // Field documentation
    Field1 Type1 `json:"field1"`
    Field2 Type2 `json:"field2,omitempty"`
}
```

### Terraform Schema

```hcl
resource "{terraform_resource}" "{name}" {
  {Complete schema example}
}
```

## RTX Command Mapping

### {Operation 1}

```
{RTX command syntax}
```

Example: `{concrete example}`

### {Operation 2}

```
{RTX command syntax}
```

Example: `{concrete example}`

## Error Handling

### Error Scenarios

1. **{Error Scenario 1}**
   - **Handling:** {How to handle}
   - **User Impact:** {What user sees}

2. **{Error Scenario 2}**
   - **Handling:** {How to handle}
   - **User Impact:** {What user sees}

## Testing Strategy

### Unit Testing

- **Parser Tests** (`{parser}_test.go`):
  - {Test coverage description}

- **Service Tests** (`{service}_service_test.go`):
  - {Test coverage description}

### Integration Testing

- **Resource Tests** (`resource_{resource}_test.go`):
  - {Test coverage description}

### End-to-End Testing

- **Acceptance Tests** (with real RTX router):
  - {Test scenarios}

## File Structure

```
internal/
├── provider/
│   ├── resource_{resource}.go
│   └── resource_{resource}_test.go
├── client/
│   ├── interfaces.go              # MODIFY: Add types and methods
│   ├── client.go                  # MODIFY: Add service initialization
│   ├── {service}_service.go
│   └── {service}_service_test.go
└── rtx/
    └── parsers/
        ├── {parser}.go
        └── {parser}_test.go
```

## Implementation Notes

{Numbered list of important implementation considerations}

1. {Note 1}
2. {Note 2}

## State Handling

- Persist only configuration attributes in Terraform state
- Operational/runtime status must not be stored in state
- {Additional notes}

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| {date} | {spec-name} | {summary of changes} |
