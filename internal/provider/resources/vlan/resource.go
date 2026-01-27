package vlan

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	_ resource.Resource                   = &VLANResource{}
	_ resource.ResourceWithImportState    = &VLANResource{}
	_ resource.ResourceWithValidateConfig = &VLANResource{}
)

// NewVLANResource creates a new VLAN resource.
func NewVLANResource() resource.Resource {
	return &VLANResource{}
}

// VLANResource defines the resource implementation.
type VLANResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *VLANResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

// Schema defines the schema for the resource.
func (r *VLANResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages VLAN interfaces on RTX routers. VLANs enable network segmentation using 802.1Q tagging on LAN interfaces.",
		Attributes: map[string]schema.Attribute{
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID (2-4094). VLAN 1 is reserved as the default native VLAN.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64PlanModifierRequiresReplace{},
				},
				Validators: []validator.Int64{
					int64validator.Between(2, 4094),
				},
			},
			"interface": schema.StringAttribute{
				Description: "The parent interface (e.g., 'lan1', 'lan2')",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^lan\d+$`),
						"must be a LAN interface (e.g., 'lan1', 'lan2')",
					),
				},
			},
			"name": schema.StringAttribute{
				Description: "VLAN name/description",
				Optional:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "IP address for the VLAN interface",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`),
						"must be a valid IPv4 address (e.g., '192.168.1.1')",
					),
				},
			},
			"ip_mask": schema.StringAttribute{
				Description: "Subnet mask for the VLAN interface (required if ip_address is set)",
				Optional:    true,
				Validators: []validator.String{
					subnetMaskValidator{},
				},
			},
			"shutdown": schema.BoolAttribute{
				Description: "Administrative shutdown state (true = disabled, false = enabled)",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"vlan_interface": schema.StringAttribute{
				Description: "The computed VLAN interface name (e.g., 'lan1/1')",
				Computed:    true,
			},
		},
	}
}

// ValidateConfig validates the resource configuration.
func (r *VLANResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data VLANModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ip_address is set, ip_mask must also be set
	ipAddressSet := !data.IPAddress.IsNull() && !data.IPAddress.IsUnknown() && data.IPAddress.ValueString() != ""
	ipMaskSet := !data.IPMask.IsNull() && !data.IPMask.IsUnknown() && data.IPMask.ValueString() != ""

	if ipAddressSet && !ipMaskSet {
		resp.Diagnostics.AddAttributeError(
			path.Root("ip_mask"),
			"Missing Required Attribute",
			"ip_mask is required when ip_address is specified",
		)
	}
	if ipMaskSet && !ipAddressSet {
		resp.Diagnostics.AddAttributeError(
			path.Root("ip_address"),
			"Missing Required Attribute",
			"ip_address is required when ip_mask is specified",
		)
	}
}

// Configure adds the provider configured client to the resource.
func (r *VLANResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *VLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VLANModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlan := data.ToClient()
	resourceID := fmt.Sprintf("%s/%d", vlan.Interface, vlan.VlanID)

	ctx = logging.WithResource(ctx, "rtx_vlan", resourceID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_vlan").Msgf("Creating VLAN: %+v", vlan)

	if err := r.client.CreateVLAN(ctx, vlan); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create VLAN",
			fmt.Sprintf("Could not create VLAN: %v", err),
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
func (r *VLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VLANModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the resource was not found, remove it from state
	if data.VlanID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the VLAN from the router.
func (r *VLANResource) read(ctx context.Context, data *VLANModel, diagnostics *diag.Diagnostics) {
	iface := data.Interface.ValueString()
	vlanID := int(data.VlanID.ValueInt64())
	resourceID := fmt.Sprintf("%s/%d", iface, vlanID)

	ctx = logging.WithResource(ctx, "rtx_vlan", resourceID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_vlan").Msgf("Reading VLAN: %s/%d", iface, vlanID)

	vlan, err := r.client.GetVLAN(ctx, iface, vlanID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_vlan").Msgf("VLAN %s/%d not found", iface, vlanID)
			data.VlanID = types.Int64Null()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read VLAN", fmt.Sprintf("Could not read VLAN %s/%d: %v", iface, vlanID, err))
		return
	}

	data.FromClient(vlan)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *VLANResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VLANModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlan := data.ToClient()
	resourceID := fmt.Sprintf("%s/%d", vlan.Interface, vlan.VlanID)

	ctx = logging.WithResource(ctx, "rtx_vlan", resourceID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_vlan").Msgf("Updating VLAN: %+v", vlan)

	if err := r.client.UpdateVLAN(ctx, vlan); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update VLAN",
			fmt.Sprintf("Could not update VLAN: %v", err),
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
func (r *VLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VLANModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()
	vlanID := int(data.VlanID.ValueInt64())
	resourceID := fmt.Sprintf("%s/%d", iface, vlanID)

	ctx = logging.WithResource(ctx, "rtx_vlan", resourceID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_vlan").Msgf("Deleting VLAN: %s/%d", iface, vlanID)

	if err := r.client.DeleteVLAN(ctx, iface, vlanID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete VLAN",
			fmt.Sprintf("Could not delete VLAN %s/%d: %v", iface, vlanID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *VLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Parse import ID as "interface/vlan_id" format (e.g., "lan1/10")
	iface, vlanID, err := parseVLANID(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid import ID format, expected 'interface/vlan_id' (e.g., 'lan1/10'): %v", err),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_vlan", importID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_vlan").Msgf("Importing VLAN: %s/%d", iface, vlanID)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("interface"), iface)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vlan_id"), int64(vlanID))...)
}

// parseVLANID parses the resource ID in "interface/vlan_id" format.
func parseVLANID(id string) (string, int, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("expected format 'interface/vlan_id', got %q", id)
	}

	iface := parts[0]
	vlanID, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid VLAN ID %q: %v", parts[1], err)
	}

	return iface, vlanID, nil
}

// int64PlanModifierRequiresReplace implements RequiresReplace for Int64 attributes.
type int64PlanModifierRequiresReplace struct{}

func (m int64PlanModifierRequiresReplace) Description(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (m int64PlanModifierRequiresReplace) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m int64PlanModifierRequiresReplace) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is no state value
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Do nothing if the values are the same
	if req.StateValue.Equal(req.ConfigValue) {
		return
	}

	resp.RequiresReplace = true
}

// subnetMaskValidator validates that a string is a valid subnet mask.
type subnetMaskValidator struct{}

func (v subnetMaskValidator) Description(ctx context.Context) string {
	return "value must be a valid subnet mask (e.g., '255.255.255.0')"
}

func (v subnetMaskValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v subnetMaskValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	// Validate dotted decimal format
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Subnet Mask",
			fmt.Sprintf("must be a valid subnet mask (e.g., '255.255.255.0'), got %q", value),
		)
		return
	}

	var nums []int
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Subnet Mask",
				fmt.Sprintf("must be a valid subnet mask (e.g., '255.255.255.0'), got %q", value),
			)
			return
		}
		nums = append(nums, num)
	}

	// Simple check: once we see a value less than 255, all following should be less than or equal
	foundPartial := false
	for i := 0; i < 4; i++ {
		if nums[i] < 255 {
			foundPartial = true
		} else if foundPartial {
			// Found 255 after a partial octet - invalid mask
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Subnet Mask",
				fmt.Sprintf("must be a valid subnet mask with contiguous bits, got %q", value),
			)
			return
		}
	}
}
