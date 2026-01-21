# Reconciliation

## Product principles
- Naming keeps Cisco-like tunnel/proposal fields while reflecting RTX L2TPv2/v3 concepts; state avoids storing session status and keeps secrets sensitive.

## Implementation alignment
- Schema supports tunnel_id, version (v2/v3), mode (lns/l2vpn), shutdown, source/destination, auth block (method/user/pass), IP pool, optional IPsec profile (PSK/tunnel_id), and L2TPv3 config (router IDs, bridge interface, cookie_size, tunnel auth). Import/CRUD implemented.
- Tunnel auth parsing present per import-fidelity fixes.
- Gaps: no PP anonymous/client-side LAC settings, no always-on/keepalive timers, no max-sessions, no PAP/CHAP per-user database, no IPsec certificate support, and limited validation of tunnel source/destination types; L2TP over IPsec integration is shallow (no rekey/phase settings).
