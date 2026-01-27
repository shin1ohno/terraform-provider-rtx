package snmp_server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SNMPServerResource{}
	_ resource.ResourceWithImportState = &SNMPServerResource{}
)

// NewSNMPServerResource creates a new SNMP server resource.
func NewSNMPServerResource() resource.Resource {
	return &SNMPServerResource{}
}

// SNMPServerResource defines the resource implementation.
type SNMPServerResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *SNMPServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snmp_server"
}

// Schema defines the schema for the resource.
func (r *SNMPServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SNMP configuration on RTX routers. This is a singleton resource - there is only one SNMP configuration per router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'snmp' for this singleton resource).",
				Computed:    true,
			},
			"location": schema.StringAttribute{
				Description: "System location (SNMP sysLocation). Describes the physical location of the device.",
				Optional:    true,
			},
			"contact": schema.StringAttribute{
				Description: "System contact (SNMP sysContact). Contact information for the device administrator.",
				Optional:    true,
			},
			"chassis_id": schema.StringAttribute{
				Description: "System name (SNMP sysName). Unique identifier for the device.",
				Optional:    true,
			},
			"enable_traps": schema.ListAttribute{
				Description: "List of trap types to enable. Valid values: all, authentication, coldstart, warmstart, linkdown, linkup, enterprise",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"community": schema.ListNestedBlock{
				Description: "SNMP community configuration",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Community string name. This is sensitive as it acts as a password for SNMP access.",
							Required:    true,
							Sensitive:   true,
						},
						"permission": schema.StringAttribute{
							Description: "Access permission: 'ro' (read-only) or 'rw' (read-write)",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("ro", "rw"),
							},
						},
						"acl": schema.StringAttribute{
							Description: "Access control list number to restrict which hosts can use this community",
							Optional:    true,
						},
					},
				},
			},
			"host": schema.ListNestedBlock{
				Description: "SNMP trap host configuration",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Description: "IP address of the SNMP trap receiver",
							Required:    true,
						},
						"community": schema.StringAttribute{
							Description: "Community string to use when sending traps to this host",
							Optional:    true,
							Sensitive:   true,
						},
						"version": schema.StringAttribute{
							Description: "SNMP version to use: '1' or '2c'",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("1", "2c"),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SNMPServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SNMPServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SNMPServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_snmp_server", "snmp")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_snmp_server").Msgf("Creating SNMP configuration: %+v", config)

	if err := r.client.CreateSNMP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SNMP configuration",
			fmt.Sprintf("Could not create SNMP configuration: %v", err),
		)
		return
	}

	// Set ID for singleton resource
	data.ID = types.StringValue("snmp")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SNMPServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SNMPServerModel

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

// read is a helper function that reads the SNMP configuration from the router.
func (r *SNMPServerResource) read(ctx context.Context, data *SNMPServerModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_snmp_server", "snmp")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_snmp_server").Msg("Reading SNMP configuration")

	var config *client.SNMPConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractSNMPServer()
			if parsed != nil {
				config = convertParsedSNMPConfig(parsed)
				logger.Debug().Str("resource", "rtx_snmp_server").Msg("Found SNMP config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_snmp_server").Msg("SNMP config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = r.client.GetSNMP(ctx)
		if err != nil {
			fwhelpers.AppendDiagError(diagnostics, "Failed to read SNMP configuration", fmt.Sprintf("Could not read SNMP configuration: %v", err))
			return
		}
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SNMPServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SNMPServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_snmp_server", "snmp")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_snmp_server").Msgf("Updating SNMP configuration: %+v", config)

	if err := r.client.UpdateSNMP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update SNMP configuration",
			fmt.Sprintf("Could not update SNMP configuration: %v", err),
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
func (r *SNMPServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SNMPServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_snmp_server", "snmp")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_snmp_server").Msg("Deleting SNMP configuration")

	if err := r.client.DeleteSNMP(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete SNMP configuration",
			fmt.Sprintf("Could not delete SNMP configuration: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *SNMPServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Only accept "snmp" as valid import ID (singleton resource)
	if req.ID != "snmp" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Invalid import ID format, expected 'snmp' for singleton resource, got: %s", req.ID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
