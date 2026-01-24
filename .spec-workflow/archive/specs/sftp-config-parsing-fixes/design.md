# Design Document: SFTP Config Parsing Fixes

## Overview

This design addresses parser issues discovered when reading RTX configuration via SFTP. The raw config is correctly retrieved, but several parsers fail to extract fields correctly. This document specifies the code changes needed to fix each issue.

## Steering Document Alignment

### Technical Standards (tech.md)
- Go 1.21+ with testify for testing
- Parser functions in `internal/rtx/parsers/`
- Table-driven tests for comprehensive coverage
- Real RTX config formats in test fixtures

### Project Structure (structure.md)
- Parser source: `internal/rtx/parsers/*.go`
- Parser tests: `internal/rtx/parsers/*_test.go`
- Config extraction: `internal/rtx/parsers/config_file.go`

## Code Reuse Analysis

### Existing Components to Leverage

- **ConfigFileParser**: Main parser that handles context-aware command extraction
- **ParseL2TPConfig**: L2TP tunnel parser in `l2tp.go`
- **ParseL2TPServiceConfig**: L2TP service parser in `l2tp_service.go`
- **ParseSystemConfig**: System config parser in `system.go`
- **ParseBindings**: DHCP binding parser in `dhcp_binding.go`
- **Existing test patterns**: `config_file_test.go` holistic test approach

### Integration Points

- **ConfigFileParser.Parse()**: Entry point for config parsing
- **ConfigFileParser.ExtractXxx()**: Resource-specific extraction methods
- **Context handling**: Tunnel context, PP context command grouping

## Architecture

### Parser Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│ SFTP Config Read                                            │
│ /system/config0 → raw string                                │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ ConfigFileParser.Parse(rawConfig)                           │
│ - Split into lines                                          │
│ - Detect context boundaries (tunnel select, pp select)      │
│ - Group commands by context                                 │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ ExtractXxx() Methods                                        │
│ - Filter commands by prefix                                 │
│ - Build raw config for specific parser                      │
│ - Call resource-specific parser                             │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Resource Parsers (l2tp.go, system.go, etc.)                 │
│ - Parse raw config lines                                    │
│ - Extract field values                                      │
│ - Return structured config                                  │
└─────────────────────────────────────────────────────────────┘
```

## Components and Fixes

### FIX-1: L2TP Tunnel Context Parsing

**File:** `internal/rtx/parsers/l2tp.go`

**Problem:** `l2tp always-on on`, `l2tp hostname`, `l2tp keepalive use on` commands in tunnel context are not parsed.

**Root Cause Analysis:**
The `ParseL2TPConfig` function receives raw config built from context commands. Need to verify:
1. Commands are correctly collected from tunnel context
2. Parser correctly handles `l2tp always-on on` format
3. Parser correctly handles `l2tp hostname <name>` format
4. Parser correctly handles `l2tp keepalive use on <interval> <retry>` format

**Solution:**
```go
// In ParseL2TPConfig, ensure these patterns are handled:

// Pattern: l2tp always-on on
if strings.HasPrefix(line, "l2tp always-on ") {
    config.AlwaysOn = strings.TrimPrefix(line, "l2tp always-on ") == "on"
}

// Pattern: l2tp hostname <name>
if strings.HasPrefix(line, "l2tp hostname ") {
    config.Name = strings.TrimPrefix(line, "l2tp hostname ")
}

// Pattern: l2tp keepalive use on <interval> <retry>
if strings.HasPrefix(line, "l2tp keepalive use ") {
    parts := strings.Fields(line)
    if len(parts) >= 4 && parts[3] == "on" {
        config.KeepaliveEnabled = true
        // Optionally parse interval and retry
    }
}
```

### FIX-2: L2TPv2 Mode Detection

**File:** `internal/rtx/parsers/l2tp.go`

**Problem:** When `tunnel encapsulation l2tp` is detected, `Mode` stays as default `"l2vpn"` instead of `"lns"`.

**Current Code (lines ~203-213):**
```go
if strings.HasPrefix(line, "tunnel encapsulation ") {
    encap := strings.TrimPrefix(line, "tunnel encapsulation ")
    if encap == "l2tp" {
        config.Version = "l2tp"
        // Mode is NOT set here - BUG
    } else if encap == "l2tpv3" {
        config.Version = "l2tpv3"
    }
}
```

**Solution:**
```go
if strings.HasPrefix(line, "tunnel encapsulation ") {
    encap := strings.TrimPrefix(line, "tunnel encapsulation ")
    if encap == "l2tp" {
        config.Version = "l2tp"
        config.Mode = "lns"  // L2TPv2 is always LNS mode
    } else if encap == "l2tpv3" {
        config.Version = "l2tpv3"
        config.Mode = "l2vpn"  // L2TPv3 is L2VPN mode
    }
}
```

### FIX-3: L2TP Service Protocol Parsing

**File:** `internal/rtx/parsers/l2tp_service.go`

**Problem:** `l2tp service on l2tpv3 l2tp` does not extract protocols.

**RTX Command Format:**
```
l2tp service on [<protocol1> [<protocol2>]]
```

**Solution:**
```go
func ParseL2TPServiceConfig(rawConfig string) (*L2TPServiceConfig, error) {
    config := &L2TPServiceConfig{}

    for _, line := range strings.Split(rawConfig, "\n") {
        line = strings.TrimSpace(line)

        // Pattern: l2tp service on [protocol1] [protocol2]
        if strings.HasPrefix(line, "l2tp service ") {
            parts := strings.Fields(line)
            // parts[0] = "l2tp", parts[1] = "service", parts[2] = "on"/"off"
            if len(parts) >= 3 {
                config.Enabled = parts[2] == "on"
                // Extract protocols after "on"
                if config.Enabled && len(parts) > 3 {
                    config.Protocols = parts[3:]
                }
            }
        }
    }

    return config, nil
}
```

### FIX-4: System Config Extraction

**File:** `internal/rtx/parsers/config_file.go` and `internal/rtx/parsers/system.go`

**Problem:** System fields (timezone, console, packet-buffer, statistics) are all null.

**Root Cause Analysis:**
1. `ExtractSystem` in `config_file.go` filters global commands
2. Commands may not be reaching `ParseSystemConfig`
3. Or `ParseSystemConfig` has parsing issues

**Debug Steps:**
1. Add logging to `ExtractSystem` to see what commands are collected
2. Verify global context commands include system lines
3. Check `ParseSystemConfig` handles all formats

**Expected Commands in Global Context:**
```
timezone +09:00
console character ja.utf8
console lines infinity
console prompt "[RTX1210] "
system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=10000 max-free=4950
system packet-buffer large max-buffer=20000 max-free=5600
statistics traffic on
statistics nat on
```

**Solution (if parsing issue):**
```go
// In ParseSystemConfig, ensure these patterns are handled:

// Pattern: timezone <offset>
if strings.HasPrefix(line, "timezone ") {
    config.Timezone = strings.TrimPrefix(line, "timezone ")
}

// Pattern: console character <charset>
if strings.HasPrefix(line, "console character ") {
    config.Console.Character = strings.TrimPrefix(line, "console character ")
}

// Pattern: console lines <value>
if strings.HasPrefix(line, "console lines ") {
    config.Console.Lines = strings.TrimPrefix(line, "console lines ")
}

// Pattern: console prompt "<prompt>"
if strings.HasPrefix(line, "console prompt ") {
    prompt := strings.TrimPrefix(line, "console prompt ")
    config.Console.Prompt = strings.Trim(prompt, "\"")
}

// Pattern: system packet-buffer <size> max-buffer=<n> max-free=<n>
if strings.HasPrefix(line, "system packet-buffer ") {
    pb := parsePacketBuffer(line)
    config.PacketBuffers = append(config.PacketBuffers, pb)
}

// Pattern: statistics traffic on/off
if strings.HasPrefix(line, "statistics traffic ") {
    config.Statistics.Traffic = strings.TrimPrefix(line, "statistics traffic ") == "on"
}

// Pattern: statistics nat on/off
if strings.HasPrefix(line, "statistics nat ") {
    config.Statistics.NAT = strings.TrimPrefix(line, "statistics nat ") == "on"
}
```

### FIX-5: DHCP Binding Scope ID

**File:** `internal/rtx/parsers/config_file.go` and `internal/rtx/parsers/dhcp_binding.go`

**Problem:** `ExtractDHCPBindings` passes `scopeID=0` to `ParseBindings`.

**Current Code (lines ~920-923):**
```go
func (p *ParsedConfig) ExtractDHCPBindings() []*DHCPBindingConfig {
    // ...
    return parser.ParseBindings(raw, 0)  // BUG: hardcoded scopeID=0
}
```

**RTX Command Format:**
```
dhcp scope bind <scope_id> <ip_address> <mac_address>
```

**Solution:**
```go
// Option 1: Parse scope ID from each line in ParseBindings
func ParseBindings(rawConfig string, defaultScopeID int) []*DHCPBindingConfig {
    var bindings []*DHCPBindingConfig

    for _, line := range strings.Split(rawConfig, "\n") {
        line = strings.TrimSpace(line)
        if !strings.HasPrefix(line, "dhcp scope bind ") {
            continue
        }

        // Parse: dhcp scope bind <scope_id> <ip> <mac...>
        parts := strings.Fields(line)
        if len(parts) < 5 {
            continue
        }

        scopeID, _ := strconv.Atoi(parts[3])
        ip := parts[4]
        mac := strings.Join(parts[5:], " ")

        binding := &DHCPBindingConfig{
            ScopeID:    scopeID,  // Extract from line, not parameter
            IPAddress:  ip,
            MACAddress: mac,
        }
        bindings = append(bindings, binding)
    }

    return bindings
}
```

### FIX-6: Update main.tf DNS Servers Order

**File:** `examples/import/main.tf`

**Change:**
```hcl
# Before
dns_servers = ["1.0.0.1", "1.1.1.1"]

# After
dns_servers = ["1.1.1.1", "1.0.0.1"]
```

This is not a parser fix but a configuration correction to match the actual RTX config.

## Testing Strategy

### Unit Testing (TDD Approach)

**Phase 1: Write Holistic Tests (Red)**
- Create `config_file_sftp_test.go`
- Use actual RTX config as test input
- Tests will fail initially (expected)

**Phase 2: Fix Implementation (Green)**
- Apply fixes FIX-1 through FIX-5
- Run tests after each fix
- Continue until all tests pass

**Phase 3: Verify and Refactor**
- Run full test suite
- Check for regressions
- Refactor if needed while keeping tests green

### Test File Structure

```go
// config_file_sftp_test.go

const sftpTestConfig = `
#
# Admin
#
login password TEST_LOGIN_PASS
...
`

func TestConfigFileParser_SFTPConfig(t *testing.T) {
    parser := NewConfigFileParser()
    result, err := parser.Parse(sftpTestConfig)
    require.NoError(t, err)

    t.Run("L2TP Tunnel 1", func(t *testing.T) {
        tunnels := result.ExtractL2TPTunnels()
        tunnel1 := findTunnel(tunnels, 1)
        require.NotNil(t, tunnel1)

        assert.True(t, tunnel1.AlwaysOn)
        assert.True(t, tunnel1.KeepaliveEnabled)
        assert.Equal(t, "test-RTX1210", tunnel1.Name)
        assert.Equal(t, "l2tpv3", tunnel1.Version)
        assert.Equal(t, "l2vpn", tunnel1.Mode)
    })

    t.Run("L2TP Tunnel 2", func(t *testing.T) {
        tunnels := result.ExtractL2TPTunnels()
        tunnel2 := findTunnel(tunnels, 2)
        require.NotNil(t, tunnel2)

        assert.Equal(t, "l2tp", tunnel2.Version)
        assert.Equal(t, "lns", tunnel2.Mode)
    })

    t.Run("L2TP Service", func(t *testing.T) {
        service := result.ExtractL2TPService()
        require.NotNil(t, service)

        assert.True(t, service.Enabled)
        assert.Equal(t, []string{"l2tpv3", "l2tp"}, service.Protocols)
    })

    t.Run("System Config", func(t *testing.T) {
        system := result.ExtractSystem()
        require.NotNil(t, system)

        assert.Equal(t, "+09:00", system.Timezone)
        assert.Equal(t, "ja.utf8", system.Console.Character)
        assert.Equal(t, "infinity", system.Console.Lines)
        assert.Equal(t, "[RTX1210] ", system.Console.Prompt)
        assert.Len(t, system.PacketBuffers, 3)
        assert.True(t, system.Statistics.Traffic)
        assert.True(t, system.Statistics.NAT)
    })

    t.Run("DHCP Bindings", func(t *testing.T) {
        bindings := result.ExtractDHCPBindings()
        assert.Len(t, bindings, 5)

        for _, b := range bindings {
            assert.Equal(t, 1, b.ScopeID)
        }

        ips := extractIPs(bindings)
        assert.Contains(t, ips, "192.168.1.20")
        assert.Contains(t, ips, "192.168.1.21")
        assert.Contains(t, ips, "192.168.1.22")
        assert.Contains(t, ips, "192.168.1.28")
        assert.Contains(t, ips, "192.168.1.29")
    })
}
```

## Error Handling

### Error Scenarios

1. **Malformed config line**
   - **Handling:** Skip line, log warning
   - **Test:** Include malformed lines in test config

2. **Missing required field**
   - **Handling:** Use default value, log warning
   - **Test:** Test with minimal config

3. **Invalid numeric value**
   - **Handling:** Use default, log error
   - **Test:** Test with invalid numbers

## RTX Command Reference

### L2TP Commands

| Command | Parser | Field |
|---------|--------|-------|
| `l2tp always-on on` | l2tp.go | AlwaysOn |
| `l2tp hostname <name>` | l2tp.go | Name |
| `l2tp keepalive use on <int> <retry>` | l2tp.go | KeepaliveEnabled |
| `tunnel encapsulation l2tp` | l2tp.go | Version="l2tp", Mode="lns" |
| `tunnel encapsulation l2tpv3` | l2tp.go | Version="l2tpv3", Mode="l2vpn" |
| `l2tp service on [proto...]` | l2tp_service.go | Enabled, Protocols |

### System Commands

| Command | Parser | Field |
|---------|--------|-------|
| `timezone <offset>` | system.go | Timezone |
| `console character <charset>` | system.go | Console.Character |
| `console lines <value>` | system.go | Console.Lines |
| `console prompt "<prompt>"` | system.go | Console.Prompt |
| `system packet-buffer <size> max-buffer=<n> max-free=<n>` | system.go | PacketBuffers |
| `statistics traffic on` | system.go | Statistics.Traffic |
| `statistics nat on` | system.go | Statistics.NAT |

### DHCP Commands

| Command | Parser | Field |
|---------|--------|-------|
| `dhcp scope bind <id> <ip> <mac>` | dhcp_binding.go | ScopeID, IPAddress, MACAddress |
