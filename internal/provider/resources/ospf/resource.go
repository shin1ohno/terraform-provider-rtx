package ospf

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/provider/validation"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &OSPFResource{}
	_ resource.ResourceWithImportState = &OSPFResource{}
)

// NewOSPFResource creates a new OSPF resource.
func NewOSPFResource() resource.Resource {
	return &OSPFResource{}
}

// OSPFResource defines the resource implementation.
type OSPFResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *OSPFResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ospf"
}

// Schema defines the schema for the resource.
func (r *OSPFResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OSPF (Open Shortest Path First) configuration on RTX routers. OSPF is a singleton resource - only one OSPF configuration can exist per router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the OSPF resource (always 'ospf').",
				Computed:    true,
			},
			"process_id": schema.Int64Attribute{
				Description: "OSPF process ID.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"router_id": schema.StringAttribute{
				Description: "OSPF router ID in IPv4 address format.",
				Required:    true,
				Validators: []validator.String{
					validation.IPv4AddressValidator(),
				},
			},
			"distance": schema.Int64Attribute{
				Description: "Administrative distance for OSPF routes.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(110),
				Validators: []validator.Int64{
					int64validator.Between(1, 255),
				},
			},
			"default_information_originate": schema.BoolAttribute{
				Description: "Originate a default route into the OSPF domain.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"redistribute_static": schema.BoolAttribute{
				Description: "Redistribute static routes into OSPF.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"redistribute_connected": schema.BoolAttribute{
				Description: "Redistribute connected routes into OSPF.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"network": schema.ListNestedBlock{
				Description: "Networks to include in OSPF.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Description: "Network IP address or interface name.",
							Required:    true,
							Validators: []validator.String{
								validation.IPv4AddressValidator(),
							},
						},
						"wildcard": schema.StringAttribute{
							Description: "Wildcard mask (inverse mask). For example, '0.0.0.255' for a /24 network.",
							Required:    true,
						},
						"area": schema.StringAttribute{
							Description: "OSPF area ID in decimal (e.g., '0') or dotted decimal (e.g., '0.0.0.0') format.",
							Required:    true,
						},
					},
				},
			},
			"area": schema.ListNestedBlock{
				Description: "OSPF area configurations.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"area_id": schema.StringAttribute{
							Description: "OSPF Area ID in decimal (e.g., '0') or dotted decimal (e.g., '0.0.0.0') format.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Area type: 'normal', 'stub', or 'nssa'.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("normal", "stub", "nssa"),
							},
						},
						"no_summary": schema.BoolAttribute{
							Description: "For stub/NSSA areas, suppress summary LSAs (totally stubby/NSSA).",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
			"neighbor": schema.ListNestedBlock{
				Description: "OSPF neighbors for NBMA networks.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Description: "Neighbor IP address.",
							Required:    true,
							Validators: []validator.String{
								validation.IPv4AddressValidator(),
							},
						},
						"priority": schema.Int64Attribute{
							Description: "Neighbor priority (0-255).",
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(1),
							Validators: []validator.Int64{
								int64validator.Between(0, 255),
							},
						},
						"cost": schema.Int64Attribute{
							Description: "Cost to reach neighbor. 0 means default cost.",
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *OSPFResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *OSPFResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OSPFModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_ospf", "ospf")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_ospf").Msgf("Creating OSPF configuration: %+v", config)

	if err := r.client.CreateOSPF(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OSPF configuration",
			fmt.Sprintf("Could not create OSPF configuration: %v", err),
		)
		return
	}

	// Set the ID
	data.ID = types.StringValue("ospf")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *OSPFResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OSPFModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was removed
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the OSPF configuration from the router.
func (r *OSPFResource) read(ctx context.Context, data *OSPFModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_ospf", "ospf")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ospf").Msg("Reading OSPF configuration")

	config, err := r.client.GetOSPF(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			logger.Debug().Str("resource", "rtx_ospf").Msg("OSPF configuration not found, removing from state")
			data.ID = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read OSPF configuration", fmt.Sprintf("Could not read OSPF configuration: %v", err))
		return
	}

	if !config.Enabled {
		logger.Debug().Str("resource", "rtx_ospf").Msg("OSPF is disabled, removing from state")
		data.ID = types.StringNull()
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *OSPFResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OSPFModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_ospf", "ospf")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_ospf").Msgf("Updating OSPF configuration: %+v", config)

	if err := r.client.UpdateOSPF(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update OSPF configuration",
			fmt.Sprintf("Could not update OSPF configuration: %v", err),
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
func (r *OSPFResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OSPFModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_ospf", "ospf")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ospf").Msg("Disabling OSPF configuration")

	if err := r.client.DeleteOSPF(ctx); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to disable OSPF",
			fmt.Sprintf("Could not disable OSPF: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *OSPFResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID must be "ospf" for this singleton resource
	if req.ID != "ospf" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be 'ospf' for this singleton resource",
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
