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

| リソース | 説明 |
|---------|------|
| rtx_dhcp_scope | DHCPスコープ管理 |
| rtx_dhcp_binding | DHCP静的バインディング（Client Identifier対応） |
| rtx_system | システム設定（timezone, console, packet_buffer, statistics） |
| rtx_ipv6_prefix | IPv6プレフィックス（static, ra, dhcpv6-pd） |
| rtx_vlan | VLAN設定（802.1Q、IP付き対応） |
| rtx_static_route | スタティックルート（マルチホップ、ECMP、フェイルオーバー） |
| rtx_interface | インターフェース設定（IP, filter, NAT, ProxyARP, MTU） |
| rtx_ipv6_interface | IPv6インターフェース設定（アドレス、RTADV、DHCPv6、MTU、フィルタ） |
| rtx_nat_static | 静的NAT（1:1マッピング、ポートベースNAT） |
| rtx_nat_masquerade | NATマスカレード（PAT、静的ポートマッピング） |
| rtx_kron_policy | Kronポリシー（コマンドリスト） |
| rtx_kron_schedule | Kronスケジュール（時刻・曜日・日付指定、起動時） |
| rtx_snmp_server | SNMP設定（community、host、trap） |
| rtx_dns_server | DNSサーバー設定（name_servers、server_select、hosts） |
| rtx_syslog | Syslog設定（hosts、facility、log levels） |
| rtx_class_map | QoSクラスマップ（トラフィック分類） |
| rtx_policy_map | QoSポリシーマップ（クラスアクション定義） |
| rtx_service_policy | QoSサービスポリシー（インターフェースへの適用） |
| rtx_shape | トラフィックシェーピング（帯域制御） |
| rtx_admin | 管理者パスワード設定（シングルトン） |
| rtx_admin_user | ユーザーアカウント管理（属性、権限） |
| rtx_httpd | HTTPDサービス設定（Webインターフェース） |
| rtx_sshd | SSHDサービス設定（SSHアクセス） |
| rtx_sftpd | SFTPDサービス設定（SFTPファイル転送） |
| rtx_access_list_ip | IPv4アクセスリスト（entries配列構造） |
| rtx_access_list_ipv6 | IPv6アクセスリスト（entries配列構造） |
| rtx_access_list_ip_dynamic | IPv4動的フィルタのグループ化（entries配列構造） |
| rtx_access_list_ipv6_dynamic | IPv6動的フィルタのグループ化（entries配列構造） |
| rtx_access_list_mac | MACアクセスリスト（entries配列構造） |
| rtx_bridge | Ethernetブリッジ（L2VPN） |
| rtx_bgp | BGP動的ルーティング |
| rtx_ospf | OSPF動的ルーティング |
| rtx_ipsec_tunnel | IPsec VPNトンネル |
| rtx_l2tp | L2TP/L2TPv3トンネル |
| rtx_pptp | PPTP VPNサーバー |

## データソース

| データソース | 説明 |
|------------|------|
| rtx_system_info | システム情報 |
| rtx_interfaces | インターフェース一覧 |
| rtx_routes | ルーティングテーブル |

---

## 現在の課題

### 既存のテストの問題（解決待ち）
- `TestPPPoERoundTrip`: LCPReconnect設定のパース問題
- ethernet_filter_service_test.go
- ip_filter_service_test.go

### SSH接続エラー
- `ssh: handshake failed: EOF` エラーが頻発
- 原因: SSHセッションプールが有効だが、simpleExecutorでは活用されていない
- 対策: SSH Session Pool統合Spec（`.spec-workflow/specs/ssh-session-pool-integration/`）

---

## SSH Session Pool 機能

SSHセッションプールを実装済み。プロバイダー設定で有効化可能。

```hcl
provider "rtx" {
  host     = "192.168.1.1"
  username = "admin"
  password = "password"

  ssh_session_pool {
    enabled      = true    # デフォルト: true
    max_sessions = 2       # デフォルト: 2
    idle_timeout = "5m"    # デフォルト: 5m
  }
}
```

**実装ファイル:**
- `internal/client/ssh_session_pool.go` - プール本体
- `internal/client/ssh_session_pool_test.go` - ユニットテスト（32テスト）
- `internal/client/ssh_session_pool_integration_test.go` - 統合テスト

---

## 最近の変更

### Dynamic Access List Import修正（2026-01-25）

RTXルーターは「名前付きアクセスリスト」の概念を持たず、フィルタ番号のみを管理。
Import時に他のアクセスリストのフィルタが漏れ込む問題を修正。

**解決策:**
- Import関数: 名前のみを設定、entriesは設定しない
- Read関数: stateにあるシーケンス番号のみを返す

**修正後のワークフロー:**
1. `terraform import` → 名前のみがstate保存
2. `terraform plan` → 設定のentriesが「追加」として表示
3. `terraform apply` → entriesがリソースにバインド

コミット: `13c58e0 import: prevent filter leakage between dynamic access lists`

### フィルタ属性統合（2026-01-25）

フィルタ管理を簡素化:
- 動的フィルタをアクセスリストリソースでグループ化
- `rtx_interface`から名前でアクセスリストを参照
- 冗長なACLバインディングリソースを削除

**破壊的変更:**
- 削除: `rtx_interface_acl`, `rtx_interface_mac_acl`, `rtx_ip_filter_dynamic`, `rtx_ipv6_filter_dynamic`
- 追加: `rtx_access_list_ip_dynamic`, `rtx_access_list_ipv6_dynamic`
- `rtx_interface`属性: `secure_filter_*` → `access_list_*`

### スキーマ属性名の標準化（2026-01-25）

業界標準の用語に合わせて属性名を変更:
- `filter_id` → `sequence` (評価順序)
- `id` → `priority` / `area_id` / `index` (用途に応じて)

---

## 次のステップ

1. **SSH Session Pool統合**: simpleExecutorでプールを活用（Spec作成済み）
2. **PPPパーサー修正**: LCPReconnect round-trip テスト修正
3. **受け入れテスト**: 実RTXでの統合テスト
