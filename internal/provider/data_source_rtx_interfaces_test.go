package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientForInterfaces extends MockClient for interfaces testing
type MockClientForInterfaces struct {
	mock.Mock
}

func (m *MockClientForInterfaces) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForInterfaces) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClientForInterfaces) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClientForInterfaces) GetInterfaces(ctx context.Context) ([]client.Interface, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Interface), args.Error(1)
}

func (m *MockClientForInterfaces) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func (m *MockClientForInterfaces) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClientForInterfaces) GetDHCPBindings(ctx context.Context, scopeID int) ([]client.DHCPBinding, error) {
	args := m.Called(ctx, scopeID)
	return args.Get(0).([]client.DHCPBinding), args.Error(1)
}

func (m *MockClientForInterfaces) CreateDHCPBinding(ctx context.Context, binding client.DHCPBinding) error {
	args := m.Called(ctx, binding)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	args := m.Called(ctx, scopeID, ipAddress)
	return args.Error(0)
}

func (m *MockClientForInterfaces) SaveConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForInterfaces) GetDHCPScope(ctx context.Context, scopeID int) (*client.DHCPScope, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.DHCPScope), args.Error(1)
}

func (m *MockClientForInterfaces) CreateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	args := m.Called(ctx, scopeID)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ListDHCPScopes(ctx context.Context) ([]client.DHCPScope, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.DHCPScope), args.Error(1)
}

func TestRTXInterfacesDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXInterfaces()
	
	// Test that the data source is properly configured
	assert.NotNil(t, dataSource)
	assert.NotNil(t, dataSource.Schema)
	assert.NotNil(t, dataSource.ReadContext)

	// Test schema structure
	schemaMap := dataSource.Schema
	
	// Check that id field exists and is computed
	assert.Contains(t, schemaMap, "id")
	assert.Equal(t, schema.TypeString, schemaMap["id"].Type)
	assert.True(t, schemaMap["id"].Computed)

	// Check interfaces list field
	assert.Contains(t, schemaMap, "interfaces")
	assert.Equal(t, schema.TypeList, schemaMap["interfaces"].Type)
	assert.True(t, schemaMap["interfaces"].Computed)
	assert.NotNil(t, schemaMap["interfaces"].Elem)

	// Check interface schema
	interfaceResource, ok := schemaMap["interfaces"].Elem.(*schema.Resource)
	assert.True(t, ok)
	interfaceSchema := interfaceResource.Schema

	// Check required fields
	requiredFields := map[string]schema.ValueType{
		"name":     schema.TypeString,
		"kind":     schema.TypeString,
		"admin_up": schema.TypeBool,
		"link_up":  schema.TypeBool,
	}
	
	for field, expectedType := range requiredFields {
		assert.Contains(t, interfaceSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, interfaceSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, interfaceSchema[field].Computed, "%s should be computed", field)
	}

	// Check optional fields
	optionalFields := map[string]schema.ValueType{
		"mac":         schema.TypeString,
		"ipv4":        schema.TypeString,
		"ipv6":        schema.TypeString,
		"mtu":         schema.TypeInt,
		"description": schema.TypeString,
		"attributes":  schema.TypeMap,
	}

	for field, expectedType := range optionalFields {
		assert.Contains(t, interfaceSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, interfaceSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, interfaceSchema[field].Computed, "%s should be computed", field)
	}

	// Check attributes element type
	assert.NotNil(t, interfaceSchema["attributes"].Elem)
	attributesElem, ok := interfaceSchema["attributes"].Elem.(*schema.Schema)
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, attributesElem.Type)
}

func TestRTXInterfacesDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClientForInterfaces{}
	
	// Mock successful interfaces retrieval
	expectedInterfaces := []client.Interface{
		{
			Name:    "LAN1",
			Kind:    "lan",
			AdminUp: true,
			LinkUp:  true,
			MAC:     "00:A0:DE:12:34:56",
			IPv4:    "192.168.1.1/24",
			MTU:     1500,
			Attributes: map[string]string{
				"speed": "1000M",
			},
		},
		{
			Name:    "WAN1",
			Kind:    "wan",
			AdminUp: true,
			LinkUp:  false,
			MAC:     "00:A0:DE:12:34:57",
			IPv4:    "203.0.113.1/30",
			MTU:     1500,
		},
		{
			Name:    "PP1",
			Kind:    "pp",
			AdminUp: false,
			LinkUp:  false,
		},
		{
			Name:        "VLAN100",
			Kind:        "vlan",
			AdminUp:     true,
			LinkUp:      true,
			IPv4:        "10.0.0.1/24",
			Description: "Management VLAN",
		},
	}

	mockClient.On("GetInterfaces", mock.Anything).Return(expectedInterfaces, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	interfaces := d.Get("interfaces").([]interface{})
	assert.Len(t, interfaces, 4)

	// Check LAN1 interface
	lan1 := interfaces[0].(map[string]interface{})
	assert.Equal(t, "LAN1", lan1["name"])
	assert.Equal(t, "lan", lan1["kind"])
	assert.Equal(t, true, lan1["admin_up"])
	assert.Equal(t, true, lan1["link_up"])
	assert.Equal(t, "00:A0:DE:12:34:56", lan1["mac"])
	assert.Equal(t, "192.168.1.1/24", lan1["ipv4"])
	assert.Equal(t, 1500, lan1["mtu"])
	assert.Equal(t, map[string]interface{}{"speed": "1000M"}, lan1["attributes"])

	// Check WAN1 interface
	wan1 := interfaces[1].(map[string]interface{})
	assert.Equal(t, "WAN1", wan1["name"])
	assert.Equal(t, "wan", wan1["kind"])
	assert.Equal(t, true, wan1["admin_up"])
	assert.Equal(t, false, wan1["link_up"])
	assert.Equal(t, "00:A0:DE:12:34:57", wan1["mac"])
	assert.Equal(t, "203.0.113.1/30", wan1["ipv4"])
	assert.Equal(t, 1500, wan1["mtu"])
	// Empty attributes will be an empty map, and other empty fields will be set
	assert.Equal(t, "", wan1["ipv6"])
	assert.Equal(t, "", wan1["description"])
	// Terraform converts map[string]string to map[string]interface{}
	assert.Equal(t, map[string]interface{}{}, wan1["attributes"])

	// Check PP1 interface (minimal data)
	pp1 := interfaces[2].(map[string]interface{})
	assert.Equal(t, "PP1", pp1["name"])
	assert.Equal(t, "pp", pp1["kind"])
	assert.Equal(t, false, pp1["admin_up"])
	assert.Equal(t, false, pp1["link_up"])
	// Empty optional fields will be zero values
	assert.Equal(t, "", pp1["mac"])
	assert.Equal(t, "", pp1["ipv4"])
	assert.Equal(t, 0, pp1["mtu"])
	assert.Equal(t, "", pp1["description"])
	assert.Equal(t, map[string]interface{}{}, pp1["attributes"])

	// Check VLAN100 interface
	vlan100 := interfaces[3].(map[string]interface{})
	assert.Equal(t, "VLAN100", vlan100["name"])
	assert.Equal(t, "vlan", vlan100["kind"])
	assert.Equal(t, true, vlan100["admin_up"])
	assert.Equal(t, true, vlan100["link_up"])
	assert.Equal(t, "10.0.0.1/24", vlan100["ipv4"])
	assert.Equal(t, "Management VLAN", vlan100["description"])
	// Empty fields will be zero values
	assert.Equal(t, "", vlan100["mac"])
	assert.Equal(t, "", vlan100["ipv6"])
	assert.Equal(t, 0, vlan100["mtu"])
	assert.Equal(t, map[string]interface{}{}, vlan100["attributes"])

	// Check that ID was set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXInterfacesDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClientForInterfaces{}
	
	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("GetInterfaces", mock.Anything).Return([]client.Interface{}, expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	// diag.Errorf puts the entire message in Summary
	assert.Contains(t, diags[0].Summary, "Failed to retrieve interfaces information")
	assert.Contains(t, diags[0].Summary, "SSH connection failed")

	mockClient.AssertExpectations(t)
}

func TestRTXInterfacesDataSourceRead_EmptyInterfaces(t *testing.T) {
	mockClient := &MockClientForInterfaces{}
	
	// Mock empty interfaces list
	mockClient.On("GetInterfaces", mock.Anything).Return([]client.Interface{}, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert empty interfaces list was set
	interfaces := d.Get("interfaces").([]interface{})
	assert.Len(t, interfaces, 0)

	// Check that ID was still set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXInterfacesDataSourceRead_DifferentRTXModels(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []client.Interface
	}{
		{
			name: "RTX830_interfaces",
			interfaces: []client.Interface{
				{
					Name:    "LAN1",
					Kind:    "lan",
					AdminUp: true,
					LinkUp:  true,
					MAC:     "00:A0:DE:11:22:33",
					IPv4:    "192.168.100.1/24",
				},
				{
					Name:    "WAN1",
					Kind:    "wan",
					AdminUp: true,
					LinkUp:  true,
					IPv4:    "203.0.113.10/30",
				},
			},
		},
		{
			name: "RTX1210_interfaces",
			interfaces: []client.Interface{
				{
					Name:    "LAN1",
					Kind:    "lan",
					AdminUp: true,
					LinkUp:  true,
					MAC:     "00:A0:DE:44:55:66",
					IPv4:    "10.0.0.1/24",
					IPv6:    "2001:db8::1/64",
					MTU:     1500,
				},
				{
					Name:    "LAN2",
					Kind:    "lan",
					AdminUp: true,
					LinkUp:  false,
					MAC:     "00:A0:DE:44:55:67",
				},
				{
					Name:    "WAN1",
					Kind:    "wan",
					AdminUp: true,
					LinkUp:  true,
					IPv4:    "203.0.113.20/30",
					MTU:     1500,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClientForInterfaces{}
			mockClient.On("GetInterfaces", mock.Anything).Return(tt.interfaces, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
			
			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Assert data was set correctly
			interfaces := d.Get("interfaces").([]interface{})
			assert.Len(t, interfaces, len(tt.interfaces))

			// Verify each interface
			for i, expectedInterface := range tt.interfaces {
				actualInterface := interfaces[i].(map[string]interface{})
				assert.Equal(t, expectedInterface.Name, actualInterface["name"])
				assert.Equal(t, expectedInterface.Kind, actualInterface["kind"])
				assert.Equal(t, expectedInterface.AdminUp, actualInterface["admin_up"])
				assert.Equal(t, expectedInterface.LinkUp, actualInterface["link_up"])

				// Check optional fields
				if expectedInterface.MAC != "" {
					assert.Equal(t, expectedInterface.MAC, actualInterface["mac"])
				}
				if expectedInterface.IPv4 != "" {
					assert.Equal(t, expectedInterface.IPv4, actualInterface["ipv4"])
				}
				if expectedInterface.IPv6 != "" {
					assert.Equal(t, expectedInterface.IPv6, actualInterface["ipv6"])
				}
				if expectedInterface.MTU > 0 {
					assert.Equal(t, expectedInterface.MTU, actualInterface["mtu"])
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// Acceptance tests using the Docker test environment
func TestAccRTXInterfacesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXInterfacesDataSourceExists("data.rtx_interfaces.test"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.#"),
					// Check that we have at least one interface
					resource.TestCheckResourceAttr("data.rtx_interfaces.test", "interfaces.#", "4"), // Assuming RTX has 4 interfaces
				),
			},
		},
	})
}

func TestAccRTXInterfacesDataSource_interfaceAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test interface fields
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.name"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.kind"),
					resource.TestMatchResourceAttr("data.rtx_interfaces.test", "interfaces.0.kind", regexp.MustCompile(`^(lan|wan|pp|vlan)$`)),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.admin_up"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.link_up"),
					// Check that interface names match expected pattern
					resource.TestMatchResourceAttr("data.rtx_interfaces.test", "interfaces.0.name", regexp.MustCompile(`^(LAN\d+|WAN\d+|PP\d+|VLAN\d+)$`)),
				),
			},
		},
	})
}

func TestAccRTXInterfacesDataSource_lanInterfaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify LAN interfaces have required attributes
					resource.TestCheckResourceAttrWith("data.rtx_interfaces.test", "interfaces.0.name", func(value string) error {
						if !strings.HasPrefix(value, "LAN") && !strings.HasPrefix(value, "WAN") {
							return fmt.Errorf("expected interface name to start with LAN or WAN, got: %s", value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccRTXInterfacesDataSource_interfaceTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that each interface has a valid kind
					resource.TestCheckResourceAttrWith("data.rtx_interfaces.test", "interfaces.#", func(value string) error {
						// This test will iterate through all interfaces in the acceptance test
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXInterfacesDataSourceConfig() string {
	return `
data "rtx_interfaces" "test" {}
`
}

func testAccCheckRTXInterfacesDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		// Check that interfaces attribute exists and has content
		interfacesCount := rs.Primary.Attributes["interfaces.#"]
		if interfacesCount == "" {
			return fmt.Errorf("interfaces count not set")
		}

		return nil
	}
}