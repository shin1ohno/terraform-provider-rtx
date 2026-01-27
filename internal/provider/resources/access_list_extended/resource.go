package access_list_extended

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AccessListExtendedResource{}
	_ resource.ResourceWithImportState = &AccessListExtendedResource{}
)

// NewAccessListExtendedResource creates a new access list extended resource.
func NewAccessListExtendedResource() resource.Resource {
	return &AccessListExtendedResource{}
}

// AccessListExtendedResource defines the resource implementation.
type AccessListExtendedResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListExtendedResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_extended"
}

// Schema defines the schema for the resource.
func (r *AccessListExtendedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IPv4 extended access lists (ACLs) on RTX routers. Extended ACLs provide granular control over packet filtering based on source/destination addresses, protocols, and ports. " +
			"Supports both manual sequence mode (explicit sequence on each entry) and auto sequence mode (sequence_start + sequence_step). " +
			"Optional apply blocks bind the ACL to interfaces.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the access list (used as identifier).",
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
				Description: "List of ACL entries.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sequence": schema.Int64Attribute{
							Description: "Sequence number (determines order). Required in manual mode (when sequence_start is not set). Auto-calculated in auto mode.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, MaxSequenceValue),
							},
						},
						"ace_rule_action": schema.StringAttribute{
							Description: "Action: 'permit' or 'deny'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("permit", "deny"),
							},
						},
						"ace_rule_protocol": schema.StringAttribute{
							Description: "Protocol: tcp, udp, icmp, ip, gre, esp, ah, or *.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("tcp", "udp", "icmp", "ip", "gre", "esp", "ah", "*"),
							},
						},
						"source_any": schema.BoolAttribute{
							Description: "Match any source address.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"source_prefix": schema.StringAttribute{
							Description: "Source IP address (e.g., '192.168.1.0').",
							Optional:    true,
						},
						"source_prefix_mask": schema.StringAttribute{
							Description: "Source wildcard mask (e.g., '0.0.0.255').",
							Optional:    true,
						},
						"source_port_equal": schema.StringAttribute{
							Description: "Source port equals (e.g., '80', '443').",
							Optional:    true,
						},
						"source_port_range": schema.StringAttribute{
							Description: "Source port range (e.g., '1024-65535').",
							Optional:    true,
						},
						"destination_any": schema.BoolAttribute{
							Description: "Match any destination address.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"destination_prefix": schema.StringAttribute{
							Description: "Destination IP address (e.g., '10.0.0.0').",
							Optional:    true,
						},
						"destination_prefix_mask": schema.StringAttribute{
							Description: "Destination wildcard mask (e.g., '0.0.0.255').",
							Optional:    true,
						},
						"destination_port_equal": schema.StringAttribute{
							Description: "Destination port equals (e.g., '80', '443').",
							Optional:    true,
						},
						"destination_port_range": schema.StringAttribute{
							Description: "Destination port range (e.g., '1024-65535').",
							Optional:    true,
						},
						"established": schema.BoolAttribute{
							Description: "Match established TCP connections (ACK or RST flag set).",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"log": schema.BoolAttribute{
							Description: "Enable logging for this entry.",
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
func (r *AccessListExtendedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListExtendedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListExtendedModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate entries
	r.validateEntries(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_access_list_extended", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_extended").Msgf("Creating access list extended: %+v", acl)

	if err := r.client.CreateAccessListExtended(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create access list extended",
			fmt.Sprintf("Could not create access list extended: %v", err),
		)
		return
	}

	// Handle apply blocks
	if len(acl.Applies) > 0 {
		if err := r.applyFiltersToInterfaces(ctx, acl); err != nil {
			resp.Diagnostics.AddError(
				"Failed to apply filters to interfaces",
				fmt.Sprintf("Could not apply filters: %v", err),
			)
			return
		}
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccessListExtendedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListExtendedModel

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

// read is a helper function that reads the ACL from the router.
func (r *AccessListExtendedResource) read(ctx context.Context, data *AccessListExtendedModel, diagnostics *diag.Diagnostics) {
	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_access_list_extended", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_extended").Msgf("Reading access list extended: %s", name)

	acl, err := r.client.GetAccessListExtended(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_access_list_extended").Msgf("Access list extended %s not found", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read access list extended", fmt.Sprintf("Could not read access list extended %s: %v", name, err))
		return
	}

	data.FromClient(acl)

	// Read apply blocks from router
	applies, err := r.readApplies(ctx, acl)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to read interface filters, apply state may be stale")
	} else {
		data.SetAppliesFromClient(applies)
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListExtendedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListExtendedModel
	var state AccessListExtendedModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate entries
	r.validateEntries(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_access_list_extended", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_extended").Msgf("Updating access list extended: %+v", acl)

	// Update entries
	if err := r.client.UpdateAccessListExtended(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update access list extended",
			fmt.Sprintf("Could not update access list extended: %v", err),
		)
		return
	}

	// Handle apply block changes
	oldACL := state.ToClient()
	if err := r.updateApplies(ctx, acl, oldACL.Applies); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update interface filters",
			fmt.Sprintf("Could not update filters: %v", err),
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
func (r *AccessListExtendedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListExtendedModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_access_list_extended", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_extended").Msgf("Deleting access list extended: %s", name)

	// Remove applies first
	acl := data.ToClient()
	if len(acl.Applies) > 0 {
		if err := r.removeFiltersFromInterfaces(ctx, acl); err != nil {
			logger.Warn().Err(err).Msg("Failed to remove filters from interfaces before delete")
		}
	}

	if err := r.client.DeleteAccessListExtended(ctx, name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete access list extended",
			fmt.Sprintf("Could not delete access list extended %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListExtendedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	ctx = logging.WithResource(ctx, "rtx_access_list_extended", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_extended").Msgf("Importing access list extended: %s", name)

	acl, err := r.client.GetAccessListExtended(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import access list extended",
			fmt.Sprintf("Could not import access list extended %s: %v", name, err),
		)
		return
	}

	var data AccessListExtendedModel
	data.Name = types.StringValue(name)
	data.FromClient(acl)

	// Set default sequence values for imported resources
	data.SequenceStart = types.Int64Null()
	data.SequenceStep = types.Int64Value(DefaultSequenceStep)

	// Read apply blocks from router
	applies, err := r.readApplies(ctx, acl)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to read interface filters during import")
	} else {
		data.SetAppliesFromClient(applies)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// validateEntries validates the ACL entries.
func (r *AccessListExtendedResource) validateEntries(ctx context.Context, data *AccessListExtendedModel, diagnostics *diag.Diagnostics) {
	if data.Entry.IsNull() || data.Entry.IsUnknown() {
		return
	}

	var entries []EntryModel
	data.Entry.ElementsAs(ctx, &entries, false)

	for i, entry := range entries {
		sourceAny := fwhelpers.GetBoolValue(entry.SourceAny)
		sourcePrefix := fwhelpers.GetStringValue(entry.SourcePrefix)
		destAny := fwhelpers.GetBoolValue(entry.DestinationAny)
		destPrefix := fwhelpers.GetStringValue(entry.DestinationPrefix)
		protocol := fwhelpers.GetStringValue(entry.AceRuleProtocol)
		established := fwhelpers.GetBoolValue(entry.Established)

		// Either source_any or source_prefix must be specified
		if !sourceAny && sourcePrefix == "" {
			diagnostics.AddAttributeError(
				path.Root("entry").AtListIndex(i),
				"Invalid entry configuration",
				"Either source_any must be true or source_prefix must be specified.",
			)
		}

		// Either destination_any or destination_prefix must be specified
		if !destAny && destPrefix == "" {
			diagnostics.AddAttributeError(
				path.Root("entry").AtListIndex(i),
				"Invalid entry configuration",
				"Either destination_any must be true or destination_prefix must be specified.",
			)
		}

		// Established is only valid for TCP
		if established && strings.ToLower(protocol) != "tcp" {
			diagnostics.AddAttributeError(
				path.Root("entry").AtListIndex(i).AtName("established"),
				"Invalid entry configuration",
				"The established attribute can only be set to true for TCP protocol.",
			)
		}
	}
}

// applyFiltersToInterfaces applies filters to interfaces.
func (r *AccessListExtendedResource) applyFiltersToInterfaces(ctx context.Context, acl client.AccessListExtended) error {
	for _, apply := range acl.Applies {
		filterIDs := apply.FilterIDs
		// If no filter_ids specified, use all entry sequences
		if len(filterIDs) == 0 {
			for _, entry := range acl.Entries {
				filterIDs = append(filterIDs, entry.Sequence)
			}
		}

		err := r.client.ApplyIPFiltersToInterface(ctx, apply.Interface, apply.Direction, filterIDs)
		if err != nil {
			return fmt.Errorf("failed to apply filters to %s %s: %w", apply.Interface, apply.Direction, err)
		}
	}
	return nil
}

// removeFiltersFromInterfaces removes filters from interfaces.
func (r *AccessListExtendedResource) removeFiltersFromInterfaces(ctx context.Context, acl client.AccessListExtended) error {
	for _, apply := range acl.Applies {
		err := r.client.RemoveIPFiltersFromInterface(ctx, apply.Interface, apply.Direction)
		if err != nil {
			return fmt.Errorf("failed to remove filters from %s %s: %w", apply.Interface, apply.Direction, err)
		}
	}
	return nil
}

// updateApplies handles changes to apply blocks.
func (r *AccessListExtendedResource) updateApplies(ctx context.Context, acl client.AccessListExtended, oldApplies []client.ExtendedApply) error {
	// Build maps for comparison
	oldMap := make(map[string]client.ExtendedApply)
	for _, a := range oldApplies {
		key := fmt.Sprintf("%s:%s", a.Interface, a.Direction)
		oldMap[key] = a
	}

	newMap := make(map[string]client.ExtendedApply)
	for _, a := range acl.Applies {
		key := fmt.Sprintf("%s:%s", a.Interface, a.Direction)
		newMap[key] = a
	}

	// Remove old applies that are not in new
	for key, oldApply := range oldMap {
		if _, exists := newMap[key]; !exists {
			err := r.client.RemoveIPFiltersFromInterface(ctx, oldApply.Interface, oldApply.Direction)
			if err != nil {
				return fmt.Errorf("failed to remove filters from %s %s: %w", oldApply.Interface, oldApply.Direction, err)
			}
		}
	}

	// Add or update new applies
	for key, newApply := range newMap {
		filterIDs := newApply.FilterIDs
		// If no filter_ids specified, use all entry sequences
		if len(filterIDs) == 0 {
			for _, entry := range acl.Entries {
				filterIDs = append(filterIDs, entry.Sequence)
			}
		}

		oldApply, exists := oldMap[key]
		if !exists || !equalIntSlices(oldApply.FilterIDs, filterIDs) {
			// Apply or update
			err := r.client.ApplyIPFiltersToInterface(ctx, newApply.Interface, newApply.Direction, filterIDs)
			if err != nil {
				return fmt.Errorf("failed to apply filters to %s %s: %w", newApply.Interface, newApply.Direction, err)
			}
		}
	}

	return nil
}

// readApplies reads the current apply state from the router.
func (r *AccessListExtendedResource) readApplies(ctx context.Context, acl *client.AccessListExtended) ([]client.ExtendedApply, error) {
	// Build set of our entry sequences for filtering
	ourSequences := make(map[int]bool)
	for _, entry := range acl.Entries {
		ourSequences[entry.Sequence] = true
	}

	result := make([]client.ExtendedApply, 0)
	for _, apply := range acl.Applies {
		filterIDs := apply.FilterIDs
		if len(filterIDs) == 0 {
			// If no explicit filter_ids, use all sequences
			for _, entry := range acl.Entries {
				filterIDs = append(filterIDs, entry.Sequence)
			}
		}

		// Verify each filter is still applied by querying the router
		currentFilters, err := r.client.GetIPInterfaceFilters(ctx, apply.Interface, apply.Direction)
		if err != nil {
			// If we can't read filters, skip this apply
			continue
		}

		// Find matching filter IDs
		currentSet := make(map[int]bool)
		for _, id := range currentFilters {
			currentSet[id] = true
		}

		matchingIDs := make([]int, 0)
		for _, id := range filterIDs {
			if currentSet[id] && ourSequences[id] {
				matchingIDs = append(matchingIDs, id)
			}
		}

		if len(matchingIDs) > 0 {
			result = append(result, client.ExtendedApply{
				Interface: apply.Interface,
				Direction: apply.Direction,
				FilterIDs: matchingIDs,
			})
		}
	}

	return result, nil
}

// equalIntSlices compares two int slices for equality.
func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SetApplyValuesFromEntries populates filter_ids from entries when not explicitly set.
func SetApplyValuesFromEntries(applies []ApplyModel, entries []client.AccessListExtendedEntry) []ApplyModel {
	result := make([]ApplyModel, len(applies))
	for i, apply := range applies {
		result[i] = apply
		if apply.FilterIDs.IsNull() || apply.FilterIDs.IsUnknown() {
			// Populate with all entry sequences
			filterIDValues := make([]attr.Value, len(entries))
			for j, entry := range entries {
				filterIDValues[j] = types.Int64Value(int64(entry.Sequence))
			}
			result[i].FilterIDs = types.ListValueMust(types.Int64Type, filterIDValues)
		}
	}
	return result
}
