package client

import (
	"sync"
	"testing"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func TestNewConfigCache(t *testing.T) {
	cache := NewConfigCache()

	if cache == nil {
		t.Fatal("NewConfigCache() returned nil")
	}
}

func TestConfigCache_Get_Empty(t *testing.T) {
	cache := NewConfigCache()

	parsed, ok := cache.Get()

	if ok {
		t.Error("Get() on empty cache should return false")
	}
	if parsed != nil {
		t.Error("Get() on empty cache should return nil")
	}
}

func TestConfigCache_Set_And_Get(t *testing.T) {
	cache := NewConfigCache()

	rawContent := "ip route default gateway 192.168.1.1\ndhcp service server"
	parsed := &parsers.ParsedConfig{
		Raw: rawContent,
	}

	cache.Set(rawContent, parsed)

	got, ok := cache.Get()

	if !ok {
		t.Error("Get() after Set() should return true")
	}
	if got == nil {
		t.Fatal("Get() after Set() should return non-nil ParsedConfig")
	}
	if got.Raw != rawContent {
		t.Errorf("Get().Raw = %q, want %q", got.Raw, rawContent)
	}
}

func TestConfigCache_GetRaw(t *testing.T) {
	cache := NewConfigCache()

	// Empty cache
	raw := cache.GetRaw()
	if raw != "" {
		t.Errorf("GetRaw() on empty cache = %q, want empty string", raw)
	}

	// After set
	expectedRaw := "ip lan1 address 192.168.1.1/24"
	cache.Set(expectedRaw, &parsers.ParsedConfig{Raw: expectedRaw})

	raw = cache.GetRaw()
	if raw != expectedRaw {
		t.Errorf("GetRaw() = %q, want %q", raw, expectedRaw)
	}
}

func TestConfigCache_Invalidate(t *testing.T) {
	cache := NewConfigCache()

	rawContent := "test config"
	parsed := &parsers.ParsedConfig{Raw: rawContent}

	cache.Set(rawContent, parsed)

	// Verify cache is set
	_, ok := cache.Get()
	if !ok {
		t.Fatal("Get() should return true after Set()")
	}

	// Invalidate
	cache.Invalidate()

	// Verify cache is cleared
	got, ok := cache.Get()
	if ok {
		t.Error("Get() after Invalidate() should return false")
	}
	if got != nil {
		t.Error("Get() after Invalidate() should return nil")
	}

	// Verify raw content is also cleared
	raw := cache.GetRaw()
	if raw != "" {
		t.Errorf("GetRaw() after Invalidate() = %q, want empty string", raw)
	}
}

func TestConfigCache_MarkDirty(t *testing.T) {
	cache := NewConfigCache()

	// Initially not dirty
	if cache.IsDirty() {
		t.Error("IsDirty() on new cache should return false")
	}

	// Mark dirty
	cache.MarkDirty()

	if !cache.IsDirty() {
		t.Error("IsDirty() after MarkDirty() should return true")
	}
}

func TestConfigCache_ClearDirty(t *testing.T) {
	cache := NewConfigCache()

	cache.MarkDirty()
	if !cache.IsDirty() {
		t.Fatal("IsDirty() after MarkDirty() should return true")
	}

	cache.ClearDirty()

	if cache.IsDirty() {
		t.Error("IsDirty() after ClearDirty() should return false")
	}
}

func TestConfigCache_Invalidate_ClearsDirty(t *testing.T) {
	cache := NewConfigCache()

	cache.MarkDirty()
	cache.Invalidate()

	// Invalidate should also clear dirty flag
	if cache.IsDirty() {
		t.Error("IsDirty() after Invalidate() should return false")
	}
}

func TestConfigCache_Set_ClearsDirty(t *testing.T) {
	cache := NewConfigCache()

	cache.MarkDirty()
	if !cache.IsDirty() {
		t.Fatal("IsDirty() after MarkDirty() should return true")
	}

	cache.Set("new content", &parsers.ParsedConfig{Raw: "new content"})

	// Set should clear dirty flag since fresh data is loaded
	if cache.IsDirty() {
		t.Error("IsDirty() after Set() should return false")
	}
}

// Thread safety tests

func TestConfigCache_ThreadSafe_ConcurrentReads(t *testing.T) {
	cache := NewConfigCache()

	rawContent := "concurrent read test config"
	parsed := &parsers.ParsedConfig{Raw: rawContent}
	cache.Set(rawContent, parsed)

	var wg sync.WaitGroup
	numReaders := 100

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, ok := cache.Get()
			if !ok {
				t.Error("Get() should return true")
			}
			if got.Raw != rawContent {
				t.Errorf("Get().Raw = %q, want %q", got.Raw, rawContent)
			}
		}()
	}

	wg.Wait()
}

func TestConfigCache_ThreadSafe_ConcurrentReadWrite(t *testing.T) {
	cache := NewConfigCache()

	var wg sync.WaitGroup
	numOperations := 100

	// Writers
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			content := "config version"
			parsed := &parsers.ParsedConfig{Raw: content}
			cache.Set(content, parsed)
		}(i)
	}

	// Readers
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// These operations should not panic or race
			cache.Get()
			cache.GetRaw()
			cache.IsDirty()
		}()
	}

	// Invalidators
	for i := 0; i < numOperations/10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Invalidate()
		}()
	}

	// Dirty markers
	for i := 0; i < numOperations/10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MarkDirty()
			cache.ClearDirty()
		}()
	}

	wg.Wait()
}

func TestConfigCache_ThreadSafe_NoRaceCondition(t *testing.T) {
	// This test is mainly to be run with -race flag
	cache := NewConfigCache()

	done := make(chan bool)
	iterations := 1000

	// Goroutine 1: Continuous writes
	go func() {
		for i := 0; i < iterations; i++ {
			content := "test"
			parsed := &parsers.ParsedConfig{Raw: content}
			cache.Set(content, parsed)
		}
		done <- true
	}()

	// Goroutine 2: Continuous reads
	go func() {
		for i := 0; i < iterations; i++ {
			cache.Get()
			cache.GetRaw()
		}
		done <- true
	}()

	// Goroutine 3: Continuous invalidation
	go func() {
		for i := 0; i < iterations; i++ {
			cache.Invalidate()
		}
		done <- true
	}()

	// Goroutine 4: Dirty flag operations
	go func() {
		for i := 0; i < iterations; i++ {
			cache.MarkDirty()
			cache.IsDirty()
			cache.ClearDirty()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestConfigCache_ValidUntil(t *testing.T) {
	cache := NewConfigCache()

	// Set with default TTL
	content := "test content"
	parsed := &parsers.ParsedConfig{Raw: content}
	cache.Set(content, parsed)

	// Should be valid immediately after set
	_, ok := cache.Get()
	if !ok {
		t.Error("Get() should return true immediately after Set()")
	}

	// The cache should have a validUntil time set
	// We can test this by checking IsValid method
	if !cache.IsValid() {
		t.Error("IsValid() should return true immediately after Set()")
	}
}

func TestConfigCache_IsValid_EmptyCache(t *testing.T) {
	cache := NewConfigCache()

	if cache.IsValid() {
		t.Error("IsValid() on empty cache should return false")
	}
}

func TestConfigCache_IsValid_AfterInvalidate(t *testing.T) {
	cache := NewConfigCache()

	content := "test"
	parsed := &parsers.ParsedConfig{Raw: content}
	cache.Set(content, parsed)

	cache.Invalidate()

	if cache.IsValid() {
		t.Error("IsValid() after Invalidate() should return false")
	}
}

func TestConfigCache_SetWithTTL(t *testing.T) {
	cache := NewConfigCache()

	content := "test"
	parsed := &parsers.ParsedConfig{Raw: content}
	ttl := 1 * time.Second

	cache.SetWithTTL(content, parsed, ttl)

	// Should be valid immediately
	if !cache.IsValid() {
		t.Error("IsValid() should return true immediately after SetWithTTL()")
	}

	// Wait for expiration
	time.Sleep(ttl + 100*time.Millisecond)

	// Should be invalid after TTL
	if cache.IsValid() {
		t.Error("IsValid() should return false after TTL expired")
	}

	// Get should still return cached value (validity check is separate)
	got, ok := cache.Get()
	if !ok {
		t.Error("Get() should still return true even after TTL expired")
	}
	if got.Raw != content {
		t.Errorf("Get().Raw = %q, want %q", got.Raw, content)
	}
}

func TestConfigCache_Multiple_Sets(t *testing.T) {
	cache := NewConfigCache()

	// First set
	content1 := "config version 1"
	parsed1 := &parsers.ParsedConfig{Raw: content1}
	cache.Set(content1, parsed1)

	// Second set (should replace)
	content2 := "config version 2"
	parsed2 := &parsers.ParsedConfig{Raw: content2}
	cache.Set(content2, parsed2)

	got, ok := cache.Get()
	if !ok {
		t.Error("Get() should return true")
	}
	if got.Raw != content2 {
		t.Errorf("Get().Raw = %q, want %q", got.Raw, content2)
	}
}
