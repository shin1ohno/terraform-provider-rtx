package access_list_ipv6_apply

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &AccessListIPv6ApplyResource{}
	_ resource.ResourceWithImportState = &AccessListIPv6ApplyResource{}
)

// NewAccessListIPv6ApplyResource creates a new IPv6 access list apply resource.
func NewAccessListIPv6ApplyResource() resource.Resource {
	return &AccessListIPv6ApplyResource{}
}

// AccessListIPv6ApplyResource defines the resource implementation.
type AccessListIPv6ApplyResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListIPv6ApplyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_ipv6_apply"
}

// Schema defines the schema for the resource.
func (r *AccessListIPv6ApplyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Applies IPv6 access list filters to an interface. " +
			"This resource manages the binding of IPv6 ACL filter sequences to a specific interface and direction. " +
			"Use this resource when you need to apply filters from multiple ACLs to the same interface, " +
			"or when you want to manage interface filter bindings separately from the ACL definitions.",
		Attributes: map[string]schema.Attribute{
			"access_list": schema.StringAttribute{
				Description: "Name of the IPv6 access list to apply. This is used for tracking purposes.",
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
					&caseInsensitivePlanModifier{},
				},
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("in", "out"),
				},
			},
			"filter_ids": schema.ListAttribute{
				Description: "List of filter IDs to apply in order. At least one filter ID must be specified.",
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

// Configure adds the provider configured client to the resource.
func (r *AccessListIPv6ApplyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListIPv6ApplyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListIPv6ApplyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normalize direction to lowercase
	data.Direction = types.StringValue(strings.ToLower(data.Direction.ValueString()))

	id := data.GetResourceID()
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", data.Interface.ValueString()).
		Str("direction", data.Direction.ValueString()).
		Str("access_list", data.AccessList.ValueString()).
		Msg("Creating IPv6 access list apply")

	filterIDs := data.GetFilterIDsAsInts()
	if len(filterIDs) == 0 {
		resp.Diagnostics.AddError(
			"Invalid filter_ids",
			"filter_ids is required: at least one filter ID must be specified",
		)
		return
	}

	err := r.client.ApplyIPv6FiltersToInterface(ctx, data.Interface.ValueString(), data.Direction.ValueString(), filterIDs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to apply IPv6 filters",
			fmt.Sprintf("Could not apply IPv6 filters to interface %s %s: %v",
				data.Interface.ValueString(), data.Direction.ValueString(), err),
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
func (r *AccessListIPv6ApplyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListIPv6ApplyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If resource was removed (interface and direction are null), remove from state
	if data.Interface.IsNull() && data.Direction.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the applied filters from the router.
func (r *AccessListIPv6ApplyResource) read(ctx context.Context, data *AccessListIPv6ApplyModel, diagnostics *diag.Diagnostics) {
	iface := data.Interface.ValueString()
	direction := data.Direction.ValueString()
	id := data.GetResourceID()

	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Reading IPv6 access list apply")

	filterIDs, err := r.client.GetIPv6InterfaceFilters(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().
				Str("resource", "rtx_access_list_ipv6_apply").
				Str("interface", iface).
				Str("direction", direction).
				Msg("IPv6 filter apply not found, removing from state")
			data.Interface = types.StringNull()
			data.Direction = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read IPv6 filter apply",
			fmt.Sprintf("Could not read IPv6 filter apply for %s %s: %v", iface, direction, err))
		return
	}

	if len(filterIDs) == 0 {
		logger.Warn().
			Str("resource", "rtx_access_list_ipv6_apply").
			Str("interface", iface).
			Str("direction", direction).
			Msg("No IPv6 filters applied, removing from state")
		data.Interface = types.StringNull()
		data.Direction = types.StringNull()
		return
	}

	data.Interface = types.StringValue(iface)
	data.Direction = types.StringValue(direction)
	data.SetFilterIDsFromInts(filterIDs)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListIPv6ApplyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListIPv6ApplyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normalize direction to lowercase
	data.Direction = types.StringValue(strings.ToLower(data.Direction.ValueString()))

	id := data.GetResourceID()
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", data.Interface.ValueString()).
		Str("direction", data.Direction.ValueString()).
		Str("access_list", data.AccessList.ValueString()).
		Msg("Updating IPv6 access list apply")

	filterIDs := data.GetFilterIDsAsInts()
	if len(filterIDs) == 0 {
		resp.Diagnostics.AddError(
			"Invalid filter_ids",
			"filter_ids is required: at least one filter ID must be specified",
		)
		return
	}

	err := r.client.ApplyIPv6FiltersToInterface(ctx, data.Interface.ValueString(), data.Direction.ValueString(), filterIDs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IPv6 filters",
			fmt.Sprintf("Could not update IPv6 filters on interface %s %s: %v",
				data.Interface.ValueString(), data.Direction.ValueString(), err),
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
func (r *AccessListIPv6ApplyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListIPv6ApplyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()
	direction := data.Direction.ValueString()
	id := data.GetResourceID()

	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Deleting IPv6 access list apply")

	err := r.client.RemoveIPv6FiltersFromInterface(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to remove IPv6 filters",
			fmt.Sprintf("Could not remove IPv6 filters from interface %s %s: %v", iface, direction, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListIPv6ApplyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: interface:direction
	id := req.ID
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Invalid import ID format: %s, expected 'interface:direction' (e.g., 'lan1:in')", id),
		)
		return
	}

	iface := parts[0]
	direction := strings.ToLower(parts[1])

	if direction != "in" && direction != "out" {
		resp.Diagnostics.AddError(
			"Invalid direction",
			fmt.Sprintf("Invalid direction: %s, must be 'in' or 'out'", parts[1]),
		)
		return
	}

	data := AccessListIPv6ApplyModel{
		AccessList: types.StringValue("imported"),
		Interface:  types.StringValue(iface),
		Direction:  types.StringValue(direction),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// caseInsensitivePlanModifier normalizes direction to lowercase.
type caseInsensitivePlanModifier struct{}

func (m *caseInsensitivePlanModifier) Description(ctx context.Context) string {
	return "Normalizes string value to lowercase for case-insensitive comparison"
}

func (m *caseInsensitivePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *caseInsensitivePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = types.StringValue(strings.ToLower(req.PlanValue.ValueString()))
}
