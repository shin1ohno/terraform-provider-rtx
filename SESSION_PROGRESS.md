# Session Progress

## rtx_tunnel 統合リソースの実装 ✅ 完了

**開始日**: 2026-02-01
**完了日**: 2026-02-01

`rtx_ipsec_tunnel` と `rtx_l2tp` を統合した新しい `rtx_tunnel` リソースを実装完了。RTXの実際のコマンド構造（`tunnel select N` が親コンテナとしてIPsecとL2TP設定を含む）を反映した設計。

### 完了済みタスク

#### Task #9: データ型の定義 ✅
- `internal/client/interfaces.go` に以下の型を追加:
  - `Tunnel` - 統合トンネル型（ID, Encapsulation, Enabled, Name, IPsec, L2TP）
  - `TunnelIPsec` - トンネル内のIPsec設定
  - `TunnelIPsecKeepalive` - IPsec keepalive/DPD設定
  - `TunnelL2TP` - トンネル内のL2TP設定
  - `TunnelL2TPKeepalive` - L2TP keepalive設定
  - `TunnelL2TPAuth` - L2TPv3トンネル認証
- Client interface に Tunnel メソッドを追加

#### Task #10: 統合パーサーの作成 ✅
- `internal/rtx/parsers/tunnel.go` を作成:
  - `TunnelParser` - 統合トンネル設定をパース
  - `ParseTunnelConfig()` - show config出力をパース
  - `BuildTunnelCommands()` - 統合コマンドビルダー
  - `ValidateTunnel()` - バリデーション関数
- `internal/rtx/parsers/tunnel_test.go` を作成:
  - IPsec, L2TPv3, L2TPv2 各モードのパーステスト
  - コマンドビルダーテスト
  - バリデーションテスト

#### Task #11: TunnelService の作成 ✅
- `internal/client/tunnel_service.go` を作成:
  - `TunnelService` - CRUDオペレーションの実装
  - `convertToParserTunnel()` / `convertFromParserTunnel()` - 型変換
- `internal/client/tunnel_service_test.go` を作成:
  - Get, Create, Delete, Update のテスト
- `internal/client/client.go` を更新:
  - `tunnelService` フィールド追加
  - Tunnel メソッドを TunnelService に委譲

#### Task #12: Tunnel リソースモデルとスキーマの作成 ✅
- `internal/provider/resources/tunnel/model.go` を作成:
  - `TunnelModel` - Terraform モデル
  - `TunnelIPsecModel`, `TunnelL2TPModel` 等ネストしたブロック
  - `ToClient()` / `FromClient()` 変換メソッド
- `internal/provider/resources/tunnel/resource.go` を作成:
  - スキーマ定義（ipsec/l2tp ネストブロック含む）
  - CRUD オペレーション
  - ImportState 実装
- `internal/provider/provider_framework.go` を更新:
  - tunnel リソースを登録
- `internal/provider/fwhelpers/helpers.go` を更新:
  - `GetStringValueWithDefault()` 関数を追加

#### Task #13: examples の更新と古いリソースの非推奨化 ✅
- `examples/tunnel/` ディレクトリを作成:
  - `main.tf` - 4つの例（IPsec, L2TPv3, L2TPv2, Security Filter付きIPsec）
  - `variables.tf` - 変数定義
- `examples/import/main.tf` を更新:
  - `rtx_ipsec_tunnel` + `rtx_l2tp` → `rtx_tunnel` に書き換え
  - ブリッジ参照を `rtx_tunnel.hnd_itm.tunnel_interface` に更新
- `examples/ipsec_tunnel/main.tf` に deprecation notice を追加
- `examples/l2tp/main.tf` に deprecation notice を追加
- `docs/resources/tunnel.md` を生成

### Encapsulation モード

| Encapsulation | IPsec Block | L2TP Block | 用途 |
|---------------|-------------|------------|------|
| `ipsec` | 必須 | 禁止 | Site-to-site IPsec VPN |
| `l2tpv3` | 任意 | 必須 | L2VPN (with optional IPsec) |
| `l2tp` | 必須 | 必須 | L2TPv2 リモートアクセス (always over IPsec) |

---

## 機能追加

### rtx_ipsec_tunnel: 3つの新機能を追加

**追加日**: 2026-01-31

`rtx_ipsec_tunnel` リソースに以下の3つの機能を追加しました。

#### 1. `secure_filter_in` / `secure_filter_out` 属性

トンネルの入出力トラフィックに対するIPフィルタを設定します。

```hcl
resource "rtx_ipsec_tunnel" "example" {
  tunnel_id = 1
  # ...
  secure_filter_in  = [200028, 200099]
  secure_filter_out = [200100, 200101, 200102]
}
```

**生成されるRTXコマンド:**
```
ip tunnel secure filter in 200028 200099
ip tunnel secure filter out 200100 200101 200102
```

#### 2. `tcp_mss_limit` 属性

トンネルのTCP MSS制限を設定します。

```hcl
resource "rtx_ipsec_tunnel" "example" {
  tunnel_id     = 1
  tcp_mss_limit = "auto"  # or numeric value like "1414"
}
```

**生成されるRTXコマンド:**
```
ip tunnel tcp mss limit auto
```

#### 3. `tunnel enable/disable` コマンドの生成

既存の `enabled` 属性に基づいて、`tunnel enable N` または `tunnel disable N` コマンドを生成するようになりました。

```hcl
resource "rtx_ipsec_tunnel" "example" {
  tunnel_id = 1
  enabled   = true  # default
}
```

**生成されるRTXコマンド:**
```
tunnel enable 1
```

#### 修正ファイル

| ファイル | 変更内容 |
|----------|----------|
| `internal/rtx/parsers/ipsec_tunnel.go` | 構造体フィールド、正規表現パターン、コマンドビルダー追加 |
| `internal/rtx/parsers/ipsec_tunnel_test.go` | パーサーとビルダーのテスト追加 |
| `internal/client/interfaces.go` | IPsecTunnel構造体にフィールド追加 |
| `internal/client/ipsec_tunnel_service.go` | コンバーター更新、Create/Updateにコマンド追加 |
| `internal/provider/resources/ipsec_tunnel/model.go` | Terraformモデルにフィールド追加、ToClient/FromClient更新 |
| `internal/provider/resources/ipsec_tunnel/resource.go` | スキーマ定義追加 |

## バグ修正

### rtx_tunnel: name 属性の不整合修正

**修正日**: 2026-02-01

`terraform apply` 時に "Provider produced inconsistent result after apply" エラーが発生する問題を修正。

**原因1**: RTX ルーターは `tunnel select N` コンテキスト内で `description` コマンドをサポートしていない。`description` コマンドを実行すると "パラメータの数が不適当です" エラーになる。

**原因2**: グローバルな `description` コマンド（ルーター全体の説明）が、インデントチェックなしでパースされ、最後のトンネルコンテキストに誤って割り当てられていた。

**修正内容**:
1. `name` 属性を読み取り専用（Computed）に変更
2. `description` コマンド生成を削除
3. パーサーにインデントチェックを追加し、コンテキスト外の description がトンネルに割り当てられないように修正

**修正ファイル**:
| ファイル | 変更内容 |
|----------|----------|
| `internal/rtx/parsers/tunnel.go` | インデントチェック追加、description コマンド生成削除 |
| `internal/rtx/parsers/tunnel_test.go` | IKE コマンドの ID を tunnel_id に修正 |
| `internal/provider/resources/tunnel/resource.go` | name 属性を Computed のみに変更 |
| `examples/import/main.tf` | name 属性を削除 |

### ipsec_tunnel: show config コマンド修正

**修正日**: 2026-01-31

`BuildShowIPsecConfigCommand()` が `show config | grep ipsec` を返していたため、`tunnel select N` 行がパースできず `tunnel_id` が null になる問題を修正。

**修正内容**: `show config` を返すように変更し、パーサーがトンネルコンテキストを正しく取得できるようにした。

### examples/import/main.tf: IPsec tunnel_id 修正

**修正日**: 2026-01-31

`tunnel_id`属性は`tunnel select N`に対応するため、main.tfを修正。

**修正前**:
- `rtx_ipsec_tunnel.ipsec101` with `tunnel_id = 101` → 存在しないトンネル
- `rtx_ipsec_tunnel.ipsec1` with `tunnel_id = 1` → 実際はtunnel select 2

**修正後**:
- `rtx_ipsec_tunnel.tunnel1` with `tunnel_id = 1` → tunnel select 1 (ipsec tunnel 101, L2TPv3)
- `rtx_ipsec_tunnel.tunnel2` with `tunnel_id = 2` → tunnel select 2 (ipsec tunnel 1, L2TP anonymous)

**ルーター設定との対応関係**:
| Terraform resource | tunnel_id | ルーターコマンド |
|--------------------|-----------|------------------|
| `rtx_ipsec_tunnel.tunnel1` | 1 | `tunnel select 1` + `ipsec tunnel 101` |
| `rtx_ipsec_tunnel.tunnel2` | 2 | `tunnel select 2` + `ipsec tunnel 1` |
