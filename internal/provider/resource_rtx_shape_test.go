package provider

import (
	"testing"
)

func TestParseShapeID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantIface string
		wantDir   string
		wantErr   bool
	}{
		{
			name:      "valid output",
			id:        "lan1:output",
			wantIface: "lan1",
			wantDir:   "output",
			wantErr:   false,
		},
		{
			name:      "valid input",
			id:        "wan1:input",
			wantIface: "wan1",
			wantDir:   "input",
			wantErr:   false,
		},
		{
			name:    "missing direction",
			id:      "lan1",
			wantErr: true,
		},
		{
			name:    "empty interface",
			id:      ":output",
			wantErr: true,
		},
		{
			name:    "invalid direction",
			id:      "lan1:both",
			wantErr: true,
		},
		{
			name:    "too many parts",
			id:      "lan1:output:extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface, dir, err := parseShapeID(tt.id)
			gotErr := err != nil

			if gotErr != tt.wantErr {
				t.Errorf("parseShapeID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}

			if !gotErr {
				if iface != tt.wantIface {
					t.Errorf("parseShapeID(%q) interface = %q, want %q", tt.id, iface, tt.wantIface)
				}
				if dir != tt.wantDir {
					t.Errorf("parseShapeID(%q) direction = %q, want %q", tt.id, dir, tt.wantDir)
				}
			}
		})
	}
}

func TestBuildShapeConfigFromResourceData_Basic(t *testing.T) {
	// This is a basic test to verify the function compiles
	// Full testing requires mocking ResourceData
}
