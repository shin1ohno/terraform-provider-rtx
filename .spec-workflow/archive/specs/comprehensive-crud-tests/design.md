# Design Document: Comprehensive CRUD Tests

## Overview

This design defines the test architecture for adding comprehensive CRUD tests to 15 Provider resources and 15 Client services. Tests follow established patterns using MockExecutor for Client tests and schema.TestResourceDataRaw for Provider tests.

## Steering Document Alignment

### Technical Standards (tech.md)
- Go 1.21+ with testify/mock for mocking
- Terraform Plugin SDK v2 testing utilities
- Table-driven tests for comprehensive coverage
- Real RTX command formats in test fixtures

### Project Structure (structure.md)
- Client tests: `internal/client/*_service_test.go`
- Provider tests: `internal/provider/resource_rtx_*_test.go`
- Test helpers reused from existing patterns

## Code Reuse Analysis

### Existing Components to Leverage

- **MockExecutor**: Mock SSH executor for Client tests (`internal/client/executor.go`)
- **schema.TestResourceDataRaw**: Terraform SDK testing utility for Provider tests
- **testify/assert**: Assertion library for readable test code
- **Existing test fixtures**: Real RTX command/response formats from parser tests

### Integration Points

- **Parser layer**: Reuse RTX command formats from `internal/rtx/parsers/*_test.go`
- **Interfaces**: Use `internal/client/interfaces.go` type definitions

## Architecture

### Test Patterns

```
┌─────────────────────────────────────────────────────────────┐
│                     Test Architecture                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Provider Tests                    Client Tests             │
│  ┌─────────────────────┐          ┌─────────────────────┐  │
│  │ TestResourceDataRaw │          │    MockExecutor     │  │
│  │         ↓           │          │         ↓           │  │
│  │ Build*FromResource  │          │   Service.Get()     │  │
│  │         ↓           │          │   Service.Create()  │  │
│  │   Assert fields     │          │   Service.Update()  │  │
│  └─────────────────────┘          │   Service.Delete()  │  │
│                                    │         ↓           │  │
│                                    │  Assert commands    │  │
│                                    │  Assert parsing     │  │
│                                    └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Client Test Pattern

- **Purpose:** Test SSH command generation and response parsing
- **Interfaces:** Table-driven tests with `mockSetup`, `expected`, `expectedErr`
- **Dependencies:** MockExecutor, testify/mock, testify/assert
- **Reuses:** MockExecutor pattern from dns_service_test.go

#### Get (単一コマンド)
```go
func TestXxxService_Get(t *testing.T) {
    tests := []struct {
        name        string
        mockSetup   func(*MockExecutor)
        expected    *XxxConfig
        expectedErr bool
        errMessage  string
    }{...}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockExecutor := new(MockExecutor)
            tt.mockSetup(mockExecutor)

            service := &XxxService{executor: mockExecutor}
            result, err := service.Get(context.Background())

            // assertions...
            mockExecutor.AssertExpectations(t)
        })
    }
}
```

#### Create/Update/Delete (バッチ送信)
```go
func TestXxxService_Create(t *testing.T) {
    tests := []struct {
        name             string
        config           *XxxConfig
        expectedCommands []string  // バッチで送信されるコマンド列
        mockResponse     string
        expectedErr      bool
    }{
        {
            name: "create basic config",
            config: &XxxConfig{
                Enabled: true,
                Value:   "test",
            },
            expectedCommands: []string{
                "xxx use on",
                "xxx value test",
            },
            mockResponse: "# xxx use on\n# xxx value test\n>",
            expectedErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockExecutor := new(MockExecutor)
            // バッチ送信をモック
            mockExecutor.On("RunBatch", mock.Anything, tt.expectedCommands).
                Return([]byte(tt.mockResponse), nil)

            service := &XxxService{executor: mockExecutor}
            err := service.Create(context.Background(), tt.config)

            if tt.expectedErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            mockExecutor.AssertExpectations(t)
        })
    }
}
```

### Provider Test Pattern

- **Purpose:** Test Terraform schema handling and data conversion
- **Interfaces:** schema.TestResourceDataRaw with map[string]interface{}
- **Dependencies:** terraform-plugin-sdk/v2/helper/schema
- **Reuses:** Pattern from resource_rtx_dns_server_test.go

```go
func TestBuildXxxFromResourceData_BasicConfig(t *testing.T) {
    resourceData := schema.TestResourceDataRaw(t, resourceRTXXxx().Schema, map[string]interface{}{
        "field1": "value1",
        "field2": 123,
    })

    config := buildXxxFromResourceData(resourceData)

    // assertions...
}
```

## Test Coverage Matrix

### Client Layer Tests (15 services)

| Service | Get | Create | Update | Delete | Edge Cases |
|---------|-----|--------|--------|--------|------------|
| bgp_service | ✓ | ✓ | ✓ | ✓ | Empty config, partial neighbors |
| dhcp_scope_service | ✓ | ✓ | ✓ | ✓ | IP range format, expire time |
| ipsec_tunnel_service | ✓ | ✓ | ✓ | ✓ | Multiple SAs, transport mode |
| nat_masquerade_service | ✓ | ✓ | ✓ | ✓ | Protocol-only entries |
| nat_static_service | ✓ | ✓ | ✓ | ✓ | Port mappings, IP ranges |
| ospf_service | ✓ | ✓ | ✓ | ✓ | Multiple areas, networks |
| pptp_service | ✓ | ✓ | ✓ | ✓ | Server/client modes |
| system_service | ✓ | ✓ | ✓ | ✓ | Timezone, console settings |
| config_service | ✓ | - | - | - | Show config parsing |
| ethernet_filter_service | ✓ | ✓ | ✓ | ✓ | MAC filter rules |
| interface_service | ✓ | ✓ | ✓ | ✓ | Multiple interface types |
| ip_filter_service | ✓ | ✓ | ✓ | ✓ | Extended ACLs |
| ipsec_transport_service | ✓ | ✓ | ✓ | ✓ | SA policies |
| ipv6_prefix_service | ✓ | ✓ | ✓ | ✓ | Prefix delegation |
| schedule_service | ✓ | ✓ | ✓ | ✓ | Time schedules |
| syslog_service | ✓ | ✓ | ✓ | ✓ | Multiple hosts |
| vlan_service | ✓ | ✓ | ✓ | ✓ | Tagged/untagged ports |

### Provider Layer Tests (15 resources)

| Resource | Build | Read | Validation | Edge Cases |
|----------|-------|------|------------|------------|
| rtx_access_list_extended | ✓ | ✓ | ✓ | Multiple rules, any/host |
| rtx_access_list_extended_ipv6 | ✓ | ✓ | ✓ | IPv6 prefixes |
| rtx_access_list_mac | ✓ | ✓ | ✓ | MAC address formats |
| rtx_bgp | ✓ | ✓ | ✓ | Multiple neighbors, AS |
| rtx_dhcp_scope | ✓ | ✓ | ✓ | IP ranges, options |
| rtx_interface_acl | ✓ | ✓ | ✓ | In/out directions |
| rtx_interface_mac_acl | ✓ | ✓ | ✓ | MAC ACL binding |
| rtx_ipsec_tunnel | ✓ | ✓ | ✓ | SA policies, IKE |
| rtx_l2tp | ✓ | ✓ | ✓ | Tunnel parameters |
| rtx_nat_masquerade | ✓ | ✓ | ✓ | Static entries, protocols |
| rtx_nat_static | ✓ | ✓ | ✓ | Port mappings |
| rtx_ospf | ✓ | ✓ | ✓ | Areas, networks |
| rtx_pptp | ✓ | ✓ | ✓ | Server settings |
| rtx_static_route | ✓ | ✓ | ✓ | Gateway, interface |
| rtx_system | ✓ | ✓ | ✓ | Timezone, console |

## Data Models

### Test Case Structure (Client)

```go
type clientTestCase struct {
    name        string                    // Test description
    mockSetup   func(*MockExecutor)       // Mock expectations
    expected    interface{}               // Expected result
    expectedErr bool                      // Error expected?
    errMessage  string                    // Error substring to match
}
```

### Test Case Structure (Provider)

```go
type providerTestCase struct {
    name     string                      // Test description
    input    map[string]interface{}      // Schema input
    expected interface{}                 // Expected config
}
```

## Error Handling

### Error Scenarios

1. **Connection failure**
   - **Handling:** MockExecutor returns error
   - **Test Verify:** Error propagated correctly

2. **Parse error**
   - **Handling:** Invalid response format
   - **Test Verify:** Meaningful error message

3. **Validation error**
   - **Handling:** Invalid configuration values
   - **Test Verify:** Validation rejects bad input

## Testing Strategy

### Unit Testing

- **Client tests:** Mock SSH execution, verify command generation
- **Provider tests:** Mock schema data, verify conversion functions
- **Coverage target:** 80%+ line coverage per file

### Test Organization

Each test file follows this structure:
1. Test Get/Read operations (successful cases)
2. Test Create operations (successful cases)
3. Test Update operations (successful cases)
4. Test Delete operations (successful cases)
5. Test error handling cases
6. Test edge cases (empty config, partial config)

## RTX Command Specification by CRUD Operation

### 1. BGP (Border Gateway Protocol)

#### Read (Get)
```
show config | grep bgp
```
**Response parsing:**
```
bgp use on
bgp autonomous-system 65001
bgp router id 10.0.0.1
bgp neighbor 1 address 192.168.1.2 as 65002
bgp neighbor 1 hold-time 90
bgp import filter 1 include 10.0.0.0/8
bgp import from static
```

#### Create
```
bgp use on
bgp autonomous-system 65001
bgp router id 10.0.0.1
bgp neighbor 1 address 192.168.1.2 as 65002
bgp neighbor 1 hold-time 90
bgp neighbor 1 keepalive 30
bgp neighbor 1 password secret123
bgp import filter 1 include 10.0.0.0/8
bgp import from static
```

#### Update
```
# Remove old settings first
no bgp neighbor 1
no bgp import filter 1
no bgp import from static
# Apply new settings
bgp neighbor 1 address 192.168.1.3 as 65003
bgp neighbor 1 hold-time 180
bgp import filter 1 include 172.16.0.0/12
# Apply configuration to running router (required)
bgp configure refresh
```

#### Delete
```
no bgp neighbor 1
no bgp import filter 1
no bgp import from static
no bgp import from connected
bgp use off
```

---

### 2. DHCP Scope

#### Read (Get)
```
show config | grep "dhcp scope"
```
**Response parsing:**
```
dhcp scope 1 192.168.1.0/24 expire 24:00
dhcp scope option 1 dns=8.8.8.8,8.8.4.4 router=192.168.1.1 domain=example.com
dhcp scope 1 except 192.168.1.1-192.168.1.10
```
Or IP range format:
```
dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253
```

#### Create
```
dhcp scope 1 192.168.1.0/24 expire 24:00
dhcp scope option 1 dns=8.8.8.8,8.8.4.4
dhcp scope option 1 router=192.168.1.1
dhcp scope option 1 domain=example.com
dhcp scope 1 except 192.168.1.1-192.168.1.10
```

#### Update
```
# Remove old scope completely
no dhcp scope 1 except 192.168.1.1-192.168.1.10
no dhcp scope option 1
no dhcp scope 1
# Create with new settings
dhcp scope 1 192.168.1.0/24 expire 12:00
dhcp scope option 1 dns=1.1.1.1
dhcp scope option 1 router=192.168.1.254
```

#### Delete
```
no dhcp scope 1 except 192.168.1.1-192.168.1.10
no dhcp scope option 1
no dhcp scope 1
```

---

### 3. IPsec Tunnel

#### Read (Get)
```
show config | grep ipsec
show config | grep "tunnel select"
```
**Response parsing:**
```
tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 192.168.1.1
ipsec ike remote address 1 203.0.113.1
ipsec ike pre-shared-key 1 text mysecret
ipsec ike encryption 1 aes256
ipsec ike hash 1 sha256
ipsec ike group 1 modp2048
ipsec ike keepalive use 1 on dpd 30 3
ipsec sa policy 101 1 esp aes256-cbc sha256-hmac
```

#### Create
```
tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 192.168.1.1
ipsec ike remote address 1 203.0.113.1
ipsec ike pre-shared-key 1 text mysecret
ipsec ike encryption 1 aes256
ipsec ike hash 1 sha256
ipsec ike group 1 modp2048
ipsec ike keepalive use 1 on dpd 30 3
ipsec sa policy 101 1 esp aes256-cbc sha256-hmac
tunnel enable 1
```

#### Update (Atomic Update Pattern)
```
# Phase 1: SendBatch - 一括送信（saveなし）
tunnel select 1
ipsec ike remote address 1 203.0.113.2
ipsec ike encryption 1 aes128
ipsec ike keepalive use 1 on dpd 60 5
tunnel disable 1
tunnel enable 1
# ← この時点で接続が切断される（想定内）

# Phase 2: Polling - VPN再確立を待つ

# Phase 3: Reconnect & Save
save
```

#### Delete
```
tunnel select 1
tunnel disable 1
no ipsec sa policy 101
no ipsec ike keepalive use 1
no ipsec ike pre-shared-key 1
no ipsec ike remote address 1
no ipsec ike local address 1
no ipsec tunnel 1
no tunnel select 1
```

---

### 4. NAT Masquerade

#### Read (Get)
```
show config | grep "nat descriptor"
```
**Response parsing:**
```
nat descriptor type 1000 masquerade
nat descriptor address outer 1000 ipcp
nat descriptor address inner 1000 192.168.1.0-192.168.1.255
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.100:80=192.168.1.100:8080 tcp
ip pp nat descriptor 1000
```

#### Create
```
nat descriptor type 1000 masquerade
nat descriptor address outer 1000 ipcp
nat descriptor address inner 1000 192.168.1.0-192.168.1.255
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.100:80=192.168.1.100:8080 tcp
ip pp nat descriptor 1000
```

#### Update
```
# Modify static entries
no nat descriptor masquerade static 1000 2
nat descriptor masquerade static 1000 2 192.168.1.101:443=192.168.1.101:8443 tcp
# Add new entry
nat descriptor masquerade static 1000 3 192.168.1.102:22=192.168.1.102:22 tcp
```

#### Delete
```
no ip pp nat descriptor 1000
no nat descriptor masquerade static 1000 1
no nat descriptor masquerade static 1000 2
no nat descriptor address inner 1000
no nat descriptor address outer 1000
no nat descriptor type 1000
```

---

### 5. NAT Static

#### Read (Get)
```
show config | grep "nat descriptor"
```
**Response parsing:**
```
nat descriptor type 2000 static
nat descriptor static 2000 203.0.113.10=192.168.1.10
nat descriptor static 2000 203.0.113.11:80=192.168.1.11:8080 tcp
ip wan nat descriptor 2000
```

#### Create
```
nat descriptor type 2000 static
nat descriptor static 2000 203.0.113.10=192.168.1.10
nat descriptor static 2000 203.0.113.11:80=192.168.1.11:8080 tcp
ip wan nat descriptor 2000
```

#### Update
```
no nat descriptor static 2000 203.0.113.10=192.168.1.10
nat descriptor static 2000 203.0.113.10=192.168.1.20
```

#### Delete
```
no ip wan nat descriptor 2000
no nat descriptor static 2000 203.0.113.10=192.168.1.10
no nat descriptor static 2000 203.0.113.11:80=192.168.1.11:8080 tcp
no nat descriptor type 2000
```

---

### 6. OSPF

#### Read (Get)
```
show config | grep ospf
```
**Response parsing:**
```
ospf use on
ospf router id 10.0.0.1
ospf area backbone
ospf area 0.0.0.1 stub
ospf area 0.0.0.2 nssa no-summary
ip lan1 ospf area backbone
ip tunnel1 ospf area 0.0.0.1
ospf import from static
ospf import from connected
```

#### Create
```
ospf use on
ospf router id 10.0.0.1
ospf area backbone
ospf area 0.0.0.1 stub
ip lan1 ospf area backbone
ip tunnel1 ospf area 0.0.0.1
ospf import from static
```

#### Update
```
# Change area type
no ospf area 0.0.0.1 stub
ospf area 0.0.0.1 nssa
# Add new interface to area
ip lan2 ospf area backbone
# Apply configuration to running router (required)
ospf configure refresh
```

#### Delete
```
no ip lan1 ospf area
no ip tunnel1 ospf area
no ospf import from static
no ospf import from connected
no ospf area 0.0.0.1
no ospf area backbone
ospf use off
```

---

### 7. PPTP

#### Read (Get)
```
show config | grep pptp
show config | grep "pp select anonymous"
show config | grep "pp auth"
```
**Response parsing:**
```
pptp service on
pptp tunnel disconnect time 900
pptp keepalive use on
pp select anonymous
pp auth accept mschap-v2
pp auth myname user1 password123
ppp ccp type mppe-128 require
ip pp remote address pool 192.168.1.100-192.168.1.150
```

#### Create
```
pptp service on
pptp tunnel disconnect time 900
pptp keepalive use on
pp select anonymous
pp auth accept mschap-v2
pp auth myname user1 password123
ppp ccp type mppe-128 require
ip pp remote address pool 192.168.1.100-192.168.1.150
```

#### Update
```
pptp tunnel disconnect time 1800
pp select anonymous
pp auth myname newuser newpassword
ip pp remote address pool 192.168.1.200-192.168.1.250
```

#### Delete
```
pp select anonymous
no ip pp remote address pool
no ppp ccp type
no pp auth myname
no pp auth accept
pptp keepalive use off
pptp service off
```

---

### 8. L2TP

#### Read (Get)
```
show config | grep l2tp
show config | grep "tunnel select"
show config | grep "tunnel encapsulation"
```
**Response parsing (L2TPv2 LNS):**
```
l2tp service on
pp select anonymous
pp bind tunnel1
pp auth accept mschap-v2
ip pp remote address pool 192.168.1.100-192.168.1.150
```
**Response parsing (L2TPv3):**
```
tunnel select 1
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 203.0.113.1
l2tp local router-id 10.0.0.1
l2tp remote router-id 10.0.0.2
l2tp remote end-id remote-router
l2tp always-on on
l2tp keepalive use on 30 3
l2tp tunnel disconnect time 900
```

#### Create (L2TPv3)
```
tunnel select 1
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 203.0.113.1
l2tp local router-id 10.0.0.1
l2tp remote router-id 10.0.0.2
l2tp remote end-id remote-router
l2tp always-on on
l2tp keepalive use on 30 3
tunnel enable 1
```

#### Update (Atomic Update Pattern)
```
# Phase 1: SendBatch - 一括送信（saveなし）
tunnel select 1
l2tp keepalive use on 60 5
l2tp tunnel disconnect time 1800
tunnel disable 1
tunnel enable 1
# ← この時点で接続が切断される（想定内）

# Phase 2: Polling - VPN再確立を待つ

# Phase 3: Reconnect & Save
save
```

#### Delete
```
tunnel select 1
tunnel disable 1
l2tp keepalive use off
no l2tp remote end-id
no l2tp remote router-id
no l2tp local router-id
no tunnel endpoint address
no tunnel encapsulation
no tunnel select 1
```

---

### 9. System Configuration

#### Read (Get)
```
show config | grep timezone
show config | grep console
show config | grep "system packet-buffer"
show config | grep statistics
```
**Response parsing:**
```
timezone +09:00
console character ja.utf8
console lines 24
console prompt "RTX>"
system packet-buffer small max-buffer=5000 max-free=1300
statistics traffic on
statistics nat on
```

#### Create/Update
```
timezone +09:00
console character ja.utf8
console lines infinity
console prompt "Router>"
system packet-buffer small max-buffer=5000 max-free=1300
statistics traffic on
statistics nat on
```

#### Delete (Reset to defaults)
```
no timezone
no console character
no console lines
no console prompt
no system packet-buffer small
no statistics traffic
no statistics nat
```

---

### 10. Static Route

#### Read (Get)
```
show config | grep "ip route"
```
**Response parsing:**
```
ip route default gateway pp 1
ip route 10.0.0.0/8 gateway 192.168.1.1
ip route 172.16.0.0/12 gateway tunnel 1
ip route 192.168.100.0/24 gateway pp 1 weight 10
```

#### Create
```
ip route default gateway pp 1
ip route 10.0.0.0/8 gateway 192.168.1.1
ip route 172.16.0.0/12 gateway tunnel 1
```

#### Update
```
no ip route 10.0.0.0/8
ip route 10.0.0.0/8 gateway 192.168.1.254
```

#### Delete
```
no ip route default
no ip route 10.0.0.0/8
no ip route 172.16.0.0/12
```

---

### 11. Access List Extended (IPv4)

#### Read (Get)
```
show config | grep "ip access-list"
show config | grep "ip .* access-group"
```
**Response parsing:**
```
ip access-list extended DENY_SSH
 10 deny tcp any any eq 22
 20 permit ip any any
ip lan1 access-group DENY_SSH in
```

#### Create
```
ip access-list extended DENY_SSH
 10 deny tcp any any eq 22
 20 permit ip any any
ip lan1 access-group DENY_SSH in
```

#### Update
```
ip access-list extended DENY_SSH
 no 10
 10 deny tcp any any eq 2222
 15 deny tcp any any eq 23
```

#### Delete
```
no ip lan1 access-group in
no ip access-list extended DENY_SSH
```

---

### 12. Access List Extended (IPv6)

#### Read (Get)
```
show config | grep "ipv6 access-list"
show config | grep "ipv6 .* access-group"
```
**Response parsing:**
```
ipv6 access-list extended DENY_SSH_V6
 10 deny tcp any any eq 22
 20 permit ipv6 any any
ipv6 lan1 access-group DENY_SSH_V6 in
```

#### Create
```
ipv6 access-list extended DENY_SSH_V6
 10 deny tcp any any eq 22
 20 permit ipv6 any any
ipv6 lan1 access-group DENY_SSH_V6 in
```

#### Update
```
ipv6 access-list extended DENY_SSH_V6
 no 10
 10 deny tcp any any eq 2222
```

#### Delete
```
no ipv6 lan1 access-group in
no ipv6 access-list extended DENY_SSH_V6
```

---

### 13. MAC Access List

#### Read (Get)
```
show config | grep "mac access-list"
show config | grep "mac .* access-group"
```
**Response parsing:**
```
mac access-list extended ALLOW_KNOWN
 10 permit host 00:11:22:33:44:55 any
 20 deny any any
mac lan1 access-group ALLOW_KNOWN in
```

#### Create
```
mac access-list extended ALLOW_KNOWN
 10 permit host 00:11:22:33:44:55 any
 20 deny any any
mac lan1 access-group ALLOW_KNOWN in
```

#### Update
```
mac access-list extended ALLOW_KNOWN
 15 permit host 00:11:22:33:44:66 any
```

#### Delete
```
no mac lan1 access-group in
no mac access-list extended ALLOW_KNOWN
```

---

### 14. Interface ACL Binding

#### Read (Get)
```
show config | grep "access-group"
```
**Response parsing:**
```
ip lan1 access-group INBOUND_FILTER in
ip lan1 access-group OUTBOUND_FILTER out
```

#### Create
```
ip lan1 access-group INBOUND_FILTER in
ip lan1 access-group OUTBOUND_FILTER out
```

#### Update
```
no ip lan1 access-group in
ip lan1 access-group NEW_INBOUND_FILTER in
```

#### Delete
```
no ip lan1 access-group in
no ip lan1 access-group out
```

---

### 15. Interface MAC ACL Binding

#### Read (Get)
```
show config | grep "mac .* access-group"
```
**Response parsing:**
```
mac lan1 access-group MAC_FILTER in
```

#### Create
```
mac lan1 access-group MAC_FILTER in
```

#### Update
```
no mac lan1 access-group in
mac lan1 access-group NEW_MAC_FILTER in
```

#### Delete
```
no mac lan1 access-group in
no mac lan1 access-group out
```

## Command Execution Strategy (コマンド送信方式)

### 現状の問題

現在の実装は**1コマンドずつ送信・応答待ち**を行っている：

```go
// 現在の実装（例: bgp_service.go）
for _, cmd := range commands {
    output, err := s.executor.Run(ctx, cmd)  // 1コマンドごとにプロンプト待ち
    if err != nil {
        return err
    }
}
```

**問題点:**
- RTXへのラウンドトリップが多い（コマンド数×往復）
- 実行が遅い
- VPN安全更新パターン（disable→enable）が実装できない

### 提案: バッチ送信方式

#### レベル1: リソース単位バッチ（必須）

1リソースの全コマンドをまとめて送信：

```go
// 提案: リソース単位バッチ
func (s *BGPService) Create(ctx context.Context, config *BGPConfig) error {
    commands := buildBGPCommands(config)
    return s.executor.RunBatch(ctx, commands)  // 一括送信
}
```

**Pros:**
- ラウンドトリップ削減（リソースあたり1往復）
- VPN安全更新パターンが実装可能
- 実行速度向上

**Cons:**
- エラー発生時、どのコマンドで失敗したか特定しにくい
- 部分適用の問題（途中でエラー→一部のみ適用済み）

#### レベル2: Terraform Apply単位バッチ（オプション）

`terraform apply`で変更される全リソースをまとめて送信：

```go
// Provider層で全リソースのコマンドを収集
type CommandCollector struct {
    commands []string
}

func (c *CommandCollector) Add(cmds ...string) {
    c.commands = append(c.commands, cmds...)
}

// Apply終了時に一括送信
func (c *CommandCollector) Flush(ctx context.Context, executor Executor) error {
    if len(c.commands) == 0 {
        return nil
    }
    return executor.RunBatch(ctx, c.commands)
}
```

**Pros:**
- 最速の実行速度
- ネットワーク効率最大化

**Cons:**
- Terraform Provider SDKとの統合が複雑
- リソース間の依存関係管理が必要
- エラー発生時のロールバックが非常に複雑
- どのリソースで失敗したか特定困難

### 推奨実装

**Phase 1（この仕様のスコープ）:** レベル1（リソース単位バッチ）を**全リソースに適用**

| リソース | Create | Update | Delete | 備考 |
|----------|--------|--------|--------|------|
| rtx_bgp | Batch | Batch | Batch | +`bgp configure refresh` |
| rtx_ospf | Batch | Batch | Batch | +`ospf configure refresh` |
| rtx_ipsec_tunnel | Batch | Batch+Polling+Save | Batch | VPN安全更新パターン |
| rtx_l2tp | Batch | Batch+Polling+Save | Batch | VPN安全更新パターン |
| rtx_pptp | Batch | Batch+Polling+Save | Batch | VPN安全更新パターン |
| rtx_dhcp_scope | Batch | Batch | Batch | 即時反映 |
| rtx_nat_masquerade | Batch | Batch | Batch | 即時反映 |
| rtx_nat_static | Batch | Batch | Batch | 即時反映 |
| rtx_system | Batch | Batch | Batch | 即時反映 |
| rtx_static_route | Batch | Batch | Batch | 即時反映 |
| rtx_access_list_extended | Batch | Batch | Batch | 即時反映 |
| rtx_access_list_extended_ipv6 | Batch | Batch | Batch | 即時反映 |
| rtx_access_list_mac | Batch | Batch | Batch | 即時反映 |
| rtx_interface_acl | Batch | Batch | Batch | 即時反映 |
| rtx_interface_mac_acl | Batch | Batch | Batch | 即時反映 |

**実装内容:**
- `Executor.RunBatch(ctx, []string)` メソッド追加
- **全15リソース**のServiceメソッドをバッチ送信に変更
- テストもバッチ送信を前提に設計

**Phase 2（将来の検討）:** レベル2（Apply単位バッチ）
- 要件とConsを精査してから判断

### 実装要件

```go
// Executor interface extension
type Executor interface {
    Run(ctx context.Context, cmd string) ([]byte, error)
    RunBatch(ctx context.Context, cmds []string) ([]byte, error)  // 新規追加
}

// Session interface extension
type Session interface {
    Send(cmd string) ([]byte, error)
    SendBatch(cmds []string) ([]byte, error)  // 新規追加
    Close() error
    SetAdminMode(bool)
}
```

---

## Configuration Application Commands (設定反映コマンド)

RTXルーターでは、設定変更が即座にルーターの動作に反映されるものと、追加のコマンドや再起動が必要なものがあります。以下は各リソースの設定反映方法を示します。

### Summary Table

| Resource | Application Method | Command | Notes |
|----------|-------------------|---------|-------|
| BGP | Refresh required | `bgp configure refresh` | BGPの設定を変更したら、ルーターを再起動するか、このコマンドを実行する必要がある |
| OSPF | Refresh required | `ospf configure refresh` | OSPF関係の設定を変更したら、ルーターを再起動するか、あるいはこのコマンドを実行しなくてはいけない |
| IPsec Tunnel | Atomic Update Pattern | `SendBatch` → polling → `save` | VPN安全更新パターン（詳細下記） |
| L2TP | Atomic Update Pattern | `SendBatch` → polling → `save` | VPN安全更新パターン（詳細下記） |
| PPTP | Atomic Update Pattern | `SendBatch` → polling → `save` | VPN安全更新パターン（詳細下記） |
| NAT Masquerade | Immediate (most) | N/A | 一部の設定変更（特にdescriptor type）はルーター再起動が必要 |
| NAT Static | Immediate (most) | N/A | 一部の設定変更はルーター再起動が必要 |
| DHCP Scope | Immediate | N/A | スコープ設定変更は即時反映 |
| System | Immediate | N/A | システム設定変更は即時反映 |
| Static Route | Immediate | N/A | 経路設定は既存設定を上書き可能、即時反映 |
| Access List | Immediate | N/A | ACL設定変更は即時反映 |
| Interface ACL | Immediate | N/A | インターフェースへのACL適用は即時反映 |
| MAC ACL | Immediate | N/A | MACフィルタ設定は即時反映 |

### Detailed Application Commands by Resource

#### BGP

**RTX Documentation Reference:** 29_BGP.md
> 「BGPの設定を変更したら、ルーターを再起動するか、このコマンドを実行する必要がある」

```
# After BGP configuration changes
bgp configure refresh
```

**Test Pattern for Update:**
```go
// MockExecutor should expect the refresh command after settings change
mockExecutor.On("Execute", ctx, "bgp configure refresh").Return("", nil)
```

#### OSPF

**RTX Documentation Reference:** 28_OSPF.md
> 「OSPF関係の設定を変更したら、ルーターを再起動するか、あるいはこのコマンドを実行しなくてはいけない」

```
# After OSPF configuration changes
ospf configure refresh
```

**Test Pattern for Update:**
```go
// MockExecutor should expect the refresh command after settings change
mockExecutor.On("Execute", ctx, "ospf configure refresh").Return("", nil)
```

#### IPsec Tunnel / L2TP / PPTP - VPN安全更新パターン (Atomic Update Pattern)

**RTX Documentation Reference:** 14_トンネリング.md
> 「トンネル先の設定を行う場合は、disable状態で行うのが望ましい」

**課題:** VPN経由でリモートRTXを管理している場合、単純に`tunnel disable`を実行すると接続が切断され、`tunnel enable`を実行できなくなる。

**解決策: Atomic Update Pattern（VPN安全更新パターン）**

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: SendBatch (saveなし)                                │
├─────────────────────────────────────────────────────────────┤
│ 複数コマンドを一括送信（プロンプト待ちなし）                    │
│ ["tunnel select 1",                                         │
│  "ipsec ike remote address 1 203.0.113.2",                  │
│  "ipsec ike encryption 1 aes128",                           │
│  "tunnel disable 1",                                        │
│  "tunnel enable 1"]                                         │
│                                                             │
│ ※ saveしないので、失敗時は再起動で元に戻る                     │
│ ※ 接続切断は想定内（エラーとして扱わない）                      │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: WaitForReconnection (Polling)                      │
├─────────────────────────────────────────────────────────────┤
│ VPN再確立を待つ                                              │
│ - SSH接続試行を繰り返す（例: 10秒間隔、最大5分）              │
│ - 成功 → Phase 3へ                                          │
│ - タイムアウト → エラー返却                                   │
│   「設定に問題がある可能性があります。リモート拠点で            │
│    ルーターを再起動してください（saveしていないので復旧します）」│
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Reconnect & Save                                   │
├─────────────────────────────────────────────────────────────┤
│ 新しいSSHセッションを確立                                     │
│ "save" を実行して設定を永続化                                 │
└─────────────────────────────────────────────────────────────┘
```

**利点:**
- ✅ 設定ミス時: 再起動で元に戻る（saveしてないから）
- ✅ 設定正常時: 自動的に接続再確立→save
- ✅ 物理アクセス不要（最悪でも再起動依頼のみ）

**実装要件:**
1. `SendBatch()` - 複数コマンドを一括送信する新メソッド
2. `WaitForReconnection()` - SSH接続試行をpollingする機能
3. 接続切断時のエラーハンドリング（切断を想定内として扱う）

**Test Pattern for Update:**
```go
func TestIPsecTunnelService_Update_AtomicPattern(t *testing.T) {
    tests := []struct {
        name              string
        batchCommands     []string
        reconnectSuccess  bool
        expectedSaveCalled bool
        expectedErr       bool
    }{
        {
            name: "successful update with reconnection",
            batchCommands: []string{
                "tunnel select 1",
                "ipsec ike remote address 1 203.0.113.2",
                "tunnel disable 1",
                "tunnel enable 1",
            },
            reconnectSuccess:   true,
            expectedSaveCalled: true,
            expectedErr:        false,
        },
        {
            name: "failed reconnection - user must restart router",
            batchCommands: []string{
                "tunnel select 1",
                "ipsec ike remote address 1 invalid",
                "tunnel disable 1",
                "tunnel enable 1",
            },
            reconnectSuccess:   false,
            expectedSaveCalled: false,
            expectedErr:        true, // Error message guides user to restart
        },
    }
    // ... test implementation
}
```

#### NAT (Masquerade/Static)

**RTX Documentation Reference:** 23_NAT_機能.md

NATの多くの設定変更は即時反映されますが、NAT descriptor typeの変更などの基本設定変更は再起動が必要な場合があります。

```
# Most changes are immediate
nat descriptor masquerade static 1000 3 192.168.1.103:443=192.168.1.103:8443 tcp

# But changing descriptor type may require restart
# (Provider should document this limitation)
```

**Test Pattern for Update:**
```go
// Simple static entry changes are immediate
mockExecutor.On("Execute", ctx, "nat descriptor masquerade static 1000 3 ...").Return("", nil)
```

#### DHCP Scope

**RTX Documentation Reference:** 12_DHCP_の設定.md

DHCPスコープの設定変更は即時反映されます。既存のリースには影響しませんが、新規リースには新しい設定が適用されます。

```
# Changes are immediate
dhcp scope 1 192.168.1.0/24 expire 12:00
dhcp scope option 1 dns=1.1.1.1
```

#### System Configuration

システム設定（timezone, console設定等）は即時反映されます。

```
# Changes are immediate
timezone +09:00
console lines infinity
```

#### Static Route

**RTX Documentation Reference:** 08_IP_の設定.md
> 「既に存在する経路を上書きすることができる」

静的ルートの変更は即時反映されます。

```
# Existing route can be overwritten
ip route 10.0.0.0/8 gateway 192.168.1.254
```

#### Access Lists

ACL設定の変更は即時反映されます。インターフェースに適用されているACLを変更すると、その瞬間から新しいルールが適用されます。

```
# Changes take effect immediately
ip access-list extended DENY_SSH
 no 10
 10 deny tcp any any eq 2222
```

## Command Parameter Variations

### IP Address Formats
- CIDR: `192.168.1.0/24`
- Dotted mask: `192.168.1.0/255.255.255.0`
- Host: `host 192.168.1.1`
- Any: `any`

### Interface Types
- `lan`, `lan1`, `lan2`
- `wan`, `wan1`
- `pp 1`, `pp 2`
- `tunnel 1`, `tunnel 2`
- `ipcp`

### Protocol Types
- `tcp`, `udp`, `icmp`
- `esp`, `ah`, `gre`
- `ip` (any IPv4)
- `ipv6` (any IPv6)

### Time Formats
- RTX format: `h:mm` (e.g., `24:00`)
- Seconds: integer value
- UTC offset: `+09:00`, `-05:00`
