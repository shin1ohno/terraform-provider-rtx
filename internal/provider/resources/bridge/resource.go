package bridge

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
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &BridgeResource{}
	_ resource.ResourceWithImportState = &BridgeResource{}
)

// NewBridgeResource creates a new bridge resource.
func NewBridgeResource() resource.Resource {
	return &BridgeResource{}
}

// BridgeResource defines the resource implementation.
type BridgeResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *BridgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bridge"
}

// Schema defines the schema for the resource.
func (r *BridgeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Ethernet bridge configurations on RTX routers. Bridges combine multiple interfaces into a single Layer 2 broadcast domain.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The bridge name (e.g., 'bridge1', 'bridge2'). Must be in format 'bridgeN'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^bridge\d+$`),
						"must be in format 'bridgeN' (e.g., 'bridge1', 'bridge2')",
					),
				},
			},
			"interface_name": schema.StringAttribute{
				Description: "The bridge interface name. Same as 'name', provided for consistency with other resources.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"members": schema.ListAttribute{
				Description: "List of member interfaces to include in the bridge (e.g., ['lan1', 'tunnel1']). Valid formats include 'lanN', 'lanN/N' (VLAN), 'tunnelN', 'ppN', 'loopbackN'.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						bridgeMemberValidator{},
					),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *BridgeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *BridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BridgeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_bridge", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	bridge := data.ToClient()
	logger.Debug().Str("resource", "rtx_bridge").Msgf("Creating bridge: %+v", bridge)

	if err := r.client.CreateBridge(ctx, bridge); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create bridge",
			fmt.Sprintf("Could not create bridge: %v", err),
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
func (r *BridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BridgeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was deleted outside of Terraform
	if data.Name.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the bridge from the router.
func (r *BridgeResource) read(ctx context.Context, data *BridgeModel, diagnostics *diag.Diagnostics) {
	name := fwhelpers.GetStringValue(data.Name)

	ctx = logging.WithResource(ctx, "rtx_bridge", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_bridge").Msgf("Reading bridge: %s", name)

	var bridge *client.BridgeConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract bridges from parsed config
			bridges := parsedConfig.ExtractBridges()
			for i := range bridges {
				if bridges[i].Name == name {
					bridge = convertParsedBridgeConfig(&bridges[i])
					logger.Debug().Str("resource", "rtx_bridge").Msg("Found bridge in SFTP cache")
					break
				}
			}
		}
		if bridge == nil {
			// Bridge not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_bridge").Msg("Bridge not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or bridge not found in cache
	if bridge == nil {
		var err error
		bridge, err = r.client.GetBridge(ctx, name)
		if err != nil {
			// Check if bridge doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_bridge").Msgf("Bridge %s not found", name)
				// Resource has been deleted outside of Terraform
				data.Name = types.StringNull()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read bridge", fmt.Sprintf("Could not read bridge %s: %v", name, err))
			return
		}
	}

	data.FromClient(bridge)
}

// convertParsedBridgeConfig converts a parser BridgeConfig to a client BridgeConfig
func convertParsedBridgeConfig(parsed *parsers.BridgeConfig) *client.BridgeConfig {
	return &client.BridgeConfig{
		Name:    parsed.Name,
		Members: parsed.Members,
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *BridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BridgeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_bridge", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	bridge := data.ToClient()
	logger.Debug().Str("resource", "rtx_bridge").Msgf("Updating bridge: %+v", bridge)

	if err := r.client.UpdateBridge(ctx, bridge); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update bridge",
			fmt.Sprintf("Could not update bridge: %v", err),
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
func (r *BridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BridgeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)

	ctx = logging.WithResource(ctx, "rtx_bridge", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_bridge").Msgf("Deleting bridge: %s", name)

	if err := r.client.DeleteBridge(ctx, name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete bridge",
			fmt.Sprintf("Could not delete bridge %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *BridgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID should be the bridge name (e.g., "bridge1")
	importID := req.ID

	// Validate bridge name format
	validNamePattern := regexp.MustCompile(`^bridge\d+$`)
	if !validNamePattern.MatchString(importID) {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected bridge name in format 'bridgeN' (e.g., 'bridge1'), got %q", importID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// bridgeMemberValidator validates a bridge member interface name.
type bridgeMemberValidator struct{}

func (v bridgeMemberValidator) Description(ctx context.Context) string {
	return "must be a valid interface name (lan*, lan*/*, tunnel*, pp*, loopback*, bridge*)"
}

func (v bridgeMemberValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v bridgeMemberValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Bridge Member",
			"Bridge member cannot be empty",
		)
		return
	}

	// Valid member patterns
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^lan\d+$`),      // lan1, lan2, etc.
		regexp.MustCompile(`^lan\d+/\d+$`),  // lan1/1 (VLAN interfaces)
		regexp.MustCompile(`^tunnel\d+$`),   // tunnel1, tunnel2, etc.
		regexp.MustCompile(`^pp\d+$`),       // pp1, pp2, etc.
		regexp.MustCompile(`^loopback\d+$`), // loopback1, etc.
		regexp.MustCompile(`^bridge\d+$`),   // nested bridge (rare)
	}

	for _, pattern := range validPatterns {
		if pattern.MatchString(value) {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Bridge Member",
		fmt.Sprintf("Value %q must be a valid interface name (lan*, lan*/*, tunnel*, pp*, loopback*, bridge*)", value),
	)
}
