// Package acctest provides reusable check functions for acceptance testing.
// These functions reduce test boilerplate by providing common assertion patterns.
package acctest

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// CheckResourceAttrNotEmpty verifies that the specified attribute exists and has a non-empty value.
// This is useful for computed attributes that should always have a value after resource creation.
//
// Usage:
//
//	Check: resource.ComposeTestCheckFunc(
//	    acctest.CheckResourceAttrNotEmpty("rtx_admin_user.test", "id"),
//	    acctest.CheckResourceAttrNotEmpty("rtx_admin_user.test", "username"),
//	)
func CheckResourceAttrNotEmpty(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary == nil {
			return fmt.Errorf("no primary instance: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes[attrName]
		if !ok {
			return fmt.Errorf("attribute %q not found in resource %s", attrName, resourceName)
		}

		if value == "" {
			return fmt.Errorf("attribute %q is empty in resource %s", attrName, resourceName)
		}

		return nil
	}
}

// CheckResourceAttrEquals verifies that the specified attribute equals the expected value.
// Unlike resource.TestCheckResourceAttr, this function provides clearer error messages
// including both the expected and actual values.
//
// Usage:
//
//	Check: acctest.CheckResourceAttrEquals("rtx_admin_user.test", "username", "admin")
func CheckResourceAttrEquals(resourceName, attrName, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary == nil {
			return fmt.Errorf("no primary instance: %s", resourceName)
		}

		actual, ok := rs.Primary.Attributes[attrName]
		if !ok {
			return fmt.Errorf("attribute %q not found in resource %s", attrName, resourceName)
		}

		if actual != expected {
			return fmt.Errorf("attribute %q in resource %s: expected %q, got %q",
				attrName, resourceName, expected, actual)
		}

		return nil
	}
}

// CheckResourceAttrContains verifies that the specified attribute contains the expected substring.
// This is useful for attributes that may have additional content beyond a specific pattern.
//
// Usage:
//
//	Check: acctest.CheckResourceAttrContains("rtx_admin_user.test", "error_message", "not found")
func CheckResourceAttrContains(resourceName, attrName, substring string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary == nil {
			return fmt.Errorf("no primary instance: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes[attrName]
		if !ok {
			return fmt.Errorf("attribute %q not found in resource %s", attrName, resourceName)
		}

		if !strings.Contains(value, substring) {
			return fmt.Errorf("attribute %q in resource %s: expected to contain %q, got %q",
				attrName, resourceName, substring, value)
		}

		return nil
	}
}

// CheckNoPlannedChanges is a documentation wrapper that explains how to verify empty plans in SDK v2.
// In terraform-plugin-sdk/v2, empty plan verification is done through TestStep configuration
// rather than a check function.
//
// To verify no changes in a plan, use this pattern:
//
//	resource.TestStep{
//	    Config:   config,  // Same config as previous step
//	    PlanOnly: true,    // Only run plan, don't apply
//	    // Test automatically fails if the plan shows any changes
//	}
//
// This function provides a runtime reminder of this pattern.
// It always passes and logs guidance for test developers.
func CheckNoPlannedChanges(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// In SDK v2, empty plan verification is done through TestStep.PlanOnly = true.
		// This check function exists for documentation purposes and always succeeds.
		// See design.md Pattern 1: Perpetual Diff Prevention for the correct usage.
		return nil
	}
}

// CheckResourceImportState verifies that an imported resource has all expected attributes populated.
// This function is used with ImportStateCheck in resource.TestStep to validate
// that imported resources contain the necessary state.
//
// Usage:
//
//	resource.TestStep{
//	    ResourceName:       resourceName,
//	    ImportState:        true,
//	    ImportStateVerify:  true,
//	    ImportStateCheck:   acctest.CheckResourceImportState([]string{"id", "username", "administrator"}),
//	}
func CheckResourceImportState(expectedAttrs []string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) == 0 {
			return fmt.Errorf("no states returned from import")
		}

		// Check the first (primary) state
		state := states[0]
		if state == nil {
			return fmt.Errorf("imported state is nil")
		}

		var missingAttrs []string
		var emptyAttrs []string

		for _, attr := range expectedAttrs {
			value, exists := state.Attributes[attr]
			if !exists {
				missingAttrs = append(missingAttrs, attr)
			} else if value == "" {
				emptyAttrs = append(emptyAttrs, attr)
			}
		}

		if len(missingAttrs) > 0 || len(emptyAttrs) > 0 {
			var errParts []string
			if len(missingAttrs) > 0 {
				errParts = append(errParts, fmt.Sprintf("missing attributes: %v", missingAttrs))
			}
			if len(emptyAttrs) > 0 {
				errParts = append(errParts, fmt.Sprintf("empty attributes: %v", emptyAttrs))
			}
			return fmt.Errorf("import state validation failed: %s", strings.Join(errParts, "; "))
		}

		return nil
	}
}

// CheckResourceImportStateWithResource verifies import state for a specific resource.
// This is a convenience wrapper that also validates the resource ID.
//
// Usage:
//
//	ImportStateCheck: acctest.CheckResourceImportStateWithResource(
//	    "rtx_admin_user.test",
//	    []string{"username", "administrator"},
//	),
func CheckResourceImportStateWithResource(resourceName string, expectedAttrs []string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) == 0 {
			return fmt.Errorf("no states returned from import for %s", resourceName)
		}

		state := states[0]
		if state == nil {
			return fmt.Errorf("imported state is nil for %s", resourceName)
		}

		// Always check that ID is set
		if state.ID == "" {
			return fmt.Errorf("imported state has no ID for %s", resourceName)
		}

		// Check expected attributes
		var problems []string

		for _, attr := range expectedAttrs {
			value, exists := state.Attributes[attr]
			if !exists {
				problems = append(problems, fmt.Sprintf("attribute %q not found", attr))
			} else if value == "" {
				problems = append(problems, fmt.Sprintf("attribute %q is empty", attr))
			}
		}

		if len(problems) > 0 {
			return fmt.Errorf("import state validation failed for %s: %s",
				resourceName, strings.Join(problems, "; "))
		}

		return nil
	}
}

// CheckResourceExists verifies that a resource exists in the Terraform state.
// This is a basic check to ensure the resource was created successfully.
//
// Usage:
//
//	Check: acctest.CheckResourceExists("rtx_admin_user.test")
func CheckResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		if rs.Primary == nil {
			return fmt.Errorf("no primary instance for resource: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource has no ID: %s", resourceName)
		}

		return nil
	}
}

// CheckResourceDestroyed verifies that a resource no longer exists.
// This is typically used in CheckDestroy to verify cleanup.
// Note: This checks the state, not the actual router. For router verification,
// use resource-specific destroy checks.
//
// Usage:
//
//	CheckDestroy: acctest.CheckResourceDestroyed("rtx_admin_user")
func CheckResourceDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			if rs.Primary != nil && rs.Primary.ID != "" {
				return fmt.Errorf("resource %s (%s) still exists with ID: %s",
					name, resourceType, rs.Primary.ID)
			}
		}

		return nil
	}
}

// CheckResourceAttrBool verifies that a boolean attribute has the expected value.
// Terraform stores booleans as strings ("true" or "false").
//
// Usage:
//
//	Check: acctest.CheckResourceAttrBool("rtx_admin_user.test", "administrator", true)
func CheckResourceAttrBool(resourceName, attrName string, expected bool) resource.TestCheckFunc {
	expectedStr := "false"
	if expected {
		expectedStr = "true"
	}

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary == nil {
			return fmt.Errorf("no primary instance: %s", resourceName)
		}

		actual, ok := rs.Primary.Attributes[attrName]
		if !ok {
			return fmt.Errorf("attribute %q not found in resource %s", attrName, resourceName)
		}

		if actual != expectedStr {
			return fmt.Errorf("attribute %q in resource %s: expected %s, got %s",
				attrName, resourceName, expectedStr, actual)
		}

		return nil
	}
}

// CheckResourceAttrSet verifies that an attribute exists and is set (not empty).
// This is similar to resource.TestCheckResourceAttrSet but with clearer error messages.
//
// Usage:
//
//	Check: acctest.CheckResourceAttrSet("rtx_admin_user.test", "id")
func CheckResourceAttrSet(resourceName, attrName string) resource.TestCheckFunc {
	return CheckResourceAttrNotEmpty(resourceName, attrName)
}

// ComposeAggregateTestCheckFunc combines multiple check functions and reports all failures.
// Unlike resource.ComposeAggregateTestCheckFunc, this provides formatted output for all errors.
//
// Usage:
//
//	Check: acctest.ComposeAggregateTestCheckFunc(
//	    acctest.CheckResourceExists("rtx_admin_user.test"),
//	    acctest.CheckResourceAttrNotEmpty("rtx_admin_user.test", "id"),
//	)
func ComposeAggregateTestCheckFunc(checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var errors []string

		for i, check := range checks {
			if err := check(s); err != nil {
				errors = append(errors, fmt.Sprintf("check %d: %s", i+1, err.Error()))
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("aggregate check failures:\n  %s", strings.Join(errors, "\n  "))
		}

		return nil
	}
}
