package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// ============================================================================
// Mock Executor for Integration Testing
// ============================================================================

// integrationMockExecutor implements Executor for integration testing
type integrationMockExecutor struct {
	responses map[string][]byte
	errors    map[string]error
	mu        sync.RWMutex
}

func newIntegrationMockExecutor() *integrationMockExecutor {
	return &integrationMockExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
	}
}

func (m *integrationMockExecutor) SetResponse(cmd string, response []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[cmd] = response
}

func (m *integrationMockExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err, ok := m.errors[cmd]; ok {
		return nil, err
	}

	if response, ok := m.responses[cmd]; ok {
		return response, nil
	}

	return nil, fmt.Errorf("unexpected command: %s", cmd)
}

func (m *integrationMockExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	var result []byte
	for _, cmd := range cmds {
		output, err := m.Run(ctx, cmd)
		if err != nil {
			return result, err
		}
		result = append(result, output...)
	}
	return result, nil
}

func (m *integrationMockExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return nil
}

func (m *integrationMockExecutor) SetLoginPassword(ctx context.Context, password string) error {
	return nil
}

// ============================================================================
// Mock SFTP Client for Integration Testing
// ============================================================================

// mockSFTPClientForIntegration is a mock SFTP client for integration testing
type mockSFTPClientForIntegration struct {
	files        map[string][]byte
	downloadErr  error
	onDownload   func()
	downloadFunc func(path string) ([]byte, error)
	closed       bool
}

func (m *mockSFTPClientForIntegration) Download(ctx context.Context, path string) ([]byte, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if m.onDownload != nil {
		m.onDownload()
	}

	// Use custom download function if provided
	if m.downloadFunc != nil {
		return m.downloadFunc(path)
	}

	if m.downloadErr != nil {
		return nil, m.downloadErr
	}

	content, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return content, nil
}

func (m *mockSFTPClientForIntegration) Close() error {
	m.closed = true
	return nil
}

// ============================================================================
// Integration Tests
// ============================================================================

// Sample RTX router configuration for testing
const sampleRTXConfig = `# RTX1210 Config
#
console character ascii
ip route default gateway pp 1
ip lan1 address 192.168.1.1/24
pp select 1
 pp always-on on
 pppoe use lan2
 pp auth accept pap chap
 pp auth myname username password
 ppp lcp mru on 1454
 ppp ipcp ipaddress on
 ppp ipcp msext on
 ppp ccp type none
 ip pp mtu 1454
 ip pp nat descriptor 1
 pp enable 1
nat descriptor type 1 masquerade
nat descriptor address outer 1 primary
dns server pp 1
dns private address spoof on
dhcp service server
dhcp server rfc2131 compliant except remain-silent
dhcp scope 1 192.168.1.100-192.168.1.200/24
`

// TestSFTPIntegration_FullFlow tests the complete flow:
// path resolution -> SFTP download -> parse -> cache
func TestSFTPIntegration_FullFlow(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock executor that returns "show environment" output
	executor := newIntegrationMockExecutor()
	showEnvOutput := `# RTX1210 Rev.14.01.40
RTX1210 BootROM Ver. 1.00
デフォルト設定ファイル: config0
`
	executor.SetResponse("show environment", []byte(showEnvOutput))

	// Create mock SFTP client
	configContent := []byte(sampleRTXConfig)
	mockSFTP := &mockSFTPClientForIntegration{
		files: map[string][]byte{
			"/system/config0": configContent,
			"/system/config1": []byte("# Empty config"),
		},
	}

	// Create config cache
	cache := NewConfigCache()

	// Test path resolution
	resolver := NewConfigPathResolver(executor)
	ctx := context.Background()

	configPath, err := resolver.Resolve(ctx)
	require.NoError(t, err)
	assert.Equal(t, "/system/config0", configPath)

	// Test SFTP download
	content, err := mockSFTP.Download(ctx, configPath)
	require.NoError(t, err)
	assert.Equal(t, configContent, content)

	// Test parsing
	parser := parsers.NewConfigFileParser()
	parsed, err := parser.Parse(string(content))
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Greater(t, parsed.CommandCount, 0)

	// Verify parsed content
	routes := parsed.ExtractStaticRoutes()
	assert.Len(t, routes, 1)
	// "ip route default" is parsed as prefix "0.0.0.0"
	assert.Equal(t, "0.0.0.0", routes[0].Prefix)

	// Test cache storage
	cache.Set(string(content), parsed)

	// Verify cache retrieval
	cachedParsed, ok := cache.Get()
	require.True(t, ok)
	assert.Equal(t, parsed.CommandCount, cachedParsed.CommandCount)

	// Verify cache validity
	assert.True(t, cache.IsValid())
	assert.False(t, cache.IsDirty())
}

// TestSFTPIntegration_Fallback tests fallback to SSH when SFTP fails
func TestSFTPIntegration_Fallback(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock executor with both show environment and show config
	executor := newIntegrationMockExecutor()
	showEnvOutput := `# RTX1210 Rev.14.01.40
デフォルト設定ファイル: config0
`
	executor.SetResponse("show environment", []byte(showEnvOutput))
	executor.SetResponse("show config", []byte(sampleRTXConfig))

	// Create failing SFTP client
	mockSFTP := &mockSFTPClientForIntegration{
		downloadErr: errors.New("SFTP connection failed"),
	}

	ctx := context.Background()

	// Attempt SFTP download (should fail)
	_, err := mockSFTP.Download(ctx, "/system/config0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SFTP connection failed")

	// Fall back to SSH "show config"
	sshContent, err := executor.Run(ctx, "show config")
	require.NoError(t, err)
	assert.Contains(t, string(sshContent), "ip route default")

	// Parse the SSH output
	parser := parsers.NewConfigFileParser()
	parsed, err := parser.Parse(string(sshContent))
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Greater(t, parsed.CommandCount, 0)
}

// TestSFTPIntegration_CacheReuse tests that cache is properly reused
func TestSFTPIntegration_CacheReuse(t *testing.T) {
	// Create mock SFTP client with download counter
	downloadCount := 0
	configContent := []byte(sampleRTXConfig)
	mockSFTP := &mockSFTPClientForIntegration{
		files: map[string][]byte{
			"/system/config0": configContent,
		},
		onDownload: func() {
			downloadCount++
		},
	}

	ctx := context.Background()
	cache := NewConfigCache()

	// First access - should download
	content, err := mockSFTP.Download(ctx, "/system/config0")
	require.NoError(t, err)
	assert.Equal(t, 1, downloadCount)

	parser := parsers.NewConfigFileParser()
	parsed, err := parser.Parse(string(content))
	require.NoError(t, err)

	cache.Set(string(content), parsed)

	// Second access - should use cache
	cachedParsed, ok := cache.Get()
	require.True(t, ok)
	assert.Equal(t, parsed.CommandCount, cachedParsed.CommandCount)
	assert.Equal(t, 1, downloadCount) // Still 1, no new download

	// Third access - cache still valid
	assert.True(t, cache.IsValid())
	cachedParsed2, ok := cache.Get()
	require.True(t, ok)
	assert.Equal(t, cachedParsed.CommandCount, cachedParsed2.CommandCount)
	assert.Equal(t, 1, downloadCount) // Still 1

	// Mark dirty - simulates a write operation
	cache.MarkDirty()
	assert.True(t, cache.IsDirty())

	// Dirty cache should trigger re-download in real implementation
	// Here we verify the dirty flag is set
	assert.True(t, cache.IsDirty())

	// Download again after dirty
	content2, err := mockSFTP.Download(ctx, "/system/config0")
	require.NoError(t, err)
	assert.Equal(t, 2, downloadCount)

	parsed2, err := parser.Parse(string(content2))
	require.NoError(t, err)
	cache.Set(string(content2), parsed2)

	// Cache should be clean again
	assert.False(t, cache.IsDirty())
}

// TestSFTPIntegration_PathResolution tests different config path scenarios
func TestSFTPIntegration_PathResolution(t *testing.T) {
	tests := []struct {
		name           string
		envOutput      string
		expectedPath   string
		expectFallback bool
	}{
		{
			name:         "Japanese config0",
			envOutput:    "デフォルト設定ファイル: config0\n",
			expectedPath: "/system/config0",
		},
		{
			name:         "Japanese config1",
			envOutput:    "デフォルト設定ファイル: config1\n",
			expectedPath: "/system/config1",
		},
		{
			name:         "English format",
			envOutput:    "Default config file: config2\n",
			expectedPath: "/system/config2",
		},
		{
			name:           "Invalid output - fallback to config0",
			envOutput:      "Some unexpected output\n",
			expectedPath:   "/system/config0",
			expectFallback: true,
		},
		{
			name:           "Empty output - fallback to config0",
			envOutput:      "",
			expectedPath:   "/system/config0",
			expectFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newIntegrationMockExecutor()
			executor.SetResponse("show environment", []byte(tt.envOutput))

			resolver := NewConfigPathResolver(executor)
			ctx := context.Background()

			path, err := resolver.Resolve(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

// TestSFTPIntegration_ConfigParsing tests parsing various config formats
func TestSFTPIntegration_ConfigParsing(t *testing.T) {
	tests := []struct {
		name          string
		config        string
		expectedCount int
		checkFunc     func(t *testing.T, parsed *parsers.ParsedConfig)
	}{
		{
			name: "Basic static route",
			config: `ip route default gateway pp 1
ip route 10.0.0.0/8 gateway 192.168.1.254
`,
			expectedCount: 2,
			checkFunc: func(t *testing.T, parsed *parsers.ParsedConfig) {
				routes := parsed.ExtractStaticRoutes()
				assert.Len(t, routes, 2)
			},
		},
		{
			name: "DHCP scope",
			config: `dhcp scope 1 192.168.1.100-192.168.1.200/24
dhcp scope 2 192.168.2.100-192.168.2.200/24
`,
			expectedCount: 2,
			checkFunc: func(t *testing.T, parsed *parsers.ParsedConfig) {
				scopes := parsed.ExtractDHCPScopes()
				assert.Len(t, scopes, 2)
			},
		},
		{
			name: "NAT masquerade",
			config: `nat descriptor type 1 masquerade
nat descriptor address outer 1 primary
`,
			expectedCount: 2,
			checkFunc: func(t *testing.T, parsed *parsers.ParsedConfig) {
				nats := parsed.ExtractNATMasquerade()
				assert.Len(t, nats, 1)
			},
		},
		{
			name: "PP context",
			config: `pp select 1
 pp always-on on
 pppoe use lan2
 pp enable 1
`,
			expectedCount: 4, // includes "pp enable 1" as a command within context
			checkFunc: func(t *testing.T, parsed *parsers.ParsedConfig) {
				// Should have PP context
				assert.Len(t, parsed.Contexts, 1)
				assert.Equal(t, parsers.ContextPP, parsed.Contexts[0].Type)
				assert.Equal(t, 1, parsed.Contexts[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := parsers.NewConfigFileParser()
			parsed, err := parser.Parse(tt.config)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, parsed.CommandCount)

			if tt.checkFunc != nil {
				tt.checkFunc(t, parsed)
			}
		})
	}
}

// TestSFTPIntegration_CacheInvalidation tests cache invalidation scenarios
func TestSFTPIntegration_CacheInvalidation(t *testing.T) {
	cache := NewConfigCache()

	// Set initial cache
	parser := parsers.NewConfigFileParser()
	parsed, err := parser.Parse(sampleRTXConfig)
	require.NoError(t, err)

	cache.Set(sampleRTXConfig, parsed)

	// Verify cache is valid
	assert.True(t, cache.IsValid())
	assert.False(t, cache.IsDirty())

	// Test 1: Mark dirty
	cache.MarkDirty()
	assert.True(t, cache.IsDirty())
	assert.True(t, cache.IsValid()) // Still valid, just dirty

	// Clear dirty
	cache.ClearDirty()
	assert.False(t, cache.IsDirty())

	// Test 2: Invalidate completely
	cache.Invalidate()
	assert.False(t, cache.IsValid())

	_, ok := cache.Get()
	assert.False(t, ok)

	// Test 3: Set with short TTL and verify expiration
	cache.SetWithTTL(sampleRTXConfig, parsed, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, cache.IsValid())
}

// TestSFTPIntegration_ErrorRecovery tests error recovery scenarios
func TestSFTPIntegration_ErrorRecovery(t *testing.T) {
	// Simulate transient SFTP errors
	failCount := 0
	maxFails := 2
	mockSFTP := &mockSFTPClientForIntegration{
		files: map[string][]byte{
			"/system/config0": []byte(sampleRTXConfig),
		},
		onDownload: func() {
			failCount++
		},
		downloadFunc: func(path string) ([]byte, error) {
			if failCount <= maxFails {
				return nil, errors.New("temporary SFTP error")
			}
			return []byte(sampleRTXConfig), nil
		},
	}

	ctx := context.Background()

	// First attempts should fail
	for i := 0; i < maxFails; i++ {
		_, err := mockSFTP.Download(ctx, "/system/config0")
		require.Error(t, err)
	}

	// Third attempt should succeed
	content, err := mockSFTP.Download(ctx, "/system/config0")
	require.NoError(t, err)
	assert.Contains(t, string(content), "ip route default")
}

// TestSFTPIntegration_ContextCancellation tests context cancellation handling
func TestSFTPIntegration_ContextCancellation(t *testing.T) {
	mockSFTP := &mockSFTPClientForIntegration{
		files: map[string][]byte{
			"/system/config0": []byte(sampleRTXConfig),
		},
	}

	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mockSFTP.Download(ctx, "/system/config0")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

// TestSFTPIntegration_LargeConfig tests handling of large configuration files
func TestSFTPIntegration_LargeConfig(t *testing.T) {
	// Generate a large config
	var sb strings.Builder
	sb.WriteString("# Large RTX Config\n")
	for i := 0; i < 1000; i++ {
		sb.WriteString(fmt.Sprintf("ip filter %d pass * * * * *\n", i+1))
	}
	largeConfig := sb.String()

	mockSFTP := &mockSFTPClientForIntegration{
		files: map[string][]byte{
			"/system/config0": []byte(largeConfig),
		},
	}

	ctx := context.Background()
	content, err := mockSFTP.Download(ctx, "/system/config0")
	require.NoError(t, err)
	assert.Equal(t, len(largeConfig), len(content))

	// Parse the large config
	parser := parsers.NewConfigFileParser()
	parsed, err := parser.Parse(string(content))
	require.NoError(t, err)
	assert.Equal(t, 1000, parsed.CommandCount)
}
