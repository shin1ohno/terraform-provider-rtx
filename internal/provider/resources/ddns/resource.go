package ddns

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// regexpHTTPURL validates that a URL starts with http:// or https://
var regexpHTTPURL = regexp.MustCompile(`^https?://`)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &DDNSResource{}
	_ resource.ResourceWithImportState = &DDNSResource{}
)

// NewDDNSResource creates a new DDNS resource.
func NewDDNSResource() resource.Resource {
	return &DDNSResource{}
}

// DDNSResource defines the resource implementation.
type DDNSResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *DDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddns"
}

// Schema defines the schema for the resource.
func (r *DDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages custom DDNS provider configuration on RTX routers. Use this resource to configure third-party DDNS services like No-IP, DynDNS, or other compatible providers.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.Int64Attribute{
				Description: "DDNS server ID (1-4). Each ID can be configured with a different DDNS provider.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4),
				},
			},
			"url": schema.StringAttribute{
				Description: "DDNS update URL (must start with http:// or https://). This is the provider's update endpoint.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexpHTTPURL,
						"must start with http:// or https://",
					),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "DDNS hostname to update (e.g., 'example.no-ip.org').",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "DDNS account username for authentication.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "DDNS account password for authentication.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *DDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DDNSModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := strconv.FormatInt(data.ServerID.ValueInt64(), 10)
	ctx = logging.WithResource(ctx, "rtx_ddns", serverID)
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_ddns").Msgf("Creating DDNS configuration: server_id=%d, hostname=%s", config.ID, config.Hostname)

	if err := r.client.ConfigureDDNS(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure DDNS",
			fmt.Sprintf("Could not configure DDNS: %v", err),
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
func (r *DDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DDNSModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If resource was not found, remove from state
	if data.ServerID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the DDNS configuration from the router.
func (r *DDNSResource) read(ctx context.Context, data *DDNSModel, diagnostics *diag.Diagnostics) {
	serverID := int(data.ServerID.ValueInt64())
	serverIDStr := strconv.Itoa(serverID)

	ctx = logging.WithResource(ctx, "rtx_ddns", serverIDStr)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ddns").Msgf("Reading DDNS configuration for server_id: %d", serverID)

	config, err := r.client.GetDDNSByID(ctx, serverID)
	if err != nil {
		fwhelpers.AppendDiagError(diagnostics, "Failed to read DDNS configuration", fmt.Sprintf("Could not read DDNS configuration: %v", err))
		data.ServerID = types.Int64Null()
		return
	}

	if config == nil {
		logger.Debug().Str("resource", "rtx_ddns").Msgf("DDNS configuration not found for server_id: %d", serverID)
		data.ServerID = types.Int64Null()
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DDNSModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := strconv.FormatInt(data.ServerID.ValueInt64(), 10)
	ctx = logging.WithResource(ctx, "rtx_ddns", serverID)
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_ddns").Msgf("Updating DDNS configuration: server_id=%d, hostname=%s", config.ID, config.Hostname)

	if err := r.client.UpdateDDNS(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update DDNS configuration",
			fmt.Sprintf("Could not update DDNS configuration: %v", err),
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
func (r *DDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DDNSModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := int(data.ServerID.ValueInt64())
	serverIDStr := strconv.Itoa(serverID)

	ctx = logging.WithResource(ctx, "rtx_ddns", serverIDStr)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ddns").Msgf("Deleting DDNS configuration for server_id: %d", serverID)

	if err := r.client.DeleteDDNS(ctx, serverID); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete DDNS configuration",
			fmt.Sprintf("Could not delete DDNS configuration for server_id %d: %v", serverID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *DDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serverID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid DDNS server ID",
			fmt.Sprintf("Invalid DDNS server ID: %s (must be integer 1-4)", req.ID),
		)
		return
	}

	if serverID < 1 || serverID > 4 {
		resp.Diagnostics.AddError(
			"Invalid DDNS server ID",
			fmt.Sprintf("Invalid DDNS server ID: %d (must be 1-4)", serverID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_id"), serverID)...)
}
