# Reconciliation

## Product principles
- Naming (`name_servers`, `domain_name`, `server_select`, `hosts`) aligns with Cisco DNS conventions; state excludes runtime resolver status.

## Implementation alignment
- Resource covers domain_lookup, domain_name, up to 3 name_servers, server_select with EDNS/record_type/query_pattern/restrict_pp/original_sender, static hosts, service_on, private_address_spoof; import supports singleton `dns`.
- Parser/build helpers handle server-select permutations and static hosts.
- Gaps: no DNS proxy/cache controls, no authoritative/server-mode toggles, no per-interface servers, TTL/control for static hosts absent, and validations for domain formats/record_type defaults are minimal.
