# Master Requirements: Filter Resources

> **DEPRECATED (2026-02-07)**: This spec has been superseded by the [access-list spec](../access-list/requirements.md).
>
> The resource names in this spec (`rtx_interface_acl`, `rtx_interface_mac_acl`, `rtx_ip_filter_dynamic`, `rtx_ipv6_filter_dynamic`) do not exist in the implementation. The actual resources are:
>
> | Former Spec Name | Actual Resource Name | Spec Location |
> |------------------|---------------------|---------------|
> | `rtx_interface_acl` | `rtx_access_list_ip_apply` / `rtx_access_list_ipv6_apply` | access-list spec |
> | `rtx_interface_mac_acl` | `rtx_access_list_mac_apply` | access-list spec |
> | `rtx_ip_filter_dynamic` | `rtx_access_list_ip_dynamic` | access-list spec |
> | `rtx_ipv6_filter_dynamic` | `rtx_access_list_ipv6_dynamic` | access-list spec |
>
> Useful content from this spec (parsing reliability requirements, RTX line wrapping handling, dynamic filter form detection) has been integrated into the access-list spec.

## Change History

| Date | Source | Changes |
|------|--------|---------|
| 2025-01-23 | Implementation Code | Initial master spec created from implementation analysis |
| 2026-01-23 | filter-number-parsing-fix | Added parsing reliability requirements for line wrapping handling |
| 2026-01-23 | terraform-plan-differences-fix | Ethernet filter parser accepts `*:*:*:*:*:*` MAC wildcard format |
| 2026-02-07 | Implementation Audit | DEPRECATED: spec replaced by access-list spec; resource names did not match implementation |
