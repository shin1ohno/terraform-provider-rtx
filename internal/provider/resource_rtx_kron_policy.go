package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXKronPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a kron policy (command list) on RTX routers. A kron policy defines a set of commands that can be executed by a kron schedule.",
		CreateContext: resourceRTXKronPolicyCreate,
		ReadContext:   resourceRTXKronPolicyRead,
		UpdateContext: resourceRTXKronPolicyUpdate,
		DeleteContext: resourceRTXKronPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXKronPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The name of the kron policy. Must start with a letter and contain only letters, numbers, underscores, and hyphens.",
				ValidateFunc: validateKronPolicyName,
			},
			"command_lines": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of commands to execute in order when the policy is triggered.",
				MinItems:    1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func resourceRTXKronPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_policy", d.Id())
	policy := buildKronPolicyFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_policy").Msgf("Creating kron policy: %+v", policy)

	err := apiClient.client.CreateKronPolicy(ctx, policy)
	if err != nil {
		return diag.Errorf("Failed to create kron policy: %v", err)
	}

	d.SetId(policy.Name)

	return resourceRTXKronPolicyRead(ctx, d, meta)
}

func resourceRTXKronPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Note: RTX routers don't have native kron policy support.
	// Policies are managed at the Terraform level only.
	// We simply read back the values from the state.

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_policy").Msgf("Reading kron policy: %s", d.Id())

	// For RTX, the policy is stored locally in Terraform state
	// No actual device query needed since RTX doesn't have native kron policy
	// Just validate the state is consistent
	if d.Id() == "" {
		return nil
	}

	return nil
}

func resourceRTXKronPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_policy", d.Id())
	policy := buildKronPolicyFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_policy").Msgf("Updating kron policy: %+v", policy)

	err := apiClient.client.UpdateKronPolicy(ctx, policy)
	if err != nil {
		return diag.Errorf("Failed to update kron policy: %v", err)
	}

	return resourceRTXKronPolicyRead(ctx, d, meta)
}

func resourceRTXKronPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_policy", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_policy").Msgf("Deleting kron policy: %s", name)

	err := apiClient.client.DeleteKronPolicy(ctx, name)
	if err != nil {
		return diag.Errorf("Failed to delete kron policy: %v", err)
	}

	return nil
}

func resourceRTXKronPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_policy").Msgf("Importing kron policy: %s", name)

	// Validate the name format
	if err := validateKronPolicyNameValue(name); err != nil {
		return nil, fmt.Errorf("invalid kron policy name for import: %v", err)
	}

	d.SetId(name)
	d.Set("name", name)
	// Note: command_lines cannot be imported from device as RTX doesn't store policies
	// User must update the resource after import with the correct commands
	d.Set("command_lines", []string{})

	return []*schema.ResourceData{d}, nil
}

// buildKronPolicyFromResourceData creates a KronPolicy from Terraform resource data
func buildKronPolicyFromResourceData(d *schema.ResourceData) client.KronPolicy {
	policy := client.KronPolicy{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("command_lines"); ok {
		commandsRaw := v.([]interface{})
		commands := make([]string, len(commandsRaw))
		for i, cmd := range commandsRaw {
			commands[i] = cmd.(string)
		}
		policy.Commands = commands
	}

	return policy
}

// validateKronPolicyName validates the kron policy name format
func validateKronPolicyName(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if err := validateKronPolicyNameValue(value); err != nil {
		return nil, []error{fmt.Errorf("%q: %v", k, err)}
	}

	return nil, nil
}

// validateKronPolicyNameValue validates the kron policy name
func validateKronPolicyNameValue(name string) error {
	if name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	// Must start with a letter and contain only letters, numbers, underscores, and hyphens
	namePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !namePattern.MatchString(name) {
		return fmt.Errorf("policy name must start with a letter and contain only letters, numbers, underscores, and hyphens, got %q", name)
	}

	// Max length check
	if len(name) > 64 {
		return fmt.Errorf("policy name must be 64 characters or less, got %d", len(name))
	}

	return nil
}
