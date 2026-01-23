# Design Document: Filter Number Parsing Fix

## Overview

This document describes the design for fixing the filter number parsing issue where numbers are incorrectly split or truncated when the RTX router output wraps at the terminal width boundary.

## Root Cause Analysis

### Problem Location

The issue is in `internal/rtx/parsers/interface_config.go`, specifically the `preprocessWrappedLines` function (lines 375-406).

### How the Bug Manifests

When the RTX router outputs configuration, it wraps long lines at approximately 80 characters. If a filter number happens to span this boundary, the current implementation incorrectly handles the wrap.

**Example of problematic output from RTX:**
```
ip lan2 secure filter in 200020 200021 200022 200023 200024 200025 200103 20010
0 200102 200104 200101 200105 200099
```

Here, `200100` is split across lines as `20010` and `0`.

**Current behavior of `preprocessWrappedLines`:**
```go
for i+1 < len(lines) && continuationPattern.MatchString(strings.TrimSpace(lines[i+1])) {
    i++
    line = line + " " + strings.TrimSpace(lines[i])  // Always adds a space!
}
```

**Result after joining:**
```
ip lan2 secure filter in ... 200103 20010 0 200102 200104 ...
```

**Result from `parseFilterList`:**
- `200103` ✓
- `20010` ✗ (truncated - should be `200100`)
- `0` ✗ (filtered out as <= 0)
- `200102` ✓

### Why This Causes the Observed Issues

1. **`200100` → `20010`**: The number is split as `20010` and `0`. With space insertion, `parseFilterList` sees them as separate numbers. `20010` is accepted, `0` is filtered out.

2. **`200027` → `2000`, `27`**: Similarly split at the line boundary. Both parts are > 0, so both are included as separate filter numbers.

3. **`101085` → `1`, `1085`**: The IPv6 dynamic filter number split in a way that produced two valid numbers.

## Proposed Solution

### Solution: Smart Line Joining

Modify `preprocessWrappedLines` to detect when a line ends with a digit and the next line starts with a digit, indicating a mid-number wrap. In this case, join without a space.

### Design Details

```go
func preprocessWrappedLines(input string) string {
    // ... existing setup code ...

    for i := 0; i < len(lines); i++ {
        line := lines[i]

        // Look ahead for continuation lines
        for i+1 < len(lines) && continuationPattern.MatchString(strings.TrimSpace(lines[i+1])) {
            i++
            nextLine := strings.TrimSpace(lines[i])

            // Check if we need to join without space (mid-number wrap)
            // Current line ends with digit AND next line starts with digit
            if endsWithDigit(line) && startsWithDigit(nextLine) {
                line = line + nextLine  // Join without space
            } else {
                line = line + " " + nextLine  // Join with space
            }
        }

        result = append(result, line)
    }

    return strings.Join(result, "\n")
}

// endsWithDigit returns true if the string ends with a digit (0-9)
func endsWithDigit(s string) bool {
    s = strings.TrimRight(s, " \t\r")
    if len(s) == 0 {
        return false
    }
    return s[len(s)-1] >= '0' && s[len(s)-1] <= '9'
}

// startsWithDigit returns true if the string starts with a digit (0-9)
func startsWithDigit(s string) bool {
    if len(s) == 0 {
        return false
    }
    return s[0] >= '0' && s[0] <= '9'
}
```

### Edge Cases to Handle

1. **Line ends with digit, next starts with `dynamic` keyword**: Should join with space
   - Example: `... 200020` + `dynamic 100080` → `... 200020 dynamic 100080`

2. **Line ends with partial number, next is just digits**: Should join without space
   - Example: `... 20010` + `0 200102` → `... 200100 200102`

3. **Line ends with complete number, next starts with another number**: Should join with space
   - Example: `... 200020` + `200021 200022` → `... 200020 200021 200022`

The key insight is: if the previous line ends in a digit AND the next line starts with a digit, the RTX likely wrapped mid-number. In all other cases (including "dynamic" keyword), use space.

### Why This Works

The RTX router always separates filter numbers with spaces. If we see:
- Line ending: `... 20010`
- Line starting: `0 200102 ...`

The only way `0` would be a standalone filter number is if it's filter ID 0, which is invalid (filters are 1-65535). Therefore, if we see digit-ends-digit-starts, it's safe to assume mid-number wrap.

## Alternatives Considered

### Alternative 1: Increase Terminal Width

Configure SSH session with larger terminal width (e.g., 200+ columns) to prevent wrapping.

**Rejected because:**
- Requires changes to SSH/session layer
- May not work for all RTX models
- Doesn't fix the underlying parsing fragility

### Alternative 2: Request RTX to Output Without Wrapping

Some devices support commands to disable line wrapping.

**Rejected because:**
- RTX routers may not support this
- Would require protocol-level changes
- Not universally reliable

### Alternative 3: Use a Different Parsing Approach

Parse the raw config character-by-character or use a more sophisticated state machine.

**Rejected because:**
- Over-engineering for this specific issue
- The simple fix (smart joining) handles all known cases
- Current regex-based approach is well-tested otherwise

## Testing Strategy

### Unit Tests

1. **Test mid-number wrap for 6-digit filter numbers:**
   ```go
   input := "ip lan2 secure filter in 200020 20010\n0 200102"
   // Should parse as: [200020, 200100, 200102]
   ```

2. **Test mid-number wrap with dynamic keyword:**
   ```go
   input := "ip lan2 secure filter out 200020 20002\n7 200099 dynamic 200080"
   // Should parse SecureFilterOut: [200020, 200027, 200099], DynamicFilterOut: [200080]
   ```

3. **Test IPv6 dynamic filter mid-number wrap:**
   ```go
   input := "ipv6 lan2 secure filter out 100020 dynamic 10108\n5 101086"
   // Should parse DynamicFilterOut: [101085, 101086]
   ```

4. **Test normal wrapping (no mid-number split):**
   ```go
   input := "ip lan2 secure filter in 200020 200021\n200022 200023"
   // Should parse as: [200020, 200021, 200022, 200023]
   ```

### Integration Tests

1. Round-trip test: Generate filter command → parse → compare
2. Test with actual RTX config samples that exhibit the issue

## Implementation Notes

### Files to Modify

1. `internal/rtx/parsers/interface_config.go`
   - Modify `preprocessWrappedLines` function
   - Add helper functions `endsWithDigit` and `startsWithDigit`

### Files for Testing

1. `internal/rtx/parsers/interface_config_test.go`
   - Add tests for mid-number wrap scenarios

## Risks and Mitigations

### Risk: False Positives

**Scenario:** A line ends with a digit that's not part of a filter number.

**Mitigation:** In the context of `secure filter in/out` commands, the only content is filter numbers and the `dynamic` keyword. False positives are unlikely.

### Risk: RTX Output Format Changes

**Scenario:** Future RTX firmware changes output format.

**Mitigation:** The fix is conservative - it only changes behavior when both conditions are met (digit-ends AND digit-starts). Other scenarios remain unchanged.

## Success Criteria

1. `terraform plan` shows no differences for correctly-configured resources
2. Filter number `200100` is read as `200100`, not `20010`
3. Filter number `200027` is read as `200027`, not split into `2000` and `27`
4. Filter number `101085` is read as `101085`, not split into `1` and `1085`
5. All existing tests continue to pass
6. Round-trip tests (parse → generate → parse) produce identical results
