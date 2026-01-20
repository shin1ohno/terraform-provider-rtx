package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestBuildDDNSServerConfigFromResourceData_BasicConfig(t *testing.T) {
	input := map[string]interface{}{
		"server_id": 1,
		"url":       "https://members.dyndns.org/nic/update",
		"hostname":  "myhost.dyndns.org",
		"username":  "testuser",
		"password":  "testpass",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXDDNS().Schema, input)
	config := buildDDNSServerConfigFromResourceData(d)

	assert.Equal(t, 1, config.ID)
	assert.Equal(t, "https://members.dyndns.org/nic/update", config.URL)
	assert.Equal(t, "myhost.dyndns.org", config.Hostname)
	assert.Equal(t, "testuser", config.Username)
	assert.Equal(t, "testpass", config.Password)
}

func TestBuildDDNSServerConfigFromResourceData_NoIPProvider(t *testing.T) {
	input := map[string]interface{}{
		"server_id": 2,
		"url":       "https://dynupdate.no-ip.com/nic/update",
		"hostname":  "example.no-ip.org",
		"username":  "noip_user",
		"password":  "noip_pass",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXDDNS().Schema, input)
	config := buildDDNSServerConfigFromResourceData(d)

	assert.Equal(t, 2, config.ID)
	assert.Equal(t, "https://dynupdate.no-ip.com/nic/update", config.URL)
	assert.Equal(t, "example.no-ip.org", config.Hostname)
	assert.Equal(t, "noip_user", config.Username)
	assert.Equal(t, "noip_pass", config.Password)
}

func TestBuildDDNSServerConfigFromResourceData_WithoutCredentials(t *testing.T) {
	input := map[string]interface{}{
		"server_id": 3,
		"url":       "https://api.example.com/update",
		"hostname":  "test.example.com",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXDDNS().Schema, input)
	config := buildDDNSServerConfigFromResourceData(d)

	assert.Equal(t, 3, config.ID)
	assert.Equal(t, "https://api.example.com/update", config.URL)
	assert.Equal(t, "test.example.com", config.Hostname)
	assert.Equal(t, "", config.Username)
	assert.Equal(t, "", config.Password)
}

func TestResourceRTXDDNSSchema(t *testing.T) {
	resource := resourceRTXDDNS()

	// Verify required fields
	assert.NotNil(t, resource.Schema["server_id"])
	assert.True(t, resource.Schema["server_id"].Required)
	assert.True(t, resource.Schema["server_id"].ForceNew)

	assert.NotNil(t, resource.Schema["url"])
	assert.True(t, resource.Schema["url"].Required)

	assert.NotNil(t, resource.Schema["hostname"])
	assert.True(t, resource.Schema["hostname"].Required)

	// Verify optional fields
	assert.NotNil(t, resource.Schema["username"])
	assert.True(t, resource.Schema["username"].Optional)

	assert.NotNil(t, resource.Schema["password"])
	assert.True(t, resource.Schema["password"].Optional)
	assert.True(t, resource.Schema["password"].Sensitive)
}

func TestBuildDDNSServerConfigFromResourceData_AllServerIDs(t *testing.T) {
	// Test all valid server IDs (1-4)
	for serverID := 1; serverID <= 4; serverID++ {
		input := map[string]interface{}{
			"server_id": serverID,
			"url":       "https://example.com/update",
			"hostname":  "test.example.com",
		}

		d := schema.TestResourceDataRaw(t, resourceRTXDDNS().Schema, input)
		config := buildDDNSServerConfigFromResourceData(d)

		assert.Equal(t, serverID, config.ID)
	}
}
