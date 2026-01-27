package pptp

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
	_ resource.Resource                = &PPTPResource{}
	_ resource.ResourceWithImportState = &PPTPResource{}
)

// NewPPTPResource creates a new PPTP resource.
func NewPPTPResource() resource.Resource {
	return &PPTPResource{}
}

// PPTPResource defines the resource implementation.
type PPTPResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *PPTPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pptp"
}

// Schema defines the schema for the resource.
func (r *PPTPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages PPTP VPN server configuration on RTX routers. PPTP is a singleton resource.\n\n" +
			"**Security Warning:** PPTP is considered insecure due to known vulnerabilities in its authentication and encryption protocols. " +
			"Consider using L2TP/IPsec or IKEv2 instead for better security.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'pptp' for this singleton resource).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"shutdown": schema.BoolAttribute{
				Description: "Administratively shut down PPTP service.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"listen_address": schema.StringAttribute{
				Description: "IP address to listen on.",
				Optional:    true,
				Computed:    true,
			},
			"max_connections": schema.Int64Attribute{
				Description: "Maximum concurrent connections. 0 means no limit.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"disconnect_time": schema.Int64Attribute{
				Description: "Idle disconnect time in seconds. 0 means no timeout.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"keepalive_enabled": schema.BoolAttribute{
				Description: "Enable keepalive for PPTP connections.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable PPTP service.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"authentication": schema.SingleNestedBlock{
				Description: "PPTP authentication settings.",
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Description: "Authentication method: 'pap', 'chap', 'mschap', or 'mschap-v2'. Note: mschap-v2 is required for MPPE encryption.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("pap", "chap", "mschap", "mschap-v2"),
						},
					},
					"username": schema.StringAttribute{
						Description: "Username for authentication.",
						Optional:    true,
					},
					"password": schema.StringAttribute{
						Description: "Password for authentication.",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
			"encryption": schema.SingleNestedBlock{
				Description: "MPPE encryption settings. Requires mschap or mschap-v2 authentication.",
				Attributes: map[string]schema.Attribute{
					"mppe_bits": schema.Int64Attribute{
						Description: "MPPE encryption strength: 40, 56, or 128 bits.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(40, 56, 128),
						},
					},
					"required": schema.BoolAttribute{
						Description: "Require encryption for all connections.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"ip_pool": schema.SingleNestedBlock{
				Description: "IP pool for PPTP clients.",
				Attributes: map[string]schema.Attribute{
					"start": schema.StringAttribute{
						Description: "Start IP address of the pool.",
						Required:    true,
					},
					"end": schema.StringAttribute{
						Description: "End IP address of the pool.",
						Required:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *PPTPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *PPTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PPTPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_pptp", "pptp")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_pptp").Msgf("Creating PPTP configuration: %+v", config)

	if err := r.client.CreatePPTP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create PPTP configuration",
			fmt.Sprintf("Could not create PPTP configuration: %v", err),
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
func (r *PPTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PPTPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If PPTP is not configured, remove from state
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the PPTP configuration from the router.
func (r *PPTPResource) read(ctx context.Context, data *PPTPModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_pptp", "pptp")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pptp").Msg("Reading PPTP configuration")

	var config *client.PPTPConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractPPTP()
			if parsed != nil {
				config = convertParsedPPTPConfig(parsed)
				logger.Debug().Str("resource", "rtx_pptp").Msg("Found PPTP config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_pptp").Msg("PPTP config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = r.client.GetPPTP(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_pptp").Msg("PPTP configuration not found, removing from state")
				data.ID = types.StringNull()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read PPTP configuration", fmt.Sprintf("Could not read PPTP configuration: %v", err))
			return
		}
	}

	if !config.Enabled {
		logger.Debug().Str("resource", "rtx_pptp").Msg("PPTP is disabled, removing from state")
		data.ID = types.StringNull()
		return
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PPTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PPTPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_pptp", "pptp")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_pptp").Msgf("Updating PPTP configuration: %+v", config)

	if err := r.client.UpdatePPTP(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update PPTP configuration",
			fmt.Sprintf("Could not update PPTP configuration: %v", err),
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
func (r *PPTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PPTPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_pptp", "pptp")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pptp").Msg("Disabling PPTP configuration")

	if err := r.client.DeletePPTP(ctx); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to disable PPTP",
			fmt.Sprintf("Could not disable PPTP: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *PPTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID must be "pptp" for this singleton resource
	if req.ID != "pptp" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be 'pptp' for this singleton resource.",
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// convertParsedPPTPConfig converts a parser PPTPConfig to a client PPTPConfig
func convertParsedPPTPConfig(parsed *parsers.PPTPConfig) *client.PPTPConfig {
	config := &client.PPTPConfig{
		Shutdown:         parsed.Shutdown,
		ListenAddress:    parsed.ListenAddress,
		MaxConnections:   parsed.MaxConnections,
		DisconnectTime:   parsed.DisconnectTime,
		KeepaliveEnabled: parsed.KeepaliveEnabled,
		Enabled:          parsed.Enabled,
	}

	// Convert authentication
	if parsed.Authentication != nil {
		config.Authentication = &client.PPTPAuth{
			Method:   parsed.Authentication.Method,
			Username: parsed.Authentication.Username,
			Password: parsed.Authentication.Password,
		}
	}

	// Convert encryption
	if parsed.Encryption != nil {
		config.Encryption = &client.PPTPEncryption{
			MPPEBits: parsed.Encryption.MPPEBits,
			Required: parsed.Encryption.Required,
		}
	}

	// Convert IP pool
	if parsed.IPPool != nil {
		config.IPPool = &client.PPTPIPPool{
			Start: parsed.IPPool.Start,
			End:   parsed.IPPool.End,
		}
	}

	return config
}
