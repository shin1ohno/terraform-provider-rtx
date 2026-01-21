# Reconciliation

## Product principles
- Resource follows Cisco-like proposal/transform naming and keeps PSKs sensitive; state excludes SA status.

## Implementation alignment
- Schema covers tunnel_id, endpoints, PSK, local/remote networks, basic IKEv2 proposal booleans (encryption/integrity/DH), IPsec transform settings, DPD toggles, and enable flag; CRUD/import implemented.
- Client/parsers handle basic IKEv2/ESP setup and lifetime fields.
- Gaps: no IKEv1/aggressive/cert auth, no transport-mode toggle (separate transport resource handles part), no PFS groups beyond 14/5/2, no NAT-T/keepalive controls, no Phase1/2 algorithm matrices (AES-GCM, SHA-512), no identity/peer FQDN handling, and no route selection/encryption-domain mapping.
