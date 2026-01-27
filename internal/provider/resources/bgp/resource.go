package bgp

import (
	"context"
	"fmt"
	"net"
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
	_ resource.Resource                = &BGPResource{}
	_ resource.ResourceWithImportState = &BGPResource{}
)

// NewBGPResource creates a new BGP resource.
func NewBGPResource() resource.Resource {
	return &BGPResource{}
}

// BGPResource defines the resource implementation.
type BGPResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *BGPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bgp"
}

// Schema defines the schema for the resource.
func (r *BGPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages BGP (Border Gateway Protocol) configuration on RTX routers. BGP is a singleton resource - only one BGP configuration can exist per router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier. Always 'bgp' for this singleton resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"asn": schema.StringAttribute{
				Description: "Autonomous System Number (1-4294967295). Supports 4-byte ASN.",
				Required:    true,
				Validators: []validator.String{
					asnValidator{},
				},
			},
			"router_id": schema.StringAttribute{
				Description: "BGP router ID in IPv4 address format. If not set, uses the highest loopback or interface IP.",
				Optional:    true,
				Validators: []validator.String{
					ipv4AddressValidator{},
				},
			},
			"default_ipv4_unicast": schema.BoolAttribute{
				Description: "Enable IPv4 unicast address family by default for new neighbors.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"log_neighbor_changes": schema.BoolAttribute{
				Description: "Log neighbor up/down changes.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"redistribute_static": schema.BoolAttribute{
				Description: "Redistribute static routes into BGP.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"redistribute_connected": schema.BoolAttribute{
				Description: "Redistribute connected routes into BGP.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"neighbor": schema.ListNestedBlock{
				Description: "BGP neighbor configurations.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Description: "Neighbor index (1-based) for RTX router configuration.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"ip": schema.StringAttribute{
							Description: "Neighbor IP address.",
							Required:    true,
							Validators: []validator.String{
								ipv4AddressValidator{},
							},
						},
						"remote_as": schema.StringAttribute{
							Description: "Remote AS number (1-4294967295).",
							Required:    true,
							Validators: []validator.String{
								asnValidator{},
							},
						},
						"hold_time": schema.Int64Attribute{
							Description: "Hold time in seconds (3-28800). Default is 90.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(3, 28800),
							},
						},
						"keepalive": schema.Int64Attribute{
							Description: "Keepalive interval in seconds (1-21845). Default is 30.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 21845),
							},
						},
						"multihop": schema.Int64Attribute{
							Description: "eBGP multihop TTL (1-255). Required for non-directly connected eBGP peers.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 255),
							},
						},
						"password": schema.StringAttribute{
							Description: "MD5 authentication password for the BGP session.",
							Optional:    true,
							Sensitive:   true,
						},
						"local_address": schema.StringAttribute{
							Description: "Local IP address for the BGP session.",
							Optional:    true,
							Validators: []validator.String{
								ipv4AddressValidator{},
							},
						},
					},
				},
			},
			"network": schema.ListNestedBlock{
				Description: "Networks to announce via BGP.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"prefix": schema.StringAttribute{
							Description: "Network prefix to announce.",
							Required:    true,
							Validators: []validator.String{
								ipv4AddressValidator{},
							},
						},
						"mask": schema.StringAttribute{
							Description: "Network mask in dotted decimal notation.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									maskRegex,
									"must be a valid IPv4 mask in dotted decimal notation",
								),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *BGPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *BGPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BGPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_bgp", "bgp")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_bgp").Msgf("Creating BGP configuration: %+v", config)

	if err := r.client.ConfigureBGP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create BGP configuration",
			fmt.Sprintf("Could not create BGP configuration: %v", err),
		)
		return
	}

	// BGP is a singleton resource
	data.ID = types.StringValue("bgp")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *BGPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BGPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If BGP was not found, remove from state
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the BGP configuration from the router.
func (r *BGPResource) read(ctx context.Context, data *BGPModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_bgp", "bgp")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_bgp").Msg("Reading BGP configuration")

	config, err := r.client.GetBGPConfig(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			logger.Debug().Str("resource", "rtx_bgp").Msg("BGP configuration not found, removing from state")
			data.ID = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read BGP configuration", fmt.Sprintf("Could not read BGP configuration: %v", err))
		return
	}

	if !config.Enabled {
		logger.Debug().Str("resource", "rtx_bgp").Msg("BGP is disabled, removing from state")
		data.ID = types.StringNull()
		return
	}

	// Preserve password values from state (they are not returned from the router)
	preservePasswords(data, config)

	data.FromClient(config)
}

// preservePasswords preserves password values from the current state in the config.
func preservePasswords(data *BGPModel, config *client.BGPConfig) {
	if data.Neighbors.IsNull() || data.Neighbors.IsUnknown() {
		return
	}

	var currentNeighbors []NeighborModel
	data.Neighbors.ElementsAs(context.TODO(), &currentNeighbors, false)

	// Build a map of passwords by index
	passwordsByIndex := make(map[int]string)
	for _, n := range currentNeighbors {
		if !n.Password.IsNull() && !n.Password.IsUnknown() {
			passwordsByIndex[fwhelpers.GetInt64Value(n.Index)] = fwhelpers.GetStringValue(n.Password)
		}
	}

	// Apply preserved passwords to config
	for i := range config.Neighbors {
		if password, ok := passwordsByIndex[config.Neighbors[i].ID]; ok {
			config.Neighbors[i].Password = password
		}
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *BGPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BGPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_bgp", "bgp")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_bgp").Msgf("Updating BGP configuration: %+v", config)

	if err := r.client.UpdateBGPConfig(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update BGP configuration",
			fmt.Sprintf("Could not update BGP configuration: %v", err),
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
func (r *BGPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BGPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_bgp", "bgp")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_bgp").Msg("Disabling BGP configuration")

	if err := r.client.ResetBGP(ctx); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to disable BGP",
			fmt.Sprintf("Could not disable BGP: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *BGPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID must be "bgp" for this singleton resource
	if req.ID != "bgp" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be 'bgp' for this singleton resource.",
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Custom validators

// asnValidator validates an Autonomous System Number.
type asnValidator struct{}

func (v asnValidator) Description(ctx context.Context) string {
	return "must be a valid AS number (1-4294967295)"
}

func (v asnValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v asnValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ASN",
			"ASN cannot be empty",
		)
		return
	}

	asn, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ASN",
			fmt.Sprintf("ASN must be a valid number (1-4294967295), got %q", value),
		)
		return
	}

	if asn == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ASN",
			fmt.Sprintf("ASN must be between 1 and 4294967295, got %q", value),
		)
	}
}

// ipv4AddressValidator validates an IPv4 address.
type ipv4AddressValidator struct{}

func (v ipv4AddressValidator) Description(ctx context.Context) string {
	return "must be a valid IPv4 address"
}

func (v ipv4AddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipv4AddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return // Optional field
	}

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Must be a valid IPv4 address, got %q", value),
		)
		return
	}

	if ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Must be an IPv4 address (not IPv6), got %q", value),
		)
	}
}

// maskRegex validates an IPv4 mask in dotted decimal notation.
var maskRegex = mustCompileRegex(`^(\d{1,3}\.){3}\d{1,3}$`)

func mustCompileRegex(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
