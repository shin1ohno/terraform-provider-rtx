# Requirements Document: Filter & NAT Enhancements

## Introduction

This specification addresses gaps in the terraform-provider-rtx's filtering and NAT functionality discovered during import reconciliation of actual RTX router configurations. The provider currently lacks support for several RTX-specific features including the `restrict` action for IP filters, protocol-specific filter types (tcpfin/tcprst), protocol-only NAT static entries (e.g., ESP), and Ethernet (MAC) filters. These features are essential for complete Infrastructure as Code management of RTX routers.

## Alignment with Product Vision

This feature directly supports the product vision of "comprehensive coverage of all RTX router features" and specifically addresses:

- **Security & Filtering**: Completing IP filter support with `restrict` action and protocol types
- **Ethernet Filters**: Layer 2 packet filtering capability mentioned in product.md
- **NAT/NAPT**: Protocol-only static mapping for IPsec/ESP support
- **Import Support**: Enabling complete import of existing RTX configurations

## Requirements

### REQ-1: IP Filter Restrict Action Support

**User Story:** As a network administrator, I want to use the `restrict` action in IP filters, so that I can implement stateful-like filtering behavior where specific traffic triggers dynamic filter activation.

#### RTX Command Reference (Validated from 08_IP_の設定.md)

**Command Syntax:**
```
ip filter filter_num pass_reject src_addr [/mask] [dest_addr [/mask] [protocol [src_port_list [dest_port_list]]]]
```

**Action (pass_reject) Variations:**
| Action | Description |
|--------|-------------|
| pass | Allow if matched (no log) |
| pass-log | Allow if matched (with log) |
| pass-nolog | Allow if matched (no log) |
| reject | Deny if matched (with log) |
| reject-log | Deny if matched (with log) |
| reject-nolog | Deny if matched (no log) |
| **restrict** | Allow if line connected, deny if disconnected (log on deny) |
| **restrict-log** | Allow if line connected, deny if disconnected (with log) |
| **restrict-nolog** | Allow if line connected, deny if disconnected (no log) |

**Examples:**
```
# Restrict action - used with dynamic filters for stateful behavior
ip filter 200026 restrict * * tcpfin * www,21,nntp
ip filter 200027 restrict * * tcprst * www,21,nntp
ip filter 500000 restrict * * * * *

# With logging variations
ip filter 500001 restrict-log * * * * *
ip filter 500002 restrict-nolog * * * * *
```

**Note:** The restrict action is useful for NTP packets and similar traffic that should only pass when the line is already connected, without triggering a connection.

#### Acceptance Criteria

1. WHEN creating an `rtx_access_list_ip` resource with `action = "restrict"` THEN the provider SHALL accept the configuration and generate the corresponding `ip filter <id> restrict ...` command
2. WHEN creating an `rtx_access_list_ip` resource with `action = "restrict-log"` or `action = "restrict-nolog"` THEN the provider SHALL also accept these variations
3. WHEN importing an existing IP filter with restrict action THEN the provider SHALL correctly parse and store the action in state
4. IF an IP filter has `restrict` action THEN the provider SHALL validate that it's used in conjunction with dynamic filters on the interface

### REQ-2: IP Filter Protocol Type Extensions

**User Story:** As a network administrator, I want to specify `tcpfin` and `tcprst` as protocol types in IP filters, so that I can create fine-grained TCP session control rules.

#### RTX Command Reference (Validated from 08_IP_の設定.md)

**Protocol Variations:**
| Protocol | Description |
|----------|-------------|
| tcp | All TCP packets (protocol number 6) |
| udp | All UDP packets (protocol number 17) |
| icmp | ICMP packets (protocol number 1) |
| icmp-error | ICMP error packets (TYPE 3-5, 11, 12) |
| icmp-info | ICMP info packets (TYPE 0, 8-10, 13-18, 30, 33-36) |
| gre | GRE protocol (protocol number 47) |
| esp | IPsec ESP (protocol number 50) |
| ah | IPsec AH (protocol number 51) |
| ipip | IP-in-IP encapsulation |
| **tcpsyn** | TCP packets with SYN flag set |
| **tcpfin** | TCP packets with FIN flag set |
| **tcprst** | TCP packets with RST flag set |
| **established** | TCP packets with ACK flag set (allows outbound connections, blocks inbound) |
| **tcpflag=FLAG/MASK** | Custom TCP flag matching (e.g., tcpflag!=0x0000/0x0007) |
| * | All protocols |
| (number) | Protocol number (1-255) |

**Examples:**
```
# tcpfin - matches TCP packets with FIN flag
ip filter 200026 restrict * * tcpfin * www,21,nntp

# tcprst - matches TCP packets with RST flag
ip filter 200027 restrict * * tcprst * www,21,nntp

# tcpsyn - matches TCP packets with SYN flag (connection initiation)
ip filter 100 reject * * tcpsyn * 1723

# established - allows return traffic for outbound connections
ip filter 101 pass * * established * *

# TCP flag matching with bitmask
ip filter 102 pass * * tcpflag!=0x0000/0x0007 * *
```

**Port Specification:**
- When protocol is TCP-based (tcp/tcpsyn/tcpfin/tcprst/established/tcpflag) or UDP, src_port_list and dest_port_list are port numbers
- When protocol is ICMP only, src_port_list is ICMP TYPE and dest_port_list is ICMP CODE
- Port lists can use: numbers, ranges (80-443), names (www, ftp), comma-separated lists

#### Acceptance Criteria

1. WHEN creating an `rtx_access_list_ip` resource with `protocol = "tcpfin"` THEN the provider SHALL accept the configuration and generate `ip filter <id> ... tcpfin ...`
2. WHEN creating an `rtx_access_list_ip` resource with `protocol = "tcprst"` THEN the provider SHALL accept the configuration and generate `ip filter <id> ... tcprst ...`
3. WHEN creating an `rtx_access_list_ip` resource with `protocol = "tcpsyn"` or `protocol = "established"` THEN the provider SHALL also accept these values
4. WHEN importing existing filters using tcpfin, tcprst, tcpsyn, or established THEN the provider SHALL correctly parse and store the protocol value

### REQ-3: Dynamic IP Filter Support

**User Story:** As a network administrator, I want to define dynamic IP filters in Terraform, so that I can configure stateful packet inspection rules for outbound traffic.

#### RTX Command Reference (Validated from 08_IP_の設定.md)

**Command Syntax (Two Forms):**

**Form 1 - Application/Protocol-based:**
```
ip filter dynamic dyn_filter_num srcaddr [/mask] dstaddr [/mask] protocol [option ...]
```

**Form 2 - Static filter reference-based:**
```
ip filter dynamic dyn_filter_num srcaddr [/mask] dstaddr [/mask] filter filter_list [in filter_list] [out filter_list] [option ...]
```

**Parameters:**
- `dyn_filter_num`: Dynamic filter number (1..21474836)
- `srcaddr`, `dstaddr`: IP address, FQDN, or * (any)
- `protocol`: Service/protocol name (see table below)
- `filter_list`: Static filter numbers defined by `ip filter` command
- `in`: Controls reverse direction access
- `out`: Controls same direction access as dynamic filter

**Available Protocols/Services:**
| Category | Values |
|----------|--------|
| Common | tcp, udp, ftp, tftp, domain, www, smtp, pop3, telnet |
| Extended | echo, discard, daytime, chargen, ssh, time, whois, dns |
| Network | gopher, finger, http, sunrpc, ident, nntp, ntp, ms-rpc |
| Microsoft | netbios_ns, netbios_dgm, netbios_ssn, ms-ds, ms-sql |
| Security | imap, imap3, ldap, https, ike, ipsec-nat-t |
| Remote | rlogin, rwho, rsh, printer |
| Routing | rip, ripng, bgp |
| Other | syslog, radius, l2tp, pptp, nfs, msblast, sip |
| Special | ping, ping6, submission, netmeeting |

**Options:**
| Option | Description |
|--------|-------------|
| syslog=on | Log connection history to syslog (default) |
| syslog=off | Do not log connection history |
| timeout=seconds | Connection timeout in seconds |

**Examples:**
```
# Form 1: Application-based dynamic filters
ip filter dynamic 200080 * * ftp syslog=off
ip filter dynamic 200081 * * domain syslog=off
ip filter dynamic 200082 * * www syslog=off
ip filter dynamic 200083 * * smtp syslog=off
ip filter dynamic 200084 * * pop3 syslog=off
ip filter dynamic 200085 * * submission syslog=off
ip filter dynamic 200098 * * tcp syslog=off
ip filter dynamic 200099 * * udp syslog=off

# Form 2: Static filter reference-based
ip filter 10 pass * * udp * snmp
ip filter dynamic 1 * * filter 10

# With specific source/destination
ip filter dynamic 100 192.168.1.0/24 * www syslog=on timeout=1800

# Usage on interface (static + dynamic filters combined)
ip lan2 secure filter out 200099 dynamic 200080 200081 200082 200083 200084 200085
```

**Dynamic Filter Timer (Global Settings):**
```
ip filter dynamic timer option=timeout [option=timeout ...]
```
| Timer Option | Default | Description |
|--------------|---------|-------------|
| tcp-syn-timeout | 30 | SYN timeout before disconnect |
| tcp-fin-timeout | 5 | FIN timeout before connection release |
| tcp-idle-time | 3600 | TCP idle timeout |
| udp-idle-time | 30 | UDP idle timeout |
| dns-timeout | 5 | DNS response timeout |

#### Acceptance Criteria

1. WHEN creating an `rtx_ip_filter_dynamic` resource with protocol/service THEN the provider SHALL generate the corresponding `ip filter dynamic <id> <src> <dst> <protocol> ...` command
2. WHEN creating an `rtx_ip_filter_dynamic` resource with filter references THEN the provider SHALL generate the corresponding `ip filter dynamic <id> <src> <dst> filter <list> ...` command
3. WHEN specifying a service type (ftp, domain, www, smtp, pop3, submission, tcp, udp, etc.) THEN the provider SHALL include it in the generated command
4. WHEN the dynamic filter includes `syslog=off` or `timeout=N` options THEN the provider SHALL include these in the generated command
5. WHEN importing existing dynamic filters THEN the provider SHALL correctly parse all parameters including both forms

### REQ-4: NAT Masquerade Protocol-Only Static Entries

**User Story:** As a network administrator, I want to create NAT masquerade static entries for protocols without port numbers (like ESP), so that I can configure IPsec passthrough for VPN connections.

#### RTX Command Reference (Validated from 23_NAT_機能.md)

**Command Syntax:**
```
nat descriptor masquerade static nat_descriptor id inner_ip protocol [outer_port=]inner_port
```

**Parameters:**
- `nat_descriptor`: NAT descriptor number (1..2147483647)
- `id`: Entry number within the descriptor
- `inner_ip`: Internal (private) IP address
- `protocol`: Protocol - esp, tcp, udp, icmp, or protocol number (1-255)
- `inner_port`: Internal port number (required for tcp/udp, optional for others)
- `outer_port=`: Optional outer port mapping (defaults to same as inner_port)

**Protocol-Only Support:**
When using protocols that don't have port numbers (esp, ah, gre, icmp), the port fields are omitted:
```
nat descriptor masquerade static <descriptor_id> <entry_num> <inner_ip> <protocol>
```

**Examples:**
```
# NAT masquerade descriptor configuration
nat descriptor type 1000 masquerade
nat descriptor address outer 1000 primary
nat descriptor masquerade incoming 1000 reject

# Protocol-only static entries (no port number)
nat descriptor masquerade static 1000 1 192.168.1.253 esp      # IPsec ESP passthrough
nat descriptor masquerade static 1000 2 192.168.1.253 ah       # IPsec AH passthrough
nat descriptor masquerade static 1000 3 192.168.1.253 gre      # GRE passthrough
nat descriptor masquerade static 1000 4 192.168.1.253 47       # Protocol number (GRE)

# Port-based static entries (with ports)
nat descriptor masquerade static 1000 10 192.168.1.253 udp 500          # IKE
nat descriptor masquerade static 1000 11 192.168.1.253 udp 4500         # NAT-T
nat descriptor masquerade static 1000 12 192.168.1.253 udp 1701         # L2TP
nat descriptor masquerade static 1000 900 192.168.1.20 tcp 55000        # Custom TCP

# With outer port mapping
nat descriptor masquerade static 1000 20 192.168.1.100 tcp 8080=80      # Map outer 8080 to inner 80
```

**Supported Protocols Without Port:**
| Protocol | Number | Use Case |
|----------|--------|----------|
| esp | 50 | IPsec Encapsulating Security Payload |
| ah | 51 | IPsec Authentication Header |
| gre | 47 | Generic Routing Encapsulation (PPTP) |
| icmp | 1 | ICMP echo passthrough |

#### Acceptance Criteria

1. WHEN creating an `rtx_nat_masquerade` with a static_entry specifying only `protocol = "esp"` (without ports) THEN the provider SHALL accept the configuration
2. WHEN creating an `rtx_nat_masquerade` with a static_entry specifying `protocol = "ah"`, `protocol = "gre"`, or protocol number THEN the provider SHALL also accept these
3. WHEN such entry is created THEN the provider SHALL generate `nat descriptor masquerade static <id> <entry_num> <ip> <protocol>`
4. WHEN importing existing protocol-only static entries THEN the provider SHALL correctly parse and store them without requiring port fields
5. WHEN the protocol requires port numbers (tcp/udp) THEN the provider SHALL require port fields

### REQ-5: Ethernet (MAC) Filter Support

**User Story:** As a network administrator, I want to define Ethernet (MAC address-based) filters in Terraform, so that I can control Layer 2 traffic based on MAC addresses.

#### RTX Command Reference (Validated from 09_イーサネットフィルタの設定.md)

**Command Syntax (Two Forms):**

**Form 1 - MAC address-based:**
```
ethernet filter num kind src_mac [dst_mac [offset byte_list]]
```

**Form 2 - DHCP-based:**
```
ethernet filter num kind type [scope] [offset byte_list]
```

**Parameters:**
- `num`: Filter number
  - 1..512 (RTX1210 Rev.14.01.16+, Rev.15.02+)
  - 1..100 (other firmware versions)
- `kind`: Action to take (see table below)
- `src_mac`: Source MAC address in `xx:xx:xx:xx:xx:xx` format or `*` (any)
- `dst_mac`: Destination MAC address (same format, optional, defaults to `*`)
- `type`: DHCP-based filter type (dhcp-bind, dhcp-not-bind)
- `scope`: DHCP scope number (1..65535) or IP address in scope
- `offset`: Byte offset from end of source MAC address (for advanced filtering)
- `byte_list`: Comma-separated hex bytes or `*` (max 16 bytes)

**Action (kind) Values:**
| Action | Description |
|--------|-------------|
| **pass-log** | Allow if matched (with log) |
| **pass-nolog** | Allow if matched (no log) |
| **reject-log** | Deny if matched (with log) |
| **reject-nolog** | Deny if matched (no log) |

**Note:** Unlike IP filters, ethernet filters do NOT have `pass`, `reject`, or `restrict` actions - only the `-log` and `-nolog` variants.

**DHCP Filter Types:**
| Type | Description |
|------|-------------|
| dhcp-bind | Match hosts with DHCP reservation in specified scope |
| dhcp-not-bind | Match hosts WITHOUT DHCP reservation in specified scope |

**Examples:**
```
# MAC address-based filters
ethernet filter 100 pass-nolog * *                              # Pass all (catch-all)
ethernet filter 1 reject-nolog bc:5c:17:05:59:3a *              # Block specific source MAC
ethernet filter 2 reject-nolog * bc:5c:17:05:59:3a              # Block specific dest MAC
ethernet filter 3 pass-log 00:11:22:33:44:55 aa:bb:cc:dd:ee:ff  # Pass specific MAC pair with log

# DHCP-based filters (match based on DHCP bindings)
ethernet filter 10 pass-nolog dhcp-bind                         # Pass DHCP reserved hosts
ethernet filter 11 reject-nolog dhcp-not-bind                   # Block non-reserved hosts
ethernet filter 12 pass-nolog dhcp-bind 1                       # Pass hosts in DHCP scope 1
ethernet filter 13 reject-nolog dhcp-not-bind 192.168.1.1       # Block non-reserved in scope containing IP

# Advanced byte-level filtering
ethernet filter 50 reject-nolog * * 0 08,00                     # Filter by EtherType

# Applying filters to interface
ethernet lan1 filter in 1 100                                   # Apply filters 1, 100 to inbound
ethernet lan1 filter out 2 100                                  # Apply filters 2, 100 to outbound
```

**Interface Application:**
```
ethernet interface filter dir list
```
- `interface`: LAN interface name (lan1, lan2, etc.) or tunnel interface (if in bridge)
- `dir`: Direction (in = from LAN, out = to LAN)
- `list`: Space-separated filter numbers (max 512 or 100 depending on firmware)

#### Acceptance Criteria

1. WHEN creating an `rtx_ethernet_filter` resource THEN the provider SHALL generate the corresponding `ethernet filter <id> <kind> <src_mac> [<dst_mac>]` command
2. WHEN the action is `pass-log`, `pass-nolog`, `reject-log`, or `reject-nolog` THEN the provider SHALL accept it
3. WHEN MAC addresses are specified as `*` (any) or specific MAC format (`xx:xx:xx:xx:xx:xx`) THEN the provider SHALL validate and accept them
4. WHEN creating a DHCP-based filter with `dhcp-bind` or `dhcp-not-bind` THEN the provider SHALL generate the corresponding command with optional scope
5. WHEN importing existing ethernet filters THEN the provider SHALL correctly parse all parameters including DHCP-based variants
6. WHEN ethernet filters are applied to interfaces (`ethernet lan1 filter in/out ...`) THEN this SHALL be managed via the interface resource or a separate `rtx_ethernet_filter_application` resource

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each new filter type should have its own parser module
- **Modular Design**: Reuse existing rtx_access_list_ip patterns for new protocol and action types
- **Dependency Management**: Dynamic filters may reference static filters by ID
- **Clear Interfaces**: Maintain consistent schema patterns across filter resources

### Performance
- No additional performance requirements beyond existing SSH command execution

### Security
- Filter configurations should be validated before sending to router
- No sensitive data (passwords) in filter configurations

### Reliability
- Parser must handle variations in RTX CLI output format
- Import must not fail on unknown filter configurations (log warning instead)

### Usability
- Schema should follow established RTX provider conventions
- Clear error messages for invalid filter configurations
- Examples in documentation for each new feature
