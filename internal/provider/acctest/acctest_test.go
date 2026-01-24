package acctest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRandomName verifies that RandomName generates unique names with the correct format.
func TestRandomName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "simple prefix",
			prefix: "test",
		},
		{
			name:   "empty prefix",
			prefix: "",
		},
		{
			name:   "long prefix",
			prefix: "my-very-long-resource-name",
		},
		{
			name:   "prefix with hyphens",
			prefix: "test-resource-name",
		},
		{
			name:   "prefix with underscores",
			prefix: "test_resource_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := RandomName(tt.prefix)

			// Verify prefix is present
			assert.True(t, strings.HasPrefix(result, tt.prefix+"-"),
				"expected name to start with prefix and hyphen, got %q", result)

			// Verify random suffix is 8 characters (default length)
			parts := strings.Split(result, "-")
			require.True(t, len(parts) >= 2, "expected at least 2 parts separated by hyphen")

			suffix := parts[len(parts)-1]
			assert.Len(t, suffix, 8, "expected random suffix to be 8 characters")
		})
	}
}

// TestRandomNameUniqueness verifies that multiple calls produce unique names.
func TestRandomNameUniqueness(t *testing.T) {
	t.Parallel()

	const iterations = 100
	names := make(map[string]struct{}, iterations)

	for i := 0; i < iterations; i++ {
		name := RandomName("test")
		if _, exists := names[name]; exists {
			t.Errorf("duplicate name generated: %s", name)
		}
		names[name] = struct{}{}
	}

	assert.Len(t, names, iterations, "expected all generated names to be unique")
}

// TestRandomNameWithLength verifies that RandomNameWithLength generates names with custom suffix length.
func TestRandomNameWithLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prefix       string
		suffixLength int
	}{
		{
			name:         "short suffix",
			prefix:       "test",
			suffixLength: 4,
		},
		{
			name:         "long suffix",
			prefix:       "test",
			suffixLength: 16,
		},
		{
			name:         "zero length",
			prefix:       "test",
			suffixLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := RandomNameWithLength(tt.prefix, tt.suffixLength)

			// Verify prefix is present
			assert.True(t, strings.HasPrefix(result, tt.prefix+"-"),
				"expected name to start with prefix and hyphen, got %q", result)

			// Verify suffix length
			parts := strings.Split(result, "-")
			suffix := parts[len(parts)-1]
			assert.Len(t, suffix, tt.suffixLength, "expected suffix length to match")
		})
	}
}

// TestRandomInt verifies that RandomInt generates integers within the specified range.
func TestRandomInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "small range",
			min:  1,
			max:  10,
		},
		{
			name: "large range",
			min:  100,
			max:  10000,
		},
		{
			name: "negative range",
			min:  -100,
			max:  -1,
		},
		{
			name: "zero included",
			min:  -10,
			max:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < 100; i++ {
				result := RandomInt(tt.min, tt.max)
				assert.GreaterOrEqual(t, result, tt.min, "result should be >= min")
				assert.LessOrEqual(t, result, tt.max, "result should be <= max")
			}
		})
	}
}

// TestRandomIP verifies that RandomIP generates valid IP addresses.
func TestRandomIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		prefix      string
		hostMin     int
		hostMax     int
		expectMatch string
	}{
		{
			name:        "class C subnet",
			prefix:      "192.168.1",
			hostMin:     1,
			hostMax:     254,
			expectMatch: "192.168.1.",
		},
		{
			name:        "class B subnet",
			prefix:      "172.16.0",
			hostMin:     1,
			hostMax:     254,
			expectMatch: "172.16.0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < 50; i++ {
				result := RandomIP(tt.prefix, tt.hostMin, tt.hostMax)

				assert.True(t, strings.HasPrefix(result, tt.expectMatch),
					"expected IP to start with %q, got %q", tt.expectMatch, result)
			}
		})
	}
}

// TestGetEnvOrDefault verifies environment variable retrieval with default fallback.
func TestGetEnvOrDefault(t *testing.T) {
	// This test does not use t.Parallel() because it manipulates environment variables

	tests := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue string
		expected     string
		setEnv       bool
	}{
		{
			name:         "env var set",
			envVar:       "TEST_ACCTEST_VAR",
			envValue:     "custom_value",
			defaultValue: "default",
			expected:     "custom_value",
			setEnv:       true,
		},
		{
			name:         "env var not set",
			envVar:       "TEST_ACCTEST_UNSET_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
			setEnv:       false,
		},
		{
			name:         "env var empty string",
			envVar:       "TEST_ACCTEST_EMPTY_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
			setEnv:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envVar, tt.envValue)
			}

			result := GetEnvOrDefault(tt.envVar, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTestCheckResourceExists verifies the resource existence check helper.
func TestTestCheckResourceExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceName string
		attrs        map[string]string
		expectError  bool
	}{
		{
			name:         "resource exists with ID",
			resourceName: "rtx_admin.test",
			attrs: map[string]string{
				"id":   "test-id",
				"name": "test-name",
			},
			expectError: false,
		},
		{
			name:         "resource not found (empty attrs)",
			resourceName: "rtx_admin.test",
			attrs:        map[string]string{},
			expectError:  true,
		},
		{
			name:         "resource with empty ID",
			resourceName: "rtx_admin.test",
			attrs: map[string]string{
				"id":   "",
				"name": "test-name",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checkFunc := TestCheckResourceExists(tt.resourceName)
			err := checkFunc(t, tt.attrs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
