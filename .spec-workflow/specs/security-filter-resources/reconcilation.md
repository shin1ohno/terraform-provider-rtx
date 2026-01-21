# Reconciliation

## Product principles
- Implements core security resources to keep configuration Cisco-aligned while covering RTX-specific filters; state remains configuration-only.

## Implementation alignment
- IP/IPv6 static and dynamic filter resources exist with restrict actions and key protocols; interface bindings are supported via `rtx_interface`/`rtx_ipv6_interface` secure_filter fields; IPsec transport resource present with tunnel_id/protocol/port; L2TP service resource exists for enablement.
- Pattern catalogs and parsers handle filter lists and IPsec transport entries.
- Gaps: Ethernet filter implemented as MAC ACL rather than numeric `ethernet filter` semantics, IPv6 filter protocol/port coverage limited, L2TP service options minimal (no protocol list/disable handling), and validation for cross-resource references (filter existence/tunnel linkage) is absent.
