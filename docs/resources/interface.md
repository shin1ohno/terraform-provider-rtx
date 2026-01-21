# rtx_interface

Manages network interface configuration on Yamaha RTX routers.

This resource configures IP addresses, security filters, dynamic filters, ethernet (L2) filters, NAT descriptors, and other interface-level settings.

## Example Usage

### Basic Interface with IP Address

```hcl
resource "rtx_interface" "lan1" {
  name = "lan1"

  ip_address {
    address = "192.168.1.1/24"
  }

  description = "Internal LAN network"
}
```

### Interface with DHCP

```hcl
resource "rtx_interface" "lan2" {
  name = "lan2"

  ip_address {
    dhcp = true
  }

  description = "WAN interface with DHCP"
}
```

### Interface with Security Filters

```hcl
resource "rtx_interface" "lan1" {
  name = "lan1"

  ip_address {
    address = "192.168.1.1/24"
  }

  # Inbound security filters (first match wins)
  secure_filter_in = [100, 101, 102]

  # Outbound security filters
  secure_filter_out = [200, 201]

  # Dynamic stateful filters
  dynamic_filter_out = [300]
}
```

### Interface with Ethernet (L2) Filters

Ethernet filters operate at Layer 2 (MAC addresses) and can be applied to LAN interfaces for filtering based on source/destination MAC addresses and Ethernet types.

```hcl
# Define ethernet filters first
resource "rtx_ethernet_filter" "allow_known_macs" {
  number     = 1
  action     = "pass"
  source_mac = "00:11:22:33:44:55"
  dest_mac   = "*"
}

resource "rtx_ethernet_filter" "block_broadcast" {
  number     = 100
  action     = "reject"
  source_mac = "*"
  dest_mac   = "ff:ff:ff:ff:ff:ff"
}

# Apply ethernet filters to the interface
resource "rtx_interface" "lan1" {
  name = "lan1"

  ip_address {
    address = "192.168.1.1/24"
  }

  # Inbound ethernet (MAC) filters - order matters (first match wins)
  ethernet_filter_in = [
    rtx_ethernet_filter.allow_known_macs.number,
    rtx_ethernet_filter.block_broadcast.number
  ]

  # Outbound ethernet filters
  ethernet_filter_out = [100]
}
```

### Interface with NAT

```hcl
resource "rtx_nat_masquerade" "main" {
  descriptor_id = 1
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"
}

resource "rtx_interface" "lan2" {
  name = "lan2"

  ip_address {
    dhcp = true
  }

  nat_descriptor = rtx_nat_masquerade.main.descriptor_id
}
```

### Complete Example with All Filter Types

```hcl
resource "rtx_interface" "lan1" {
  name        = "lan1"
  description = "Internal LAN with comprehensive filtering"

  ip_address {
    address = "192.168.1.1/24"
  }

  # Layer 3/4 security filters
  secure_filter_in  = [100, 101, 102]
  secure_filter_out = [200, 201]

  # Dynamic stateful inspection
  dynamic_filter_out = [300, 301]

  # Layer 2 ethernet filters
  ethernet_filter_in  = [1, 2, 100]
  ethernet_filter_out = [50]

  # NAT configuration
  nat_descriptor = 1

  # Enable ProxyARP
  proxyarp = true

  # Custom MTU
  mtu = 1500
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) Interface name. Valid values are:
  - `lan1`, `lan2`, `lan3`, etc. - LAN interfaces
  - `bridge1`, `bridge2`, etc. - Bridge interfaces
  - `pp1`, `pp2`, etc. - PPP interfaces
  - `tunnel1`, `tunnel2`, etc. - Tunnel interfaces

* `description` - (Optional) Interface description.

* `ip_address` - (Optional) IP address configuration block. Contains:
  - `address` - (Optional) Static IP address in CIDR notation (e.g., `192.168.1.1/24`).
  - `dhcp` - (Optional) Set to `true` to use DHCP for IP address assignment. Either `address` or `dhcp` should be set, but not both.

* `secure_filter_in` - (Optional) List of inbound security filter numbers. Order matters - first match wins. These are IP-level (Layer 3/4) filters defined with `rtx_ip_filter` resources.

* `secure_filter_out` - (Optional) List of outbound security filter numbers. Order matters - first match wins.

* `dynamic_filter_out` - (Optional) List of dynamic filter numbers for outbound stateful inspection. Defined with `rtx_ip_filter_dynamic` resources.

* `ethernet_filter_in` - (Optional) List of inbound Ethernet (L2) filter numbers. These filters operate at Layer 2 using MAC addresses. Order matters - first match wins. Valid filter numbers are 1-512. Define filters using `rtx_ethernet_filter` resources.

* `ethernet_filter_out` - (Optional) List of outbound Ethernet (L2) filter numbers. Valid filter numbers are 1-512.

* `nat_descriptor` - (Optional) NAT descriptor ID to bind to this interface. Use `rtx_nat_masquerade` or `rtx_nat_static` to define the descriptor. Set to `0` or omit to disable NAT.

* `proxyarp` - (Optional) Enable Proxy ARP on this interface. Default is `false`.

* `mtu` - (Optional) Maximum Transmission Unit size (64-65535). Set to `0` to use the default MTU.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The interface name (same as `name`).

## Import

RTX interfaces can be imported using the interface name:

```shell
terraform import rtx_interface.lan1 lan1
```

## RTX Router Commands

This resource generates the following RTX router commands:

### IP Address Configuration

```
ip lan1 address 192.168.1.1/24
```

or for DHCP:

```
ip lan1 address dhcp
```

### Security Filters

```
ip lan1 secure filter in 100 101 102
ip lan1 secure filter out 200 201
```

### Dynamic Filters

```
ip lan1 dynamic filter out 300
```

### Ethernet Filters

```
ethernet lan1 filter in 1 2 100
ethernet lan1 filter out 50
```

### NAT Descriptor

```
ip lan1 nat descriptor 1
```

## Notes

- Ethernet filters can only be applied to LAN interfaces (lan1, lan2, etc.)
- Filter numbers in the lists are applied in order - the first matching filter determines the action
- Security filters are defined using `rtx_ip_filter` resources
- Ethernet filters are defined using `rtx_ethernet_filter` resources
- Dynamic filters provide stateful packet inspection and are defined using `rtx_ip_filter_dynamic` resources
