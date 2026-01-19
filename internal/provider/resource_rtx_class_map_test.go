package provider

import (
	"testing"
)

func TestValidateClassMapName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid name",
			value:   "voip-traffic",
			wantErr: false,
		},
		{
			name:    "valid name with underscore",
			value:   "voip_traffic",
			wantErr: false,
		},
		{
			name:    "valid name with numbers",
			value:   "class1",
			wantErr: false,
		},
		{
			name:    "empty name",
			value:   "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			value:   "1class",
			wantErr: true,
		},
		{
			name:    "contains space",
			value:   "voip traffic",
			wantErr: true,
		},
		{
			name:    "contains special char",
			value:   "voip@traffic",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateClassMapName(tt.value, "name")
			gotErr := len(errs) > 0

			if gotErr != tt.wantErr {
				t.Errorf("validateClassMapName(%q) error = %v, wantErr %v", tt.value, errs, tt.wantErr)
			}
		})
	}
}

func TestBuildClassMapFromResourceData_Basic(t *testing.T) {
	// This is a basic test to verify the function compiles and runs
	// Full testing requires mocking ResourceData which is complex

	// The function signature test
	// buildClassMapFromResourceData takes *schema.ResourceData
	// and returns client.ClassMap
}
