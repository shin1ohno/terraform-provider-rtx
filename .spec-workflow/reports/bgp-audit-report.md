# 差分レポート: bgp

## リファレンス情報
- ファイル: docs/RTX-commands/29_BGP.md
- 対象セクション: 29.1〜29.18

---

## 差分一覧

| # | 項目 | リファレンス（正） | 現在の記述（誤） | 場所 | 修正優先度 |
|---|------|------------------|-----------------|------|-----------|
| 1 | **bgp neighbor コマンド構文** | `bgp neighbor neighbor_id remote_as remote_address [parameter...]` | `bgp neighbor 1 address 203.0.113.1 as 65002` | specs/bgp/config.yaml:75 | **高** |
| 2 | **bgp neighbor パーサー** | `bgp neighbor neighbor_id remote_as remote_address` | `bgp neighbor (\d+) address ([0-9.]+) as (\d+)` | internal/rtx/parsers/bgp.go:66 | **高** |
| 3 | **bgp neighbor ビルダー** | `bgp neighbor %d %s %s` (ID AS IP) | `bgp neighbor %d address %s as %s` | internal/rtx/parsers/bgp.go:227 | **高** |
| 4 | **bgp autonomous-system AS範囲** | `1..65535` (2バイトAS) | `1..4294967295` (4バイトAS) | specs/bgp/config.yaml:53-56, 189-191 | **高** |
| 5 | **bgp autonomous-system バリデーション** | `1..65535` | `1..4294967295` | internal/rtx/parsers/bgp.go:301-305 | **高** |
| 6 | **bgp neighbor password コマンド** | `bgp neighbor pre-shared-key neighbor_id text text_key` | `bgp neighbor 1 password mysecret` | specs/bgp/config.yaml:122-127 | **高** |
| 7 | **bgp neighbor password パーサー** | `bgp neighbor pre-shared-key (\d+) text (.+)` | `bgp neighbor (\d+) password (.+)` | internal/rtx/parsers/bgp.go:70 | **高** |
| 8 | **bgp neighbor password ビルダー** | `bgp neighbor pre-shared-key %d text %s` | `bgp neighbor %d password %s` | internal/rtx/parsers/bgp.go:250-252 | **高** |
| 9 | **bgp neighbor hold-time 範囲** | `3..28,800` | `3..65535` | specs/bgp/config.yaml:215, internal/rtx/parsers/bgp.go:330 | 中 |
| 10 | **bgp neighbor remote_as 範囲** | `1..65535` (2バイトAS) | `1..4294967295` | specs/bgp/config.yaml:90-96 | **高** |

---

## 詳細分析

### 1. bgp neighbor コマンド構文の差分（最重要）

**リファレンス (29.12節, 324-326行目)**:
```
[書式]
bgp neighbor neighbor_id remote_as remote_address [parameter...]
no bgp neighbor neighbor_id [remote_as remote_address [parameter...]]
```

**現在のspec/実装**:
```
bgp neighbor 1 address 203.0.113.1 as 65002
```

**問題点**:
- リファレンスでは `address` や `as` というキーワードは使用されない
- パラメータの順序は `neighbor_id remote_as remote_address`
- 現在の実装は `neighbor_id address remote_address as remote_as` という独自構文

**正しい形式の例**:
```
bgp neighbor 1 65002 203.0.113.1
bgp neighbor 1 65002 203.0.113.1 hold-time=90
```

### 2. bgp autonomous-system AS番号範囲の差分

**リファレンス (29.4節)**:
```
• [設定値] : AS 番号 (1..65535)
```

**現在のspec**:
```yaml
- name: bgp_asn_4byte
  rtx: "bgp autonomous-system 4200000001"
  terraform:
    asn: "4200000001"
```

**問題点**: リファレンスは2バイトAS (1-65535) のみサポート。4バイトASは未サポート。

### 3. bgp neighbor pre-shared-key コマンドの差分

**リファレンス (29.13節)**:
```
[書式]
bgp neighbor pre-shared-key neighbor_id text text_key
no bgp neighbor pre-shared-key neighbor_id [text text_key]
```

**現在のspec/実装**:
```
bgp neighbor 1 password mysecret
```

**問題点**: 完全に異なるコマンド構文。`pre-shared-key` サブコマンドと `text` キーワードが必要。

---

## 影響範囲

- [x] specs/ の修正が必要
- [ ] master-specs/ の修正が必要（bgp用のmaster-specsは存在しない）
- [x] 実装（パーサー/サービス）の修正が必要

---

## 次のステップ

### spec-fixer に渡す内容

```yaml
resource: bgp
fixes:
  - location: specs/bgp/config.yaml
    issues:
      - line: 75
        current: 'bgp neighbor 1 address 203.0.113.1 as 65002'
        correct: 'bgp neighbor 1 65002 203.0.113.1'

      - line: 83
        current: 'bgp neighbor 2 address 203.0.113.2 as 65003'
        correct: 'bgp neighbor 2 65003 203.0.113.2'

      - line: 91
        current: 'bgp neighbor 1 address 10.0.0.1 as 4200000001'
        correct: 'bgp neighbor 1 65001 10.0.0.1'
        note: 'AS番号も2バイトAS範囲内に変更必要'

      - line: 49-56
        issue: '4バイトASNのテストケースを削除または修正'
        note: 'リファレンスは1-65535のみサポート'

      - line: 122-127
        current: 'bgp neighbor 1 password mysecret'
        correct: 'bgp neighbor pre-shared-key 1 text mysecret'

      - line: 189-191
        current: '4294967295を最大値'
        correct: '65535を最大値に変更'

      - line: 205-214
        current: 'hold_time max: 65535'
        correct: 'hold_time max: 28800'
```

### 実装修正に渡す内容

```
ファイル: internal/rtx/parsers/bgp.go

1. 行66: bgpNeighborAddrPattern の修正
   現在: `bgp\s+neighbor\s+(\d+)\s+address\s+([0-9.]+)\s+as\s+(\d+)`
   正しい: `bgp\s+neighbor\s+(\d+)\s+(\d+)\s+([0-9.]+)`

2. 行70: bgpNeighborPasswordPattern の修正
   現在: `bgp\s+neighbor\s+(\d+)\s+password\s+(.+)`
   正しい: `bgp\s+neighbor\s+pre-shared-key\s+(\d+)\s+text\s+(.+)`

3. 行227: BuildBGPNeighborCommand の修正
   現在: fmt.Sprintf("bgp neighbor %d address %s as %s", ...)
   正しい: fmt.Sprintf("bgp neighbor %d %s %s", neighbor.ID, neighbor.RemoteAS, neighbor.IP)

4. 行250-252: BuildBGPNeighborPasswordCommand の修正
   現在: fmt.Sprintf("bgp neighbor %d password %s", ...)
   正しい: fmt.Sprintf("bgp neighbor pre-shared-key %d text %s", ...)

5. 行301-305: ASN バリデーション範囲の修正
   現在: 1-4294967295
   正しい: 1-65535

6. 行330: hold-time バリデーション範囲の修正
   現在: 3-65535
   正しい: 3-28800
```

---

## 判定

**[AUDIT:1]** — 差分が見つかった

重大な構文差分が複数存在し、早急な修正が必要。
