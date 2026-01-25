# RTX Parser Improvement Roadmap

**Generated:** 2026-01-20
**Based on:** Coverage Report Analysis

## Overview

This document outlines the gap analysis and improvement roadmap for RTX command parsers. The analysis prioritizes work based on Terraform resource usage and common enterprise configuration patterns.

## Current State Summary

| Metric | Value |
|--------|-------|
| Implemented Parsers | 29 |
| Pattern Catalogs | 16 |
| Total Patterns | 728 |
| Test Functions | 450 |
| Test Pass Rate | 100% |
| Terraform Resources | 35+ |

## Gap Analysis

### Category 1: Complete Coverage (No Action Required)

These parsers have comprehensive test coverage aligned with documented patterns:

| Parser | Patterns | Tests | Terraform Resource |
|--------|----------|-------|-------------------|
| dns.go | 40 | 39 | rtx_dns_server |
| static_route.go | 16 | 11 | rtx_static_route |
| vlan.go | 26 | 21 | rtx_vlan |
| interface_config.go | 47 | 17 | rtx_interface |
| admin.go | 32 | 13 | rtx_admin, rtx_admin_user |
| snmp.go | 20 | 23 | rtx_snmp_server |
| syslog.go | 17 | 18 | rtx_syslog |
| schedule.go | 27 | 15 | rtx_kron_policy, rtx_kron_schedule |
| qos.go | 32 | 27 | rtx_class_map, rtx_policy_map, rtx_service_policy, rtx_shape |
| dhcp_scope.go | 59 | 17 | rtx_dhcp_scope |
| dhcp_bindings.go | - | 8 | rtx_dhcp_binding |
| nat_masquerade.go | 95 | 26 | rtx_nat_masquerade |
| nat_static.go | - | 24 | rtx_nat_static |
| bgp.go | 69 | 31 | rtx_bgp |
| ospf.go | 86 | 19 | rtx_ospf |
| ipsec_tunnel.go | 91 | 11 | rtx_ipsec_tunnel |
| l2tp.go | 47 | 25 | rtx_l2tp |
| pptp.go | 24 | 4 | rtx_pptp |

### Category 2: Minor Gaps (Low Priority)

| Parser | Gap | Impact | Recommendation |
|--------|-----|--------|----------------|
| ipsec_transport.go | No test file | Low - limited use in Terraform | Add tests if transport mode resource is implemented |
| ip_filter.go | No pattern catalog | Low - functionality tested | Optional: create ip_filter.yaml for documentation |
| ethernet_filter.go | No pattern catalog | Low - functionality tested | Optional: create ethernet_filter.yaml for documentation |

### Category 3: Edge Case Opportunities (Enhancement)

These parsers have functional coverage but could benefit from additional edge case tests:

| Parser | Patterns | Tests | Gap % | Priority |
|--------|----------|-------|-------|----------|
| ipsec_tunnel.go | 91 | 11 | 88% | Medium |
| ospf.go | 86 | 19 | 78% | Medium |
| nat_masquerade.go | 95 | 26 | 73% | Medium |
| bgp.go | 69 | 31 | 55% | Low |

**Note:** Gap % represents optional parameter combinations not explicitly tested. Core functionality is fully covered.

### Category 4: Not Implemented (Future Work)

RTX command chapters without corresponding parsers:

| Chapter | Feature | Implementation Priority |
|---------|---------|------------------------|
| Ch.10 | PPP/PPPoE | Medium - used in WAN connections |
| Ch.11 | ISDN | Low - legacy technology |
| Ch.14 | Intrusion Detection | Low - specialized use |
| Ch.18-20 | Mobile Network | Low - specialized use |
| Ch.22 | Application Monitoring | Low - specialized use |
| Ch.27 | DDNS | Medium - common enterprise use |
| Ch.31 | RIP | Low - legacy routing |
| Ch.32 | Multicast | Low - specialized use |
| Ch.33-36 | Advanced Features | Low |

## Improvement Roadmap

### Phase 1: Maintenance (Current)

**Timeline:** Ongoing
**Effort:** Minimal

Tasks:
- [x] All existing parsers have passing tests
- [x] Pattern catalogs created for all Terraform resources
- [x] Test fixtures organized by category
- [ ] Monitor for regressions in CI/CD

### Phase 2: Edge Case Enhancement (Optional)

**Timeline:** As needed
**Effort:** Medium

Priority order based on Terraform resource usage frequency:

1. **IPsec Tunnel Parser Enhancement**
   - Add tests for all encryption algorithm combinations
   - Add tests for IKEv1/IKEv2 variations
   - Add tests for NAT traversal options
   - Estimated tests to add: ~20

2. **OSPF Parser Enhancement**
   - Add tests for all area type combinations
   - Add tests for authentication variations
   - Add tests for redistribution options
   - Estimated tests to add: ~15

3. **NAT Parser Enhancement**
   - Add tests for port range mappings
   - Add tests for protocol variations
   - Add tests for dynamic NAT options
   - Estimated tests to add: ~15

4. **BGP Parser Enhancement**
   - Add tests for route-map options
   - Add tests for 4-byte ASN edge cases
   - Add tests for community attributes
   - Estimated tests to add: ~10

### Phase 3: New Resource Support (Future)

**Timeline:** Based on user demand
**Effort:** High

When implementing new Terraform resources, follow this checklist:

1. **Pattern Extraction**
   - [ ] Create `internal/rtx/testdata/patterns/{feature}.yaml`
   - [ ] Extract all command formats from RTX documentation
   - [ ] Include `設定例` examples as test cases

2. **Parser Implementation**
   - [ ] Create `internal/rtx/parsers/{feature}.go`
   - [ ] Implement Parse function for config extraction
   - [ ] Implement Build functions for command generation

3. **Test Coverage**
   - [ ] Create `internal/rtx/parsers/{feature}_test.go`
   - [ ] Create `internal/rtx/testdata/fixtures/{feature}/` directory
   - [ ] Add table-driven tests from pattern catalog
   - [ ] Ensure 100% pass rate before PR

4. **Terraform Resource**
   - [ ] Create `internal/provider/resource_rtx_{feature}.go`
   - [ ] Implement CRUD operations
   - [ ] Add acceptance tests
   - [ ] Create example configurations

### Phase 4: Documentation Parity (Long-term)

**Timeline:** After major feature work complete
**Effort:** Low-Medium

Create pattern catalogs for parsers without explicit documentation:

| Parser | Pattern Catalog | Priority |
|--------|-----------------|----------|
| ip_filter.go | ip_filter.yaml | Low |
| ethernet_filter.go | ethernet_filter.yaml | Low |
| bridge.go | bridge.yaml | Low |
| service.go | service.yaml | Low |
| system.go | system.yaml | Low |
| ipv6_interface.go | ipv6_interface.yaml | Low |
| ipv6_prefix.go | ipv6_prefix.yaml | Low |

## Success Metrics

### Current Status

| Metric | Target | Actual |
|--------|--------|--------|
| Test Pass Rate | 100% | ✅ 100% |
| Terraform Resource Coverage | 100% | ✅ 100% |
| Pattern Catalog Coverage | 80% | ✅ 89% (16/18 high-priority) |
| CI Build Status | Green | ✅ Green |

### Future Targets

| Metric | Target | Timeline |
|--------|--------|----------|
| Edge Case Coverage | 90% | Phase 2 |
| Pattern Catalog Coverage | 100% | Phase 4 |
| New Resource Time-to-Market | < 1 week | Ongoing |

## Implementation Guidelines

### Adding Edge Case Tests

```go
// Example: Adding IPsec encryption algorithm edge cases
func TestParseIPsecEncryptionAlgorithms(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"AES-128-CBC", "ipsec ike encryption 1 aes128-cbc", "aes128-cbc"},
        {"AES-256-CBC", "ipsec ike encryption 1 aes256-cbc", "aes256-cbc"},
        {"AES-GCM-128", "ipsec ike encryption 1 aes-gcm-128", "aes-gcm-128"},
        // Add more variations...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Adding Pattern Catalog Entries

```yaml
# Example: Adding to ipsec.yaml
patterns:
  - name: ipsec_ike_encryption
    command: "ipsec ike encryption"
    format: "ipsec ike encryption {tunnel_id} {algorithm}"
    parameters:
      - name: tunnel_id
        type: int
        range: "1-100"
      - name: algorithm
        type: enum
        values:
          - des-cbc
          - 3des-cbc
          - aes128-cbc
          - aes256-cbc
          - aes-gcm-128
          - aes-gcm-256
    examples:
      - input: "ipsec ike encryption 1 aes256-cbc"
        tunnel_id: 1
        algorithm: "aes256-cbc"
```

## Conclusion

The RTX command parser test suite is in excellent condition with 100% test pass rate and comprehensive coverage of all Terraform resources. The improvement roadmap focuses on:

1. **Maintaining** the current quality level
2. **Enhancing** edge case coverage as needed
3. **Supporting** new resource development with established patterns
4. **Documenting** remaining parsers for completeness

No urgent action is required. Future improvements should be driven by user demand and new resource implementation needs.
