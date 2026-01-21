# rtx_ipv6_filter_dynamic

Manages IPv6 dynamic (stateful) IP filters on Yamaha RTX routers.

Dynamic filters provide stateful packet inspection for common protocols, automatically allowing return traffic for established connections. This is essential for implementing proper firewall rules that allow outbound connections while blocking unsolicited inbound traffic.

## Example Usage

### Basic Dynamic Filter for Web Traffic

```hcl
resource "rtx_ipv6_filter_dynamic" "main" {
  entry {
    number      = 101080
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = false
  }
}
```

### Multiple Protocol Filters

```hcl
resource "rtx_ipv6_filter_dynamic" "outbound" {
  # FTP traffic
  entry {
    number      = 101020
    source      = "*"
    destination = "*"
    protocol    = "ftp"
    syslog      = false
  }

  # Web traffic (HTTP/HTTPS)
  entry {
    number      = 101080
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = false
  }

  # SMTP email
  entry {
    number      = 101025
    source      = "*"
    destination = "*"
    protocol    = "smtp"
    syslog      = false
  }

  # DNS queries
  entry {
    number      = 101053
    source      = "*"
    destination = "*"
    protocol    = "dns"
    syslog      = false
  }
}
```

### Email Submission Protocol (Port 587)

The `submission` protocol is used for email client to mail server communication on port 587, as defined in RFC 4409/6409.

```hcl
resource "rtx_ipv6_filter_dynamic" "email" {
  # Traditional SMTP (port 25) for server-to-server
  entry {
    number      = 101025
    source      = "*"
    destination = "*"
    protocol    = "smtp"
    syslog      = false
  }

  # Email Submission (port 587) for client-to-server
  entry {
    number      = 101587
    source      = "*"
    destination = "*"
    protocol    = "submission"
    syslog      = false
  }

  # POP3 for email retrieval
  entry {
    number      = 101110
    source      = "*"
    destination = "*"
    protocol    = "pop3"
    syslog      = false
  }
}
```

### Comprehensive Outbound Filter Set

```hcl
resource "rtx_ipv6_filter_dynamic" "comprehensive" {
  # FTP (with data channel handling)
  entry {
    number      = 101020
    source      = "*"
    destination = "*"
    protocol    = "ftp"
    syslog      = false
  }

  # SSH
  entry {
    number      = 101022
    source      = "*"
    destination = "*"
    protocol    = "ssh"
    syslog      = false
  }

  # Telnet
  entry {
    number      = 101023
    source      = "*"
    destination = "*"
    protocol    = "telnet"
    syslog      = false
  }

  # SMTP (server-to-server)
  entry {
    number      = 101025
    source      = "*"
    destination = "*"
    protocol    = "smtp"
    syslog      = false
  }

  # DNS
  entry {
    number      = 101053
    source      = "*"
    destination = "*"
    protocol    = "dns"
    syslog      = false
  }

  # HTTP/HTTPS
  entry {
    number      = 101080
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = false
  }

  # POP3
  entry {
    number      = 101110
    source      = "*"
    destination = "*"
    protocol    = "pop3"
    syslog      = false
  }

  # Email Submission (port 587)
  entry {
    number      = 101587
    source      = "*"
    destination = "*"
    protocol    = "submission"
    syslog      = false
  }

  # Generic TCP (catch-all for other protocols)
  entry {
    number      = 101999
    source      = "*"
    destination = "*"
    protocol    = "tcp"
    syslog      = false
  }

  # Generic UDP
  entry {
    number      = 102000
    source      = "*"
    destination = "*"
    protocol    = "udp"
    syslog      = false
  }
}
```

### With Syslog Enabled for Monitoring

```hcl
resource "rtx_ipv6_filter_dynamic" "monitored" {
  entry {
    number      = 101080
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = true  # Log matching traffic
  }

  entry {
    number      = 101022
    source      = "*"
    destination = "*"
    protocol    = "ssh"
    syslog      = true  # Log SSH connections
  }
}
```

## Argument Reference

The following arguments are supported:

* `entry` - (Required) List of dynamic filter entries. Each entry supports:

  * `number` - (Required) Filter number (1-65535). This is a unique identifier for the filter entry.

  * `source` - (Required) Source IPv6 address or `*` for any.

  * `destination` - (Required) Destination IPv6 address or `*` for any.

  * `protocol` - (Required) Protocol type. Valid values are:
    - `ftp` - FTP with data channel tracking (ports 20, 21)
    - `www` - HTTP/HTTPS web traffic (ports 80, 443)
    - `smtp` - Simple Mail Transfer Protocol (port 25)
    - `submission` - Email submission protocol (port 587, RFC 6409)
    - `pop3` - Post Office Protocol v3 (port 110)
    - `dns` - Domain Name System (port 53)
    - `domain` - Alias for `dns`
    - `telnet` - Telnet protocol (port 23)
    - `ssh` - Secure Shell (port 22)
    - `tcp` - Generic TCP traffic
    - `udp` - Generic UDP traffic
    - `*` - Any protocol

  * `syslog` - (Optional) Enable syslog logging for this filter. Default is `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource identifier (`ipv6_filter_dynamic`).

## Import

IPv6 dynamic filters can be imported using the resource ID:

```shell
terraform import rtx_ipv6_filter_dynamic.main ipv6_filter_dynamic
```

## RTX Router Commands

This resource generates the following RTX router commands:

```
ipv6 filter dynamic 101020 * * ftp syslog=off
ipv6 filter dynamic 101080 * * www syslog=off
ipv6 filter dynamic 101587 * * submission syslog=off
```

With syslog enabled:

```
ipv6 filter dynamic 101080 * * www syslog=on
```

## Notes

### Protocol Details

| Protocol | Port(s) | Description |
|----------|---------|-------------|
| `ftp` | 20, 21 | FTP with active/passive mode handling |
| `www` | 80, 443 | HTTP and HTTPS web traffic |
| `smtp` | 25 | Server-to-server email relay |
| `submission` | 587 | Client-to-server email submission (RFC 6409) |
| `pop3` | 110 | Email retrieval |
| `dns` | 53 | DNS queries (TCP and UDP) |
| `telnet` | 23 | Telnet remote access |
| `ssh` | 22 | Secure Shell |
| `tcp` | Any | Generic TCP connection tracking |
| `udp` | Any | Generic UDP session tracking |

### Email Protocol Recommendations

For modern email configurations:

- Use `submission` (port 587) for email clients sending mail to your server
- Use `smtp` (port 25) only for server-to-server relay
- The `submission` protocol requires authentication (SMTP AUTH) and typically uses STARTTLS encryption

### Applying Dynamic Filters

Dynamic filters must be applied to interfaces using the `rtx_ipv6_interface` resource:

```hcl
resource "rtx_ipv6_interface" "lan1" {
  interface = "lan1"

  # Apply dynamic filters for outbound traffic
  dynamic_filter_out = [
    101020,  # FTP
    101080,  # WWW
    101587,  # Submission
  ]
}
```

## Related Resources

- [rtx_ipv6_filter](./ipv6_filter.md) - Static IPv6 packet filters
- [rtx_ipv6_interface](./ipv6_interface.md) - IPv6 interface configuration
- [rtx_ip_filter_dynamic](./ip_filter_dynamic.md) - IPv4 dynamic filters
