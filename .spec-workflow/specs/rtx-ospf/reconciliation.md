# Reconciliation

## Product principles
- Resource name/fields mirror Cisco IOS XE (process_id/router_id/networks/neighbors); state avoids operational neighbor status.

## Implementation alignment
- Schema includes process_id, router_id, distance, default_information_originate, networks (ip/wildcard/area), areas (id/type/no_summary), neighbors (ip/priority/cost), and redistribution of static/connected; CRUD/import implemented via client/parsers.
- Parser registry present; import treats OSPF as singleton.
- Gaps: no area types beyond stub/NSSA options, no interface-level timers/authentication/cost/priority, no redistribution metrics/metric-type or route filtering, no passive-interface handling, and no enable/disable toggle beyond implicit Enabled flag; validations for router/area IDs minimal.
