# Terraform Provider for Yamaha RTX Routers

Terraform provider for managing Yamaha RTX series routers - enterprise-grade network routers widely used in Japan.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.11 (**required** for WriteOnly attribute support)
- Go >= 1.22 (for building from source)
- Yamaha RTX router with SSH access enabled

> **Important:** This provider requires Terraform 1.11 or later. The WriteOnly attribute feature used for sensitive fields (passwords, pre-shared keys) is not available in earlier versions.

## Migration Notes (v0.7+)

Starting with version 0.7, this provider has been migrated from Terraform SDK v2 to the Terraform Plugin Framework:

- **WriteOnly Attributes**: Sensitive fields (`password`, `admin_password`, `private_key`, `private_key_passphrase`, `pre_shared_key` in resources) now use WriteOnly attributes. These values are sent to the provider but never stored in Terraform state, improving security.
- **No Breaking Changes**: Resource schemas remain unchanged. Existing configurations will continue to work.
- **State File**: After upgrading, sensitive values will no longer appear in state files. This is expected behavior.

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
    }
  }
}
```

Then run `terraform init`.

### From GitHub Releases

1. Download the appropriate archive for your platform from [GitHub Releases](https://github.com/shin1ohno/terraform-provider-rtx/releases)
2. Extract and install to your Terraform plugins directory:

```bash
# Linux/macOS
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/shin1ohno/rtx/0.9.1/linux_amd64
unzip terraform-provider-rtx_0.9.1_linux_amd64.zip -d ~/.terraform.d/plugins/registry.terraform.io/shin1ohno/rtx/0.9.1/linux_amd64
```

3. Configure Terraform to use the local provider:

```hcl
terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "0.9.1"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/shin1ohno/terraform-provider-rtx.git
cd terraform-provider-rtx
make install
```

## Before You Begin

### SSH Host Key Setup

Before using this provider, you must configure SSH host key verification. Choose one of these options:

**Option 1: Add router's host key to known_hosts (Recommended)**

```bash
ssh-keyscan -t rsa 192.168.1.1 >> ~/.ssh/known_hosts
```

**Option 2: Skip host key verification (Testing only)**

```hcl
provider "rtx" {
  # ... other settings ...
  skip_host_key_check = true
}
```

## Quick Start

```hcl
terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
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
| **Security** | [access_list_ip](docs/resources/access_list_ip.md), [access_list_ipv6](docs/resources/access_list_ipv6.md), [access_list_ip_dynamic](docs/resources/access_list_ip_dynamic.md), [access_list_ipv6_dynamic](docs/resources/access_list_ipv6_dynamic.md), [access_list_mac](docs/resources/access_list_mac.md), [access_list_extended](docs/resources/access_list_extended.md), [access_list_extended_ipv6](docs/resources/access_list_extended_ipv6.md) |
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
| `RTX_ADMIN_PASSWORD` | Admin password for config changes |
| `RTX_USE_SFTP` | Enable SFTP-based config reading |
| `RTX_SFTP_CONFIG_PATH` | SFTP path to config file (auto-detected if empty) |
| `RTX_SSH_HOST_KEY` | SSH host public key (base64 encoded) |
| `RTX_KNOWN_HOSTS_FILE` | Path to known_hosts file (default: ~/.ssh/known_hosts) |
| `RTX_SKIP_HOST_KEY_CHECK` | Skip SSH host key verification (insecure) |
| `RTX_MAX_PARALLELISM` | Max concurrent operations (default: 4) |

#### Priority

Provider configuration takes precedence over environment variables:

```hcl
provider "rtx" {
  host = "192.168.1.1"  # This value is used, even if RTX_HOST is set
}
```

If a value is not set in the provider block, the environment variable is used as the default.

#### Usage Patterns

**Local development** - Use environment variables to avoid committing credentials:

```bash
export RTX_HOST="192.168.1.1"
export RTX_USERNAME="admin"
export RTX_PASSWORD="secret"
terraform plan
```

**CI/CD** - Set environment variables in your pipeline:

```yaml
# GitHub Actions example
env:
  RTX_HOST: ${{ secrets.RTX_HOST }}
  RTX_USERNAME: ${{ secrets.RTX_USERNAME }}
  RTX_PASSWORD: ${{ secrets.RTX_PASSWORD }}
```

**Multiple routers** - Use provider configuration with variables:

```hcl
variable "routers" {
  type = map(object({
    host     = string
    username = string
    password = string
  }))
}

provider "rtx" {
  alias    = "router1"
  host     = var.routers["router1"].host
  username = var.routers["router1"].username
  password = var.routers["router1"].password
}
```

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
