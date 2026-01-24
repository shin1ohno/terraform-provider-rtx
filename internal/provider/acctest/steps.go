// Package acctest provides test step builder functions for standardized
// acceptance test construction in Terraform provider tests.
package acctest

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// BasicCreateStep returns a TestStep configured for basic resource creation.
// This is typically the first step in an acceptance test that creates a new resource.
//
// Parameters:
//   - config: The Terraform HCL configuration to apply
//   - checks: Optional check functions to verify the resource state after creation
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic(), resource.TestCheckResourceAttr(...)),
//		},
//	})
func BasicCreateStep(config string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := resource.TestStep{
		Config: config,
	}

	if len(checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}

	return step
}

// UpdateStep returns a TestStep configured for resource updates.
// Use this after a BasicCreateStep to test updating an existing resource.
// The config should contain modified attribute values compared to the create config.
//
// Parameters:
//   - config: The updated Terraform HCL configuration to apply
//   - checks: Optional check functions to verify the resource state after update
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.UpdateStep(testAccConfig_updated(), resource.TestCheckResourceAttr(...)),
//		},
//	})
func UpdateStep(config string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := resource.TestStep{
		Config: config,
	}

	if len(checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}

	return step
}

// ImportStep returns a TestStep configured for import testing.
// This step imports an existing resource and verifies that the imported state
// matches the state created by Terraform.
//
// Parameters:
//   - resourceName: The resource address to import (e.g., "rtx_admin_user.test")
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.ImportStep("rtx_admin_user.test"),
//		},
//	})
func ImportStep(resourceName string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      resourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

// ImportStepWithIgnore returns a TestStep configured for import testing
// with specific attributes ignored during verification.
// Use this when certain attributes (like passwords) cannot be imported.
//
// Parameters:
//   - resourceName: The resource address to import (e.g., "rtx_admin_user.test")
//   - ignoreFields: List of attribute names to ignore during import verification
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.ImportStepWithIgnore("rtx_admin_user.test", "password"),
//		},
//	})
func ImportStepWithIgnore(resourceName string, ignoreFields ...string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            resourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: ignoreFields,
	}
}

// ImportStepWithID returns a TestStep configured for import testing
// with a specific import ID. Use this when the import ID differs from
// the resource's Terraform ID.
//
// Parameters:
//   - resourceName: The resource address to import (e.g., "rtx_admin_user.test")
//   - importID: The ID to use for importing the resource
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.ImportStepWithID("rtx_admin_user.test", "admin_username"),
//		},
//	})
func ImportStepWithID(resourceName, importID string) resource.TestStep {
	return resource.TestStep{
		ResourceName:  resourceName,
		ImportState:   true,
		ImportStateId: importID,
	}
}

// ImportStepOptions provides a fully configurable import test step.
// Use this when you need more control over the import test behavior.
type ImportStepOptions struct {
	// ResourceName is the resource address to import (required)
	ResourceName string

	// ImportID is the ID to use for import. If empty, uses the resource's ID from state.
	ImportID string

	// IgnoreFields are attribute names to skip during import verification
	IgnoreFields []string

	// Verify enables ImportStateVerify to compare imported state with created state
	Verify bool

	// Checks are additional check functions to run after import
	Checks []resource.TestCheckFunc
}

// ImportStepWithOpts returns a TestStep configured with the given options.
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.ImportStepWithOptions(acctest.ImportStepOptions{
//				ResourceName: "rtx_admin_user.test",
//				IgnoreFields: []string{"password"},
//				Verify:       true,
//			}),
//		},
//	})
func ImportStepWithOpts(opts ImportStepOptions) resource.TestStep {
	step := resource.TestStep{
		ResourceName:            opts.ResourceName,
		ImportState:             true,
		ImportStateVerify:       opts.Verify,
		ImportStateVerifyIgnore: opts.IgnoreFields,
	}

	if opts.ImportID != "" {
		step.ImportStateId = opts.ImportID
	}

	if len(opts.Checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(opts.Checks...)
	}

	return step
}

// NoChangeStep returns a TestStep that verifies no changes are planned.
// This is used to detect perpetual diff issues where Terraform plans changes
// even when the configuration and state should be identical.
//
// The step uses PlanOnly mode, which will cause the test to fail if any
// changes are planned. This is a key test for ensuring resources don't
// have "perpetual diff" bugs.
//
// Parameters:
//   - config: The Terraform HCL configuration to verify (should match current state)
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.NoChangeStep(testAccConfig_basic()),  // Should plan no changes
//		},
//	})
func NoChangeStep(config string) resource.TestStep {
	return resource.TestStep{
		Config:   config,
		PlanOnly: true,
		// In terraform-plugin-sdk/v2, PlanOnly mode automatically fails
		// if there are any planned changes (non-empty plan).
	}
}

// NoChangeStepWithChecks returns a TestStep that verifies no changes are planned
// and also runs additional check functions. This combines perpetual diff detection
// with state verification.
//
// Parameters:
//   - config: The Terraform HCL configuration to verify
//   - checks: Check functions to verify the current state
//
// Example:
//
//	resource.Test(t, resource.TestCase{
//		Steps: []resource.TestStep{
//			acctest.BasicCreateStep(testAccConfig_basic()),
//			acctest.NoChangeStepWithChecks(testAccConfig_basic(),
//				resource.TestCheckResourceAttr("rtx_admin_user.test", "username", "admin"),
//			),
//		},
//	})
func NoChangeStepWithChecks(config string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := resource.TestStep{
		Config:   config,
		PlanOnly: true,
	}

	if len(checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}

	return step
}

// DestroyStep returns a TestStep that destroys all resources.
// This can be used as an explicit step to verify destroy behavior,
// although the test framework automatically destroys resources after tests.
//
// Parameters:
//   - config: An empty or minimal configuration that will cause resource deletion
func DestroyStep() resource.TestStep {
	return resource.TestStep{
		Config:  "",
		Destroy: true,
	}
}

// RefreshOnlyStep returns a TestStep that only refreshes the state without applying.
// This is useful for testing that state refresh correctly reads the current
// resource state from the remote system.
//
// Parameters:
//   - config: The Terraform HCL configuration
//   - checks: Check functions to verify the refreshed state
func RefreshOnlyStep(config string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := resource.TestStep{
		Config:       config,
		RefreshState: true,
	}

	if len(checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}

	return step
}

// ExpectNonEmptyPlanStep returns a TestStep that expects a non-empty plan.
// This is useful for testing that certain configuration changes are detected
// as requiring updates.
//
// Parameters:
//   - config: The Terraform HCL configuration that should cause planned changes
//   - checks: Optional check functions to verify the planned changes
func ExpectNonEmptyPlanStep(config string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := resource.TestStep{
		Config:             config,
		PlanOnly:           true,
		ExpectNonEmptyPlan: true,
	}

	if len(checks) > 0 {
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}

	return step
}

// StepSequence is a builder for creating a sequence of test steps.
// It provides a fluent interface for building complex test scenarios.
type StepSequence struct {
	steps []resource.TestStep
}

// NewStepSequence creates a new StepSequence builder.
func NewStepSequence() *StepSequence {
	return &StepSequence{
		steps: make([]resource.TestStep, 0),
	}
}

// Create adds a BasicCreateStep to the sequence.
func (s *StepSequence) Create(config string, checks ...resource.TestCheckFunc) *StepSequence {
	s.steps = append(s.steps, BasicCreateStep(config, checks...))
	return s
}

// Update adds an UpdateStep to the sequence.
func (s *StepSequence) Update(config string, checks ...resource.TestCheckFunc) *StepSequence {
	s.steps = append(s.steps, UpdateStep(config, checks...))
	return s
}

// Import adds an ImportStep to the sequence.
func (s *StepSequence) Import(resourceName string) *StepSequence {
	s.steps = append(s.steps, ImportStep(resourceName))
	return s
}

// ImportIgnore adds an ImportStepWithIgnore to the sequence.
func (s *StepSequence) ImportIgnore(resourceName string, ignoreFields ...string) *StepSequence {
	s.steps = append(s.steps, ImportStepWithIgnore(resourceName, ignoreFields...))
	return s
}

// NoChange adds a NoChangeStep to the sequence.
func (s *StepSequence) NoChange(config string) *StepSequence {
	s.steps = append(s.steps, NoChangeStep(config))
	return s
}

// Custom adds a custom TestStep to the sequence.
func (s *StepSequence) Custom(step resource.TestStep) *StepSequence {
	s.steps = append(s.steps, step)
	return s
}

// Build returns the accumulated test steps.
func (s *StepSequence) Build() []resource.TestStep {
	return s.steps
}
