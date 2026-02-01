package tunnel

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
	_ resource.Resource                = &TunnelResource{}
	_ resource.ResourceWithImportState = &TunnelResource{}
)

// NewTunnelResource creates a new unified tunnel resource.
func NewTunnelResource() resource.Resource {
	return &TunnelResource{}
}

// TunnelResource defines the resource implementation.
type TunnelResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *TunnelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tunnel"
}

// Schema defines the schema for the resource.
func (r *TunnelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages unified tunnel configuration on RTX routers. Supports IPsec, L2TPv3, and L2TPv2 tunnels.",
		Attributes: map[string]schema.Attribute{
			"tunnel_id": schema.Int64Attribute{
				Description: "Tunnel ID (tunnel select N, 1-6000).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 6000),
				},
			},
			"encapsulation": schema.StringAttribute{
				Description: "Tunnel encapsulation type: 'ipsec' (site-to-site VPN), 'l2tpv3' (L2VPN), or 'l2tp' (L2TPv2 remote access).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ipsec", "l2tpv3", "l2tp"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the tunnel.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				Description: "Tunnel description/name. Read-only - RTX does not support setting description within tunnel context. Use rtx_interface to set the tunnel interface description if needed.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tunnel_interface": schema.StringAttribute{
				Description: "The tunnel interface name (e.g., 'tunnel1'). Computed from tunnel_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ipsec": schema.SingleNestedBlock{
				Description: "IPsec configuration for the tunnel.",
				Attributes: map[string]schema.Attribute{
					"ipsec_tunnel_id": schema.Int64Attribute{
						Description: "IPsec tunnel ID (ipsec tunnel N). Defaults to tunnel_id if not specified.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"local_address": schema.StringAttribute{
						Description: "Local IKE endpoint address.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"remote_address": schema.StringAttribute{
						Description: "Remote IKE endpoint address or FQDN.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"pre_shared_key": schema.StringAttribute{
						Description: "IKE pre-shared key. This value is write-only and will not be stored in state.",
						Required:    true,
						Sensitive:   true,
						WriteOnly:   true,
					},
					"secure_filter_in": schema.ListAttribute{
						Description: "Inbound security filter IDs.",
						Optional:    true,
						ElementType: types.Int64Type,
					},
					"secure_filter_out": schema.ListAttribute{
						Description: "Outbound security filter IDs.",
						Optional:    true,
						ElementType: types.Int64Type,
					},
					"tcp_mss_limit": schema.StringAttribute{
						Description: "TCP MSS limit: 'auto' or numeric value.",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"ipsec_transform": schema.SingleNestedBlock{
						Description: "IPsec Phase 2 transform settings.",
						Attributes: map[string]schema.Attribute{
							"protocol": schema.StringAttribute{
								Description: "Protocol: 'esp' or 'ah'.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("esp"),
								Validators: []validator.String{
									stringvalidator.OneOf("esp", "ah"),
								},
							},
							"encryption_aes256": schema.BoolAttribute{
								Description: "Use AES-256 encryption.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"encryption_aes128": schema.BoolAttribute{
								Description: "Use AES-128 encryption.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"encryption_3des": schema.BoolAttribute{
								Description: "Use 3DES encryption.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"integrity_sha256": schema.BoolAttribute{
								Description: "Use SHA-256-HMAC integrity.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"integrity_sha1": schema.BoolAttribute{
								Description: "Use SHA-1-HMAC integrity.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"integrity_md5": schema.BoolAttribute{
								Description: "Use MD5-HMAC integrity.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
						},
					},
					"keepalive": schema.SingleNestedBlock{
						Description: "IPsec keepalive/DPD settings.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Enable keepalive.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"mode": schema.StringAttribute{
								Description: "Keepalive mode: 'dpd' or 'heartbeat'.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("dpd"),
								Validators: []validator.String{
									stringvalidator.OneOf("dpd", "heartbeat"),
								},
							},
							"interval": schema.Int64Attribute{
								Description: "Keepalive interval in seconds.",
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
							"retry": schema.Int64Attribute{
								Description: "Retry count.",
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
			"l2tp": schema.SingleNestedBlock{
				Description: "L2TP configuration for the tunnel.",
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Description: "L2TP hostname for negotiation.",
						Optional:    true,
					},
					"local_router_id": schema.StringAttribute{
						Description: "Local router ID (L2TPv3).",
						Optional:    true,
					},
					"remote_router_id": schema.StringAttribute{
						Description: "Remote router ID (L2TPv3).",
						Optional:    true,
					},
					"remote_end_id": schema.StringAttribute{
						Description: "Remote end ID (L2TPv3).",
						Optional:    true,
					},
					"always_on": schema.BoolAttribute{
						Description: "Keep connection always active.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
				Blocks: map[string]schema.Block{
					"tunnel_auth": schema.SingleNestedBlock{
						Description: "L2TP tunnel authentication.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Enable tunnel authentication.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"password": schema.StringAttribute{
								Description: "Tunnel authentication password. This value is write-only.",
								Optional:    true,
								Sensitive:   true,
								WriteOnly:   true,
							},
						},
					},
					"keepalive": schema.SingleNestedBlock{
						Description: "L2TP keepalive settings.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Enable keepalive.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"interval": schema.Int64Attribute{
								Description: "Keepalive interval in seconds.",
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
							"retry": schema.Int64Attribute{
								Description: "Retry count.",
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *TunnelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*fwhelpers.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *fwhelpers.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *TunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TunnelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource context for logging
	ctx = logging.WithResource(ctx, "rtx_tunnel", strconv.FormatInt(data.TunnelID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	tunnel := data.ToClient()
	logger.Debug().Str("resource", "rtx_tunnel").Msgf("Creating tunnel: %+v", tunnel)

	if err := r.client.CreateTunnel(ctx, tunnel); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create tunnel",
			fmt.Sprintf("Could not create tunnel: %v", err),
		)
		return
	}

	// Set the ID
	data.TunnelInterface = types.StringValue(fmt.Sprintf("tunnel%d", tunnel.ID))

	// Read back the created resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *TunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TunnelModel

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
func (r *TunnelResource) read(ctx context.Context, data *TunnelModel, diagnostics *diag.Diagnostics) {
	tunnelID := int(data.TunnelID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_tunnel", strconv.Itoa(tunnelID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_tunnel").Msgf("Reading tunnel: %d", tunnelID)

	tunnel, err := r.client.GetTunnel(ctx, tunnelID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			diagnostics.AddWarning(
				"Tunnel not found",
				fmt.Sprintf("Tunnel %d was not found on the router. It may have been deleted outside of Terraform.", tunnelID),
			)
			return
		}
		diagnostics.AddError(
			"Failed to read tunnel",
			fmt.Sprintf("Could not read tunnel %d: %v", tunnelID, err),
		)
		return
	}

	// Update the model with the read data, preserving write-only values
	data.FromClient(tunnel)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *TunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TunnelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_tunnel", strconv.FormatInt(data.TunnelID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	tunnel := data.ToClient()
	logger.Debug().Str("resource", "rtx_tunnel").Msgf("Updating tunnel: %+v", tunnel)

	if err := r.client.UpdateTunnel(ctx, tunnel); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update tunnel",
			fmt.Sprintf("Could not update tunnel: %v", err),
		)
		return
	}

	// Read back the updated resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *TunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TunnelModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnelID := int(data.TunnelID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_tunnel", strconv.Itoa(tunnelID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_tunnel").Msgf("Deleting tunnel: %d", tunnelID)

	if err := r.client.DeleteTunnel(ctx, tunnelID); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete tunnel",
			fmt.Sprintf("Could not delete tunnel %d: %v", tunnelID, err),
		)
		return
	}
}

// ImportState imports an existing resource by ID.
func (r *TunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tunnelID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected numeric tunnel_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tunnel_id"), tunnelID)...)
}
