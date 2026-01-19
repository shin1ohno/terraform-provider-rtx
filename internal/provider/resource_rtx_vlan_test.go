package provider

import (
	"testing"
)

func TestParseVLANID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		wantIface   string
		wantVlanID  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid ID",
			id:         "lan1/10",
			wantIface:  "lan1",
			wantVlanID: 10,
			wantErr:    false,
		},
		{
			name:       "valid ID with lan2",
			id:         "lan2/100",
			wantIface:  "lan2",
			wantVlanID: 100,
			wantErr:    false,
		},
		{
			name:       "valid ID with high VLAN",
			id:         "lan1/4094",
			wantIface:  "lan1",
			wantVlanID: 4094,
			wantErr:    false,
		},
		{
			name:        "missing VLAN ID",
			id:          "lan1",
			wantErr:     true,
			errContains: "expected format",
		},
		{
			name:        "too many parts",
			id:          "lan1/10/extra",
			wantErr:     true,
			errContains: "expected format",
		},
		{
			name:        "non-numeric VLAN ID",
			id:          "lan1/abc",
			wantErr:     true,
			errContains: "invalid VLAN ID",
		},
		{
			name:        "empty string",
			id:          "",
			wantErr:     true,
			errContains: "expected format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface, vlanID, err := parseVLANID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseVLANID() expected error, got nil")
					return
				}
				if tt.errContains != "" && !containsStr(err.Error(), tt.errContains) {
					t.Errorf("parseVLANID() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("parseVLANID() unexpected error: %v", err)
				return
			}

			if iface != tt.wantIface {
				t.Errorf("parseVLANID() iface = %q, want %q", iface, tt.wantIface)
			}
			if vlanID != tt.wantVlanID {
				t.Errorf("parseVLANID() vlanID = %d, want %d", vlanID, tt.wantVlanID)
			}
		})
	}
}

func TestValidateVLANInterfaceName(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		key       string
		wantWarn  bool
		wantError bool
	}{
		{
			name:      "valid lan1",
			value:     "lan1",
			key:       "interface",
			wantWarn:  false,
			wantError: false,
		},
		{
			name:      "valid lan2",
			value:     "lan2",
			key:       "interface",
			wantWarn:  false,
			wantError: false,
		},
		{
			name:      "valid lan10",
			value:     "lan10",
			key:       "interface",
			wantWarn:  false,
			wantError: false,
		},
		{
			name:      "empty string",
			value:     "",
			key:       "interface",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid eth0",
			value:     "eth0",
			key:       "interface",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid pp1",
			value:     "pp1",
			key:       "interface",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid lan",
			value:     "lan",
			key:       "interface",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid lanX",
			value:     "lanX",
			key:       "interface",
			wantWarn:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns, errs := validateVLANInterfaceName(tt.value, tt.key)

			if tt.wantWarn && len(warns) == 0 {
				t.Error("validateVLANInterfaceName() expected warnings, got none")
			}
			if !tt.wantWarn && len(warns) > 0 {
				t.Errorf("validateVLANInterfaceName() unexpected warnings: %v", warns)
			}
			if tt.wantError && len(errs) == 0 {
				t.Error("validateVLANInterfaceName() expected errors, got none")
			}
			if !tt.wantError && len(errs) > 0 {
				t.Errorf("validateVLANInterfaceName() unexpected errors: %v", errs)
			}
		})
	}
}

func TestValidateVLANSubnetMask(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		key       string
		wantError bool
	}{
		{
			name:      "valid 255.255.255.0",
			value:     "255.255.255.0",
			key:       "ip_mask",
			wantError: false,
		},
		{
			name:      "valid 255.255.0.0",
			value:     "255.255.0.0",
			key:       "ip_mask",
			wantError: false,
		},
		{
			name:      "valid 255.0.0.0",
			value:     "255.0.0.0",
			key:       "ip_mask",
			wantError: false,
		},
		{
			name:      "valid 255.255.255.240",
			value:     "255.255.255.240",
			key:       "ip_mask",
			wantError: false,
		},
		{
			name:      "empty string (optional)",
			value:     "",
			key:       "ip_mask",
			wantError: false,
		},
		{
			name:      "invalid single number",
			value:     "24",
			key:       "ip_mask",
			wantError: true,
		},
		{
			name:      "invalid CIDR",
			value:     "/24",
			key:       "ip_mask",
			wantError: true,
		},
		{
			name:      "invalid IP-like",
			value:     "192.168.1.0",
			key:       "ip_mask",
			wantError: false, // Valid format, even if not a typical mask
		},
		{
			name:      "invalid non-mask value",
			value:     "abc.def.ghi.jkl",
			key:       "ip_mask",
			wantError: true,
		},
		{
			name:      "invalid out of range",
			value:     "256.255.255.0",
			key:       "ip_mask",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateSubnetMask(tt.value, tt.key)

			if tt.wantError && len(errs) == 0 {
				t.Error("validateSubnetMask() expected errors, got none")
			}
			if !tt.wantError && len(errs) > 0 {
				t.Errorf("validateSubnetMask() unexpected errors: %v", errs)
			}
		})
	}
}

// containsStr checks if s contains substr
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
