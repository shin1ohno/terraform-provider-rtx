package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXAccessListIPv6_Schema(t *testing.T) {
	resource := resourceRTXAccessListIPv6()

	// Test required fields
	requiredFields := []string{"sequence", "action", "source", "destination", "protocol"}
	for _, field := range requiredFields {
		if resource.Schema[field].Required != true {
			t.Errorf("%s should be required", field)
		}
	}

	// Test sequence is ForceNew
	if resource.Schema["sequence"].ForceNew != true {
		t.Error("sequence should force new")
	}

	// Test optional fields
	optionalFields := []string{"source_port", "dest_port"}
	for _, field := range optionalFields {
		if resource.Schema[field].Optional != true {
			t.Errorf("%s should be optional", field)
		}
	}

	// Test Computed (for import compatibility)
	if resource.Schema["source_port"].Computed != true {
		t.Errorf("source_port should be computed, got %v", resource.Schema["source_port"].Computed)
	}
	if resource.Schema["dest_port"].Computed != true {
		t.Errorf("dest_port should be computed, got %v", resource.Schema["dest_port"].Computed)
	}
}

func TestResourceRTXAccessListIPv6_ValidProtocols(t *testing.T) {
	// Verify icmp6 is in the valid protocols list
	found := false
	for _, proto := range ValidIPv6FilterProtocols {
		if proto == "icmp6" {
			found = true
			break
		}
	}
	if !found {
		t.Error("icmp6 should be in ValidIPv6FilterProtocols")
	}

	// Verify common protocols are present
	expectedProtocols := []string{"tcp", "udp", "icmp6", "ip", "*"}
	for _, expected := range expectedProtocols {
		found := false
		for _, proto := range ValidIPv6FilterProtocols {
			if proto == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s should be in ValidIPv6FilterProtocols", expected)
		}
	}
}

func TestBuildIPv6FilterFromResourceData(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"sequence":    101000,
		"action":      "pass",
		"source":      "*",
		"destination": "*",
		"protocol":    "icmp6",
		"source_port": "*",
		"dest_port":   "*",
	})

	filter := buildIPv6FilterFromResourceData(d)

	if filter.Number != 101000 {
		t.Errorf("Number = %d, want 101000", filter.Number)
	}
	if filter.Action != "pass" {
		t.Errorf("Action = %s, want pass", filter.Action)
	}
	if filter.SourceAddress != "*" {
		t.Errorf("SourceAddress = %s, want *", filter.SourceAddress)
	}
	if filter.DestAddress != "*" {
		t.Errorf("DestAddress = %s, want *", filter.DestAddress)
	}
	if filter.Protocol != "icmp6" {
		t.Errorf("Protocol = %s, want icmp6", filter.Protocol)
	}
}

func TestBuildIPv6FilterFromResourceData_WithPorts(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"sequence":    101001,
		"action":      "reject",
		"source":      "2001:db8::/32",
		"destination": "*",
		"protocol":    "tcp",
		"source_port": "1024-65535",
		"dest_port":   "80",
	})

	filter := buildIPv6FilterFromResourceData(d)

	if filter.Number != 101001 {
		t.Errorf("Number = %d, want 101001", filter.Number)
	}
	if filter.Action != "reject" {
		t.Errorf("Action = %s, want reject", filter.Action)
	}
	if filter.SourceAddress != "2001:db8::/32" {
		t.Errorf("SourceAddress = %s, want 2001:db8::/32", filter.SourceAddress)
	}
	if filter.Protocol != "tcp" {
		t.Errorf("Protocol = %s, want tcp", filter.Protocol)
	}
	if filter.SourcePort != "1024-65535" {
		t.Errorf("SourcePort = %s, want 1024-65535", filter.SourcePort)
	}
	if filter.DestPort != "80" {
		t.Errorf("DestPort = %s, want 80", filter.DestPort)
	}
}

func TestResourceRTXAccessListIPv6_ActionValidation(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	actionSchema := resource.Schema["action"]

	if actionSchema.ValidateFunc == nil {
		t.Error("action should have validation function")
	}

	// Test valid actions
	validActions := []string{"pass", "reject", "restrict", "restrict-log"}
	for _, action := range validActions {
		_, errs := actionSchema.ValidateFunc(action, "action")
		if len(errs) > 0 {
			t.Errorf("action %s should be valid, got errors: %v", action, errs)
		}
	}
}

func TestResourceRTXAccessListIPv6_ProtocolValidation(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	protocolSchema := resource.Schema["protocol"]

	if protocolSchema.ValidateFunc == nil {
		t.Error("protocol should have validation function")
	}

	// Test valid protocols including IPv6-specific icmp6
	validProtocols := []string{"tcp", "udp", "icmp6", "ip", "*"}
	for _, proto := range validProtocols {
		_, errs := protocolSchema.ValidateFunc(proto, "protocol")
		if len(errs) > 0 {
			t.Errorf("protocol %s should be valid, got errors: %v", proto, errs)
		}
	}
}

func TestResourceRTXAccessListIPv6_FilterIDValidation(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	filterIDSchema := resource.Schema["sequence"]

	if filterIDSchema.ValidateFunc == nil {
		t.Error("sequence should have validation function")
	}

	// Test valid filter IDs
	validIDs := []int{1, 100, 65535}
	for _, id := range validIDs {
		_, errs := filterIDSchema.ValidateFunc(id, "sequence")
		if len(errs) > 0 {
			t.Errorf("sequence %d should be valid, got errors: %v", id, errs)
		}
	}

	// Test invalid filter IDs (0 and -1 are invalid, 65536 is valid since range is 1-2147483647)
	invalidIDs := []int{0, -1}
	for _, id := range invalidIDs {
		_, errs := filterIDSchema.ValidateFunc(id, "sequence")
		if len(errs) == 0 {
			t.Errorf("sequence %d should be invalid", id)
		}
	}

	// 65536 is valid since range is 1-2147483647
	_, errs := filterIDSchema.ValidateFunc(65536, "sequence")
	if len(errs) > 0 {
		t.Errorf("sequence 65536 should be valid, got errors: %v", errs)
	}
}
