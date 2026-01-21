package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXAdminUser_Schema(t *testing.T) {
	resource := resourceRTXAdminUser()

	// Test username schema
	if s, ok := resource.Schema["username"]; !ok {
		t.Error("username schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("username should be TypeString")
		}
		if !s.Required {
			t.Error("username should be Required")
		}
		if !s.ForceNew {
			t.Error("username should be ForceNew")
		}
	}

	// Test password schema
	if s, ok := resource.Schema["password"]; !ok {
		t.Error("password schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("password should be TypeString")
		}
		if !s.Required {
			t.Error("password should be Required")
		}
		if !s.Sensitive {
			t.Error("password should be Sensitive")
		}
	}

	// Test encrypted schema
	if s, ok := resource.Schema["encrypted"]; !ok {
		t.Error("encrypted schema should exist")
	} else {
		if s.Type != schema.TypeBool {
			t.Error("encrypted should be TypeBool")
		}
		if !s.Optional {
			t.Error("encrypted should be Optional")
		}
		if !s.Computed {
			t.Error("encrypted should be Computed for import compatibility")
		}
	}

	// Test administrator schema
	if s, ok := resource.Schema["administrator"]; !ok {
		t.Error("administrator schema should exist")
	} else {
		if s.Type != schema.TypeBool {
			t.Error("administrator should be TypeBool")
		}
		if !s.Optional {
			t.Error("administrator should be Optional")
		}
		if !s.Computed {
			t.Error("administrator should be Computed for import compatibility")
		}
	}

	// Test connection_methods schema
	if s, ok := resource.Schema["connection_methods"]; !ok {
		t.Error("connection_methods schema should exist")
	} else {
		if s.Type != schema.TypeSet {
			t.Error("connection_methods should be TypeSet")
		}
		if !s.Optional {
			t.Error("connection_methods should be Optional")
		}
	}

	// Test gui_pages schema
	if s, ok := resource.Schema["gui_pages"]; !ok {
		t.Error("gui_pages schema should exist")
	} else {
		if s.Type != schema.TypeSet {
			t.Error("gui_pages should be TypeSet")
		}
		if !s.Optional {
			t.Error("gui_pages should be Optional")
		}
	}

	// Test login_timer schema
	if s, ok := resource.Schema["login_timer"]; !ok {
		t.Error("login_timer schema should exist")
	} else {
		if s.Type != schema.TypeInt {
			t.Error("login_timer should be TypeInt")
		}
		if !s.Optional {
			t.Error("login_timer should be Optional")
		}
		if !s.Computed {
			t.Error("login_timer should be Computed for import fidelity")
		}
	}
}

func TestResourceRTXAdminUser_CRUD(t *testing.T) {
	resource := resourceRTXAdminUser()

	// Verify CRUD functions exist
	if resource.CreateContext == nil {
		t.Error("CreateContext should be defined")
	}
	if resource.ReadContext == nil {
		t.Error("ReadContext should be defined")
	}
	if resource.UpdateContext == nil {
		t.Error("UpdateContext should be defined")
	}
	if resource.DeleteContext == nil {
		t.Error("DeleteContext should be defined")
	}
}

func TestResourceRTXAdminUser_Importer(t *testing.T) {
	resource := resourceRTXAdminUser()

	if resource.Importer == nil {
		t.Error("Importer should be defined")
	}
	if resource.Importer.StateContext == nil {
		t.Error("StateContext should be defined")
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid username",
			value:   "admin",
			wantErr: false,
		},
		{
			name:    "valid username with underscore",
			value:   "admin_user",
			wantErr: false,
		},
		{
			name:    "valid username with numbers",
			value:   "admin123",
			wantErr: false,
		},
		{
			name:    "empty username",
			value:   "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			value:   "1admin",
			wantErr: true,
		},
		{
			name:    "contains special characters",
			value:   "admin@user",
			wantErr: true,
		},
		{
			name:    "contains spaces",
			value:   "admin user",
			wantErr: true,
		},
		{
			name:    "contains hyphen",
			value:   "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateUsername(tt.value, "username")
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("validateUsername() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}
