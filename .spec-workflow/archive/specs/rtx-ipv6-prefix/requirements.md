# Requirements: rtx_ipv6_prefix

## Overview
Terraform resource for managing IPv6 prefix definitions on Yamaha RTX routers. IPv6 prefixes are used for address assignment (SLAAC), Router Advertisement, and DHCPv6 prefix delegation.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Define IPv6 prefix
- **Read**: Query current prefix configuration
- **Update**: Modify prefix parameters
- **Delete**: Remove prefix definition

### 2. Static Prefix
- Define static IPv6 prefix
- Configure prefix length (e.g., /64)

### 3. RA-Derived Prefix
- Define prefix from upstream Router Advertisement
- Specify source interface (WAN interface)
- Commonly used with Japanese IPoE/MAP-E ISPs

### 4. DHCPv6 Prefix Delegation
- Define prefix from DHCPv6-PD
- Specify source interface
- Support delegated prefix lengths

### 5. Prefix Identification
- Unique prefix ID (1-255)
- Reference by ID in interface configuration

### 6. Import Support
- Import existing prefix definition by ID

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned prefix configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete prefix |
| `terraform destroy` | ✅ Required | Remove prefix definition |
| `terraform import` | ✅ Required | Import existing prefix into state |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<prefix_id>` (e.g., `1`)
- **Import Command**: `terraform import rtx_ipv6_prefix.ra_prefix 1`
- **Post-Import**: All attributes must be populated

## Non-Functional Requirements

### 7. Validation
- Validate prefix ID is in range 1-255
- Validate IPv6 prefix format
- Validate prefix length is in range 1-128
- Validate source interface exists (for RA/PD)

### 8. Dependencies
- Referenced by `rtx_ipv6_interface` for address assignment

## RTX Commands Reference
```
ipv6 prefix <id> <prefix>::/<length>
ipv6 prefix <id> ra-prefix@<interface>::/<length>
ipv6 prefix <id> dhcp-prefix@<interface>::/<length>
no ipv6 prefix <id>
show config | grep "ipv6 prefix"
```

## Example Usage
```hcl
# Static prefix
resource "rtx_ipv6_prefix" "static" {
  id            = 1
  source        = "static"
  prefix        = "2001:db8:1234::"
  prefix_length = 64
}

# RA-derived prefix (from upstream ISP)
resource "rtx_ipv6_prefix" "ra_prefix" {
  id            = 2
  source        = "ra"
  interface     = "lan2"
  prefix_length = 64
}

# DHCPv6 Prefix Delegation
resource "rtx_ipv6_prefix" "pd" {
  id            = 3
  source        = "dhcpv6-pd"
  interface     = "lan2"
  prefix_length = 48
}

# Use prefix in interface
resource "rtx_ipv6_interface" "lan" {
  interface = "lan1"

  address {
    prefix_ref   = rtx_ipv6_prefix.ra_prefix.id
    interface_id = "::1"
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
