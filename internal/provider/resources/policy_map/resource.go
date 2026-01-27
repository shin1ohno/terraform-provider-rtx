package policy_map

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &PolicyMapResource{}
	_ resource.ResourceWithImportState = &PolicyMapResource{}
	_ resource.ResourceWithModifyPlan  = &PolicyMapResource{}
)

// NewPolicyMapResource creates a new policy map resource.
func NewPolicyMapResource() resource.Resource {
	return &PolicyMapResource{}
}

// PolicyMapResource defines the resource implementation.
type PolicyMapResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *PolicyMapResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_map"
}

// Schema defines the schema for the resource.
func (r *PolicyMapResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages QoS policy-map configurations on RTX routers. Policy-maps define actions to take on classified traffic.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The policy-map name. Must start with a letter and contain only alphanumeric characters, underscores, and hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
						"must start with a letter and contain only alphanumeric characters, underscores, and hyphens",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"class": schema.ListNestedBlock{
				Description: "List of class definitions within this policy-map",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Class name (references a class-map)",
							Required:    true,
						},
						"priority": schema.StringAttribute{
							Description: "Priority level: 'high', 'normal', or 'low'",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("high", "normal", "low"),
							},
						},
						"bandwidth_percent": schema.Int64Attribute{
							Description: "Bandwidth percentage allocation (1-100)",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
						},
						"police_cir": schema.Int64Attribute{
							Description: "Committed Information Rate in bps for policing",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
						"queue_limit": schema.Int64Attribute{
							Description: "Queue depth limit",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *PolicyMapResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan implements custom plan modification to validate total bandwidth.
func (r *PolicyMapResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var data PolicyMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate total bandwidth_percent doesn't exceed 100%
	if !data.Classes.IsNull() && !data.Classes.IsUnknown() {
		elements := data.Classes.Elements()
		var totalBandwidth int64

		for _, elem := range elements {
			objVal := elem.(types.Object)
			if objVal.IsNull() || objVal.IsUnknown() {
				continue
			}
			attrs := objVal.Attributes()
			if bwAttr, ok := attrs["bandwidth_percent"]; ok {
				if bwVal, ok := bwAttr.(types.Int64); ok && !bwVal.IsNull() && !bwVal.IsUnknown() {
					totalBandwidth += bwVal.ValueInt64()
				}
			}
		}

		if totalBandwidth > 100 {
			resp.Diagnostics.AddError(
				"Invalid bandwidth allocation",
				fmt.Sprintf("Total bandwidth_percent across all classes (%d%%) exceeds 100%%", totalBandwidth),
			)
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *PolicyMapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PolicyMapModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_policy_map", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	pm := data.ToClient()
	logger.Debug().Str("resource", "rtx_policy_map").Msgf("Creating policy-map: %s", pm.Name)

	if err := r.client.CreatePolicyMap(ctx, pm); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create policy-map",
			fmt.Sprintf("Could not create policy-map: %v", err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *PolicyMapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PolicyMapModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the resource no longer exists, remove from state
	if data.Name.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the policy map from the router.
func (r *PolicyMapResource) read(ctx context.Context, data *PolicyMapModel, diagnostics *diag.Diagnostics) {
	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_policy_map", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_policy_map").Msgf("Reading policy-map: %s", name)

	pm, err := r.client.GetPolicyMap(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_policy_map").Msgf("Policy-map %s not found", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read policy-map", fmt.Sprintf("Could not read policy-map %s: %v", name, err))
		return
	}

	data.FromClient(pm)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PolicyMapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PolicyMapModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_policy_map", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	pm := data.ToClient()
	logger.Debug().Str("resource", "rtx_policy_map").Msgf("Updating policy-map: %s", pm.Name)

	if err := r.client.UpdatePolicyMap(ctx, pm); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update policy-map",
			fmt.Sprintf("Could not update policy-map: %v", err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *PolicyMapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PolicyMapModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_policy_map", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_policy_map").Msgf("Deleting policy-map: %s", name)

	if err := r.client.DeletePolicyMap(ctx, name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete policy-map",
			fmt.Sprintf("Could not delete policy-map %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *PolicyMapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
