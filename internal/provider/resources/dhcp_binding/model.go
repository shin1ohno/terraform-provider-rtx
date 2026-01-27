package dhcp_binding

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// DHCPBindingModel describes the resource data model.
type DHCPBindingModel struct {
	ID               types.String `tfsdk:"id"`
	ScopeID          types.Int64  `tfsdk:"scope_id"`
	IPAddress        types.String `tfsdk:"ip_address"`
	MACAddress       types.String `tfsdk:"mac_address"`
	UseMACAsClientID types.Bool   `tfsdk:"use_mac_as_client_id"`
	ClientIdentifier types.String `tfsdk:"client_identifier"`
	Hostname         types.String `tfsdk:"hostname"`
	Description      types.String `tfsdk:"description"`
}

// ToClient converts the Terraform model to a client.DHCPBinding.
func (m *DHCPBindingModel) ToClient() client.DHCPBinding {
	binding := client.DHCPBinding{
		ScopeID:   fwhelpers.GetInt64Value(m.ScopeID),
		IPAddress: fwhelpers.GetStringValue(m.IPAddress),
	}

	// Handle client identification method
	if !m.MACAddress.IsNull() && m.MACAddress.ValueString() != "" {
		binding.MACAddress = fwhelpers.GetStringValue(m.MACAddress)
		binding.UseClientIdentifier = fwhelpers.GetBoolValue(m.UseMACAsClientID)
	} else if !m.ClientIdentifier.IsNull() && m.ClientIdentifier.ValueString() != "" {
		binding.ClientIdentifier = fwhelpers.GetStringValue(m.ClientIdentifier)
		binding.UseClientIdentifier = true
	}

	return binding
}

// FromClient updates the Terraform model from a client.DHCPBinding.
func (m *DHCPBindingModel) FromClient(binding *client.DHCPBinding) {
	m.ScopeID = types.Int64Value(int64(binding.ScopeID))
	m.IPAddress = types.StringValue(binding.IPAddress)

	if binding.MACAddress != "" {
		normalizedMAC, _ := normalizeMACAddress(binding.MACAddress)
		m.MACAddress = types.StringValue(normalizedMAC)
	} else {
		m.MACAddress = types.StringNull()
	}

	m.UseMACAsClientID = types.BoolValue(binding.UseClientIdentifier)

	if binding.ClientIdentifier != "" {
		m.ClientIdentifier = types.StringValue(binding.ClientIdentifier)
	} else {
		m.ClientIdentifier = types.StringNull()
	}

	// Compute ID based on scope_id and MAC address
	normalizedMAC, _ := normalizeMACAddress(binding.MACAddress)
	m.ID = types.StringValue(fmt.Sprintf("%d:%s", binding.ScopeID, normalizedMAC))
}

// normalizeMACAddress normalizes MAC address to lowercase colon-separated format.
func normalizeMACAddress(mac string) (string, error) {
	// Remove all separators
	cleaned := strings.ToLower(mac)
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	// Validate length
	if len(cleaned) != 12 {
		return "", fmt.Errorf("MAC address must be 12 hex digits, got %d", len(cleaned))
	}

	// Validate characters
	for _, c := range cleaned {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", fmt.Errorf("MAC address contains invalid characters")
		}
	}

	// Format with colons
	result := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		cleaned[0:2], cleaned[2:4], cleaned[4:6],
		cleaned[6:8], cleaned[8:10], cleaned[10:12])

	return result, nil
}

// normalizeClientIdentifier normalizes client identifier format.
func normalizeClientIdentifier(cid string) string {
	if cid == "" {
		return ""
	}

	// Normalize: lowercase, consistent colon format
	cleaned := strings.ToLower(cid)
	cleaned = strings.ReplaceAll(cleaned, "-", ":")
	cleaned = strings.ReplaceAll(cleaned, " ", ":")

	// Remove duplicate colons
	for strings.Contains(cleaned, "::") {
		cleaned = strings.ReplaceAll(cleaned, "::", ":")
	}

	return cleaned
}

// normalizeIPAddress normalizes IP address format.
func normalizeIPAddress(ip string) string {
	return strings.TrimSpace(ip)
}
