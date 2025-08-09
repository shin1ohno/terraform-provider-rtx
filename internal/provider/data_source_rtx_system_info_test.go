package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
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

// Mock client for unit tests
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRTXSystemInfoDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXSystemInfo()
	
	// Test that the data source is properly configured
	assert.NotNil(t, dataSource)
	assert.NotNil(t, dataSource.Schema)
	assert.NotNil(t, dataSource.ReadContext)

	// Test schema structure
	schemaMap := dataSource.Schema
	
	// Check required fields
	requiredFields := []string{"model", "firmware_version", "serial_number", "mac_address", "uptime"}
	for _, field := range requiredFields {
		assert.Contains(t, schemaMap, field, "Schema should contain %s field", field)
		assert.Equal(t, schema.TypeString, schemaMap[field].Type, "%s should be of type string", field)
		assert.True(t, schemaMap[field].Computed, "%s should be computed", field)
	}

	// Check that id field exists and is computed
	assert.Contains(t, schemaMap, "id")
	assert.Equal(t, schema.TypeString, schemaMap["id"].Type)
	assert.True(t, schemaMap["id"].Computed)
}

func TestRTXSystemInfoDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClient{}
	
	// Mock successful show environment command
	expectedResponse := client.Result{
		Raw: []byte(`Model: RTX1210
Firmware Version: Rev.14.01.27
Serial Number: ABC123456789
MAC Address: 00:a0:de:12:34:56
Uptime: 15 days, 10:30:25`),
		Parsed: nil,
	}

	mockClient.On("Run", mock.Anything, mock.MatchedBy(func(cmd client.Command) bool {
		return cmd.Key == "show environment" && cmd.Payload == "show environment"
	})).Return(expectedResponse, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXSystemInfo().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXSystemInfoRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	assert.Equal(t, "RTX1210", d.Get("model").(string))
	assert.Equal(t, "Rev.14.01.27", d.Get("firmware_version").(string))
	assert.Equal(t, "ABC123456789", d.Get("serial_number").(string))
	assert.Equal(t, "00:a0:de:12:34:56", d.Get("mac_address").(string))
	assert.Equal(t, "15 days, 10:30:25", d.Get("uptime").(string))
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXSystemInfoDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClient{}
	
	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("Run", mock.Anything, mock.AnythingOfType("client.Command")).Return(client.Result{}, expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXSystemInfo().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXSystemInfoRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Failed to retrieve system information")

	mockClient.AssertExpectations(t)
}

func TestRTXSystemInfoDataSourceRead_ParseError(t *testing.T) {
	mockClient := &MockClient{}
	
	// Mock response with unparseable output
	expectedResponse := client.Result{
		Raw:    []byte("Invalid output format"),
		Parsed: nil,
	}

	mockClient.On("Run", mock.Anything, mock.AnythingOfType("client.Command")).Return(expectedResponse, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXSystemInfo().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXSystemInfoRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Failed to parse system information")

	mockClient.AssertExpectations(t)
}

// TestRTXSystemInfoDataSourceRead_CommandError removed since client.Result doesn't have Error field

// Test parsing of various RTX output formats
func TestParseSystemInfo(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedModel  string
		expectedFW     string
		expectedSerial string
		expectedMAC    string
		expectedUptime string
		shouldError    bool
	}{
		{
			name: "Standard RTX1210 format",
			output: `Model: RTX1210
Firmware Version: Rev.14.01.27
Serial Number: ABC123456789
MAC Address: 00:a0:de:12:34:56
Uptime: 15 days, 10:30:25`,
			expectedModel:  "RTX1210",
			expectedFW:     "Rev.14.01.27",
			expectedSerial: "ABC123456789",
			expectedMAC:    "00:a0:de:12:34:56",
			expectedUptime: "15 days, 10:30:25",
			shouldError:    false,
		},
		{
			name: "RTX830 format with different spacing",
			output: `Model:    RTX830
Firmware Version:  Rev.15.02.10
Serial Number:   XYZ987654321
MAC Address:    aa:bb:cc:dd:ee:ff
Uptime:   3 days, 5:15:42`,
			expectedModel:  "RTX830",
			expectedFW:     "Rev.15.02.10",
			expectedSerial: "XYZ987654321",
			expectedMAC:    "aa:bb:cc:dd:ee:ff",
			expectedUptime: "3 days, 5:15:42",
			shouldError:    false,
		},
		{
			name: "Short uptime format",
			output: `Model: RTX1200
Firmware Version: Rev.10.01.80
Serial Number: DEF456789123
MAC Address: 12:34:56:78:9a:bc
Uptime: 2:15:30`,
			expectedModel:  "RTX1200",
			expectedFW:     "Rev.10.01.80",
			expectedSerial: "DEF456789123",
			expectedMAC:    "12:34:56:78:9a:bc",
			expectedUptime: "2:15:30",
			shouldError:    false,
		},
		{
			name:        "Missing model field",
			output:      `Firmware Version: Rev.14.01.27`,
			shouldError: true,
		},
		{
			name:        "Empty output",
			output:      "",
			shouldError: true,
		},
		{
			name:        "Invalid format",
			output:      "This is not a valid RTX output",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSystemInfo(tt.output)
			
			if tt.shouldError {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedModel, result.Model)
			assert.Equal(t, tt.expectedFW, result.FirmwareVersion)
			assert.Equal(t, tt.expectedSerial, result.SerialNumber)
			assert.Equal(t, tt.expectedMAC, result.MACAddress)
			assert.Equal(t, tt.expectedUptime, result.Uptime)
		})
	}
}

// Acceptance tests using Docker test environment
func TestAccRTXSystemInfoDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSystemInfoDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXSystemInfoDataSourceExists("data.rtx_system_info.test"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "model"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "firmware_version"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "mac_address"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "uptime"),
					// Test that model matches RTX pattern
					resource.TestMatchResourceAttr("data.rtx_system_info.test", "model", regexp.MustCompile(`^RTX\d+$`)),
					// Test that firmware version has correct format
					resource.TestMatchResourceAttr("data.rtx_system_info.test", "firmware_version", regexp.MustCompile(`^Rev\.\d+\.\d+\.\d+$`)),
					// Test that MAC address has correct format
					resource.TestMatchResourceAttr("data.rtx_system_info.test", "mac_address", regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)),
				),
			},
		},
	})
}

func TestAccRTXSystemInfoDataSource_attributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSystemInfoDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that all required attributes are present and non-empty
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "model", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("model should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "firmware_version", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("firmware_version should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "serial_number", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("serial_number should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "mac_address", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("mac_address should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "uptime", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("uptime should not be empty")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXSystemInfoDataSourceConfig() string {
	return `
data "rtx_system_info" "test" {}
`
}

func testAccCheckRTXSystemInfoDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		return nil
	}
}

// Helper functions for tests
func testAccPreCheck(t *testing.T) {
	// Skip if not running acceptance tests
	if !isAccTest() {
		t.Skip("Skipping acceptance test")
	}
	
	// Check required environment variables for Docker test environment
	requiredEnvVars := []string{
		"RTX_HOST",
		"RTX_USERNAME", 
		"RTX_PASSWORD",
	}
	
	for _, envVar := range requiredEnvVars {
		if value := getEnvVar(envVar); value == "" {
			t.Fatalf("Environment variable %s must be set for acceptance tests", envVar)
		}
	}
}

// isAccTest checks if acceptance tests are enabled
func isAccTest() bool {
	return os.Getenv("TF_ACC") != ""
}

// getEnvVar gets environment variable value
func getEnvVar(name string) string {
	return os.Getenv(name)
}

// testAccProviderFactories provides test provider factories for acceptance tests
var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"rtx": func() (*schema.Provider, error) {
		return New("test"), nil
	},
}

// SystemInfo represents parsed system information
type SystemInfo struct {
	Model           string
	FirmwareVersion string
	SerialNumber    string
	MACAddress      string
	Uptime          string
}

// parseSystemInfo parses RTX system information from command output
func parseSystemInfo(output string) (*SystemInfo, error) {
	if strings.TrimSpace(output) == "" {
		return nil, errors.New("empty output")
	}

	info := &SystemInfo{}
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "Model":
			info.Model = value
		case "Firmware Version":
			info.FirmwareVersion = value
		case "Serial Number":
			info.SerialNumber = value
		case "MAC Address":
			info.MACAddress = value
		case "Uptime":
			info.Uptime = value
		}
	}
	
	// Validate that we got all required fields
	if info.Model == "" {
		return nil, errors.New("model not found in output")
	}
	if info.FirmwareVersion == "" {
		return nil, errors.New("firmware version not found in output")
	}
	if info.SerialNumber == "" {
		return nil, errors.New("serial number not found in output")  
	}
	if info.MACAddress == "" {
		return nil, errors.New("MAC address not found in output")
	}
	if info.Uptime == "" {
		return nil, errors.New("uptime not found in output")
	}
	
	return info, nil
}