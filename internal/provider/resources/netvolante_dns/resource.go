package netvolante_dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
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
	_ resource.Resource                = &NetVolanteDNSResource{}
	_ resource.ResourceWithImportState = &NetVolanteDNSResource{}
)

// NewNetVolanteDNSResource creates a new NetVolante DNS resource.
func NewNetVolanteDNSResource() resource.Resource {
	return &NetVolanteDNSResource{}
}

// NetVolanteDNSResource defines the resource implementation.
type NetVolanteDNSResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *NetVolanteDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_netvolante_dns"
}

// Schema defines the schema for the resource.
func (r *NetVolanteDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages NetVolante DNS (Yamaha's free DDNS service) configuration on RTX routers. Use this resource to register your router's IP address with a *.netvolante.jp hostname.",
		Attributes: map[string]schema.Attribute{
			"interface": schema.StringAttribute{
				Description: "Interface to use for DDNS updates (e.g., 'pp 1', 'lan1'). This determines which IP address is registered.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "NetVolante DNS hostname (e.g., 'example.aa0.netvolante.jp'). Must end with .netvolante.jp.",
				Required:    true,
			},
			"server": schema.Int64Attribute{
				Description: "NetVolante DNS server number (1 or 2). Default is 1.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 2),
				},
			},
			"timeout": schema.Int64Attribute{
				Description: "Update timeout in seconds (1-3600). Default is 60.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(1, 3600),
				},
			},
			"ipv6_enabled": schema.BoolAttribute{
				Description: "Enable IPv6 address registration with NetVolante DNS. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"auto_hostname": schema.BoolAttribute{
				Description: "Enable automatic hostname generation. When enabled, the router generates a unique hostname.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NetVolanteDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NetVolanteDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetVolanteDNSModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_netvolante_dns", data.Interface.ValueString())
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_netvolante_dns").Msgf("Creating NetVolante DNS configuration: %+v", config)

	if err := r.client.ConfigureNetVolanteDNS(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create NetVolante DNS configuration",
			fmt.Sprintf("Could not create NetVolante DNS configuration: %v", err),
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
func (r *NetVolanteDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetVolanteDNSModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the resource was not found, remove from state
	if data.Interface.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the configuration from the router.
func (r *NetVolanteDNSResource) read(ctx context.Context, data *NetVolanteDNSModel, diagnostics *diag.Diagnostics) {
	iface := data.Interface.ValueString()

	ctx = logging.WithResource(ctx, "rtx_netvolante_dns", iface)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_netvolante_dns").Msgf("Reading NetVolante DNS configuration for interface: %s", iface)

	config, err := r.client.GetNetVolanteDNSByInterface(ctx, iface)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_netvolante_dns").Msgf("NetVolante DNS configuration for interface %s not found", iface)
			data.Interface = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read NetVolante DNS configuration", fmt.Sprintf("Could not read NetVolante DNS configuration for interface %s: %v", iface, err))
		return
	}

	if config == nil {
		logger.Debug().Str("resource", "rtx_netvolante_dns").Msgf("NetVolante DNS configuration for interface %s not found", iface)
		data.Interface = types.StringNull()
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NetVolanteDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetVolanteDNSModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_netvolante_dns", data.Interface.ValueString())
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_netvolante_dns").Msgf("Updating NetVolante DNS configuration: %+v", config)

	if err := r.client.UpdateNetVolanteDNS(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update NetVolante DNS configuration",
			fmt.Sprintf("Could not update NetVolante DNS configuration: %v", err),
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
func (r *NetVolanteDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetVolanteDNSModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface := data.Interface.ValueString()

	ctx = logging.WithResource(ctx, "rtx_netvolante_dns", iface)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_netvolante_dns").Msgf("Deleting NetVolante DNS configuration for interface: %s", iface)

	if err := r.client.DeleteNetVolanteDNS(ctx, iface); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete NetVolante DNS configuration",
			fmt.Sprintf("Could not delete NetVolante DNS configuration for interface %s: %v", iface, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *NetVolanteDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("interface"), req, resp)
}
