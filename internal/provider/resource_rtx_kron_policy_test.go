package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXKronPolicySchema(t *testing.T) {
	resource := resourceRTXKronPolicy()

	// Test that the resource has the expected schema
	if resource.Schema["name"] == nil {
		t.Error("expected 'name' field in schema")
	}
	if resource.Schema["command_lines"] == nil {
		t.Error("expected 'command_lines' field in schema")
	}
}

func TestResourceRTXKronPolicySchemaAttributes(t *testing.T) {
	resource := resourceRTXKronPolicy()

	tests := []struct {
		field    string
		expected struct {
			Type     schema.ValueType
			Required bool
			ForceNew bool
		}
	}{
		{
			field: "name",
			expected: struct {
				Type     schema.ValueType
				Required bool
				ForceNew bool
			}{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		{
			field: "command_lines",
			expected: struct {
				Type     schema.ValueType
				Required bool
				ForceNew bool
			}{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			s := resource.Schema[tt.field]
			if s == nil {
				t.Fatalf("field %q not found in schema", tt.field)
			}

			if s.Type != tt.expected.Type {
				t.Errorf("field %q: Type = %v, want %v", tt.field, s.Type, tt.expected.Type)
			}
			if s.Required != tt.expected.Required {
				t.Errorf("field %q: Required = %v, want %v", tt.field, s.Required, tt.expected.Required)
			}
			if s.ForceNew != tt.expected.ForceNew {
				t.Errorf("field %q: ForceNew = %v, want %v", tt.field, s.ForceNew, tt.expected.ForceNew)
			}
		})
	}
}

func TestValidateKronPolicyName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid simple name", value: "backup", wantErr: false},
		{name: "valid with underscore", value: "daily_backup", wantErr: false},
		{name: "valid with hyphen", value: "daily-backup", wantErr: false},
		{name: "valid with numbers", value: "backup1", wantErr: false},
		{name: "valid mixed", value: "Daily_Backup-1", wantErr: false},
		{name: "starts with number", value: "1backup", wantErr: true},
		{name: "contains spaces", value: "my backup", wantErr: true},
		{name: "empty", value: "", wantErr: true},
		{name: "special characters", value: "backup@home", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := validateKronPolicyName(tt.value, "name")

			if tt.wantErr && len(errors) == 0 {
				t.Errorf("expected error for value %q, got none", tt.value)
			}
			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("unexpected error for value %q: %v", tt.value, errors)
			}
		})
	}
}

func TestValidateKronPolicyNameValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: "policy1", wantErr: false},
		{name: "valid long", input: "my_long_policy_name_with_many_characters", wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "too long", input: "this_is_a_very_long_policy_name_that_exceeds_the_sixty_four_character_limit_allowed", wantErr: true},
		{name: "starts with underscore", input: "_policy", wantErr: true},
		{name: "starts with hyphen", input: "-policy", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKronPolicyNameValue(tt.input)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}
