package dhcp_scope

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &DHCPScopeResource{}
	_ resource.ResourceWithImportState = &DHCPScopeResource{}
)

// NewDHCPScopeResource creates a new DHCP scope resource.
func NewDHCPScopeResource() resource.Resource {
	return &DHCPScopeResource{}
}

// DHCPScopeResource defines the resource implementation.
type DHCPScopeResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *DHCPScopeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_scope"
}

// Schema defines the schema for the resource.
func (r *DHCPScopeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DHCP scopes on RTX routers. A DHCP scope defines the IP address range and associated network parameters for DHCP address allocation.",
		Attributes: map[string]schema.Attribute{
			"scope_id": schema.Int64Attribute{
				Description: "The DHCP scope ID (positive integer).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"network": schema.StringAttribute{
				Description: "The network address in CIDR notation (e.g., '192.168.1.0/24').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"range_start": schema.StringAttribute{
				Description: "Start IP address of the DHCP allocation range (parsed from IP range format).",
				Optional:    true,
				Computed:    true,
			},
			"range_end": schema.StringAttribute{
				Description: "End IP address of the DHCP allocation range (parsed from IP range format).",
				Optional:    true,
				Computed:    true,
			},
			"lease_time": schema.StringAttribute{
				Description: "DHCP lease duration in Go duration format (e.g., '72h', '30m') or 'infinite'.",
				Optional:    true,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"exclude_ranges": schema.ListNestedBlock{
				Description: "IP address ranges to exclude from DHCP allocation.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"start": schema.StringAttribute{
							Description: "Start IP address of the exclusion range.",
							Required:    true,
						},
						"end": schema.StringAttribute{
							Description: "End IP address of the exclusion range.",
							Required:    true,
						},
					},
				},
			},
			"options": schema.SingleNestedBlock{
				Description: "DHCP options for client configuration (Cisco-compatible naming).",
				Attributes: map[string]schema.Attribute{
					"routers": schema.ListAttribute{
						Description: "Default gateway addresses for DHCP clients (maximum 3).",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtMost(3),
						},
					},
					"dns_servers": schema.ListAttribute{
						Description: "DNS server addresses for DHCP clients (maximum 3).",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtMost(3),
						},
					},
					"domain_name": schema.StringAttribute{
						Description: "Domain name for DHCP clients.",
						Optional:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *DHCPScopeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*fwhelpers.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *fwhelpers.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *DHCPScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DHCPScopeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource context for logging
	ctx = logging.WithResource(ctx, "rtx_dhcp_scope", strconv.FormatInt(data.ScopeID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	scope := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("Creating DHCP scope: %+v", scope)

	if err := r.client.CreateDHCPScope(ctx, scope); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create DHCP scope",
			fmt.Sprintf("Could not create DHCP scope: %v", err),
		)
		return
	}

	// Read back the created resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *DHCPScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DHCPScopeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// Check if resource was removed (ScopeID set to null)
		if data.ScopeID.IsNull() {
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the scope from the router.
func (r *DHCPScopeResource) read(ctx context.Context, data *DHCPScopeModel, diagnostics *diag.Diagnostics) {
	scopeID := int(data.ScopeID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_dhcp_scope", strconv.Itoa(scopeID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("Reading DHCP scope: %d", scopeID)

	var scope *client.DHCPScope
	var err error

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, cacheErr := r.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract DHCP scopes from parsed config
			scopes := parsedConfig.ExtractDHCPScopes()
			for i := range scopes {
				if scopes[i].ScopeID == scopeID {
					scope = convertParsedDHCPScope(&scopes[i])
					logger.Debug().Str("resource", "rtx_dhcp_scope").Msg("Found scope in SFTP cache")
					break
				}
			}
		}
		if scope == nil {
			// Scope not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_dhcp_scope").Msg("Scope not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or scope not found in cache
	if scope == nil {
		scope, err = r.client.GetDHCPScope(ctx, scopeID)
		if err != nil {
			// Check if scope doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("DHCP scope %d not found, removing from state", scopeID)
				data.ScopeID = types.Int64Null()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read DHCP scope", fmt.Sprintf("Could not read DHCP scope %d: %v", scopeID, err))
			return
		}
	}

	// Update data from the scope
	data.FromClient(ctx, scope, diagnostics)
}

// convertParsedDHCPScope converts a parser DHCPScope to a client DHCPScope.
func convertParsedDHCPScope(parsed *parsers.DHCPScope) *client.DHCPScope {
	scope := &client.DHCPScope{
		ScopeID:    parsed.ScopeID,
		Network:    parsed.Network,
		RangeStart: parsed.RangeStart,
		RangeEnd:   parsed.RangeEnd,
		LeaseTime:  parsed.LeaseTime,
		Options: client.DHCPScopeOptions{
			Routers:    parsed.Options.Routers,
			DNSServers: parsed.Options.DNSServers,
			DomainName: parsed.Options.DomainName,
		},
		ExcludeRanges: make([]client.ExcludeRange, len(parsed.ExcludeRanges)),
	}
	for i, r := range parsed.ExcludeRanges {
		scope.ExcludeRanges[i] = client.ExcludeRange{
			Start: r.Start,
			End:   r.End,
		}
	}
	return scope
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DHCPScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DHCPScopeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_dhcp_scope", strconv.FormatInt(data.ScopeID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	scope := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("Updating DHCP scope: %+v", scope)

	if err := r.client.UpdateDHCPScope(ctx, scope); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update DHCP scope",
			fmt.Sprintf("Could not update DHCP scope: %v", err),
		)
		return
	}

	// Read back the updated resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DHCPScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DHCPScopeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeID := int(data.ScopeID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_dhcp_scope", strconv.Itoa(scopeID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("Deleting DHCP scope: %d", scopeID)

	if err := r.client.DeleteDHCPScope(ctx, scopeID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete DHCP scope",
			fmt.Sprintf("Could not delete DHCP scope %d: %v", scopeID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *DHCPScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scopeID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected scope ID as integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope_id"), int64(scopeID))...)
}
