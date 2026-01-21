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
	optionalFields := []string{"description", "ip_address", "secure_filter_in", "secure_filter_out", "dynamic_filter_out", "nat_descriptor", "proxyarp", "mtu", "ethernet_filter_in", "ethernet_filter_out"}
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

func TestBuildInterfaceConfigFromResourceData_EthernetFilter(t *testing.T) {
	// Create a mock ResourceData with ethernet filters
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name":                "lan1",
		"ethernet_filter_in":  []interface{}{1, 10, 100},
		"ethernet_filter_out": []interface{}{2, 20, 200},
	})

	config := buildInterfaceConfigFromResourceData(d)

	// Verify ethernet_filter_in
	if len(config.EthernetFilterIn) != 3 {
		t.Errorf("EthernetFilterIn length = %d, want 3", len(config.EthernetFilterIn))
	}
	expectedIn := []int{1, 10, 100}
	for i, v := range expectedIn {
		if config.EthernetFilterIn[i] != v {
			t.Errorf("EthernetFilterIn[%d] = %d, want %d", i, config.EthernetFilterIn[i], v)
		}
	}

	// Verify ethernet_filter_out
	if len(config.EthernetFilterOut) != 3 {
		t.Errorf("EthernetFilterOut length = %d, want 3", len(config.EthernetFilterOut))
	}
	expectedOut := []int{2, 20, 200}
	for i, v := range expectedOut {
		if config.EthernetFilterOut[i] != v {
			t.Errorf("EthernetFilterOut[%d] = %d, want %d", i, config.EthernetFilterOut[i], v)
		}
	}
}

func TestBuildInterfaceConfigFromResourceData_EthernetFilterEmpty(t *testing.T) {
	// Create a mock ResourceData without ethernet filters
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name": "lan1",
	})

	config := buildInterfaceConfigFromResourceData(d)

	// Verify ethernet filters are nil when not set
	if config.EthernetFilterIn != nil {
		t.Errorf("EthernetFilterIn should be nil, got %v", config.EthernetFilterIn)
	}
	if config.EthernetFilterOut != nil {
		t.Errorf("EthernetFilterOut should be nil, got %v", config.EthernetFilterOut)
	}
}

func TestEthernetFilterSchemaValidation(t *testing.T) {
	resource := resourceRTXInterface()

	// Get the validation function from the schema
	ethernetFilterInSchema := resource.Schema["ethernet_filter_in"]
	if ethernetFilterInSchema.Type != schema.TypeList {
		t.Error("ethernet_filter_in should be TypeList")
	}

	elemSchema, ok := ethernetFilterInSchema.Elem.(*schema.Schema)
	if !ok {
		t.Fatal("ethernet_filter_in Elem should be *schema.Schema")
	}

	if elemSchema.ValidateFunc == nil {
		t.Fatal("ethernet_filter_in element should have ValidateFunc")
	}

	// Test validation with valid values
	validValues := []int{1, 256, 512}
	for _, v := range validValues {
		_, errors := elemSchema.ValidateFunc(v, "test")
		if len(errors) > 0 {
			t.Errorf("ValidateFunc(%d) should pass, got errors: %v", v, errors)
		}
	}

	// Test validation with invalid values
	invalidValues := []int{0, -1, 513, 1000}
	for _, v := range invalidValues {
		_, errors := elemSchema.ValidateFunc(v, "test")
		if len(errors) == 0 {
			t.Errorf("ValidateFunc(%d) should fail, but passed", v)
		}
	}

	// Verify ethernet_filter_out has the same validation
	ethernetFilterOutSchema := resource.Schema["ethernet_filter_out"]
	outElemSchema, ok := ethernetFilterOutSchema.Elem.(*schema.Schema)
	if !ok {
		t.Fatal("ethernet_filter_out Elem should be *schema.Schema")
	}

	if outElemSchema.ValidateFunc == nil {
		t.Fatal("ethernet_filter_out element should have ValidateFunc")
	}

	// Test outbound filter validation
	for _, v := range validValues {
		_, errors := outElemSchema.ValidateFunc(v, "test")
		if len(errors) > 0 {
			t.Errorf("ethernet_filter_out ValidateFunc(%d) should pass, got errors: %v", v, errors)
		}
	}
	for _, v := range invalidValues {
		_, errors := outElemSchema.ValidateFunc(v, "test")
		if len(errors) == 0 {
			t.Errorf("ethernet_filter_out ValidateFunc(%d) should fail, but passed", v)
		}
	}
}

func TestFlattenInterfaceConfigToResourceData_EthernetFilter(t *testing.T) {
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name": "lan1",
	})

	// Simulate setting ethernet filters from config (as done in resourceRTXInterfaceRead)
	ethernetFilterIn := []int{1, 50, 512}
	ethernetFilterOut := []int{10, 100, 500}

	if err := d.Set("ethernet_filter_in", ethernetFilterIn); err != nil {
		t.Fatalf("Failed to set ethernet_filter_in: %v", err)
	}
	if err := d.Set("ethernet_filter_out", ethernetFilterOut); err != nil {
		t.Fatalf("Failed to set ethernet_filter_out: %v", err)
	}

	// Verify the values are correctly set and retrievable
	gotIn := d.Get("ethernet_filter_in").([]interface{})
	if len(gotIn) != 3 {
		t.Errorf("ethernet_filter_in length = %d, want 3", len(gotIn))
	}
	for i, expected := range ethernetFilterIn {
		if gotIn[i].(int) != expected {
			t.Errorf("ethernet_filter_in[%d] = %d, want %d", i, gotIn[i], expected)
		}
	}

	gotOut := d.Get("ethernet_filter_out").([]interface{})
	if len(gotOut) != 3 {
		t.Errorf("ethernet_filter_out length = %d, want 3", len(gotOut))
	}
	for i, expected := range ethernetFilterOut {
		if gotOut[i].(int) != expected {
			t.Errorf("ethernet_filter_out[%d] = %d, want %d", i, gotOut[i], expected)
		}
	}
}

func TestFlattenInterfaceConfigToResourceData_EthernetFilterEmpty(t *testing.T) {
	resource := resourceRTXInterface()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name": "lan1",
	})

	// Simulate setting empty/nil ethernet filters (as done in resourceRTXInterfaceRead)
	var emptyFilters []int

	if err := d.Set("ethernet_filter_in", emptyFilters); err != nil {
		t.Fatalf("Failed to set ethernet_filter_in: %v", err)
	}
	if err := d.Set("ethernet_filter_out", emptyFilters); err != nil {
		t.Fatalf("Failed to set ethernet_filter_out: %v", err)
	}

	// Verify empty filters are handled correctly
	gotIn := d.Get("ethernet_filter_in").([]interface{})
	if len(gotIn) != 0 {
		t.Errorf("ethernet_filter_in should be empty, got %v", gotIn)
	}

	gotOut := d.Get("ethernet_filter_out").([]interface{})
	if len(gotOut) != 0 {
		t.Errorf("ethernet_filter_out should be empty, got %v", gotOut)
	}
}
