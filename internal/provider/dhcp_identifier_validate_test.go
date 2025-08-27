package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateClientIdentification(t *testing.T) {
	tests := []struct {
		name        string
		macAddress  interface{}
		clientID    interface{}
		useClientID interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Both mac_address and client_identifier set",
			macAddress:  "00:11:22:33:44:55",
			clientID:    "01:00:11:22:33:44:55",
			useClientID: false,
			expectError: true,
			errorMsg:    "exactly one of mac_address or client_identifier must be specified",
		},
		{
			name:        "Neither mac_address nor client_identifier set",
			macAddress:  nil,
			clientID:    nil,
			useClientID: false,
			expectError: true,
			errorMsg:    "exactly one of mac_address or client_identifier must be specified",
		},
		{
			name:        "Only mac_address set - valid",
			macAddress:  "00:11:22:33:44:55",
			clientID:    nil,
			useClientID: false,
			expectError: false,
		},
		{
			name:        "Only mac_address set with use_mac_as_client_id - valid",
			macAddress:  "00:11:22:33:44:55",
			clientID:    nil,
			useClientID: true,
			expectError: false,
		},
		{
			name:        "Only client_identifier set - valid",
			macAddress:  nil,
			clientID:    "01:00:11:22:33:44:55",
			useClientID: false,
			expectError: false,
		},
		{
			name:        "client_identifier with use_mac_as_client_id conflict",
			macAddress:  nil,
			clientID:    "01:00:11:22:33:44:55",
			useClientID: true,
			expectError: true,
			errorMsg:    "use_mac_as_client_id cannot be used with client_identifier",
		},
		{
			name:        "Empty strings should be treated as unset",
			macAddress:  "",
			clientID:    "",
			useClientID: false,
			expectError: true,
			errorMsg:    "exactly one of mac_address or client_identifier must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock resource data
			resourceSchema := map[string]*schema.Schema{
				"mac_address": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"client_identifier": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"use_mac_as_client_id": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			}

			d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
				"mac_address":          tt.macAddress,
				"client_identifier":    tt.clientID,
				"use_mac_as_client_id": tt.useClientID,
			})

			err := validateClientIdentificationWithResourceData(context.Background(), d)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateClientIdentifierFormat(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		expectError bool
		errorMsg    string
	}{
		// Valid MAC-based (01 prefix)
		{
			name:        "Valid MAC-based identifier",
			identifier:  "01:00:11:22:33:44:55",
			expectError: false,
		},
		{
			name:        "Valid MAC-based identifier uppercase",
			identifier:  "01:AA:BB:CC:DD:EE:FF",
			expectError: false,
		},
		// Valid ASCII-based (02 prefix)  
		{
			name:        "Valid ASCII-based identifier",
			identifier:  "02:68:6f:73:74:6e:61:6d:65", // "hostname" in hex
			expectError: false,
		},
		// Valid vendor-specific (FF prefix)
		{
			name:        "Valid vendor-specific identifier", 
			identifier:  "ff:00:01:02:03:04:05",
			expectError: false,
		},
		// Invalid formats
		{
			name:        "Invalid prefix - not supported",
			identifier:  "03:00:11:22:33:44:55",
			expectError: true,
			errorMsg:    "client identifier prefix must be 01 (MAC), 02 (ASCII), or ff (vendor-specific)",
		},
		{
			name:        "No data after prefix",
			identifier:  "01:",
			expectError: true,
			errorMsg:    "client identifier must have data after type prefix",
		},
		{
			name:        "No prefix",
			identifier:  "001122334455",
			expectError: true,
			errorMsg:    "client identifier must be in format 'type:data'",
		},
		{
			name:        "Non-hex characters",
			identifier:  "01:zz:11:22:33:44:55",
			expectError: true,
			errorMsg:    "client identifier contains invalid hex characters",
		},
		{
			name:        "Too long identifier",
			identifier:  "01:" + repeatString("aa:", 127) + "bb",
			expectError: true,
			errorMsg:    "client identifier too long (max 255 bytes)",
		},
		{
			name:        "Empty string",
			identifier:  "",
			expectError: true,
			errorMsg:    "client identifier cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClientIdentifierFormatSimple(tt.identifier)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to create strings of specific length for testing
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}