package dhcp_binding

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	_ resource.Resource                = &DHCPBindingResource{}
	_ resource.ResourceWithImportState = &DHCPBindingResource{}
)

// NewDHCPBindingResource creates a new DHCP binding resource.
func NewDHCPBindingResource() resource.Resource {
	return &DHCPBindingResource{}
}

// DHCPBindingResource defines the resource implementation.
type DHCPBindingResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *DHCPBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_binding"
}

// Schema defines the schema for the resource.
func (r *DHCPBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DHCP static lease bindings on RTX routers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the resource in the format 'scope_id:mac_address'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scope_id": schema.Int64Attribute{
				Description: "The DHCP scope ID.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "The IP address to assign.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac_address": schema.StringAttribute{
				Description: "The MAC address of the device (e.g., '00:11:22:33:44:55'). Conflicts with client_identifier.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("client_identifier")),
				},
			},
			"use_mac_as_client_id": schema.BoolAttribute{
				Description: "When true with mac_address, automatically generates '01:MAC' client identifier.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					requiresReplaceOnChange{},
				},
			},
			"client_identifier": schema.StringAttribute{
				Description: "DHCP Client Identifier in hex format (e.g., '01:aa:bb:cc:dd:ee:ff' for MAC-based, '02:12:34:56:78' for custom). Conflicts with mac_address.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("mac_address")),
					stringvalidator.ConflictsWith(path.MatchRoot("use_mac_as_client_id")),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname for the device (for documentation purposes).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the DHCP binding (for documentation purposes).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// requiresReplaceOnChange is a custom plan modifier for Bool attributes that requires replacement on change.
type requiresReplaceOnChange struct{}

func (m requiresReplaceOnChange) Description(ctx context.Context) string {
	return "Requires resource replacement when the value changes."
}

func (m requiresReplaceOnChange) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m requiresReplaceOnChange) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// If there's no state, this is a create operation
	if req.State.Raw.IsNull() {
		return
	}

	// If the config value is unknown, we can't compare
	if req.ConfigValue.IsUnknown() {
		return
	}

	// If the state value is unknown, we can't compare
	if req.StateValue.IsUnknown() {
		return
	}

	// Compare values - trigger replacement if different
	if req.ConfigValue.ValueBool() != req.StateValue.ValueBool() {
		resp.RequiresReplace = true
	}
}

// Configure adds the provider configured client to the resource.
func (r *DHCPBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig validates the resource configuration.
func (r *DHCPBindingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data DHCPBindingModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one of mac_address or client_identifier is specified
	// Skip validation if values are unknown (e.g., from for_each expressions)
	macAddressSet := !data.MACAddress.IsNull() && !data.MACAddress.IsUnknown() && data.MACAddress.ValueString() != ""
	clientIdentifierSet := !data.ClientIdentifier.IsNull() && !data.ClientIdentifier.IsUnknown() && data.ClientIdentifier.ValueString() != ""
	macAddressUnknown := data.MACAddress.IsUnknown()
	clientIdentifierUnknown := data.ClientIdentifier.IsUnknown()

	// Only validate if both are known (or null)
	if !macAddressUnknown && !clientIdentifierUnknown && !macAddressSet && !clientIdentifierSet {
		resp.Diagnostics.AddError(
			"Missing Client Identification",
			"Exactly one of 'mac_address' or 'client_identifier' must be specified.",
		)
		return
	}

	// Validate client_identifier format if specified
	if clientIdentifierSet {
		cid := data.ClientIdentifier.ValueString()
		normalized := normalizeClientIdentifier(cid)
		parts := strings.Split(normalized, ":")

		if len(parts) < 2 {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_identifier"),
				"Invalid Client Identifier Format",
				"Client identifier must be in format 'type:data' (e.g., '01:aa:bb:cc:dd:ee:ff', '02:66:6f:6f').",
			)
			return
		}

		// Validate each part is valid hex
		for i, part := range parts {
			if len(part) != 2 {
				resp.Diagnostics.AddAttributeError(
					path.Root("client_identifier"),
					"Invalid Client Identifier Format",
					fmt.Sprintf("Client identifier must contain 2-character hex octets at position %d, got %q.", i, part),
				)
				return
			}

			for _, c := range part {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					resp.Diagnostics.AddAttributeError(
						path.Root("client_identifier"),
						"Invalid Client Identifier Format",
						fmt.Sprintf("Client identifier contains invalid hex character '%c' at position %d.", c, i),
					)
					return
				}
			}
		}
	}

	// Validate MAC address format if specified
	if macAddressSet {
		mac := data.MACAddress.ValueString()
		if _, err := normalizeMACAddress(mac); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("mac_address"),
				"Invalid MAC Address Format",
				fmt.Sprintf("MAC address is invalid: %v.", err),
			)
			return
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *DHCPBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DHCPBindingModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource context for logging
	ctx = logging.WithResource(ctx, "rtx_dhcp_binding", "")
	logger := logging.FromContext(ctx)

	binding := data.ToClient()
	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Creating DHCP binding: scope_id=%d, ip=%s", binding.ScopeID, binding.IPAddress)

	if err := r.client.CreateDHCPBinding(ctx, binding); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create DHCP binding",
			fmt.Sprintf("Could not create DHCP binding: %v", err),
		)
		return
	}

	// Set the ID
	var identifier string
	if !data.MACAddress.IsNull() && data.MACAddress.ValueString() != "" {
		normalizedMAC, _ := normalizeMACAddress(data.MACAddress.ValueString())
		identifier = normalizedMAC
	} else if !data.ClientIdentifier.IsNull() && data.ClientIdentifier.ValueString() != "" {
		identifier = normalizeClientIdentifier(data.ClientIdentifier.ValueString())
	}
	data.ID = types.StringValue(fmt.Sprintf("%d:%s", binding.ScopeID, identifier))

	// Read back the created resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *DHCPBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DHCPBindingModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was deleted outside of Terraform
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the binding from the router.
func (r *DHCPBindingResource) read(ctx context.Context, data *DHCPBindingModel, diagnostics *diag.Diagnostics) {
	ctx = logging.WithResource(ctx, "rtx_dhcp_binding", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Reading DHCP binding: ID=%s", data.ID.ValueString())

	// Parse the ID or use the scope_id from state
	var scopeID int
	var identifier string

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		var err error
		scopeID, identifier, err = parseDHCPBindingID(data.ID.ValueString())
		if err != nil {
			fwhelpers.AppendDiagError(diagnostics, "Invalid resource ID", fmt.Sprintf("Could not parse resource ID: %v", err))
			return
		}
	} else {
		scopeID = fwhelpers.GetInt64Value(data.ScopeID)
		if !data.MACAddress.IsNull() && data.MACAddress.ValueString() != "" {
			identifier, _ = normalizeMACAddress(data.MACAddress.ValueString())
		}
	}

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Reading DHCP binding: scope_id=%d, identifier=%s", scopeID, identifier)

	// Get all bindings for the scope
	bindings, err := r.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		fwhelpers.AppendDiagError(diagnostics, "Failed to read DHCP binding", fmt.Sprintf("Could not retrieve DHCP bindings: %v", err))
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Retrieved %d bindings for scope %d", len(bindings), scopeID)

	// Find the matching binding
	var found *client.DHCPBinding
	for i := range bindings {
		binding := &bindings[i]
		normalizedBindingMAC, _ := normalizeMACAddress(binding.MACAddress)
		logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Checking binding: MAC=%s (normalized=%s), IP=%s against target=%s",
			binding.MACAddress, normalizedBindingMAC, binding.IPAddress, identifier)

		// Try to match by MAC address first
		if normalizedBindingMAC == identifier && binding.ScopeID == scopeID {
			found = binding
			break
		}
		// Fall back to IP address match for backward compatibility
		if binding.IPAddress == identifier && binding.ScopeID == scopeID {
			found = binding
			break
		}
	}

	if found == nil {
		logger.Debug().Str("resource", "rtx_dhcp_binding").Msg("Binding not found, marking as deleted")
		data.ID = types.StringNull()
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Found binding: %+v", found)

	// Preserve hostname and description from state (not stored on router)
	hostname := data.Hostname
	description := data.Description

	// Update data from the binding
	data.FromClient(found)

	// Restore hostname and description
	data.Hostname = hostname
	data.Description = description
}

// Update updates the resource and sets the updated Terraform state on success.
// Note: DHCP bindings are ForceNew for all fields, so Update is essentially a no-op.
func (r *DHCPBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DHCPBindingModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Since all fields are ForceNew, this should not be called in normal operation.
	// But we implement it for completeness.
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DHCPBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DHCPBindingModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_dhcp_binding", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	scopeID, identifier, err := parseDHCPBindingID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid resource ID",
			fmt.Sprintf("Could not parse resource ID: %v", err),
		)
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Deleting DHCP binding: scope_id=%d, identifier=%s", scopeID, identifier)

	// Get all bindings to find the IP address for this MAC address
	bindings, err := r.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete DHCP binding",
			fmt.Sprintf("Could not retrieve DHCP bindings: %v", err),
		)
		return
	}

	// Find the binding with matching MAC address to get its IP address
	var ipToDelete string
	for _, binding := range bindings {
		normalizedBindingMAC, _ := normalizeMACAddress(binding.MACAddress)
		if normalizedBindingMAC == identifier {
			ipToDelete = binding.IPAddress
			break
		}
		// Fall back to IP address match
		if binding.IPAddress == identifier {
			ipToDelete = binding.IPAddress
			break
		}
	}

	if ipToDelete == "" {
		// Binding already doesn't exist, consider this success
		logger.Debug().Str("resource", "rtx_dhcp_binding").Msg("Binding not found, already deleted")
		return
	}

	if err := r.client.DeleteDHCPBinding(ctx, scopeID, ipToDelete); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete DHCP binding",
			fmt.Sprintf("Could not delete DHCP binding: %v", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *DHCPBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	logger := logging.FromContext(ctx)
	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Importing DHCP binding: %s", importID)

	// Parse the import ID - can be either "scope_id:mac_address" or "scope_id:ip_address"
	scopeID, identifier, err := parseDHCPBindingID(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format 'scope_id:mac_address' or 'scope_id:ip_address', got %q: %v", importID, err),
		)
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Importing DHCP binding: scope_id=%d, identifier=%s", scopeID, identifier)

	// Get all bindings for the scope to find the requested binding
	bindings, err := r.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Import DHCP Binding",
			fmt.Sprintf("Could not retrieve DHCP bindings for scope %d: %v", scopeID, err),
		)
		return
	}

	// Determine if identifier is MAC address or IP address and find the binding
	var targetBinding *client.DHCPBinding

	// Check if identifier looks like a MAC address
	if _, err := normalizeMACAddress(identifier); err == nil {
		// It's a MAC address - search by MAC
		normalizedIdentifier, _ := normalizeMACAddress(identifier)
		for i := range bindings {
			binding := &bindings[i]
			normalizedBindingMAC, _ := normalizeMACAddress(binding.MACAddress)
			if normalizedBindingMAC == normalizedIdentifier {
				targetBinding = binding
				break
			}
		}
	} else {
		// It's likely an IP address - search by IP
		for i := range bindings {
			binding := &bindings[i]
			if binding.IPAddress == identifier {
				targetBinding = binding
				break
			}
		}
	}

	if targetBinding == nil {
		resp.Diagnostics.AddError(
			"DHCP Binding Not Found",
			fmt.Sprintf("DHCP binding with scope_id=%d and identifier=%s not found.", scopeID, identifier),
		)
		return
	}

	logger.Debug().Str("resource", "rtx_dhcp_binding").Msgf("Found binding for import: %+v", targetBinding)

	// Set the state values
	normalizedMAC, _ := normalizeMACAddress(targetBinding.MACAddress)
	finalID := fmt.Sprintf("%d:%s", scopeID, normalizedMAC)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), finalID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope_id"), int64(scopeID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_address"), targetBinding.IPAddress)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mac_address"), normalizedMAC)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("use_mac_as_client_id"), targetBinding.UseClientIdentifier)...)
}

// parseDHCPBindingID parses the composite ID into scope_id and identifier.
func parseDHCPBindingID(id string) (int, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("expected format 'scope_id:identifier', got %s", id)
	}

	scopeID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid scope_id: %v", err)
	}

	identifier := parts[1]
	return scopeID, identifier, nil
}
