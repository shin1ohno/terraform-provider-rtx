package access_list_ip_dynamic

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
	_ resource.Resource                = &AccessListIPDynamicResource{}
	_ resource.ResourceWithImportState = &AccessListIPDynamicResource{}
)

// NewAccessListIPDynamicResource creates a new dynamic IP access list resource.
func NewAccessListIPDynamicResource() resource.Resource {
	return &AccessListIPDynamicResource{}
}

// AccessListIPDynamicResource defines the resource implementation.
type AccessListIPDynamicResource struct {
	client client.Client
}

// validProtocols is the list of valid protocols for dynamic IP filters.
var validProtocols = []string{
	"ftp", "www", "smtp", "pop3", "dns", "domain", "telnet", "ssh",
	"tcp", "udp", "*",
	"tftp", "submission", "https", "imap", "imaps", "pop3s", "smtps",
	"ldap", "ldaps", "bgp", "sip", "ipsec-nat-t", "ntp", "snmp",
	"rtsp", "h323", "pptp", "l2tp", "ike", "esp",
}

// Metadata returns the resource type name.
func (r *AccessListIPDynamicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_list_ip_dynamic"
}

// Schema defines the schema for the resource.
func (r *AccessListIPDynamicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manages a named collection of IPv4 dynamic (stateful) IP filters on RTX routers.

Dynamic filters provide stateful packet inspection for various protocols. This resource groups
multiple dynamic filter entries under a single name for easier management and reference.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Access list name (identifier).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entry": schema.ListNestedBlock{
				Description: "List of dynamic filter entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sequence": schema.Int64Attribute{
							Description: "Sequence number (determines order and filter number).",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"source": schema.StringAttribute{
							Description: "Source address or '*' for any. Can be an IP address, network in CIDR notation, or '*'.",
							Required:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination address or '*' for any. Can be an IP address, network in CIDR notation, or '*'.",
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
						"timeout": schema.Int64Attribute{
							Description: "Timeout value in seconds. If not specified, uses system default.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AccessListIPDynamicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AccessListIPDynamicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessListIPDynamicModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Creating dynamic IP access list: %s", acl.Name)

	if err := r.client.CreateAccessListIPDynamic(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create dynamic IP access list",
			fmt.Sprintf("Could not create dynamic IP access list: %v", err),
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
func (r *AccessListIPDynamicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessListIPDynamicModel

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

// read is a helper function that reads the access list from the router.
func (r *AccessListIPDynamicResource) read(ctx context.Context, data *AccessListIPDynamicModel, diagnostics *diag.Diagnostics) {
	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Reading dynamic IP access list: %s", name)

	// Get current sequences from state to filter results
	// This prevents other access lists' filters from leaking into this resource's state
	currentSeqs := data.GetCurrentSequences()

	acl, err := r.client.GetAccessListIPDynamic(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Dynamic IP access list %s not found, removing from state", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read dynamic IP access list", fmt.Sprintf("Could not read dynamic IP access list %s: %v", name, err))
		return
	}

	data.FromClient(acl, currentSeqs)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccessListIPDynamicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessListIPDynamicModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	acl := data.ToClient()
	logger.Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Updating dynamic IP access list: %s", acl.Name)

	if err := r.client.UpdateAccessListIPDynamic(ctx, acl); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update dynamic IP access list",
			fmt.Sprintf("Could not update dynamic IP access list: %v", err),
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
func (r *AccessListIPDynamicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessListIPDynamicModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Deleting dynamic IP access list: %s", name)

	// Collect filter numbers to delete
	filterNums := data.GetFilterNumbers()

	if err := r.client.DeleteAccessListIPDynamic(ctx, name, filterNums); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete dynamic IP access list",
			fmt.Sprintf("Could not delete dynamic IP access list %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AccessListIPDynamicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Importing dynamic IP access list: %s", name)

	// Import only sets the name - entries are intentionally NOT imported.
	// This is because RTX doesn't track which filters belong to which "named list".
	// The Terraform configuration defines which entries belong to this access list.
	// After import, run `terraform apply` to bind the configured entries to this resource.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}
