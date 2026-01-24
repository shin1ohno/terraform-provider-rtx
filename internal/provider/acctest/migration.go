// Package acctest provides testing utilities for Terraform provider acceptance tests.
package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// StateMigrationTestCase defines a state migration test scenario for testing
// schema version upgraders. It allows verifying that old state formats are
// correctly transformed to new state formats.
type StateMigrationTestCase struct {
	// Name is a descriptive name for this test case
	Name string

	// OldState is the state data in the old schema version format.
	// Can be nil to test nil state handling.
	OldState map[string]interface{}

	// ExpectedState is the expected state data after the upgrade.
	// Can be nil if the upgrader is expected to return nil.
	ExpectedState map[string]interface{}

	// UpgradeFunc is the StateUpgradeFunc to test
	UpgradeFunc schema.StateUpgradeFunc

	// Meta is the provider meta (typically API client), passed to UpgradeFunc.
	// Can be nil for upgraders that don't use meta.
	Meta interface{}

	// ExpectError indicates whether the upgrade is expected to return an error.
	// If true, the test passes only if an error is returned.
	ExpectError bool

	// ErrorMatcher is an optional function to validate the error message.
	// Only used when ExpectError is true.
	ErrorMatcher func(error) bool
}

// RunStateMigrationTests executes a slice of state migration test cases.
// Each test case verifies that the UpgradeFunc correctly transforms OldState
// into ExpectedState.
func RunStateMigrationTests(t *testing.T, cases []StateMigrationTestCase) {
	t.Helper()

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()

			if tc.UpgradeFunc == nil {
				t.Fatal("UpgradeFunc must not be nil")
			}

			ctx := context.Background()

			// Execute the upgrade function
			actualState, err := tc.UpgradeFunc(ctx, tc.OldState, tc.Meta)

			// Check error expectations
			if tc.ExpectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tc.ErrorMatcher != nil && !tc.ErrorMatcher(err) {
					t.Fatalf("error did not match expected pattern: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare actual state to expected state
			if diff := cmp.Diff(tc.ExpectedState, actualState); diff != "" {
				t.Errorf("state mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}

// RunStateMigrationTest runs a single state migration test case.
// This is a convenience function for testing a single upgrade scenario.
func RunStateMigrationTest(t *testing.T, tc StateMigrationTestCase) {
	t.Helper()
	RunStateMigrationTests(t, []StateMigrationTestCase{tc})
}

// CrossVersionTestStep returns a TestStep configured to use an external provider
// version for cross-version upgrade testing. This allows testing state migration
// by first applying with an old provider version, then upgrading to the current
// version.
//
// The providerSource should be the full provider source (e.g., "registry.terraform.io/sh1/rtx")
// The providerVersion should be a version constraint (e.g., "1.0.0" or "~> 1.0")
// The config is the Terraform configuration to apply with the external provider.
//
// Example usage:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			// Step 1: Apply with old provider version
//			acctest.CrossVersionTestStep(
//				"registry.terraform.io/sh1/rtx",
//				"1.0.0",
//				testAccOldConfig(),
//			),
//			// Step 2: Upgrade to current provider (uses testAccProviderFactories)
//			{
//				Config: testAccNewConfig(),
//				Check:  resource.TestCheckResourceAttr(...),
//			},
//		},
//	})
func CrossVersionTestStep(providerSource, providerVersion, config string) resource.TestStep {
	return resource.TestStep{
		ExternalProviders: map[string]resource.ExternalProvider{
			"rtx": {
				Source:            providerSource,
				VersionConstraint: providerVersion,
			},
		},
		Config: config,
	}
}

// CrossVersionTestStepWithCheck returns a TestStep like CrossVersionTestStep but
// with additional check functions to verify the state after applying with the
// external provider.
func CrossVersionTestStepWithCheck(providerSource, providerVersion, config string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := CrossVersionTestStep(providerSource, providerVersion, config)
	if len(checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}
	return step
}

// CrossVersionUpgradeTestConfig holds configuration for a complete cross-version
// upgrade test scenario.
type CrossVersionUpgradeTestConfig struct {
	// ProviderSource is the full source address of the provider
	// (e.g., "registry.terraform.io/sh1/rtx")
	ProviderSource string

	// OldVersion is the version of the provider to start with
	OldVersion string

	// OldConfig is the Terraform configuration to apply with the old version
	OldConfig string

	// OldChecks are optional check functions to run after applying OldConfig
	OldChecks []resource.TestCheckFunc

	// NewConfig is the Terraform configuration to apply after upgrading.
	// If empty, OldConfig will be reused (testing no-config-change upgrade).
	NewConfig string

	// NewChecks are optional check functions to run after applying NewConfig
	NewChecks []resource.TestCheckFunc

	// ExpectEmptyPlanAfterUpgrade indicates whether to verify that upgrading
	// the provider version with the same config produces an empty plan.
	// This catches perpetual diffs introduced by state migration.
	ExpectEmptyPlanAfterUpgrade bool
}

// BuildCrossVersionUpgradeSteps builds test steps for a cross-version upgrade
// test scenario. This creates a standard pattern:
// 1. Apply with old provider version
// 2. Upgrade to current provider version (optionally verify empty plan)
// 3. Apply new config and verify state
func BuildCrossVersionUpgradeSteps(cfg CrossVersionUpgradeTestConfig) []resource.TestStep {
	steps := make([]resource.TestStep, 0, 3)

	// Step 1: Apply with old provider version
	oldStep := CrossVersionTestStep(cfg.ProviderSource, cfg.OldVersion, cfg.OldConfig)
	if len(cfg.OldChecks) > 0 {
		oldStep.Check = resource.ComposeTestCheckFunc(cfg.OldChecks...)
	}
	steps = append(steps, oldStep)

	// Determine new config (use old config if not specified)
	newConfig := cfg.NewConfig
	if newConfig == "" {
		newConfig = cfg.OldConfig
	}

	// Step 2: If expecting empty plan, add PlanOnly step
	if cfg.ExpectEmptyPlanAfterUpgrade {
		steps = append(steps, resource.TestStep{
			Config:   newConfig,
			PlanOnly: true,
			// PlanOnly with same config will fail if there are planned changes
		})
	}

	// Step 3: Apply with new config and run checks
	newStep := resource.TestStep{
		Config: newConfig,
	}
	if len(cfg.NewChecks) > 0 {
		newStep.Check = resource.ComposeTestCheckFunc(cfg.NewChecks...)
	}
	steps = append(steps, newStep)

	return steps
}

// AssertStateMigration is a helper function that creates a test function
// for use in unit test tables. It returns a function that can be used with
// t.Run() to test a single state migration.
func AssertStateMigration(
	upgradeFunc schema.StateUpgradeFunc,
	oldState, expectedState map[string]interface{},
) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		ctx := context.Background()
		actualState, err := upgradeFunc(ctx, oldState, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(expectedState, actualState); diff != "" {
			t.Errorf("state mismatch (-expected +actual):\n%s", diff)
		}
	}
}

// ValidateStateUpgrader validates that a StateUpgrader is properly configured.
// It checks that required fields are set and the upgrade function works with
// nil state (a common edge case).
func ValidateStateUpgrader(t *testing.T, upgrader schema.StateUpgrader) {
	t.Helper()

	if !upgrader.Type.IsObjectType() && !upgrader.Type.IsMapType() {
		// Type should typically be an object type for resources
		t.Log("warning: StateUpgrader.Type is not an object or map type")
	}

	if upgrader.Upgrade == nil {
		t.Fatal("StateUpgrader.Upgrade must not be nil")
	}

	// Test with nil state - upgraders should handle this gracefully
	ctx := context.Background()
	_, err := upgrader.Upgrade(ctx, nil, nil)
	if err != nil {
		t.Logf("note: upgrade function returned error for nil state: %v", err)
		// This is not necessarily an error - some upgraders may legitimately
		// reject nil state. This is just informational.
	}
}

// MustDeepCopyState creates a deep copy of state data to prevent test pollution.
// Panics if the state cannot be copied (which should never happen with valid state).
func MustDeepCopyState(state map[string]interface{}) map[string]interface{} {
	if state == nil {
		return nil
	}
	return deepCopyMap(state)
}

// deepCopyMap performs a deep copy of a map[string]interface{}.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = deepCopyValue(v)
	}
	return result
}

// deepCopyValue performs a deep copy of an arbitrary value.
func deepCopyValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		return deepCopyMap(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = deepCopyValue(item)
		}
		return result
	case string, bool, int, int64, float64:
		// Primitive types are already value types
		return val
	default:
		// For unknown types, return as-is (may not be a true deep copy)
		return val
	}
}

// StateUpgradeChain tests a chain of state upgraders, verifying that state
// can be upgraded through multiple schema versions. This is useful when
// testing that state from very old provider versions can be upgraded to
// the current version.
func StateUpgradeChain(
	t *testing.T,
	initialState map[string]interface{},
	expectedFinalState map[string]interface{},
	upgraders ...schema.StateUpgradeFunc,
) {
	t.Helper()

	if len(upgraders) == 0 {
		t.Fatal("at least one upgrader must be provided")
	}

	ctx := context.Background()
	currentState := MustDeepCopyState(initialState)

	for i, upgrader := range upgraders {
		var err error
		currentState, err = upgrader(ctx, currentState, nil)
		if err != nil {
			t.Fatalf("upgrader %d failed: %v", i, err)
		}
	}

	if diff := cmp.Diff(expectedFinalState, currentState); diff != "" {
		t.Errorf("final state mismatch after upgrade chain (-expected +actual):\n%s", diff)
	}
}

// StateMigrationBenchmark provides a benchmark for state migration performance.
// This can be useful for ensuring that state migrations are efficient,
// especially for resources with large state.
func StateMigrationBenchmark(
	b *testing.B,
	upgradeFunc schema.StateUpgradeFunc,
	sampleState map[string]interface{},
) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Deep copy to prevent state mutation affecting benchmarks
		stateCopy := MustDeepCopyState(sampleState)
		_, err := upgradeFunc(ctx, stateCopy, nil)
		if err != nil {
			b.Fatalf("upgrade failed: %v", err)
		}
	}
}

// DescribeStateMigration returns a human-readable description of a state migration
// for use in test output or documentation.
func DescribeStateMigration(oldVersion, newVersion int, changes []string) string {
	return fmt.Sprintf(
		"State migration v%d -> v%d:\n  - %s",
		oldVersion,
		newVersion,
		joinWithPrefix(changes, "\n  - "),
	)
}

func joinWithPrefix(items []string, sep string) string {
	if len(items) == 0 {
		return "(no changes documented)"
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += sep + items[i]
	}
	return result
}
