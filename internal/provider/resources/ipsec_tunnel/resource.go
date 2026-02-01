package ipsec_tunnel

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
	_ resource.Resource                = &IPsecTunnelResource{}
	_ resource.ResourceWithImportState = &IPsecTunnelResource{}
)

// NewIPsecTunnelResource creates a new IPsec tunnel resource.
func NewIPsecTunnelResource() resource.Resource {
	return &IPsecTunnelResource{}
}

// IPsecTunnelResource defines the resource implementation.
type IPsecTunnelResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *IPsecTunnelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_tunnel"
}

// Schema defines the schema for the resource.
func (r *IPsecTunnelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IPsec VPN tunnel configuration on RTX routers. Supports IKEv2 with pre-shared key authentication.",
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
			"ipsec_tunnel_id": schema.Int64Attribute{
				Description: "IPsec tunnel ID (ipsec tunnel N). If not specified, defaults to tunnel_id.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Tunnel description/name.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_address": schema.StringAttribute{
				Description: "Local endpoint IP address.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"remote_address": schema.StringAttribute{
				Description: "Remote endpoint IP address or hostname (for dynamic DNS).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pre_shared_key": schema.StringAttribute{
				Description: "Pre-shared key for IKE authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"local_network": schema.StringAttribute{
				Description: "Local network in CIDR notation (e.g., '192.168.1.0/24').",
				Optional:    true,
			},
			"remote_network": schema.StringAttribute{
				Description: "Remote network in CIDR notation (e.g., '10.0.0.0/24').",
				Optional:    true,
			},
			"dpd_enabled": schema.BoolAttribute{
				Description: "Enable Dead Peer Detection.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dpd_interval": schema.Int64Attribute{
				Description: "DPD interval in seconds.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dpd_retry": schema.Int64Attribute{
				Description: "DPD retry count before declaring peer dead (0 means disabled).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"keepalive_mode": schema.StringAttribute{
				Description: "Keepalive mode: 'dpd' (Dead Peer Detection) or 'heartbeat'. Defaults to 'dpd' if dpd_enabled is true.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("dpd", "heartbeat"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the IPsec tunnel.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"tunnel_interface": schema.StringAttribute{
				Description: "The tunnel interface name (e.g., 'tunnel1'). Computed from tunnel_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secure_filter_in": schema.ListAttribute{
				Description: "IP filter IDs for incoming traffic on this tunnel (ip tunnel secure filter in).",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"secure_filter_out": schema.ListAttribute{
				Description: "IP filter IDs for outgoing traffic on this tunnel (ip tunnel secure filter out).",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"tcp_mss_limit": schema.StringAttribute{
				Description: "TCP MSS limit for this tunnel: 'auto' or a numeric value (ip tunnel tcp mss limit).",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"ikev2_proposal": schema.SingleNestedBlock{
				Description: "IKE Phase 1 proposal settings.",
				Attributes: map[string]schema.Attribute{
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
						Description: "Use SHA-256 integrity.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"integrity_sha1": schema.BoolAttribute{
						Description: "Use SHA-1 integrity.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"integrity_md5": schema.BoolAttribute{
						Description: "Use MD5 integrity.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"group_fourteen": schema.BoolAttribute{
						Description: "Use DH group 14 (2048-bit).",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"group_five": schema.BoolAttribute{
						Description: "Use DH group 5 (1536-bit).",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"group_two": schema.BoolAttribute{
						Description: "Use DH group 2 (1024-bit).",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"lifetime_seconds": schema.Int64Attribute{
						Description: "IKE SA lifetime in seconds.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(60),
						},
					},
				},
			},
			"ipsec_transform": schema.SingleNestedBlock{
				Description: "IPsec Phase 2 transform settings.",
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Description: "IPsec protocol: 'esp' or 'ah'.",
						Optional:    true,
						Computed:    true,
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
					"pfs_group_fourteen": schema.BoolAttribute{
						Description: "Use PFS with DH group 14.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"pfs_group_five": schema.BoolAttribute{
						Description: "Use PFS with DH group 5.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"pfs_group_two": schema.BoolAttribute{
						Description: "Use PFS with DH group 2.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"lifetime_seconds": schema.Int64Attribute{
						Description: "IPsec SA lifetime in seconds.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(60),
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPsecTunnelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IPsecTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPsecTunnelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource context for logging
	ctx = logging.WithResource(ctx, "rtx_ipsec_tunnel", strconv.FormatInt(data.TunnelID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	tunnel := data.ToClient()
	logger.Debug().Str("resource", "rtx_ipsec_tunnel").Msgf("Creating IPsec tunnel: %+v", tunnel)

	if err := r.client.CreateIPsecTunnel(ctx, tunnel); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create IPsec tunnel",
			fmt.Sprintf("Could not create IPsec tunnel: %v", err),
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
func (r *IPsecTunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPsecTunnelModel

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
func (r *IPsecTunnelResource) read(ctx context.Context, data *IPsecTunnelModel, diagnostics *diag.Diagnostics) {
	tunnelID := int(data.TunnelID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_ipsec_tunnel", strconv.Itoa(tunnelID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipsec_tunnel").Msgf("Reading IPsec tunnel: %d", tunnelID)

	tunnel, err := r.client.GetIPsecTunnel(ctx, tunnelID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_ipsec_tunnel").Msgf("IPsec tunnel %d not found", tunnelID)
			// Resource has been deleted outside of Terraform
			data.TunnelID = types.Int64Null()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read IPsec tunnel", fmt.Sprintf("Could not read IPsec tunnel %d: %v", tunnelID, err))
		return
	}

	// Update data from the tunnel
	data.FromClient(tunnel)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPsecTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IPsecTunnelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_ipsec_tunnel", strconv.FormatInt(data.TunnelID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	// Preserve planned values that router may not return consistently
	plannedDPDEnabled := data.DPDEnabled
	plannedDPDInterval := data.DPDInterval
	plannedDPDRetry := data.DPDRetry
	plannedKeepaliveMode := data.KeepaliveMode
	plannedIKEv2Proposal := data.IKEv2Proposal

	tunnel := data.ToClient()
	logger.Debug().Str("resource", "rtx_ipsec_tunnel").Msgf("Updating IPsec tunnel: %+v", tunnel)

	if err := r.client.UpdateIPsecTunnel(ctx, tunnel); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IPsec tunnel",
			fmt.Sprintf("Could not update IPsec tunnel: %v", err),
		)
		return
	}

	// Read back the updated resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore planned DPD values - router may return different defaults
	// Only restore if planned values are known (not unknown)
	if !plannedDPDEnabled.IsUnknown() {
		data.DPDEnabled = plannedDPDEnabled
	}
	if !plannedDPDInterval.IsUnknown() {
		data.DPDInterval = plannedDPDInterval
	}
	if !plannedDPDRetry.IsUnknown() {
		data.DPDRetry = plannedDPDRetry
	}
	if !plannedKeepaliveMode.IsUnknown() {
		data.KeepaliveMode = plannedKeepaliveMode
	}

	// Preserve planned IKEv2Proposal - if plan doesn't have it, don't create it
	// This handles the case where user removes the ikev2_proposal block
	data.IKEv2Proposal = plannedIKEv2Proposal

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IPsecTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPsecTunnelModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnelID := int(data.TunnelID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_ipsec_tunnel", strconv.Itoa(tunnelID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipsec_tunnel").Msgf("Deleting IPsec tunnel: %d", tunnelID)

	if err := r.client.DeleteIPsecTunnel(ctx, tunnelID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete IPsec tunnel",
			fmt.Sprintf("Could not delete IPsec tunnel %d: %v", tunnelID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *IPsecTunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
