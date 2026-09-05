# TODO

- **ipv6_interface: destroy leaves `ipv6 <iface> address` lines on the device**
  Observed 2026-08-03 on RTX1210 Rev.14.01.42 (home-monitor #117/#119): destroying an
  `rtx_ipv6_interface` removed the rtadv / dhcp-service lines but left
  `ipv6 lan1 address ra-prefix@lan2::2/64` and `ipv6 bridge1 address ra-prefix@lan2::1/64`
  in the running config. Terraform state dropped the resource, so the leftovers became
  invisible drift that needed a manual admin-session cleanup.
  First step: acceptance test that reads back `show config` after destroy, then make the
  delete path issue `no ipv6 <iface> address <value>` for every configured address block.

- **access_list_ipv6: READ loses the `dynamic ...` suffix of an out-direction apply line**
  Observed 2026-08-12 on RTX1210: the live device has
  `ipv6 lan2 secure filter out 100 105 110 115 120 125 130 135 140 145 150 155 160 165 dynamic 1 6 11 16 21 26 31 36`
  but every refresh reads `dynamic_sequences` back as empty, so terraform plans a
  perpetual in-place update re-adding the sequences (phantom drift, documented in
  home-monitor rtx-hnd.tf). The line is long enough to console-wrap in `show config`
  over SSH — likely the same wrap-handling class as #13/#14, not applied to
  interface secure-filter apply parsing (in-direction dynamic refs on the shorter
  v4 line read back fine).
  First step: unit test ParseInterfaceSecureFilter (or its ipv6 equivalent) against
  a console-wrapped out line with a dynamic suffix; fix with the shared de-wrap
  helper from parsers/.

- **ipv6_interface: schema accepts `rtadv` / `dhcpv6_service` on bridge interfaces the firmware rejects**
  `ipv6 bridge1 rtadv send 1 o_flag=on m_flag=off` fails on-device with コマンド名エラー —
  per the official IPv6 command reference, `rtadv send` / `dhcp service` take LAN
  interfaces only, on ALL current models (rt-common applies to RTX3510/1300/1220/1210/840/830/vRX).
  The shared interface regex `^(lan|bridge|pp|tunnel)\d+$` lets bridgeN through validation,
  so the failure only surfaces mid-apply (caused the home-monitor #118 revert).
  First step: restrict rtadv / dhcpv6_service to lanN (plus pp where documented) via schema
  validators, with a test asserting bridgeN is rejected at plan time.

- **kron_schedule: `policy_list` is accepted but never written to the device**
  Found while fixing the missing `*` execution-context token (0.16.4). The schema offers
  `policy_list` as an alternative to `command_lines`, but `CreateSchedule` only iterates
  `Commands` — a schedule configured with `policy_list` alone issues no command at all,
  and (now that the read path works) fails with "not found after create". RTX has no
  kron policy list; `CreateKronPolicy` is a logging no-op by design.
  First step: reject `policy_list` at plan time in `ValidateConfig` with a message naming
  `command_lines` as the replacement, or drop the attribute in the next minor.

- **kron_schedule: a one-time `date` schedule can never satisfy the consistency check**
  `recurring` carries `booldefault.StaticBool(true)`, so a config with `date = "2025/01/15"`
  and no explicit `recurring` plans `recurring=true` while the read path correctly returns
  false for a YYYY/MM/DD date — a guaranteed "Provider produced inconsistent result after
  apply". The daily and startup shapes are unaffected. Do NOT fix by dropping the default:
  that puts `(known after apply)` on `recurring` in every existing user's plan.
  First step: implement `ModifyPlan` to set `recurring=false` when config leaves it unset and
  the schedule is `on_startup` or has a one-time date (`parsers.IsScheduleOneTimeDate`), plus
  a `ValidateConfig` error for an explicit `recurring = true` on those shapes.

- **`containsError` does not recognise the RTX's syntax-rejection messages generically**
  `internal/client/dhcp_service.go:112` decides success by substring match. `エラー:` is in
  the list, so the common rejections are caught, but a message that does not start with it
  would let a rejected write return nil — the failure then surfaces much later as a
  read-back mismatch. 0.16.4 papers over this for kron_schedule specifically
  (`missingAfterWrite` names it at the resource layer); every other service still relies on
  the substring list.
  First step: capture the actual rejection strings from a live RTX for a handful of
  malformed commands (`schedule at ?`-style probing via `scripts/rtx_shell_probe.go`), then
  widen the list with a test per string. Note this changes failure behaviour for every
  service that calls `runCommand`, so it wants its own PR.

- **`schedule_spec_test.go` is stale relative to `specs/schedule/config.yaml`**
  The checked-in generated file predates the spec rewrite that introduced the `*` execution
  context: 22 cases in the file vs 21+1 in the spec, no `*` anywhere, and a whole
  `schedule pp N <day> <time> connect` family the current spec does not define. Nothing
  catches it — CI's Generate Check runs `go generate ./...`, which does not invoke specgen.
  First step: `go run ./tools/specgen -spec specs/schedule/config.yaml -test-only`, inspect
  the ~22-case diff, and commit it on its own. Consider adding specgen to `go generate` so
  the drift cannot recur silently.

- **kron_schedule cannot express `+TIMER` schedules or a non-global execution context**
  `schedule at <id> +<seconds> * <command>` has no attribute, so the parser skips those
  lines entirely and such a schedule reads as absent. `pp <n>` / `tunnel <n>` / `switch <sw>`
  contexts are parsed and carried on `parsers.Schedule.Context` (the resource warns when it
  sees one), but there is no way to configure them, so importing such a schedule and applying
  rewrites it into the global `*` context.
  First step: decide whether these belong on `rtx_kron_schedule` as `timer` / `context`
  attributes or on a separate resource, then add the builder + parser round-trip cases from
  `specs/schedule/config.yaml:279-300`.
