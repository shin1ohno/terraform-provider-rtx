# Bug Report: `rtx_snmp_server.enable_traps` Inconsistent Result After Apply

**Date**: 2026-05-08
**Observed in**: home-monitor project (`terraform apply` during setup PR #37 / Phase B-4-1)
**Affected resource**: `rtx_snmp_server`
**Pattern class**: Empty list vs null (recurrence of [Bug 2 in drift-bug-report-2026-02-08.md](drift-bug-report-2026-02-08.md))

## Summary

`terraform apply` aborted with:

```
Error: Provider produced inconsistent result after apply

When applying changes to rtx_snmp_server.hnd, provider
"provider[\"registry.terraform.io/shin1ohno/rtx\"]" produced an unexpected
new value: .enable_traps: was null, but now cty.ListValEmpty(cty.String).
```

The `enable_traps` attribute was `null` in plan (config does not set it) but the provider's `Read`/post-apply step produced an empty list `[]` instead — Terraform's framework rejects this as a contract violation.

This is the same class of bug fixed for `rtx_tunnel.secure_filter_in` in commit [`4030206`](https://github.com/shin1ohno/terraform-provider-rtx/commit/4030206) (2026-02-08, "tunnel: Fix inconsistent result for empty secure_filter"). `rtx_snmp_server` was missed in that sweep.

## Reproduction

home-monitor `rtx_snmp_server.hnd` resource (no `enable_traps` block in config). Any `terraform apply` that touches the resource — including a no-op `description` refresh — re-triggers the error every time.

```hcl
# home-monitor/rtx-snmp.tf (excerpt)
resource "rtx_snmp_server" "hnd" {
  sys_name     = "..."
  sys_location = "..."
  sys_contact  = "..."
  # enable_traps NOT specified — should remain null in state
  ...
}
```

Apply fails at `rtx_snmp_server.hnd: Modifying... [id=snmp]`. Other resources in the same plan complete successfully; only this resource trips on the post-apply consistency check.

## Root Cause

`internal/provider/resources/snmp_server/model.go:143-152`:

```go
// Convert enable_traps
if len(config.TrapEnable) > 0 {
    trapValues := make([]attr.Value, len(config.TrapEnable))
    for i, t := range config.TrapEnable {
        trapValues[i] = types.StringValue(t)
    }
    m.EnableTraps = types.ListValueMust(types.StringType, trapValues)
} else {
    m.EnableTraps = types.ListValueMust(types.StringType, []attr.Value{})  // ← writes empty list
}
```

When the router has no traps enabled (and the config did not set `enable_traps` at all), the `else` branch produces an empty `cty.ListValEmpty(cty.String)`. Terraform's framework expected `null` (the planned value, since the attribute was unset in config), so it raises the inconsistency.

The same `else`-branch pattern exists at `model.go:140` for `Hosts` — also a candidate to verify against the same scenario (e.g. `rtx_snmp_server` with no `host` block at all).

## Fix Shape

Mirror the `rtx_tunnel.secure_filter_in` fix from commit `4030206`:

1. **In `model.go:143-152`** — when `config.TrapEnable` is empty AND the schema attribute is null/unset in the plan, return `types.ListNull(types.StringType)` instead of `types.ListValueMust(types.StringType, []attr.Value{})`. The planned value of `null` is preserved.

   Approximate shape:
   ```go
   if len(config.TrapEnable) > 0 {
       // ... existing build path ...
       m.EnableTraps = types.ListValueMust(types.StringType, trapValues)
   } else {
       m.EnableTraps = types.ListNull(types.StringType)
   }
   ```

   The exact decision rule (always-null-when-empty vs preserve-explicit-empty-list-from-config) is the same one the tunnel fix made — match its convention.

2. **Apply the same audit to `Hosts`** at `model.go:140`. Empty hosts list with no `host` block in config will trip the same way once a deployment exercises that path.

3. **Update path** — if `Update()` reads the planned-value before calling `FromClient`, ensure the plan value (which may legitimately be `[]` if user wrote `enable_traps = []`) is preserved instead of being overwritten by the read-back. This is the second half of the tunnel fix — see commit `4030206`'s `Update()` changes for the exact pattern.

## Test Coverage

Add to `internal/provider/resources/snmp_server/`:

- `model_test.go`: case where `config.TrapEnable` is empty slice → expect `EnableTraps == types.ListNull(types.StringType)` (not `ListValueMust(... []attr.Value{})`)
- Acceptance test (`testdata/`): apply a `rtx_snmp_server` config with no `enable_traps` block → second `terraform plan` shows no diff (regression guard for this exact symptom)

The 2026-02-08 bug report's Bug 2 test additions can be used as a template.

## Verification

After fix, the home-monitor apply should pass cleanly:

```bash
cd ~/ManagedProjects/home-monitor
terraform apply -target=rtx_snmp_server.hnd
# Expected: Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
# Re-running terraform plan should show: No changes.
```

## Related

- Sibling bug 2026-02-08: `docs/dev/drift-bug-report-2026-02-08.md` — Bug 2 (empty list vs null, multiple resources)
- Sibling fix 2026-02-08: commit `4030206` — `rtx_tunnel` empty `secure_filter_*` handling
- Triggered by: home-monitor PR #37 (Phase B-4-1, `/host-registry/outputs/*` SSM tree publish). `rtx_snmp_server.hnd` was unrelated to that PR's intent — the resource was already drifted on `main`, so any apply that includes it surfaces this bug. Confirmed by running plan from `main` HEAD (without PR #37 changes): same `Plan: 0 to add, 17 to change, 0 to destroy` showing `rtx_snmp_server.hnd will be updated in-place`.
