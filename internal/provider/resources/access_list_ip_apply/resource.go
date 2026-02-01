package access_list_ip_apply

import (
	"context"
	"fmt"
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
	_ resource.Resource                = &AccessListIPApplyResource{}
	_ resource.ResourceWithImportState = &AccessListIPApplyResource{}
)

// NewAccessListIPApplyResource creates a new access list IP apply resource.
func NewAccessListIPApplyResource() resource.Resource {
	return &AccessListIPApplyResource{}
}

// AccessListIPApplyResource defines the resource implementation.
type AccessListIPApplyResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListIPApplyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_ip_apply"
}

// Schema defines the schema for the resource.
func (r *AccessListIPApplyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Applies IP access list filters to an interface. " +
			"This resource manages the binding of IP ACL filter sequences to a specific interface and direction. " +
			"Use this resource when you need to apply filters from multiple ACLs to the same interface, " +
			"or when you want to manage interface filter bindings separately from the ACL definitions.",
		Attributes: map[string]schema.Attribute{
			"access_list": schema.StringAttribute{
				Description: "Name of the IP access list to apply. This is used for tracking purposes.",
				Required:    true,
			},
			"interface": schema.StringAttribute{
				Description: "Interface name to apply the filters to (e.g., lan1, pp1, tunnel1)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Description: "Traffic direction: 'in' for incoming traffic, 'out' for outgoing traffic",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					&caseInsensitiveStringPlanModifier{},
				},
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("in", "out"),
				},
			},
			"sequences": schema.ListAttribute{
				Description: "List of sequence numbers to apply in order. At least one sequence must be specified.",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.ValueInt64sAre(
						int64validator.AtLeast(1),
					),
				},
			},
		},
	}
}

// caseInsensitiveStringPlanModifier normalizes direction to lowercase.
type caseInsensitiveStringPlanModifier struct{}

func (m *caseInsensitiveStringPlanModifier) Description(ctx context.Context) string {
	return "Normalizes string value to lowercase"
}

func (m *caseInsensitiveStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *caseInsensitiveStringPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	resp.PlanValue = types.StringValue(strings.ToLower(req.PlanValue.ValueString()))
}

// Configure adds the provider configured client to the resource.
func (r *AccessListIPApplyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListIPApplyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListIPApplyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()
	direction := strings.ToLower(data.Direction.ValueString())
	aclName := data.AccessList.ValueString()

	// Build resource ID
	id := fmt.Sprintf("%s:%s", iface, direction)

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ip_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Creating IP access list apply")

	// Get sequences
	sequences := data.GetSequencesAsInts()

	if len(sequences) == 0 {
		resp.Diagnostics.AddError(
			"Invalid sequences",
			"sequences is required: at least one sequence must be specified",
		)
		return
	}

	// Apply filters to interface
	if err := r.client.ApplyIPFiltersToInterface(ctx, iface, direction, sequences); err != nil {
		resp.Diagnostics.AddError(
			"Failed to apply IP filters",
			fmt.Sprintf("Could not apply IP filters to interface %s %s: %v", iface, direction, err),
		)
		return
	}

	// Normalize direction in the data
	data.Direction = types.StringValue(direction)

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccessListIPApplyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListIPApplyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was removed
	if data.Interface.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the apply state from the router.
func (r *AccessListIPApplyResource) read(ctx context.Context, data *AccessListIPApplyModel, diagnostics *diag.Diagnostics) {
	iface := data.Interface.ValueString()
	direction := strings.ToLower(data.Direction.ValueString())
	id := fmt.Sprintf("%s:%s", iface, direction)

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ip_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Reading IP access list apply")

	// Get current filter IDs from router
	filterIDs, err := r.client.GetIPInterfaceFilters(ctx, iface, direction)
	if err != nil {
		// If not found, the resource has been removed
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().
				Str("resource", "rtx_access_list_ip_apply").
				Str("interface", iface).
				Str("direction", direction).
				Msg("IP filter apply not found, removing from state")
			data.Interface = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read IP filter apply", fmt.Sprintf("Could not read IP filter apply for %s %s: %v", iface, direction, err))
		return
	}

	// If no filters are applied, resource doesn't exist
	if len(filterIDs) == 0 {
		logger.Warn().
			Str("resource", "rtx_access_list_ip_apply").
			Str("interface", iface).
			Str("direction", direction).
			Msg("No IP filters applied, removing from state")
		data.Interface = types.StringNull()
		return
	}

	// Update state
	data.Interface = types.StringValue(iface)
	data.Direction = types.StringValue(direction)
	data.SetSequencesFromInts(filterIDs)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListIPApplyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListIPApplyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()
	direction := strings.ToLower(data.Direction.ValueString())
	aclName := data.AccessList.ValueString()
	id := fmt.Sprintf("%s:%s", iface, direction)

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ip_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Updating IP access list apply")

	// Get sequences
	sequences := data.GetSequencesAsInts()

	if len(sequences) == 0 {
		resp.Diagnostics.AddError(
			"Invalid sequences",
			"sequences is required: at least one sequence must be specified",
		)
		return
	}

	// Apply filters to interface (this will replace existing filters)
	if err := r.client.ApplyIPFiltersToInterface(ctx, iface, direction, sequences); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IP filters",
			fmt.Sprintf("Could not update IP filters on interface %s %s: %v", iface, direction, err),
		)
		return
	}

	// Normalize direction in the data
	data.Direction = types.StringValue(direction)

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccessListIPApplyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListIPApplyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()
	direction := strings.ToLower(data.Direction.ValueString())
	id := fmt.Sprintf("%s:%s", iface, direction)

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ip_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Deleting IP access list apply")

	// Remove filters from interface
	if err := r.client.RemoveIPFiltersFromInterface(ctx, iface, direction); err != nil {
		// Ignore "not found" errors
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to remove IP filters",
			fmt.Sprintf("Could not remove IP filters from interface %s %s: %v", iface, direction, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListIPApplyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: interface:direction
	id := req.ID
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Invalid import ID format: %s, expected 'interface:direction' (e.g., 'lan1:in')", id),
		)
		return
	}

	iface := parts[0]
	direction := strings.ToLower(parts[1])

	// Validate direction
	if direction != "in" && direction != "out" {
		resp.Diagnostics.AddError(
			"Invalid direction",
			fmt.Sprintf("Invalid direction: %s, must be 'in' or 'out'", direction),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("interface"), iface)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("direction"), direction)...)
	// access_list will need to be set manually after import
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access_list"), "imported")...)
}
