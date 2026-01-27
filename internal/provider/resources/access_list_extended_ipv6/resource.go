package access_list_extended_ipv6

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
	_ resource.Resource                   = &AccessListExtendedIPv6Resource{}
	_ resource.ResourceWithImportState    = &AccessListExtendedIPv6Resource{}
	_ resource.ResourceWithValidateConfig = &AccessListExtendedIPv6Resource{}
)

// NewAccessListExtendedIPv6Resource creates a new access list extended IPv6 resource.
func NewAccessListExtendedIPv6Resource() resource.Resource {
	return &AccessListExtendedIPv6Resource{}
}

// AccessListExtendedIPv6Resource defines the resource implementation.
type AccessListExtendedIPv6Resource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AccessListExtendedIPv6Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_extended_ipv6"
}

// Schema defines the schema for the resource.
func (r *AccessListExtendedIPv6Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IPv6 extended access lists (ACLs) on RTX routers. Extended ACLs provide granular control over IPv6 packet filtering based on source/destination addresses, protocols, and ports.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the access list (used as identifier)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entry": schema.ListNestedBlock{
				Description: "List of ACL entries",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sequence": schema.Int64Attribute{
							Description: "Sequence number (determines order, typically 10, 20, 30...)",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"ace_rule_action": schema.StringAttribute{
							Description: "Action: 'permit' or 'deny'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("permit", "deny"),
							},
						},
						"ace_rule_protocol": schema.StringAttribute{
							Description: "Protocol: tcp, udp, icmpv6, ipv6, ip, or *",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("tcp", "udp", "icmpv6", "ipv6", "ip", "*"),
							},
						},
						"source_any": schema.BoolAttribute{
							Description: "Match any source address",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"source_prefix": schema.StringAttribute{
							Description: "Source IPv6 address (e.g., '2001:db8::')",
							Optional:    true,
						},
						"source_prefix_length": schema.Int64Attribute{
							Description: "Source prefix length (e.g., 64)",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(0, 128),
							},
						},
						"source_port_equal": schema.StringAttribute{
							Description: "Source port equals (e.g., '80', '443')",
							Optional:    true,
						},
						"source_port_range": schema.StringAttribute{
							Description: "Source port range (e.g., '1024-65535')",
							Optional:    true,
						},
						"destination_any": schema.BoolAttribute{
							Description: "Match any destination address",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"destination_prefix": schema.StringAttribute{
							Description: "Destination IPv6 address (e.g., '2001:db8:1::')",
							Optional:    true,
						},
						"destination_prefix_length": schema.Int64Attribute{
							Description: "Destination prefix length (e.g., 64)",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(0, 128),
							},
						},
						"destination_port_equal": schema.StringAttribute{
							Description: "Destination port equals (e.g., '80', '443')",
							Optional:    true,
						},
						"destination_port_range": schema.StringAttribute{
							Description: "Destination port range (e.g., '1024-65535')",
							Optional:    true,
						},
						"established": schema.BoolAttribute{
							Description: "Match established TCP connections (ACK or RST flag set)",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"log": schema.BoolAttribute{
							Description: "Enable logging for this entry",
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

// ValidateConfig validates the configuration.
func (r *AccessListExtendedIPv6Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data AccessListExtendedIPv6Model

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate entries
	if data.Entries.IsNull() || data.Entries.IsUnknown() {
		return
	}

	var entries []EntryModel
	resp.Diagnostics.Append(data.Entries.ElementsAs(ctx, &entries, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, entry := range entries {
		// Either source_any or source_prefix must be specified
		sourceAny := fwhelpers.GetBoolValue(entry.SourceAny)
		sourcePrefix := fwhelpers.GetStringValue(entry.SourcePrefix)
		if !sourceAny && sourcePrefix == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("entry").AtListIndex(i),
				"Invalid Source Configuration",
				"Either source_any must be true or source_prefix must be specified",
			)
		}

		// Either destination_any or destination_prefix must be specified
		destAny := fwhelpers.GetBoolValue(entry.DestinationAny)
		destPrefix := fwhelpers.GetStringValue(entry.DestinationPrefix)
		if !destAny && destPrefix == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("entry").AtListIndex(i),
				"Invalid Destination Configuration",
				"Either destination_any must be true or destination_prefix must be specified",
			)
		}

		// Established is only valid for TCP
		established := fwhelpers.GetBoolValue(entry.Established)
		protocol := strings.ToLower(fwhelpers.GetStringValue(entry.AceRuleProtocol))
		if established && protocol != "tcp" {
			resp.Diagnostics.AddAttributeError(
				path.Root("entry").AtListIndex(i).AtName("established"),
				"Invalid Established Configuration",
				"established can only be set to true for tcp protocol",
			)
		}
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListExtendedIPv6Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListExtendedIPv6Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListExtendedIPv6Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", name)
	logger := logging.FromContext(ctx)

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Creating IPv6 access list extended: %+v", acl)

	if err := r.client.CreateAccessListExtendedIPv6(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create IPv6 access list extended",
			fmt.Sprintf("Could not create IPv6 access list extended: %v", err),
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
func (r *AccessListExtendedIPv6Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListExtendedIPv6Model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ACL was not found, remove from state
	if data.Name.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the ACL from the router.
func (r *AccessListExtendedIPv6Resource) read(ctx context.Context, data *AccessListExtendedIPv6Model, diagnostics *diag.Diagnostics) {
	name := fwhelpers.GetStringValue(data.Name)

	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Reading IPv6 access list extended: %s", name)

	acl, err := r.client.GetAccessListExtendedIPv6(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("IPv6 access list extended %s not found", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read IPv6 access list extended", fmt.Sprintf("Could not read IPv6 access list extended %s: %v", name, err))
		return
	}

	data.FromClient(acl)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListExtendedIPv6Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListExtendedIPv6Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", name)
	logger := logging.FromContext(ctx)

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Updating IPv6 access list extended: %+v", acl)

	if err := r.client.UpdateAccessListExtendedIPv6(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IPv6 access list extended",
			fmt.Sprintf("Could not update IPv6 access list extended: %v", err),
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
func (r *AccessListExtendedIPv6Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListExtendedIPv6Model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwhelpers.GetStringValue(data.Name)
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Deleting IPv6 access list extended: %s", name)

	if err := r.client.DeleteAccessListExtendedIPv6(ctx, name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete IPv6 access list extended",
			fmt.Sprintf("Could not delete IPv6 access list extended %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListExtendedIPv6Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
