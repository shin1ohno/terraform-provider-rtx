# Design Document: Import Fidelity Fix v3

## Technical Approach

This design addresses three parser issues identified in requirements.md. Each fix targets a specific parsing function with minimal code changes.

## REQ-1: Filter List Line-Wrap Handling

### Current Behavior

```
Input (RTX output):
  ip lan2 secure filter in 200020 200021 ... 20010
  0 200102 ...

After preprocessWrappedLines:
  ip lan2 secure filter in 200020 200021 ... 20010 0 200102 ...
                                            ^^^^^ ^
                                            Split number becomes two tokens
```

The `preprocessWrappedLines` function joins continuation lines with a space, breaking numbers that span lines.

### Proposed Solution

Modify `preprocessWrappedLines` to detect and handle split numbers:

```go
// In preprocessWrappedLines, when joining lines:
// If current line ends with digit AND next line starts with digit,
// join WITHOUT space to reconstruct the split number

for i+1 < len(lines) && continuationPattern.MatchString(strings.TrimSpace(lines[i+1])) {
    i++
    nextLine := strings.TrimSpace(lines[i])

    // Check if this is a split number (line ends with digit, next starts with digit)
    if endsWithDigit(line) && startsWithDigit(nextLine) {
        line = line + nextLine  // No space - reconstruct number
    } else {
        line = line + " " + nextLine  // Normal join with space
    }
}
```

### Helper Functions

```go
func endsWithDigit(s string) bool {
    if len(s) == 0 { return false }
    return s[len(s)-1] >= '0' && s[len(s)-1] <= '9'
}

func startsWithDigit(s string) bool {
    if len(s) == 0 { return false }
    return s[0] >= '0' && s[0] <= '9'
}
```

### Files Modified

- `internal/rtx/parsers/interface_config.go`

### Test Cases

```go
// Test: Number split across lines
input := `ip lan2 secure filter in 200020 20010
0 200102`
// Expected: filters = [200020, 200100, 200102]

// Test: Dynamic keyword with split numbers
input := `ip lan2 secure filter out 200050 200051 dynamic 20008
5 200086`
// Expected: dynamic_filters = [200085, 200086]
```

---

## REQ-2: DNS Service Recursive Support

### Current Behavior

```go
// dns.go line 83
dnsServicePattern := regexp.MustCompile(`^\s*dns\s+service\s+(on|off)\s*$`)
```

This pattern does not match `dns service recursive`, causing `ServiceOn = false` when the router uses recursive mode.

### Proposed Solution

Expand the regex pattern to include `recursive`:

```go
// Match: "dns service on", "dns service off", "dns service recursive"
dnsServicePattern := regexp.MustCompile(`^\s*dns\s+service\s+(on|off|recursive)\s*$`)

// In parsing logic:
if matches := dnsServicePattern.FindStringSubmatch(line); len(matches) >= 2 {
    config.ServiceOn = (matches[1] == "on" || matches[1] == "recursive")
    continue
}
```

### Builder Update

Update `BuildDNSServiceCommand` to output `recursive` when enabled (preferred form per REQ-2 AC-4):

```go
func BuildDNSServiceCommand(enable bool) string {
    if enable {
        return "dns service recursive"  // Preferred form
    }
    return "dns service off"
}
```

### Files Modified

- `internal/rtx/parsers/dns.go`

### Test Cases

```go
// Test: dns service recursive
input := "dns service recursive"
// Expected: ServiceOn = true

// Test: dns service on (backward compat)
input := "dns service on"
// Expected: ServiceOn = true
```

---

## REQ-3: DNS Server Select Multi-Server EDNS Parsing

### Current Behavior

```
Config: dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .

Current parsing:
  Phase 1: servers = [2606:4700:4700::1111] (stops at edns=on)
  Phase 2: edns = true
  Phase 3: record_type not found (next token is IP)
  Phase 4: query_pattern = "2606:4700:4700::1001" (WRONG!)
```

The parser stops collecting servers at the first `edns=on`, treating subsequent servers as query patterns.

### Proposed Solution

Modify server parsing to handle interleaved `edns=on` options:

```go
func parseDNSServerSelectFields(id int, rest string) *DNSServerSelect {
    fields := strings.Fields(rest)
    // ...

    // Phase 1: Parse servers with interleaved edns=on
    // Format: <server1> [edns=on] [<server2> [edns=on]] ...
    const maxServers = 2
    for i < len(fields) && len(sel.Servers) < maxServers {
        if isValidIPForDNS(fields[i]) {
            sel.Servers = append(sel.Servers, fields[i])
            i++
            // Check for edns=on after this server
            if i < len(fields) && fields[i] == "edns=on" {
                sel.EDNS = true
                i++
            }
        } else {
            break  // Not an IP, move to next phase
        }
    }

    // Phase 2: Skip if we see another edns=on (edge case)
    // Already handled in Phase 1

    // Phase 3: Check for record_type
    if i < len(fields) && validRecordTypes[fields[i]] && fields[i] != "." {
        sel.RecordType = fields[i]
        i++
    }

    // Phase 4: Parse query_pattern (required)
    if i < len(fields) {
        sel.QueryPattern = fields[i]
        i++
    }

    // ... rest unchanged
}
```

### Key Change

Move `edns=on` detection into the server loop to handle:
- `<server1> edns=on <server2> edns=on` (interleaved)
- `<server1> <server2> edns=on` (trailing, existing format)
- `<server1> edns=on` (single server with EDNS)

### Files Modified

- `internal/rtx/parsers/dns.go`

### Test Cases

```go
// Test: Multi-server with interleaved EDNS
input := "dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa ."
// Expected:
//   servers = ["2606:4700:4700::1111", "2606:4700:4700::1001"]
//   edns = true
//   record_type = "aaaa"
//   query_pattern = "."

// Test: Single server with EDNS
input := "dns server select 1 1.1.1.1 edns=on ."
// Expected:
//   servers = ["1.1.1.1"]
//   edns = true
//   query_pattern = "."

// Test: Two servers, trailing EDNS (backward compat)
input := "dns server select 2 8.8.8.8 8.8.4.4 edns=on ."
// Expected:
//   servers = ["8.8.8.8", "8.8.4.4"]
//   edns = true
//   query_pattern = "."
```

---

## Implementation Order

1. **REQ-2** (DNS service recursive) - Simplest change, isolated regex update
2. **REQ-3** (DNS server select EDNS) - Moderate complexity, self-contained in one function
3. **REQ-1** (Line-wrap handling) - Most complex, affects filter parsing globally

This order minimizes risk by starting with isolated changes before modifying shared preprocessing logic.

## Backward Compatibility

All changes are backward compatible:
- REQ-2: Existing `on`/`off` values continue to work
- REQ-3: Existing single-server and trailing-EDNS formats continue to work
- REQ-1: Standard (non-wrapped) lines are processed identically

## Testing Strategy

1. Unit tests for each parser fix
2. Round-trip tests (parse → build → parse) for DNS commands
3. Integration test with actual RTX output samples (line-wrapped)
4. Regression tests for existing functionality
