# RTX Spec vs コマンドリファレンス対照表

このドキュメントは、`specs/` ディレクトリ内の各リソース定義と `docs/RTX-commands/` のコマンドリファレンスを比較したものです。

**最終更新日**: 2026-02-07

## 概要

| ステータス | 件数 | 説明 |
|-----------|------|------|
| ✅ 一致 | 13 | Specとドキュメントの構文が完全一致 |
| ⚠️ 差分あり | 7 | 構文やオプションに差異がある |

---

## 対照表

### 1. admin (specs/admin/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_admin_config` |
| **説明** | 管理者・ユーザー設定 |
| **参照** | 04_機器の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
login password <password>
administrator password <password>
login user <username> <password>
login user <username> encrypted <password>
user attribute <user> <attribute>=<value> ...
```

**ドキュメント内のコマンド構文**:
```
login password <password>
administrator password <password>
login user <username> <password>
user attribute [user] attribute=value [attribute=value ...]
```

**差分**: なし

---

### 2. bridge (specs/bridge/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_bridge` |
| **説明** | ブリッジ設定 |
| **参照** | 45_ブリッジインタフェース.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
bridge member <bridge_interface> <interface> <interface> [...]
```

**ドキュメント内のコマンド構文**:
```
bridge member bridge_interface interface interface [...]
no bridge member bridge_interface [interface ...]
```

**差分**: なし

---

### 3. ddns (specs/ddns/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_ddns` |
| **説明** | DDNS設定 |
| **参照** | 34_ネットボランチ_DNS_サービスの設定.md |
| **ステータス** | ⚠️ 差分あり |

**Spec内のコマンド構文**:
```
netvolante-dns hostname host <interface> <hostname>
netvolante-dns server <server_num>
netvolante-dns timeout <timeout>
netvolante-dns use ipv6 <interface> <on|off>
netvolante-dns auto hostname <interface> <on|off>
netvolante-dns use <interface> <on|off>
netvolante-dns go <interface>
```

**ドキュメント内のコマンド構文**:
```
netvolante-dns hostname host interface [server=server_num] host [duplicate]
netvolante-dns hostname host pp [server=server_num] host [duplicate]
netvolante-dns server ip_address  または  netvolante-dns server name
netvolante-dns timeout interface time
netvolante-dns use interface switch
netvolante-dns auto hostname interface [server=server_num] switch
netvolante-dns go interface
```

**差分**:
- `netvolante-dns server`: Specは数値、ドキュメントはIPアドレス形式
- ドキュメントには `[server=server_num]` オプションが存在
- ドキュメントには `[duplicate]` オプションが存在

---

### 4. dns (specs/dns/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_dns` |
| **説明** | DNS設定 |
| **参照** | 24_DNS_の設定.md |
| **ステータス** | ⚠️ 差分あり |

**Spec内のコマンド構文**:
```
dns domain lookup <on|off>
dns domain <domain_name>
dns server <ip_address> [ip_address ...]
dns server select <id> <servers> <record_type> <query_pattern> [restrict pp <connection_pp>]
dns static <name> <address>
dns service <recursive|off>
dns private address spoof <on|off>
```

**ドキュメント内のコマンド構文**:
```
dns service service  (service: on, off, recursive, fallback)
dns server ip_address [edns=sw] [nat46=tunnel_num] ...
dns domain domain_name
dns server select id server [edns=sw] [nat46=tunnel_num] ... [type] query [original-sender] [restrict pp connection-pp]
dns static type name value [ttl=ttl]
dns private address spoof spoof
```

**差分**:
- `dns domain lookup <on|off>` はドキュメントに存在しない（`dns service` で制御）
- `dns static`: Specは `<name> <address>` 形式、ドキュメントは `type name value [ttl=ttl]` 形式

---

### 5. ethernet_filter (specs/ethernet_filter/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_ethernet_filter` |
| **説明** | イーサネットフィルタ設定 |
| **参照** | 09_イーサネットフィルタの設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
ethernet filter <num> <kind> <src_mac> <dst_mac> [offset byte_list]
ethernet filter <num> <kind> <type> [scope] [offset byte_list]
ethernet <interface> filter <dir> <list>
```

**ドキュメント内のコマンド構文**:
```
ethernet filter num kind src_mac [dst_mac [offset byte_list]]
ethernet filter num kind type [scope] [offset byte_list]
ethernet interface filter dir list
```

**差分**: なし

---

### 6. interface (specs/interface/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_interface` |
| **説明** | インタフェース設定 |
| **参照** | 08_IP_の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
ip <interface> address <address>
ip <interface> secure filter <in|out> <filter_list>
ip <interface> nat descriptor <descriptor_num>
ip <interface> proxyarp <on|off>
ip <interface> mtu <mtu>
description <interface> "<description>"
```

**ドキュメント内のコマンド構文**:
```
ip interface address ip_address[/netmask]
ip interface secure filter in|out filter_list [dynamic dynamic_filter_list]
ip interface nat descriptor descriptor_num
ip interface proxyarp switch
ip interface mtu mtu
description interface description
```

**差分**: なし

---

### 7. ip_filter (specs/ip_filter/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_ip_filter` |
| **説明** | IPフィルタ設定 |
| **参照** | 08_IP_の設定.md (8.1.14) |
| **ステータス** | ⚠️ 差分あり |

**Spec内のコマンド構文**:
```
ip filter <num> <action> <src_addr> <dst_addr> <protocol> [src_port] [dst_port]
ip filter dynamic <num> <src> <dst> <protocol> [syslog on] [timeout=<time>]
ip filter dynamic <num> <src> <dst> filter <filter_list> [in <filter_list>] [out <filter_list>]
ipv6 filter <num> <action> <src_addr> <dst_addr> <protocol> [src_port] [dst_port]
ipv6 filter dynamic <num> <src> <dst> <protocol>
```

**ドキュメント内のコマンド構文**:
```
ip filter num action src_address dst_address protocol [src_port] [dst_port]
ip filter dynamic dyn_filter_num srcaddr[/mask] dstaddr[/mask] protocol [option ...]
ip filter dynamic dyn_filter_num srcaddr[/mask] dstaddr[/mask] filter filter_list [in filter_list] [out filter_list] [option ...]
  option: syslog=on|off, timeout=time
ipv6 filter num action src_address dst_address protocol [src_port] [dst_port]
ipv6 filter dynamic num src dst protocol
```

**差分**:
- Spec: `syslog on` vs ドキュメント: `syslog=on` (等号の有無)
- ドキュメント: srcaddr/dstaddrに `[/mask]` オプションがあるがSpecでは`*`のみをテスト

---

### 8. ipv6 (specs/ipv6/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_ipv6_interface` |
| **説明** | IPv6設定 |
| **参照** | 30_IPv6.md |
| **ステータス** | ⚠️ 差分あり |

**Spec内のコマンド構文**:
```
ipv6 <interface> address <address>/<prefix_len>
ipv6 <interface> rtadv send <prefix_id> [options]
ipv6 <interface> dhcp service <client|server>
ipv6 <interface> mtu <mtu>
ipv6 <interface> secure filter <in|out> <filter_list>
ipv6 prefix <id> <prefix>/<prefix_len>
```

**ドキュメント内のコマンド構文**:
```
ipv6 interface address ipv6_address/prefix_len [address_type]
ipv6 interface rtadv send prefix_id [prefix_id ...] [option=value ...]
ipv6 interface dhcp service type
ipv6 interface mtu mtu
ipv6 interface secure filter in|out filter_list [dynamic dynamic_filter_list]
ipv6 prefix prefix_id prefix/prefix_len [preferred_lifetime=time] [valid_lifetime=time] [l_flag=switch] [a_flag=switch]
```

**差分**:
- `ipv6 prefix` の追加オプション（`preferred_lifetime`, `valid_lifetime`, `l_flag`, `a_flag`）がSpecに未記載

---

### 9. l2tp (specs/l2tp/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_l2tp` |
| **説明** | L2TP設定 |
| **参照** | 16_L2TP_機能の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
l2tp service <on|off> [l2tpv3] [l2tp]
tunnel encapsulation <l2tp|l2tpv3>
tunnel endpoint address <local> <remote>
l2tp local router-id <ip>
l2tp remote router-id <ip>
l2tp always-on <on|off>
l2tp keepalive use <on|off> <interval> <count>
l2tp tunnel disconnect time <time>
l2tp tunnel auth <on|off> [password]
l2tp syslog <on|off>
```

**ドキュメント内のコマンド構文**:
```
l2tp service service [version [version]]
tunnel encapsulation (14_トンネリング.md)
l2tp tunnel auth switch [password]
l2tp tunnel disconnect time time
l2tp keepalive use switch [interval [count]]
l2tp always-on sw
```

**差分**: なし（タイプミス "intarval" → "interval" は修正済み）

---

### 10. nat/masquerade (specs/nat/masquerade.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_nat_masquerade` |
| **説明** | NATマスカレード設定 |
| **参照** | 23_NAT_機能.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
nat descriptor type <num> masquerade
nat descriptor address outer <num> <address>
nat descriptor address inner <num> <address_range>
```

**ドキュメント内のコマンド構文**:
```
nat descriptor type num type
nat descriptor address outer num address
nat descriptor address inner num address
```

**差分**: なし

---

### 11. nat/static (specs/nat/static.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_nat_static` |
| **説明** | 静的NAT設定 |
| **参照** | 23_NAT_機能.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
nat descriptor type <num> static
nat descriptor static <num> <id> <outer>=<inner>
```

**ドキュメント内のコマンド構文**:
```
nat descriptor type num type
nat descriptor static num id outer=inner
```

**差分**: なし

---

### 12. ospf (specs/ospf/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_ospf` |
| **説明** | OSPF設定 |
| **参照** | 28_OSPF.md |
| **ステータス** | ⚠️ 差分あり |

**Spec内のコマンド構文**:
```
ospf use <on|off>
ospf router id <ip_address>
ospf area <area_id> [stub [no-summary]|nssa [no-summary]]
ip <interface> ospf area <area_id>
ospf import from <static|connected>
```

**ドキュメント内のコマンド構文**:
```
ospf use use
ospf router id ip_address
ospf area area [auth=auth] [stub [cost=cost]]
ip interface ospf area area
ospf import from protocol [filter filter_num ...]
```

**差分**:
- NSSA サポートがドキュメントに明記されていない
- `[auth=auth]` オプションがSpecに未記載

---

### 13. ppp (specs/ppp/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_pppoe` |
| **説明** | PPP/PPPoE設定 |
| **参照** | 11_PPP_の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
pp select <num>
pppoe use <interface>
pp bind <interface>
pp auth accept <chap|pap|chap pap>
pp auth myname <username> <password>
pp always-on <on|off>
pp disconnect time <time>
pp keepalive interval <interval> retry-interval <retry>
ip pp address <address>
ip pp mtu <mtu>
ip pp tcp mss limit <limit>
ip pp nat descriptor <num>
```

**ドキュメント内のコマンド構文**:
```
pp select peer_num
pppoe use interface
pp bind interface [/info] [interface [/info]]
pp auth accept accept [accept]
pp auth myname myname password
pp always-on switch [time]
pp keepalive interval interval retry-interval retry
```

**差分**: なし

---

### 14. pptp (specs/pptp/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_pptp` |
| **説明** | PPTP設定 |
| **参照** | 17_PPTP_機能の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
pptp service <on|off>
pptp tunnel disconnect time <time>
pptp keepalive use <on|off>
ppp ccp type <mppe-128|mppe-56|mppe-40> [require]
ip pp remote address pool <start>-<end>
```

**ドキュメント内のコマンド構文**:
```
pptp service service
pptp tunnel disconnect time time
pptp keepalive use switch
ppp ccp type type
ip pp remote address pool range
```

**差分**: なし

---

### 15. schedule (specs/schedule/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_schedule` |
| **説明** | スケジュール設定 |
| **参照** | 37_スケジュール.md |
| **ステータス** | ⚠️ 差分あり |

**Spec内のコマンド構文**:
```
schedule at <id> <time> <command>
schedule at <id> startup <command>
schedule at <id> <date> <time> <command>
schedule pp <pp_num> <day_of_week> <time> <action>
```

**ドキュメント内のコマンド構文**:
```
schedule at id [date] time * command ...
schedule at id [date] time pp peer_num command ...
schedule at id [date] time tunnel tunnel_num command ...
schedule at id +timer * command ...
```

**差分**:
- Specの `schedule pp` 構文はドキュメントの `schedule at id [date] time pp peer_num command ...` と形式が異なる

---

### 16. service (specs/service/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_service` |
| **説明** | サービス設定（HTTP/SSH/SFTP） |
| **参照** | 04_機器の設定.md, 33_HTTP_サーバー機能.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
httpd host <any|interface>
httpd proxy-access l2ms permit <on|off>
sshd service <on|off>
sshd host <interface> [interface ...]
sshd auth method <password|publickey>
sftpd host <interface> [interface ...]
```

**ドキュメント内のコマンド構文**:
```
httpd host ip_range [ip_range ...] / httpd host any / httpd host none / httpd host lan
httpd proxy-access l2ms permit permit
sshd service switch
sshd host interface [interface ...]
sshd auth method method
sftpd host interface [interface ...]
```

**差分**: なし

---

### 17. snmp (specs/snmp/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_snmp` |
| **説明** | SNMP設定 |
| **参照** | 21_SNMP_の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
snmp sysname <name>
snmp syslocation <location>
snmp syscontact <contact>
snmp community read-only <community> [acl]
snmp community read-write <community> [acl]
snmp host <ip_address> [community community] [version version]
snmp trap community <community>
snmp trap enable snmp <trap_type>
```

**ドキュメント内のコマンド構文**:
```
snmp sysname name
snmp syslocation location
snmp syscontact contact
snmp community read-only community [acl]
snmp community read-write community [acl]
snmp host ip_address [community community] [version version]
snmp trap community community
snmp trap enable snmp trap_type
```

**差分**: なし

---

### 18. syslog (specs/syslog/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_syslog` |
| **説明** | Syslog設定 |
| **参照** | 04_機器の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
syslog host <ip_address> [port]
syslog local address <ip_address>
syslog facility <facility>
syslog notice <on|off>
syslog info <on|off>
syslog debug <on|off>
```

**ドキュメント内のコマンド構文**:
```
syslog host ip_address [port]
syslog local address ip_address
syslog facility facility
syslog notice on|off
syslog info on|off
syslog debug on|off
```

**差分**: なし

---

### 19. system (specs/system/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_system_config` |
| **説明** | システム設定 |
| **参照** | 04_機器の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
timezone <timezone>
console character <encoding>
console lines <num|infinity>
console prompt <prompt>
system packet-buffer <size> max-buffer=<num> max-free=<num>
statistics traffic <on|off>
statistics nat <on|off>
```

**ドキュメント内のコマンド構文**:
```
timezone timezone
console character encoding
console lines lines
console prompt prompt
system packet-buffer size [max-buffer=num] [max-free=num]
statistics traffic on|off
statistics nat on|off
```

**差分**: なし

---

### 20. tunnel (specs/tunnel/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_tunnel` |
| **説明** | トンネル設定 |
| **参照** | 14_トンネリング.md, 15_IPsec_の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
tunnel select <id>
tunnel encapsulation <ipsec|l2tp|l2tpv3>
tunnel endpoint name <name> [fqdn]
tunnel endpoint address <local> <remote>
description "<description>"
tunnel enable <id>
tunnel disable <id>
ipsec ike local address <num> <ip>
ipsec ike remote address <num> <ip>
ipsec ike pre-shared-key <num> text <key>
ipsec ike nat-traversal <num> <on|off>
l2tp hostname <name>
l2tp always-on <on|off>
l2tp syslog <on|off>
```

**ドキュメント内のコマンド構文**:
```
tunnel select tunnel_num
tunnel encapsulation type
tunnel endpoint address local remote
tunnel endpoint name name [fqdn]
tunnel enable tunnel_num
tunnel disable tunnel_num
ipsec ike local address num ip_address
ipsec ike remote address num ip_address
ipsec ike pre-shared-key num type key
ipsec ike nat-traversal num switch
```

**差分**: なし（GREはboundary_testsで`valid: false`と明記済み）

---

### 21. vlan (specs/vlan/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_vlan` |
| **説明** | VLAN設定 |
| **参照** | 38_VLAN_の設定.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
vlan <interface>/<slot> 802.1q vid=<vlan_id>
ip <interface>/<slot> address <address>
description <interface>/<slot> "<name>"
```

**ドキュメント内のコマンド構文**:
```
vlan interface/sub_interface 802.1q vid=vid [name=name]
ip interface/sub_interface address ip_address/netmask
```

**差分**: なし（2026-02-07 Reconcile完了 - VLAN ID範囲を2-4094に修正、`[name=name]`は任意オプションのため影響なし）

---

### 22. bgp (specs/bgp/config.yaml)

| 項目 | 内容 |
|------|------|
| **リソース名** | `rtx_bgp` |
| **説明** | BGP設定 |
| **参照** | 29_BGP.md |
| **ステータス** | ✅ 一致 |

**Spec内のコマンド構文**:
```
bgp use <on|off>
bgp autonomous-system <asn>
bgp router id <ip_address>
bgp neighbor <n> <remote_as> <remote_address> [hold-time=<sec>] [local-address=<ip>] [passive=on|off]
bgp neighbor pre-shared-key <n> text <password>
bgp import filter <id> include <prefix>/<cidr>
bgp import from <static|connected>
```

**ドキュメント内のコマンド構文**:
```
bgp use use
bgp autonomous-system as (1-65535)
bgp router id ip_address
bgp neighbor neighbor_id remote_as remote_address [parameter ...]
bgp neighbor pre-shared-key neighbor_id text text_key
bgp import filter filter_num [reject] kind ip_address/mask ... [parameter ...]
```

**差分**: なし（2026-02-07 Reconcile完了）

---

## 差分サマリー

### 対応が必要な項目

| リソース | 差分タイプ | 対応優先度 | 詳細 |
|---------|-----------|-----------|------|
| **dns** | 存在しないコマンド | 中 | `dns domain lookup` がドキュメントに存在しない |
| **schedule** | 構文形式 | 中 | `schedule pp` の形式がドキュメントと異なる |
| **ddns** | 引数形式・オプション未記載 | 低 | `netvolante-dns server` の引数形式、追加オプション |
| **ipv6** | オプション未記載 | 低 | `ipv6 prefix` の追加オプション |
| **ospf** | オプション未記載 | 低 | NSSA、`[auth=auth]` オプション |

---

## 更新履歴

- 2026-02-07: BGP Reconcile完了 - neighbor構文、pre-shared-key、2-byte ASN、hold-time範囲を修正
- 2026-02-07: ip_filterの差分を追加 (`syslog on` vs `syslog=on`)
- 2026-02-07: L2TP/IPIPドキュメントのタイプミス修正 ("intarval" → "interval")
- 2026-02-07: 初版作成 - 全22リソースの対照表を作成
