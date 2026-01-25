package acctest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunStateMigrationTests(t *testing.T) {
	// Simple upgrader that renames a field
	renameUpgrader := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		if rawState == nil {
			return nil, nil
		}
		result := make(map[string]interface{})
		for k, v := range rawState {
			if k == "old_name" {
				result["new_name"] = v
			} else {
				result[k] = v
			}
		}
		return result, nil
	}

	// Upgrader that always returns an error
	errorUpgrader := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		return nil, errors.New("upgrade failed")
	}

	tests := []struct {
		name       string
		testCase   StateMigrationTestCase
		shouldFail bool
	}{
		{
			name: "successful field rename",
			testCase: StateMigrationTestCase{
				Name: "rename old_name to new_name",
				OldState: map[string]interface{}{
					"old_name": "value",
					"other":    "preserved",
				},
				ExpectedState: map[string]interface{}{
					"new_name": "value",
					"other":    "preserved",
				},
				UpgradeFunc: renameUpgrader,
			},
			shouldFail: false,
		},
		{
			name: "nil state handling",
			testCase: StateMigrationTestCase{
				Name:          "nil state returns nil",
				OldState:      nil,
				ExpectedState: nil,
				UpgradeFunc:   renameUpgrader,
			},
			shouldFail: false,
		},
		{
			name: "expected error occurs",
			testCase: StateMigrationTestCase{
				Name:        "error expected",
				OldState:    map[string]interface{}{"key": "value"},
				UpgradeFunc: errorUpgrader,
				ExpectError: true,
			},
			shouldFail: false,
		},
		{
			name: "expected error with matcher",
			testCase: StateMigrationTestCase{
				Name:        "error with pattern",
				OldState:    map[string]interface{}{"key": "value"},
				UpgradeFunc: errorUpgrader,
				ExpectError: true,
				ErrorMatcher: func(err error) bool {
					return strings.Contains(err.Error(), "upgrade failed")
				},
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test for expected test failures in a sub-test,
			// so we just run the successful cases
			if !tt.shouldFail {
				RunStateMigrationTests(t, []StateMigrationTestCase{tt.testCase})
			}
		})
	}
}

func TestRunStateMigrationTest(t *testing.T) {
	addFieldUpgrader := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		if rawState == nil {
			return nil, nil
		}
		result := make(map[string]interface{})
		for k, v := range rawState {
			result[k] = v
		}
		result["new_field"] = "default_value"
		return result, nil
	}

	RunStateMigrationTest(t, StateMigrationTestCase{
		Name: "add new field with default",
		OldState: map[string]interface{}{
			"existing": "value",
		},
		ExpectedState: map[string]interface{}{
			"existing":  "value",
			"new_field": "default_value",
		},
		UpgradeFunc: addFieldUpgrader,
	})
}

func TestCrossVersionTestStep(t *testing.T) {
	step := CrossVersionTestStep(
		"shin1ohno/rtx",
		"1.0.0",
		`resource "rtx_admin" "test" { name = "test" }`,
	)

	assert.NotNil(t, step.ExternalProviders)
	assert.Contains(t, step.ExternalProviders, "rtx")
	assert.Equal(t, "shin1ohno/rtx", step.ExternalProviders["rtx"].Source)
	assert.Equal(t, "1.0.0", step.ExternalProviders["rtx"].VersionConstraint)
	assert.Contains(t, step.Config, "rtx_admin")
}

func TestCrossVersionTestStepWithCheck(t *testing.T) {
	checkCalled := false
	checkFunc := func(s *terraform.State) error {
		checkCalled = true
		return nil
	}

	step := CrossVersionTestStepWithCheck(
		"shin1ohno/rtx",
		"1.0.0",
		`resource "rtx_admin" "test" { name = "test" }`,
		checkFunc,
	)

	assert.NotNil(t, step.ExternalProviders)
	assert.NotNil(t, step.Check)
	// Note: We can't easily execute the check without a full terraform test framework
	_ = checkCalled // suppress unused warning
}

func TestBuildCrossVersionUpgradeSteps(t *testing.T) {
	cfg := CrossVersionUpgradeTestConfig{
		ProviderSource: "shin1ohno/rtx",
		OldVersion:     "1.0.0",
		OldConfig:      `resource "rtx_admin" "test" { name = "old" }`,
		NewConfig:      `resource "rtx_admin" "test" { name = "new" }`,
	}

	steps := BuildCrossVersionUpgradeSteps(cfg)
	require.Len(t, steps, 2)

	// First step should use external provider
	assert.NotNil(t, steps[0].ExternalProviders)
	assert.Contains(t, steps[0].Config, "old")

	// Second step should use new config
	assert.Contains(t, steps[1].Config, "new")
}

func TestBuildCrossVersionUpgradeSteps_WithEmptyPlanCheck(t *testing.T) {
	cfg := CrossVersionUpgradeTestConfig{
		ProviderSource:              "shin1ohno/rtx",
		OldVersion:                  "1.0.0",
		OldConfig:                   `resource "rtx_admin" "test" { name = "test" }`,
		ExpectEmptyPlanAfterUpgrade: true,
	}

	steps := BuildCrossVersionUpgradeSteps(cfg)
	require.Len(t, steps, 3)

	// First step should use external provider
	assert.NotNil(t, steps[0].ExternalProviders)

	// Second step should be PlanOnly
	assert.True(t, steps[1].PlanOnly)

	// Third step should apply the config
	assert.False(t, steps[2].PlanOnly)
}

func TestBuildCrossVersionUpgradeSteps_UsesOldConfigWhenNewEmpty(t *testing.T) {
	cfg := CrossVersionUpgradeTestConfig{
		ProviderSource: "shin1ohno/rtx",
		OldVersion:     "1.0.0",
		OldConfig:      `resource "rtx_admin" "test" { name = "same" }`,
		NewConfig:      "", // Empty - should reuse OldConfig
	}

	steps := BuildCrossVersionUpgradeSteps(cfg)
	require.Len(t, steps, 2)

	// Both steps should have the same config
	assert.Equal(t, cfg.OldConfig, steps[1].Config)
}

func TestAssertStateMigration(t *testing.T) {
	upgrader := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		result := make(map[string]interface{})
		for k, v := range rawState {
			result[k] = v
		}
		result["version"] = 2
		return result, nil
	}

	// This should not fail
	testFunc := AssertStateMigration(
		upgrader,
		map[string]interface{}{"key": "value"},
		map[string]interface{}{"key": "value", "version": 2},
	)
	testFunc(t)
}

func TestMustDeepCopyState(t *testing.T) {
	t.Run("nil state", func(t *testing.T) {
		result := MustDeepCopyState(nil)
		assert.Nil(t, result)
	})

	t.Run("simple state", func(t *testing.T) {
		original := map[string]interface{}{
			"string": "value",
			"int":    42,
			"bool":   true,
			"float":  3.14,
		}
		copied := MustDeepCopyState(original)

		assert.Equal(t, original, copied)

		// Modify copied, original should not change
		copied["string"] = "modified"
		assert.Equal(t, "value", original["string"])
	})

	t.Run("nested state", func(t *testing.T) {
		original := map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"list": []interface{}{"a", "b", "c"},
		}
		copied := MustDeepCopyState(original)

		assert.Equal(t, original, copied)

		// Modify nested map in copied, original should not change
		copied["nested"].(map[string]interface{})["key"] = "modified"
		assert.Equal(t, "value", original["nested"].(map[string]interface{})["key"])

		// Modify list in copied, original should not change
		copied["list"].([]interface{})[0] = "modified"
		assert.Equal(t, "a", original["list"].([]interface{})[0])
	})
}

func TestStateUpgradeChain(t *testing.T) {
	// Upgrader v0 -> v1: rename "name" to "username"
	upgraderV0toV1 := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		if rawState == nil {
			return nil, nil
		}
		result := make(map[string]interface{})
		for k, v := range rawState {
			if k == "name" {
				result["username"] = v
			} else {
				result[k] = v
			}
		}
		return result, nil
	}

	// Upgrader v1 -> v2: add "version" field
	upgraderV1toV2 := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		if rawState == nil {
			return nil, nil
		}
		result := make(map[string]interface{})
		for k, v := range rawState {
			result[k] = v
		}
		result["schema_version"] = 2
		return result, nil
	}

	initialState := map[string]interface{}{
		"name":  "admin",
		"other": "value",
	}

	expectedFinalState := map[string]interface{}{
		"username":       "admin",
		"other":          "value",
		"schema_version": 2,
	}

	StateUpgradeChain(t, initialState, expectedFinalState, upgraderV0toV1, upgraderV1toV2)
}

func TestDescribeStateMigration(t *testing.T) {
	description := DescribeStateMigration(0, 1, []string{
		"Renamed 'name' field to 'username'",
		"Added 'schema_version' field",
	})

	assert.Contains(t, description, "v0 -> v1")
	assert.Contains(t, description, "Renamed 'name' field to 'username'")
	assert.Contains(t, description, "Added 'schema_version' field")
}

func TestDescribeStateMigration_NoChanges(t *testing.T) {
	description := DescribeStateMigration(1, 2, []string{})
	assert.Contains(t, description, "(no changes documented)")
}
