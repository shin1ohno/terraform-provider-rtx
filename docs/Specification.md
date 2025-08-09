# Terraform Provider RTX - Specification

## DHCP Static Lease (MAC Address-based IP Assignment) Implementation Plan

### Overview
This specification defines the implementation of DHCP static lease functionality for Yamaha RTX routers through Terraform. This feature allows administrators to assign fixed IP addresses to devices based on their MAC addresses using the `dhcp scope bind` command.

### Resource: `rtx_dhcp_binding`

#### Resource Design

```hcl
resource "rtx_dhcp_binding" "example" {
  scope_id    = 1
  ip_address  = "192.168.1.50"
  mac_address = "00:a0:c5:12:34:56"
  
  # Optional: Use Client-Identifier format
  use_client_identifier = false
}
```

#### Schema Definition

1. **Required Fields**:
   - `scope_id` (int): DHCP scope ID
   - `ip_address` (string): IP address to assign
   - `mac_address` (string): Device MAC address

2. **Optional Fields**:
   - `use_client_identifier` (bool): Use Client-Identifier (ethernet) format (default: false)

3. **Computed Fields**:
   - `id` (string): Resource unique identifier (format: `{scope_id}:{ip_address}`)

### Implementation Architecture

#### 1. Client Layer Extension

**`internal/client/interfaces.go`** additions:
```go
type DHCPBinding struct {
    ScopeID            int    `json:"scope_id"`
    IPAddress          string `json:"ip_address"`
    MACAddress         string `json:"mac_address"`
    UseClientIdentifier bool   `json:"use_client_identifier"`
}

type Client interface {
    // Existing methods...
    
    // DHCP Binding management
    GetDHCPBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error)
    CreateDHCPBinding(ctx context.Context, binding DHCPBinding) error
    DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error
}
```

**`internal/client/dhcp_service.go`** (new file):
```go
type DHCPService struct {
    executor Executor
}

func (s *DHCPService) CreateBinding(ctx context.Context, binding DHCPBinding) error
func (s *DHCPService) DeleteBinding(ctx context.Context, scopeID int, ipAddress string) error
func (s *DHCPService) ListBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error)
```

#### 2. Parser Implementation

**`internal/rtx/parsers/dhcp_bindings.go`** (new file):
- Parse `show dhcp scope bind {scope_id}` command output
- Support different RTX models (RTX830, RTX1210, etc.)

#### 3. Resource Implementation

**`internal/provider/resource_rtx_dhcp_binding.go`** (new file):
- CRUD operations:
  - **Create**: Execute `dhcp scope bind` command
  - **Read**: Verify binding with `show dhcp scope bind`
  - **Update**: Delete and recreate (RTX doesn't support direct update)
  - **Delete**: Execute `no dhcp scope bind` command
- Import functionality: Import by `{scope_id}:{ip_address}` format

### Implementation Steps

1. **Phase 1: Foundation**
   - Create DHCPService
   - Extend Client interface
   - Define DHCPBinding data structure

2. **Phase 2: Parser Implementation**
   - `show dhcp scope bind` output parser
   - Model-specific parser registration
   - Test case creation

3. **Phase 3: Resource Implementation**
   - TDD approach with test creation
   - CRUD lifecycle implementation
   - Acceptance test creation

4. **Phase 4: Docker Environment Extension**
   - Add DHCP-related command simulation
   - Prepare test data

### Error Handling and Validation

1. **Input Validation**:
   - MAC address format validation (allow colons, hyphens, or no separators)
   - IP address within scope range validation
   - Duplicate binding check

2. **Error Handling**:
   - Command execution errors
   - Parse errors
   - Network connection errors

### Security Considerations

1. Document MAC address spoofing risks
2. Emphasize proper DHCP scope configuration

### Testing Strategy

1. **Unit Tests**:
   - Parser tests (each RTX model)
   - Resource CRUD operation tests
   - Error handling tests

2. **Acceptance Tests**:
   - Basic binding creation/deletion
   - Multiple bindings management
   - Import functionality test

### Documentation

1. **Usage Examples**:
   ```hcl
   # DHCP scope configuration (prerequisite)
   # Configure on RTX router: dhcp scope 1 192.168.1.0/24

   # Fixed IP for printer
   resource "rtx_dhcp_binding" "printer" {
     scope_id    = 1
     ip_address  = "192.168.1.100"
     mac_address = "00:11:22:33:44:55"
   }

   # Managing multiple devices
   variable "static_devices" {
     type = map(object({
       ip_address  = string
       mac_address = string
     }))
   }

   resource "rtx_dhcp_binding" "devices" {
     for_each = var.static_devices
     
     scope_id    = 1
     ip_address  = each.value.ip_address
     mac_address = each.value.mac_address
   }
   ```

### Future Extensions

1. **DHCP Scope Resource** (`rtx_dhcp_scope`)
2. **DHCP Options** management
3. **Lease Information** read-only data source

### Implementation Priority

1. Basic DHCP binding CRUD functionality
2. Multiple scope support
3. Client-Identifier support
4. Wildcard (vendor OUI) support

### RTX Router Commands Reference

Based on official Yamaha documentation:

1. **Create binding**:
   ```
   dhcp scope bind <scope-id> <ip-address> <mac-address>
   ```
   Example: `dhcp scope bind 1 192.168.1.50 00:a0:c5:12:34:56`

2. **Create binding with Client-Identifier**:
   ```
   dhcp scope bind <scope-id> <ip-address> ethernet <mac-address>
   ```
   Example: `dhcp scope bind 1 192.168.1.50 ethernet 00:a0:c5:12:34:56`

3. **Delete binding**:
   ```
   no dhcp scope bind <scope-id> <ip-address>
   ```
   Example: `no dhcp scope bind 1 192.168.1.50`

4. **Show bindings**:
   ```
   show dhcp scope bind <scope-id>
   ```
   Example: `show dhcp scope bind 1`

### Notes

- MAC address format: Colons, hyphens, or no separators are accepted by RTX routers
- IP addresses must be within the configured DHCP scope range
- Only one IP can be bound to a MAC address within the same scope
- Bindings take effect immediately but existing leases are honored until expiry
- Configuration must be saved with `save` command to persist across reboots