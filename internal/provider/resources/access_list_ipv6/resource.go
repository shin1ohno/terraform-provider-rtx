package access_list_ipv6

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// MaxSequenceValue is the maximum valid sequence number for RTX filters.
// RTX routers support filter numbers up to 2147483647, but practical usage is typically under 1000000.
const MaxSequenceValue = 2147483647

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AccessListIPv6Resource{}
	_ resource.ResourceWithImportState = &AccessListIPv6Resource{}
)

// NewAccessListIPv6Resource creates a new access list IPv6 resource.
func NewAccessListIPv6Resource() resource.Resource {
	return &AccessListIPv6Resource{}
}

// AccessListIPv6Resource defines the resource implementation.
type AccessListIPv6Resource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListIPv6Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_ipv6"
}

// Schema defines the schema for the resource.
func (r *AccessListIPv6Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a group of IPv6 static filters (access list) on RTX routers. " +
			"This resource manages multiple IPv6 filter rules as a single group using the RTX native 'ipv6 filter' command. " +
			"Supports automatic sequence numbering or manual sequence assignment.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "ACL group identifier. This name is used to reference the ACL in other resources and for Terraform state management.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sequence_start": schema.Int64Attribute{
				Description: "Starting sequence number for automatic sequence calculation. When set, sequence numbers are automatically assigned to entries based on their definition order. Mutually exclusive with entry-level sequence attributes.",
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
				Description: "List of IPv6 filter entries. Each entry defines a single filter rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sequence": schema.Int64Attribute{
							Description: "Sequence number determines the order of evaluation. Required when sequence_start is not set (manual mode). Auto-calculated when sequence_start is set (auto mode).",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, MaxSequenceValue),
							},
						},
						"action": schema.StringAttribute{
							Description: "Filter action: pass, reject, restrict, or restrict-log",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("pass", "reject", "restrict", "restrict-log"),
							},
						},
						"source": schema.StringAttribute{
							Description: "Source IPv6 address/prefix (e.g., '2001:db8::/32') or '*' for any",
							Required:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination IPv6 address/prefix (e.g., '2001:db8::1/128') or '*' for any",
							Required:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol: tcp, udp, icmp6, ip, gre, esp, ah, or * for any",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("*"),
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("tcp", "udp", "icmp6", "ip", "gre", "esp", "ah", "*"),
							},
						},
						"source_port": schema.StringAttribute{
							Description: "Source port number, range (e.g., '1024-65535'), or '*' for any. Only valid for TCP/UDP.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("*"),
						},
						"dest_port": schema.StringAttribute{
							Description: "Destination port number, range (e.g., '80'), or '*' for any. Only valid for TCP/UDP.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("*"),
						},
						"log": schema.BoolAttribute{
							Description: "Enable logging when this entry matches traffic.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
			"apply": schema.ListNestedBlock{
				Description: "List of interface bindings. Each apply block binds this ACL to an interface in a specific direction.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"interface": schema.StringAttribute{
							Description: "Interface to apply the ACL to (e.g., lan1, bridge1, pp1, tunnel1).",
							Required:    true,
						},
						"direction": schema.StringAttribute{
							Description: "Direction to apply the ACL: 'in' for incoming traffic, 'out' for outgoing traffic.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("in", "out"),
							},
						},
						"filter_ids": schema.ListAttribute{
							Description: "Specific filter IDs (sequence numbers) to apply in order. If omitted, all entry sequences are applied in order.",
							Optional:    true,
							Computed:    true,
							ElementType: types.Int64Type,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListIPv6Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListIPv6Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListIPv6Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Creating IPv6 access list group: %s", name)

	// Build and create IPv6 filters
	filters := data.ToFilters(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, filter := range filters {
		if err := r.client.CreateIPv6Filter(ctx, filter); err != nil {
			resp.Diagnostics.AddError(
				"Failed to create IPv6 filter",
				fmt.Sprintf("Could not create IPv6 filter %d: %v", filter.Number, err),
			)
			return
		}
	}

	// Handle apply blocks
	if err := r.applyFiltersToInterfaces(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to apply IPv6 filters to interfaces",
			err.Error(),
		)
		return
	}

	// Read back to get computed values
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccessListIPv6Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListIPv6Model

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

// read is a helper function that reads the filters from the router.
func (r *AccessListIPv6Resource) read(ctx context.Context, data *AccessListIPv6Model, diagnostics *diag.Diagnostics) {
	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Reading IPv6 access list group: %s", name)

	// Get the expected sequences from state to query
	sequences := data.GetExpectedSequences()

	// Read each filter
	filters := make([]*client.IPFilter, 0, len(sequences))
	foundAny := false
	for _, seq := range sequences {
		filter, err := r.client.GetIPv6Filter(ctx, seq)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read IPv6 filter", fmt.Sprintf("Could not read IPv6 filter %d: %v", seq, err))
			return
		}
		foundAny = true
		filters = append(filters, filter)
	}

	// If no entries found and we expected some, mark as deleted
	if !foundAny && len(sequences) > 0 {
		logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("IPv6 access list %s not found, removing from state", name)
		data.Name = types.StringNull()
		return
	}

	// Update entries from filters
	data.FromFilters(ctx, filters, diagnostics)

	// Read and update apply blocks
	if err := r.readApplyBlocks(ctx, data); err != nil {
		logger.Warn().Err(err).Msg("Failed to read apply blocks")
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListIPv6Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListIPv6Model
	var state AccessListIPv6Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Updating IPv6 access list group: %s", name)

	// Get old and new sequences
	oldSequences := state.GetExpectedSequences()
	newSequences := data.GetExpectedSequences()

	// Delete removed sequences
	toDelete := FindRemovedSequences(oldSequences, newSequences)
	for _, seq := range toDelete {
		if err := r.client.DeleteIPv6Filter(ctx, seq); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				logger.Warn().Err(err).Msgf("Failed to delete IPv6 filter %d", seq)
			}
		}
	}

	// Create/update filters
	filters := data.ToFilters(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, filter := range filters {
		if err := r.client.UpdateIPv6Filter(ctx, filter); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update IPv6 filter",
				fmt.Sprintf("Could not update IPv6 filter %d: %v", filter.Number, err),
			)
			return
		}
	}

	// Handle apply changes - remove old applies
	for i := range state.Apply {
		apply := &state.Apply[i]
		iface := fwhelpers.GetStringValue(apply.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(apply.Direction))

		if err := r.client.RemoveIPv6FiltersFromInterface(ctx, iface, direction); err != nil {
			logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
		}
	}

	// Apply new applies
	if err := r.applyFiltersToInterfaces(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to apply IPv6 filters to interfaces",
			err.Error(),
		)
		return
	}

	// Read back to get computed values
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccessListIPv6Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListIPv6Model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Deleting IPv6 access list group: %s", name)

	// First remove apply blocks to free up filter references
	for i := range data.Apply {
		apply := &data.Apply[i]
		iface := fwhelpers.GetStringValue(apply.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(apply.Direction))

		if err := r.client.RemoveIPv6FiltersFromInterface(ctx, iface, direction); err != nil {
			logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
		}
	}

	// Get sequences to delete
	sequences := data.GetExpectedSequences()

	// Delete all entries
	for _, seq := range sequences {
		if err := r.client.DeleteIPv6Filter(ctx, seq); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				resp.Diagnostics.AddError(
					"Failed to delete IPv6 filter",
					fmt.Sprintf("Could not delete IPv6 filter %d: %v", seq, err),
				)
				return
			}
		}
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListIPv6Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: name:seq1,seq2,seq3 or just name
	importID := req.ID
	parts := strings.Split(importID, ":")

	name := parts[0]
	var sequences []int

	if len(parts) > 1 {
		// Parse comma-separated sequence numbers
		seqStrs := strings.Split(parts[1], ",")
		for _, s := range seqStrs {
			seq, err := strconv.Atoi(strings.TrimSpace(s))
			if err == nil {
				sequences = append(sequences, seq)
			}
		}
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Importing IPv6 access list: %s with sequences %v", name, sequences)

	// Set the name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)

	// If sequences provided, import those specific ones
	if len(sequences) > 0 {
		entries := make([]EntryModel, 0, len(sequences))

		for _, seq := range sequences {
			filter, err := r.client.GetIPv6Filter(ctx, seq)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					continue
				}
				resp.Diagnostics.AddError(
					"Failed to read IPv6 filter",
					fmt.Sprintf("Could not read IPv6 filter %d: %v", seq, err),
				)
				return
			}

			entry := EntryModel{
				Sequence:    types.Int64Value(int64(filter.Number)),
				Action:      types.StringValue(filter.Action),
				Source:      types.StringValue(filter.SourceAddress),
				Destination: types.StringValue(filter.DestAddress),
				Protocol:    types.StringValue(filter.Protocol),
				SourcePort:  types.StringValue(normalizePort(filter.SourcePort)),
				DestPort:    types.StringValue(normalizePort(filter.DestPort)),
				Log:         types.BoolValue(false),
			}
			entries = append(entries, entry)
		}

		if len(entries) == 0 {
			resp.Diagnostics.AddError(
				"No IPv6 filters found",
				fmt.Sprintf("No IPv6 filters found with sequences %v", sequences),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entry"), entries)...)
	}
}

// applyFiltersToInterfaces applies filters to interfaces based on apply blocks.
func (r *AccessListIPv6Resource) applyFiltersToInterfaces(ctx context.Context, data *AccessListIPv6Model) error {
	if len(data.Apply) == 0 {
		return nil
	}

	for i := range data.Apply {
		apply := &data.Apply[i]
		iface := fwhelpers.GetStringValue(apply.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(apply.Direction))
		filterIDs := data.GetApplyFilterIDs(apply)

		if len(filterIDs) > 0 {
			if err := r.client.ApplyIPv6FiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
				return fmt.Errorf("failed to apply filters to interface %s %s: %w", iface, direction, err)
			}
		}
	}

	return nil
}

// readApplyBlocks reads apply block state from the router.
func (r *AccessListIPv6Resource) readApplyBlocks(ctx context.Context, data *AccessListIPv6Model) error {
	if len(data.Apply) == 0 {
		return nil
	}

	for i := range data.Apply {
		apply := &data.Apply[i]
		iface := fwhelpers.GetStringValue(apply.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(apply.Direction))

		filterIDs, err := r.client.GetIPv6InterfaceFilters(ctx, iface, direction)
		if err != nil {
			return fmt.Errorf("failed to get filters for interface %s %s: %w", iface, direction, err)
		}

		SetApplyFilterIDs(apply, filterIDs)
	}

	return nil
}
