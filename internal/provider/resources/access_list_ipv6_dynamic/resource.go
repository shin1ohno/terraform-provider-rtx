package access_list_ipv6_dynamic

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// MaxSequenceValue is the maximum allowed sequence number.
const MaxSequenceValue = 65535

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AccessListIPv6DynamicResource{}
	_ resource.ResourceWithImportState = &AccessListIPv6DynamicResource{}
)

// NewAccessListIPv6DynamicResource creates a new access list IPv6 dynamic resource.
func NewAccessListIPv6DynamicResource() resource.Resource {
	return &AccessListIPv6DynamicResource{}
}

// AccessListIPv6DynamicResource defines the resource implementation.
type AccessListIPv6DynamicResource struct {
	client client.Client
}

// validProtocols lists all valid protocols for dynamic IPv6 filters.
var validProtocols = []string{
	"ftp", "www", "smtp", "pop3", "dns", "domain", "telnet", "ssh",
	"tcp", "udp", "*",
	"tftp", "submission", "https", "imap", "imaps", "pop3s", "smtps",
	"ldap", "ldaps", "bgp", "sip", "ipsec-nat-t", "ntp", "snmp",
	"rtsp", "h323", "pptp", "l2tp", "ike", "esp",
}

// Metadata returns the resource type name.
func (r *AccessListIPv6DynamicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_ipv6_dynamic"
}

// Schema defines the schema for the resource.
func (r *AccessListIPv6DynamicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manages a named collection of IPv6 dynamic (stateful) filters on RTX routers.

Dynamic filters provide stateful packet inspection for various protocols. This resource groups
multiple dynamic IPv6 filter entries under a single name for easier management and reference.

Note: Unlike IPv4 dynamic filters, IPv6 dynamic filters do NOT support the timeout attribute.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Access list name (identifier).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sequence_start": schema.Int64Attribute{
				Description: "Starting sequence number for automatic sequence calculation. When set, sequence numbers are automatically assigned to entries based on their definition order.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, MaxSequenceValue),
				},
			},
			"sequence_step": schema.Int64Attribute{
				Description: fmt.Sprintf("Increment value for automatic sequence calculation. Only used when sequence_start is set. Default is %d.", DefaultSequenceStep),
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(DefaultSequenceStep),
				Validators: []validator.Int64{
					int64validator.Between(1, MaxSequenceValue),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entry": schema.ListNestedBlock{
				Description: "List of dynamic IPv6 filter entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sequence": schema.Int64Attribute{
							Description: "Sequence number (determines order and filter number). Required in manual mode, auto-calculated when sequence_start is set.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"source": schema.StringAttribute{
							Description: "Source IPv6 address or '*' for any. Can be an IPv6 address, network in CIDR notation, or '*'.",
							Required:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination IPv6 address or '*' for any. Can be an IPv6 address, network in CIDR notation, or '*'.",
							Required:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol for stateful inspection. Valid values: ftp, www, smtp, pop3, dns, domain, " +
								"telnet, ssh, tcp, udp, *, tftp, submission, https, imap, imaps, pop3s, smtps, ldap, ldaps, bgp, sip, " +
								"ipsec-nat-t, ntp, snmp, rtsp, h323, pptp, l2tp, ike, esp.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(validProtocols...),
							},
						},
						"syslog": schema.BoolAttribute{
							Description: "Enable syslog logging for this filter.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListIPv6DynamicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListIPv6DynamicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListIPv6DynamicModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_dynamic", name)
	logger := logging.FromContext(ctx)

	// Check for sequence conflicts with existing IPv6 dynamic filters on the router
	r.checkSequenceConflicts(ctx, &data, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_ipv6_dynamic").Msgf("Creating dynamic IPv6 access list: %s", acl.Name)

	if err := r.client.CreateAccessListIPv6Dynamic(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create dynamic IPv6 access list",
			fmt.Sprintf("Could not create dynamic IPv6 access list: %v", err),
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
func (r *AccessListIPv6DynamicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListIPv6DynamicModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// Check if resource was removed
		if data.Name.IsNull() {
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the access list from the router.
func (r *AccessListIPv6DynamicResource) read(ctx context.Context, data *AccessListIPv6DynamicModel, diagnostics *diag.Diagnostics) {
	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_dynamic", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6_dynamic").Msgf("Reading dynamic IPv6 access list: %s", name)

	// Get current sequences from state to filter results
	// This prevents other access lists' filters from leaking into this resource's state
	currentSeqs := data.GetCurrentSequences()

	acl, err := r.client.GetAccessListIPv6Dynamic(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().Str("resource", "rtx_access_list_ipv6_dynamic").Msgf("Dynamic IPv6 access list %s not found, removing from state", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read dynamic IPv6 access list", fmt.Sprintf("Could not read dynamic IPv6 access list %s: %v", name, err))
		return
	}

	data.FromClient(acl, currentSeqs)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListIPv6DynamicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListIPv6DynamicModel
	var state AccessListIPv6DynamicModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_dynamic", name)
	logger := logging.FromContext(ctx)

	// Check for sequence conflicts with existing IPv6 dynamic filters on the router
	// Pass current state sequences so our own sequences are not flagged as conflicts
	currentStateSequences := state.GetFilterNumbers()
	r.checkSequenceConflicts(ctx, &data, currentStateSequences, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_ipv6_dynamic").Msgf("Updating dynamic IPv6 access list: %s", acl.Name)

	if err := r.client.UpdateAccessListIPv6Dynamic(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update dynamic IPv6 access list",
			fmt.Sprintf("Could not update dynamic IPv6 access list: %v", err),
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
func (r *AccessListIPv6DynamicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListIPv6DynamicModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_dynamic", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6_dynamic").Msgf("Deleting dynamic IPv6 access list: %s", name)

	// Collect filter numbers to delete
	filterNums := data.GetFilterNumbers()

	if err := r.client.DeleteAccessListIPv6Dynamic(ctx, name, filterNums); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete dynamic IPv6 access list",
			fmt.Sprintf("Could not delete dynamic IPv6 access list %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListIPv6DynamicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6_dynamic").Msgf("Importing dynamic IPv6 access list: %s", name)

	// Set the name first
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a model and read entries from the router
	// This populates the state with all existing dynamic IPv6 filters
	data := AccessListIPv6DynamicModel{
		Name: types.StringValue(name),
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the full state including entries
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// checkSequenceConflicts checks for sequence conflicts with existing IPv6 dynamic filters on the router.
// currentState contains sequences that this resource already owns (for update operations).
func (r *AccessListIPv6DynamicResource) checkSequenceConflicts(ctx context.Context, data *AccessListIPv6DynamicModel, currentState []int, diagnostics *diag.Diagnostics) {
	logger := logging.FromContext(ctx)

	// Get planned sequences
	plannedSequences := data.GetFilterNumbers()
	if len(plannedSequences) == 0 {
		return
	}

	// Get all existing IPv6 dynamic filter sequences from the router
	existingSequences, err := r.client.GetAllIPv6FilterDynamicSequences(ctx)
	if err != nil {
		// Log warning but don't fail - this is a best-effort check
		logger.Warn().Err(err).Msg("Could not check for sequence conflicts")
		return
	}

	// Check for conflicts
	conflicts := fwhelpers.CheckSequenceConflicts(plannedSequences, existingSequences, currentState)
	if len(conflicts) > 0 {
		diagnostics.AddError(
			"Sequence conflict detected",
			fwhelpers.FormatSequenceConflictError("rtx_access_list_ipv6_dynamic", data.Name.ValueString(), conflicts),
		)
	}
}
