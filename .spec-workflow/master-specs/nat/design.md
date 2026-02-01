# Master Design: NAT Resources

## Overview

This document describes the technical design and architecture of NAT resources in the Terraform RTX Provider. The implementation follows a layered architecture pattern with clear separation between Terraform resource definitions, client services, and RTX command parsers.

## Resource Summary

### rtx_nat_static

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_nat_static` |
| Service File | `internal/client/nat_static_service.go` |
| Parser File | `internal/rtx/parsers/nat_static.go` |
| Resource File | `internal/provider/resource_rtx_nat_static.go` |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation-derived |

### rtx_nat_masquerade

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_nat_masquerade` |
| Service File | `internal/client/nat_masquerade_service.go` |
| Parser File | `internal/rtx/parsers/nat_masquerade.go` |
| Resource File | `internal/provider/resource_rtx_nat_masquerade.go` |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation-derived |

## Steering Document Alignment

### Technical Standards (tech.md)

- Uses **Terraform Plugin Framework** for resource implementation
- Follows Go standard project layout
- Implements context-based cancellation pattern
- Uses batch command execution for atomic operations
- Validates all inputs before command generation

### Project Structure (structure.md)

- Provider layer: `internal/provider/resource_rtx_nat_*.go`
- Client layer: `internal/client/nat_*_service.go`
- Parser layer: `internal/rtx/parsers/nat_*.go`
- Interfaces: `internal/client/interfaces.go`

## Code Reuse Analysis

### Existing Components to Leverage

- **Executor Interface**: Reuses `Executor` interface for command execution
- **RunBatch Pattern**: Follows established batch command pattern
- **IP Validation**: Shared `validateIPAddress`, `validateIPRange` functions
- **Logging**: Uses `logging.FromContext` for consistent debug output
- **SaveConfig**: Integrates with client's configuration persistence

### Integration Points

- **Client Interface**: Extends `Client` interface with NAT methods
- **rtxClient**: Services receive reference to main client for SaveConfig
- **Provider Registration**: Resources registered in provider schema

---

## Architecture

The NAT resources follow a three-tier architecture with clear responsibilities:

```
+------------------------------------------------------------------+
|                     Terraform Provider Layer                       |
|  +------------------------------+  +-----------------------------+  |
|  | resource_rtx_nat_static.go   |  | resource_rtx_nat_masquerade |  |
|  |                              |  |            .go              |  |
|  | - Schema definition          |  | - Schema definition         |  |
|  | - CRUD handlers              |  | - CRUD handlers             |  |
|  | - Import support             |  | - Import support            |  |
|  | - CustomizeDiff validation   |  | - Validation functions      |  |
|  +------------------------------+  +-----------------------------+  |
+------------------------------------------------------------------+
                                |
                                v
+------------------------------------------------------------------+
|                        Client Layer                               |
|  +------------------------------+  +-----------------------------+  |
|  | nat_static_service.go        |  | nat_masquerade_service.go   |  |
|  |                              |  |                             |  |
|  | - Create/Get/Update/Delete   |  | - Create/Get/Update/Delete  |  |
|  | - List                       |  | - List                      |  |
|  | - Validation                 |  | - Validation                |  |
|  | - Type conversion            |  | - Type conversion           |  |
|  +------------------------------+  +-----------------------------+  |
+------------------------------------------------------------------+
                                |
                                v
+------------------------------------------------------------------+
|                        Parser Layer                               |
|  +------------------------------+  +-----------------------------+  |
|  | parsers/nat_static.go        |  | parsers/nat_masquerade.go   |  |
|  |                              |  |                             |  |
|  | - Parse command output       |  | - Parse command output      |  |
|  | - Build RTX commands         |  | - Build RTX commands        |  |
|  | - Entry validation           |  | - Entry validation          |  |
|  | - Command patterns           |  | - Command patterns          |  |
|  +------------------------------+  +-----------------------------+  |
+------------------------------------------------------------------+
                                |
                                v
+------------------------------------------------------------------+
|                      RTX Router (SSH)                             |
+------------------------------------------------------------------+
```

### Modular Design Principles

- **Single File Responsibility**: Each file handles one specific concern
- **Component Isolation**: Services and parsers are independently testable
- **Service Layer Separation**: Business logic in services, I/O in executor
- **Utility Modularity**: Validation functions are reusable

---

## Components and Interfaces

### Component 1: NATStaticService (`internal/client/nat_static_service.go`)

- **Purpose:** Orchestrates NAT static CRUD operations
- **Interfaces:**
  ```go
  type NATStaticService struct {
      executor Executor
      client   *rtxClient // For SaveConfig
  }

  func NewNATStaticService(executor Executor, client *rtxClient) *NATStaticService

  func (s *NATStaticService) Create(ctx context.Context, nat NATStatic) error
  func (s *NATStaticService) Get(ctx context.Context, descriptorID int) (*NATStatic, error)
  func (s *NATStaticService) Update(ctx context.Context, nat NATStatic) error
  func (s *NATStaticService) Delete(ctx context.Context, descriptorID int) error
  func (s *NATStaticService) List(ctx context.Context) ([]NATStatic, error)
  ```
- **Dependencies:** Executor interface, parsers package
- **Reuses:** Executor.RunBatch, Executor.Run, SaveConfig pattern

### Component 2: NATMasqueradeService (`internal/client/nat_masquerade_service.go`)

- **Purpose:** Orchestrates NAT masquerade CRUD operations
- **Interfaces:**
  ```go
  type NATMasqueradeService struct {
      executor Executor
      client   *rtxClient
  }

  func NewNATMasqueradeService(executor Executor, client *rtxClient) *NATMasqueradeService

  func (s *NATMasqueradeService) Create(ctx context.Context, nat NATMasquerade) error
  func (s *NATMasqueradeService) Get(ctx context.Context, descriptorID int) (*NATMasquerade, error)
  func (s *NATMasqueradeService) Update(ctx context.Context, nat NATMasquerade) error
  func (s *NATMasqueradeService) Delete(ctx context.Context, descriptorID int) error
  func (s *NATMasqueradeService) List(ctx context.Context) ([]NATMasquerade, error)
  ```
- **Dependencies:** Executor interface, parsers package
- **Reuses:** Executor.RunBatch, Executor.Run, SaveConfig pattern

### Component 3: NAT Static Parser (`internal/rtx/parsers/nat_static.go`)

- **Purpose:** Parses RTX output and builds commands for static NAT
- **Interfaces:**
  ```go
  type NATStatic struct {
      DescriptorID int              `json:"descriptor_id"`
      Entries      []NATStaticEntry `json:"entries,omitempty"`
  }

  type NATStaticEntry struct {
      InsideLocal       string `json:"inside_local"`
      InsideLocalPort   int    `json:"inside_local_port,omitempty"`
      OutsideGlobal     string `json:"outside_global"`
      OutsideGlobalPort int    `json:"outside_global_port,omitempty"`
      Protocol          string `json:"protocol,omitempty"`
  }

  func ParseNATStaticConfig(raw string) ([]NATStatic, error)
  func (p *NATStaticParser) ParseSingleNATStatic(raw string, descriptorID int) (*NATStatic, error)

  // Command builders
  func BuildNATDescriptorTypeStaticCommand(id int) string
  func BuildNATStaticMappingCommand(id int, entry NATStaticEntry) string
  func BuildNATStaticPortMappingCommand(id int, entry NATStaticEntry) string
  func BuildDeleteNATStaticCommand(id int) string
  func BuildDeleteNATStaticMappingCommand(id int, entry NATStaticEntry) string
  func BuildDeleteNATStaticPortMappingCommand(id int, entry NATStaticEntry) string
  func BuildShowNATStaticCommand(descriptorID int) string
  func BuildShowAllNATStaticCommand() string

  // Validation
  func ValidateNATStatic(nat NATStatic) error
  func ValidateNATStaticEntry(entry NATStaticEntry) error
  func IsPortBasedNAT(entry NATStaticEntry) bool
  ```
- **Dependencies:** Standard library (regexp, strconv, strings)
- **Reuses:** ValidateDescriptorID, ValidateNATPort (shared functions)

### Component 4: NAT Masquerade Parser (`internal/rtx/parsers/nat_masquerade.go`)

- **Purpose:** Parses RTX output and builds commands for NAT masquerade
- **Interfaces:**
  ```go
  type NATMasquerade struct {
      DescriptorID  int                     `json:"descriptor_id"`
      OuterAddress  string                  `json:"outer_address"`
      InnerNetwork  string                  `json:"inner_network"`
      StaticEntries []MasqueradeStaticEntry `json:"static_entries,omitempty"`
  }

  type MasqueradeStaticEntry struct {
      EntryNumber       int    `json:"entry_number"`
      InsideLocal       string `json:"inside_local"`
      InsideLocalPort   *int   `json:"inside_local_port,omitempty"`
      OutsideGlobal     string `json:"outside_global,omitempty"`
      OutsideGlobalPort *int   `json:"outside_global_port,omitempty"`
      Protocol          string `json:"protocol,omitempty"`
  }

  func ParseNATMasqueradeConfig(raw string) ([]NATMasquerade, error)

  // Command builders
  func BuildNATDescriptorTypeMasqueradeCommand(id int) string
  func BuildNATDescriptorAddressOuterCommand(id int, address string) string
  func BuildNATDescriptorAddressInnerCommand(id int, network string) string
  func BuildNATMasqueradeStaticCommand(id int, entryNum int, entry MasqueradeStaticEntry) string
  func BuildDeleteNATMasqueradeCommand(id int) string
  func BuildDeleteNATMasqueradeStaticCommand(id int, entryNum int) string
  func BuildShowNATDescriptorCommand(id int) string
  func BuildShowAllNATDescriptorsCommand() string

  // Validation
  func ValidateNATMasquerade(nat NATMasquerade) error
  func ValidateDescriptorID(id int) error
  func ValidateNATPort(port int) error
  func ValidateNATProtocol(protocol string) error
  func ValidateOuterAddress(address string) error
  func IsProtocolOnly(protocol string) bool
  ```
- **Dependencies:** Standard library (regexp, strconv, strings, net)
- **Reuses:** Shared validation functions

### Component 5: Terraform Resource - NAT Static (`internal/provider/resource_rtx_nat_static.go`)

- **Purpose:** Terraform resource implementation for static NAT
- **Interfaces:**
  ```go
  func resourceRTXNATStatic() *schema.Resource
  func resourceRTXNATStaticCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATStaticRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATStaticUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATStaticDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATStaticImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error)

  // Helper functions
  func buildNATStaticFromResourceData(d *schema.ResourceData) client.NATStatic
  func expandNATStaticEntries(entries []interface{}) []client.NATStaticEntry
  func flattenNATStaticEntries(entries []client.NATStaticEntry) []interface{}
  func validateNATIPAddress(v interface{}, k string) ([]string, []error)
  func validateNATStaticEntries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error
  ```
- **Dependencies:** terraform-plugin-sdk/v2, client package
- **Reuses:** apiClient pattern, logging package

### Component 6: Terraform Resource - NAT Masquerade (`internal/provider/resource_rtx_nat_masquerade.go`)

- **Purpose:** Terraform resource implementation for NAT masquerade
- **Interfaces:**
  ```go
  func resourceRTXNATMasquerade() *schema.Resource
  func resourceRTXNATMasqueradeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATMasqueradeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATMasqueradeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATMasqueradeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
  func resourceRTXNATMasqueradeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error)

  // Helper functions
  func buildNATMasqueradeFromResourceData(d *schema.ResourceData) client.NATMasquerade
  func expandStaticEntries(entries []interface{}) []client.MasqueradeStaticEntry
  func flattenStaticEntries(entries []client.MasqueradeStaticEntry) []interface{}
  func parseNATMasqueradeID(id string) (int, error)
  func validateIPRange(v interface{}, k string) ([]string, []error)
  func validateIPAddress(v interface{}, k string) ([]string, []error)
  ```
- **Dependencies:** terraform-plugin-sdk/v2, client package
- **Reuses:** apiClient pattern, logging package

### Component 7: Client Interface Extension (`internal/client/interfaces.go`)

- **Purpose:** Defines NAT methods on Client interface
- **Interfaces:**
  ```go
  // NAT Static methods
  GetNATStatic(ctx context.Context, descriptorID int) (*NATStatic, error)
  CreateNATStatic(ctx context.Context, nat NATStatic) error
  UpdateNATStatic(ctx context.Context, nat NATStatic) error
  DeleteNATStatic(ctx context.Context, descriptorID int) error
  ListNATStatics(ctx context.Context) ([]NATStatic, error)

  // NAT Masquerade methods
  GetNATMasquerade(ctx context.Context, descriptorID int) (*NATMasquerade, error)
  CreateNATMasquerade(ctx context.Context, nat NATMasquerade) error
  UpdateNATMasquerade(ctx context.Context, nat NATMasquerade) error
  DeleteNATMasquerade(ctx context.Context, descriptorID int) error
  ListNATMasquerades(ctx context.Context) ([]NATMasquerade, error)
  ```
- **Dependencies:** context package
- **Reuses:** Existing interface pattern

---

## Data Models

### NATStatic (Client Layer)

```go
// NATStatic represents a static NAT descriptor configuration on an RTX router
type NATStatic struct {
    DescriptorID int              `json:"descriptor_id"` // NAT descriptor ID (1-65535)
    Entries      []NATStaticEntry `json:"entries,omitempty"`
}

// NATStaticEntry represents a single static NAT mapping entry
type NATStaticEntry struct {
    InsideLocal       string `json:"inside_local"`                  // Inside local IP address
    InsideLocalPort   *int   `json:"inside_local_port,omitempty"`   // Inside local port (for port NAT)
    OutsideGlobal     string `json:"outside_global"`                // Outside global IP address
    OutsideGlobalPort *int   `json:"outside_global_port,omitempty"` // Outside global port (for port NAT)
    Protocol          string `json:"protocol,omitempty"`            // Protocol: tcp, udp (for port NAT)
}
```

### NATMasquerade (Client Layer)

```go
// NATMasquerade represents a NAT masquerade configuration on an RTX router
type NATMasquerade struct {
    DescriptorID  int                     `json:"descriptor_id"`            // NAT descriptor ID (1-65535)
    OuterAddress  string                  `json:"outer_address"`            // "ipcp", interface name, or specific IP
    InnerNetwork  string                  `json:"inner_network"`            // IP range: "192.168.1.0-192.168.1.255"
    StaticEntries []MasqueradeStaticEntry `json:"static_entries,omitempty"` // Static port mappings
}

// MasqueradeStaticEntry represents a static port mapping entry for NAT masquerade
type MasqueradeStaticEntry struct {
    EntryNumber       int    `json:"entry_number"`                  // Entry number for identification
    InsideLocal       string `json:"inside_local"`                  // Internal IP address
    InsideLocalPort   *int   `json:"inside_local_port,omitempty"`   // Internal port (nil for protocol-only)
    OutsideGlobal     string `json:"outside_global"`                // External IP address (or "ipcp")
    OutsideGlobalPort *int   `json:"outside_global_port,omitempty"` // External port (nil for protocol-only)
    Protocol          string `json:"protocol,omitempty"`            // "tcp", "udp", "esp", "ah", "gre", or empty
}
```

### Terraform Schema - NAT Static

```hcl
resource "rtx_nat_static" "example" {
  descriptor_id = 1  # Required, ForceNew, 1-65535

  entry {                       # Required, List
    inside_local        = ""    # Required, IPv4
    inside_local_port   = 0     # Optional, 1-65535
    outside_global      = ""    # Required, IPv4
    outside_global_port = 0     # Optional, 1-65535
    protocol            = ""    # Optional, "tcp" or "udp"
  }
}
```

### Terraform Schema - NAT Masquerade

```hcl
resource "rtx_nat_masquerade" "example" {
  descriptor_id = 1                             # Required, ForceNew, 1-65535
  outer_address = "ipcp"                        # Required, "ipcp"/interface/IP
  inner_network = "192.168.1.0-192.168.1.255"   # Optional, IP range

  static_entry {                                # Optional, List
    entry_number        = 1                     # Required, >= 1
    inside_local        = "192.168.1.100"       # Required, IPv4
    inside_local_port   = 8080                  # Optional, 1-65535
    outside_global      = "ipcp"                # Optional, default "ipcp"
    outside_global_port = 80                    # Optional, 1-65535
    protocol            = "tcp"                 # Optional
  }
}
```

---

## RTX Command Mapping

### NAT Static Type Definition

```
nat descriptor type <id> static
```

Example: `nat descriptor type 1 static`

### NAT Static 1:1 Mapping

```
nat descriptor static <id> <outside_ip>=<inside_ip>
```

Example: `nat descriptor static 1 203.0.113.1=192.168.1.1`

### NAT Static Port-Based Mapping

```
nat descriptor static <id> <outside_ip>:<port>=<inside_ip>:<port> <protocol>
```

Example: `nat descriptor static 1 203.0.113.1:80=192.168.1.1:8080 tcp`

### NAT Masquerade Type Definition

```
nat descriptor type <id> masquerade
```

Example: `nat descriptor type 1 masquerade`

### NAT Masquerade Outer Address

```
nat descriptor address outer <id> <address>
```

Examples:
- `nat descriptor address outer 1 ipcp` (PPPoE-assigned)
- `nat descriptor address outer 1 pp1` (interface)
- `nat descriptor address outer 1 203.0.113.1` (specific IP)

### NAT Masquerade Inner Network

```
nat descriptor address inner <id> <range>
```

Example: `nat descriptor address inner 1 192.168.1.0-192.168.1.255`

### NAT Masquerade Static Entry (with ports)

```
nat descriptor masquerade static <id> <entry_num> <outer>:<port>=<inner>:<port> [protocol]
```

Example: `nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:8080 tcp`

### NAT Masquerade Static Entry (protocol-only)

```
nat descriptor masquerade static <id> <entry_num> <inner_ip> <protocol>
```

Example: `nat descriptor masquerade static 1000 1 192.168.1.253 esp`

### Delete Commands

```
no nat descriptor type <id>                           # Delete entire descriptor
no nat descriptor static <id> <mapping>               # Delete specific static mapping
no nat descriptor masquerade static <id> <entry_num>  # Delete specific masquerade entry
```

### Show Commands

```
# Note: RTX routers do not support grep -E (extended regex)
# Use simple wildcard patterns instead
show config | grep "nat descriptor.*<id>"                          # Show specific descriptor
show config | grep "nat descriptor"                                 # Show all descriptors
```

---

## Error Handling

### Error Scenarios

1. **Invalid Descriptor ID**
   - **Handling:** Validation error before command execution
   - **User Impact:** Clear error message with valid range (1-65535)

2. **Invalid IP Address**
   - **Handling:** Validation error in schema or service layer
   - **User Impact:** Error message indicating invalid IP format

3. **Port/Protocol Mismatch**
   - **Handling:** CustomizeDiff validation for static NAT, service validation for masquerade
   - **User Impact:** Clear error about required port/protocol combination

4. **Descriptor Not Found**
   - **Handling:** Read returns nil, clears from state; Delete ignores error
   - **User Impact:** Resource recreated on next apply

5. **Connection Failed**
   - **Handling:** Error propagated with context
   - **User Impact:** Error message indicating connection failure

6. **Command Execution Error**
   - **Handling:** containsError() checks output for "Error:" prefix
   - **User Impact:** Error with command output for debugging

7. **Context Cancelled**
   - **Handling:** Early return with context.Canceled error
   - **User Impact:** Operation aborted cleanly

---

## Testing Strategy

### Unit Testing

- **Parser Tests** (`nat_static_test.go`, `nat_masquerade_test.go`):
  - Test parsing various command output formats
  - Test command building for all scenarios
  - Test validation functions
  - Test edge cases (empty input, malformed output)
  - Test IsPortBasedNAT and IsProtocolOnly helpers

- **Service Tests** (`nat_static_service_test.go`, `nat_masquerade_service_test.go`):
  - Test Create with valid/invalid inputs
  - Test Get with found/not found scenarios
  - Test Update with differential changes
  - Test Delete including idempotency
  - Test List with empty/populated results
  - Test context cancellation
  - Test batch command generation
  - Use MockExecutor for isolation

### Integration Testing

- **Resource Tests** (`resource_rtx_nat_static_test.go`, `resource_rtx_nat_masquerade_test.go`):
  - Test schema validation
  - Test expand/flatten functions
  - Test import ID parsing
  - Test CustomizeDiff validation

### End-to-End Testing

- **Acceptance Tests** (with real RTX router):
  - Create NAT static with 1:1 mapping
  - Create NAT static with port-based mapping
  - Update NAT static entries
  - Delete NAT static
  - Import existing NAT static
  - Create NAT masquerade with basic config
  - Create NAT masquerade with static entries
  - Create NAT masquerade with protocol-only entries
  - Update NAT masquerade configuration
  - Delete NAT masquerade
  - Import existing NAT masquerade

---

## File Structure

```
internal/
├── provider/
│   ├── resource_rtx_nat_static.go           # Static NAT resource
│   ├── resource_rtx_nat_static_test.go      # Resource unit tests
│   ├── resource_rtx_nat_masquerade.go       # Masquerade resource
│   └── resource_rtx_nat_masquerade_test.go  # Resource unit tests
├── client/
│   ├── interfaces.go                        # NAT types and interface methods
│   ├── client.go                            # Service initialization
│   ├── nat_static_service.go                # Static NAT service
│   ├── nat_static_service_test.go           # Service unit tests
│   ├── nat_masquerade_service.go            # Masquerade service
│   └── nat_masquerade_service_test.go       # Service unit tests
└── rtx/
    └── parsers/
        ├── nat_static.go                    # Static NAT parser
        ├── nat_static_test.go               # Parser unit tests
        ├── nat_masquerade.go                # Masquerade parser
        └── nat_masquerade_test.go           # Parser unit tests
```

---

## Implementation Notes

1. **Batch Command Execution**: All create/update/delete operations use `RunBatch` to send multiple commands in a single SSH session, ensuring atomic configuration and reducing network overhead.

2. **Differential Updates**: The Update operation compares current state with desired state and only issues commands for actual changes, minimizing router configuration churn.

3. **Type Conversion**: Services maintain separate types from parsers (`client.NATStatic` vs `parsers.NATStatic`) to allow independent evolution and clearer API boundaries.

4. **Port Handling**: Static NAT entries use `int` for ports in parser layer but `*int` in client layer to distinguish "not set" from "0".

5. **Protocol-Only Entries**: NAT masquerade supports ESP, AH, GRE, ICMP protocols without port numbers for VPN passthrough scenarios.

6. **Save Configuration**: All mutating operations call `SaveConfig` to persist changes to non-volatile storage.

7. **Error String Detection**: The `containsError()` function checks for "Error:" prefix in command output to detect RTX command failures.

8. **Regex Patterns**: Parser uses multiple regex patterns to handle different command output formats (1:1 NAT, port NAT, alternate formats).

9. **Import Format**: Import uses simple descriptor_id integer, parsed and validated during import.

10. **CustomizeDiff**: Static NAT uses CustomizeDiff for complex port/protocol validation that can't be expressed in schema alone.

11. **RTX Grep Compatibility**: RTX routers do not support `grep -E` (extended regex) or the `\|` OR operator in grep patterns. All grep commands must use simple wildcard patterns (e.g., `grep "nat descriptor.*<id>"` instead of `grep -E "nat descriptor( <id>| <id>$)"`).

12. **OutsideGlobal Default**: For NAT masquerade static entries, when `OutsideGlobal` is not specified, it defaults to "ipcp" to indicate using the PPPoE-assigned address.

---

## State Handling

- **Resource ID**: String representation of descriptor_id
- **Persisted State**: All configuration attributes (descriptor_id, entries, outer_address, inner_network, static_entries)
- **Not Persisted**: Runtime statistics, active session counts
- **Drift Detection**: Full read on every plan operation
- **Missing Resource**: Cleared from state when "not found" during read

---

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation Analysis | Initial master design created from implementation code |
| 2026-01-23 | terraform-plan-differences-fix | Updated grep patterns for RTX compatibility; documented OutsideGlobal default to "ipcp" |
| 2026-02-01 | Implementation Audit | Update to Terraform Plugin Framework (not SDK v2) |
