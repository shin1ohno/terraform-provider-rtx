# Requirements: rtx_ip_filter

## Overview
Terraform resource for managing IP packet filters (firewall rules) on Yamaha RTX routers.

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
resource "rtx_ip_filter" "allow_http" {
  number = 100
  action = "pass"

  source      = "*"
  destination = "192.168.1.0/24"
  protocol    = "tcp"
  dest_port   = "80,443"
}

resource "rtx_ip_filter" "block_all" {
  number = 999
  action = "reject"

  source      = "*"
  destination = "*"
  protocol    = "*"
}

resource "rtx_ip_filter_set" "wan_inbound" {
  number  = 1
  filters = [100, 101, 102, 999]
}

resource "rtx_ip_filter_apply" "wan" {
  interface = "pp1"
  direction = "in"
  filter_set = 1
}
```
