# Reconciliation

## Product principles
- Resource treats PPTP as legacy but keeps Cisco-style admin/encryption/auth naming; secrets marked sensitive and no runtime status stored.

## Implementation alignment
- Current schema covers enabling service, authentication block, IP pool, and simple encryption toggle (MPPE bits), plus shutdown flag; CRUD/import use singleton ID.
- Parser/client manage basic PPTP on/off and credentials.
- Gaps: no client-mode settings, keepalive/disconnect timers, max-connections, stateless/required MPPE options, DNS assignment, or PP anonymous binding; import validation minimal.
