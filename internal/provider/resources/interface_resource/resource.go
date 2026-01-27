package interface_resource

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
	_ resource.Resource                = &InterfaceResource{}
	_ resource.ResourceWithImportState = &InterfaceResource{}
)

// NewInterfaceResource creates a new interface resource.
func NewInterfaceResource() resource.Resource {
	return &InterfaceResource{}
}

// InterfaceResource defines the resource implementation.
type InterfaceResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *InterfaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface"
}

// Schema defines the schema for the resource.
func (r *InterfaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages network interface configuration on RTX routers. This includes IP address assignment, security filters, NAT descriptors, and other interface-level settings.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
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
			"interface_name": schema.StringAttribute{
				Description: "The interface name. Same as 'name', provided for consistency with other resources.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Interface description.",
				Optional:    true,
			},
			"nat_descriptor": schema.Int64Attribute{
				Description: "NAT descriptor ID to bind to this interface. Use rtx_nat_masquerade or rtx_nat_static to define the descriptor.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"proxyarp": schema.BoolAttribute{
				Description: "Enable ProxyARP on this interface.",
				Optional:    true,
				Computed:    true,
			},
			"mtu": schema.Int64Attribute{
				Description: "Maximum Transmission Unit size. Set to 0 to use the default MTU.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ip_address": schema.SingleNestedBlock{
				Description: "IP address configuration block. Either 'address' or 'dhcp' must be set, but not both.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Description: "Static IP address in CIDR notation (e.g., '192.168.1.1/24').",
						Optional:    true,
					},
					"dhcp": schema.BoolAttribute{
						Description: "Use DHCP for IP address assignment.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *InterfaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *InterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource context for logging
	ctx = logging.WithResource(ctx, "rtx_interface", fwhelpers.GetStringValue(data.Name))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_interface").Msgf("Creating interface configuration: %+v", config)

	if err := r.client.ConfigureInterface(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create interface configuration",
			fmt.Sprintf("Could not configure interface: %v", err),
		)
		return
	}

	// Set computed attributes
	data.InterfaceName = types.StringValue(config.Name)

	// Read back the created resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *InterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the resource was deleted externally
	if data.Name.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the interface configuration from the router.
func (r *InterfaceResource) read(ctx context.Context, data *InterfaceModel, diagnostics *diag.Diagnostics) {
	interfaceName := fwhelpers.GetStringValue(data.Name)

	ctx = logging.WithResource(ctx, "rtx_interface", interfaceName)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_interface").Msgf("Reading interface configuration: %s", interfaceName)

	config, err := r.client.GetInterfaceConfig(ctx, interfaceName)
	if err != nil {
		// Check if interface doesn't have any configuration
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_interface").Msgf("Interface %s configuration not found, removing from state", interfaceName)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read interface configuration", fmt.Sprintf("Could not read interface %s: %v", interfaceName, err))
		return
	}

	// Update data from the configuration
	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *InterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InterfaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_interface", fwhelpers.GetStringValue(data.Name))
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_interface").Msgf("Updating interface configuration: %+v", config)

	if err := r.client.UpdateInterfaceConfig(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update interface configuration",
			fmt.Sprintf("Could not update interface: %v", err),
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
func (r *InterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InterfaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	interfaceName := fwhelpers.GetStringValue(data.Name)

	ctx = logging.WithResource(ctx, "rtx_interface", interfaceName)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_interface").Msgf("Resetting interface configuration: %s", interfaceName)

	if err := r.client.ResetInterface(ctx, interfaceName); err != nil {
		// Check if it's already reset/clean
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to reset interface configuration",
			fmt.Sprintf("Could not reset interface %s: %v", interfaceName, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *InterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Validate interface name format
	pattern := regexp.MustCompile(`^(lan|bridge|pp|tunnel)\d+$`)
	if !pattern.MatchString(importID) {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be a valid interface name (e.g., 'lan1', 'lan2', 'bridge1', 'pp1', 'tunnel1'), got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), importID)...)
}
