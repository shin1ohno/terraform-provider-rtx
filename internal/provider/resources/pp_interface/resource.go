package pp_interface

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &PPInterfaceResource{}
	_ resource.ResourceWithImportState = &PPInterfaceResource{}
)

// NewPPInterfaceResource creates a new PP interface resource.
func NewPPInterfaceResource() resource.Resource {
	return &PPInterfaceResource{}
}

// PPInterfaceResource defines the resource implementation.
type PPInterfaceResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *PPInterfaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pp_interface"
}

// Schema defines the schema for the resource.
func (r *PPInterfaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages PP (Point-to-Point) interface IP configuration on RTX routers. This resource configures IP-level settings for PP interfaces used by PPPoE connections.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The resource ID (same as pp_number as string).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pp_number": schema.Int64Attribute{
				Description: "PP interface number (1-based). This identifies the PP interface to configure.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "IP address for the PP interface. Use 'ipcp' for dynamic IP assignment from the ISP, or specify a static IP address in CIDR notation.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"mtu": schema.Int64Attribute{
				Description: "Maximum Transmission Unit size for the PP interface. Valid range: 64-1500. 0 means use default.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 1500),
				},
			},
			"tcp_mss": schema.Int64Attribute{
				Description: "TCP Maximum Segment Size limit. Valid range: 1-1460. 0 means not set.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 1460),
				},
			},
			"nat_descriptor": schema.Int64Attribute{
				Description: "NAT descriptor ID to bind to this PP interface. Use rtx_nat_masquerade or rtx_nat_static to define the descriptor.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"pp_interface": schema.StringAttribute{
				Description: "The PP interface name (e.g., 'pp1'). Computed from pp_number.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *PPInterfaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *PPInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PPInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum := int(data.PPNumber.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_pp_interface", fmt.Sprintf("%d", ppNum))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_pp_interface").Msgf("Creating PP interface IP configuration for PP %d", ppNum)

	if err := r.client.ConfigurePPInterface(ctx, ppNum, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure PP interface",
			fmt.Sprintf("Could not configure PP interface %d: %v", ppNum, err),
		)
		return
	}

	// Set ID and computed fields
	data.ID = fwhelpers.StringValueOrNull(fmt.Sprintf("%d", ppNum))
	data.PPInterface = fwhelpers.StringValueOrNull(fmt.Sprintf("pp%d", ppNum))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *PPInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PPInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid PP number",
			fmt.Sprintf("Invalid PP number in resource ID: %v", err),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_pp_interface", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pp_interface").Msgf("Reading PP interface IP configuration for PP %d", ppNum)

	config, err := r.client.GetPPInterfaceConfig(ctx, ppNum)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_pp_interface").Msgf("PP interface configuration for PP %d not found, removing from state", ppNum)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read PP interface configuration",
			fmt.Sprintf("Could not read PP interface %d: %v", ppNum, err),
		)
		return
	}

	data.FromClient(ppNum, config)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PPInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PPInterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum := int(data.PPNumber.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_pp_interface", fmt.Sprintf("%d", ppNum))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_pp_interface").Msgf("Updating PP interface IP configuration for PP %d", ppNum)

	if err := r.client.UpdatePPInterfaceConfig(ctx, ppNum, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update PP interface configuration",
			fmt.Sprintf("Could not update PP interface %d: %v", ppNum, err),
		)
		return
	}

	// Ensure ID and computed fields are set
	data.ID = fwhelpers.StringValueOrNull(fmt.Sprintf("%d", ppNum))
	data.PPInterface = fwhelpers.StringValueOrNull(fmt.Sprintf("pp%d", ppNum))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *PPInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PPInterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid PP number",
			fmt.Sprintf("Invalid PP number in resource ID: %v", err),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_pp_interface", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pp_interface").Msgf("Resetting PP interface IP configuration for PP %d", ppNum)

	if err := r.client.ResetPPInterfaceConfig(ctx, ppNum); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to reset PP interface configuration",
			fmt.Sprintf("Could not reset PP interface %d: %v", ppNum, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *PPInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Passthrough the ID to the "id" attribute - Read will populate the rest
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
