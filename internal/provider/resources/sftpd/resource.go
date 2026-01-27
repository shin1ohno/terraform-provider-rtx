package sftpd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &SFTPDResource{}
	_ resource.ResourceWithImportState = &SFTPDResource{}
)

// NewSFTPDResource creates a new SFTPD resource.
func NewSFTPDResource() resource.Resource {
	return &SFTPDResource{}
}

// SFTPDResource defines the resource implementation.
type SFTPDResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *SFTPDResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sftpd"
}

// Schema defines the schema for the resource.
func (r *SFTPDResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SFTP daemon (sftpd) configuration on RTX routers. " +
			"This is a singleton resource - only one instance should exist per router. " +
			"SFTPD requires SSHD to be enabled for the service to work.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'sftpd' for this singleton resource).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hosts": schema.ListAttribute{
				Description: "List of interfaces to listen on. At least one interface must be specified.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^(lan\d+|pp\d+|bridge\d+|tunnel\d+)$`),
							"must be a valid interface name (e.g., lan1, pp1, bridge1, tunnel1)",
						),
					),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SFTPDResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SFTPDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SFTPDModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_sftpd", "sftpd")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_sftpd").Msgf("Creating SFTPD configuration: hosts=%v", config.Hosts)

	if err := r.client.ConfigureSFTPD(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure SFTPD",
			fmt.Sprintf("Could not configure SFTPD: %v", err),
		)
		return
	}

	// Set the ID for singleton resource
	data.ID = types.StringValue("sftpd")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SFTPDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SFTPDModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If hosts is empty, the resource no longer exists
	if data.Hosts.IsNull() || len(data.Hosts.Elements()) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the SFTPD config from the router.
func (r *SFTPDResource) read(ctx context.Context, data *SFTPDModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_sftpd", "sftpd")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sftpd").Msg("Reading SFTPD configuration")

	config, err := r.client.GetSFTPD(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			logger.Debug().Str("resource", "rtx_sftpd").Msg("SFTPD not configured")
			data.Hosts = types.ListNull(types.StringType)
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read SFTPD configuration", fmt.Sprintf("Could not read SFTPD configuration: %v", err))
		return
	}

	if len(config.Hosts) == 0 {
		logger.Debug().Str("resource", "rtx_sftpd").Msg("SFTPD hosts not configured")
		data.Hosts = types.ListNull(types.StringType)
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SFTPDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SFTPDModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_sftpd", "sftpd")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_sftpd").Msgf("Updating SFTPD configuration: hosts=%v", config.Hosts)

	if err := r.client.UpdateSFTPD(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update SFTPD configuration",
			fmt.Sprintf("Could not update SFTPD configuration: %v", err),
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
func (r *SFTPDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SFTPDModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_sftpd", "sftpd")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sftpd").Msg("Deleting SFTPD configuration")

	if err := r.client.ResetSFTPD(ctx); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to remove SFTPD configuration",
			fmt.Sprintf("Could not remove SFTPD configuration: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *SFTPDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
