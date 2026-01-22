# Reconciliation

## Product principles
- Catalogs/documentation reinforce predictable parsing and align with reuse of Cisco-like schemas while keeping parsing/state concerns separated.

## Implementation alignment
- YAML catalogs exist in `internal/rtx/testdata/patterns` for IP filters, ethernet filters, bridge, service, system, IPv6 interface/prefix, NAT, OSPF, BGP, etc.; `ipsec_transport_test.go` exercises transport mode parsing.
- Catalog schema follows shared `schema.yaml`, enabling parser validation and fixture reuse.
- Gaps: catalogs lack visible version/last-updated metadata and some parameters lack documented defaults/source references; automation to generate coverage reports from catalogs is not in place.
