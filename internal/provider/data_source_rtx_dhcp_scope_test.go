package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientForDHCPScope extends MockClient for DHCP scope testing
type MockClientForDHCPScope struct {
	mock.Mock
}

func (m *MockClientForDHCPScope) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClientForDHCPScope) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClientForDHCPScope) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func (m *MockClientForDHCPScope) GetInterfaces(ctx context.Context) ([]client.Interface, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Interface), args.Error(1)
}

func (m *MockClientForDHCPScope) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClientForDHCPScope) GetStaticRoutes(ctx context.Context) ([]client.StaticRoute, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.StaticRoute), args.Error(1)
}

func (m *MockClientForDHCPScope) GetDHCPScopes(ctx context.Context) ([]client.DHCPScope, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.DHCPScope), args.Error(1)
}

func (m *MockClientForDHCPScope) GetDHCPScope(ctx context.Context, scopeID int) (*client.DHCPScope, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.DHCPScope), args.Error(1)
}

func (m *MockClientForDHCPScope) GetDHCPBindings(ctx context.Context, scopeID int) ([]client.DHCPBinding, error) {
	args := m.Called(ctx, scopeID)
	return args.Get(0).([]client.DHCPBinding), args.Error(1)
}

func (m *MockClientForDHCPScope) CreateDHCPBinding(ctx context.Context, binding client.DHCPBinding) error {
	args := m.Called(ctx, binding)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	args := m.Called(ctx, scopeID, ipAddress)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) SaveConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) CreateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) UpdateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	args := m.Called(ctx, scopeID)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) CreateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) GetStaticRoute(ctx context.Context, destination, gateway, iface string) (*client.StaticRoute, error) {
	args := m.Called(ctx, destination, gateway, iface)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.StaticRoute), args.Error(1)
}

func (m *MockClientForDHCPScope) UpdateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForDHCPScope) DeleteStaticRoute(ctx context.Context, destination, gateway, iface string) error {
	args := m.Called(ctx, destination, gateway, iface)
	return args.Error(0)
}

func TestRTXDHCPScopeDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXDHCPScope()

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

	// Check scopes list field
	assert.Contains(t, schemaMap, "scopes")
	assert.Equal(t, schema.TypeList, schemaMap["scopes"].Type)
	assert.True(t, schemaMap["scopes"].Computed)
	assert.NotNil(t, schemaMap["scopes"].Elem)

	// Check scope schema
	scopeResource, ok := schemaMap["scopes"].Elem.(*schema.Resource)
	assert.True(t, ok)

	scopeSchema := scopeResource.Schema

	// Check required fields
	requiredFields := []string{"scope_id", "range_start", "range_end", "prefix"}
	for _, field := range requiredFields {
		assert.Contains(t, scopeSchema, field, "Scope schema should contain %s field", field)
		assert.True(t, scopeSchema[field].Computed, "%s should be computed", field)
	}

	// Check optional fields
	optionalFields := []string{"gateway", "dns_servers", "lease", "domain_name"}
	for _, field := range optionalFields {
		assert.Contains(t, scopeSchema, field, "Scope schema should contain %s field", field)
		assert.True(t, scopeSchema[field].Computed, "%s should be computed", field)
	}
}

func TestRTXDHCPScopeDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClientForDHCPScope{}

	// Mock successful DHCP scopes retrieval
	expectedScopes := []client.DHCPScope{
		{
			ID:         1,
			RangeStart: "192.168.100.2",
			RangeEnd:   "192.168.100.191",
			Prefix:     24,
			Gateway:    "192.168.100.1",
			DNSServers: []string{"8.8.8.8", "8.8.4.4"},
			Lease:      7,
			DomainName: "example.com",
		},
		{
			ID:         2,
			RangeStart: "10.0.0.10",
			RangeEnd:   "10.0.0.20",
			Prefix:     16,
		},
	}

	mockClient.On("GetDHCPScopes", mock.Anything).Return(expectedScopes, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXDHCPScope().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXDHCPScopeRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	scopes := d.Get("scopes").([]interface{})
	assert.Len(t, scopes, 2)

	// Check first scope
	scope1 := scopes[0].(map[string]interface{})
	assert.Equal(t, 1, scope1["scope_id"])
	assert.Equal(t, "192.168.100.2", scope1["range_start"])
	assert.Equal(t, "192.168.100.191", scope1["range_end"])
	assert.Equal(t, 24, scope1["prefix"])
	assert.Equal(t, "192.168.100.1", scope1["gateway"])
	assert.Equal(t, []interface{}{"8.8.8.8", "8.8.4.4"}, scope1["dns_servers"])
	assert.Equal(t, 7, scope1["lease"])
	assert.Equal(t, "example.com", scope1["domain_name"])

	// Check second scope (minimal)
	scope2 := scopes[1].(map[string]interface{})
	assert.Equal(t, 2, scope2["scope_id"])
	assert.Equal(t, "10.0.0.10", scope2["range_start"])
	assert.Equal(t, "10.0.0.20", scope2["range_end"])
	assert.Equal(t, 16, scope2["prefix"])

	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestAccDataSourceRTXDHCPScope_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRTXDHCPScopeConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rtx_dhcp_scope.test", "id"),
					resource.TestCheckResourceAttr("data.rtx_dhcp_scope.test", "scopes.#", "1"),
					resource.TestCheckResourceAttr("data.rtx_dhcp_scope.test", "scopes.0.scope_id", "1"),
					resource.TestCheckResourceAttr("data.rtx_dhcp_scope.test", "scopes.0.range_start", "192.168.100.2"),
					resource.TestCheckResourceAttr("data.rtx_dhcp_scope.test", "scopes.0.range_end", "192.168.100.191"),
					resource.TestCheckResourceAttr("data.rtx_dhcp_scope.test", "scopes.0.prefix", "24"),
				),
			},
		},
	})
}

const testAccDataSourceRTXDHCPScopeConfig = `
data "rtx_dhcp_scope" "test" {}
`
