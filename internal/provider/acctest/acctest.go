// Package acctest provides acceptance test utilities for the RTX Terraform provider.
// This package contains shared test infrastructure including PreCheck functions,
// random name generators, and provider factory setup.
package acctest

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/provider"
)

// ProviderFactories is a map of provider factory functions for use in acceptance tests.
// It provides a configured "rtx" provider instance for Terraform test cases.
var ProviderFactories = map[string]func() (*schema.Provider, error){
	"rtx": func() (*schema.Provider, error) {
		return provider.New("test"), nil
	},
}

// RequiredEnvVars lists the environment variables required for acceptance tests.
var RequiredEnvVars = []string{
	"RTX_HOST",
	"RTX_USERNAME",
	"RTX_PASSWORD",
}

// OptionalEnvVars lists optional environment variables that enhance test functionality.
var OptionalEnvVars = []string{
	"RTX_ADMIN_PASSWORD",
	"RTX_PORT",
	"RTX_SSH_HOST_KEY",
	"RTX_KNOWN_HOSTS_FILE",
	"RTX_SKIP_HOST_KEY_CHECK",
}

// PreCheck verifies that all required prerequisites for acceptance tests are met.
// It checks for required environment variables and validates router connectivity settings.
// Call this function in the PreCheck field of resource.TestCase.
func PreCheck(t *testing.T) {
	t.Helper()

	// Verify acceptance test mode is enabled
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	// Check required environment variables
	var missingVars []string
	for _, envVar := range RequiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		t.Fatalf("Required environment variables not set: %v\n"+
			"Please configure these variables to run acceptance tests:\n"+
			"  RTX_HOST     - Hostname or IP address of the RTX router\n"+
			"  RTX_USERNAME - Username for SSH authentication\n"+
			"  RTX_PASSWORD - Password for SSH authentication",
			missingVars)
	}

	// Validate that host is not empty (additional safety check)
	if host := os.Getenv("RTX_HOST"); host == "" {
		t.Fatal("RTX_HOST environment variable is set but empty")
	}
}

// PreCheckWithAdminPassword verifies prerequisites including admin password.
// Use this for tests that require administrative privileges to modify router configuration.
func PreCheckWithAdminPassword(t *testing.T) {
	t.Helper()

	PreCheck(t)

	// Admin password is optional but some tests may require it explicitly
	if os.Getenv("RTX_ADMIN_PASSWORD") == "" && os.Getenv("RTX_PASSWORD") == "" {
		t.Fatal("RTX_ADMIN_PASSWORD or RTX_PASSWORD must be set for tests requiring admin access")
	}
}

// PreCheckFunc returns a function suitable for use as resource.TestCase PreCheck.
// This is a convenience wrapper for PreCheck that returns a closure.
func PreCheckFunc(t *testing.T) func() {
	return func() { PreCheck(t) }
}

// RandomName generates a unique resource name with the given prefix.
// The generated name is suitable for use in parallel tests to avoid resource conflicts.
// The name format is: prefix-<random_suffix>
// Example: RandomName("test") might return "test-abc123"
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, acctest.RandString(8))
}

// RandomNameWithLength generates a unique resource name with a specified random suffix length.
// Use this when you need more control over the name length.
func RandomNameWithLength(prefix string, length int) string {
	return fmt.Sprintf("%s-%s", prefix, acctest.RandString(length))
}

// RandomInt generates a random integer suitable for test resource identifiers.
// Returns an integer in the range [min, max].
func RandomInt(min, max int) int {
	return acctest.RandIntRange(min, max)
}

// RandomIP generates a random IP address in the specified subnet.
// The subnet should be in CIDR notation (e.g., "192.168.1.0/24").
// This is useful for generating unique IP addresses in tests.
func RandomIP(subnetPrefix string, hostMin, hostMax int) string {
	host := RandomInt(hostMin, hostMax)
	return fmt.Sprintf("%s.%d", subnetPrefix, host)
}

// SkipIfEnvNotSet skips the test if the specified environment variable is not set.
// Use this for tests that depend on optional configuration.
func SkipIfEnvNotSet(t *testing.T, envVar string) {
	t.Helper()

	if os.Getenv(envVar) == "" {
		t.Skipf("Environment variable %s not set, skipping test", envVar)
	}
}

// GetEnvOrDefault returns the value of an environment variable or a default value.
func GetEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvOrSkip returns the value of an environment variable or skips the test.
func GetEnvOrSkip(t *testing.T, envVar string) string {
	t.Helper()

	value := os.Getenv(envVar)
	if value == "" {
		t.Skipf("Environment variable %s not set, skipping test", envVar)
	}
	return value
}

// TestCheckResourceExists is a helper that verifies a resource exists in state.
// It returns an error if the resource is not found or has no ID set.
func TestCheckResourceExists(resourceName string) func(*testing.T, map[string]string) error {
	return func(t *testing.T, attrs map[string]string) error {
		if len(attrs) == 0 {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		if attrs["id"] == "" {
			return fmt.Errorf("resource %s has no ID set", resourceName)
		}

		return nil
	}
}
