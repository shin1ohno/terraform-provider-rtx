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
	optionalFields := []string{
		"description", "ip_address", "nat_descriptor", "proxyarp", "mtu",
		"access_list_ip_in", "access_list_ip_out",
		"access_list_ipv6_in", "access_list_ipv6_out",
		"access_list_ip_dynamic_in", "access_list_ip_dynamic_out",
		"access_list_ipv6_dynamic_in", "access_list_ipv6_dynamic_out",
		"access_list_mac_in", "access_list_mac_out",
	}
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
		"access_list_ip_in":  "wan-in-acl",
		"access_list_ip_out": "wan-out-acl",
		"nat_descriptor":     1000,
		"proxyarp":           false,
		"mtu":                0,
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
	if config.AccessListIPIn != "wan-in-acl" {
		t.Errorf("AccessListIPIn = %q, want %q", config.AccessListIPIn, "wan-in-acl")
	}
	if config.AccessListIPOut != "wan-out-acl" {
		t.Errorf("AccessListIPOut = %q, want %q", config.AccessListIPOut, "wan-out-acl")
	}
	if config.NATDescriptor != 1000 {
		t.Errorf("NATDescriptor = %d, want 1000", config.NATDescriptor)
	}
}

func TestBuildInterfaceConfigFromResourceData_AccessLists(t *testing.T) {
	// Create a mock ResourceData with access list attributes
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name":                         "lan1",
		"access_list_ip_in":            "ip-in-acl",
		"access_list_ip_out":           "ip-out-acl",
		"access_list_ipv6_in":          "ipv6-in-acl",
		"access_list_ipv6_out":         "ipv6-out-acl",
		"access_list_ip_dynamic_in":    "ip-dyn-in-acl",
		"access_list_ip_dynamic_out":   "ip-dyn-out-acl",
		"access_list_ipv6_dynamic_in":  "ipv6-dyn-in-acl",
		"access_list_ipv6_dynamic_out": "ipv6-dyn-out-acl",
		"access_list_mac_in":           "mac-in-acl",
		"access_list_mac_out":          "mac-out-acl",
	})

	config := buildInterfaceConfigFromResourceData(d)

	// Verify access list attributes
	if config.AccessListIPIn != "ip-in-acl" {
		t.Errorf("AccessListIPIn = %q, want %q", config.AccessListIPIn, "ip-in-acl")
	}
	if config.AccessListIPOut != "ip-out-acl" {
		t.Errorf("AccessListIPOut = %q, want %q", config.AccessListIPOut, "ip-out-acl")
	}
	if config.AccessListIPv6In != "ipv6-in-acl" {
		t.Errorf("AccessListIPv6In = %q, want %q", config.AccessListIPv6In, "ipv6-in-acl")
	}
	if config.AccessListIPv6Out != "ipv6-out-acl" {
		t.Errorf("AccessListIPv6Out = %q, want %q", config.AccessListIPv6Out, "ipv6-out-acl")
	}
	if config.AccessListIPDynamicIn != "ip-dyn-in-acl" {
		t.Errorf("AccessListIPDynamicIn = %q, want %q", config.AccessListIPDynamicIn, "ip-dyn-in-acl")
	}
	if config.AccessListIPDynamicOut != "ip-dyn-out-acl" {
		t.Errorf("AccessListIPDynamicOut = %q, want %q", config.AccessListIPDynamicOut, "ip-dyn-out-acl")
	}
	if config.AccessListIPv6DynamicIn != "ipv6-dyn-in-acl" {
		t.Errorf("AccessListIPv6DynamicIn = %q, want %q", config.AccessListIPv6DynamicIn, "ipv6-dyn-in-acl")
	}
	if config.AccessListIPv6DynamicOut != "ipv6-dyn-out-acl" {
		t.Errorf("AccessListIPv6DynamicOut = %q, want %q", config.AccessListIPv6DynamicOut, "ipv6-dyn-out-acl")
	}
	if config.AccessListMACIn != "mac-in-acl" {
		t.Errorf("AccessListMACIn = %q, want %q", config.AccessListMACIn, "mac-in-acl")
	}
	if config.AccessListMACOut != "mac-out-acl" {
		t.Errorf("AccessListMACOut = %q, want %q", config.AccessListMACOut, "mac-out-acl")
	}
}

func TestBuildInterfaceConfigFromResourceData_AccessListsEmpty(t *testing.T) {
	// Create a mock ResourceData without access list attributes
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name": "lan1",
	})

	config := buildInterfaceConfigFromResourceData(d)

	// Verify access list attributes are empty when not set
	if config.AccessListIPIn != "" {
		t.Errorf("AccessListIPIn should be empty, got %q", config.AccessListIPIn)
	}
	if config.AccessListIPOut != "" {
		t.Errorf("AccessListIPOut should be empty, got %q", config.AccessListIPOut)
	}
	if config.AccessListMACIn != "" {
		t.Errorf("AccessListMACIn should be empty, got %q", config.AccessListMACIn)
	}
	if config.AccessListMACOut != "" {
		t.Errorf("AccessListMACOut should be empty, got %q", config.AccessListMACOut)
	}
}

func TestAccessListSchemaTypes(t *testing.T) {
	resource := resourceRTXInterface()

	accessListFields := []string{
		"access_list_ip_in", "access_list_ip_out",
		"access_list_ipv6_in", "access_list_ipv6_out",
		"access_list_ip_dynamic_in", "access_list_ip_dynamic_out",
		"access_list_ipv6_dynamic_in", "access_list_ipv6_dynamic_out",
		"access_list_mac_in", "access_list_mac_out",
	}

	for _, field := range accessListFields {
		s := resource.Schema[field]
		if s.Type != schema.TypeString {
			t.Errorf("%s should be TypeString, got %v", field, s.Type)
		}
		if !s.Optional {
			t.Errorf("%s should be optional", field)
		}
	}
}

func TestFlattenInterfaceConfigToResourceData_AccessLists(t *testing.T) {
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name": "lan1",
	})

	// Simulate setting access list attributes from config (as done in resourceRTXInterfaceRead)
	if err := d.Set("access_list_ip_in", "test-acl-in"); err != nil {
		t.Fatalf("Failed to set access_list_ip_in: %v", err)
	}
	if err := d.Set("access_list_ip_out", "test-acl-out"); err != nil {
		t.Fatalf("Failed to set access_list_ip_out: %v", err)
	}
	if err := d.Set("access_list_mac_in", "mac-acl-in"); err != nil {
		t.Fatalf("Failed to set access_list_mac_in: %v", err)
	}
	if err := d.Set("access_list_mac_out", "mac-acl-out"); err != nil {
		t.Fatalf("Failed to set access_list_mac_out: %v", err)
	}

	// Verify the values are correctly set and retrievable
	gotIPIn := d.Get("access_list_ip_in").(string)
	if gotIPIn != "test-acl-in" {
		t.Errorf("access_list_ip_in = %q, want %q", gotIPIn, "test-acl-in")
	}

	gotIPOut := d.Get("access_list_ip_out").(string)
	if gotIPOut != "test-acl-out" {
		t.Errorf("access_list_ip_out = %q, want %q", gotIPOut, "test-acl-out")
	}

	gotMACIn := d.Get("access_list_mac_in").(string)
	if gotMACIn != "mac-acl-in" {
		t.Errorf("access_list_mac_in = %q, want %q", gotMACIn, "mac-acl-in")
	}

	gotMACOut := d.Get("access_list_mac_out").(string)
	if gotMACOut != "mac-acl-out" {
		t.Errorf("access_list_mac_out = %q, want %q", gotMACOut, "mac-acl-out")
	}
}
