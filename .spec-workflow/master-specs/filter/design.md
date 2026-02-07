# Master Design: Filter Resources

> **DEPRECATED (2026-02-07)**: This spec has been superseded by the [access-list design](../access-list/design.md).
>
> The resource names in this spec (`rtx_interface_acl`, `rtx_interface_mac_acl`, `rtx_ip_filter_dynamic`, `rtx_ipv6_filter_dynamic`) do not exist in the implementation. The actual resources are all under the `access_list_*` namespace. See the access-list spec for the authoritative design document.

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2025-01-23 | Implementation Code | Initial master design created from implementation analysis |
| 2026-01-25 | Resource consolidation | Removed `rtx_ethernet_filter` resource |
| 2026-02-01 | Structure Sync | Updated file paths to resources/{name}/ modular structure |
| 2026-02-07 | Implementation Audit | DEPRECATED: spec replaced by access-list design; resource names did not match implementation |
