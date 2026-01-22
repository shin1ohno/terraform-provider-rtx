# Reconciliation

## Product principles
- Resource naming matches Cisco-style admin/user split and keeps secrets out of state outputs.
- State stores configuration inputs only; operational login status is not persisted.

## Implementation alignment
- `rtx_admin` handles login/admin passwords; `rtx_admin_user` covers administrator flag, connection methods, GUI pages, login_timer, and import reads attributes.
- Passwords are marked sensitive and not read back, consistent with security expectations.
- Gaps: no support for encrypted password variants or validation of hash formats; `rtx_admin` import only sets ID without verifying existing config; no warnings when router-side attributes differ from state defaults.
