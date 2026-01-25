# Master Design Document: SFTP-Based Configuration Reading

## Overview

This document describes the architecture and implementation of the SFTP-based configuration reading feature for the RTX Terraform provider. When enabled, the provider downloads the router's complete configuration file via SFTP, caches it in memory, and parses it to serve resource read operations.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Provider Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐ │
│  │ provider.go │  │ resource_*  │  │ data_source_*               │ │
│  │             │  │ Read()      │  │ Read()                      │ │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬────────────────┘ │
│         │                │                       │                  │
│         │     ┌──────────▼───────────────────────▼──────────┐      │
│         │     │              apiClient                       │      │
│         │     │  client.SFTPEnabled()                       │      │
│         │     │  client.GetCachedConfig(ctx)                │      │
│         └─────┤                                              │      │
│               └──────────────────┬───────────────────────────┘      │
└──────────────────────────────────┼──────────────────────────────────┘
                                   │
┌──────────────────────────────────▼──────────────────────────────────┐
│                          Client Layer                                │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                         rtxClient                             │  │
│  │  ┌─────────────────┐  ┌───────────────┐  ┌────────────────┐ │  │
│  │  │ configCache     │  │ sftpEnabled   │  │ sftpConfigPath │ │  │
│  │  │ *ConfigCache    │  │ bool          │  │ string         │ │  │
│  │  └────────┬────────┘  └───────────────┘  └────────────────┘ │  │
│  │           │                                                   │  │
│  │  ┌────────▼────────────────────────────────────────────────┐ │  │
│  │  │ GetCachedConfig(ctx) (*ParsedConfig, error)             │ │  │
│  │  │  1. Check cache validity                                 │ │  │
│  │  │  2. If valid, return cached                             │ │  │
│  │  │  3. If invalid, download via SFTP                       │ │  │
│  │  │  4. Parse config                                        │ │  │
│  │  │  5. Store in cache                                      │ │  │
│  │  │  6. Return parsed                                       │ │  │
│  │  └─────────────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌───────────────────┐  ┌───────────────────┐  ┌─────────────────┐ │
│  │ ConfigCache       │  │ ConfigPathResolver│  │ SFTPClient      │ │
│  │ config_cache.go   │  │ config_path_*.go  │  │ sftp_client.go  │ │
│  └───────────────────┘  └───────────────────┘  └─────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
                                   │
┌──────────────────────────────────▼──────────────────────────────────┐
│                          Parser Layer                                │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ ConfigFileParser (internal/rtx/parsers/config_file.go)       │  │
│  │                                                               │  │
│  │  Parse(content) -> *ParsedConfig                             │  │
│  │                                                               │  │
│  │  ┌─────────────────────────────────────────────────────────┐ │  │
│  │  │ ParsedConfig                                            │ │  │
│  │  │  - Raw string                                           │ │  │
│  │  │  - Commands []ParsedCommand                             │ │  │
│  │  │  - Contexts []ParseContext                              │ │  │
│  │  │                                                          │ │  │
│  │  │  ExtractStaticRoutes() []StaticRoute                    │ │  │
│  │  │  ExtractDHCPScopes() []DHCPScope                        │ │  │
│  │  │  ExtractNATMasquerade() []NATMasquerade                 │ │  │
│  │  │  ExtractIPFilters() []IPFilter                          │ │  │
│  │  │  ExtractPasswords() *ExtractedPasswords                 │ │  │
│  │  └─────────────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
```

## Components

### Component 1: SFTPClient

**Location:** `internal/client/sftp_client.go`

**Purpose:** Download configuration file from RTX router via SFTP

**Interface:**
```go
type SFTPClient interface {
    Download(ctx context.Context, path string) ([]byte, error)
    Close() error
}
```

**Implementation:**
```go
type sftpClientImpl struct {
    sshClient  sshClientInterface
    sftpClient sftpClientInterface
    mu         sync.Mutex
    closed     bool
}

func NewSFTPClient(config *Config) (SFTPClient, error)
```

**Dependencies:**
- `golang.org/x/crypto/ssh` - SSH connection
- `github.com/pkg/sftp` - SFTP protocol

**Error Handling:**
- `ErrSFTPClosed` - Operation on closed client
- Connection errors - Propagated with context
- File not found - Clear error message with path

### Component 2: ConfigCache

**Location:** `internal/client/config_cache.go`

**Purpose:** Thread-safe in-memory cache for downloaded configuration

**Interface:**
```go
type ConfigCache struct {
    mu         sync.RWMutex
    content    string
    parsed     *parsers.ParsedConfig
    validUntil time.Time
    dirty      bool
}

func NewConfigCache() *ConfigCache
func (c *ConfigCache) Get() (*parsers.ParsedConfig, bool)
func (c *ConfigCache) GetRaw() (string, bool)
func (c *ConfigCache) Set(content string, parsed *parsers.ParsedConfig)
func (c *ConfigCache) SetWithTTL(content string, parsed *parsers.ParsedConfig, ttl time.Duration)
func (c *ConfigCache) Invalidate()
func (c *ConfigCache) MarkDirty()
func (c *ConfigCache) ClearDirty()
func (c *ConfigCache) IsDirty() bool
func (c *ConfigCache) IsValid() bool
```

**Constants:**
- `DefaultCacheTTL` - Default time-to-live for cached data

**Thread Safety:**
- Uses `sync.RWMutex` for concurrent read access
- Write operations acquire exclusive lock

### Component 3: ConfigPathResolver

**Location:** `internal/client/config_path_resolver.go`

**Purpose:** Determine correct config file path by querying router's startup config

**Interface:**
```go
type ConfigPathResolver struct {
    executor Executor
}

func NewConfigPathResolver(executor Executor) *ConfigPathResolver
func (r *ConfigPathResolver) Resolve(ctx context.Context) (string, error)
```

**Logic:**
1. Execute `show environment` via SSH
2. Parse output to find startup config line:
   - Japanese: `デフォルト設定ファイル: config{N}`
   - English: `Default config file: config{N}`
3. Return SFTP path: `/system/config{N}`

**Constants:**
- `DefaultConfigPath = "/system/config0"` - Fallback path

### Component 4: ConfigFileParser

**Location:** `internal/rtx/parsers/config_file.go`

**Purpose:** Parse full config.txt into structured data

**Types:**
```go
type ContextType int
const (
    ContextGlobal ContextType = iota
    ContextTunnel
    ContextPP
    ContextIPsecTunnel
)

type ParseContext struct {
    Type ContextType
    ID   int
    Name string
}

type ParsedCommand struct {
    Line        string
    Context     *ParseContext
    LineNumber  int
    IndentLevel int
}

type ParsedConfig struct {
    Raw          string
    LineCount    int
    CommandCount int
    Contexts     []ParseContext
    Commands     []ParsedCommand
}
```

**Interface:**
```go
type ConfigFileParser struct {
    tunnelSelectPattern       *regexp.Regexp
    ppSelectPattern           *regexp.Regexp
    ppSelectAnonymousPattern  *regexp.Regexp
    ipsecTunnelPattern        *regexp.Regexp
}

func NewConfigFileParser() *ConfigFileParser
func (p *ConfigFileParser) Parse(content string) (*ParsedConfig, error)
```

**Extract Methods:**
```go
func (c *ParsedConfig) ExtractStaticRoutes() []StaticRoute
func (c *ParsedConfig) ExtractDHCPScopes() []DHCPScope
func (c *ParsedConfig) ExtractNATMasquerade() []NATMasquerade
func (c *ParsedConfig) ExtractIPFilters() []IPFilter
func (c *ParsedConfig) ExtractIPFiltersDynamic() []IPFilterDynamic
func (c *ParsedConfig) ExtractPasswords() *ExtractedPasswords
```

**Context-Aware Parsing:**

RTX config uses hierarchical contexts:
```
tunnel select 1          # Enter tunnel 1 context
 tunnel encapsulation l2tpv3
 ipsec tunnel 101        # Enter ipsec tunnel subcontext
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
 tunnel enable 1
tunnel select 2          # Exit tunnel 1, enter tunnel 2
```

Parser maintains context stack to associate commands with correct parent.

### Component 5: Provider Configuration

**Location:** `internal/provider/provider.go`

**Schema Fields:**
```go
"use_sftp": {
    Type:        schema.TypeBool,
    Optional:    true,
    Default:     false,
    DefaultFunc: schema.EnvDefaultFunc("RTX_USE_SFTP", false),
    Description: "Use SFTP-based bulk configuration reading.",
},
"sftp_config_path": {
    Type:        schema.TypeString,
    Optional:    true,
    Default:     "",
    DefaultFunc: schema.EnvDefaultFunc("RTX_SFTP_CONFIG_PATH", ""),
    Description: "Path to config file. Empty means auto-detect.",
},
```

### Component 6: Client Config Extension

**Location:** `internal/client/interfaces.go`

**Config Fields:**
```go
type Config struct {
    // ... existing fields ...
    SFTPEnabled    bool
    SFTPConfigPath string
}
```

### Component 7: Client Integration

**Location:** `internal/client/client.go`

**Methods:**
```go
func (c *rtxClient) SFTPEnabled() bool
func (c *rtxClient) GetCachedConfig(ctx context.Context) (*parsers.ParsedConfig, error)
func (c *rtxClient) InvalidateCache()
func (c *rtxClient) MarkCacheDirty()
```

**GetCachedConfig Flow:**
1. Check if SFTP is enabled
2. Check if cache is valid and not dirty
3. If cache hit, return cached ParsedConfig
4. If cache miss:
   a. Resolve config path (if not specified)
   b. Create SFTP client
   c. Download config file
   d. Parse config
   e. Store in cache
   f. Return parsed config
5. On any SFTP error: log warning, return nil (caller falls back to SSH)

## Data Flow

### Resource Read with SFTP Cache

```
Resource.Read()
    │
    ▼
apiClient.client.SFTPEnabled()
    │
    ├─── false ──► Use SSH CLI (existing behavior)
    │
    └─── true ───►
                  │
                  ▼
          apiClient.client.GetCachedConfig(ctx)
                  │
                  ├─── Cache Hit ──► Return ParsedConfig
                  │
                  └─── Cache Miss ──►
                                     │
                                     ▼
                              ConfigPathResolver.Resolve(ctx)
                                     │
                                     ▼
                              SFTPClient.Download(ctx, path)
                                     │
                                     ▼
                              ConfigFileParser.Parse(content)
                                     │
                                     ▼
                              ConfigCache.Set(content, parsed)
                                     │
                                     ▼
                              Return ParsedConfig
    │
    ▼
ParsedConfig.ExtractXXX()
    │
    ├─── Found ──► Set resource attributes
    │
    └─── Not Found ──► Fall back to SSH CLI
```

## Error Handling

### Fallback Strategy

```
┌─────────────────┐
│ SFTP Operation  │
└────────┬────────┘
         │
         ▼
    ┌────────────┐
    │ Error?     │
    └────┬───────┘
         │
    ┌────┴────┐
    │         │
   No        Yes
    │         │
    ▼         ▼
┌───────┐  ┌─────────────────────┐
│Return │  │Log Warning          │
│Result │  │(tflog.Warn)         │
└───────┘  └──────────┬──────────┘
                      │
                      ▼
           ┌──────────────────────┐
           │Return nil            │
           │(Caller uses SSH CLI) │
           └──────────────────────┘
```

### Error Types

| Error | Handling | User Impact |
|-------|----------|-------------|
| SFTP Connection Refused | Log warning, fallback to SSH | Slower but works |
| Authentication Failed | Log warning, fallback to SSH | Check admin_password |
| File Not Found | Log warning with path, fallback | Check sftp_config_path |
| Parse Error | Log error with line, fallback | Config syntax issue |

## Testing Strategy

### Unit Tests

| File | Coverage |
|------|----------|
| `sftp_client_test.go` | Mock SSH, download success/failure |
| `config_cache_test.go` | Concurrent access, invalidation, TTL |
| `config_path_resolver_test.go` | Japanese/English output parsing |
| `config_file_test.go` | Context tracking, password extraction |

### Integration Tests

**File:** `sftp_integration_test.go`

| Test | Description |
|------|-------------|
| `TestSFTPIntegration_FullFlow` | Path resolution → download → parse → cache |
| `TestSFTPIntegration_Fallback` | SFTP failure triggers SSH fallback |
| `TestSFTPIntegration_CacheReuse` | Multiple reads use single download |
| `TestSFTPIntegration_CacheInvalidation` | Dirty flag, TTL expiration |
| `TestSFTPIntegration_PathResolution` | Various `show environment` outputs |
| `TestSFTPIntegration_ConfigParsing` | Static routes, DHCP, NAT, tunnels |
| `TestSFTPIntegration_ErrorRecovery` | Transient SFTP errors |
| `TestSFTPIntegration_ContextCancellation` | Context cancellation handling |
| `TestSFTPIntegration_LargeConfig` | Performance with large configs |

## Password Extraction Patterns

The parser recognizes these password patterns:

| Pattern | Field |
|---------|-------|
| `login password <plaintext>` | LoginPassword |
| `administrator password <plaintext>` | AdminPassword |
| `login user <name> <password>` | Users (plaintext) |
| `login user <name> encrypted <hash>` | Users (encrypted=true) |
| `pp auth username <user> <password>` | PPAuth |
| `ipsec ike pre-shared-key <n> text <secret>` | IPsecPSK |
| `l2tp tunnel auth on <secret>` | L2TPAuth |

## Resource Conversion Functions

Each resource has a converter function to transform parser types to client types:

| Resource | Function | Location |
|----------|----------|----------|
| Static Route | `convertParsedStaticRoute()` | `resource_rtx_static_route.go` |
| DHCP Scope | `convertParsedDHCPScope()` | `resource_rtx_dhcp_scope.go` |
| NAT Masquerade | `convertParsedNATMasquerade()` | `resource_rtx_nat_masquerade.go` |

## Dependencies

| Package | Purpose | License |
|---------|---------|---------|
| `golang.org/x/crypto/ssh` | SSH connection | BSD-3-Clause |
| `github.com/pkg/sftp` | SFTP protocol | MIT |
| `sync` | Thread-safe cache | Go stdlib |
| `regexp` | Context pattern matching | Go stdlib |
