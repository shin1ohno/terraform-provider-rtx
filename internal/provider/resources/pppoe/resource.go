package pppoe

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &PPPoEResource{}
	_ resource.ResourceWithImportState = &PPPoEResource{}
)

// NewPPPoEResource creates a new PPPoE resource.
func NewPPPoEResource() resource.Resource {
	return &PPPoEResource{}
}

// PPPoEResource defines the resource implementation.
type PPPoEResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *PPPoEResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pppoe"
}

// Schema defines the schema for the resource.
func (r *PPPoEResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages PPPoE connection configuration on RTX routers. This resource configures PPPoE (Point-to-Point Protocol over Ethernet) connections for WAN connectivity.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The resource identifier (PP number as string).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pp_number": schema.Int64Attribute{
				Description: "PP interface number (1-based). This identifies the PPPoE connection.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "Connection name or description.",
				Optional:    true,
			},
			"bind_interface": schema.StringAttribute{
				Description: "Physical interface to bind for PPPoE (e.g., 'lan2').",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "PPPoE authentication username.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "PPPoE authentication password.",
				Required:    true,
				Sensitive:   true,
			},
			"service_name": schema.StringAttribute{
				Description: "PPPoE service name (optional). Used to specify a particular service when multiple services are available.",
				Optional:    true,
			},
			"ac_name": schema.StringAttribute{
				Description: "PPPoE Access Concentrator name (optional).",
				Optional:    true,
			},
			"auth_method": schema.StringAttribute{
				Description: "Authentication method. Valid values: 'pap', 'chap', 'mschap', 'mschap-v2'. Defaults to 'chap'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("chap"),
				Validators: []validator.String{
					stringvalidator.OneOf("pap", "chap", "mschap", "mschap-v2"),
				},
			},
			"always_on": schema.BoolAttribute{
				Description: "Keep connection always active. Defaults to true if not specified.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"disconnect_timeout": schema.Int64Attribute{
				Description: "Idle disconnect timeout in seconds. 0 means no automatic disconnect.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"reconnect_interval": schema.Int64Attribute{
				Description: "Seconds between reconnect attempts (keepalive retry interval).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"reconnect_attempts": schema.Int64Attribute{
				Description: "Maximum reconnect attempts (0 = unlimited).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the PP interface is enabled. Defaults to true if not specified.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
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
func (r *PPPoEResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *PPPoEResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PPPoEModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum := int(data.PPNumber.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_pppoe", strconv.Itoa(ppNum))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_pppoe").Msgf("Creating PPPoE configuration for PP %d", config.Number)

	if err := r.client.CreatePPPoE(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create PPPoE configuration",
			fmt.Sprintf("Could not create PPPoE configuration: %v", err),
		)
		return
	}

	// Set the ID
	data.ID = types.StringValue(strconv.Itoa(ppNum))

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *PPPoEResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PPPoEModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was not found
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the PPPoE configuration from the router.
func (r *PPPoEResource) read(ctx context.Context, data *PPPoEModel, diagnostics *diag.Diagnostics) {
	ppNum := int(data.PPNumber.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_pppoe", strconv.Itoa(ppNum))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pppoe").Msgf("Reading PPPoE configuration for PP %d", ppNum)

	config, err := r.client.GetPPPoE(ctx, ppNum)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_pppoe").Msgf("PPPoE configuration for PP %d not found, removing from state", ppNum)
			data.ID = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read PPPoE configuration", fmt.Sprintf("Could not read PPPoE configuration for PP %d: %v", ppNum, err))
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PPPoEResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PPPoEModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum := int(data.PPNumber.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_pppoe", strconv.Itoa(ppNum))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_pppoe").Msgf("Updating PPPoE configuration for PP %d", config.Number)

	if err := r.client.UpdatePPPoE(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update PPPoE configuration",
			fmt.Sprintf("Could not update PPPoE configuration: %v", err),
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
func (r *PPPoEResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PPPoEModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppNum := int(data.PPNumber.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_pppoe", strconv.Itoa(ppNum))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pppoe").Msgf("Deleting PPPoE configuration for PP %d", ppNum)

	if err := r.client.DeletePPPoE(ctx, ppNum); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete PPPoE configuration",
			fmt.Sprintf("Could not delete PPPoE configuration for PP %d: %v", ppNum, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *PPPoEResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Parse PP number from import ID
	ppNum, err := strconv.Atoi(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid import ID format: expected PP number (e.g., '1'), got '%s'", importID),
		)
		return
	}

	if ppNum < 1 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Invalid PP number: must be >= 1",
		)
		return
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("Importing PPPoE configuration for PP %d", ppNum)

	// Set the ID and pp_number attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.Itoa(ppNum))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pp_number"), int64(ppNum))...)
}
