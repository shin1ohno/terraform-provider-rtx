package provider

import (
	"testing"
)

func TestValidatePolicyMapName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid name",
			value:   "qos-policy",
			wantErr: false,
		},
		{
			name:    "valid name with underscore",
			value:   "qos_policy",
			wantErr: false,
		},
		{
			name:    "valid name with numbers",
			value:   "policy1",
			wantErr: false,
		},
		{
			name:    "empty name",
			value:   "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			value:   "1policy",
			wantErr: true,
		},
		{
			name:    "contains space",
			value:   "qos policy",
			wantErr: true,
		},
		{
			name:    "contains special char",
			value:   "qos@policy",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validatePolicyMapName(tt.value, "name")
			gotErr := len(errs) > 0

			if gotErr != tt.wantErr {
				t.Errorf("validatePolicyMapName(%q) error = %v, wantErr %v", tt.value, errs, tt.wantErr)
			}
		})
	}
}
