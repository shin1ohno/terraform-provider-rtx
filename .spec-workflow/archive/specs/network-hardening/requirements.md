# Requirements Document: Network Hardening Gap Fixes

## Introduction

This spec bundles three high-risk gaps to align implementation with RTX semantics and product principles:
- **Ethernet filter parity**: Implement native RTX `ethernet filter` semantics instead of MAC ACL stand-in.
- **PPP/PPPoE/L2TP session robustness**: Add keepalive/reconnect controls and session limits for WAN/VPN stability.
- **IPsec/IKE feature depth**: Cover NAT-T, broader algorithms, and IKEv1 support for interoperability.

## Alignment with Product Vision
- Matches “comprehensive RTX feature coverage” and “Cisco-compatible syntax” while keeping configuration-only state.
- Improves reliability for WAN/VPN connectivity and firewall correctness.

## Requirements

### REQ-1: MAC ACL Expansion to Cover RTX Ethernet Filter Semantics
**User Story:** As a network administrator, I want Cisco-style MAC ACLs to express RTX `ethernet filter` capabilities without introducing a new resource.

**Acceptance Criteria**
1. Extend existing MAC ACL resource to optionally accept a `filter_id` (numeric). If set, generate/parse RTX `ethernet filter <id> ...` commands; if omitted, preserve current name-based ACL behavior.
2. Support RTX actions: `pass-log`, `pass-nolog`, `reject-log`, `reject-nolog` in addition to existing permit/deny. Action is scoped per entry and mutually exclusive with permit/deny to avoid ambiguity.
3. Support DHCP-based matches: entry-level `dhcp_match` block with `type` (`dhcp-bind` | `dhcp-not-bind`) and optional `scope`/`ip`. Disallow mixing DHCP match with explicit MAC addresses in the same entry.
4. Support offset/byte_list matching: optional `offset` (int) and `byte_list` (hex bytes or `*`, max 16) per entry to mirror RTX offset matching.
5. Preserve ordering and application: optional `apply` block for `interface`/`direction` (in/out) with ordered list of filter_ids; when `filter_id` is unset, behavior remains unchanged (no apply commands).
6. Validation: enforce MAC pattern (`xx:xx:xx:xx:xx:xx` or `*`), filter_id bounds (platform limits), byte_list format/length, action exclusivity, and disallow conflicting DHCP/MAC fields.
7. Backward compatibility: default behavior unchanged when new fields are omitted; imports must not break existing state. Provide migration guidance for users adopting `filter_id`/RTX actions.

### REQ-2: PPP/PPPoE/L2TP Session Controls
**User Story:** As a WAN/VPN operator, I want keepalive and reconnect controls so PPP/PPPoE and L2TP sessions recover reliably.

**Acceptance Criteria**
1. Add LCP echo knobs to PPP/PPPoE: `lcp-echo-interval`, `lcp-echo-failure`, and reconnect/backoff timers; import existing values.
2. Support connection state controls: auto-connect on boot, manual connect/disconnect commands, and max reconnect attempts.
3. Expose MTU/MRU/MSS adjust fields aligned with PPPoE defaults; validate ranges.
4. L2TP: add keepalive/log options, always-on toggle, max-sessions, and per-tunnel disconnect timer; preserve tunnel_auth fields.
5. Validation rejects conflicting settings (e.g., keepalive disabled with failure threshold set) and enforces required pairs (interval + failure).
6. Tests cover DHCP/STATIC IPCP cases, failover/reconnect scenarios, and import of encrypted passwords (store as sensitive without diff).

### REQ-3: IPsec/IKE Feature Depth
**User Story:** As a VPN administrator, I need broader IPsec/IKE coverage to interoperate with varied peers.

**Acceptance Criteria**
1. Support NAT-T and keepalive controls (on/off, interval) for tunnels; map to RTX CLI.
2. Extend algorithms: AES-GCM (128/256), SHA-512, additional DH groups (14/16/19/20 if supported), with validation of legal combos.
3. Add IKEv1 (main/aggressive) option with identity (address/fqdn/user-fqdn) and pre-shared-key handling; preserve IKEv2 behavior.
4. Allow certificate-based auth stubs (schema + validation) even if CLI wiring ships later behind feature flag; mark private keys/certs sensitive.
5. Add PFS groups and lifetimes per phase with metric validation; import must not overwrite missing router fields with defaults silently.
6. Tests: round-trip new algorithms, NAT-T on/off, IKEv1 aggressive with identities, and AES-GCM interop cases.

## Non-Functional Requirements
- **Validation**: Strong type/range checks; cross-field validation for dependent parameters.
- **Backward compatibility**: New fields optional with safe defaults; existing configs/imports continue to work.
- **Security**: All secrets (PSK, passwords, keys) marked sensitive; logs sanitized via logging utilities.
- **Testing**: Table-driven parser/service/resource tests for new CLI paths; fixture updates for import.

## Deliverables
- Requirements implemented in provider schema, client services, parsers, and tests.
- Migration notes for Ethernet filter transition and new PPP/IPsec options.
