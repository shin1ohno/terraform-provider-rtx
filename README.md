# Terraform Provider for Yamaha RTX Routers

Terraform provider for managing Yamaha RTX series routers - enterprise-grade network routers widely used in Japan.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Go >= 1.22 (for building from source)
- Yamaha RTX router with SSH access enabled

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    rtx = {
      source = "sh1/rtx"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/sh1/terraform-provider-rtx.git
cd terraform-provider-rtx
make install
```

## Quick Start

```hcl
terraform {
  required_providers {
    rtx = {
      source = "sh1/rtx"
    }
  }
}

provider "rtx" {
  host     = "192.168.1.1"
  username = "admin"
  password = var.rtx_password
}

# Configure a VLAN
resource "rtx_vlan" "guest" {
  vlan_id    = 10
  interface  = "lan1"
  name       = "Guest Network"
  ip_address = "192.168.10.1"
  ip_mask    = "255.255.255.0"
}
```

## Documentation

- [Provider Configuration](docs/index.md)

### Resources

| Category | Resources |
|----------|-----------|
| **Interfaces** | [interface](docs/resources/interface.md), [bridge](docs/resources/bridge.md), [vlan](docs/resources/vlan.md), [pp_interface](docs/resources/pp_interface.md), [ipv6_interface](docs/resources/ipv6_interface.md), [ipv6_prefix](docs/resources/ipv6_prefix.md) |
| **Connectivity** | [pppoe](docs/resources/pppoe.md), [static_route](docs/resources/static_route.md), [bgp](docs/resources/bgp.md), [ospf](docs/resources/ospf.md) |
| **VPN** | [ipsec_tunnel](docs/resources/ipsec_tunnel.md), [ipsec_transport](docs/resources/ipsec_transport.md), [l2tp](docs/resources/l2tp.md), [l2tp_service](docs/resources/l2tp_service.md), [pptp](docs/resources/pptp.md) |
| **NAT** | [nat_masquerade](docs/resources/nat_masquerade.md), [nat_static](docs/resources/nat_static.md) |
| **Security** | [access_list_ip](docs/resources/access_list_ip.md), [access_list_ipv6](docs/resources/access_list_ipv6.md), [access_list_mac](docs/resources/access_list_mac.md), [access_list_extended](docs/resources/access_list_extended.md), [ethernet_filter](docs/resources/ethernet_filter.md), [ip_filter_dynamic](docs/resources/ip_filter_dynamic.md), [interface_acl](docs/resources/interface_acl.md) |
| **DHCP & DNS** | [dhcp_scope](docs/resources/dhcp_scope.md), [dhcp_binding](docs/resources/dhcp_binding.md), [dns_server](docs/resources/dns_server.md), [ddns](docs/resources/ddns.md), [netvolante_dns](docs/resources/netvolante_dns.md) |
| **QoS** | [class_map](docs/resources/class_map.md), [policy_map](docs/resources/policy_map.md), [service_policy](docs/resources/service_policy.md), [shape](docs/resources/shape.md) |
| **Services** | [sshd](docs/resources/sshd.md), [sftpd](docs/resources/sftpd.md), [httpd](docs/resources/httpd.md), [snmp_server](docs/resources/snmp_server.md), [syslog](docs/resources/syslog.md) |
| **Administration** | [admin](docs/resources/admin.md), [admin_user](docs/resources/admin_user.md), [system](docs/resources/system.md), [kron_schedule](docs/resources/kron_schedule.md), [kron_policy](docs/resources/kron_policy.md) |

## Examples

| Example | Description |
|---------|-------------|
| [bridge](examples/bridge/) | Layer 2 bridge configuration |
| [vlan](examples/vlan/) | 802.1Q VLAN setup |
| [pppoe](examples/pppoe/) | PPPoE connection for ISP |
| [ipv6_interface](examples/ipv6_interface/) | IPv6 with SLAAC and DHCPv6 |
| [ipsec_tunnel](examples/ipsec_tunnel/) | Site-to-site IPsec VPN |
| [l2tp](examples/l2tp/) | L2TP/L2TPv3 VPN tunnels |
| [nat_masquerade](examples/nat_masquerade/) | NAT/PAT configuration |
| [static_route](examples/static_route/) | Static routing |
| [dhcp](examples/dhcp/) | DHCP server setup |
| [dns_server](examples/dns_server/) | DNS server and forwarding |
| [qos](examples/qos/) | QoS with class/policy maps |
| [bgp](examples/bgp/) | BGP routing |
| [ospf](examples/ospf/) | OSPF routing |
| [schedule](examples/schedule/) | Scheduled tasks (kron) |
| [admin](examples/admin/) | Administrative settings |
| [import](examples/import/) | Comprehensive import example |

## Provider Configuration

```hcl
provider "rtx" {
  host     = "192.168.1.1"      # Router IP or hostname
  username = "admin"            # SSH username
  password = var.rtx_password   # SSH password (sensitive)
  port     = 22                 # SSH port (default: 22)
  timeout  = 30                 # Connection timeout in seconds

  # Optional: Use SFTP for faster config reading
  use_sftp = true

  # Optional: SSH session pooling for better performance
  ssh_session_pool {
    enabled      = true
    max_sessions = 2
    idle_timeout = "5m"
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `RTX_HOST` | Router hostname or IP |
| `RTX_USERNAME` | SSH username |
| `RTX_PASSWORD` | SSH password |
| `RTX_PORT` | SSH port |
| `RTX_ADMIN_PASSWORD` | Admin password for config changes |
| `RTX_USE_SFTP` | Enable SFTP-based config reading |
| `RTX_SKIP_HOST_KEY_CHECK` | Skip SSH host key verification (insecure) |

## Resource References

Resources provide computed attributes for referencing in other resources:

```hcl
# Interface resources
resource "rtx_interface" "lan1" {
  name = "lan1"
}

resource "rtx_pppoe" "wan" {
  pp_number      = 1
  bind_interface = rtx_interface.lan2.interface_name  # "lan2"
  # ...
}

resource "rtx_l2tp" "tunnel1" {
  tunnel_id = 1
  # ...
}

resource "rtx_bridge" "internal" {
  name    = "bridge1"
  members = [
    rtx_interface.lan1.interface_name,  # "lan1"
    rtx_l2tp.tunnel1.tunnel_interface   # "tunnel1"
  ]
}

# Use computed pp_interface for NAT
resource "rtx_nat_masquerade" "main" {
  descriptor_id = 1
  outer_address = rtx_pppoe.wan.pp_interface  # "pp1"
  inner_network = "192.168.1.0-192.168.1.255"
}
```

### Computed Interface Attributes

| Resource | Attribute | Example |
|----------|-----------|---------|
| `rtx_interface` | `interface_name` | `"lan1"` |
| `rtx_bridge` | `interface_name` | `"bridge1"` |
| `rtx_vlan` | `vlan_interface` | `"lan1/1"` |
| `rtx_pppoe` | `pp_interface` | `"pp1"` |
| `rtx_pp_interface` | `pp_interface` | `"pp1"` |
| `rtx_l2tp` | `tunnel_interface` | `"tunnel1"` |
| `rtx_ipsec_tunnel` | `tunnel_interface` | `"tunnel1"` |

## Importing Existing Configuration

```bash
# Import a VLAN
terraform import rtx_vlan.guest lan1/10

# Import a PPPoE connection
terraform import rtx_pppoe.wan 1

# Import an IPsec tunnel
terraform import rtx_ipsec_tunnel.vpn 1

# Import a bridge
terraform import rtx_bridge.internal bridge1
```

## Supported RTX Models

This provider is designed for Yamaha RTX series routers including:

- RTX830
- RTX1210
- RTX1220
- RTX3500
- RTX5000
- NVR510
- NVR700W

## Development

```bash
# Build
make build

# Run tests
make test

# Run acceptance tests (requires RTX router)
TF_ACC=1 make testacc

# Generate documentation
make docs

# Lint
make lint
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

[Contribution guidelines here]
