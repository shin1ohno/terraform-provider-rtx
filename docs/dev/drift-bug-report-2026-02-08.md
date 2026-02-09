# Bug Report: Persistent Drift After Apply

**Date**: 2026-02-08
**Observed in**: home-monitor project (`terraform apply` → `terraform plan`)
**Affected resources**: `rtx_dns_server`, `rtx_tunnel`, `rtx_ipsec_tunnel`, `rtx_class_map`, `rtx_kron_schedule`, `rtx_dhcp_scope`, `rtx_sftpd`

## Summary

After a successful `terraform apply`, a subsequent `terraform plan` still reports changes for `rtx_dns_server.main` and `rtx_tunnel.hnd_itm`. The RTX router accepts the configuration, but the provider's Read function does not correctly reconcile the state.

Bug 2 (empty list vs null) is a systemic issue affecting multiple resources across the provider.

---

## Bug 1: `rtx_dns_server` — server_select ordering and ghost entries

### Symptoms

1. `priority` fields show as `X -> (known after apply)`
2. `record_type` and `address` values are swapped between server_select blocks (A ↔ AAAA)
3. A 4th server_select block (priority 11, AAAA) appears in state despite only 3 blocks in config

### Terraform Config (abbreviated)

```hcl
resource "rtx_dns_server" "main" {
  priority_start = 1
  priority_step  = 5

  server_select {  # expected priority: 1
    query_pattern = "home.local"
    record_type   = "any"
    server { address = "100.100.100.100" }
  }
  server_select {  # expected priority: 6
    query_pattern = "."
    record_type   = "a"
    server { address = "1.1.1.1" }
    server { address = "1.0.0.1" }
  }
  server_select {  # expected priority: 11
    query_pattern = "."
    record_type   = "aaaa"
    server { address = "2606:4700:4700::1111" }
    server { address = "2606:4700:4700::1001" }
  }
}
```

### Root Cause

**`reorderServerSelectToMatchPlan()` in `model.go:313-366`**

The function builds `currentByPriority` from the state read back from the router, then tries to match entries by calculated priority. Two problems:

1. **No fallback when priority lookup fails**: If `currentByPriority[priority]` misses (e.g. router returned entries in a different order or with unexpected IDs), `reorderedValues[i]` remains `nil`. This nil value propagates into the state list, causing `(known after apply)` in the plan.

2. **Router returns extra entries**: The router may echo back entries that overlap or differ from what was sent. The function only maps entries it can find by priority — unmatched router entries are silently appended or lost, causing ghost entries (the 4th server_select at priority 11).

**Parser issue (`internal/rtx/dns.go:222-312`)**: The `parseDNSServerSelectFields` function may misidentify field positions, swapping `record_type` with server addresses when the RTX command format varies.

### Relevant Code

- `internal/provider/resources/dns_server/model.go:308-375` — `reorderServerSelectToMatchPlan()`
- `internal/provider/resources/dns_server/model.go:174-179` — `FromClient()` sorts by ID
- `internal/rtx/dns.go:222-312` — `parseDNSServerSelectFields()`

---

## Bug 2: `rtx_tunnel` — `secure_filter_in = []` perpetual drift

### Symptoms

Every `terraform plan` shows:

```diff
  ipsec {
+   secure_filter_in = []
  }
```

The config explicitly sets `secure_filter_in = []`, but the state always stores it as `null`.

### Root Cause

**Empty list vs null distinction lost in `convertIntSliceToList()` / `convertListToIntSlice()`**

In `model.go:14-48`:

```go
// convertListToIntSlice — line 15
func convertListToIntSlice(list types.List) []int {
    if list.IsNull() || list.IsUnknown() {
        return nil
    }
    elements := list.Elements()
    if len(elements) == 0 {
        return nil  // BUG: explicit empty list [] becomes nil
    }
    // ...
}

// convertIntSliceToList — line 36
func convertIntSliceToList(ints []int) types.List {
    if len(ints) == 0 {
        return types.ListNull(types.Int64Type)  // BUG: nil and empty both become null
    }
    // ...
}
```

The round-trip is:

1. Config: `secure_filter_in = []` (empty list, NOT null)
2. `convertListToIntSlice([])` → returns `nil` (loses empty vs null distinction)
3. Router has no filter line → parser returns `nil`
4. `convertIntSliceToList(nil)` → returns `types.ListNull()` (null, NOT empty list)
5. State stores `null`, config says `[]` → Terraform sees a diff

**Partial workaround exists**: `resource.go:483-515` preserves planned values during `Update()`, but this does not cover `Read()` or `Create()` → `Read()` flow.

### Relevant Code

- `internal/provider/resources/tunnel/model.go:14-48` — `convertListToIntSlice()` / `convertIntSliceToList()`
- `internal/provider/resources/tunnel/resource.go:483-515` — partial workaround in `Update()`

---

## Suggested Fixes

### For Bug 1 (dns_server)

- In `reorderServerSelectToMatchPlan()`, add explicit handling for missing priority keys — fall back to matching by `(query_pattern, record_type)` tuple instead of solely by priority.
- Ensure `FromClient()` filters out router entries that don't correspond to any planned server_select block, preventing ghost entries.
- Add parser tests with RTX command variations to catch field-position misidentification.

### For Bug 2 (tunnel)

Distinguish `nil` (null/unset) from `[]int{}` (explicitly empty):

```go
func convertListToIntSlice(list types.List) []int {
    if list.IsNull() || list.IsUnknown() {
        return nil
    }
    elements := list.Elements()
    if len(elements) == 0 {
        return []int{}  // preserve empty vs nil
    }
    // ...
}

func convertIntSliceToList(ints []int) types.List {
    if ints == nil {
        return types.ListNull(types.Int64Type)
    }
    if len(ints) == 0 {
        elements := make([]attr.Value, 0)
        list, _ := types.ListValue(types.Int64Type, elements)
        return list  // empty list, not null
    }
    // ...
}
```

Also apply the same preservation logic from `Update()` to `Create()` and standalone `Read()`.

---

## Bug 2 Impact Analysis: All Affected Resources

The empty-list-vs-null bug is systemic. The same pattern exists in multiple resources.

### Severity: CRITICAL — Identical duplicated functions

These resources have their own copy of `convertIntSliceToList()`/`convertListToIntSlice()` with the exact same bug:

#### `rtx_ipsec_tunnel`

- **File**: `internal/provider/resources/ipsec_tunnel/model.go:14-48`
- **Affected attributes**: `secure_filter_in`, `secure_filter_out`
- **FromClient**: lines 198-199
- **Identical** duplicated functions — same bug as `rtx_tunnel`

#### `rtx_class_map`

- **File**: `internal/provider/resources/class_map/model.go:64-75`
- **Affected attributes**: `match_destination_port`, `match_source_port`
- **Function**: `intSliceToList()` (different name, same bug pattern)
- **FromClient**: lines 43-44

```go
// class_map/model.go:64-75
func intSliceToList(slice []int) types.List {
    if len(slice) == 0 {
        return types.ListNull(types.Int64Type)  // BUG: same pattern
    }
    // ...
}
```

### Severity: HIGH — Direct ListNull in FromClient

#### `rtx_kron_schedule`

- **File**: `internal/provider/resources/kron_schedule/model.go:81`
- **Affected attribute**: `command_lines` (types.List of strings)
- **Pattern**: Directly returns `types.ListNull(types.StringType)` when command list is empty

### Severity: MEDIUM — Nested block attributes

#### `rtx_dhcp_scope`

- **File**: `internal/provider/resources/dhcp_scope/model.go:134, 150`
- **Affected attributes**: `options.routers`, `options.dns_servers`
- **Context**: Inside optional `options` nested block — drift occurs if user writes `options { routers = [] }`

### Severity: LOW — Inconsistent handling

#### `rtx_sftpd`

- **File**: `internal/provider/resources/sftpd/resource.go:164, 173`
- **Affected attribute**: `hosts`
- **Issue**: `resource.go` sets `types.ListNull()` for empty hosts, but `model.go` `FromClient()` correctly creates an empty `types.ListValue()`. Inconsistency between code paths may cause drift depending on which path executes.

### Not Affected (safe usages)

The following use `types.ListNull()` appropriately for optional nested block lists where null means "block not present":

| Resource | File | Attribute | Why Safe |
|----------|------|-----------|----------|
| `rtx_access_list_mac` | access_list_mac/model.go:267,298,322 | Nested block attrs | Null = block absent |
| `rtx_access_list_extended` | access_list_extended/model.go:209 | `apply` block list | Null = no apply blocks |
| `rtx_ospf` | ospf/model.go:157,175,193 | `networks`, `areas`, `neighbors` | Null = block absent |
| `rtx_nat_masquerade` | nat_masquerade/model.go:126 | `static_entry` block list | Null = no entries |

---

## Suggested Fix: Centralized Conversion Functions

The root cause is duplicated conversion functions across 3 packages (`tunnel`, `ipsec_tunnel`, `class_map`). These should be consolidated into `fwhelpers` and fixed once:

```go
// fwhelpers/list_helpers.go

// IntSliceToList converts []int to types.List, preserving nil vs empty distinction.
func IntSliceToList(ints []int) types.List {
    if ints == nil {
        return types.ListNull(types.Int64Type)
    }
    elements := make([]attr.Value, len(ints))
    for i, v := range ints {
        elements[i] = types.Int64Value(int64(v))
    }
    list, _ := types.ListValue(types.Int64Type, elements)
    return list
}

// ListToIntSlice converts types.List to []int, preserving nil vs empty distinction.
func ListToIntSlice(list types.List) []int {
    if list.IsNull() || list.IsUnknown() {
        return nil
    }
    elements := list.Elements()
    result := make([]int, 0, len(elements))
    for _, elem := range elements {
        if int64Val, ok := elem.(types.Int64); ok && !int64Val.IsNull() && !int64Val.IsUnknown() {
            result = append(result, int(int64Val.ValueInt64()))
        }
    }
    return result
}

// StringSliceToList converts []string to types.List, preserving nil vs empty distinction.
func StringSliceToList(strs []string) types.List {
    if strs == nil {
        return types.ListNull(types.StringType)
    }
    elements := make([]attr.Value, len(strs))
    for i, v := range strs {
        elements[i] = types.StringValue(v)
    }
    list, _ := types.ListValue(types.StringType, elements)
    return list
}
```

Then replace all per-package copies with `fwhelpers.IntSliceToList()` / `fwhelpers.ListToIntSlice()`.

For `rtx_kron_schedule` and `rtx_dhcp_scope`, use `fwhelpers.StringSliceToList()` with the same nil-vs-empty logic.
