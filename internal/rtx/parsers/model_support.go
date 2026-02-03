// Package parsers provides RTX command parsing and building utilities.
package parsers

// SupportedModels defines the standard set of router models supported by this provider.
// Based on Yamaha RTX router command references.
var SupportedModels = []string{
	"vRX",
	"RTX5000",
	"RTX3510",
	"RTX3500",
	"RTX1300",
	"RTX1220",
	"RTX1210",
	"RTX840",
	"RTX830",
}

// modelSupportMap defines which router models support which commands.
// This is derived from Yamaha RTX router command references and spec files.
var modelSupportMap = map[string][]string{
	// Admin commands
	"admin_config": SupportedModels,

	// BGP configuration
	"bgp_config": SupportedModels,

	// Bridge configuration
	"bridge_config": SupportedModels,

	// DDNS configuration
	"ddns_config": SupportedModels,

	// DHCP configuration
	"dhcp_scope":                SupportedModels,
	"dhcp_scope_bind":           SupportedModels,
	"dhcp_service":              SupportedModels,
	"dhcp_relay_server":         SupportedModels,
	"ip_interface_dhcp_service": SupportedModels,
	"dhcp_scope_lease_type":     SupportedModels,
	"dhcp_client_hostname":      SupportedModels,

	// DNS configuration
	"dns_config": SupportedModels,

	// Ethernet filter
	"ethernet_filter": SupportedModels,

	// Interface configuration
	"interface_config": SupportedModels,

	// IP filter
	"ip_filter": SupportedModels,

	// IP route
	"ip_route": SupportedModels,

	// IPsec
	"ipsec_tunnel": SupportedModels,

	// IPv6 configuration
	"ipv6_config": SupportedModels,

	// L2TP configuration
	"l2tp_config": SupportedModels,

	// NAT configuration
	"nat_masquerade": SupportedModels,
	"nat_static":     SupportedModels,

	// OSPF configuration
	"ospf_config": SupportedModels,

	// PPP/PPPoE configuration
	"ppp_config": SupportedModels,

	// PPTP configuration
	"pptp_config": SupportedModels,

	// QoS configuration
	"qos_queue": SupportedModels,

	// Schedule configuration
	"schedule": SupportedModels,

	// Service configuration (HTTP/SSH/SFTP)
	"service_config": SupportedModels,

	// SNMP configuration
	"snmp_config": SupportedModels,

	// Syslog configuration
	"syslog_config": SupportedModels,

	// System configuration
	"system_config": SupportedModels,

	// Tunnel configuration
	"tunnel_config": SupportedModels,

	// VLAN configuration
	"vlan_config": SupportedModels,
}

// AllKnownModels returns all known RTX router models including older/unsupported ones
func AllKnownModels() []string {
	return []string{
		"vRX",
		"RTX5000",
		"RTX3510",
		"RTX3500",
		"RTX1300",
		"RTX1220",
		"RTX1210",
		"RTX840",
		"RTX830",
		"RTX810",
		"NVR700W",
		"NVR510",
		"NVR500",
	}
}

// IsModelSupported checks if a specific command is supported on a given router model.
// Returns true if the command is supported on the model, false otherwise.
func IsModelSupported(command, model string) bool {
	supportedModels, exists := modelSupportMap[command]
	if !exists {
		// If command is not in the map, assume it's supported on all standard models
		for _, m := range SupportedModels {
			if m == model {
				return true
			}
		}
		return false
	}

	for _, m := range supportedModels {
		if m == model {
			return true
		}
	}
	return false
}

// GetSupportedModels returns the list of models that support a given command.
// Returns nil if the command is not found.
func GetSupportedModels(command string) []string {
	return modelSupportMap[command]
}

// GetUnsupportedModels returns the list of models that do NOT support a given command.
func GetUnsupportedModels(command string) []string {
	supportedModels := modelSupportMap[command]
	if supportedModels == nil {
		supportedModels = SupportedModels
	}

	supportedMap := make(map[string]bool)
	for _, m := range supportedModels {
		supportedMap[m] = true
	}

	var unsupported []string
	for _, m := range AllKnownModels() {
		if !supportedMap[m] {
			unsupported = append(unsupported, m)
		}
	}
	return unsupported
}
