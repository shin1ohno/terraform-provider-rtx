# Session Progress

## Terraform Provider for Yamaha RTX

### プロジェクト概要

Yamaha RTXシリーズルーター用Terraformプロバイダーの開発プロジェクト。

**プロバイダー設定**:
- `host`: RTXルーターのIPアドレス/ホスト名
- `username`: 認証用ユーザー名
- `password`: 認証用パスワード
- `port`: SSHポート（デフォルト: 22）
- `timeout`: 接続タイムアウト秒数（デフォルト: 30）

環境変数: `RTX_HOST`, `RTX_USERNAME`, `RTX_PASSWORD`

---

## 実装完了済みリソース

| リソース | ステータス | 説明 |
|---------|----------|------|
| rtx_dhcp_scope | ✅ 完了 | DHCPスコープ管理 |
| rtx_dhcp_binding | ✅ 完了 | DHCP静的バインディング（Client Identifier対応） |
| rtx_system | ✅ 完了 | システム設定（timezone, console, packet_buffer, statistics） |
| rtx_ipv6_prefix | ✅ 完了 | IPv6プレフィックス（static, ra, dhcpv6-pd） |
| rtx_vlan | ✅ 完了 | VLAN設定（802.1Q、IP付き対応） |
| rtx_static_route | ✅ 完了 | スタティックルート（マルチホップ、ECMP、フェイルオーバー） |
| rtx_interface | ✅ 完了 | インターフェース設定（IP, filter, NAT, ProxyARP, MTU） |
| rtx_nat_static | ✅ 完了 | 静的NAT（1:1マッピング、ポートベースNAT） |
| rtx_nat_masquerade | ✅ 完了 | NATマスカレード（PAT、静的ポートマッピング） |
| rtx_kron_policy | ✅ 完了 | Kronポリシー（コマンドリスト） |
| rtx_kron_schedule | ✅ 完了 | Kronスケジュール（時刻・曜日・日付指定、起動時） |
| rtx_snmp_server | ✅ 完了 | SNMP設定（シングルトン、community、host、trap） |
| rtx_dns_server | ✅ 完了 | DNSサーバー設定（シングルトン、name_servers、server_select、hosts） |
| rtx_syslog | ✅ 完了 | Syslog設定（シングルトン、hosts、facility、log levels） |
| rtx_class_map | ✅ 完了 | QoSクラスマップ（トラフィック分類） |
| rtx_policy_map | ✅ 完了 | QoSポリシーマップ（クラスアクション定義） |
| rtx_service_policy | ✅ 完了 | QoSサービスポリシー（インターフェースへの適用） |
| rtx_shape | ✅ 完了 | トラフィックシェーピング（帯域制御） |
| rtx_admin | ✅ 完了 | 管理者パスワード設定（シングルトン） |
| rtx_admin_user | ✅ 完了 | ユーザーアカウント管理（属性、権限） |
| rtx_httpd | ✅ 完了 | HTTPDサービス設定（Webインターフェース） |
| rtx_sshd | ✅ 完了 | SSHDサービス設定（SSHアクセス） |
| rtx_sftpd | ✅ 完了 | SFTPDサービス設定（SFTPファイル転送） |
| rtx_ipv6_interface | ✅ 完了 | IPv6インターフェース設定（アドレス、RTADV、DHCPv6、MTU、フィルタ） |
| rtx_access_list_extended | ✅ 完了 | IPv4アクセスリスト（Cisco互換、entries配列構造） |
| rtx_access_list_extended_ipv6 | ✅ 完了 | IPv6アクセスリスト（Cisco互換、entries配列構造） |
| rtx_ip_filter_dynamic | ✅ 完了 | IPv4動的フィルタ（ステートフル検査） |
| rtx_ipv6_filter_dynamic | ✅ 完了 | IPv6動的フィルタ（ステートフル検査） |
| rtx_interface_acl | ✅ 完了 | インターフェースへのACL適用（IPv4/IPv6、静的/動的） |
| rtx_access_list_mac | ✅ 完了 | MACアクセスリスト（Cisco互換、entries配列構造） |
| rtx_interface_mac_acl | ✅ 完了 | インターフェースへのMAC ACL適用 |

## データソース

| データソース | ステータス |
|------------|----------|
| rtx_system_info | ✅ 完了 |
| rtx_interfaces | ✅ 完了 |
| rtx_routes | ✅ 完了 |

---

## タスク定義済みSpec（23件）

すべてのSpecに4フェーズ構成のtasks.mdが作成済み：

**基盤リソース**: rtx-interface✅, rtx-static-route✅, rtx-vlan✅, rtx-bridge, rtx-system✅

**ルーティング**: rtx-bgp, rtx-ospf

**NAT**: rtx-nat-static, rtx-nat-masquerade (client layer ✅)

**フィルタ・セキュリティ**: rtx-ip-filter, rtx-ethernet-filter

**VPN**: rtx-ipsec-tunnel, rtx-l2tp, rtx-pptp

**サービス・監視**: rtx-dns-server, rtx-snmp, rtx-qos, rtx-schedule, rtx-syslog

**システム管理**: rtx-service✅, rtx-admin✅, rtx-ipv6-interface✅, rtx-ipv6-prefix✅

---

## Wave並列開発計画

### Wave 1: 基盤リソース ✅ 完了
- rtx-interface ✅
- rtx-static-route ✅
- rtx-vlan ✅
- rtx-system ✅
- rtx-ipv6-prefix ✅

### Wave 2: フィルタ/NAT ✅ 完了
- rtx-ip-filter ✅
- rtx-ethernet-filter ✅
- rtx-nat-static ✅
- rtx-nat-masquerade ✅

### Wave 3: VPN/ルーティング ✅ 完了
- rtx-bgp ✅ (BGP動的ルーティング)
- rtx-ospf ✅ (OSPF動的ルーティング)
- rtx-ipsec-tunnel ✅ (IPsec VPN)
- rtx-l2tp ✅ (L2TP/L2TPv3トンネル)
- rtx-pptp ✅ (PPTP VPNサーバー)

### Wave 4: サービス・監視 ✅ 完了
- rtx-dns-server ✅ (DNSサーバー設定)
- rtx-snmp ✅ (SNMP監視設定)
- rtx-schedule ✅ (スケジュール実行)
- rtx-syslog ✅ (Syslog設定)
- rtx-qos ✅ (QoS/帯域制御)

### Wave 5: 管理・サービス ✅ 完了
- rtx-admin ✅ (管理者パスワード、ユーザーアカウント)
- rtx-service ✅ (HTTPD/SSHD/SFTPD サービス設定)

### Wave 6: 依存リソース ✅ 完了
- rtx-bridge ✅ (Ethernetブリッジ、L2VPN)
- rtx-ipv6-interface ✅ (IPv6アドレス、RTADV、DHCPv6、MTU、フィルタ)

---

## 現在の課題

### モッククライアントの更新が必要
既存のテストファイルで使用しているモッククライアントが新しいインターフェースメソッドを実装していない：
- `data_source_rtx_interfaces_test.go`
- `data_source_rtx_routes_test.go`
- `data_source_rtx_system_info_test.go`

必要なメソッド追加：
- System系: GetSystemConfig, ConfigureSystem, UpdateSystemConfig, ResetSystem
- IPv6Prefix系: GetIPv6Prefix, CreateIPv6Prefix, UpdateIPv6Prefix, DeleteIPv6Prefix, ListIPv6Prefixes
- VLAN系: GetVLAN, CreateVLAN, UpdateVLAN, DeleteVLAN, ListVLANs
- StaticRoute系: GetStaticRoute, CreateStaticRoute, UpdateStaticRoute, DeleteStaticRoute, ListStaticRoutes
- Interface系: GetInterfaceConfig, ConfigureInterface, UpdateInterfaceConfig, ResetInterfaceConfig

---

## 次のステップ

1. **モッククライアント修正**: 全テスト成功を確認
2. **Wave 2-5の実装開始**: 並列開発可能
3. **受け入れテスト**: Docker RTXシミュレーター or 実RTXでの統合テスト
4. **Dashboard**: http://localhost:5000 でステータス確認可能

---

## BUG: Spec/Design vs 実装の乖離 ✅ 解決済み

### 発見日: 2026-01-19
### 解決日: 2026-01-19

### 概要
Wave 2/3/4のSpec/Design文書と実際の実装に乖離が発見された。

### 解決内容

#### Phase 1: IP Filter拡張 ✅ 完了

| タスク | ステータス | 説明 |
|--------|----------|------|
| 1.1 rtx_access_list_extended | ✅ 完了 | Cisco互換スキーマでIPv4アクセスリスト実装（entries配列） |
| 1.2 rtx_access_list_extended_ipv6 | ✅ 完了 | IPv6アクセスリスト実装 |
| 1.3 rtx_ip_filter_dynamic | ✅ 完了 | IPv4動的フィルタ実装 |
| 1.4 rtx_ipv6_filter_dynamic | ✅ 完了 | IPv6動的フィルタ実装 |
| 1.5 rtx_interface_acl | ✅ 完了 | インターフェースACL適用リソース |

#### Phase 2: Ethernet Filter拡張 ✅ 完了

| タスク | ステータス | 説明 |
|--------|----------|------|
| 2.1 rtx_access_list_mac | ✅ 完了 | Cisco互換スキーマでMACアクセスリスト実装 |
| 2.2 rtx_interface_mac_acl | ✅ 完了 | インターフェースMAC ACL適用リソース |

#### Phase 3: 既存リソースの位置づけ

**決定**: Option A（削除）
- `rtx_ip_filter`: 削除済み → `rtx_access_list_extended` に置き換え
- `rtx_ethernet_filter`: 削除済み → `rtx_access_list_mac` に置き換え

### 作成/修正されたファイル

**プロバイダーリソース（新規）:**
- `internal/provider/resource_rtx_interface_acl.go`
- `internal/provider/resource_rtx_access_list_mac.go`
- `internal/provider/resource_rtx_interface_mac_acl.go`

**プロバイダーリソース（修正）:**
- `internal/provider/resource_rtx_access_list_extended.go` - Delete関数シグネチャ修正
- `internal/provider/resource_rtx_access_list_extended_ipv6.go` - Delete関数シグネチャ修正
- `internal/provider/resource_rtx_ip_filter_dynamic.go` - Delete関数シグネチャ修正
- `internal/provider/resource_rtx_ipv6_filter_dynamic.go` - Delete関数シグネチャ修正

**provider.go更新:**
- 7リソースをResourcesMapに追加登録

**テストファイル（モッククライアント更新）:**
- `data_source_rtx_interfaces_test.go`
- `data_source_rtx_routes_test.go`
- `data_source_rtx_system_info_test.go`

### ビルド・テスト結果

- ビルド: ✅ 成功 (`go build ./...`)
- プロバイダーテスト: ✅ 成功
- パーサーテスト: ✅ 成功
- クライアントテスト: ⚠️ 既存の問題あり（ethernet_filter_service_test.go, ip_filter_service_test.go）

---

## 最近のセッション履歴

### セッション26: Wave 6 並列実装完了
Wave 6の2リソース（rtx-bridge、rtx-ipv6-interface）を2並列エージェントで開発:

**rtx_bridge** (Ethernetブリッジ):
- Parser: `internal/rtx/parsers/bridge.go` - BridgeConfig（Name, Members）
  - コマンド: bridge member, no bridge member, show config | grep bridge
  - バリデーション: bridge名フォーマット（bridge[0-9]+）、メンバーインターフェース名
- Client: `internal/client/bridge_service.go` - BridgeService
  - Create/Get/Update/Delete/Listメソッド
- Provider: `internal/provider/resource_rtx_bridge.go`
  - スキーマ: name (ForceNew), members[]
  - CRUD + Import機能
- Examples: `examples/bridge/main.tf`
  - 基本ブリッジ、複数メンバー、L2VPN、VLANブリッジ、PPブリッジ

**rtx_ipv6_interface** (IPv6インターフェース設定):
- Parser: `internal/rtx/parsers/ipv6_interface.go` - IPv6InterfaceConfig, IPv6Address, RTADVConfig
  - コマンド: ipv6 address, rtadv send, dhcp service, mtu, secure filter in/out
  - バリデーション: インターフェース名、MTU（1280-65535）、フィルター番号
- Client: `internal/client/ipv6_interface_service.go` - IPv6InterfaceService
  - Configure/Get/Update/Reset/Listメソッド
- Provider: `internal/provider/resource_rtx_ipv6_interface.go`
  - スキーマ: interface, address[], rtadv{}, dhcpv6_service, mtu, secure_filter_in[], secure_filter_out[], dynamic_filter_out[]
  - CRUD + Import機能
- Examples: `examples/ipv6_interface/main.tf`
  - WAN DHCPv6クライアント、LAN RTADV+DHCPv6サーバー、Bridge静的アドレス、セキュリティフィルタ

ビルド結果: ✅ 成功
テスト結果: ✅ 新規追加分（bridge, ipv6_interface）パス
※ 既存テスト（ethernet_filter_service_test.go, ip_filter_service_test.go）に別問題あり

### セッション25: rtx_ipv6_interface 実装 (セッション26にマージ)

### セッション24: Wave 5 並列実装完了
Wave 5の2リソースを2並列エージェントで開発:

**rtx-admin** (管理者認証設定):
- Parser: `internal/rtx/parsers/admin.go` - ParseAdminConfig, BuildUserCommand等
- Client: `internal/client/admin_service.go` - AdminService実装
- Provider: `resource_rtx_admin.go` (シングルトン、パスワード)
- Provider: `resource_rtx_admin_user.go` (ユーザー管理)
- Examples: `examples/admin/main.tf`
- 機能: login_password, admin_password, ユーザーアカウントCRUD
- 属性: administrator, connection[], gui_pages[], login_timer

**rtx-service** (ネットワークサービス設定):
- Parser: `internal/rtx/parsers/service.go` - HTTPD/SSHD/SFTPD パーサー
- Client: `internal/client/service_manager.go` - ServiceManager実装
- Provider: `resource_rtx_httpd.go` (host, proxy_access)
- Provider: `resource_rtx_sshd.go` (enabled, hosts, host_key)
- Provider: `resource_rtx_sftpd.go` (hosts)
- Examples: `examples/services/{httpd,sshd,sftpd}/main.tf`
- 機能: Webインターフェース、SSH、SFTPサービス管理

ビルド結果: ✅ 成功
テスト結果: ✅ パーサー/クライアント/プロバイダーテスト全件成功

### セッション23: Wave 3 プロバイダーレイヤー完了
Wave 3のプロバイダーリソースとサンプル設定を完了:

**プロバイダーリソース（Phase 3）**:
- `resource_rtx_bgp.go` - BGP動的ルーティング
- `resource_rtx_ospf.go` - OSPF動的ルーティング
- `resource_rtx_ipsec_tunnel.go` - IPsec VPNトンネル
- `resource_rtx_l2tp.go` - L2TP/L2TPv3トンネル
- `resource_rtx_pptp.go` - PPTP VPNサーバー

**サンプル設定（Phase 4）**:
- `examples/bgp/` - iBGP/eBGP設定例
- `examples/ospf/` - OSPF multi-area設定例
- `examples/ipsec_tunnel/` - Site-to-Site VPN設定例
- `examples/l2tp/` - L2TPv2 LNS/L2TPv3設定例
- `examples/pptp/` - PPTP VPN設定例

ビルド・テスト結果: ✅ パーサーテスト全件成功

### セッション22: Wave 4 並列実装完了
Wave 4の5リソースを5並列エージェントで開発:
- **rtx_dns_server**: DNSサーバー設定（シングルトンリソース）
  - スキーマ: domain_lookup, domain_name, name_servers, server_select, hosts, private_address_spoof
  - ドメインベースサーバー選択、静的ホストエントリ対応
- **rtx_snmp_server**: SNMP監視設定（シングルトンリソース）
  - スキーマ: location, contact, communities, hosts, enable_traps
  - SNMPv1/v2c対応、トラップ設定
- **rtx_syslog**: Syslog設定（シングルトンリソース）
  - スキーマ: hosts (address/port), local_address, facility, notice/info/debug
  - 複数ホスト、カスタムポート対応
- **rtx_kron_policy/rtx_kron_schedule**: スケジュール実行（2リソース構成）
  - ポリシー: コマンドリスト定義
  - スケジュール: 時刻、曜日、日付、起動時トリガー
- **rtx_class_map/rtx_policy_map/rtx_service_policy/rtx_shape**: QoS設定（4リソース構成）
  - クラスマップ: トラフィック分類ルール
  - ポリシーマップ: クラスアクション定義（優先度、帯域）
  - サービスポリシー: インターフェースへの適用
  - シェーピング: トラフィック帯域制御

ファイル作成:
- Parser: dns.go, snmp.go, schedule.go, syslog.go, qos.go + tests
- Client: dns_service.go, snmp_service.go, schedule_service.go, syslog_service.go, qos_service.go + tests
- Provider: resource_rtx_{dns_server,snmp_server,kron_policy,kron_schedule,syslog,class_map,policy_map,service_policy,shape}.go + tests
- Examples: dns_server/, snmp/, schedule/, syslog/, qos/

ビルド結果: ✅ 成功

### セッション21: Wave 2 並列実装完了
Wave 2の4リソースを並列開発:
- **rtx_ip_filter**: IPフィルタ（ACL）リソース
  - スキーマ: number, action, source_address, dest_address, protocol, ports, established
  - CRUD操作、インポート機能
- **rtx_ethernet_filter**: Ethernetフィルタリソース
  - スキーマ: number, action, source_mac, dest_mac, ether_type, vlan_id
  - MACアドレスバリデーション
- **rtx_nat_static**: 静的NATリソース
  - スキーマ: descriptor_id, entry (inside_local, outside_global, ports, protocol)
  - 1:1マッピングとポートベースNAT対応
- **rtx_nat_masquerade**: NATマスカレードリソース
  - スキーマ: descriptor_id, outer_address, inner_network, static_entry
  - PAT、静的ポートマッピング対応

追加修正:
- Wave 3サービス（BGP, OSPF, IPsec, L2TP, PPTP）のコンパイルエラー修正
- モッククライアントの全インターフェースメソッド実装
- ビルド成功確認

### セッション20: rtx-schedule 実装
- Parser: Schedule, KronPolicy データモデル
- Commands: schedule at, schedule at startup, schedule at datetime, schedule pp
- Client: ScheduleService
- Provider: rtx_kron_policy, rtx_kron_schedule リソース
- 機能:
  - 日次定期スケジュール (at_time)
  - 週次スケジュール (day_of_week)
  - スタートアップスケジュール (on_startup)
  - 一回限りの日時指定スケジュール (date)
  - コマンドリスト (KronPolicy)
  - スケジュールとポリシーの連携

### セッション19: rtx-static-route 実装
- Parser: StaticRoute, NextHopデータモデル、マルチホップ対応
- Client: StaticRouteService、ECMP/フェイルオーバー対応
- Provider: rtx_static_route リソース
- 機能: デフォルトルート、ロードバランシング、フェイルオーバー、IPフィルタ付きルート

### セッション18: rtx-vlan 実装
- Parser: VLAN データモデル、スロット自動割り当て
- Client: VLANService
- Provider: rtx_vlan リソース
- 機能: 802.1Qタギング、IP付きVLAN、同一インターフェース上の複数VLAN

### セッション17: rtx-ipv6-prefix 実装
- Parser: IPv6Prefix データモデル（static, ra, dhcpv6-pd）
- Client: IPv6PrefixService
- Provider: rtx_ipv6_prefix リソース
- 機能: 静的プレフィックス、RA派生、DHCPv6-PD

### セッション16: rtx-system 実装
- Parser: SystemConfig（Timezone, Console, PacketBuffers, Statistics）
- Client: SystemService
- Provider: rtx_system リソース
- 機能: タイムゾーン、コンソール設定、パケットバッファチューニング、統計収集

### rtx_interface 実装
- Parser: InterfaceConfig（IP, フィルタ、NAT、ProxyARP、MTU）
- Client: InterfaceService
- Provider: rtx_interface リソース
- 機能: DHCP/静的IP、セキュリティフィルタ、動的フィルタ、NAT記述子、ProxyARP、MTU

### Wave 3 サービスファイルのコンパイルエラー修正
修正対象ファイル:
- `internal/client/bgp_service.go`
- `internal/client/ospf_service.go`
- `internal/client/ipsec_tunnel_service.go`
- `internal/client/l2tp_service.go`
- `internal/client/pptp_service.go`

修正内容:
1. `s.executor.Execute` を `s.executor.Run` に置換（Executorインターフェースのメソッド名修正）
2. 戻り値の型を `string` から `[]byte` に合わせて修正（`containsError(output)` を `containsError(string(output))` に変更）
3. `bgp_service.go` から重複した `containsError` 関数を削除（`dhcp_service.go` に定義済み）
4. `bgp_service.go` から不要な `"strings"` インポートを削除
5. `config_service.go` から重複した `DNSHost` 型定義を削除（`interfaces.go` に定義済み）
6. `client.go` に DNS メソッドを追加（GetDNS, ConfigureDNS, UpdateDNS, ResetDNS）→ 既存のため削除

ビルド結果: ✅ 成功（`go build ./...`）
