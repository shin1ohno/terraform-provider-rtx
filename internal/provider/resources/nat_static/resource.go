package nat_static

import (
	"context"
	"fmt"
	"net"
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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &NATStaticResource{}
	_ resource.ResourceWithImportState = &NATStaticResource{}
)

// NewNATStaticResource creates a new NAT static resource.
func NewNATStaticResource() resource.Resource {
	return &NATStaticResource{}
}

// NATStaticResource defines the resource implementation.
type NATStaticResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *NATStaticResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_static"
}

// Schema defines the schema for the resource.
func (r *NATStaticResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages static NAT (Network Address Translation) on RTX routers. Static NAT provides one-to-one mapping between inside local and outside global addresses.",
		Attributes: map[string]schema.Attribute{
			"descriptor_id": schema.Int64Attribute{
				Description: "The NAT descriptor ID (1-65535)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entry": schema.ListNestedBlock{
				Description: "List of static NAT mapping entries",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"inside_local": schema.StringAttribute{
							Description: "Inside local IP address (internal address)",
							Required:    true,
							Validators: []validator.String{
								ipv4AddressValidator{},
							},
						},
						"inside_local_port": schema.Int64Attribute{
							Description: "Inside local port (1-65535, required if protocol is specified)",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"outside_global": schema.StringAttribute{
							Description: "Outside global IP address (external address)",
							Required:    true,
							Validators: []validator.String{
								ipv4AddressValidator{},
							},
						},
						"outside_global_port": schema.Int64Attribute{
							Description: "Outside global port (1-65535, required if protocol is specified)",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol for port-based NAT: 'tcp' or 'udp' (required if ports are specified)",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp"),
							},
						},
					},
				},
			},
		},
	}
}

// ipv4AddressValidator validates that a string is a valid IPv4 address.
type ipv4AddressValidator struct{}

func (v ipv4AddressValidator) Description(ctx context.Context) string {
	return "value must be a valid IPv4 address"
}

func (v ipv4AddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipv4AddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IP Address",
			"IP address cannot be empty",
		)
		return
	}

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IP Address",
			fmt.Sprintf("Value %q is not a valid IP address", value),
		)
		return
	}

	if ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Value %q must be a valid IPv4 address", value),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *NATStaticResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NATStaticResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NATStaticModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate entry configuration
	r.validateEntries(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptorID := int(data.DescriptorID.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_nat_static", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	nat := data.ToClient()
	logger.Debug().Str("resource", "rtx_nat_static").Msgf("Creating NAT static: %+v", nat)

	if err := r.client.CreateNATStatic(ctx, nat); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create NAT static",
			fmt.Sprintf("Could not create NAT static: %v", err),
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
func (r *NATStaticResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NATStaticModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was removed
	if data.DescriptorID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the NAT static from the router.
func (r *NATStaticResource) read(ctx context.Context, data *NATStaticModel, diagnostics *diag.Diagnostics) {
	descriptorID := fwhelpers.GetInt64Value(data.DescriptorID)

	ctx = logging.WithResource(ctx, "rtx_nat_static", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_nat_static").Msgf("Reading NAT static: %d", descriptorID)

	var nat *client.NATStatic

	// Try to use SFTP cache if enabled
	if r.client.SFTPEnabled() {
		parsedConfig, err := r.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			nats := parsedConfig.ExtractNATStatic()
			for i := range nats {
				if nats[i].DescriptorID == descriptorID {
					nat = convertParsedNATStatic(&nats[i])
					logger.Debug().Str("resource", "rtx_nat_static").Msg("Found NAT static in SFTP cache")
					break
				}
			}
		}
		if nat == nil {
			logger.Debug().Str("resource", "rtx_nat_static").Msg("NAT static not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or NAT static not found in cache
	if nat == nil {
		var err error
		nat, err = r.client.GetNATStatic(ctx, descriptorID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_nat_static").Msgf("NAT static %d not found, removing from state", descriptorID)
				data.DescriptorID = types.Int64Null()
				return
			}
			fwhelpers.AppendDiagError(diagnostics, "Failed to read NAT static", fmt.Sprintf("Could not read NAT static %d: %v", descriptorID, err))
			return
		}
	}

	data.FromClient(nat)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NATStaticResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NATStaticModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate entry configuration
	r.validateEntries(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptorID := int(data.DescriptorID.ValueInt64())
	ctx = logging.WithResource(ctx, "rtx_nat_static", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	nat := data.ToClient()
	logger.Debug().Str("resource", "rtx_nat_static").Msgf("Updating NAT static: %+v", nat)

	if err := r.client.UpdateNATStatic(ctx, nat); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update NAT static",
			fmt.Sprintf("Could not update NAT static: %v", err),
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
func (r *NATStaticResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NATStaticModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptorID := fwhelpers.GetInt64Value(data.DescriptorID)

	ctx = logging.WithResource(ctx, "rtx_nat_static", strconv.Itoa(descriptorID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_nat_static").Msgf("Deleting NAT static: %d", descriptorID)

	if err := r.client.DeleteNATStatic(ctx, descriptorID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete NAT static",
			fmt.Sprintf("Could not delete NAT static %d: %v", descriptorID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *NATStaticResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	descriptorID, err := strconv.Atoi(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Invalid import ID format, expected descriptor_id (integer), got %q: %v", importID, err),
		)
		return
	}

	if descriptorID < 1 || descriptorID > 65535 {
		resp.Diagnostics.AddError(
			"Invalid descriptor_id",
			fmt.Sprintf("descriptor_id must be between 1 and 65535, got %d", descriptorID),
		)
		return
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_static").Msgf("Importing NAT static: %d", descriptorID)

	// Verify NAT static exists
	nat, err := r.client.GetNATStatic(ctx, descriptorID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import NAT static",
			fmt.Sprintf("Could not import NAT static %d: %v", descriptorID, err),
		)
		return
	}

	var data NATStaticModel
	data.FromClient(nat)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("descriptor_id"), types.Int64Value(int64(descriptorID)))...)
}

// validateEntries validates that entries have consistent port/protocol configuration.
func (r *NATStaticResource) validateEntries(ctx context.Context, data *NATStaticModel, diagnostics *diag.Diagnostics) {
	if data.Entry.IsNull() || data.Entry.IsUnknown() {
		return
	}

	elements := data.Entry.Elements()
	for i, elem := range elements {
		objVal, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attrs := objVal.Attributes()

		var insideLocalPort, outsideGlobalPort int64
		var protocol string

		if v, ok := attrs["inside_local_port"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
			insideLocalPort = v.ValueInt64()
		}
		if v, ok := attrs["outside_global_port"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
			outsideGlobalPort = v.ValueInt64()
		}
		if v, ok := attrs["protocol"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
			protocol = v.ValueString()
		}

		// If any port is specified, protocol must be specified
		if (insideLocalPort > 0 || outsideGlobalPort > 0) && protocol == "" {
			diagnostics.AddError(
				"Invalid entry configuration",
				fmt.Sprintf("entry[%d]: protocol is required when ports are specified", i),
			)
		}

		// If protocol is specified, both ports should be specified
		if protocol != "" && (insideLocalPort == 0 || outsideGlobalPort == 0) {
			diagnostics.AddError(
				"Invalid entry configuration",
				fmt.Sprintf("entry[%d]: both inside_local_port and outside_global_port are required when protocol is specified", i),
			)
		}
	}
}

// convertParsedNATStatic converts a parser NATStatic to a client NATStatic.
func convertParsedNATStatic(parsed *parsers.NATStatic) *client.NATStatic {
	nat := &client.NATStatic{
		DescriptorID: parsed.DescriptorID,
		Entries:      make([]client.NATStaticEntry, len(parsed.Entries)),
	}
	for i, entry := range parsed.Entries {
		nat.Entries[i] = client.NATStaticEntry{
			InsideLocal:   entry.InsideLocal,
			OutsideGlobal: entry.OutsideGlobal,
			Protocol:      entry.Protocol,
		}
		if entry.InsideLocalPort > 0 {
			port := entry.InsideLocalPort
			nat.Entries[i].InsideLocalPort = &port
		}
		if entry.OutsideGlobalPort > 0 {
			port := entry.OutsideGlobalPort
			nat.Entries[i].OutsideGlobalPort = &port
		}
	}
	return nat
}
