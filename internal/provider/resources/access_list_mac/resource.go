package access_list_mac

import (
	"context"
	"fmt"
	"strings"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// MaxSequence is the maximum allowed sequence number for RTX ACL entries.
const MaxSequence = 65535

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AccessListMACResource{}
	_ resource.ResourceWithImportState = &AccessListMACResource{}
)

// NewAccessListMACResource creates a new MAC access list resource.
func NewAccessListMACResource() resource.Resource {
	return &AccessListMACResource{}
}

// AccessListMACResource defines the resource implementation.
type AccessListMACResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListMACResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_mac"
}

// Schema defines the schema for the resource.
func (r *AccessListMACResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages MAC address access lists on RTX routers. " +
			"MAC ACLs filter traffic based on source and destination MAC addresses. " +
			"Supports automatic sequence numbering (auto mode) or manual sequence assignment.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Access list name (identifier)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filter_id": schema.Int64Attribute{
				Description: "Optional RTX filter ID to enable numeric ethernet filter mode. If not specified, derived from first entry.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"sequence_start": schema.Int64Attribute{
				Description: "Starting sequence number for automatic sequence calculation. When set, sequence numbers are automatically assigned to entries based on their definition order. Mutually exclusive with entry-level sequence attributes.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, MaxSequence),
				},
			},
			"sequence_step": schema.Int64Attribute{
				Description: fmt.Sprintf("Increment value for automatic sequence calculation. Only used when sequence_start is set. Default is %d.", DefaultSequenceStep),
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(DefaultSequenceStep),
				Validators: []validator.Int64{
					int64validator.Between(1, MaxSequence),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"apply": schema.ListNestedBlock{
				Description: "List of interface bindings. Each apply block binds this ACL to an interface in a specific direction. Multiple apply blocks are supported.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"interface": schema.StringAttribute{
							Description: "Interface to apply filters (e.g., lan1, bridge1). MAC ACLs cannot be applied to PP or Tunnel interfaces.",
							Required:    true,
						},
						"direction": schema.StringAttribute{
							Description: "Direction to apply filters (in or out)",
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
								listvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
							},
						},
					},
				},
			},
			"entry": schema.ListNestedBlock{
				Description: "List of MAC ACL entries",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sequence": schema.Int64Attribute{
							Description: "Sequence number (determines order of evaluation). Required in manual mode (when sequence_start is not set). Auto-calculated in auto mode (when sequence_start is set).",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, MaxSequence),
							},
						},
						"ace_action": schema.StringAttribute{
							Description: "Action to take (permit/deny or RTX pass/reject with log/nolog)",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive(
									"permit", "deny",
									"pass-log", "pass-nolog", "reject-log", "reject-nolog",
									"pass", "reject",
								),
							},
						},
						"source_any": schema.BoolAttribute{
							Description: "Match any source MAC address",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"source_address": schema.StringAttribute{
							Description: "Source MAC address (e.g., 00:00:00:00:00:00)",
							Optional:    true,
						},
						"source_address_mask": schema.StringAttribute{
							Description: "Source MAC wildcard mask",
							Optional:    true,
						},
						"destination_any": schema.BoolAttribute{
							Description: "Match any destination MAC address",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"destination_address": schema.StringAttribute{
							Description: "Destination MAC address (e.g., 00:00:00:00:00:00)",
							Optional:    true,
						},
						"destination_address_mask": schema.StringAttribute{
							Description: "Destination MAC wildcard mask",
							Optional:    true,
						},
						"ether_type": schema.StringAttribute{
							Description: "Ethernet type (e.g., 0x0800 for IPv4, 0x0806 for ARP)",
							Optional:    true,
						},
						"vlan_id": schema.Int64Attribute{
							Description: "VLAN ID to match",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 4094),
							},
						},
						"log": schema.BoolAttribute{
							Description: "Enable logging for this entry",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"filter_id": schema.Int64Attribute{
							Description: "Explicit filter number for this entry (overrides sequence)",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"offset": schema.Int64Attribute{
							Description: "Offset for byte matching",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
						"byte_list": schema.ListAttribute{
							Description: "Byte list (hex) for offset matching",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						"dhcp_match": schema.SingleNestedBlock{
							Description: "DHCP-based match settings",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "DHCP match type (dhcp-bind or dhcp-not-bind)",
									Optional:    true,
									Validators: []validator.String{
										stringvalidator.OneOf("dhcp-bind", "dhcp-not-bind"),
									},
								},
								"scope": schema.Int64Attribute{
									Description: "DHCP scope number",
									Optional:    true,
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListMACResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListMACResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListMACModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", name)
	logger := logging.FromContext(ctx)

	acl := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_access_list_mac").Msgf("Creating MAC access list: %s", acl.Name)

	if err := r.client.CreateAccessListMAC(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create MAC access list",
			fmt.Sprintf("Could not create MAC access list: %v", err),
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
func (r *AccessListMACResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListMACModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// Check if resource was removed (Name set to null)
		if data.Name.IsNull() {
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the MAC ACL from the router.
func (r *AccessListMACResource) read(ctx context.Context, data *AccessListMACModel, diagnostics *diag.Diagnostics) {
	name := fwhelpers.GetStringValue(data.Name)

	ctx = logging.WithResource(ctx, "rtx_access_list_mac", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_mac").Msgf("Reading MAC access list: %s", name)

	// Build a set of expected sequence numbers from current model
	expectedSequences := make(map[int]struct{})
	sequenceStart := fwhelpers.GetInt64Value(data.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(data.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	for i, entry := range data.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit sequence or filter_id
			seq = fwhelpers.GetInt64Value(entry.Sequence)
			if seq == 0 {
				seq = fwhelpers.GetInt64Value(entry.FilterID)
			}
		}
		if seq > 0 {
			expectedSequences[seq] = struct{}{}
		}
	}

	acl, err := r.client.GetAccessListMAC(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().Str("resource", "rtx_access_list_mac").Msgf("MAC access list %s not found, removing from state", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read MAC access list", fmt.Sprintf("Could not read MAC access list %s: %v", name, err))
		return
	}

	// Filter ACL entries to only include those with expected sequences
	if len(expectedSequences) > 0 {
		filteredEntries := make([]client.AccessListMACEntry, 0, len(data.Entries))
		for _, entry := range acl.Entries {
			seq := entry.Sequence
			if seq == 0 {
				seq = entry.FilterID
			}
			if _, ok := expectedSequences[seq]; ok {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		acl.Entries = filteredEntries
	}

	// Preserve Applies from the current model instead of reading from router
	// This is necessary because Applies reference filter IDs that may include
	// filters not managed by this resource
	acl.Applies = make([]client.MACApply, 0, len(data.Applies))
	for _, apply := range data.Applies {
		var filterIDs []int
		if !apply.FilterIDs.IsNull() && !apply.FilterIDs.IsUnknown() {
			var tfFilterIDs []types.Int64
			if diags := apply.FilterIDs.ElementsAs(ctx, &tfFilterIDs, false); !diags.HasError() {
				for _, id := range tfFilterIDs {
					filterIDs = append(filterIDs, int(id.ValueInt64()))
				}
			}
		}
		acl.Applies = append(acl.Applies, client.MACApply{
			Interface: fwhelpers.GetStringValue(apply.Interface),
			Direction: fwhelpers.GetStringValue(apply.Direction),
			FilterIDs: filterIDs,
		})
	}

	data.FromClient(ctx, acl, diagnostics)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListMACResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListMACModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", name)
	logger := logging.FromContext(ctx)

	acl := data.ToClient(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_access_list_mac").Msgf("Updating MAC access list: %s", acl.Name)

	if err := r.client.UpdateAccessListMAC(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update MAC access list",
			fmt.Sprintf("Could not update MAC access list: %v", err),
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
func (r *AccessListMACResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListMACModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_mac").Msgf("Deleting MAC access list: %s", name)

	// Get filter numbers to delete
	filterNums := data.GetFilterNumbersForDelete(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAccessListMAC(ctx, name, filterNums); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete MAC access list",
			fmt.Sprintf("Could not delete MAC access list %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListMACResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
