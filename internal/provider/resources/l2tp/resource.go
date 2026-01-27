package l2tp

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &L2TPResource{}
	_ resource.ResourceWithImportState = &L2TPResource{}
)

// NewL2TPResource creates a new L2TP resource.
func NewL2TPResource() resource.Resource {
	return &L2TPResource{}
}

// L2TPResource defines the resource implementation.
type L2TPResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *L2TPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_l2tp"
}

// Schema defines the schema for the resource.
func (r *L2TPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages L2TP/L2TPv3 tunnel configuration on RTX routers. Supports both L2TPv2 (LNS for remote access VPN) and L2TPv3 (L2VPN for site-to-site).",
		Attributes: map[string]schema.Attribute{
			"tunnel_id": schema.Int64Attribute{
				Description: "Tunnel ID (1-6000).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 6000),
				},
			},
			"tunnel_interface": schema.StringAttribute{
				Description: "The tunnel interface name (e.g., 'tunnel1'). Computed from tunnel_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Tunnel description.",
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: "L2TP version: 'l2tp' (v2) or 'l2tpv3' (v3).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("l2tp", "l2tpv3"),
				},
			},
			"mode": schema.StringAttribute{
				Description: "Operating mode: 'lns' (L2TPv2 server) or 'l2vpn' (L2TPv3 site-to-site).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("lns", "l2vpn"),
				},
			},
			"shutdown": schema.BoolAttribute{
				Description: "Administratively shut down the tunnel.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"tunnel_source": schema.StringAttribute{
				Description: "Source IP address or interface.",
				Optional:    true,
			},
			"tunnel_destination": schema.StringAttribute{
				Description: "Destination IP address or FQDN.",
				Optional:    true,
			},
			"tunnel_dest_type": schema.StringAttribute{
				Description: "Destination type: 'ip' or 'fqdn'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ip", "fqdn"),
				},
			},
			"keepalive_enabled": schema.BoolAttribute{
				Description: "Enable keepalive.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"keepalive_interval": schema.Int64Attribute{
				Description: "Keepalive interval in seconds.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"keepalive_retry": schema.Int64Attribute{
				Description: "Keepalive retry count.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"disconnect_time": schema.Int64Attribute{
				Description: "Idle disconnect time in seconds. 0 means no timeout.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"always_on": schema.BoolAttribute{
				Description: "Enable always-on mode.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the L2TP tunnel.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"authentication": schema.SingleNestedBlock{
				Description: "L2TP authentication settings. Required for L2TP LNS mode, not needed for L2TPv3.",
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Description: "Authentication method: 'pap', 'chap', 'mschap', or 'mschap-v2'.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("pap", "chap", "mschap", "mschap-v2"),
						},
					},
					"username": schema.StringAttribute{
						Description: "Username for authentication.",
						Optional:    true,
					},
					"password": schema.StringAttribute{
						Description: "Password for authentication.",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
			"ip_pool": schema.SingleNestedBlock{
				Description: "IP pool for L2TP LNS clients. Required for LNS mode, not needed for L2TPv3.",
				Attributes: map[string]schema.Attribute{
					"start": schema.StringAttribute{
						Description: "Start IP address of the pool.",
						Optional:    true,
					},
					"end": schema.StringAttribute{
						Description: "End IP address of the pool.",
						Optional:    true,
					},
				},
			},
			"ipsec_profile": schema.SingleNestedBlock{
				Description: "IPsec encryption settings for L2TP.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable IPsec encryption.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"pre_shared_key": schema.StringAttribute{
						Description: "IPsec pre-shared key.",
						Optional:    true,
						Sensitive:   true,
					},
					"tunnel_id": schema.Int64Attribute{
						Description: "Associated IPsec tunnel ID.",
						Optional:    true,
					},
				},
			},
			"l2tpv3_config": schema.SingleNestedBlock{
				Description: "L2TPv3-specific configuration.",
				Attributes: map[string]schema.Attribute{
					"local_router_id": schema.StringAttribute{
						Description: "Local router ID.",
						Optional:    true,
						Computed:    true,
					},
					"remote_router_id": schema.StringAttribute{
						Description: "Remote router ID.",
						Optional:    true,
						Computed:    true,
					},
					"remote_end_id": schema.StringAttribute{
						Description: "Remote end ID (hostname).",
						Optional:    true,
					},
					"session_id": schema.Int64Attribute{
						Description: "Session ID.",
						Optional:    true,
					},
					"cookie_size": schema.Int64Attribute{
						Description: "Cookie size: 0, 4, or 8 bytes.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(0, 4, 8),
						},
					},
					"bridge_interface": schema.StringAttribute{
						Description: "Bridge interface for L2VPN.",
						Optional:    true,
					},
					"tunnel_auth_enabled": schema.BoolAttribute{
						Description: "Enable tunnel authentication.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"tunnel_auth_password": schema.StringAttribute{
						Description: "Tunnel authentication password. This value is write-only and will not be stored in state.",
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *L2TPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *L2TPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data L2TPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_l2tp", strconv.FormatInt(data.TunnelID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_l2tp").Msgf("Creating L2TP tunnel: %+v", config)

	if err := r.client.CreateL2TP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create L2TP tunnel",
			fmt.Sprintf("Could not create L2TP tunnel: %v", err),
		)
		return
	}

	data.TunnelInterface = types.StringValue(fmt.Sprintf("tunnel%d", config.ID))

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *L2TPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data L2TPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the tunnel from the router.
func (r *L2TPResource) read(ctx context.Context, data *L2TPModel, diagnostics *diag.Diagnostics) {
	tunnelID := int(data.TunnelID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_l2tp", strconv.Itoa(tunnelID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_l2tp").Msgf("Reading L2TP tunnel: %d", tunnelID)

	config, err := r.client.GetL2TP(ctx, tunnelID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_l2tp").Msgf("L2TP tunnel %d not found", tunnelID)
			data.TunnelID = types.Int64Null()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read L2TP tunnel", fmt.Sprintf("Could not read L2TP tunnel %d: %v", tunnelID, err))
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *L2TPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data L2TPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_l2tp", strconv.FormatInt(data.TunnelID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_l2tp").Msgf("Updating L2TP tunnel: %+v", config)

	if err := r.client.UpdateL2TP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update L2TP tunnel",
			fmt.Sprintf("Could not update L2TP tunnel: %v", err),
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
func (r *L2TPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data L2TPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnelID := int(data.TunnelID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_l2tp", strconv.Itoa(tunnelID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_l2tp").Msgf("Deleting L2TP tunnel: %d", tunnelID)

	if err := r.client.DeleteL2TP(ctx, tunnelID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete L2TP tunnel",
			fmt.Sprintf("Could not delete L2TP tunnel %d: %v", tunnelID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *L2TPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tunnelID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected tunnel ID as integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tunnel_id"), int64(tunnelID))...)
}
