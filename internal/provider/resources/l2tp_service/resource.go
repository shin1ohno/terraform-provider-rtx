package l2tp_service

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &L2TPServiceResource{}
	_ resource.ResourceWithImportState = &L2TPServiceResource{}
)

// NewL2TPServiceResource creates a new L2TP service resource.
func NewL2TPServiceResource() resource.Resource {
	return &L2TPServiceResource{}
}

// L2TPServiceResource defines the resource implementation.
type L2TPServiceResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *L2TPServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_l2tp_service"
}

// Schema defines the schema for the resource.
func (r *L2TPServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages L2TP service configuration on RTX routers. " +
			"This is a singleton resource - only one instance can exist per router. " +
			"The L2TP service must be enabled for L2TP/L2TPv3 tunnels to function.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier. Always 'default' for this singleton resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable or disable the L2TP service. When disabled, all L2TP/L2TPv3 tunnels will be inactive.",
				Required:    true,
			},
			"protocols": schema.ListAttribute{
				Description: "List of L2TP protocols to enable. Valid values are 'l2tp' (L2TPv2) and 'l2tpv3'. If not specified, defaults to all protocols when enabled.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listStringValidator{allowedValues: []string{"l2tp", "l2tpv3"}},
				},
			},
		},
	}
}

// listStringValidator validates that list elements are in the allowed set.
type listStringValidator struct {
	allowedValues []string
}

func (v listStringValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("value must be one of: %v", v.allowedValues)
}

func (v listStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v listStringValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()
	for i, elem := range elements {
		if strVal, ok := elem.(types.String); ok && !strVal.IsNull() && !strVal.IsUnknown() {
			val := strVal.ValueString()
			valid := false
			for _, allowed := range v.allowedValues {
				if val == allowed {
					valid = true
					break
				}
			}
			if !valid {
				resp.Diagnostics.AddAttributeError(
					req.Path.AtListIndex(i),
					"Invalid Protocol Value",
					fmt.Sprintf("protocol value must be one of %v, got: %s", v.allowedValues, val),
				)
			}
		}
	}
}

// Ensure listStringValidator implements validator.List
var _ validator.List = listStringValidator{}

// Configure adds the provider configured client to the resource.
func (r *L2TPServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *L2TPServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data L2TPServiceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_l2tp_service", "default")
	logger := logging.FromContext(ctx)

	enabled, protocols := data.ToClient()
	logger.Debug().Str("resource", "rtx_l2tp_service").Msgf("Creating L2TP service configuration: enabled=%v, protocols=%v", enabled, protocols)

	if err := r.client.SetL2TPServiceState(ctx, enabled, protocols); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure L2TP service",
			fmt.Sprintf("Could not configure L2TP service: %v", err),
		)
		return
	}

	// Set the ID for singleton resource
	data.ID = types.StringValue("default")

	// Read back to ensure consistency
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *L2TPServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data L2TPServiceModel

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

// read is a helper function that reads the L2TP service state from the router.
func (r *L2TPServiceResource) read(ctx context.Context, data *L2TPServiceModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_l2tp_service", "default")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_l2tp_service").Msg("Reading L2TP service configuration")

	var state *client.L2TPServiceState

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract L2TP service from parsed config
			service := parsedConfig.ExtractL2TPService()
			if service != nil {
				state = &client.L2TPServiceState{
					Enabled:   service.Enabled,
					Protocols: service.Protocols,
				}
				logger.Debug().Str("resource", "rtx_l2tp_service").Msg("Found L2TP service in SFTP cache")
			}
		}
		if state == nil {
			// Service not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_l2tp_service").Msg("L2TP service not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or service not found in cache
	if state == nil {
		var err error
		state, err = r.client.GetL2TPServiceState(ctx)
		if err != nil {
			// Check if service is not configured
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_l2tp_service").Msg("L2TP service not configured, removing from state")
				data.ID = types.StringNull()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read L2TP service configuration", fmt.Sprintf("Could not read L2TP service: %v", err))
			return
		}
	}

	// Update the model from client state
	data.FromClient(state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *L2TPServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data L2TPServiceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_l2tp_service", "default")
	logger := logging.FromContext(ctx)

	enabled, protocols := data.ToClient()
	logger.Debug().Str("resource", "rtx_l2tp_service").Msgf("Updating L2TP service configuration: enabled=%v, protocols=%v", enabled, protocols)

	if err := r.client.SetL2TPServiceState(ctx, enabled, protocols); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update L2TP service configuration",
			fmt.Sprintf("Could not update L2TP service: %v", err),
		)
		return
	}

	// Read back to ensure consistency
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *L2TPServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data L2TPServiceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_l2tp_service", "default")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_l2tp_service").Msg("Deleting L2TP service configuration (disabling service)")

	// Disable L2TP service on delete
	if err := r.client.SetL2TPServiceState(ctx, false, nil); err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to disable L2TP service",
			fmt.Sprintf("Could not disable L2TP service: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *L2TPServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Accept "default" as the import ID (singleton resource)
	if req.ID != "default" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected 'default' for this singleton resource, got: %s", req.ID),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_l2tp_service", "default")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_l2tp_service").Msg("Importing L2TP service configuration")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "default")...)
}
