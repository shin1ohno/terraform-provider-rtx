package ipv6_interface

import (
	"context"
	"fmt"
	"regexp"
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
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &IPv6InterfaceResource{}
	_ resource.ResourceWithImportState = &IPv6InterfaceResource{}
)

// NewIPv6InterfaceResource creates a new IPv6 interface resource.
func NewIPv6InterfaceResource() resource.Resource {
	return &IPv6InterfaceResource{}
}

// IPv6InterfaceResource defines the resource implementation.
type IPv6InterfaceResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *IPv6InterfaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_interface"
}

// Schema defines the schema for the resource.
func (r *IPv6InterfaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IPv6 interface configuration on RTX routers. This includes IPv6 addresses, Router Advertisement (RTADV), DHCPv6, MTU, and security filters.",
		Attributes: map[string]schema.Attribute{
			"interface": schema.StringAttribute{
				Description: "Interface name (e.g., 'lan1', 'lan2', 'bridge1', 'pp1', 'tunnel1').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(lan|bridge|pp|tunnel)\d+$`),
						"must be a valid interface name (e.g., 'lan1', 'lan2', 'bridge1', 'pp1', 'tunnel1')",
					),
				},
			},
			"dhcpv6_service": schema.StringAttribute{
				Description: "DHCPv6 service mode: 'server', 'client', or '' (disabled).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("", "server", "client"),
				},
			},
			"mtu": schema.Int64Attribute{
				Description: "IPv6 MTU size (minimum 1280 for IPv6). Set to 0 to use the default MTU.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"address": schema.ListNestedBlock{
				Description: "IPv6 address configuration blocks. Multiple addresses can be configured on a single interface.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "Static IPv6 address in CIDR notation (e.g., '2001:db8::1/64'). Either 'address' or 'prefix_ref' with 'interface_id' must be specified.",
							Optional:    true,
						},
						"prefix_ref": schema.StringAttribute{
							Description: "Prefix reference for dynamic address (e.g., 'ra-prefix@lan2', 'dhcp-prefix@lan2'). Must be used with 'interface_id'.",
							Optional:    true,
						},
						"interface_id": schema.StringAttribute{
							Description: "Interface identifier with prefix length (e.g., '::1/64'). Used with 'prefix_ref'.",
							Optional:    true,
						},
					},
				},
			},
			"rtadv": schema.SingleNestedBlock{
				Description: "Router Advertisement (RTADV) configuration for this interface.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable Router Advertisement on this interface.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"prefix_id": schema.Int64Attribute{
						Description: "IPv6 prefix ID to advertise. Must match an rtx_ipv6_prefix resource.",
						Required:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"o_flag": schema.BoolAttribute{
						Description: "Other Configuration Flag (O flag). When set, clients should use DHCPv6 for other configuration (e.g., DNS).",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"m_flag": schema.BoolAttribute{
						Description: "Managed Address Configuration Flag (M flag). When set, clients should use DHCPv6 for address assignment.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"lifetime": schema.Int64Attribute{
						Description: "Router lifetime in seconds. Set to 0 to use the default value.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPv6InterfaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IPv6InterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPv6InterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	interfaceName := data.Interface.ValueString()
	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", interfaceName)
	logger := logging.FromContext(ctx)

	config := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("Creating IPv6 interface configuration: %+v", config)

	if err := r.client.ConfigureIPv6Interface(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure IPv6 interface",
			fmt.Sprintf("Could not configure IPv6 interface: %v", err),
		)
		return
	}

	// Read back the created resource (interface name is the ID)
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *IPv6InterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPv6InterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// Check if resource was removed (Interface set to null)
		if data.Interface.IsNull() {
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the IPv6 interface configuration from the router.
func (r *IPv6InterfaceResource) read(ctx context.Context, data *IPv6InterfaceModel, diagnostics *diag.Diagnostics) {
	interfaceName := data.Interface.ValueString()

	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", interfaceName)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("Reading IPv6 interface configuration: %s", interfaceName)

	var config *client.IPv6InterfaceConfig
	var err error

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, cacheErr := r.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract IPv6 interfaces from parsed config
			interfaces := parsedConfig.ExtractIPv6Interfaces()
			if parsed, ok := interfaces[interfaceName]; ok {
				config = convertParsedIPv6InterfaceConfig(parsed)
				logger.Debug().Str("resource", "rtx_ipv6_interface").Msg("Found IPv6 interface in SFTP cache")
			}
		}
		if config == nil {
			// Interface not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_ipv6_interface").Msg("IPv6 interface not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or interface not found in cache
	if config == nil {
		config, err = r.client.GetIPv6InterfaceConfig(ctx, interfaceName)
		if err != nil {
			// Check if interface doesn't have any configuration
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("IPv6 interface %s configuration not found, removing from state", interfaceName)
				data.Interface = types.StringNull()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read IPv6 interface configuration", fmt.Sprintf("Could not read IPv6 interface %s: %v", interfaceName, err))
			return
		}
	}

	// Update data from the config
	data.FromClient(ctx, config, diagnostics)
}

// convertParsedIPv6InterfaceConfig converts a parser IPv6InterfaceConfig to a client IPv6InterfaceConfig.
func convertParsedIPv6InterfaceConfig(parsed *parsers.IPv6InterfaceConfig) *client.IPv6InterfaceConfig {
	config := &client.IPv6InterfaceConfig{
		Interface:     parsed.Interface,
		DHCPv6Service: parsed.DHCPv6Service,
		MTU:           parsed.MTU,
	}

	// Convert addresses
	for _, addr := range parsed.Addresses {
		config.Addresses = append(config.Addresses, client.IPv6Address{
			Address:     addr.Address,
			PrefixRef:   addr.PrefixRef,
			InterfaceID: addr.InterfaceID,
		})
	}

	// Convert RTADV config
	if parsed.RTADV != nil {
		config.RTADV = &client.RTADVConfig{
			Enabled:  parsed.RTADV.Enabled,
			PrefixID: parsed.RTADV.PrefixID,
			OFlag:    parsed.RTADV.OFlag,
			MFlag:    parsed.RTADV.MFlag,
			Lifetime: parsed.RTADV.Lifetime,
		}
	}

	return config
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPv6InterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IPv6InterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	interfaceName := data.Interface.ValueString()
	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", interfaceName)
	logger := logging.FromContext(ctx)

	// Deep copy planned RTADV block since router may not return it consistently
	var plannedRTADV *RTADVModel
	if data.RTADV != nil {
		plannedRTADV = &RTADVModel{
			Enabled:  data.RTADV.Enabled,
			PrefixID: data.RTADV.PrefixID,
			OFlag:    data.RTADV.OFlag,
			MFlag:    data.RTADV.MFlag,
			Lifetime: data.RTADV.Lifetime,
		}
	}

	config := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("Updating IPv6 interface configuration: %+v", config)

	if err := r.client.UpdateIPv6InterfaceConfig(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IPv6 interface configuration",
			fmt.Sprintf("Could not update IPv6 interface: %v", err),
		)
		return
	}

	// Read back the updated resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore planned RTADV values, but only for known (not unknown) attributes
	// Unknown attributes should get their value from the router
	if plannedRTADV != nil {
		if data.RTADV == nil {
			data.RTADV = &RTADVModel{}
		}
		if !plannedRTADV.Enabled.IsUnknown() {
			data.RTADV.Enabled = plannedRTADV.Enabled
		}
		if !plannedRTADV.PrefixID.IsUnknown() {
			data.RTADV.PrefixID = plannedRTADV.PrefixID
		}
		if !plannedRTADV.OFlag.IsUnknown() {
			data.RTADV.OFlag = plannedRTADV.OFlag
		}
		if !plannedRTADV.MFlag.IsUnknown() {
			data.RTADV.MFlag = plannedRTADV.MFlag
		}
		if !plannedRTADV.Lifetime.IsUnknown() {
			data.RTADV.Lifetime = plannedRTADV.Lifetime
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IPv6InterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPv6InterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	interfaceName := data.Interface.ValueString()

	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", interfaceName)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("Resetting IPv6 interface configuration: %s", interfaceName)

	if err := r.client.ResetIPv6Interface(ctx, interfaceName); err != nil {
		// Check if it's already reset/clean
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to reset IPv6 interface configuration",
			fmt.Sprintf("Could not reset IPv6 interface %s: %v", interfaceName, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *IPv6InterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	interfaceName := req.ID

	// Validate interface name format
	pattern := regexp.MustCompile(`^(lan|bridge|pp|tunnel)\d+$`)
	if !pattern.MatchString(interfaceName) {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected interface name (e.g., 'lan1', 'bridge1', 'pp1', 'tunnel1'), got: %s", interfaceName),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("interface"), req, resp)
}
