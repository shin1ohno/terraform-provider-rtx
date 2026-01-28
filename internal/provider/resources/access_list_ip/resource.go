package access_list_ip

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AccessListIPResource{}
	_ resource.ResourceWithImportState = &AccessListIPResource{}
)

// MaxSequenceValue is the maximum valid sequence number for RTX filters.
// RTX routers support filter numbers up to 2147483647, but practical usage is typically under 1000000.
const MaxSequenceValue = 2147483647

// NewAccessListIPResource creates a new IP access list resource.
func NewAccessListIPResource() resource.Resource {
	return &AccessListIPResource{}
}

// AccessListIPResource defines the resource implementation.
type AccessListIPResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_ip"
}

// Schema defines the schema for the resource.
func (r *AccessListIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a group of IPv4 static filters (access list) on RTX routers. " +
			"This resource manages multiple IP filter rules as a single group using the RTX native 'ip filter' command. " +
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
							Validators: []validator.List{
								listvalidator.ValueInt64sAre(
									int64validator.Between(1, MaxSequenceValue),
								),
							},
						},
					},
				},
			},
			"entry": schema.ListNestedBlock{
				Description: "List of IP filter entries. Each entry defines a single filter rule.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
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
							Description: "Source IP address/network in CIDR notation (e.g., '10.0.0.0/8') or '*' for any",
							Required:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination IP address/network in CIDR notation (e.g., '192.168.1.0/24') or '*' for any",
							Required:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol: tcp, udp, icmp, ip, gre, esp, ah, or * for any",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("*"),
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("tcp", "udp", "udp,tcp", "tcp,udp", "icmp", "ip", "gre", "esp", "ah", "tcpfin", "tcprst", "*"),
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
						"established": schema.BoolAttribute{
							Description: "Match established TCP connections only. Only valid for TCP protocol.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
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
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListIPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Creating IP access list group: %s", name)

	// Validate the configuration
	r.validateConfig(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for sequence conflicts with existing filters on the router
	r.checkSequenceConflicts(ctx, &data, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build and create IP filters
	filters := data.ToClientFilters()
	for _, filter := range filters {
		if err := r.client.CreateIPFilter(ctx, filter); err != nil {
			resp.Diagnostics.AddError(
				"Failed to create IP filter",
				fmt.Sprintf("Could not create IP filter %d: %v", filter.Number, err),
			)
			return
		}
	}

	// Handle apply blocks
	if err := r.applyFiltersToInterfaces(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to apply IP filters to interfaces",
			err.Error(),
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
func (r *AccessListIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListIPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was deleted externally
	if data.Name.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the filters from the router.
func (r *AccessListIPResource) read(ctx context.Context, data *AccessListIPModel, diagnostics *diag.Diagnostics) {
	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Reading IP access list group: %s", name)

	// Get the expected sequences from state to query
	sequences := data.GetExpectedSequences()

	// Read each filter
	filters := make([]client.IPFilter, 0, len(sequences))
	foundAny := false
	for _, seq := range sequences {
		filter, err := r.client.GetIPFilter(ctx, seq)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read IP filter", fmt.Sprintf("Could not read IP filter %d: %v", seq, err))
			return
		}
		foundAny = true
		filters = append(filters, *filter)
	}

	// If no entries found and we expected some, mark as deleted
	if !foundAny && len(sequences) > 0 {
		logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("IP access list %s not found, removing from state", name)
		data.Name = types.StringNull()
		return
	}

	// Set entries
	data.SetEntriesFromFilters(filters)

	// Read and set apply blocks
	if err := r.readApplyBlocks(ctx, data); err != nil {
		logger.Warn().Err(err).Msg("Failed to read apply blocks")
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListIPModel
	var state AccessListIPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Updating IP access list group: %s", name)

	// Validate the configuration
	r.validateConfig(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get old and new sequences
	oldSequences := state.GetExpectedSequences()
	newSequences := data.GetExpectedSequences()

	// Check for sequence conflicts with existing filters on the router
	// Pass oldSequences as currentState so our own sequences are not flagged as conflicts
	r.checkSequenceConflicts(ctx, &data, oldSequences, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete removed sequences
	toDelete := findRemovedSequences(oldSequences, newSequences)
	for _, seq := range toDelete {
		if err := r.client.DeleteIPFilter(ctx, seq); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				logger.Warn().Err(err).Msgf("Failed to delete IP filter %d", seq)
			}
		}
	}

	// Create/update filters
	filters := data.ToClientFilters()
	for _, filter := range filters {
		if err := r.client.UpdateIPFilter(ctx, filter); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update IP filter",
				fmt.Sprintf("Could not update IP filter %d: %v", filter.Number, err),
			)
			return
		}
	}

	// Handle apply changes
	oldApplies := state.GetApplies()
	newApplies := data.GetApplies()

	// Remove old applies
	for _, a := range oldApplies {
		iface := fwhelpers.GetStringValue(a.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(a.Direction))

		if err := r.client.RemoveIPFiltersFromInterface(ctx, iface, direction); err != nil {
			logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
		}
	}

	// Apply new applies
	for _, a := range newApplies {
		iface := fwhelpers.GetStringValue(a.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(a.Direction))
		filterIDs := r.extractFilterIDs(a, &data)

		if len(filterIDs) > 0 {
			if err := r.client.ApplyIPFiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
				resp.Diagnostics.AddError(
					"Failed to apply filters to interface",
					fmt.Sprintf("Could not apply filters to interface %s %s: %v", iface, direction, err),
				)
				return
			}
		}
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccessListIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListIPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Deleting IP access list group: %s", name)

	// First remove apply blocks to free up filter references
	applies := data.GetApplies()
	for _, a := range applies {
		iface := fwhelpers.GetStringValue(a.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(a.Direction))

		if err := r.client.RemoveIPFiltersFromInterface(ctx, iface, direction); err != nil {
			logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
		}
	}

	// Get sequences to delete
	sequences := data.GetExpectedSequences()

	// Delete all entries
	for _, seq := range sequences {
		if err := r.client.DeleteIPFilter(ctx, seq); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				resp.Diagnostics.AddError(
					"Failed to delete IP filter",
					fmt.Sprintf("Could not delete IP filter %d: %v", seq, err),
				)
				return
			}
		}
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: name:seq1,seq2,seq3 or just name
	importID := req.ID
	parts := strings.Split(importID, ":")

	name := parts[0]
	var sequences []int

	if len(parts) > 1 {
		// Parse comma-separated sequence numbers
		seqStrs := strings.Split(parts[1], ",")
		for _, s := range seqStrs {
			var seq int
			if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &seq); err == nil {
				sequences = append(sequences, seq)
			}
		}
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip").Msgf("Importing IP access list: %s with sequences %v", name, sequences)

	// Set name in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)

	// If sequences provided, import those specific ones
	if len(sequences) > 0 {
		filters := make([]client.IPFilter, 0, len(sequences))

		for _, seq := range sequences {
			filter, err := r.client.GetIPFilter(ctx, seq)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					continue
				}
				resp.Diagnostics.AddError(
					"Failed to read IP filter",
					fmt.Sprintf("Could not read IP filter %d: %v", seq, err),
				)
				return
			}
			filters = append(filters, *filter)
		}

		if len(filters) == 0 {
			resp.Diagnostics.AddError(
				"No IP filters found",
				fmt.Sprintf("No IP filters found with sequences %v", sequences),
			)
			return
		}

		// Build entry list
		entryValues := make([]attr.Value, len(filters))
		for i, filter := range filters {
			entry := EntryModel{
				Sequence:    types.Int64Value(int64(filter.Number)),
				Action:      types.StringValue(filter.Action),
				Source:      types.StringValue(filter.SourceAddress),
				Destination: types.StringValue(filter.DestAddress),
				Protocol:    types.StringValue(normalizePort(filter.Protocol)),
				SourcePort:  types.StringValue(normalizePort(filter.SourcePort)),
				DestPort:    types.StringValue(normalizePort(filter.DestPort)),
				Established: types.BoolValue(filter.Established),
				Log:         types.BoolValue(false),
			}
			entryValues[i] = entryToObjectValue(entry)
		}

		entryList := types.ListValueMust(types.ObjectType{AttrTypes: EntryModelAttrTypes()}, entryValues)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entry"), entryList)...)
	}
}

// validateConfig validates the ACL configuration for auto/manual mode consistency.
func (r *AccessListIPResource) validateConfig(ctx context.Context, data *AccessListIPModel, diagnostics *diag.Diagnostics) {
	sequenceStart := fwhelpers.GetInt64Value(data.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(data.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	if data.Entry.IsNull() || data.Entry.IsUnknown() {
		return
	}

	var entries []EntryModel
	data.Entry.ElementsAs(ctx, &entries, false)

	autoMode := sequenceStart > 0
	usedSequences := make(map[int]int) // sequence -> entry index

	for i, entry := range entries {
		entrySeq := fwhelpers.GetInt64Value(entry.Sequence)
		protocol := strings.ToLower(fwhelpers.GetStringValue(entry.Protocol))
		established := fwhelpers.GetBoolValue(entry.Established)
		sourcePort := fwhelpers.GetStringValue(entry.SourcePort)
		destPort := fwhelpers.GetStringValue(entry.DestPort)

		if autoMode {
			// Auto mode: entry-level sequence should not be specified
			if entrySeq > 0 {
				diagnostics.AddError(
					"Invalid configuration",
					fmt.Sprintf("entry[%d]: sequence cannot be specified when sequence_start is set (auto mode). Remove the sequence attribute or use manual mode by removing sequence_start", i),
				)
				return
			}

			// Calculate the sequence for overflow check
			calculatedSeq := sequenceStart + (i * sequenceStep)
			if calculatedSeq > MaxSequenceValue {
				diagnostics.AddError(
					"Sequence overflow",
					fmt.Sprintf("entry[%d]: calculated sequence %d exceeds maximum value %d. Reduce sequence_start or sequence_step, or reduce number of entries", i, calculatedSeq, MaxSequenceValue),
				)
				return
			}

			// Check for duplicates
			if prevIdx, exists := usedSequences[calculatedSeq]; exists {
				diagnostics.AddError(
					"Duplicate sequence",
					fmt.Sprintf("entry[%d]: calculated sequence %d conflicts with entry[%d]. Increase sequence_step to avoid collisions", i, calculatedSeq, prevIdx),
				)
				return
			}
			usedSequences[calculatedSeq] = i
		} else {
			// Manual mode: entry-level sequence is required
			if entrySeq <= 0 {
				diagnostics.AddError(
					"Invalid configuration",
					fmt.Sprintf("entry[%d]: sequence must be specified when sequence_start is not set (manual mode). Add a sequence attribute to each entry or use auto mode by setting sequence_start", i),
				)
				return
			}

			// Check for duplicates
			if prevIdx, exists := usedSequences[entrySeq]; exists {
				diagnostics.AddError(
					"Duplicate sequence",
					fmt.Sprintf("entry[%d]: sequence %d is already used by entry[%d]. Each entry must have a unique sequence number", i, entrySeq, prevIdx),
				)
				return
			}
			usedSequences[entrySeq] = i
		}

		// Established is only valid for TCP
		if established && protocol != "tcp" {
			diagnostics.AddError(
				"Invalid configuration",
				fmt.Sprintf("entry[%d]: established can only be set to true for tcp protocol", i),
			)
			return
		}

		// Port specifications valid for TCP/UDP and TCP-based protocols
		tcpBasedProtocols := protocol == "tcp" || protocol == "udp" || protocol == "tcp,udp" || protocol == "udp,tcp" || protocol == "tcpfin" || protocol == "tcprst"
		if !tcpBasedProtocols {
			if sourcePort != "*" && sourcePort != "" {
				diagnostics.AddError(
					"Invalid configuration",
					fmt.Sprintf("entry[%d]: source_port can only be specified for tcp, udp, tcpfin, or tcprst protocols", i),
				)
				return
			}
			if destPort != "*" && destPort != "" {
				diagnostics.AddError(
					"Invalid configuration",
					fmt.Sprintf("entry[%d]: dest_port can only be specified for tcp, udp, tcpfin, or tcprst protocols", i),
				)
				return
			}
		}
	}
}

// applyFiltersToInterfaces handles the apply blocks during create.
func (r *AccessListIPResource) applyFiltersToInterfaces(ctx context.Context, data *AccessListIPModel) error {
	applies := data.GetApplies()

	for _, a := range applies {
		iface := fwhelpers.GetStringValue(a.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(a.Direction))
		filterIDs := r.extractFilterIDs(a, data)

		if len(filterIDs) > 0 {
			if err := r.client.ApplyIPFiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
				return fmt.Errorf("failed to apply filters to interface %s %s: %w", iface, direction, err)
			}
		}
	}

	return nil
}

// readApplyBlocks reads apply block state from the router.
func (r *AccessListIPResource) readApplyBlocks(ctx context.Context, data *AccessListIPModel) error {
	applies := data.GetApplies()
	if len(applies) == 0 {
		return nil
	}

	updatedApplies := make([]ApplyModel, 0, len(applies))

	for _, a := range applies {
		iface := fwhelpers.GetStringValue(a.Interface)
		direction := strings.ToLower(fwhelpers.GetStringValue(a.Direction))

		filterIDs, err := r.client.GetIPInterfaceFilters(ctx, iface, direction)
		if err != nil {
			return fmt.Errorf("failed to get filters for interface %s %s: %w", iface, direction, err)
		}

		filterIDValues := make([]attr.Value, len(filterIDs))
		for i, id := range filterIDs {
			filterIDValues[i] = types.Int64Value(int64(id))
		}

		updatedApply := ApplyModel{
			Interface: types.StringValue(iface),
			Direction: types.StringValue(direction),
			FilterIDs: types.ListValueMust(types.Int64Type, filterIDValues),
		}
		updatedApplies = append(updatedApplies, updatedApply)
	}

	data.SetAppliesFromRouter(updatedApplies)
	return nil
}

// extractFilterIDs extracts filter IDs from apply config, falling back to entry sequences.
func (r *AccessListIPResource) extractFilterIDs(apply ApplyModel, data *AccessListIPModel) []int {
	if !apply.FilterIDs.IsNull() && !apply.FilterIDs.IsUnknown() {
		var filterIDs []int64
		apply.FilterIDs.ElementsAs(context.TODO(), &filterIDs, false)
		if len(filterIDs) > 0 {
			ids := make([]int, len(filterIDs))
			for i, id := range filterIDs {
				ids[i] = int(id)
			}
			return ids
		}
	}

	// Fall back to all entry sequences
	return data.GetExpectedSequences()
}

// findRemovedSequences finds sequences that were in old but not in new.
func findRemovedSequences(old, new []int) []int {
	newSet := make(map[int]bool)
	for _, seq := range new {
		newSet[seq] = true
	}

	var removed []int
	for _, seq := range old {
		if !newSet[seq] {
			removed = append(removed, seq)
		}
	}

	return removed
}

// checkSequenceConflicts checks for sequence conflicts with existing filters on the router.
// currentState contains sequences that this resource already owns (for update operations).
func (r *AccessListIPResource) checkSequenceConflicts(ctx context.Context, data *AccessListIPModel, currentState []int, diagnostics *diag.Diagnostics) {
	logger := logging.FromContext(ctx)

	// Get planned sequences
	plannedSequences := data.GetExpectedSequences()
	if len(plannedSequences) == 0 {
		return
	}

	// Get all existing sequences from the router
	existingSequences, err := r.client.GetAllIPFilterSequences(ctx)
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
			fwhelpers.FormatSequenceConflictError("rtx_access_list_ip", fwhelpers.GetStringValue(data.Name), conflicts),
		)
	}
}
