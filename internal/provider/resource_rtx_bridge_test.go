package provider

import (
	"testing"
)

func TestValidateBridgeName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid bridge1",
			value:   "bridge1",
			wantErr: false,
		},
		{
			name:    "valid bridge10",
			value:   "bridge10",
			wantErr: false,
		},
		{
			name:    "valid bridge999",
			value:   "bridge999",
			wantErr: false,
		},
		{
			name:    "empty name",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - no number",
			value:   "bridge",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong prefix",
			value:   "br1",
			wantErr: true,
		},
		{
			name:    "invalid format - lan interface",
			value:   "lan1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateBridgeName(tt.value, "name")

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
				}
				return
			}

			if len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestValidateBridgeMember(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid lan1",
			value:   "lan1",
			wantErr: false,
		},
		{
			name:    "valid lan10",
			value:   "lan10",
			wantErr: false,
		},
		{
			name:    "valid VLAN interface",
			value:   "lan1/1",
			wantErr: false,
		},
		{
			name:    "valid VLAN interface 2",
			value:   "lan2/10",
			wantErr: false,
		},
		{
			name:    "valid tunnel1",
			value:   "tunnel1",
			wantErr: false,
		},
		{
			name:    "valid tunnel99",
			value:   "tunnel99",
			wantErr: false,
		},
		{
			name:    "valid pp1",
			value:   "pp1",
			wantErr: false,
		},
		{
			name:    "valid pp10",
			value:   "pp10",
			wantErr: false,
		},
		{
			name:    "valid loopback1",
			value:   "loopback1",
			wantErr: false,
		},
		{
			name:    "valid bridge interface",
			value:   "bridge2",
			wantErr: false,
		},
		{
			name:    "empty member",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - eth0",
			value:   "eth0",
			wantErr: true,
		},
		{
			name:    "invalid format - no number",
			value:   "lan",
			wantErr: true,
		},
		{
			name:    "invalid format - arbitrary string",
			value:   "interface1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateBridgeMember(tt.value, "member")

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
				}
				return
			}

			if len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestBuildBridgeFromResourceData(t *testing.T) {
	// This test would require mocking schema.ResourceData
	// For now, we test the validation functions which are the critical pieces
	t.Skip("Requires mocking schema.ResourceData")
}
