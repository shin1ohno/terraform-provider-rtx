package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXAdmin_Schema(t *testing.T) {
	resource := resourceRTXAdmin()

	// Test login_password schema
	if s, ok := resource.Schema["login_password"]; !ok {
		t.Error("login_password schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("login_password should be TypeString")
		}
		if !s.Optional {
			t.Error("login_password should be Optional")
		}
		if !s.Sensitive {
			t.Error("login_password should be Sensitive")
		}
	}

	// Test admin_password schema
	if s, ok := resource.Schema["admin_password"]; !ok {
		t.Error("admin_password schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("admin_password should be TypeString")
		}
		if !s.Optional {
			t.Error("admin_password should be Optional")
		}
		if !s.Sensitive {
			t.Error("admin_password should be Sensitive")
		}
	}
}

func TestResourceRTXAdmin_CRUD(t *testing.T) {
	resource := resourceRTXAdmin()

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

func TestResourceRTXAdmin_Importer(t *testing.T) {
	resource := resourceRTXAdmin()

	if resource.Importer == nil {
		t.Error("Importer should be defined")
	}
	if resource.Importer.StateContext == nil {
		t.Error("StateContext should be defined")
	}
}
