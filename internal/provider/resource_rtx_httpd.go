package provider

import (
	"context"
	"regexp"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXHTTPD() *schema.Resource {
	return &schema.Resource{
		Description: "Manages HTTP daemon (httpd) configuration on RTX routers. " +
			"This is a singleton resource - only one instance should exist per router. " +
			"The HTTPD service provides the web management interface for the router.",
		CreateContext: resourceRTXHTTPDCreate,
		ReadContext:   resourceRTXHTTPDRead,
		UpdateContext: resourceRTXHTTPDUpdate,
		DeleteContext: resourceRTXHTTPDDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXHTTPDImport,
		},

		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Interface to listen on. Use 'any' for all interfaces, or specify an interface name (e.g., 'lan1', 'pp1', 'bridge1', 'tunnel1').",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^(any|lan\d+|pp\d+|bridge\d+|tunnel\d+)$`),
					"must be 'any' or a valid interface name (e.g., lan1, pp1, bridge1, tunnel1)",
				),
			},
			"proxy_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable L2MS proxy access for HTTP. When enabled, allows proxy access via L2MS protocol.",
			},
		},
	}
}

func resourceRTXHTTPDCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildHTTPDConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_httpd").Msgf("Creating HTTPD configuration: %+v", config)

	err := apiClient.client.ConfigureHTTPD(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure HTTPD: %v", err)
	}

	// Use fixed ID for singleton resource
	d.SetId("httpd")

	// Read back to ensure consistency
	return resourceRTXHTTPDRead(ctx, d, meta)
}

func resourceRTXHTTPDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_httpd").Msg("Reading HTTPD configuration")

	var config *client.HTTPDConfig

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractHTTPD()
			if parsed != nil {
				config = convertParsedHTTPDConfig(parsed)
				logger.Debug().Str("resource", "rtx_httpd").Msg("Found HTTPD config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_httpd").Msg("HTTPD config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = apiClient.client.GetHTTPD(ctx)
		if err != nil {
			// Check if not configured
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_httpd").Msg("HTTPD not configured, removing from state")
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read HTTPD configuration: %v", err)
		}
	}

	// If no host is configured, the resource doesn't exist
	if config.Host == "" {
		logger.Debug().Str("resource", "rtx_httpd").Msg("HTTPD host not configured, removing from state")
		d.SetId("")
		return nil
	}

	// Update the state
	if err := d.Set("host", config.Host); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("proxy_access", config.ProxyAccess); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// convertParsedHTTPDConfig converts a parser HTTPDConfig to a client HTTPDConfig
func convertParsedHTTPDConfig(parsed *parsers.HTTPDConfig) *client.HTTPDConfig {
	return &client.HTTPDConfig{
		Host:        parsed.Host,
		ProxyAccess: parsed.ProxyAccess,
	}
}

func resourceRTXHTTPDUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildHTTPDConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_httpd").Msgf("Updating HTTPD configuration: %+v", config)

	err := apiClient.client.UpdateHTTPD(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update HTTPD configuration: %v", err)
	}

	return resourceRTXHTTPDRead(ctx, d, meta)
}

func resourceRTXHTTPDDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_httpd").Msg("Deleting HTTPD configuration")

	err := apiClient.client.ResetHTTPD(ctx)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to remove HTTPD configuration: %v", err)
	}

	return nil
}

func resourceRTXHTTPDImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_httpd").Msg("Importing HTTPD configuration")

	// Verify HTTPD is configured
	config, err := apiClient.client.GetHTTPD(ctx)
	if err != nil {
		return nil, err
	}

	if config.Host == "" {
		return nil, nil // Not configured, nothing to import
	}

	// Set the ID
	d.SetId("httpd")

	// Set attributes
	d.Set("host", config.Host)
	d.Set("proxy_access", config.ProxyAccess)

	return []*schema.ResourceData{d}, nil
}

// buildHTTPDConfigFromResourceData creates an HTTPDConfig from Terraform resource data
func buildHTTPDConfigFromResourceData(d *schema.ResourceData) client.HTTPDConfig {
	return client.HTTPDConfig{
		Host:        d.Get("host").(string),
		ProxyAccess: d.Get("proxy_access").(bool),
	}
}
