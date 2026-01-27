package kron_policy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &KronPolicyResource{}
	_ resource.ResourceWithImportState = &KronPolicyResource{}
)

// NewKronPolicyResource creates a new kron policy resource.
func NewKronPolicyResource() resource.Resource {
	return &KronPolicyResource{}
}

// KronPolicyResource defines the resource implementation.
type KronPolicyResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *KronPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kron_policy"
}

// Schema defines the schema for the resource.
func (r *KronPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a kron policy (command list) on RTX routers. A kron policy defines a set of commands that can be executed by a kron schedule.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the kron policy. Must start with a letter and contain only letters, numbers, underscores, and hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
						"must start with a letter and contain only letters, numbers, underscores, and hyphens",
					),
				},
			},
			"command_lines": schema.ListAttribute{
				Description: "List of commands to execute in order when the policy is triggered.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *KronPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*fwhelpers.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *fwhelpers.ProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *KronPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KronPolicyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_kron_policy", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	policy := data.ToClient()
	logger.Debug().Str("resource", "rtx_kron_policy").Msgf("Creating kron policy: %+v", policy)

	if err := r.client.CreateKronPolicy(ctx, policy); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create kron policy",
			fmt.Sprintf("Could not create kron policy: %v", err),
		)
		return
	}

	// For kron policy, state is managed by Terraform (no read-back from device)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *KronPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KronPolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: RTX routers don't have native kron policy support.
	// Policies are managed at the Terraform level only.
	// We simply read back the values from the state.

	ctx = logging.WithResource(ctx, "rtx_kron_policy", data.Name.ValueString())
	logger := logging.FromContext(ctx)
	logger.Debug().Str("resource", "rtx_kron_policy").Msgf("Reading kron policy: %s", data.Name.ValueString())

	// For RTX, the policy is stored locally in Terraform state
	// No actual device query needed since RTX doesn't have native kron policy
	// Just validate the state is consistent
	if data.Name.IsNull() || data.Name.ValueString() == "" {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *KronPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KronPolicyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_kron_policy", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	policy := data.ToClient()
	logger.Debug().Str("resource", "rtx_kron_policy").Msgf("Updating kron policy: %+v", policy)

	if err := r.client.UpdateKronPolicy(ctx, policy); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update kron policy",
			fmt.Sprintf("Could not update kron policy: %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *KronPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KronPolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_kron_policy", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_kron_policy").Msgf("Deleting kron policy: %s", name)

	if err := r.client.DeleteKronPolicy(ctx, name); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete kron policy",
			fmt.Sprintf("Could not delete kron policy %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *KronPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	ctx = logging.WithResource(ctx, "rtx_kron_policy", name)
	logger := logging.FromContext(ctx)
	logger.Debug().Str("resource", "rtx_kron_policy").Msgf("Importing kron policy: %s", name)

	// Validate the name format
	if err := validateKronPolicyNameValue(name); err != nil {
		resp.Diagnostics.AddError(
			"Invalid kron policy name for import",
			fmt.Sprintf("Invalid kron policy name: %v", err),
		)
		return
	}

	// Set the name attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)

	// Note: command_lines cannot be imported from device as RTX doesn't store policies
	// User must update the resource after import with the correct commands
	// Set an empty list for command_lines
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("command_lines"), []string{})...)
}

// validateKronPolicyNameValue validates the kron policy name.
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
