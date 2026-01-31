# Session Progress

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
