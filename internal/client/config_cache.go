package client

import (
	"sync"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// DefaultCacheTTL is the default time-to-live for cached configuration.
// This determines how long the cache is considered "valid" before requiring
// a refresh check.
const DefaultCacheTTL = 5 * time.Minute

// ConfigCache provides thread-safe caching for router configuration data.
// It stores both the raw configuration content and the parsed structure,
// with support for TTL-based validity checking and dirty flag management.
type ConfigCache struct {
	mu         sync.RWMutex
	content    string                // Raw config file content
	parsed     *parsers.ParsedConfig // Parsed configuration data
	validUntil time.Time             // Cache invalidation timestamp
	dirty      bool                  // True if write occurred, requiring refresh
}

// NewConfigCache creates a new empty configuration cache.
func NewConfigCache() *ConfigCache {
	return &ConfigCache{}
}

// Get returns the cached parsed configuration and a boolean indicating
// whether valid data exists in the cache.
// This method is safe for concurrent access.
func (c *ConfigCache) Get() (*parsers.ParsedConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.parsed == nil {
		return nil, false
	}
	return c.parsed, true
}

// GetRaw returns the raw configuration content.
// Returns an empty string if the cache is empty.
// This method is safe for concurrent access.
func (c *ConfigCache) GetRaw() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.content
}

// Set stores the raw content and parsed configuration in the cache
// with the default TTL. This also clears the dirty flag.
// This method is safe for concurrent access.
func (c *ConfigCache) Set(content string, parsed *parsers.ParsedConfig) {
	c.SetWithTTL(content, parsed, DefaultCacheTTL)
}

// SetWithTTL stores the raw content and parsed configuration in the cache
// with a custom TTL. This also clears the dirty flag.
// This method is safe for concurrent access.
func (c *ConfigCache) SetWithTTL(content string, parsed *parsers.ParsedConfig, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content = content
	c.parsed = parsed
	c.validUntil = time.Now().Add(ttl)
	c.dirty = false
}

// Invalidate clears all cached data and resets the validity timestamp.
// This also clears the dirty flag.
// This method is safe for concurrent access.
func (c *ConfigCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content = ""
	c.parsed = nil
	c.validUntil = time.Time{}
	c.dirty = false
}

// MarkDirty sets the dirty flag to indicate that the router configuration
// has been modified and the cache may be stale.
// This method is safe for concurrent access.
func (c *ConfigCache) MarkDirty() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dirty = true
}

// ClearDirty resets the dirty flag.
// This method is safe for concurrent access.
func (c *ConfigCache) ClearDirty() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dirty = false
}

// IsDirty returns true if the cache has been marked dirty,
// indicating that the router configuration may have changed.
// This method is safe for concurrent access.
func (c *ConfigCache) IsDirty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.dirty
}

// IsValid returns true if the cache contains data and the TTL has not expired.
// Note: This does not consider the dirty flag; use IsDirty() separately
// to check if a refresh is recommended.
// This method is safe for concurrent access.
func (c *ConfigCache) IsValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.parsed == nil {
		return false
	}
	return time.Now().Before(c.validUntil)
}
