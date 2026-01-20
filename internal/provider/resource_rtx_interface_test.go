package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXInterface_Schema(t *testing.T) {
	resource := resourceRTXInterface()

	// Test required fields
	if resource.Schema["name"].Required != true {
		t.Error("name should be required")
	}
	if resource.Schema["name"].ForceNew != true {
		t.Error("name should force new")
	}

	// Test optional fields
	optionalFields := []string{"description", "ip_address", "secure_filter_in", "secure_filter_out", "dynamic_filter_out", "nat_descriptor", "proxyarp", "mtu"}
	for _, field := range optionalFields {
		if resource.Schema[field].Optional != true {
			t.Errorf("%s should be optional", field)
		}
	}

	// Test computed fields (for import compatibility, these don't have Default values)
	if resource.Schema["nat_descriptor"].Computed != true {
		t.Error("nat_descriptor should be computed")
	}
	if resource.Schema["proxyarp"].Computed != true {
		t.Error("proxyarp should be computed")
	}
	if resource.Schema["mtu"].Computed != true {
		t.Error("mtu should be computed")
	}

	// Test ip_address block structure
	ipAddressResource := resource.Schema["ip_address"].Elem.(*schema.Resource)
	if ipAddressResource.Schema["address"].Optional != true {
		t.Error("ip_address.address should be optional")
	}
	if ipAddressResource.Schema["dhcp"].Optional != true {
		t.Error("ip_address.dhcp should be optional")
	}
	if ipAddressResource.Schema["dhcp"].Computed != true {
		t.Error("ip_address.dhcp should be computed (for import compatibility)")
	}
}

func TestValidateInterfaceConfigName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lan1", "lan1", false},
		{"valid lan2", "lan2", false},
		{"valid lan3", "lan3", false},
		{"valid bridge1", "bridge1", false},
		{"valid bridge10", "bridge10", false},
		{"valid pp1", "pp1", false},
		{"valid pp10", "pp10", false},
		{"valid tunnel1", "tunnel1", false},
		{"valid tunnel100", "tunnel100", false},
		{"invalid empty", "", true},
		{"invalid format", "invalid", true},
		{"invalid eth0", "eth0", true},
		{"invalid lan", "lan", true},
		{"invalid vlan1", "vlan1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := validateInterfaceConfigName(tt.input, "test")
			hasErr := len(errors) > 0

			if hasErr != tt.wantErr {
				if tt.wantErr {
					t.Errorf("validateInterfaceConfigName(%q) expected error, got none", tt.input)
				} else {
					t.Errorf("validateInterfaceConfigName(%q) unexpected error: %v", tt.input, errors)
				}
			}
		})
	}
}

func TestValidateCIDROptional(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid CIDR", "192.168.1.1/24", false},
		{"valid CIDR /16", "10.0.0.1/16", false},
		{"valid CIDR /32", "192.168.1.1/32", false},
		{"empty string", "", false},
		{"invalid no prefix", "192.168.1.1", true},
		{"invalid prefix", "192.168.1.1/33", true},
		{"invalid IP", "invalid/24", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := validateCIDROptional(tt.input, "test")
			hasErr := len(errors) > 0

			if hasErr != tt.wantErr {
				if tt.wantErr {
					t.Errorf("validateCIDROptional(%q) expected error, got none", tt.input)
				} else {
					t.Errorf("validateCIDROptional(%q) unexpected error: %v", tt.input, errors)
				}
			}
		})
	}
}

func TestBuildInterfaceConfigFromResourceData(t *testing.T) {
	// Create a mock ResourceData
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name":        "lan2",
		"description": "WAN connection",
		"ip_address": []interface{}{
			map[string]interface{}{
				"address": "",
				"dhcp":    true,
			},
		},
		"secure_filter_in":  []interface{}{200020, 200021, 200099},
		"secure_filter_out": []interface{}{200020, 200021, 200099},
		"nat_descriptor":    1000,
		"proxyarp":          false,
		"mtu":               0,
	})

	config := buildInterfaceConfigFromResourceData(d)

	if config.Name != "lan2" {
		t.Errorf("Name = %q, want %q", config.Name, "lan2")
	}
	if config.Description != "WAN connection" {
		t.Errorf("Description = %q, want %q", config.Description, "WAN connection")
	}
	if config.IPAddress == nil {
		t.Error("IPAddress should not be nil")
	} else {
		if config.IPAddress.DHCP != true {
			t.Error("IPAddress.DHCP should be true")
		}
	}
	if len(config.SecureFilterIn) != 3 {
		t.Errorf("SecureFilterIn length = %d, want 3", len(config.SecureFilterIn))
	}
	if config.NATDescriptor != 1000 {
		t.Errorf("NATDescriptor = %d, want 1000", config.NATDescriptor)
	}
}
