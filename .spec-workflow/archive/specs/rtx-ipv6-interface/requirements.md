# Requirements: rtx_ipv6_interface

## Overview
Terraform resource for managing IPv6 interface configurations on Yamaha RTX routers. This includes IPv6 address assignment, Router Advertisement (RA), and DHCPv6 server/client configuration.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure IPv6 interface settings
- **Read**: Query current IPv6 interface configuration
- **Update**: Modify IPv6 interface parameters
- **Delete**: Remove IPv6 configuration from interface

### 2. IPv6 Address Configuration
- Static IPv6 address with prefix length
- Prefix-based address using `rtx_ipv6_prefix` reference
- Interface identifier (::1, ::2, etc.)

### 3. Router Advertisement (RA) Configuration
- Enable/disable RA on interface
- Configure prefix to advertise
- O flag (Other config - DHCPv6 for options)
- M flag (Managed address - DHCPv6 for addresses)
- Router lifetime

### 4. DHCPv6 Configuration
- DHCPv6 server mode
- DHCPv6 client mode
- Information Request (IR) mode

### 5. IPv6 MTU Configuration
- Configure IPv6-specific MTU size

### 6. IPv6 Security Filters
- Inbound security filters
- Outbound security filters
- Dynamic filters for stateful inspection

### 7. Import Support
- Import existing IPv6 interface configuration by name

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned IPv6 interface changes |
| `terraform apply` | ✅ Required | Create, update, or delete IPv6 settings |
| `terraform destroy` | ✅ Required | Remove IPv6 configuration from interface |
| `terraform import` | ✅ Required | Import existing IPv6 interface config |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<interface_name>` (e.g., `lan1`)
- **Import Command**: `terraform import rtx_ipv6_interface.lan lan1`
- **Post-Import**: All attributes must be populated

## Non-Functional Requirements

### 8. Validation
- Validate interface name format
- Validate IPv6 address format
- Validate prefix reference exists
- Validate filter numbers

### 9. Dependencies
- IPv6 prefixes (use `rtx_ipv6_prefix` for prefix definitions)
- IPv6 filters (use `rtx_ip_filter` for filter definitions)

## RTX Commands Reference
```
ipv6 <interface> address <address>/<prefix_length>
ipv6 <interface> address <prefix_ref>::<interface_id>/<length>
ipv6 <interface> rtadv send <prefix_id> [o_flag=on|off] [m_flag=on|off]
ipv6 <interface> dhcp service server
ipv6 <interface> dhcp service client [ir=on|off]
ipv6 <interface> mtu <size>
ipv6 <interface> secure filter in <filter_list>
ipv6 <interface> secure filter out <filter_list> [dynamic <dynamic_filter_list>]
```

## Example Usage
```hcl
# WAN interface with DHCPv6 client
resource "rtx_ipv6_interface" "wan" {
  interface = "lan2"

  dhcpv6_service = "client"

  mtu = 1500

  secure_filter_in   = [101000, 101002, 101099]
  secure_filter_out  = [101099]
  dynamic_filter_out = [101080, 101081, 101082]
}

# LAN interface with RA and DHCPv6 server
resource "rtx_ipv6_interface" "lan" {
  interface = "lan1"

  address {
    prefix_ref   = 1
    interface_id = "::2"
  }

  rtadv {
    enabled   = true
    prefix_id = 1
    o_flag    = true
    m_flag    = false
  }

  dhcpv6_service = "server"
}

# Bridge interface with prefix-based address
resource "rtx_ipv6_interface" "bridge" {
  interface = "bridge1"

  address {
    prefix_ref   = 1
    interface_id = "::1"
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
