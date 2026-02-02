package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// DHCPClientConfig represents DHCP client configuration
type DHCPClientConfig struct {
	Interface  string `json:"interface"`            // Interface name
	Hostname   string `json:"hostname,omitempty"`   // Client hostname
	ClientID   string `json:"client_id,omitempty"`  // Client identifier
	VendorID   string `json:"vendor_id,omitempty"`  // Vendor class identifier
	RequireDNS bool   `json:"require_dns"`          // Request DNS servers
	ReleaseOn  string `json:"release_on,omitempty"` // Release condition
}

// DHCPClientParser parses DHCP client configurations
type DHCPClientParser struct{}

// NewDHCPClientParser creates a new DHCP client parser
func NewDHCPClientParser() *DHCPClientParser {
	return &DHCPClientParser{}
}

// ParseClientConfig parses the output of "show config | grep dhcp client"
func (p *DHCPClientParser) ParseClientConfig(raw string) ([]DHCPClientConfig, error) {
	var configs []DHCPClientConfig
	configMap := make(map[string]*DHCPClientConfig)
	lines := strings.Split(raw, "\n")

	// Pattern: dhcp client hostname INTERFACE HOSTNAME
	hostnamePattern := regexp.MustCompile(`^\s*dhcp\s+client\s+hostname\s+(\S+)\s+(.+)\s*$`)
	// Pattern: dhcp client client-identifier INTERFACE TYPE:DATA
	clientIDPattern := regexp.MustCompile(`^\s*dhcp\s+client\s+client-identifier\s+(\S+)\s+(.+)\s*$`)
	// Pattern: dhcp client vendor-class-identifier INTERFACE VENDOR
	vendorIDPattern := regexp.MustCompile(`^\s*dhcp\s+client\s+vendor-class-identifier\s+(\S+)\s+(.+)\s*$`)
	// Pattern: dhcp client require-dns INTERFACE (on|off)
	requireDNSPattern := regexp.MustCompile(`^\s*dhcp\s+client\s+require-dns\s+(\S+)\s+(on|off)\s*$`)
	// Pattern: dhcp client release linkdown INTERFACE
	releaseLinkdownPattern := regexp.MustCompile(`^\s*dhcp\s+client\s+release\s+linkdown\s+(\S+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Helper to get or create config for interface
		getConfig := func(iface string) *DHCPClientConfig {
			if cfg, exists := configMap[iface]; exists {
				return cfg
			}
			cfg := &DHCPClientConfig{Interface: iface}
			configMap[iface] = cfg
			return cfg
		}

		if matches := hostnamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			cfg := getConfig(matches[1])
			cfg.Hostname = matches[2]
			continue
		}

		if matches := clientIDPattern.FindStringSubmatch(line); len(matches) >= 3 {
			cfg := getConfig(matches[1])
			cfg.ClientID = matches[2]
			continue
		}

		if matches := vendorIDPattern.FindStringSubmatch(line); len(matches) >= 3 {
			cfg := getConfig(matches[1])
			cfg.VendorID = matches[2]
			continue
		}

		if matches := requireDNSPattern.FindStringSubmatch(line); len(matches) >= 3 {
			cfg := getConfig(matches[1])
			cfg.RequireDNS = matches[2] == "on"
			continue
		}

		if matches := releaseLinkdownPattern.FindStringSubmatch(line); len(matches) >= 2 {
			cfg := getConfig(matches[1])
			cfg.ReleaseOn = "linkdown"
			continue
		}
	}

	// Convert map to slice
	for _, cfg := range configMap {
		configs = append(configs, *cfg)
	}

	return configs, nil
}

// BuildDHCPClientHostnameCommand builds the dhcp client hostname command
func BuildDHCPClientHostnameCommand(iface string, hostname string) string {
	return fmt.Sprintf("dhcp client hostname %s %s", iface, hostname)
}

// BuildDeleteDHCPClientHostnameCommand builds the no dhcp client hostname command
func BuildDeleteDHCPClientHostnameCommand(iface string) string {
	return fmt.Sprintf("no dhcp client hostname %s", iface)
}

// BuildDHCPClientClientIDCommand builds the dhcp client client-identifier command
func BuildDHCPClientClientIDCommand(iface string, clientID string) string {
	return fmt.Sprintf("dhcp client client-identifier %s %s", iface, clientID)
}

// BuildDeleteDHCPClientClientIDCommand builds the no dhcp client client-identifier command
func BuildDeleteDHCPClientClientIDCommand(iface string) string {
	return fmt.Sprintf("no dhcp client client-identifier %s", iface)
}

// BuildDHCPClientVendorIDCommand builds the dhcp client vendor-class-identifier command
func BuildDHCPClientVendorIDCommand(iface string, vendorID string) string {
	return fmt.Sprintf("dhcp client vendor-class-identifier %s %s", iface, vendorID)
}

// BuildDeleteDHCPClientVendorIDCommand builds the no dhcp client vendor-class-identifier command
func BuildDeleteDHCPClientVendorIDCommand(iface string) string {
	return fmt.Sprintf("no dhcp client vendor-class-identifier %s", iface)
}

// BuildDHCPClientRequireDNSCommand builds the dhcp client require-dns command
func BuildDHCPClientRequireDNSCommand(iface string, enabled bool) string {
	if enabled {
		return fmt.Sprintf("dhcp client require-dns %s on", iface)
	}
	return fmt.Sprintf("dhcp client require-dns %s off", iface)
}

// BuildDHCPClientReleaseLinkdownCommand builds the dhcp client release linkdown command
func BuildDHCPClientReleaseLinkdownCommand(iface string) string {
	return fmt.Sprintf("dhcp client release linkdown %s", iface)
}

// BuildDeleteDHCPClientReleaseLinkdownCommand builds the no dhcp client release linkdown command
func BuildDeleteDHCPClientReleaseLinkdownCommand(iface string) string {
	return fmt.Sprintf("no dhcp client release linkdown %s", iface)
}

// BuildShowDHCPClientCommand builds the show command for DHCP client
func BuildShowDHCPClientCommand(iface string) string {
	return fmt.Sprintf("show config | grep \"dhcp client.*%s\"", iface)
}
