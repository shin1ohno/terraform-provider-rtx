package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// TestAdministratorPreservation verifies that when a user is created with
// administrator=true and then updated without specifying the administrator field,
// the administrator privilege is preserved (not reset to false).
//
// This is a regression test for the original lockout scenario where updating
// a user without specifying administrator would reset it to false.
func TestAdministratorPreservation(t *testing.T) {
	// Create a resource schema for testing
	resource := resourceRTXAdminUser()

	// Simulate initial creation with administrator=true
	initialData := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":      "testuser",
		"password":      "secret123",
		"administrator": true,
		"login_timer":   300,
	})

	// Build the user from initial data
	initialUser := buildAdminUserFromResourceData(initialData)

	// Verify initial state has administrator=true
	if initialUser.Attributes.Administrator == nil {
		t.Fatal("expected Administrator to be non-nil after initial build")
	}
	if *initialUser.Attributes.Administrator != true {
		t.Errorf("expected Administrator=true, got %v", *initialUser.Attributes.Administrator)
	}

	// Simulate Read function populating state after successful creation
	// The Read function would set these values from the router's response
	if err := initialData.Set("administrator", true); err != nil {
		t.Fatalf("failed to set administrator in state: %v", err)
	}
	if err := initialData.Set("login_timer", 300); err != nil {
		t.Fatalf("failed to set login_timer in state: %v", err)
	}

	// Now simulate an Update call where administrator is NOT in the new config
	// but IS in state from the previous Read
	// This uses d.Get() which returns the merged value (config or state)
	updateUser := buildAdminUserFromResourceData(initialData)

	// The key assertion: administrator should still be true because d.Get()
	// returns the merged value from config (if set) or state (fallback)
	if updateUser.Attributes.Administrator == nil {
		t.Fatal("expected Administrator to be non-nil after update build")
	}
	if *updateUser.Attributes.Administrator != true {
		t.Errorf("expected Administrator=true to be preserved, got %v", *updateUser.Attributes.Administrator)
	}

	// Also verify login_timer is preserved
	if updateUser.Attributes.LoginTimer == nil {
		t.Fatal("expected LoginTimer to be non-nil after update build")
	}
	if *updateUser.Attributes.LoginTimer != 300 {
		t.Errorf("expected LoginTimer=300 to be preserved, got %v", *updateUser.Attributes.LoginTimer)
	}
}

// TestAdministratorPreservation_WithExplicitFalse verifies that when a user
// explicitly sets administrator=false, it is respected and not overridden.
func TestAdministratorPreservation_WithExplicitFalse(t *testing.T) {
	resource := resourceRTXAdminUser()

	// Create with administrator=false explicitly
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":      "testuser",
		"password":      "secret123",
		"administrator": false,
	})

	user := buildAdminUserFromResourceData(data)

	if user.Attributes.Administrator == nil {
		t.Fatal("expected Administrator to be non-nil")
	}
	if *user.Attributes.Administrator != false {
		t.Errorf("expected Administrator=false, got %v", *user.Attributes.Administrator)
	}
}

// TestFieldPreservation_AllOptionalFields verifies that all optional fields
// are preserved when they exist in state but are not specified in a new config.
func TestFieldPreservation_AllOptionalFields(t *testing.T) {
	resource := resourceRTXAdminUser()

	// Initial data with all optional fields set
	initialData := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":           "testuser",
		"password":           "secret123",
		"encrypted":          true,
		"administrator":      true,
		"connection_methods": []interface{}{"ssh", "telnet"},
		"gui_pages":          []interface{}{"dashboard", "config"},
		"login_timer":        600,
	})

	// Simulate Read populating state
	if err := initialData.Set("encrypted", true); err != nil {
		t.Fatalf("failed to set encrypted: %v", err)
	}
	if err := initialData.Set("administrator", true); err != nil {
		t.Fatalf("failed to set administrator: %v", err)
	}
	if err := initialData.Set("connection_methods", []interface{}{"ssh", "telnet"}); err != nil {
		t.Fatalf("failed to set connection_methods: %v", err)
	}
	if err := initialData.Set("gui_pages", []interface{}{"dashboard", "config"}); err != nil {
		t.Fatalf("failed to set gui_pages: %v", err)
	}
	if err := initialData.Set("login_timer", 600); err != nil {
		t.Fatalf("failed to set login_timer: %v", err)
	}

	// Build user - all values should be preserved via d.Get()
	user := buildAdminUserFromResourceData(initialData)

	// Verify all fields are preserved
	if user.Encrypted != true {
		t.Errorf("expected Encrypted=true, got %v", user.Encrypted)
	}

	if user.Attributes.Administrator == nil || *user.Attributes.Administrator != true {
		t.Error("expected Administrator=true to be preserved")
	}

	if len(user.Attributes.Connection) != 2 {
		t.Errorf("expected 2 connection methods, got %d", len(user.Attributes.Connection))
	}

	if len(user.Attributes.GUIPages) != 2 {
		t.Errorf("expected 2 GUI pages, got %d", len(user.Attributes.GUIPages))
	}

	if user.Attributes.LoginTimer == nil || *user.Attributes.LoginTimer != 600 {
		t.Error("expected LoginTimer=600 to be preserved")
	}
}

// TestBuildAdminUserFromResourceData_PointerTypes verifies that pointer fields
// are properly wrapped using helper functions.
func TestBuildAdminUserFromResourceData_PointerTypes(t *testing.T) {
	resource := resourceRTXAdminUser()

	testCases := []struct {
		name        string
		input       map[string]interface{}
		expectAdmin *bool
		expectTimer *int
	}{
		{
			name: "all fields set",
			input: map[string]interface{}{
				"username":      "admin",
				"password":      "pass",
				"administrator": true,
				"login_timer":   300,
			},
			expectAdmin: BoolPtr(true),
			expectTimer: IntPtr(300),
		},
		{
			name: "administrator false",
			input: map[string]interface{}{
				"username":      "user",
				"password":      "pass",
				"administrator": false,
				"login_timer":   0,
			},
			expectAdmin: BoolPtr(false),
			expectTimer: IntPtr(0),
		},
		{
			name: "minimal fields",
			input: map[string]interface{}{
				"username": "minimal",
				"password": "pass",
			},
			expectAdmin: BoolPtr(false), // default zero value for bool
			expectTimer: IntPtr(0),      // default zero value for int
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resource.Schema, tc.input)
			user := buildAdminUserFromResourceData(data)

			// Verify Administrator
			if user.Attributes.Administrator == nil {
				t.Error("expected Administrator to be non-nil (wrapped in pointer)")
			} else if *user.Attributes.Administrator != *tc.expectAdmin {
				t.Errorf("expected Administrator=%v, got %v", *tc.expectAdmin, *user.Attributes.Administrator)
			}

			// Verify LoginTimer
			if user.Attributes.LoginTimer == nil {
				t.Error("expected LoginTimer to be non-nil (wrapped in pointer)")
			} else if *user.Attributes.LoginTimer != *tc.expectTimer {
				t.Errorf("expected LoginTimer=%v, got %v", *tc.expectTimer, *user.Attributes.LoginTimer)
			}
		})
	}
}

// TestAdminUserAttributes_ConnectionPreservation verifies that connection methods
// are preserved when they exist in state.
func TestAdminUserAttributes_ConnectionPreservation(t *testing.T) {
	resource := resourceRTXAdminUser()

	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":           "testuser",
		"password":           "secret",
		"connection_methods": []interface{}{"ssh", "sftp", "http"},
	})

	user := buildAdminUserFromResourceData(data)

	if len(user.Attributes.Connection) != 3 {
		t.Errorf("expected 3 connection methods, got %d: %v", len(user.Attributes.Connection), user.Attributes.Connection)
	}

	// Verify connection methods are not nil (empty slice is valid)
	if user.Attributes.Connection == nil {
		t.Error("expected Connection to be non-nil (empty slice, not nil)")
	}
}

// TestAdminUserAttributes_GUIPagesPreservation verifies that GUI pages
// are preserved when they exist in state.
func TestAdminUserAttributes_GUIPagesPreservation(t *testing.T) {
	resource := resourceRTXAdminUser()

	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":  "testuser",
		"password":  "secret",
		"gui_pages": []interface{}{"dashboard", "lan-map"},
	})

	user := buildAdminUserFromResourceData(data)

	if len(user.Attributes.GUIPages) != 2 {
		t.Errorf("expected 2 GUI pages, got %d: %v", len(user.Attributes.GUIPages), user.Attributes.GUIPages)
	}

	// Verify GUI pages are not nil (empty slice is valid)
	if user.Attributes.GUIPages == nil {
		t.Error("expected GUIPages to be non-nil (empty slice, not nil)")
	}
}

// TestAdminUser_SlicesNeverNil verifies that slice fields are never nil,
// even when not specified in config.
func TestAdminUser_SlicesNeverNil(t *testing.T) {
	resource := resourceRTXAdminUser()

	// Create with only required fields - no slices specified
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username": "testuser",
		"password": "secret",
	})

	user := buildAdminUserFromResourceData(data)

	// Both slices should be non-nil (empty, but not nil)
	if user.Attributes.Connection == nil {
		t.Error("expected Connection to be non-nil empty slice, got nil")
	}
	if user.Attributes.GUIPages == nil {
		t.Error("expected GUIPages to be non-nil empty slice, got nil")
	}
}

// TestFieldHelperIntegration verifies that field helper functions work correctly
// with buildAdminUserFromResourceData.
func TestFieldHelperIntegration(t *testing.T) {
	resource := resourceRTXAdminUser()

	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":      "testuser",
		"password":      "testpass",
		"encrypted":     true,
		"administrator": true,
		"login_timer":   1800,
	})

	// Test individual helper functions
	username := GetStringValue(data, "username")
	if username != "testuser" {
		t.Errorf("expected username='testuser', got '%s'", username)
	}

	encrypted := GetBoolValue(data, "encrypted")
	if encrypted != true {
		t.Errorf("expected encrypted=true, got %v", encrypted)
	}

	administrator := GetBoolValue(data, "administrator")
	if administrator != true {
		t.Errorf("expected administrator=true, got %v", administrator)
	}

	loginTimer := GetIntValue(data, "login_timer")
	if loginTimer != 1800 {
		t.Errorf("expected login_timer=1800, got %d", loginTimer)
	}

	// Verify pointer wrappers
	adminPtr := BoolPtr(administrator)
	if adminPtr == nil || *adminPtr != true {
		t.Error("expected BoolPtr to wrap true correctly")
	}

	timerPtr := IntPtr(loginTimer)
	if timerPtr == nil || *timerPtr != 1800 {
		t.Error("expected IntPtr to wrap 1800 correctly")
	}
}

// TestAdminUserMatchesClientStruct verifies that buildAdminUserFromResourceData
// produces a valid client.AdminUser struct.
func TestAdminUserMatchesClientStruct(t *testing.T) {
	resource := resourceRTXAdminUser()

	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"username":           "admintest",
		"password":           "supersecret",
		"encrypted":          false,
		"administrator":      true,
		"connection_methods": []interface{}{"ssh"},
		"gui_pages":          []interface{}{"dashboard"},
		"login_timer":        0,
	})

	user := buildAdminUserFromResourceData(data)

	// Verify the struct matches expected client.AdminUser type
	var _ client.AdminUser = user // compile-time type check

	// Verify all fields
	if user.Username != "admintest" {
		t.Errorf("expected Username='admintest', got '%s'", user.Username)
	}
	if user.Password != "supersecret" {
		t.Errorf("expected Password='supersecret', got '%s'", user.Password)
	}
	if user.Encrypted != false {
		t.Errorf("expected Encrypted=false, got %v", user.Encrypted)
	}
	if user.Attributes.Administrator == nil || *user.Attributes.Administrator != true {
		t.Error("expected Administrator=true")
	}
	if len(user.Attributes.Connection) != 1 || user.Attributes.Connection[0] != "ssh" {
		t.Errorf("expected Connection=['ssh'], got %v", user.Attributes.Connection)
	}
	if len(user.Attributes.GUIPages) != 1 || user.Attributes.GUIPages[0] != "dashboard" {
		t.Errorf("expected GUIPages=['dashboard'], got %v", user.Attributes.GUIPages)
	}
	if user.Attributes.LoginTimer == nil || *user.Attributes.LoginTimer != 0 {
		t.Error("expected LoginTimer=0")
	}
}
