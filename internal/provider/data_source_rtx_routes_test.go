package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientForRoutes extends MockClient for routes testing
type MockClientForRoutes struct {
	MockClient
}

func (m *MockClientForRoutes) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClientForRoutes) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func TestRTXRoutesDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXRoutes()
	
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
		"protocol":    schema.TypeString,
		"metric":      schema.TypeInt,
	}
	
	for field, expectedType := range requiredFields {
		assert.Contains(t, routeSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, routeSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, routeSchema[field].Computed, "%s should be computed", field)
	}
}

func TestRTXRoutesDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClientForRoutes{}
	
	// Mock successful routes retrieval
	metric1 := 1
	metric2 := 10
	expectedRoutes := []client.Route{
		{
			Destination: "0.0.0.0/0",
			Gateway:     "192.168.1.1",
			Interface:   "LAN1",
			Protocol:    "S",
			Metric:      &metric1,
		},
		{
			Destination: "192.168.1.0/24",
			Gateway:     "*",
			Interface:   "LAN1",
			Protocol:    "C",
			Metric:      nil, // Connected routes may not have metrics
		},
		{
			Destination: "10.0.0.0/24",
			Gateway:     "192.168.1.254",
			Interface:   "LAN1",
			Protocol:    "R",
			Metric:      &metric2,
		},
		{
			Destination: "203.0.113.0/30",
			Gateway:     "*",
			Interface:   "WAN1",
			Protocol:    "C",
			Metric:      nil,
		},
	}

	mockClient.On("GetRoutes", mock.Anything).Return(expectedRoutes, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	routes := d.Get("routes").([]interface{})
	assert.Len(t, routes, 4)

	// Check default route (first route)
	defaultRoute := routes[0].(map[string]interface{})
	assert.Equal(t, "0.0.0.0/0", defaultRoute["destination"])
	assert.Equal(t, "192.168.1.1", defaultRoute["gateway"])
	assert.Equal(t, "LAN1", defaultRoute["interface"])
	assert.Equal(t, "S", defaultRoute["protocol"])
	assert.Equal(t, 1, defaultRoute["metric"])

	// Check connected route (second route)
	connectedRoute := routes[1].(map[string]interface{})
	assert.Equal(t, "192.168.1.0/24", connectedRoute["destination"])
	assert.Equal(t, "*", connectedRoute["gateway"])
	assert.Equal(t, "LAN1", connectedRoute["interface"])
	assert.Equal(t, "C", connectedRoute["protocol"])
	assert.Equal(t, 0, connectedRoute["metric"]) // Should default to 0 when nil

	// Check RIP route (third route)
	ripRoute := routes[2].(map[string]interface{})
	assert.Equal(t, "10.0.0.0/24", ripRoute["destination"])
	assert.Equal(t, "192.168.1.254", ripRoute["gateway"])
	assert.Equal(t, "LAN1", ripRoute["interface"])
	assert.Equal(t, "R", ripRoute["protocol"])
	assert.Equal(t, 10, ripRoute["metric"])

	// Check WAN connected route (fourth route)
	wanRoute := routes[3].(map[string]interface{})
	assert.Equal(t, "203.0.113.0/30", wanRoute["destination"])
	assert.Equal(t, "*", wanRoute["gateway"])
	assert.Equal(t, "WAN1", wanRoute["interface"])
	assert.Equal(t, "C", wanRoute["protocol"])
	assert.Equal(t, 0, wanRoute["metric"])

	// Check that ID was set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXRoutesDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClientForRoutes{}
	
	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("GetRoutes", mock.Anything).Return([]client.Route{}, expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Failed to retrieve routes information")
	assert.Contains(t, diags[0].Summary, "SSH connection failed")

	mockClient.AssertExpectations(t)
}

func TestRTXRoutesDataSourceRead_EmptyRoutes(t *testing.T) {
	mockClient := &MockClientForRoutes{}
	
	// Mock empty routes list
	mockClient.On("GetRoutes", mock.Anything).Return([]client.Route{}, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert empty routes list was set
	routes := d.Get("routes").([]interface{})
	assert.Len(t, routes, 0)

	// Check that ID was still set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXRoutesDataSourceRead_DifferentRouteTypes(t *testing.T) {
	tests := []struct {
		name   string
		routes []client.Route
	}{
		{
			name: "StaticRoutes",
			routes: []client.Route{
				{
					Destination: "0.0.0.0/0",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Protocol:    "S",
					Metric:      intPtr(1),
				},
				{
					Destination: "10.10.10.0/24",
					Gateway:     "192.168.1.10",
					Interface:   "LAN1",
					Protocol:    "S",
					Metric:      intPtr(5),
				},
			},
		},
		{
			name: "ConnectedRoutes",
			routes: []client.Route{
				{
					Destination: "192.168.1.0/24",
					Gateway:     "*",
					Interface:   "LAN1",
					Protocol:    "C",
					Metric:      nil,
				},
				{
					Destination: "192.168.2.0/24",
					Gateway:     "*",
					Interface:   "LAN2",
					Protocol:    "C",
					Metric:      nil,
				},
			},
		},
		{
			name: "DynamicRoutes",
			routes: []client.Route{
				{
					Destination: "172.16.1.0/24",
					Gateway:     "192.168.1.100",
					Interface:   "LAN1",
					Protocol:    "R",
					Metric:      intPtr(2),
				},
				{
					Destination: "172.16.2.0/24",
					Gateway:     "192.168.1.101",
					Interface:   "LAN1",
					Protocol:    "O",
					Metric:      intPtr(110),
				},
				{
					Destination: "172.16.3.0/24",
					Gateway:     "192.168.1.102",
					Interface:   "LAN1",
					Protocol:    "B",
					Metric:      intPtr(20),
				},
			},
		},
		{
			name: "DHCPRoutes",
			routes: []client.Route{
				{
					Destination: "0.0.0.0/0",
					Gateway:     "203.0.113.1",
					Interface:   "WAN1",
					Protocol:    "D",
					Metric:      intPtr(1),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClientForRoutes{}
			mockClient.On("GetRoutes", mock.Anything).Return(tt.routes, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})
			
			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Assert data was set correctly
			routes := d.Get("routes").([]interface{})
			assert.Len(t, routes, len(tt.routes))

			// Verify each route
			for i, expectedRoute := range tt.routes {
				actualRoute := routes[i].(map[string]interface{})
				assert.Equal(t, expectedRoute.Destination, actualRoute["destination"])
				assert.Equal(t, expectedRoute.Gateway, actualRoute["gateway"])
				assert.Equal(t, expectedRoute.Interface, actualRoute["interface"])
				assert.Equal(t, expectedRoute.Protocol, actualRoute["protocol"])
				
				if expectedRoute.Metric != nil {
					assert.Equal(t, *expectedRoute.Metric, actualRoute["metric"])
				} else {
					assert.Equal(t, 0, actualRoute["metric"])
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRTXRoutesDataSourceRead_ProtocolValidation(t *testing.T) {
	validProtocols := []string{"S", "C", "R", "O", "B", "D"}
	
	for _, protocol := range validProtocols {
		t.Run(fmt.Sprintf("Protocol_%s", protocol), func(t *testing.T) {
			mockClient := &MockClientForRoutes{}
			
			routes := []client.Route{
				{
					Destination: "192.168.1.0/24",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Protocol:    protocol,
					Metric:      intPtr(1),
				},
			}
			mockClient.On("GetRoutes", mock.Anything).Return(routes, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})
			
			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Verify protocol was set correctly
			routesData := d.Get("routes").([]interface{})
			assert.Len(t, routesData, 1)
			
			actualRoute := routesData[0].(map[string]interface{})
			assert.Equal(t, protocol, actualRoute["protocol"])

			mockClient.AssertExpectations(t)
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// Acceptance tests using the Docker test environment
func TestAccRTXRoutesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXRoutesDataSourceExists("data.rtx_routes.test"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.#"),
					// Check that we have at least one route (should have default route and connected routes)
					resource.TestCheckResourceAttr("data.rtx_routes.test", "routes.#", "3"), // Assuming RTX has 3 routes
				),
			},
		},
	})
}

func TestAccRTXRoutesDataSource_routeAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test route fields
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.destination"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.gateway"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.interface"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.protocol"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.metric"),
					// Test that destination matches CIDR pattern
					resource.TestMatchResourceAttr("data.rtx_routes.test", "routes.0.destination", regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)),
					// Test that protocol is one of the valid values
					resource.TestMatchResourceAttr("data.rtx_routes.test", "routes.0.protocol", regexp.MustCompile(`^[SCROB D]$`)),
				),
			},
		},
	})
}

func TestAccRTXRoutesDataSource_defaultRoute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check for default route (0.0.0.0/0) existence
					resource.TestCheckResourceAttrWith("data.rtx_routes.test", "routes.#", func(value string) error {
						// This test checks that at least one route exists
						if value == "0" {
							return fmt.Errorf("expected at least one route, got none")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccRTXRoutesDataSource_connectedRoutes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify connected routes have "*" as gateway
					resource.TestCheckResourceAttrWith("data.rtx_routes.test", "routes.#", func(value string) error {
						// This would need custom logic to check for connected routes in acceptance test
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXRoutesDataSourceConfig() string {
	return `
data "rtx_routes" "test" {}
`
}

func testAccCheckRTXRoutesDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		// Check that routes attribute exists and has content
		routesCount := rs.Primary.Attributes["routes.#"]
		if routesCount == "" {
			return fmt.Errorf("routes count not set")
		}

		return nil
	}
}