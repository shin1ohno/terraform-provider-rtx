package dns_server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &DNSServerResource{}
	_ resource.ResourceWithImportState = &DNSServerResource{}
)

// NewDNSServerResource creates a new DNS server resource.
func NewDNSServerResource() resource.Resource {
	return &DNSServerResource{}
}

// DNSServerResource defines the resource implementation.
type DNSServerResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *DNSServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_server"
}

// MaxPriorityValue is the maximum valid priority number for DNS server select entries.
const MaxPriorityValue = 65535

// DefaultPriorityStep is the default step between priority numbers.
const DefaultPriorityStep = 10

// Schema defines the schema for the DNS server resource.
func (r *DNSServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DNS server configuration on RTX routers. This is a singleton resource - there is only one DNS server configuration per router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (always 'dns' for this singleton resource)",
				Computed:    true,
			},
			"domain_lookup": schema.BoolAttribute{
				Description: "Enable DNS domain lookup (dns domain lookup on/off)",
				Optional:    true,
				Computed:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "Default domain name for DNS queries (dns domain <name>)",
				Optional:    true,
			},
			"name_servers": schema.ListAttribute{
				Description: "List of DNS server IP addresses (up to 3)",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtMost(3),
				},
			},
			"service_on": schema.BoolAttribute{
				Description: "Enable DNS service (dns service on/off)",
				Optional:    true,
				Computed:    true,
			},
			"private_address_spoof": schema.BoolAttribute{
				Description: "Enable DNS private address spoofing (dns private address spoof on/off)",
				Optional:    true,
				Computed:    true,
			},
			"priority_start": schema.Int64Attribute{
				Description: "Starting priority number for automatic priority calculation in server_select entries. When set, priority numbers are automatically assigned based on definition order. Mutually exclusive with entry-level priority attributes.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, MaxPriorityValue),
				},
			},
			"priority_step": schema.Int64Attribute{
				Description: fmt.Sprintf("Increment value for automatic priority calculation. Only used when priority_start is set. Default is %d.", DefaultPriorityStep),
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(DefaultPriorityStep),
				Validators: []validator.Int64{
					int64validator.Between(1, MaxPriorityValue),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"server_select": schema.ListNestedBlock{
				Description: "Domain-based DNS server selection entries",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"priority": schema.Int64Attribute{
							Description: "Priority for DNS server selection. Lower numbers have higher priority. Required when priority_start is not set (manual mode). Auto-calculated when priority_start is set (auto mode).",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, MaxPriorityValue),
							},
							PlanModifiers: []planmodifier.Int64{
								AutoPriorityModifier(),
							},
						},
						"record_type": schema.StringAttribute{
							Description: "DNS record type to match: a, aaaa, ptr, mx, ns, cname, any",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("a"),
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("a", "aaaa", "ptr", "mx", "ns", "cname", "any"),
							},
						},
						"query_pattern": schema.StringAttribute{
							Description: "Domain pattern to match (e.g., '.', '*.example.com', 'internal.net')",
							Required:    true,
						},
						"original_sender": schema.StringAttribute{
							Description: "Source IP/CIDR restriction for DNS queries",
							Optional:    true,
						},
						"restrict_pp": schema.Int64Attribute{
							Description: "PP session restriction (0 = no restriction)",
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"server": schema.ListNestedBlock{
							Description: "DNS servers for this selector (1-2 servers with per-server EDNS settings)",
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 2),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Description: "DNS server IP address (IPv4 or IPv6)",
										Required:    true,
									},
									"edns": schema.BoolAttribute{
										Description: "Enable EDNS (Extension mechanisms for DNS) for this server",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
						},
					},
				},
			},
			"hosts": schema.ListNestedBlock{
				Description: "Static DNS host entries (dns static)",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Hostname",
							Required:    true,
						},
						"address": schema.StringAttribute{
							Description: "IP address",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *DNSServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DNSServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSServerModel
	var planData DNSServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save a copy of the plan to preserve server_select ordering
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_dns_server", "dns")
	logger := logging.FromContext(ctx)

	// Validate the configuration
	r.validateConfig(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_dns_server").Msgf("Creating DNS server configuration: %+v", config)

	if err := r.client.ConfigureDNS(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure DNS server",
			fmt.Sprintf("Could not configure DNS server: %v", err),
		)
		return
	}

	// Set the singleton ID
	data.ID = types.StringValue("dns")

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reorder server_select to match the plan ordering
	data.reorderServerSelectToMatchPlan(ctx, &planData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *DNSServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSServerModel

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

// read is a helper function that reads the DNS configuration from the router.
func (r *DNSServerResource) read(ctx context.Context, data *DNSServerModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_dns_server", "dns")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_dns_server").Msg("Reading DNS server configuration")

	var config *client.DNSConfig

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractDNSServer()
			if parsed != nil {
				config = convertParsedDNSConfig(parsed)
				logger.Debug().Str("resource", "rtx_dns_server").Msg("Found DNS config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_dns_server").Msg("DNS config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = r.client.GetDNS(ctx)
		if err != nil {
			fwhelpers.AppendDiagError(diagnostics, "Failed to read DNS server configuration", fmt.Sprintf("Could not read DNS server configuration: %v", err))
			return
		}
	}

	data.FromClient(ctx, config, diagnostics)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DNSServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSServerModel
	var planData DNSServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save a copy of the plan to preserve server_select ordering
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_dns_server", "dns")
	logger := logging.FromContext(ctx)

	// Validate the configuration
	r.validateConfig(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_dns_server").Msgf("Updating DNS server configuration: %+v", config)

	if err := r.client.UpdateDNS(ctx, config); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update DNS server configuration",
			fmt.Sprintf("Could not update DNS server configuration: %v", err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reorder server_select to match the plan ordering
	data.reorderServerSelectToMatchPlan(ctx, &planData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DNSServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_dns_server", "dns")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_dns_server").Msg("Deleting (resetting) DNS server configuration")

	if err := r.client.ResetDNS(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to reset DNS server configuration",
			fmt.Sprintf("Could not reset DNS server configuration: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *DNSServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Only accept "dns" as valid import ID (singleton resource)
	if req.ID != "dns" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid import ID format, expected 'dns' for singleton resource, got: %s", req.ID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// convertParsedDNSConfig converts a parser DNSConfig to a client DNSConfig.
func convertParsedDNSConfig(parsed *parsers.DNSConfig) *client.DNSConfig {
	config := &client.DNSConfig{
		DomainLookup: parsed.DomainLookup,
		DomainName:   parsed.DomainName,
		ServiceOn:    parsed.ServiceOn,
		PrivateSpoof: parsed.PrivateSpoof,
		NameServers:  make([]string, len(parsed.NameServers)),
		ServerSelect: make([]client.DNSServerSelect, len(parsed.ServerSelect)),
		Hosts:        make([]client.DNSHost, len(parsed.Hosts)),
	}

	// Copy name servers
	copy(config.NameServers, parsed.NameServers)

	// Convert server select entries
	for i, sel := range parsed.ServerSelect {
		servers := make([]client.DNSServer, len(sel.Servers))
		for j, srv := range sel.Servers {
			servers[j] = client.DNSServer{
				Address: srv.Address,
				EDNS:    srv.EDNS,
			}
		}
		config.ServerSelect[i] = client.DNSServerSelect{
			ID:             sel.ID,
			Servers:        servers,
			RecordType:     sel.RecordType,
			QueryPattern:   sel.QueryPattern,
			OriginalSender: sel.OriginalSender,
			RestrictPP:     sel.RestrictPP,
		}
	}

	// Convert hosts
	for i, host := range parsed.Hosts {
		config.Hosts[i] = client.DNSHost{
			Name:    host.Name,
			Address: host.Address,
		}
	}

	return config
}

// validateConfig validates the DNS server configuration for auto/manual mode consistency.
func (r *DNSServerResource) validateConfig(ctx context.Context, data *DNSServerModel, diagnostics *diag.Diagnostics) {
	priorityStart := fwhelpers.GetInt64Value(data.PriorityStart)
	priorityStep := fwhelpers.GetInt64Value(data.PriorityStep)
	if priorityStep == 0 {
		priorityStep = DefaultPriorityStep
	}

	if data.ServerSelect.IsNull() || data.ServerSelect.IsUnknown() {
		return
	}

	var serverSelects []DNSServerSelectModel
	data.ServerSelect.ElementsAs(ctx, &serverSelects, false)

	autoMode := priorityStart > 0
	usedPriorities := make(map[int]int) // priority -> entry index

	for i, sel := range serverSelects {
		entryPriority := fwhelpers.GetInt64Value(sel.Priority)

		if autoMode {
			// Auto mode: entry-level priority should not be specified
			if entryPriority > 0 {
				diagnostics.AddError(
					"Invalid configuration",
					fmt.Sprintf("server_select[%d]: priority cannot be specified when priority_start is set (auto mode). Remove the priority attribute or use manual mode by removing priority_start", i),
				)
				return
			}

			// Calculate the priority for overflow check
			calculatedPriority := priorityStart + (i * priorityStep)
			if calculatedPriority > MaxPriorityValue {
				diagnostics.AddError(
					"Priority overflow",
					fmt.Sprintf("server_select[%d]: calculated priority %d exceeds maximum value %d. Reduce priority_start or priority_step, or reduce number of entries", i, calculatedPriority, MaxPriorityValue),
				)
				return
			}

			// Check for duplicates
			if prevIdx, exists := usedPriorities[calculatedPriority]; exists {
				diagnostics.AddError(
					"Duplicate priority",
					fmt.Sprintf("server_select[%d]: calculated priority %d conflicts with server_select[%d]. Increase priority_step to avoid collisions", i, calculatedPriority, prevIdx),
				)
				return
			}
			usedPriorities[calculatedPriority] = i
		} else {
			// Manual mode: entry-level priority is required
			if entryPriority <= 0 {
				diagnostics.AddError(
					"Invalid configuration",
					fmt.Sprintf("server_select[%d]: priority must be specified when priority_start is not set (manual mode). Add a priority attribute to each entry or use auto mode by setting priority_start", i),
				)
				return
			}

			// Check for duplicates
			if prevIdx, exists := usedPriorities[int(entryPriority)]; exists {
				diagnostics.AddError(
					"Duplicate priority",
					fmt.Sprintf("server_select[%d]: priority %d is already used by server_select[%d]. Each entry must have a unique priority number", i, entryPriority, prevIdx),
				)
				return
			}
			usedPriorities[int(entryPriority)] = i
		}
	}
}
