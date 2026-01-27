package system

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SystemResource{}
	_ resource.ResourceWithImportState = &SystemResource{}
)

// NewSystemResource creates a new system resource.
func NewSystemResource() resource.Resource {
	return &SystemResource{}
}

// SystemResource defines the resource implementation.
type SystemResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *SystemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

// Schema defines the schema for the resource.
func (r *SystemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages system-level settings on RTX routers. This is a singleton resource - there is only one system configuration per router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'system' for this singleton resource).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone as UTC offset (e.g., '+09:00' for JST, '-05:00' for EST).",
				Optional:    true,
				Validators: []validator.String{
					timezoneValidator{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"console": schema.ListNestedBlock{
				Description: "Console settings.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"character": schema.StringAttribute{
							Description: "Character encoding (ja.utf8, ja.sjis, ascii, euc-jp).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("ja.utf8", "ja.sjis", "ascii", "euc-jp"),
							},
						},
						"lines": schema.StringAttribute{
							Description: "Lines per page (positive integer or 'infinity').",
							Optional:    true,
							Validators: []validator.String{
								consoleLinesValidator{},
							},
						},
						"prompt": schema.StringAttribute{
							Description: "Custom prompt string.",
							Optional:    true,
						},
					},
				},
			},
			"packet_buffer": schema.ListNestedBlock{
				Description: "Packet buffer tuning settings (small, middle, large).",
				Validators: []validator.List{
					listvalidator.SizeAtMost(3),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"size": schema.StringAttribute{
							Description: "Buffer size category (small, middle, large).",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("small", "middle", "large"),
							},
						},
						"max_buffer": schema.Int64Attribute{
							Description: "Maximum buffer count.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"max_free": schema.Int64Attribute{
							Description: "Maximum free buffer count.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
				},
			},
			"statistics": schema.ListNestedBlock{
				Description: "Statistics collection settings.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"traffic": schema.BoolAttribute{
							Description: "Enable traffic statistics collection.",
							Optional:    true,
							Computed:    true,
						},
						"nat": schema.BoolAttribute{
							Description: "Enable NAT statistics collection.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SystemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SystemModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_system", "system")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_system").Msgf("Creating system configuration: %+v", config)

	if err := r.client.ConfigureSystem(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create system configuration",
			fmt.Sprintf("Could not create system configuration: %v", err),
		)
		return
	}

	// Set ID for singleton resource
	data.ID = types.StringValue("system")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SystemModel

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

// read is a helper function that reads the system config from the router.
func (r *SystemResource) read(ctx context.Context, data *SystemModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_system", "system")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_system").Msg("Reading system configuration")

	var config *client.SystemConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract system config from parsed config
			parsedSystem := parsedConfig.ExtractSystem()
			if parsedSystem != nil {
				config = convertParsedSystemConfig(parsedSystem)
				logger.Debug().Str("resource", "rtx_system").Msg("Found system config in SFTP cache")
			}
		}
		if config == nil {
			// Config not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_system").Msg("System config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = r.client.GetSystemConfig(ctx)
		if err != nil {
			fwhelpers.AppendDiagError(diagnostics, "Failed to read system configuration", fmt.Sprintf("Could not read system configuration: %v", err))
			return
		}
	}

	data.FromClient(config)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SystemModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_system", "system")
	logger := logging.FromContext(ctx)

	config := data.ToClient()
	logger.Debug().Str("resource", "rtx_system").Msgf("Updating system configuration: %+v", config)

	if err := r.client.UpdateSystemConfig(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update system configuration",
			fmt.Sprintf("Could not update system configuration: %v", err),
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
func (r *SystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SystemModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_system", "system")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_system").Msg("Deleting (resetting) system configuration")

	if err := r.client.ResetSystem(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to reset system configuration",
			fmt.Sprintf("Could not reset system configuration: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *SystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Only accept "system" as valid import ID (singleton resource)
	if importID != "system" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Invalid import ID format, expected 'system' for singleton resource, got: %s", importID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// convertParsedSystemConfig converts a parser SystemConfig to a client SystemConfig
func convertParsedSystemConfig(parsed *parsers.SystemConfig) *client.SystemConfig {
	config := &client.SystemConfig{
		Timezone:      parsed.Timezone,
		PacketBuffers: make([]client.PacketBufferConfig, len(parsed.PacketBuffers)),
	}

	// Convert console config
	if parsed.Console != nil {
		config.Console = &client.ConsoleConfig{
			Character: parsed.Console.Character,
			Lines:     parsed.Console.Lines,
			Prompt:    parsed.Console.Prompt,
		}
	}

	// Convert packet buffers
	for i, pb := range parsed.PacketBuffers {
		config.PacketBuffers[i] = client.PacketBufferConfig{
			Size:      pb.Size,
			MaxBuffer: pb.MaxBuffer,
			MaxFree:   pb.MaxFree,
		}
	}

	// Convert statistics config
	if parsed.Statistics != nil {
		config.Statistics = &client.StatisticsConfig{
			Traffic: parsed.Statistics.Traffic,
			NAT:     parsed.Statistics.NAT,
		}
	}

	return config
}

// Custom validators

// timezoneValidator validates timezone format (+-HH:MM).
type timezoneValidator struct{}

func (v timezoneValidator) Description(ctx context.Context) string {
	return "value must be a valid UTC offset (e.g., '+09:00', '-05:00')"
}

func (v timezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timezoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	pattern := regexp.MustCompile(`^[\+\-]\d{2}:\d{2}$`)
	if !pattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Timezone Format",
			fmt.Sprintf("Value must be a valid UTC offset (e.g., '+09:00', '-05:00'), got: %s", value),
		)
	}
}

// consoleLinesValidator validates console lines setting.
type consoleLinesValidator struct{}

func (v consoleLinesValidator) Description(ctx context.Context) string {
	return "value must be a positive integer or 'infinity'"
}

func (v consoleLinesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v consoleLinesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	if value == "infinity" {
		return
	}

	// Try to parse as positive integer
	lines := strings.TrimSpace(value)
	n, err := strconv.Atoi(lines)
	if err != nil || n <= 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Console Lines Value",
			fmt.Sprintf("Value must be a positive integer or 'infinity', got: %s", value),
		)
	}
}
