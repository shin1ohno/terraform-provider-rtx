package nat_masquerade

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	_ resource.Resource                = &NATMasqueradeResource{}
	_ resource.ResourceWithImportState = &NATMasqueradeResource{}
)

// NewNATMasqueradeResource creates a new NAT masquerade resource.
func NewNATMasqueradeResource() resource.Resource {
	return &NATMasqueradeResource{}
}

// NATMasqueradeResource defines the resource implementation.
type NATMasqueradeResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *NATMasqueradeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_masquerade"
}

// Schema defines the schema for the resource.
func (r *NATMasqueradeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages NAT masquerade (PAT/NAPT) configurations on RTX routers. NAT masquerade allows multiple internal hosts to share a single external IP address using port address translation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (same as descriptor_id).",
				Computed:    true,
			},
			"descriptor_id": schema.Int64Attribute{
				Description: "NAT descriptor ID (1-65535).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"outer_address": schema.StringAttribute{
				Description: "Outer (external) address: 'ipcp' for PPPoE-assigned address, interface name (e.g., 'pp1'), or specific IP address.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"inner_network": schema.StringAttribute{
				Description: "Inner (internal) network range in format 'start_ip-end_ip' (e.g., '192.168.1.0-192.168.1.255').",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"static_entry": schema.ListNestedBlock{
				Description: "Static port mapping entries for port forwarding.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"entry_number": schema.Int64Attribute{
							Description: "Entry number for identification.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"inside_local": schema.StringAttribute{
							Description: "Internal IP address.",
							Required:    true,
						},
						"inside_local_port": schema.Int64Attribute{
							Description: "Internal port number (1-65535). Required for tcp/udp, omit for protocol-only entries (esp, ah, gre, icmp).",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"outside_global": schema.StringAttribute{
							Description: "External IP address or 'ipcp' for PPPoE-assigned address.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("ipcp"),
						},
						"outside_global_port": schema.Int64Attribute{
							Description: "External port number (1-65535). Required for tcp/udp, omit for protocol-only entries (esp, ah, gre, icmp).",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol: 'tcp', 'udp' (require ports), or 'esp', 'ah', 'gre', 'icmp' (protocol-only, no ports).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("tcp", "udp", "esp", "ah", "gre", "icmp"),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NATMasqueradeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NATMasqueradeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NATMasqueradeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptorID := strconv.FormatInt(data.DescriptorID.ValueInt64(), 10)
	ctx = logging.WithResource(ctx, "rtx_nat_masquerade", descriptorID)
	logger := logging.FromContext(ctx)

	nat, diags := data.ToClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_nat_masquerade").Msgf("Creating NAT Masquerade: %+v", nat)

	if err := r.client.CreateNATMasquerade(ctx, nat); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create NAT masquerade",
			fmt.Sprintf("Could not create NAT masquerade: %v", err),
		)
		return
	}

	// Set the ID
	data.ID = types.StringValue(descriptorID)

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *NATMasqueradeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NATMasqueradeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was deleted
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the NAT masquerade from the router.
func (r *NATMasqueradeResource) read(ctx context.Context, data *NATMasqueradeModel, diagnostics *diag.Diagnostics) {
	descriptorID := fwhelpers.GetInt64Value(data.DescriptorID)
	if descriptorID == 0 {
		// Try to parse from ID
		id := fwhelpers.GetStringValue(data.ID)
		if id != "" {
			parsed, err := strconv.Atoi(id)
			if err == nil {
				descriptorID = parsed
			}
		}
	}

	ctx = logging.WithResource(ctx, "rtx_nat_masquerade", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_nat_masquerade").Msgf("Reading NAT Masquerade: %d", descriptorID)

	var nat *client.NATMasquerade
	var err error

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, cacheErr := r.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract NAT masquerade from parsed config
			nats := parsedConfig.ExtractNATMasquerade()
			for i := range nats {
				if nats[i].DescriptorID == descriptorID {
					nat = convertParsedNATMasquerade(&nats[i])
					logger.Debug().Str("resource", "rtx_nat_masquerade").Msg("Found NAT masquerade in SFTP cache")
					break
				}
			}
		}
		if nat == nil {
			logger.Debug().Str("resource", "rtx_nat_masquerade").Msg("NAT masquerade not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or NAT not found in cache
	if nat == nil {
		nat, err = r.client.GetNATMasquerade(ctx, descriptorID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_nat_masquerade").Msgf("NAT Masquerade %d not found, removing from state", descriptorID)
				data.ID = types.StringNull()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read NAT masquerade", fmt.Sprintf("Could not read NAT masquerade %d: %v", descriptorID, err))
			return
		}
	}

	diagnostics.Append(data.FromClient(ctx, nat)...)
	data.ID = types.StringValue(strconv.Itoa(nat.DescriptorID))
}

// convertParsedNATMasquerade converts a parser NATMasquerade to a client NATMasquerade.
func convertParsedNATMasquerade(parsed *parsers.NATMasquerade) *client.NATMasquerade {
	nat := &client.NATMasquerade{
		DescriptorID:  parsed.DescriptorID,
		OuterAddress:  parsed.OuterAddress,
		InnerNetwork:  parsed.InnerNetwork,
		StaticEntries: make([]client.MasqueradeStaticEntry, len(parsed.StaticEntries)),
	}
	for i, entry := range parsed.StaticEntries {
		nat.StaticEntries[i] = client.MasqueradeStaticEntry{
			EntryNumber:       entry.EntryNumber,
			InsideLocal:       entry.InsideLocal,
			InsideLocalPort:   entry.InsideLocalPort,
			OutsideGlobal:     entry.OutsideGlobal,
			OutsideGlobalPort: entry.OutsideGlobalPort,
			Protocol:          entry.Protocol,
		}
	}
	return nat
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NATMasqueradeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NATMasqueradeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptorID := strconv.FormatInt(data.DescriptorID.ValueInt64(), 10)
	ctx = logging.WithResource(ctx, "rtx_nat_masquerade", descriptorID)
	logger := logging.FromContext(ctx)

	nat, diags := data.ToClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logger.Debug().Str("resource", "rtx_nat_masquerade").Msgf("Updating NAT Masquerade: %+v", nat)

	if err := r.client.UpdateNATMasquerade(ctx, nat); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update NAT masquerade",
			fmt.Sprintf("Could not update NAT masquerade: %v", err),
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
func (r *NATMasqueradeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NATMasqueradeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptorID := fwhelpers.GetInt64Value(data.DescriptorID)
	ctx = logging.WithResource(ctx, "rtx_nat_masquerade", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_nat_masquerade").Msgf("Deleting NAT Masquerade: %d", descriptorID)

	if err := r.client.DeleteNATMasquerade(ctx, descriptorID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete NAT masquerade",
			fmt.Sprintf("Could not delete NAT masquerade %d: %v", descriptorID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *NATMasqueradeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Parse import ID as descriptor_id
	descriptorID, err := strconv.Atoi(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Invalid import ID format, expected descriptor_id (e.g., '1'): %v", err),
		)
		return
	}

	if descriptorID < 1 || descriptorID > 65535 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("descriptor_id must be between 1 and 65535, got %d", descriptorID),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_nat_masquerade", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_nat_masquerade").Msgf("Importing NAT Masquerade: %d", descriptorID)

	// Set both id and descriptor_id
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("descriptor_id"), int64(descriptorID))...)
}
