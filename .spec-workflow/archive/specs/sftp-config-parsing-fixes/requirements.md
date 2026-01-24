# Requirements: SFTP Config Parsing Fixes

## Overview

When using SFTP-based config reading (`sftp_enabled = true`), several resources show incorrect state compared to SSH-based reading. This causes `terraform plan` to show changes that shouldn't exist.

## Problem Statement

After switching from SSH to SFTP for configuration reading, `terraform plan` shows unexpected differences for several resources. The raw config from RTX is correctly retrieved via SFTP, but the parsing logic has issues extracting certain fields.

## Raw RTX Config (Masked)

The following is the actual config retrieved via SFTP from RTX router, with sensitive information replaced with test data:

```
#
# Admin
#
login password TEST_LOGIN_PASS
administrator password TEST_ADMIN_PASS
login user testuser TEST_USER_PASS
user attribute administrator=off connection=off gui-page=dashboard,lan-map,config login-timer=300
user attribute testuser connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600
timezone +09:00
console character ja.utf8
console lines infinity
console prompt "[RTX1210] "
system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=10000 max-free=4950
system packet-buffer large max-buffer=20000 max-free=5600
switch control use bridge1 on terminal=on
lan-map snapshot use bridge1 on terminal=wired-only

httpd host any
httpd proxy-access l2ms permit on

operation http revision-up permit on
description yno RTX1210-Test
sshd service on
sshd host lan2 bridge1
sshd host key generate ... [SSH_KEY_OMITTED]
sftpd host bridge1
external-memory statistics filename prefix usb1:rtx1210
statistics traffic on
statistics nat on
switch control watch interval 2 5
lan-map terminal watch interval 1800 10
lan-map sysname RTX1210_Test

syslog host 192.168.1.20
syslog local address 192.168.1.253
syslog facility local0
syslog notice on
syslog info on
syslog debug on

#
# WAN connection
#

description lan2 wan
ip lan2 address dhcp
ip lan2 nat descriptor 1000
ip lan2 secure filter in  200020 200021 200022 200023 200024 200025 200103 200100 200102 200104 200101 200105 200099
ip lan2 secure filter out 200020 200021 200022 200023 200024 200025 200026 200027 200099 dynamic 200080 200081 200082 200083 200084 200085

ipv6 lan2 mtu 1500
ipv6 lan2 dhcp service client ir=on
ipv6 lan2 secure filter in 101000 101002 101099
ipv6 lan2 secure filter out 101099 dynamic 101080 101081 101082 101083 101084 101085 101098 101099

ngn type lan2 off
sip use off

#
# IP configuration
#

ip routing process fast
ip route change log on
ip filter source-route on
ip filter directed-broadcast on
ip route default gateway dhcp lan2
ip route 10.33.128.0/21 gateway 192.168.1.20 gateway 192.168.1.21
ip route 100.64.0.0/10 gateway 192.168.1.20 gateway 192.168.1.21

ipv6 routing process fast
ipv6 prefix 1 ra-prefix@lan2::/64
ipv6 bridge1 address ra-prefix@lan2::1/64
ipv6 lan1 address ra-prefix@lan2::2/64
ipv6 lan1 rtadv send 1 o_flag=on
ipv6 lan1 dhcp service server

bridge member bridge1 lan1 tunnel1
ip bridge1 address 192.168.1.253/16
ip lan1 proxyarp on

#
# Services
#

dhcp service server
dhcp server rfc2131 compliant except remain-silent
dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 maxexpire 24:00
dhcp scope bind 1 192.168.1.20 01 00 30 93 11 0e 33
dhcp scope bind 1 192.168.1.21 01 00 3e e1 c3 54 b4
dhcp scope bind 1 192.168.1.22 01 00 3e e1 c3 54 b5
dhcp scope bind 1 192.168.1.28 24:59:e5:54:5e:5a
dhcp scope bind 1 192.168.1.29 b8:0b:da:ef:77:33
dhcp scope option 1 dns=1.1.1.1,1.0.0.1
dhcp scope option 1 router=192.168.1.253

dhcp client release linkdown on
dns host bridge1
dns service recursive
dns cache use off
dns server select 1 100.100.100.100 edns=on any home.local
dns server select 500000 1.1.1.1 edns=on 1.0.0.1 edns=on a .
dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .
dns private address spoof on

#
# Tunnels
#

pp disable all
pp select anonymous
 pp bind tunnel2
 pp auth request mschap-v2
 pp auth username testuser TEST_PP_PASS
 ppp ipcp ipaddress on
 ppp ipcp msext on
 ppp ccp type none
 ip pp remote address pool dhcp
 ip pp mtu 1258
 pp enable anonymous
no tunnel enable all
tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name remote.example.com fqdn
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike keepalive log 1 off
  ipsec ike keepalive use 1 on heartbeat 10 6
  ipsec ike local address 1 192.168.1.253
  ipsec ike log 1 key-info message-info payload-info
  ipsec ike nat-traversal 1 on
  ipsec ike pre-shared-key 1 text TEST_IPSEC_PSK
  ipsec ike remote address 1 remote.example.com
  ipsec ike remote name 1 l2tpv3 key-id
 l2tp always-on on
 l2tp hostname test-RTX1210
 l2tp tunnel auth on TEST_L2TP_AUTH
 l2tp tunnel disconnect time off
 l2tp keepalive use on 60 3
 l2tp keepalive log off
 l2tp syslog on
 l2tp local router-id 192.168.1.253
 l2tp remote router-id 192.168.1.254
 l2tp remote end-id remote1
 ip tunnel secure filter in 200028 200099
 ip tunnel tcp mss limit auto
 tunnel enable 1

 tunnel select 2
 tunnel encapsulation l2tp
 ipsec tunnel 1
  ipsec sa policy 1 2 esp aes-cbc sha-hmac
  ipsec ike keepalive use 2 off
  ipsec ike nat-traversal 2 on
  ipsec ike pre-shared-key 2 text TEST_IPSEC_PSK2
  ipsec ike remote address 2 any
 l2tp tunnel disconnect time off
 ip tunnel tcp mss limit auto
 tunnel enable 2

... [IP filters omitted - not relevant to this issue]

ipsec auto refresh on
ipsec transport 1 101 udp 1701
ipsec transport 3 3 udp 1701
l2tp service on l2tpv3 l2tp

... [IPv6 filters omitted - not relevant to this issue]
```

## Identified Discrepancies

### 1. rtx_l2tp.tunnel1 - L2TPv3 tunnel fields not parsed

**Raw Config (tunnel select 1 context):**
```
tunnel select 1
 tunnel encapsulation l2tpv3
 ...
 l2tp always-on on
 l2tp hostname test-RTX1210
 l2tp keepalive use on 60 3
 ...
 tunnel enable 1
```

**Current Parse Result:**
```hcl
always_on         = false
keepalive_enabled = false
name              = null
```

**Expected Result (from main.tf):**
```hcl
always_on          = true
keepalive_enabled  = true
name               = "ebisu-RTX1210"  # from l2tp hostname
```

**Root Cause Analysis:**
- The `l2tp always-on on`, `l2tp hostname`, and `l2tp keepalive use on` commands are inside `tunnel select 1` context
- `ExtractL2TPTunnels` in `config_file.go` builds raw config from context commands and passes to `ParseL2TPConfig`
- However, the parser may not be receiving these commands correctly due to context handling

### 2. rtx_l2tp.tunnel2 - Mode incorrectly detected as l2vpn

**Raw Config (tunnel select 2 context):**
```
tunnel select 2
 tunnel encapsulation l2tp
 ipsec tunnel 1
 ...
 tunnel enable 2
```

**Current Parse Result:**
```hcl
mode = "l2vpn"
```

**Expected Result:**
```hcl
mode = "lns"
```

**Root Cause Analysis:**
- In `l2tp.go` line 203-213, when `tunnel encapsulation l2tp` is detected, only `Version` is set to `"l2tp"`, but `Mode` is not changed from the default `"l2vpn"`
- For L2TPv2 (`tunnel encapsulation l2tp`), the mode should be `"lns"` (L2TP Network Server)

### 3. rtx_l2tp_service.main - Protocols not parsed

**Raw Config:**
```
l2tp service on l2tpv3 l2tp
```

**Current Parse Result:**
```hcl
protocols = []
```

**Expected Result:**
```hcl
protocols = ["l2tpv3", "l2tp"]
```

**Root Cause Analysis:**
- `ParseL2TPServiceConfig` function is not parsing the protocol list from `l2tp service on <protocol1> <protocol2>` format
- Need to verify `l2tp_service.go` parser implementation

### 4. rtx_system.main - All system fields null

**Raw Config:**
```
timezone +09:00
console character ja.utf8
console lines infinity
console prompt "[RTX1210] "
system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=10000 max-free=4950
system packet-buffer large max-buffer=20000 max-free=5600
statistics traffic on
statistics nat on
```

**Current Parse Result:**
```hcl
timezone = null
console = null
packet_buffer = null
statistics = null
```

**Expected Result:**
```hcl
timezone = "+09:00"
console {
  character = "ja.utf8"
  lines     = "infinity"
  prompt    = "[RTX1210] "
}
packet_buffer {
  size       = "small"
  max_buffer = 5000
  max_free   = 1300
}
packet_buffer {
  size       = "middle"
  max_buffer = 10000
  max_free   = 4950
}
packet_buffer {
  size       = "large"
  max_buffer = 20000
  max_free   = 5600
}
statistics {
  traffic = true
  nat     = true
}
```

**Root Cause Analysis:**
- `ExtractSystem` in `config_file.go` filters global commands starting with `timezone `, `console `, `system packet-buffer `, `statistics `
- Need to verify these commands are being found in global context
- The system parser `ParseSystemConfig` may have issues

### 5. rtx_dhcp_binding - All bindings show as "create"

**Raw Config:**
```
dhcp scope bind 1 192.168.1.20 01 00 30 93 11 0e 33
dhcp scope bind 1 192.168.1.21 01 00 3e e1 c3 54 b4
dhcp scope bind 1 192.168.1.22 01 00 3e e1 c3 54 b5
dhcp scope bind 1 192.168.1.28 24:59:e5:54:5e:5a
dhcp scope bind 1 192.168.1.29 b8:0b:da:ef:77:33
```

**Current Parse Result:**
- All 5 bindings appear to be missing from state, causing "create" actions

**Expected Result:**
- Bindings should match existing terraform state and show no changes

**Root Cause Analysis:**
- In `config_file.go` line 920-923, `ExtractDHCPBindings` calls `parser.ParseBindings(raw, 0)` with `scopeID=0`
- The `ParseBindings` function may be using this passed `scopeID` instead of extracting it from each line
- Resource ID format may not match between SFTP read and existing state

### 6. rtx_dhcp_scope.scope1 - DNS servers order reversed

**Raw Config:**
```
dhcp scope option 1 dns=1.1.1.1,1.0.0.1
```

**Current Parse Result:**
```hcl
dns_servers = ["1.1.1.1", "1.0.0.1"]
```

**main.tf Definition:**
```hcl
dns_servers = ["1.0.0.1", "1.1.1.1"]
```

**Analysis:**
- The parser correctly extracts the order from config: `1.1.1.1` first, then `1.0.0.1`
- The `main.tf` definition has the order reversed
- This is NOT a parser bug - the `main.tf` should be corrected to match the actual config

## Requirements

### REQ-1: Fix L2TP tunnel context parsing
- **User Story:** As a Terraform user, I want L2TP tunnel attributes (always_on, keepalive_enabled, name) to be correctly parsed from tunnel context
- **Acceptance Criteria:**
  - `l2tp always-on on` in tunnel context sets `AlwaysOn = true`
  - `l2tp hostname <name>` in tunnel context sets `Name = <name>`
  - `l2tp keepalive use on <interval> <retry>` in tunnel context sets `KeepaliveEnabled = true`

### REQ-2: Fix L2TPv2 mode detection
- **User Story:** As a Terraform user, I want L2TPv2 tunnels to be correctly identified with mode="lns"
- **Acceptance Criteria:**
  - When `tunnel encapsulation l2tp` is detected, `Mode` should be set to `"lns"`
  - L2TPv3 tunnels (`tunnel encapsulation l2tpv3`) should keep `Mode = "l2vpn"`

### REQ-3: Fix L2TP service protocol parsing
- **User Story:** As a Terraform user, I want the L2TP service protocols to be parsed correctly
- **Acceptance Criteria:**
  - `l2tp service on l2tpv3 l2tp` should result in `protocols = ["l2tpv3", "l2tp"]`
  - `l2tp service on` (without protocols) should result in empty protocols list

### REQ-4: Fix system config extraction
- **User Story:** As a Terraform user, I want system configuration (timezone, console, packet_buffer, statistics) to be correctly parsed
- **Acceptance Criteria:**
  - All system commands from global context should be extracted
  - Parser should correctly parse timezone, console, packet-buffer, and statistics settings

### REQ-5: Fix DHCP binding scope ID extraction
- **User Story:** As a Terraform user, I want DHCP bindings to maintain consistent resource IDs
- **Acceptance Criteria:**
  - Scope ID should be extracted from each `dhcp scope bind` line
  - Resource IDs should match format used in terraform state

### REQ-6: Update main.tf dns_servers order
- **User Story:** As a Terraform user, I want the main.tf to reflect the actual router configuration
- **Acceptance Criteria:**
  - Update `dns_servers` order in main.tf to match actual config: `["1.1.1.1", "1.0.0.1"]`

### REQ-7: Create holistic integration tests using actual RTX config

**User Story:** As a developer, I want comprehensive tests that validate the entire parsing pipeline using realistic RTX configuration data, following the same holistic test pattern established for SSH-based parsing.

**Acceptance Criteria:**

1. WHEN a holistic test file is created THEN it SHALL be named `config_file_sftp_test.go` in `internal/rtx/parsers/`
2. WHEN the actual RTX config is used as test input THEN sensitive data SHALL be masked with test values
3. WHEN testing L2TPv3 tunnel (tunnel 1) THEN the test SHALL verify:
   - `AlwaysOn = true` (from `l2tp always-on on`)
   - `KeepaliveEnabled = true` (from `l2tp keepalive use on 60 3`)
   - `Name = "test-RTX1210"` (from `l2tp hostname test-RTX1210`)
   - `Version = "l2tpv3"` (from `tunnel encapsulation l2tpv3`)
   - `Mode = "l2vpn"` (default for L2TPv3)
4. WHEN testing L2TPv2 LNS tunnel (tunnel 2) THEN the test SHALL verify:
   - `Mode = "lns"` (implied by `tunnel encapsulation l2tp`)
   - `Version = "l2tp"` (from `tunnel encapsulation l2tp`)
5. WHEN testing L2TP service THEN the test SHALL verify:
   - `Enabled = true` (from `l2tp service on`)
   - `Protocols = ["l2tpv3", "l2tp"]` (from `l2tp service on l2tpv3 l2tp`)
6. WHEN testing system config THEN the test SHALL verify:
   - `Timezone = "+09:00"`
   - `Console.Character = "ja.utf8"`
   - `Console.Lines = "infinity"`
   - `Console.Prompt = "[RTX1210] "`
   - `PacketBuffer` has 3 entries (small, middle, large)
   - `Statistics.Traffic = true`
   - `Statistics.NAT = true`
7. WHEN testing DHCP bindings THEN the test SHALL verify:
   - All 5 bindings are extracted
   - Each binding has correct `ScopeID = 1`
   - IP addresses match: 192.168.1.20, .21, .22, .28, .29
8. IF a test fails before fixes are applied THEN it is expected (TDD Red phase)
9. WHEN fixes are applied THEN all tests SHALL pass (TDD Green phase)

**Test Coverage Matrix:**

| Resource | Extract Function | Expected Fields | Test Priority |
|----------|-----------------|-----------------|---------------|
| L2TP Tunnel 1 | `ExtractL2TPTunnels` | AlwaysOn, KeepaliveEnabled, Name | High |
| L2TP Tunnel 2 | `ExtractL2TPTunnels` | Mode="lns", Version="l2tp" | High |
| L2TP Service | `ExtractL2TPService` | Protocols=["l2tpv3","l2tp"] | High |
| System | `ExtractSystem` | Timezone, Console, PacketBuffer, Statistics | High |
| DHCP Bindings | `ExtractDHCPBindings` | 5 bindings with ScopeID=1 | High |

**Reuse Existing Test Patterns:**

The holistic test approach was established in `config_file_test.go` during SSH-based parsing development. The same patterns SHALL be reused:

```go
// Pattern 1: Full config parsing with resource extraction verification
// Reference: TestConfigFileParser_ExtractFromSampleConfig
func TestConfigFileParser_SFTPConfig(t *testing.T) {
    sampleConfig := `...actual RTX config from SFTP...`
    parser := NewConfigFileParser()
    result, err := parser.Parse(sampleConfig)
    require.NoError(t, err)

    // Verify each resource extraction matches expected values
}

// Pattern 2: Table-driven tests for specific extraction functions
// Reference: TestConfigFileParser_ExtractL2TPTunnels
func TestConfigFileParser_SFTPConfig_L2TPTunnels(t *testing.T) {
    tests := []struct {
        name     string
        tunnelID int
        expected *L2TPConfig
    }{
        {
            name:     "L2TPv3 tunnel with keepalive",
            tunnelID: 1,
            expected: &L2TPConfig{
                AlwaysOn:         true,
                KeepaliveEnabled: true,
                Name:             "test-RTX1210",
                // ...
            },
        },
        // ...
    }
}

// Pattern 3: Verify raw config is correctly captured
// Reference: TestConfigFileParser_Context handling
func TestConfigFileParser_SFTPConfig_RawCommands(t *testing.T) {
    // Verify tunnel context commands are properly collected
}
```

**TDD Cycle:**

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Write Tests (Red)                                   │
│ - Create config_file_sftp_test.go with expected values      │
│ - Tests WILL fail initially (this is expected)              │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Fix Implementation (Green)                          │
│ - Fix l2tp.go for Mode detection                            │
│ - Fix l2tp_service.go for protocol parsing                  │
│ - Fix config_file.go for context handling                   │
│ - Fix system.go for system config parsing                   │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Verify Green                                        │
│ - Run: go test ./internal/rtx/parsers/... -v -run SFTP      │
│ - All tests must pass                                       │
└─────────────────────────────────────────────────────────────┘
```

## Out of Scope

- Changes to SFTP connection handling
- Changes to SSH-based config reading
- Performance optimizations
