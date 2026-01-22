# Requirements Document: Parser Import Bugs

## Introduction

Terraform import機能のテスト中に複数のパーサーバグが発見されました。これらのバグにより、RTXルータからの設定のインポートが不完全になり、`terraform plan`で差分が発生します。

本specは、発見された5つのパーサーバグを修正し、完全なterraformインポートを実現することを目的とします。

## Alignment with Product Vision

terraform-provider-rtxの目標は、RTXルータの設定を完全にTerraformで管理できるようにすることです。インポート機能が正しく動作しないと、既存の設定をTerraformに移行できず、プロバイダーの価値が大幅に低下します。

## Requirements

### Requirement 1: Filter List Parsing Bug Fix

**User Story:** As a network administrator, I want to import interface configurations with long filter lists, so that all security filters are correctly imported into Terraform state.

#### RTX Command Format

```
# IPv4 secure filter (static filters)
ip lan2 secure filter in 200020 200021 200022 200023 200024 200025 200103 200100
 200102 200104 200101 200105 200099
ip lan2 secure filter out 200000 200001 200002 200010 200011 200012 200026 200027
 200099

# IPv4 secure filter with dynamic filters
ip lan2 secure filter out 200000 200001 200002 200010 200011 200012 200026 200027
 200099 dynamic 200080 200081 200082 200083 200084 200085

# IPv6 secure filter
ipv6 lan2 secure filter in 101000 101002 402100 101098 101099
ipv6 lan2 secure filter out 101098 101099 dynamic 101080 101081 101082 101083
 101084 101085 101098 101099
```

Note: 行が80文字を超えると、次の行の先頭にスペースが1つ入って継続される。

#### Acceptance Criteria

1. WHEN an interface has a secure_filter_in list with 10+ filter numbers THEN the parser SHALL import all filter numbers without truncation
2. WHEN an interface has a secure_filter_out list with dynamic filters THEN the parser SHALL correctly separate static and dynamic filter numbers
3. WHEN the RTX output wraps filter numbers to multiple lines THEN the preprocessWrappedLines function SHALL correctly join continuation lines
4. IF a filter number is 200100 THEN the parser SHALL NOT parse it as 20010 (truncation bug)

**Affected Files:**
- `internal/rtx/parsers/interface_config.go` - IPv4 interface parsing
- `internal/rtx/parsers/ipv6_interface.go` - IPv6 interface parsing

**Root Cause Analysis:**
- RTX output wraps long filter lists at terminal width (80 chars)
- Continuation lines start with numbers but may include "dynamic" keyword
- Current regex `^\d` may incorrectly identify non-continuation lines

### Requirement 2: DNS server_select Parsing Bug Fix

**User Story:** As a network administrator, I want to import DNS server configurations with multiple servers and query patterns, so that DNS resolution works correctly after import.

#### RTX Command Format

```
# 基本形式: dns server select <id> <server1> [<server2>] [edns=on] [<record_type>] <query_pattern> [<original_sender>]
dns server select 500000 1.1.1.1 1.0.0.1 edns=on a .
dns server select 500100 2606:4700:4700::1111 2606:4700:4700::1001 edns=on aaaa .

# より単純な例
dns server select 1 192.168.1.1 .internal.example.com
dns server select 2 8.8.8.8 8.8.4.4 .
```

Note: フィールドの順序が重要。server1, server2, edns, record_type, query_pattern, original_senderの順。

#### Acceptance Criteria

1. WHEN a dns server select entry has 2 servers THEN the parser SHALL import both server IP addresses
2. WHEN a dns server select entry has query_pattern "." THEN the parser SHALL import "." not the second server IP
3. WHEN a dns server select entry has record_type "aaaa" THEN the parser SHALL import "aaaa" not "a"
4. WHEN a dns server select entry has edns=on THEN the parser SHALL set EDNS to true

**Affected Files:**
- `internal/rtx/parsers/dns.go` - DNS configuration parsing

**Root Cause Analysis:**
- `parseDNSServerSelectFields` function parses fields in incorrect order
- Second server IP is being parsed as query_pattern or original_sender
- record_type detection may fail for certain patterns

### Requirement 3: DHCP Scope Parsing Bug Fix

**User Story:** As a network administrator, I want to import DHCP scope configurations completely, so that DHCP service continues to work after Terraform management begins.

#### RTX Command Format

```
# DHCP scope definition
dhcp scope 1 192.168.0.0/16 expire 24:00

# DHCP scope options
dhcp scope option 1 default_gateway=192.168.1.1
dhcp scope option 1 dns=192.168.1.1,1.1.1.1

# DHCP scope bind (MAC address binding)
dhcp scope bind 1 192.168.1.50 00:30:93:11:0e:33
dhcp scope bind 1 192.168.1.60 ethernet 00:3e:e1:c3:54:b4
```

Note: `show config` の出力では `dhcp scope 1 192.168.0.0/16` のようにネットワークCIDRが含まれる。

#### Acceptance Criteria

1. WHEN a DHCP scope has a network configuration THEN the parser SHALL import the network CIDR
2. WHEN a DHCP scope has a lease_time/expire configuration THEN the parser SHALL import the lease time
3. WHEN a DHCP scope has dns options THEN the parser SHALL import all DNS servers
4. IF the scope regex pattern doesn't match THEN the parser SHALL log a warning and attempt alternative parsing

**Affected Files:**
- `internal/rtx/parsers/dhcp_scope.go` - DHCP scope parsing
- `internal/provider/resource_dhcp_scope.go` - DHCP scope resource read function

**Root Cause Analysis:**
- Scope pattern regex may be too strict for some RTX output formats
- Network field not being populated in imported state

### Requirement 4: NAT Masquerade Import Implementation

**User Story:** As a network administrator, I want to import NAT masquerade configurations, so that NAT rules are preserved when moving to Terraform management.

#### RTX Command Format

```
# NAT descriptor type definition
nat descriptor type 1000 masquerade

# NAT descriptor outer address
nat descriptor address outer 1000 primary

# NAT masquerade static entries
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.253 udp 500
nat descriptor masquerade static 1000 3 192.168.1.253 udp 4500
nat descriptor masquerade static 1000 4 192.168.1.253 udp 1701
nat descriptor masquerade static 1000 900 192.168.1.20 tcp 55000
```

Note: `nat descriptor masquerade static <descriptor_id> <entry_number> <inside_local> <protocol> [<port>]`

#### Acceptance Criteria

1. WHEN importing a NAT masquerade resource THEN the importer SHALL find the NAT descriptor by ID
2. WHEN a NAT masquerade has static entries THEN the parser SHALL import all static NAT mappings
3. IF the NAT masquerade descriptor ID does not exist THEN the importer SHALL return a clear error message

**Affected Files:**
- `internal/rtx/parsers/nat.go` - NAT configuration parsing (if exists)
- `internal/provider/resource_nat_masquerade.go` - NAT masquerade resource

**Root Cause Analysis:**
- Import function returns "not found" even when descriptor exists
- Read function may not be parsing all NAT configuration lines

### Requirement 5: IPv6 Filter Dynamic Read Implementation

**User Story:** As a network administrator, I want to import IPv6 dynamic filter configurations, so that IPv6 traffic filtering is preserved in Terraform state.

#### RTX Command Format

```
# IPv6 dynamic filter entries
ipv6 filter dynamic 101080 * * ftp
ipv6 filter dynamic 101081 * * domain
ipv6 filter dynamic 101082 * * www
ipv6 filter dynamic 101083 * * smtp
ipv6 filter dynamic 101084 * * pop3
ipv6 filter dynamic 101085 * * submission
ipv6 filter dynamic 101098 * * tcp
ipv6 filter dynamic 101099 * * udp
```

Note: `ipv6 filter dynamic <number> <source> <destination> <protocol> [syslog=on]`
プロトコルには tcp, udp, ftp, domain, www, smtp, pop3, submission などが使用可能。

#### Acceptance Criteria

1. WHEN importing an ipv6_filter_dynamic resource THEN the read function SHALL return the current filter entries
2. WHEN an IPv6 dynamic filter has multiple entries THEN the parser SHALL import all entries with their properties
3. IF no IPv6 dynamic filters exist THEN the read function SHALL return an empty configuration not an error

**Affected Files:**
- `internal/rtx/parsers/ipv6_filter.go` - IPv6 filter parsing (needs implementation or fix)
- `internal/provider/resource_ipv6_filter_dynamic.go` - IPv6 filter dynamic resource

**Root Cause Analysis:**
- Error message states "IPv6 filter dynamic config not implemented"
- Read function needs to be implemented or fixed

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each parser function handles one configuration type
- **Modular Design**: Parser functions are isolated in the parsers package
- **Dependency Management**: Parsers depend only on standard library and internal types
- **Clear Interfaces**: Parser functions return structured types, not raw strings

### Performance
- Parser functions must complete within 100ms for typical configuration sizes
- No external network calls during parsing

### Security
- No sensitive data (passwords) should be logged during parsing
- Input validation must prevent regex denial-of-service attacks

### Reliability
- Parsers must not panic on malformed input
- Unknown configuration lines should be skipped with optional logging
- All parsers must have comprehensive unit tests

### Usability
- Error messages must indicate which configuration line failed to parse
- Debug logging should show the raw input for troubleshooting
