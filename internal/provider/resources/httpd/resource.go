package httpd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &HTTPDResource{}
	_ resource.ResourceWithImportState = &HTTPDResource{}
)

// NewHTTPDResource creates a new HTTPD resource.
func NewHTTPDResource() resource.Resource {
	return &HTTPDResource{}
}

// HTTPDResource defines the resource implementation.
type HTTPDResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *HTTPDResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_httpd"
}

// Schema defines the schema for the resource.
func (r *HTTPDResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages HTTP daemon (httpd) configuration on RTX routers. " +
			"This is a singleton resource - only one instance should exist per router. " +
			"The HTTPD service provides the web management interface for the router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier. Always 'httpd' for this singleton resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				Description: "Interface to listen on. Use 'any' for all interfaces, or specify an interface name (e.g., 'lan1', 'pp1', 'bridge1', 'tunnel1').",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(any|lan\d+|pp\d+|bridge\d+|tunnel\d+)$`),
						"must be 'any' or a valid interface name (e.g., lan1, pp1, bridge1, tunnel1)",
					),
				},
			},
			"proxy_access": schema.BoolAttribute{
				Description: "Enable L2MS proxy access for HTTP. When enabled, allows proxy access via L2MS protocol.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *HTTPDResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *HTTPDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HTTPDModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_httpd", "httpd")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_httpd").Msgf("Creating HTTPD configuration: %+v", config)

	if err := r.client.ConfigureHTTPD(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure HTTPD",
			fmt.Sprintf("Could not configure HTTPD: %v", err),
		)
		return
	}

	// Set the ID
	data.ID = fwhelpers.StringValueOrNull("httpd")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *HTTPDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HTTPDModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If resource was removed, clear state
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the HTTPD configuration from the router.
func (r *HTTPDResource) read(ctx context.Context, data *HTTPDModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_httpd", "httpd")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_httpd").Msg("Reading HTTPD configuration")

	var config *client.HTTPDConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractHTTPD()
			if parsed != nil {
				config = convertParsedHTTPDConfig(parsed)
				logger.Debug().Str("resource", "rtx_httpd").Msg("Found HTTPD config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_httpd").Msg("HTTPD config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = r.client.GetHTTPD(ctx)
		if err != nil {
			// Check if not configured
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_httpd").Msg("HTTPD not configured, removing from state")
				data.ID = fwhelpers.StringValueOrNull("")
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read HTTPD configuration", fmt.Sprintf("Could not read HTTPD configuration: %v", err))
			return
		}
	}

	// If no host is configured, the resource doesn't exist
	if config.Host == "" {
		logger.Debug().Str("resource", "rtx_httpd").Msg("HTTPD host not configured, removing from state")
		data.ID = fwhelpers.StringValueOrNull("")
		return
	}

	data.FromClient(config)
}

// convertParsedHTTPDConfig converts a parser HTTPDConfig to a client HTTPDConfig.
func convertParsedHTTPDConfig(parsed *parsers.HTTPDConfig) *client.HTTPDConfig {
	return &client.HTTPDConfig{
		Host:        parsed.Host,
		ProxyAccess: parsed.ProxyAccess,
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *HTTPDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HTTPDModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_httpd", "httpd")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_httpd").Msgf("Updating HTTPD configuration: %+v", config)

	if err := r.client.UpdateHTTPD(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update HTTPD configuration",
			fmt.Sprintf("Could not update HTTPD configuration: %v", err),
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
func (r *HTTPDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HTTPDModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_httpd", "httpd")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_httpd").Msg("Deleting HTTPD configuration")

	if err := r.client.ResetHTTPD(ctx); err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to remove HTTPD configuration",
			fmt.Sprintf("Could not remove HTTPD configuration: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *HTTPDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For singleton resources, we ignore the import ID and use "httpd"
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
