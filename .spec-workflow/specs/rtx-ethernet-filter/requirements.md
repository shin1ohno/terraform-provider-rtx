# Requirements: rtx_ethernet_filter

## Overview
Terraform resource for managing Ethernet (Layer 2) packet filters on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_access_list_mac` (MAC access lists)

## Cisco Compatibility

This resource follows Cisco MAC ACL naming patterns:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Access list name |
| `entries` | `entries` | List of filter entries |
| `action` | `ace_action` | permit/deny |
| `source_mac` | `source_address` | Source MAC address |
| `source_mac_mask` | `source_address_mask` | Source MAC mask |
| `destination_mac` | `destination_address` | Destination MAC address |
| `ethertype` | `ethertype` | Ethernet type filter |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Define Ethernet filter rules
- **Read**: Query filter configuration
- **Update**: Modify filter rules
- **Delete**: Remove filter rules

### 2. Filter Definition
- Filter number (1-65535)
- Action: pass, reject
- Source MAC address (with wildcards)
- Destination MAC address
- Ethernet type (0x0800 for IPv4, 0x0806 for ARP, etc.)
- VLAN ID filtering

### 3. MAC Address Matching
- Exact MAC address match
- Wildcard matching
- MAC address ranges

### 4. Protocol Filtering
- Filter by EtherType
- ARP filtering
- Non-IP protocol filtering

### 5. Interface Application
- Apply to LAN interface
- Direction: in, out
- Bridge interface support

### 6. Import Support
- Import existing filter by number

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned Ethernet filter changes |
| `terraform apply` | ✅ Required | Create, update, or delete Ethernet filters |
| `terraform destroy` | ✅ Required | Remove filter rules from router |
| `terraform import` | ✅ Required | Import existing filters into state |
| `terraform refresh` | ✅ Required | Sync state with actual filter configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<filter_number>` (e.g., `1`)
- **Import Command**: `terraform import rtx_ethernet_filter.allow_known_macs 1`
- **Post-Import**: All filter parameters must be populated from router

## Non-Functional Requirements

### 7. Validation
- Validate MAC address format
- Validate EtherType values
- Validate VLAN ID range (1-4094)

### 8. Order Sensitivity
- Filters evaluated in order
- First match wins

## RTX Commands Reference
```
ethernet filter <n> <action> <src_mac> <dst_mac> [<eth_type>]
ethernet <interface> filter <direction> <filter_list>
```

## Example Usage
```hcl
# MAC access list - Cisco-compatible naming
resource "rtx_access_list_mac" "trusted_macs" {
  name = "TRUSTED_MACS"

  entries = [
    {
      sequence            = 10
      ace_action          = "permit"
      source_address      = "0011.2200.0000"
      source_address_mask = "0000.00ff.ffff"
      destination_any     = true
    },
    {
      sequence              = 20
      ace_action            = "deny"
      source_any            = true
      destination_address   = "ffff.ffff.ffff"
      ethertype             = "0x0806"  # ARP
    }
  ]
}

# Apply MAC ACL to interface
resource "rtx_interface_mac_acl" "lan1" {
  interface             = "lan1"
  mac_access_group_in   = "TRUSTED_MACS"
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
