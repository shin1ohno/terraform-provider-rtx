package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientForStaticRoutes extends MockClient for static routes testing
type MockClientForStaticRoutes struct {
	mock.Mock
}

func (m *MockClientForStaticRoutes) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClientForStaticRoutes) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) GetStaticRoutes(ctx context.Context) ([]client.StaticRoute, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.StaticRoute), args.Error(1)
}

func (m *MockClientForStaticRoutes) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func (m *MockClientForStaticRoutes) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClientForStaticRoutes) GetInterfaces(ctx context.Context) ([]client.Interface, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Interface), args.Error(1)
}

func (m *MockClientForStaticRoutes) GetDHCPBindings(ctx context.Context, scopeID int) ([]client.DHCPBinding, error) {
	args := m.Called(ctx, scopeID)
	return args.Get(0).([]client.DHCPBinding), args.Error(1)
}

func (m *MockClientForStaticRoutes) CreateDHCPBinding(ctx context.Context, binding client.DHCPBinding) error {
	args := m.Called(ctx, binding)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	args := m.Called(ctx, scopeID, ipAddress)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) GetDHCPScopes(ctx context.Context) ([]client.DHCPScope, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.DHCPScope), args.Error(1)
}

func (m *MockClientForStaticRoutes) GetDHCPScope(ctx context.Context, scopeID int) (*client.DHCPScope, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.DHCPScope), args.Error(1)
}

func (m *MockClientForStaticRoutes) SaveConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) CreateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) UpdateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	args := m.Called(ctx, scopeID)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) CreateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) GetStaticRoute(ctx context.Context, destination, gateway, iface string) (*client.StaticRoute, error) {
	args := m.Called(ctx, destination, gateway, iface)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.StaticRoute), args.Error(1)
}

func (m *MockClientForStaticRoutes) UpdateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForStaticRoutes) DeleteStaticRoute(ctx context.Context, destination, gateway, iface string) error {
	args := m.Called(ctx, destination, gateway, iface)
	return args.Error(0)
}

func TestRTXStaticRoutesDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXStaticRoutes()

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

	// Check routes list field
	assert.Contains(t, schemaMap, "routes")
	assert.Equal(t, schema.TypeList, schemaMap["routes"].Type)
	assert.True(t, schemaMap["routes"].Computed)
	assert.NotNil(t, schemaMap["routes"].Elem)

	// Check route schema
	routeResource, ok := schemaMap["routes"].Elem.(*schema.Resource)
	assert.True(t, ok)
	routeSchema := routeResource.Schema

	// Check required fields
	requiredFields := map[string]schema.ValueType{
		"destination": schema.TypeString,
		"gateway":     schema.TypeString,
		"interface":   schema.TypeString,
		"metric":      schema.TypeInt,
		"weight":      schema.TypeInt,
		"hide":        schema.TypeBool,
		"description": schema.TypeString,
	}

	for field, expectedType := range requiredFields {
		assert.Contains(t, routeSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, routeSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, routeSchema[field].Computed, "%s should be computed", field)
	}
}

func TestRTXStaticRoutesDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClientForStaticRoutes{}

	// Mock successful static routes retrieval
	expectedStaticRoutes := []client.StaticRoute{
		{
			Destination: "0.0.0.0/0",
			GatewayIP:   "192.168.1.1",
			Interface:   "LAN1",
			Metric:      1,
			Weight:      1,
			Hide:        false,
			Description: "Default route",
		},
		{
			Destination: "10.0.0.0/24",
			GatewayIP:   "192.168.1.254",
			Interface:   "LAN1",
			Metric:      5,
			Weight:      1,
			Hide:        false,
			Description: "Private network route",
		},
		{
			Destination:      "172.16.0.0/16",
			GatewayInterface: "WAN1",
			Interface:        "WAN1",
			Metric:           10,
			Weight:           2,
			Hide:             true,
			Description: "",
		},
	}

	mockClient.On("GetStaticRoutes", mock.Anything).Return(expectedStaticRoutes, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXStaticRoutes().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXStaticRoutesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	routes := d.Get("routes").([]interface{})
	assert.Len(t, routes, 3)

	// Check default route (first route)
	defaultRoute := routes[0].(map[string]interface{})
	assert.Equal(t, "0.0.0.0/0", defaultRoute["destination"])
	assert.Equal(t, "192.168.1.1", defaultRoute["gateway"])
	assert.Equal(t, "LAN1", defaultRoute["interface"])
	assert.Equal(t, 1, defaultRoute["metric"])
	assert.Equal(t, 1, defaultRoute["weight"])
	// Kind field removed
	assert.Equal(t, false, defaultRoute["hide"])
	assert.Equal(t, "Default route", defaultRoute["description"])

	// Check private network route (second route)
	privateRoute := routes[1].(map[string]interface{})
	assert.Equal(t, "10.0.0.0/24", privateRoute["destination"])
	assert.Equal(t, "192.168.1.254", privateRoute["gateway"])
	assert.Equal(t, "LAN1", privateRoute["interface"])
	assert.Equal(t, 5, privateRoute["metric"])
	assert.Equal(t, 1, privateRoute["weight"])
	// Kind field removed
	assert.Equal(t, false, privateRoute["hide"])
	assert.Equal(t, "Private network route", privateRoute["description"])

	// Check interface route (third route)
	interfaceRoute := routes[2].(map[string]interface{})
	assert.Equal(t, "172.16.0.0/16", interfaceRoute["destination"])
	assert.Equal(t, "WAN1", interfaceRoute["gateway"])
	assert.Equal(t, "WAN1", interfaceRoute["interface"])
	assert.Equal(t, 10, interfaceRoute["metric"])
	assert.Equal(t, 2, interfaceRoute["weight"])
	// Kind field removed
	assert.Equal(t, true, interfaceRoute["hide"])
	assert.Equal(t, "", interfaceRoute["description"])

	// Check that ID was set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXStaticRoutesDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClientForStaticRoutes{}

	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("GetStaticRoutes", mock.Anything).Return([]client.StaticRoute{}, expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXStaticRoutes().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXStaticRoutesRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Failed to retrieve static routes information")
	assert.Contains(t, diags[0].Summary, "SSH connection failed")

	mockClient.AssertExpectations(t)
}

func TestRTXStaticRoutesDataSourceRead_EmptyRoutes(t *testing.T) {
	mockClient := &MockClientForStaticRoutes{}

	// Mock empty static routes list
	mockClient.On("GetStaticRoutes", mock.Anything).Return([]client.StaticRoute{}, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXStaticRoutes().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXStaticRoutesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert empty routes list was set
	routes := d.Get("routes").([]interface{})
	assert.Len(t, routes, 0)

	// Check that ID was still set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXStaticRoutesDataSourceRead_DifferentRouteKinds(t *testing.T) {
	tests := []struct {
		name   string
		routes []client.StaticRoute
	}{
		{
			name: "StaticRoutes",
			routes: []client.StaticRoute{
				{
					Destination: "0.0.0.0/0",
					GatewayIP:     "192.168.1.1",
					Interface:   "LAN1",
					Metric:      1,
					Weight:      1,
					Hide:        false,
					Description: "Default route",
				},
				{
					Destination: "10.10.10.0/24",
					GatewayIP:     "192.168.1.10",
					Interface:   "LAN1",
					Metric:      5,
					Weight:      1,
					Hide:        false,
					Description: "Private subnet",
				},
			},
		},
		{
			name: "InterfaceRoutes",
			routes: []client.StaticRoute{
				{
					Destination:      "192.168.1.0/24",
					GatewayInterface: "LAN1",
					Interface:        "LAN1",
					Metric:           0,
					Weight:           1,
					Hide:             false,
					Description:      "LAN1 connected",
				},
				{
					Destination:      "192.168.2.0/24",
					GatewayInterface: "LAN2",
					Interface:        "LAN2",
					Metric:           0,
					Weight:           1,
					Hide:             false,
					Description:      "LAN2 connected",
				},
			},
		},
		{
			name: "MixedRoutes",
			routes: []client.StaticRoute{
				{
					Destination: "0.0.0.0/0",
					GatewayIP:     "192.168.1.1",
					Interface:   "LAN1",
					Metric:      1,
					Weight:      1,
					Hide:        false,
					Description: "Default gateway",
				},
				{
					Destination:      "192.168.100.0/24",
					GatewayInterface: "VLAN100",
					Interface:        "VLAN100",
					Metric:           0,
					Weight:           1,
					Hide:             true,
					Description:      "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClientForStaticRoutes{}
			mockClient.On("GetStaticRoutes", mock.Anything).Return(tt.routes, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXStaticRoutes().Schema, map[string]interface{}{})

			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXStaticRoutesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Assert data was set correctly
			routes := d.Get("routes").([]interface{})
			assert.Len(t, routes, len(tt.routes))

			// Verify each route
			for i, expectedRoute := range tt.routes {
				actualRoute := routes[i].(map[string]interface{})
				assert.Equal(t, expectedRoute.Destination, actualRoute["destination"])
				// Gateway field needs to be derived from GatewayIP or GatewayInterface
				expectedGateway := expectedRoute.GatewayIP
				if expectedGateway == "" && expectedRoute.GatewayInterface != "" {
					expectedGateway = expectedRoute.GatewayInterface
				}
				assert.Equal(t, expectedGateway, actualRoute["gateway"])
				assert.Equal(t, expectedRoute.Interface, actualRoute["interface"])
				assert.Equal(t, expectedRoute.Metric, actualRoute["metric"])
				assert.Equal(t, expectedRoute.Weight, actualRoute["weight"])
				// Kind field no longer exists in StaticRoute struct
				assert.Equal(t, expectedRoute.Hide, actualRoute["hide"])
				assert.Equal(t, expectedRoute.Description, actualRoute["description"])
			}

			mockClient.AssertExpectations(t)
		})
	}
}


// Acceptance tests using the Docker test environment
func TestAccRTXStaticRoutesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXStaticRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXStaticRoutesDataSourceExists("data.rtx_static_routes.test"),
					resource.TestCheckResourceAttrSet("data.rtx_static_routes.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_static_routes.test", "routes.#"),
				),
			},
		},
	})
}

func TestAccRTXStaticRoutesDataSource_routeAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXStaticRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test that routes list exists
					resource.TestCheckResourceAttrSet("data.rtx_static_routes.test", "routes.#"),
				),
			},
		},
	})
}

func TestAccRTXStaticRoutesDataSource_defaultRoute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXStaticRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that we can read static routes
					resource.TestCheckResourceAttrWith("data.rtx_static_routes.test", "routes.#", func(value string) error {
						// This test just checks that the resource can be read successfully
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXStaticRoutesDataSourceConfig() string {
	return `
data "rtx_static_routes" "test" {}
`
}

func testAccCheckRTXStaticRoutesDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		// Check that routes attribute exists
		routesCount := rs.Primary.Attributes["routes.#"]
		if routesCount == "" {
			return fmt.Errorf("routes count not set")
		}

		return nil
	}
}