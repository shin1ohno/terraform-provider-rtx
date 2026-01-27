package access_list_mac_apply

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
	_ resource.Resource                = &AccessListMACApplyResource{}
	_ resource.ResourceWithImportState = &AccessListMACApplyResource{}
)

// NewAccessListMACApplyResource creates a new access list MAC apply resource.
func NewAccessListMACApplyResource() resource.Resource {
	return &AccessListMACApplyResource{}
}

// AccessListMACApplyResource defines the resource implementation.
type AccessListMACApplyResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListMACApplyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_mac_apply"
}

// Schema defines the schema for the resource.
func (r *AccessListMACApplyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Applies MAC access list filters to an interface. " +
			"This resource manages the binding of MAC ACL filter sequences to a specific interface and direction. " +
			"Use this resource when you need to apply filters from multiple ACLs to the same interface, " +
			"or when you want to manage interface filter bindings separately from the ACL definitions. " +
			"Note: MAC filters are only supported on Ethernet interfaces (lan, bridge). " +
			"PP and Tunnel interfaces are not supported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier in the format 'interface:direction'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_list": schema.StringAttribute{
				Description: "Name of the MAC access list to apply. This is used for tracking purposes.",
				Required:    true,
			},
			"interface": schema.StringAttribute{
				Description: "Interface name to apply the filters to (e.g., lan1, bridge1). PP and Tunnel interfaces are not supported for MAC filters.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Description: "Traffic direction: 'in' for incoming traffic, 'out' for outgoing traffic.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("in", "out"),
				},
			},
			"filter_ids": schema.ListAttribute{
				Description: "List of filter IDs to apply in order. At least one filter ID must be specified.",
				Required:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueInt64sAre(
						int64validator.AtLeast(1),
					),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListMACApplyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig performs custom validation on the configuration.
func (r *AccessListMACApplyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data AccessListMACApplyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate interface type if known
	if !data.Interface.IsNull() && !data.Interface.IsUnknown() {
		if err := validateMACInterfaceType(data.Interface.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("interface"),
				"Invalid Interface Type",
				err.Error(),
			)
		}
	}
}

// validateMACInterfaceType validates that the interface type supports MAC filters.
// MAC filters are only supported on Ethernet interfaces (lan, bridge).
// PP and Tunnel interfaces are NOT supported.
func validateMACInterfaceType(iface string) error {
	ifaceLower := strings.ToLower(iface)

	// Check for unsupported interface types
	if strings.HasPrefix(ifaceLower, "pp") {
		// Ensure it's actually a PP interface (pp followed by a number)
		rest := ifaceLower[2:]
		if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			return fmt.Errorf("MAC filters are not supported on PP interfaces: %s. Use lan or bridge interfaces instead", iface)
		}
	}

	if strings.HasPrefix(ifaceLower, "tunnel") {
		// Ensure it's actually a tunnel interface (tunnel followed by a number)
		rest := ifaceLower[6:]
		if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			return fmt.Errorf("MAC filters are not supported on Tunnel interfaces: %s. Use lan or bridge interfaces instead", iface)
		}
	}

	// Validate it's a supported interface type (lan or bridge)
	if !strings.HasPrefix(ifaceLower, "lan") && !strings.HasPrefix(ifaceLower, "bridge") {
		return fmt.Errorf("MAC filters are only supported on Ethernet interfaces (lan, bridge), not %s", iface)
	}

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *AccessListMACApplyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListMACApplyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()
	direction := strings.ToLower(data.Direction.ValueString())
	aclName := data.AccessList.ValueString()

	// Build resource ID
	id := fmt.Sprintf("%s:%s", iface, direction)

	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Creating MAC access list apply")

	// Validate interface type
	if err := validateMACInterfaceType(iface); err != nil {
		resp.Diagnostics.AddError("Invalid Interface Type", err.Error())
		return
	}

	// Get filter IDs
	filterIDs := data.GetFilterIDsAsInts()

	if len(filterIDs) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"filter_ids is required: at least one filter ID must be specified",
		)
		return
	}

	// Apply filters to interface
	if err := r.client.ApplyMACFiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
		resp.Diagnostics.AddError(
			"Failed to apply MAC filters",
			fmt.Sprintf("Could not apply MAC filters to interface %s %s: %v", iface, direction, err),
		)
		return
	}

	data.ID = types.StringValue(id)

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccessListMACApplyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListMACApplyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the resource was not found (ID is now null), remove it from state
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the MAC filter apply from the router.
func (r *AccessListMACApplyResource) read(ctx context.Context, data *AccessListMACApplyModel, diagnostics *diag.Diagnostics) {
	// Parse ID
	id := data.ID.ValueString()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Invalid resource ID format: %s, expected 'interface:direction'", id),
		)
		return
	}

	iface := parts[0]
	direction := parts[1]

	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Reading MAC access list apply")

	// Get current filter IDs from router
	filterIDs, err := r.client.GetMACInterfaceFilters(ctx, iface, direction)
	if err != nil {
		// If not found, the resource has been removed
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().
				Str("resource", "rtx_access_list_mac_apply").
				Str("interface", iface).
				Str("direction", direction).
				Msg("MAC filter apply not found, removing from state")
			data.ID = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read MAC filter apply",
			fmt.Sprintf("Could not read MAC filter apply for %s %s: %v", iface, direction, err))
		return
	}

	// If no filters are applied, resource doesn't exist
	if len(filterIDs) == 0 {
		logger.Warn().
			Str("resource", "rtx_access_list_mac_apply").
			Str("interface", iface).
			Str("direction", direction).
			Msg("No MAC filters applied, removing from state")
		data.ID = types.StringNull()
		return
	}

	// Update state
	data.Interface = types.StringValue(iface)
	data.Direction = types.StringValue(direction)
	data.SetFilterIDsFromInts(filterIDs)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListMACApplyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListMACApplyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID
	id := data.ID.ValueString()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Invalid resource ID format: %s, expected 'interface:direction'", id),
		)
		return
	}

	iface := parts[0]
	direction := parts[1]
	aclName := data.AccessList.ValueString()

	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Updating MAC access list apply")

	// Get filter IDs
	filterIDs := data.GetFilterIDsAsInts()

	if len(filterIDs) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"filter_ids is required: at least one filter ID must be specified",
		)
		return
	}

	// Apply filters to interface (this will replace existing filters)
	if err := r.client.ApplyMACFiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update MAC filters",
			fmt.Sprintf("Could not update MAC filters on interface %s %s: %v", iface, direction, err),
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
func (r *AccessListMACApplyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListMACApplyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID
	id := data.ID.ValueString()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Invalid resource ID format: %s, expected 'interface:direction'", id),
		)
		return
	}

	iface := parts[0]
	direction := parts[1]

	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Deleting MAC access list apply")

	// Remove filters from interface
	if err := r.client.RemoveMACFiltersFromInterface(ctx, iface, direction); err != nil {
		// Ignore "not found" errors
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete MAC filter apply",
			fmt.Sprintf("Could not remove MAC filters from interface %s %s: %v", iface, direction, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListMACApplyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: interface:direction
	id := req.ID
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid import ID format: %s, expected 'interface:direction' (e.g., 'lan1:in')", id),
		)
		return
	}

	iface := parts[0]
	direction := strings.ToLower(parts[1])

	// Validate direction
	if direction != "in" && direction != "out" {
		resp.Diagnostics.AddError(
			"Invalid Direction",
			fmt.Sprintf("Invalid direction: %s, must be 'in' or 'out'", direction),
		)
		return
	}

	// Validate interface type
	if err := validateMACInterfaceType(iface); err != nil {
		resp.Diagnostics.AddError("Invalid Interface Type", err.Error())
		return
	}

	// Set the ID and basic attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%s:%s", iface, direction))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("interface"), iface)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("direction"), direction)...)
	// access_list will need to be set manually after import
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access_list"), "imported")...)
}
