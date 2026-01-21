# Reconciliation

## Product principles
- Resources reuse Cisco-style PP/PPPoE naming and keep credentials sensitive; state excludes live session status.

## Implementation alignment
- `rtx_pppoe`/`rtx_pp_interface` provide basic fields for interface selection, username/password, service/ac name, auth method, and link to pp interface; IPCP/DNS and default route handling are minimal.
- Parsers for PPP exist and cover secure filters import.
- Gaps: no LCP echo/keepalive controls, no reconnect/timeout knobs, no compression/MRU/MSS tuning, limited routing/DNS/static IP options, no session priority or multi-session management, and imports do not preserve encrypted passwords.
