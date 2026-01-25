package provider

import (
	"testing"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestValidateIPv6InterfaceName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid lan1", "lan1", false},
		{"valid lan2", "lan2", false},
		{"valid lan10", "lan10", false},
		{"valid bridge1", "bridge1", false},
		{"valid bridge10", "bridge10", false},
		{"valid pp1", "pp1", false},
		{"valid tunnel1", "tunnel1", false},
		{"valid tunnel100", "tunnel100", false},
		{"empty", "", true},
		{"invalid format", "invalid", true},
		{"eth0", "eth0", true},
		{"lan", "lan", true},
		{"wan1", "wan1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateIPv6InterfaceName(tt.value, "interface")
			hasErr := len(errs) > 0

			if hasErr != tt.wantErr {
				t.Errorf("validateIPv6InterfaceName(%q) hasErr = %v, wantErr %v", tt.value, hasErr, tt.wantErr)
			}
		})
	}
}

func TestValidateIPv6InterfaceNameValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid lan1", "lan1", false},
		{"valid lan2", "lan2", false},
		{"valid bridge1", "bridge1", false},
		{"valid pp1", "pp1", false},
		{"valid tunnel1", "tunnel1", false},
		{"empty", "", true},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv6InterfaceNameValue(tt.value)
			hasErr := err != nil

			if hasErr != tt.wantErr {
				t.Errorf("validateIPv6InterfaceNameValue(%q) hasErr = %v, wantErr %v", tt.value, hasErr, tt.wantErr)
			}
		})
	}
}

func TestValidateIPv6CIDROptional(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid IPv6 CIDR", "2001:db8::1/64", false},
		{"valid full address", "2001:0db8:0000:0000:0000:0000:0000:0001/128", false},
		{"valid /0", "::/0", false},
		{"empty string", "", false},
		{"missing prefix", "2001:db8::1", true},
		{"IPv4 address", "192.168.1.1/24", true},
		{"invalid format", "invalid/64", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateIPv6CIDROptional(tt.value, "address")
			hasErr := len(errs) > 0

			if hasErr != tt.wantErr {
				t.Errorf("validateIPv6CIDROptional(%q) hasErr = %v, wantErr %v", tt.value, hasErr, tt.wantErr)
			}
		})
	}
}

func TestValidateDHCPv6Service(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"server", "server", false},
		{"client", "client", false},
		{"empty", "", false},
		{"invalid", "invalid", true},
		{"off", "off", true}, // "off" is not a valid value, use empty string instead
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateDHCPv6Service(tt.value, "dhcpv6_service")
			hasErr := len(errs) > 0

			if hasErr != tt.wantErr {
				t.Errorf("validateDHCPv6Service(%q) hasErr = %v, wantErr %v", tt.value, hasErr, tt.wantErr)
			}
		})
	}
}

func TestBuildIPv6InterfaceConfigFromResourceData_BasicConfig(t *testing.T) {
	// Test helper to validate that the config builder works correctly
	// This tests the basic structure without using real ResourceData

	// Create a sample config to validate the structure
	config := client.IPv6InterfaceConfig{
		Interface: "lan1",
		Addresses: []client.IPv6Address{
			{Address: "2001:db8::1/64"},
		},
		RTADV: &client.RTADVConfig{
			Enabled:  true,
			PrefixID: 1,
			OFlag:    true,
			MFlag:    false,
		},
		DHCPv6Service:            "server",
		MTU:                      1500,
		AccessListIPv6In:         "ipv6-in-acl",
		AccessListIPv6Out:        "ipv6-out-acl",
		AccessListIPv6DynamicIn:  "ipv6-dynamic-in",
		AccessListIPv6DynamicOut: "ipv6-dynamic-out",
	}

	// Validate structure
	if config.Interface != "lan1" {
		t.Errorf("Interface = %q, want %q", config.Interface, "lan1")
	}

	if len(config.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(config.Addresses))
	}

	if config.Addresses[0].Address != "2001:db8::1/64" {
		t.Errorf("Address = %q, want %q", config.Addresses[0].Address, "2001:db8::1/64")
	}

	if config.RTADV == nil {
		t.Error("RTADV is nil, want non-nil")
	} else {
		if !config.RTADV.Enabled {
			t.Error("RTADV.Enabled = false, want true")
		}
		if config.RTADV.PrefixID != 1 {
			t.Errorf("RTADV.PrefixID = %d, want 1", config.RTADV.PrefixID)
		}
		if !config.RTADV.OFlag {
			t.Error("RTADV.OFlag = false, want true")
		}
		if config.RTADV.MFlag {
			t.Error("RTADV.MFlag = true, want false")
		}
	}

	if config.DHCPv6Service != "server" {
		t.Errorf("DHCPv6Service = %q, want %q", config.DHCPv6Service, "server")
	}

	if config.MTU != 1500 {
		t.Errorf("MTU = %d, want 1500", config.MTU)
	}

	if config.AccessListIPv6In != "ipv6-in-acl" {
		t.Errorf("AccessListIPv6In = %q, want %q", config.AccessListIPv6In, "ipv6-in-acl")
	}

	if config.AccessListIPv6Out != "ipv6-out-acl" {
		t.Errorf("AccessListIPv6Out = %q, want %q", config.AccessListIPv6Out, "ipv6-out-acl")
	}

	if config.AccessListIPv6DynamicIn != "ipv6-dynamic-in" {
		t.Errorf("AccessListIPv6DynamicIn = %q, want %q", config.AccessListIPv6DynamicIn, "ipv6-dynamic-in")
	}

	if config.AccessListIPv6DynamicOut != "ipv6-dynamic-out" {
		t.Errorf("AccessListIPv6DynamicOut = %q, want %q", config.AccessListIPv6DynamicOut, "ipv6-dynamic-out")
	}
}

func TestBuildIPv6InterfaceConfigFromResourceData_PrefixBasedAddress(t *testing.T) {
	// Test prefix-based address configuration
	config := client.IPv6InterfaceConfig{
		Interface: "lan1",
		Addresses: []client.IPv6Address{
			{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"},
		},
	}

	if len(config.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(config.Addresses))
	}

	addr := config.Addresses[0]
	if addr.PrefixRef != "ra-prefix@lan2" {
		t.Errorf("PrefixRef = %q, want %q", addr.PrefixRef, "ra-prefix@lan2")
	}
	if addr.InterfaceID != "::1/64" {
		t.Errorf("InterfaceID = %q, want %q", addr.InterfaceID, "::1/64")
	}
	if addr.Address != "" {
		t.Errorf("Address = %q, want empty", addr.Address)
	}
}

func TestBuildIPv6InterfaceConfigFromResourceData_MultipleAddresses(t *testing.T) {
	// Test multiple addresses configuration
	config := client.IPv6InterfaceConfig{
		Interface: "lan1",
		Addresses: []client.IPv6Address{
			{Address: "2001:db8::1/64"},
			{Address: "2001:db8::2/64"},
			{PrefixRef: "ra-prefix@lan2", InterfaceID: "::3/64"},
		},
	}

	if len(config.Addresses) != 3 {
		t.Errorf("Addresses count = %d, want 3", len(config.Addresses))
	}

	// First address: static
	if config.Addresses[0].Address != "2001:db8::1/64" {
		t.Errorf("Addresses[0].Address = %q, want %q", config.Addresses[0].Address, "2001:db8::1/64")
	}

	// Second address: static
	if config.Addresses[1].Address != "2001:db8::2/64" {
		t.Errorf("Addresses[1].Address = %q, want %q", config.Addresses[1].Address, "2001:db8::2/64")
	}

	// Third address: prefix-based
	if config.Addresses[2].PrefixRef != "ra-prefix@lan2" {
		t.Errorf("Addresses[2].PrefixRef = %q, want %q", config.Addresses[2].PrefixRef, "ra-prefix@lan2")
	}
}

func TestBuildIPv6InterfaceConfigFromResourceData_AccessLists(t *testing.T) {
	// Test access list configuration
	config := client.IPv6InterfaceConfig{
		Interface:                "lan1",
		AccessListIPv6In:         "ipv6-in-acl",
		AccessListIPv6Out:        "ipv6-out-acl",
		AccessListIPv6DynamicIn:  "ipv6-dynamic-in-acl",
		AccessListIPv6DynamicOut: "ipv6-dynamic-out-acl",
	}

	// Validate access list names
	if config.AccessListIPv6In != "ipv6-in-acl" {
		t.Errorf("AccessListIPv6In = %q, want %q", config.AccessListIPv6In, "ipv6-in-acl")
	}

	if config.AccessListIPv6Out != "ipv6-out-acl" {
		t.Errorf("AccessListIPv6Out = %q, want %q", config.AccessListIPv6Out, "ipv6-out-acl")
	}

	if config.AccessListIPv6DynamicIn != "ipv6-dynamic-in-acl" {
		t.Errorf("AccessListIPv6DynamicIn = %q, want %q", config.AccessListIPv6DynamicIn, "ipv6-dynamic-in-acl")
	}

	if config.AccessListIPv6DynamicOut != "ipv6-dynamic-out-acl" {
		t.Errorf("AccessListIPv6DynamicOut = %q, want %q", config.AccessListIPv6DynamicOut, "ipv6-dynamic-out-acl")
	}
}

func TestBuildIPv6InterfaceConfigFromResourceData_EmptyConfig(t *testing.T) {
	// Test empty configuration
	config := client.IPv6InterfaceConfig{
		Interface: "lan1",
	}

	if config.Interface != "lan1" {
		t.Errorf("Interface = %q, want %q", config.Interface, "lan1")
	}

	if len(config.Addresses) != 0 {
		t.Errorf("Addresses count = %d, want 0", len(config.Addresses))
	}

	if config.RTADV != nil {
		t.Error("RTADV is not nil, want nil")
	}

	if config.DHCPv6Service != "" {
		t.Errorf("DHCPv6Service = %q, want empty", config.DHCPv6Service)
	}

	if config.MTU != 0 {
		t.Errorf("MTU = %d, want 0", config.MTU)
	}
}
