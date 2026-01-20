package provider

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXSyslog() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages syslog configuration on RTX routers. This is a singleton resource - only one instance can exist per router.",
		CreateContext: resourceRTXSyslogCreate,
		ReadContext:   resourceRTXSyslogRead,
		UpdateContext: resourceRTXSyslogUpdate,
		DeleteContext: resourceRTXSyslogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXSyslogImport,
		},

		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Syslog destination hosts (one or more)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IP address or hostname of the syslog server",
							ValidateFunc: validateSyslogHostAddress,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "UDP port (default 514, use 0 to use default)",
							ValidateFunc: validation.IntBetween(0, 65535),
						},
					},
				},
			},
			"local_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Source IP address for syslog messages",
				ValidateFunc: validateIPAddress,
			},
			"facility": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Syslog facility (user, local0-local7)",
				ValidateFunc: validateSyslogFacility,
			},
			"notice": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable notice level logging",
			},
			"info": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable info level logging",
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable debug level logging",
			},
		},
	}
}

func resourceRTXSyslogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSyslogConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_syslog").Msgf("Creating syslog configuration: %+v", config)

	err := apiClient.client.ConfigureSyslog(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create syslog configuration: %v", err)
	}

	// Singleton resource - use fixed ID
	d.SetId("syslog")

	// Read back to ensure consistency
	return resourceRTXSyslogRead(ctx, d, meta)
}

func resourceRTXSyslogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_syslog").Msg("Reading syslog configuration")

	config, err := apiClient.client.GetSyslogConfig(ctx)
	if err != nil {
		// Check if resource doesn't exist (no configuration)
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_syslog").Msg("Syslog configuration not found, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read syslog configuration: %v", err)
	}

	// Update the state
	hosts := make([]interface{}, len(config.Hosts))
	for i, host := range config.Hosts {
		hostMap := map[string]interface{}{
			"address": host.Address,
			"port":    host.Port,
		}
		hosts[i] = hostMap
	}
	if err := d.Set("host", hosts); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("local_address", config.LocalAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("facility", config.Facility); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notice", config.Notice); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("info", config.Info); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("debug", config.Debug); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXSyslogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSyslogConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_syslog").Msgf("Updating syslog configuration: %+v", config)

	err := apiClient.client.UpdateSyslogConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update syslog configuration: %v", err)
	}

	return resourceRTXSyslogRead(ctx, d, meta)
}

func resourceRTXSyslogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_syslog").Msg("Deleting syslog configuration")

	err := apiClient.client.ResetSyslog(ctx)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete syslog configuration: %v", err)
	}

	return nil
}

func resourceRTXSyslogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Accept "syslog" as the import ID (singleton resource)
	if importID != "syslog" {
		return nil, fmt.Errorf("invalid import ID format, expected 'syslog', got %q", importID)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_syslog").Msg("Importing syslog configuration")

	// Verify syslog configuration exists
	config, err := apiClient.client.GetSyslogConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import syslog configuration: %v", err)
	}

	// Set all attributes
	d.SetId("syslog")

	hosts := make([]interface{}, len(config.Hosts))
	for i, host := range config.Hosts {
		hostMap := map[string]interface{}{
			"address": host.Address,
			"port":    host.Port,
		}
		hosts[i] = hostMap
	}
	d.Set("host", hosts)
	d.Set("local_address", config.LocalAddress)
	d.Set("facility", config.Facility)
	d.Set("notice", config.Notice)
	d.Set("info", config.Info)
	d.Set("debug", config.Debug)

	return []*schema.ResourceData{d}, nil
}

// buildSyslogConfigFromResourceData creates a SyslogConfig from Terraform resource data
func buildSyslogConfigFromResourceData(d *schema.ResourceData) client.SyslogConfig {
	config := client.SyslogConfig{
		Hosts:    []client.SyslogHost{},
		Facility: d.Get("facility").(string),
		Notice:   d.Get("notice").(bool),
		Info:     d.Get("info").(bool),
		Debug:    d.Get("debug").(bool),
	}

	if v, ok := d.GetOk("local_address"); ok {
		config.LocalAddress = v.(string)
	}

	if v, ok := d.GetOk("host"); ok {
		hostSet := v.(*schema.Set)
		for _, hostItem := range hostSet.List() {
			hostMap := hostItem.(map[string]interface{})
			host := client.SyslogHost{
				Address: hostMap["address"].(string),
				Port:    hostMap["port"].(int),
			}
			config.Hosts = append(config.Hosts, host)
		}
	}

	return config
}

// validateSyslogHostAddress validates the syslog host address format
func validateSyslogHostAddress(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	// Check if it's a valid IP address
	if isValidIPv4(value) || isValidIPv6(value) {
		return nil, nil
	}

	// Check if it's a valid hostname (basic validation)
	if isValidHostname(value) {
		return nil, nil
	}

	return nil, []error{fmt.Errorf("%q must be a valid IP address or hostname, got %q", k, value)}
}

// validateSyslogFacility validates the syslog facility value
func validateSyslogFacility(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	validFacilities := []string{
		"user", "local0", "local1", "local2", "local3",
		"local4", "local5", "local6", "local7",
	}

	for _, valid := range validFacilities {
		if value == valid {
			return nil, nil
		}
	}

	return nil, []error{fmt.Errorf("%q must be one of: user, local0-local7, got %q", k, value)}
}

// isValidIPv4 checks if a string is a valid IPv4 address
func isValidIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		num := 0
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
			num = num*10 + int(c-'0')
		}
		if num > 255 {
			return false
		}
		// No leading zeros (except "0" itself)
		if len(part) > 1 && part[0] == '0' {
			return false
		}
	}
	return true
}

// isValidIPv6 checks if a string is a valid IPv6 address (basic check)
func isValidIPv6(ip string) bool {
	// Simple check for IPv6 format
	if !strings.Contains(ip, ":") {
		return false
	}
	// Must not have more than 8 groups
	parts := strings.Split(ip, ":")
	if len(parts) > 8 {
		return false
	}
	return true
}

// isValidHostname checks if a string is a valid hostname
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Hostname labels
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		// Must start and end with alphanumeric
		if !isAlphanumeric(rune(label[0])) || !isAlphanumeric(rune(label[len(label)-1])) {
			return false
		}
		// Can only contain alphanumeric and hyphens
		for _, c := range label {
			if !isAlphanumeric(c) && c != '-' {
				return false
			}
		}
	}
	return true
}

// isAlphanumeric checks if a rune is alphanumeric
func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
