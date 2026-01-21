# Design: Network Hardening (MAC ACL expansion, PPP/L2TP session controls, IPsec/IKE depth)

## Goals
- Keep Cisco-aligned naming while covering RTX semantics for Ethernet filters, PPP/L2TP robustness, and IPsec/IKE interoperability.
- Avoid breaking existing configs: new fields optional, legacy behavior unchanged by default.

## Scope
1) MAC ACL拡張でRTX `ethernet filter`機能を吸収  
2) PPP/PPPoE/L2TPのキープアライブ・再接続・MTU/MRU/MSS制御追加  
3) IPsec/IKE機能拡張（NAT-T、アルゴリズム拡充、IKEv1対応）

## MAC ACL Expansion
- **Schema (rtx_access_list_mac)**  
  - Optional `filter_id` (int) at ACL level; when set, emit RTX `ethernet filter` commands.  
  - Entry action: existing `permit/deny` plus `pass-log`, `pass-nolog`, `reject-log`, `reject-nolog` (mutually exclusive set).  
  - Entry `dhcp_match` block: `type` (`dhcp-bind`/`dhcp-not-bind`), optional `scope` or `ip`. Disallow pairing with explicit MACs.  
  - Entry offset matching: `offset` (int>=0) + `byte_list` (hex bytes or `*`, max 16).  
  - Optional `apply` block: `interface`, `direction` (in/out), ordered `filter_ids`. Only used when `filter_id` is set.
- **Command generation/parsing**  
  - `filter_id` present: use RTX `ethernet filter <id> ...` with action/mac/DHCP/offset/ethertype/vlan.  
  - `filter_id` absent: keep current ACL-style output; ignore RTX-only fields to preserve compatibility.  
  - Apply commands: `ethernet <if> filter in|out <ids...>`; preserve order; delete with `no ethernet ...`.
- **Validation**  
  - MAC format `xx:xx:xx:xx:xx:xx` or `*`; byte_list length ≤16; action set exclusivity; DHCP and MAC mutually exclusive.  
  - filter_id bounds per platform (1..512/100) with conservative validation.  
  - apply requires filter_ids when present.
- **Migration**  
  - Default: no new fields -> legacy behavior.  
  - Users enabling RTX semantics set `filter_id` and new fields; state schema tolerant to missing fields on import.

## PPP/PPPoE/L2TP Session Controls
- **Schema additions**  
  - PPP/PPPoE: `lcp_echo_interval`, `lcp_echo_failure`, `reconnect_interval`, `reconnect_attempts`, `auto_connect`, `mtu`, `mru`, `mss_adjust`, `pppoe_service/ac`, `ipcp_dns` toggles.  
  - L2TP: `always_on`, `keepalive_interval`, `keepalive_log`, `disconnect_timer`, `max_sessions`.  
- **Command mapping**  
  - LCP: map to `pp lcp echo on interval <n> failure <m>` (or RTX equivalent).  
  - Reconnect: reconnect interval/attempts; auto_connect on boot.  
  - MTU/MRU/MSS: apply to PPP interface (`pp <id>`).  
  - L2TP: keepalive/log, always_on, disconnect timer, max-sessions.  
- **Validation**  
  - Interval/failure paired; ranges per RTX defaults; MSS/MTU bounds; attempts >=0.  
  - Avoid overwriting router defaults on import when field absent.

## IPsec/IKE Depth
- **Schema additions**  
  - NAT-T: `nat_traversal` (on/off) + `nat_keepalive_interval`.  
  - Algorithms: AES-GCM (128/256), SHA-512, DH groups (14/16/19/20 where supported), PFS groups.  
  - IKEv1: `ike_version` (v1/v2), `mode` (main/aggressive for v1), `identity` (address/fqdn/user-fqdn).  
  - Cert placeholders: `local_cert`, `remote_id` (Sensitive).  
- **Command mapping**  
  - Map NAT-T to RTX ipsec ike commands; include keepalive when set.  
  - Algorithm flags in proposal/transform; validate legal combos (e.g., GCM implies integrity off).  
  - IKEv1 mode/identity to appropriate commands; keep v2 defaults unchanged.
- **Validation**  
  - Algorithm compatibility, group ranges, required identity for aggressive mode.  
  - Sensitive fields flagged; no default overwrite on import.

## Testing Strategy (TDD徹底)
- RTXコマンドバリエーションを網羅するテーブル駆動テストを先に生成し、RED→GREENで実装。  
- 生成元: 既存のパターンカタログに加え、新規フィールドの組合せを列挙（MAC: action/DHCP/offset/apply、PPP: keepalive/reconnect/MTU、IPsec: NAT-T/IKEv1/algorithms）。  
- ラウンドトリップ必須: build→parse→buildが同一コマンド列/構造になることを全ケースで検証。  
- リグレッション: 新フィールド未指定の従来コマンド群を固定フィクスチャとして保持し、diffが出ないことを確認。  
- Importテスト: Router出力に欠落フィールドがある場合でもstateがデフォルト上書きされないことをケース化。

## Rollout / Compatibility
- All new fields optional; no behavior change without opt-in.  
- Documentation to guide: how to enable RTX semantics for MAC filters, how to set PPP/L2TP keepalives, and how to configure NAT-T/IKEv1.
