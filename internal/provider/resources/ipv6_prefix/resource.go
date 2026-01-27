package ipv6_prefix

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	_ resource.Resource                   = &IPv6PrefixResource{}
	_ resource.ResourceWithImportState    = &IPv6PrefixResource{}
	_ resource.ResourceWithValidateConfig = &IPv6PrefixResource{}
)

// NewIPv6PrefixResource creates a new IPv6 prefix resource.
func NewIPv6PrefixResource() resource.Resource {
	return &IPv6PrefixResource{}
}

// IPv6PrefixResource defines the resource implementation.
type IPv6PrefixResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *IPv6PrefixResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_prefix"
}

// Schema defines the schema for the resource.
func (r *IPv6PrefixResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IPv6 prefix definitions on RTX routers. IPv6 prefixes can be configured as static, RA-derived, or DHCPv6-PD delegated.",
		Attributes: map[string]schema.Attribute{
			"prefix_id": schema.Int64Attribute{
				Description: "The IPv6 prefix ID (1-255)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 255),
				},
			},
			"prefix": schema.StringAttribute{
				Description: "Static IPv6 prefix value (e.g., '2001:db8::') - required when source is 'static'",
				Optional:    true,
				Validators: []validator.String{
					ipv6PrefixValidator{},
				},
			},
			"prefix_length": schema.Int64Attribute{
				Description: "Prefix length in bits (1-128)",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 128),
				},
			},
			"source": schema.StringAttribute{
				Description: "Prefix source type: 'static', 'ra' (Router Advertisement), or 'dhcpv6-pd' (Prefix Delegation)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("static", "ra", "dhcpv6-pd"),
				},
			},
			"interface": schema.StringAttribute{
				Description: "Source interface name (required for 'ra' and 'dhcpv6-pd' sources, e.g., 'lan2', 'pp1')",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// ipv6PrefixValidator validates that a string is a valid IPv6 prefix.
type ipv6PrefixValidator struct{}

func (v ipv6PrefixValidator) Description(ctx context.Context) string {
	return "must be a valid IPv6 prefix (e.g., '2001:db8::')"
}

func (v ipv6PrefixValidator) MarkdownDescription(ctx context.Context) string {
	return "must be a valid IPv6 prefix (e.g., `2001:db8::`)"
}

func (v ipv6PrefixValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return // Allow empty for non-static sources
	}

	// Basic validation - should contain only valid IPv6 characters
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == ':') {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid IPv6 Prefix",
				fmt.Sprintf("Value %q contains invalid characters for an IPv6 prefix", value),
			)
			return
		}
	}

	// Should contain at least one colon
	if !strings.Contains(value, ":") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv6 Prefix",
			fmt.Sprintf("Value %q must be a valid IPv6 prefix (e.g., '2001:db8::')", value),
		)
	}
}

// ValidateConfig performs custom validation on the resource configuration.
func (r *IPv6PrefixResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data IPv6PrefixModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source := data.Source.ValueString()
	prefix := data.Prefix.ValueString()
	iface := data.Interface.ValueString()

	switch source {
	case "static":
		if data.Prefix.IsNull() || prefix == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("prefix"),
				"Missing Required Attribute",
				"'prefix' is required when source is 'static'",
			)
		}
		if !data.Interface.IsNull() && iface != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("interface"),
				"Invalid Attribute Combination",
				"'interface' should not be set when source is 'static'",
			)
		}
	case "ra", "dhcpv6-pd":
		if data.Interface.IsNull() || iface == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("interface"),
				"Missing Required Attribute",
				fmt.Sprintf("'interface' is required when source is '%s'", source),
			)
		}
		if !data.Prefix.IsNull() && prefix != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("prefix"),
				"Invalid Attribute Combination",
				fmt.Sprintf("'prefix' should not be set when source is '%s' (it is derived dynamically)", source),
			)
		}
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPv6PrefixResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IPv6PrefixResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPv6PrefixModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefixID := strconv.Itoa(fwhelpers.GetInt64Value(data.PrefixID))
	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", prefixID)
	logger := logging.FromContext(ctx)

	prefix := data.ToClient()
	logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Creating IPv6 prefix: %+v", prefix)

	if err := r.client.CreateIPv6Prefix(ctx, prefix); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create IPv6 prefix",
			fmt.Sprintf("Could not create IPv6 prefix: %v", err),
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
func (r *IPv6PrefixResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPv6PrefixModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was removed
	if data.PrefixID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the prefix from the router.
func (r *IPv6PrefixResource) read(ctx context.Context, data *IPv6PrefixModel, diagnostics *diag.Diagnostics) {
	prefixID := fwhelpers.GetInt64Value(data.PrefixID)

	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", strconv.Itoa(prefixID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Reading IPv6 prefix: %d", prefixID)

	var prefix *client.IPv6Prefix
	var err error

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, cacheErr := r.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract IPv6 prefixes from parsed config
			prefixes := parsedConfig.ExtractIPv6Prefixes()
			for i := range prefixes {
				if prefixes[i].ID == prefixID {
					prefix = convertParsedIPv6Prefix(&prefixes[i])
					logger.Debug().Str("resource", "rtx_ipv6_prefix").Msg("Found IPv6 prefix in SFTP cache")
					break
				}
			}
		}
		if prefix == nil {
			// Prefix not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_ipv6_prefix").Msg("IPv6 prefix not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or prefix not found in cache
	if prefix == nil {
		prefix, err = r.client.GetIPv6Prefix(ctx, prefixID)
		if err != nil {
			// Check if prefix doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("IPv6 prefix %d not found, removing from state", prefixID)
				data.PrefixID = types.Int64Null()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read IPv6 prefix", fmt.Sprintf("Could not read IPv6 prefix %d: %v", prefixID, err))
			return
		}
	}

	data.FromClient(prefix)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPv6PrefixResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IPv6PrefixModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefixID := strconv.Itoa(fwhelpers.GetInt64Value(data.PrefixID))
	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", prefixID)
	logger := logging.FromContext(ctx)

	prefix := data.ToClient()
	logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Updating IPv6 prefix: %+v", prefix)

	if err := r.client.UpdateIPv6Prefix(ctx, prefix); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IPv6 prefix",
			fmt.Sprintf("Could not update IPv6 prefix: %v", err),
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
func (r *IPv6PrefixResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPv6PrefixModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefixID := fwhelpers.GetInt64Value(data.PrefixID)

	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", strconv.Itoa(prefixID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Deleting IPv6 prefix: %d", prefixID)

	if err := r.client.DeleteIPv6Prefix(ctx, prefixID); err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete IPv6 prefix",
			fmt.Sprintf("Could not delete IPv6 prefix %d: %v", prefixID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *IPv6PrefixResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID as prefix_id
	prefixID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid import ID format, expected prefix_id (integer): %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prefix_id"), prefixID)...)
}

// convertParsedIPv6Prefix converts a parser IPv6Prefix to a client IPv6Prefix.
func convertParsedIPv6Prefix(parsed *parsers.IPv6Prefix) *client.IPv6Prefix {
	return &client.IPv6Prefix{
		ID:           parsed.ID,
		Prefix:       parsed.Prefix,
		PrefixLength: parsed.PrefixLength,
		Source:       parsed.Source,
		Interface:    parsed.Interface,
	}
}
