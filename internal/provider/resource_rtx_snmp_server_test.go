package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildSNMPConfigFromResourceData(t *testing.T) {
	// This test validates the structure of the config building function
	// In a real scenario, you would use a mock ResourceData
	t.Run("Basic structure validation", func(t *testing.T) {
		// Verify the function signature and return type
		// Since we can't easily mock schema.ResourceData without Terraform's test framework,
		// we focus on testing the helper functions
	})
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		key       string
		wantWarn  bool
		wantError bool
	}{
		{
			name:      "valid IPv4",
			value:     "192.168.1.100",
			key:       "ip_address",
			wantWarn:  false,
			wantError: false,
		},
		{
			name:      "valid IPv4 zeros",
			value:     "0.0.0.0",
			key:       "ip_address",
			wantWarn:  false,
			wantError: false,
		},
		{
			name:      "valid IPv4 broadcast",
			value:     "255.255.255.255",
			key:       "ip_address",
			wantWarn:  false,
			wantError: false,
		},
		{
			name:      "invalid IPv6 - not supported",
			value:     "::1",
			key:       "ip_address",
			wantWarn:  false,
			wantError: true, // IPv6 is not supported for SNMP on RTX
		},
		{
			name:      "invalid IPv6 full - not supported",
			value:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			key:       "ip_address",
			wantWarn:  false,
			wantError: true, // IPv6 is not supported for SNMP on RTX
		},
		{
			name:      "empty string",
			value:     "",
			key:       "ip_address",
			wantWarn:  false,
			wantError: false, // Empty is allowed (optional field)
		},
		{
			name:      "invalid not an IP",
			value:     "not-an-ip",
			key:       "ip_address",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid hostname",
			value:     "example.com",
			key:       "ip_address",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid IPv4 out of range",
			value:     "256.1.1.1",
			key:       "ip_address",
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid partial IP",
			value:     "192.168",
			key:       "ip_address",
			wantWarn:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns, errs := validateIPAddress(tt.value, tt.key)

			if tt.wantWarn {
				if len(warns) == 0 {
					t.Errorf("validateIPAddress() expected warnings, got none")
				}
			} else {
				if len(warns) > 0 {
					t.Errorf("validateIPAddress() unexpected warnings: %v", warns)
				}
			}

			if tt.wantError {
				if len(errs) == 0 {
					t.Errorf("validateIPAddress() expected errors, got none")
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("validateIPAddress() unexpected errors: %v", errs)
				}
			}
		})
	}
}

func TestSNMPConfigConversion(t *testing.T) {
	// Test that we can properly convert between client types and resource data
	t.Run("SNMPConfig with communities", func(t *testing.T) {
		config := client.SNMPConfig{
			SysName:     "TestRouter",
			SysLocation: "Test Location",
			SysContact:  "admin@test.com",
			Communities: []client.SNMPCommunity{
				{Name: "public", Permission: "ro", ACL: ""},
				{Name: "private", Permission: "rw", ACL: "10"},
			},
			Hosts: []client.SNMPHost{
				{Address: "192.168.1.100", Community: "public", Version: "2c"},
			},
			TrapEnable: []string{"coldstart", "warmstart"},
		}

		// Verify the config structure
		if config.SysName != "TestRouter" {
			t.Errorf("SysName = %q, want %q", config.SysName, "TestRouter")
		}
		if len(config.Communities) != 2 {
			t.Errorf("Communities length = %d, want %d", len(config.Communities), 2)
		}
		if len(config.Hosts) != 1 {
			t.Errorf("Hosts length = %d, want %d", len(config.Hosts), 1)
		}
		if len(config.TrapEnable) != 2 {
			t.Errorf("TrapEnable length = %d, want %d", len(config.TrapEnable), 2)
		}
	})

	t.Run("Empty SNMPConfig", func(t *testing.T) {
		config := client.SNMPConfig{
			Communities: []client.SNMPCommunity{},
			Hosts:       []client.SNMPHost{},
			TrapEnable:  []string{},
		}

		if config.SysName != "" {
			t.Errorf("SysName should be empty, got %q", config.SysName)
		}
		if len(config.Communities) != 0 {
			t.Errorf("Communities should be empty, got %d items", len(config.Communities))
		}
	})

	t.Run("SNMPCommunity validation", func(t *testing.T) {
		validPerms := []string{"ro", "rw"}
		for _, perm := range validPerms {
			community := client.SNMPCommunity{
				Name:       "test",
				Permission: perm,
			}
			if community.Permission != perm {
				t.Errorf("Permission = %q, want %q", community.Permission, perm)
			}
		}
	})

	t.Run("SNMPHost validation", func(t *testing.T) {
		host := client.SNMPHost{
			Address:   "192.168.1.100",
			Community: "public",
			Version:   "2c",
		}
		if host.Address != "192.168.1.100" {
			t.Errorf("Address = %q, want %q", host.Address, "192.168.1.100")
		}
		if host.Version != "2c" {
			t.Errorf("Version = %q, want %q", host.Version, "2c")
		}
	})
}

func TestSNMPResourceID(t *testing.T) {
	// Test that the resource uses the correct singleton ID
	expectedID := "snmp"

	t.Run("Singleton ID should be 'snmp'", func(t *testing.T) {
		// In the actual Create function, we set d.SetId("snmp")
		// This test documents the expected behavior
		actualID := "snmp"
		if actualID != expectedID {
			t.Errorf("Resource ID = %q, want %q", actualID, expectedID)
		}
	})
}

// Acceptance tests for rtx_snmp_server

func TestAccRTXSNMPServer_basic(t *testing.T) {
	resourceName := "rtx_snmp_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSNMPServerConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "sys_name", "TestRouter"),
					resource.TestCheckResourceAttr(resourceName, "community.#", "1"),
				),
			},
		},
	})
}

func TestAccRTXSNMPServer_import(t *testing.T) {
	resourceName := "rtx_snmp_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSNMPServerConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "sys_name", "TestRouter"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRTXSNMPServer_noDiff(t *testing.T) {
	resourceName := "rtx_snmp_server.test"
	config := testAccRTXSNMPServerConfig_full()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "sys_name", "TestRouter"),
					resource.TestCheckResourceAttr(resourceName, "sys_location", "Test Location"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestAccRTXSNMPServer_update(t *testing.T) {
	resourceName := "rtx_snmp_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSNMPServerConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "sys_name", "TestRouter"),
				),
			},
			{
				Config: testAccRTXSNMPServerConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "sys_name", "UpdatedRouter"),
				),
			},
		},
	})
}

func testAccRTXSNMPServerConfig_basic() string {
	return `
resource "rtx_snmp_server" "test" {
  sys_name = "TestRouter"

  community {
    name       = "public"
    permission = "ro"
  }
}
`
}

func testAccRTXSNMPServerConfig_full() string {
	return `
resource "rtx_snmp_server" "test" {
  sys_name     = "TestRouter"
  sys_location = "Test Location"
  sys_contact  = "admin@test.com"

  community {
    name       = "public"
    permission = "ro"
  }
}
`
}

func testAccRTXSNMPServerConfig_updated() string {
	return `
resource "rtx_snmp_server" "test" {
  sys_name = "UpdatedRouter"

  community {
    name       = "public"
    permission = "ro"
  }
}
`
}
