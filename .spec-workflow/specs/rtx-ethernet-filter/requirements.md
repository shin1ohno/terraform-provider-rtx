# Requirements: rtx_ethernet_filter

## Overview
Terraform resource for managing Ethernet (Layer 2) packet filters on Yamaha RTX routers.

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
resource "rtx_ethernet_filter" "allow_known_macs" {
  number = 1
  action = "pass"

  source_mac      = "00:11:22:*:*:*"
  destination_mac = "*"
}

resource "rtx_ethernet_filter" "block_broadcast" {
  number = 10
  action = "reject"

  source_mac      = "*"
  destination_mac = "ff:ff:ff:ff:ff:ff"
  ethernet_type   = "0x0806"  # ARP
}

resource "rtx_ethernet_filter_apply" "lan1" {
  interface = "lan1"
  direction = "in"
  filters   = [1, 10]
}
```
