package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AdminResource{}
	_ resource.ResourceWithImportState = &AdminResource{}
)

// NewAdminResource creates a new admin resource.
func NewAdminResource() resource.Resource {
	return &AdminResource{}
}

// AdminResource defines the resource implementation.
type AdminResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AdminResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin"
}

// Schema defines the schema for the resource.
func (r *AdminResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages admin password configuration on RTX routers. This is a singleton resource - only one instance can exist per router. " +
			"Note: Changing passwords requires the provider's admin_password to be set to the current password.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'admin' for this singleton resource).",
				Computed:    true,
			},
			"login_password": schema.StringAttribute{
				Description: "Login password for the RTX router. This password is used for initial authentication when connecting to the router.",
				Optional:    true,
				Sensitive:   true,
			},
			"admin_password": schema.StringAttribute{
				Description: "Administrator password for the RTX router. This password is required for entering administrator mode to make configuration changes.",
				Optional:    true,
				Sensitive:   true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last password update performed by Terraform (RFC3339 format).",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AdminResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AdminResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AdminModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use fixed ID for singleton resource
	data.ID = types.StringValue("admin")

	ctx = logging.WithResource(ctx, "rtx_admin", "admin")
	logger := logging.FromContext(ctx)

	config := data.ToClient()

	// Only configure if passwords are provided
	if config.AdminPassword != "" || config.LoginPassword != "" {
		logger.Debug().Str("resource", "rtx_admin").Msg("Creating admin configuration")

		if err := r.client.ConfigureAdmin(ctx, config); err != nil {
			resp.Diagnostics.AddError(
				"Failed to configure admin",
				fmt.Sprintf("Could not configure admin: %v", err),
			)
			return
		}

		// Record the timestamp of successful password update
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	} else {
		logger.Debug().Str("resource", "rtx_admin").Msg("Creating admin configuration (no password changes)")
		data.LastUpdated = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AdminResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AdminModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_admin", fwhelpers.GetStringValue(data.ID))
	logger := logging.FromContext(ctx)

	// Passwords cannot be read back from the router for security reasons
	// The resource exists if it was created, and passwords remain in state
	if data.ID.IsNull() || data.ID.ValueString() == "" {
		return
	}

	logger.Debug().Str("resource", "rtx_admin").Msg("Reading admin configuration (passwords cannot be read from router)")

	// Try to use SFTP cache if enabled to verify admin config exists
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract admin config from parsed config
			parsedAdmin := parsedConfig.ExtractAdmin()
			if parsedAdmin != nil {
				// Admin config exists in cache - passwords cannot be read for security
				logger.Debug().Str("resource", "rtx_admin").Msg("Found admin config in SFTP cache (passwords not readable)")
			}
		}
	}

	// Resource still exists - keep state as is
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AdminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AdminModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_admin", fwhelpers.GetStringValue(data.ID))
	logger := logging.FromContext(ctx)

	config := data.ToClient()

	// Check if passwords are provided for update
	if config.AdminPassword != "" || config.LoginPassword != "" {
		logger.Debug().Str("resource", "rtx_admin").Msg("Updating admin configuration")

		if err := r.client.ConfigureAdmin(ctx, config); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update admin configuration",
				fmt.Sprintf("Could not update admin configuration: %v", err),
			)
			return
		}

		// Record the timestamp of successful password update
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AdminResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AdminModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: RTX password removal also requires interactive commands.
	// To remove passwords on RTX, use the console:
	//   administrator password    (then press Enter without entering a password)
	//   login password            (then press Enter without entering a password)
	//
	// This delete only removes the state record, not the actual router passwords.

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin").Msg("Deleting admin configuration (state record only - RTX password commands are interactive)")

	// State is automatically removed by Terraform
}

// ImportState imports an existing resource into Terraform.
func (r *AdminResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For import, we just set the ID to "admin"
	// Passwords must be provided in the Terraform configuration after import
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Override with fixed ID for singleton
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "admin")...)

	logging.FromContext(ctx).Info().Str("resource", "rtx_admin").Msg("Admin configuration imported. Note: Passwords must be set in configuration as they cannot be read from the router.")
}
