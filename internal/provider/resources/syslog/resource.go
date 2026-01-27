package syslog

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
	_ resource.Resource                = &SyslogResource{}
	_ resource.ResourceWithImportState = &SyslogResource{}
)

// NewSyslogResource creates a new syslog resource.
func NewSyslogResource() resource.Resource {
	return &SyslogResource{}
}

// SyslogResource defines the resource implementation.
type SyslogResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *SyslogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_syslog"
}

// Schema defines the schema for the resource.
func (r *SyslogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages syslog configuration on RTX routers. This is a singleton resource - only one instance can exist per router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'syslog' for this singleton resource).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_address": schema.StringAttribute{
				Description: "Source IP address for syslog messages.",
				Optional:    true,
				Validators: []validator.String{
					ipAddressValidator{},
				},
			},
			"facility": schema.StringAttribute{
				Description: "Syslog facility (user, local0-local7).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(
						"user", "local0", "local1", "local2", "local3",
						"local4", "local5", "local6", "local7",
					),
				},
			},
			"notice": schema.BoolAttribute{
				Description: "Enable notice level logging.",
				Optional:    true,
				Computed:    true,
			},
			"info": schema.BoolAttribute{
				Description: "Enable info level logging.",
				Optional:    true,
				Computed:    true,
			},
			"debug": schema.BoolAttribute{
				Description: "Enable debug level logging.",
				Optional:    true,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"host": schema.SetNestedBlock{
				Description: "Syslog destination hosts (one or more).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "IP address or hostname of the syslog server.",
							Required:    true,
							Validators: []validator.String{
								syslogHostAddressValidator{},
							},
						},
						"port": schema.Int64Attribute{
							Description: "UDP port (default 514, use 0 to use default).",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(0, 65535),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SyslogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SyslogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SyslogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_syslog", "syslog")
	logger := logging.FromContext(ctx)

	config, diags := data.ToClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_syslog").Msgf("Creating syslog configuration: %+v", config)

	if err := r.client.ConfigureSyslog(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create syslog configuration",
			fmt.Sprintf("Could not create syslog configuration: %v", err),
		)
		return
	}

	// Set ID for singleton resource
	data.ID = types.StringValue("syslog")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SyslogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SyslogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the resource was not found, remove from state
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the syslog config from the router.
func (r *SyslogResource) read(ctx context.Context, data *SyslogModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_syslog", "syslog")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_syslog").Msg("Reading syslog configuration")

	var config *client.SyslogConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractSyslog()
			if parsed != nil {
				config = convertParsedSyslogConfig(parsed)
				logger.Debug().Str("resource", "rtx_syslog").Msg("Found syslog config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_syslog").Msg("Syslog config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = r.client.GetSyslogConfig(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_syslog").Msg("Syslog configuration not found, removing from state")
				data.ID = types.StringNull()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read syslog configuration", fmt.Sprintf("Could not read syslog configuration: %v", err))
			return
		}
	}

	diagnostics.Append(data.FromClient(ctx, config)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SyslogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SyslogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_syslog", "syslog")
	logger := logging.FromContext(ctx)

	config, diags := data.ToClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_syslog").Msgf("Updating syslog configuration: %+v", config)

	if err := r.client.UpdateSyslogConfig(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update syslog configuration",
			fmt.Sprintf("Could not update syslog configuration: %v", err),
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
func (r *SyslogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SyslogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_syslog", "syslog")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_syslog").Msg("Deleting syslog configuration")

	if err := r.client.ResetSyslog(ctx); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete syslog configuration",
			fmt.Sprintf("Could not delete syslog configuration: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *SyslogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Accept "syslog" as the import ID (singleton resource)
	if importID != "syslog" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID 'syslog', got %q", importID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// convertParsedSyslogConfig converts a parser SyslogConfig to a client SyslogConfig.
func convertParsedSyslogConfig(parsed *parsers.SyslogConfig) *client.SyslogConfig {
	config := &client.SyslogConfig{
		LocalAddress: parsed.LocalAddress,
		Facility:     parsed.Facility,
		Notice:       parsed.Notice,
		Info:         parsed.Info,
		Debug:        parsed.Debug,
		Hosts:        make([]client.SyslogHost, len(parsed.Hosts)),
	}
	for i, host := range parsed.Hosts {
		config.Hosts[i] = client.SyslogHost{
			Address: host.Address,
			Port:    host.Port,
		}
	}
	return config
}

// Custom validators

// ipAddressValidator validates IP addresses.
type ipAddressValidator struct{}

func (v ipAddressValidator) Description(ctx context.Context) string {
	return "must be a valid IPv4 or IPv6 address"
}

func (v ipAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	if !isValidIPv4(value) && !isValidIPv6(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IP Address",
			fmt.Sprintf("Value %q is not a valid IPv4 or IPv6 address.", value),
		)
	}
}

// syslogHostAddressValidator validates syslog host address (IP or hostname).
type syslogHostAddressValidator struct{}

func (v syslogHostAddressValidator) Description(ctx context.Context) string {
	return "must be a valid IP address or hostname"
}

func (v syslogHostAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v syslogHostAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Host Address",
			"Host address cannot be empty.",
		)
		return
	}

	if !isValidIPv4(value) && !isValidIPv6(value) && !isValidHostname(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Host Address",
			fmt.Sprintf("Value %q is not a valid IP address or hostname.", value),
		)
	}
}

// isValidIPv4 checks if a string is a valid IPv4 address.
func isValidIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		num := 0
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
			num = num*10 + int(c-'0')
		}
		if num > 255 {
			return false
		}
		// No leading zeros (except "0" itself)
		if len(part) > 1 && part[0] == '0' {
			return false
		}
	}
	return true
}

// isValidIPv6 checks if a string is a valid IPv6 address (basic check).
func isValidIPv6(ip string) bool {
	if !strings.Contains(ip, ":") {
		return false
	}
	parts := strings.Split(ip, ":")
	return len(parts) <= 8
}

// isValidHostname checks if a string is a valid hostname.
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		// Must start and end with alphanumeric
		if !isAlphanumeric(rune(label[0])) || !isAlphanumeric(rune(label[len(label)-1])) {
			return false
		}
		// Can only contain alphanumeric and hyphens
		for _, c := range label {
			if !isAlphanumeric(c) && c != '-' {
				return false
			}
		}
	}
	return true
}

// isAlphanumeric checks if a rune is alphanumeric.
func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
