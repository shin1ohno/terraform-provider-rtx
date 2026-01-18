# Requirements: rtx_ip_filter

## Overview
Terraform resource for managing IP packet filters (firewall rules) on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_access_list_extended`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Access list name |
| `entries` | `entries` | List of ACL entries |
| `sequence` | `sequence` | Entry sequence number |
| `ace_rule_action` | `ace_rule_action` | permit/deny (pass/reject) |
| `ace_rule_protocol` | `ace_rule_protocol` | Protocol (tcp/udp/icmp/ip) |
| `source_prefix` | `source_prefix` | Source network |
| `source_prefix_mask` | `source_prefix_mask` | Source wildcard mask |
| `destination_prefix` | `destination_prefix` | Destination network |
| `destination_port_equal` | `destination_port_equal` | Destination port |
| `log` | `log` | Enable logging |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Define IP filter rules
- **Read**: Query filter configuration
- **Update**: Modify filter rules
- **Delete**: Remove filter rules

### 2. Filter Definition
- Filter number (1-65535)
- Action: pass, reject, restrict, restrict-log
- Source address/network (with negation support)
- Destination address/network
- Protocol (tcp, udp, icmp, ip, etc.)
- Source port(s)
- Destination port(s)

### 3. Filter Set
- Group multiple filters into a set
- Apply filter set to interface
- Direction: in, out

### 4. Stateful Filtering (Dynamic Filter)
- Track connection state
- Allow return traffic automatically
- Session timeout configuration

### 5. Protocol-Specific Options
- TCP flags filtering (SYN, ACK, FIN, RST)
- ICMP type and code filtering
- Established connection matching

### 6. Logging
- Log matched packets
- Syslog integration

### 7. Import Support
- Import existing filter by number

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned filter rule changes |
| `terraform apply` | ✅ Required | Create, update, or delete IP filter rules |
| `terraform destroy` | ✅ Required | Remove filter rules from router |
| `terraform import` | ✅ Required | Import existing filter rules into state |
| `terraform refresh` | ✅ Required | Sync state with actual filter configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<filter_number>` (e.g., `100`)
- **Import Command**: `terraform import rtx_ip_filter.allow_http 100`
- **Post-Import**: All filter parameters must be populated from router

## Non-Functional Requirements

### 8. Validation
- Validate IP address formats
- Validate port ranges (1-65535)
- Validate protocol names
- Prevent filter number conflicts

### 9. Order Sensitivity
- Filters evaluated in order
- First match wins

## RTX Commands Reference
```
ip filter <n> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>]
ip filter dynamic <n> <src> <dst> <protocol> [<options>]
ip filter set <set_n> <filter_list>
ip <interface> secure filter <direction> <set>
```

## Example Usage
```hcl
# Access list - Cisco-compatible naming
resource "rtx_access_list_extended" "web_access" {
  name = "WEB_ACCESS"

  entries = [
    {
      sequence                 = 10
      remark                   = "Allow HTTP/HTTPS to internal servers"
      ace_rule_action          = "permit"
      ace_rule_protocol        = "tcp"
      source_prefix            = "0.0.0.0"
      source_prefix_mask       = "255.255.255.255"
      destination_prefix       = "192.168.1.0"
      destination_prefix_mask  = "0.0.0.255"
      destination_port_equal   = "80"
      log                      = true
    },
    {
      sequence                 = 20
      ace_rule_action          = "permit"
      ace_rule_protocol        = "tcp"
      source_prefix            = "0.0.0.0"
      source_prefix_mask       = "255.255.255.255"
      destination_prefix       = "192.168.1.0"
      destination_prefix_mask  = "0.0.0.255"
      destination_port_equal   = "443"
    },
    {
      sequence          = 999
      remark            = "Deny all other traffic"
      ace_rule_action   = "deny"
      ace_rule_protocol = "ip"
      source_any        = true
      destination_any   = true
    }
  ]
}

# Apply access list to interface
resource "rtx_interface_acl" "wan" {
  interface            = "pp1"
  ip_access_group_in   = "WEB_ACCESS"
}
```
