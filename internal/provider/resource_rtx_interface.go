package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXInterface() *schema.Resource {
	return &schema.Resource{
		Description: "Manages network interface configuration on RTX routers. This includes IP address assignment, security filters, NAT descriptors, and other interface-level settings.",

		CreateContext: resourceRTXInterfaceCreate,
		ReadContext:   resourceRTXInterfaceRead,
		UpdateContext: resourceRTXInterfaceUpdate,
		DeleteContext: resourceRTXInterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXInterfaceImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Interface name (e.g., 'lan1', 'lan2', 'bridge1', 'pp1', 'tunnel1')",
				ValidateFunc: validateInterfaceConfigName,
			},
			"interface_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The interface name. Same as 'name', provided for consistency with other resources.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Interface description",
			},
			"ip_address": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "IP address configuration block. Either 'address' or 'dhcp' must be set, but not both.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Static IP address in CIDR notation (e.g., '192.168.1.1/24')",
							ValidateFunc: validateCIDROptional,
						},
						"dhcp": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Use DHCP for IP address assignment",
						},
					},
				},
			},
			"nat_descriptor": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "NAT descriptor ID to bind to this interface. Use rtx_nat_masquerade or rtx_nat_static to define the descriptor.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"proxyarp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable ProxyARP on this interface",
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Maximum Transmission Unit size. Set to 0 to use the default MTU.",
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"access_list_ip_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound IP access list name",
			},
			"access_list_ip_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound IP access list name",
			},
			"access_list_ipv6_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound IPv6 access list name",
			},
			"access_list_ipv6_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound IPv6 access list name",
			},
			"access_list_ip_dynamic_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound dynamic IP access list name",
			},
			"access_list_ip_dynamic_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound dynamic IP access list name",
			},
			"access_list_ipv6_dynamic_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound dynamic IPv6 access list name",
			},
			"access_list_ipv6_dynamic_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound dynamic IPv6 access list name",
			},
			"access_list_mac_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound MAC access list name",
			},
			"access_list_mac_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound MAC access list name",
			},
		},
	}
}

func resourceRTXInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface", d.Id())
	config := buildInterfaceConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface").Msgf("Creating interface configuration: %+v", config)

	err := apiClient.client.ConfigureInterface(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure interface: %v", err)
	}

	// Use interface name as the resource ID
	d.SetId(config.Name)

	// Set interface_name (computed attribute) to match name
	if err := d.Set("interface_name", config.Name); err != nil {
		return diag.FromErr(err)
	}

	// Explicitly set access list values to match the config.
	// The RTX router stores filter numbers, not access list names.
	// We must set these values explicitly to ensure the state matches the config.
	if err := d.Set("access_list_ip_in", config.AccessListIPIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ip_out", config.AccessListIPOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_in", config.AccessListIPv6In); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_out", config.AccessListIPv6Out); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ip_dynamic_in", config.AccessListIPDynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ip_dynamic_out", config.AccessListIPDynamicOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_mac_in", config.AccessListMACIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_mac_out", config.AccessListMACOut); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface", d.Id())
	logger := logging.FromContext(ctx)

	interfaceName := d.Id()

	logger.Debug().Str("resource", "rtx_interface").Msgf("Reading interface configuration: %s", interfaceName)

	var config *client.InterfaceConfig
	var err error

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, cacheErr := apiClient.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract interfaces from parsed config
			interfaces := parsedConfig.ExtractInterfaces()
			if parsed, ok := interfaces[interfaceName]; ok {
				config = convertParsedInterfaceConfig(parsed)
				logger.Debug().Str("resource", "rtx_interface").Msg("Found interface in SFTP cache")
			}
		}
		if config == nil {
			// Interface not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_interface").Msg("Interface not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or interface not found in cache
	if config == nil {
		config, err = apiClient.client.GetInterfaceConfig(ctx, interfaceName)
		if err != nil {
			// Check if interface doesn't have any configuration
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_interface").Msgf("Interface %s configuration not found, removing from state", interfaceName)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read interface configuration: %v", err)
		}
	}

	// Update the state
	if err := d.Set("name", config.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interface_name", config.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", config.Description); err != nil {
		return diag.FromErr(err)
	}

	// Set IP address block
	if config.IPAddress != nil && (config.IPAddress.Address != "" || config.IPAddress.DHCP) {
		ipAddress := []map[string]interface{}{
			{
				"address": config.IPAddress.Address,
				"dhcp":    config.IPAddress.DHCP,
			},
		}
		if err := d.Set("ip_address", ipAddress); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("ip_address", []map[string]interface{}{}); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set access list attributes - preserve values when service returns empty.
	// RTX router stores filter numbers, not access list names. The service layer cannot
	// reverse-lookup names from numbers, so it returns empty values.
	//
	// We use a fallback chain:
	// 1. Service value (if not empty)
	// 2. Config value via GetRawConfig() (if available)
	// 3. Prior state value via d.Get()
	//
	// This preserves access list names across plan/apply cycles.
	rawConfig := d.GetRawConfig()
	preserveOrSet := func(attrName, serviceValue string) error {
		value := serviceValue
		if value == "" {
			// Try to get from raw config (Terraform configuration)
			if !rawConfig.IsNull() {
				configVal := rawConfig.GetAttr(attrName)
				if !configVal.IsNull() && !configVal.IsKnown() {
					// Value is unknown (being computed) - use prior state
					value = d.Get(attrName).(string)
				} else if !configVal.IsNull() && configVal.AsString() != "" {
					value = configVal.AsString()
				} else {
					// Config is empty or null, use prior state
					value = d.Get(attrName).(string)
				}
			} else {
				// No raw config available, use prior state
				value = d.Get(attrName).(string)
			}
		}
		return d.Set(attrName, value)
	}

	if err := preserveOrSet("access_list_ip_in", config.AccessListIPIn); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ip_out", config.AccessListIPOut); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_in", config.AccessListIPv6In); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_out", config.AccessListIPv6Out); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ip_dynamic_in", config.AccessListIPDynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ip_dynamic_out", config.AccessListIPDynamicOut); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_mac_in", config.AccessListMACIn); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_mac_out", config.AccessListMACOut); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("nat_descriptor", config.NATDescriptor); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("proxyarp", config.ProxyARP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mtu", config.MTU); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface", d.Id())
	config := buildInterfaceConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface").Msgf("Updating interface configuration: %+v", config)

	err := apiClient.client.UpdateInterfaceConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update interface configuration: %v", err)
	}

	// Explicitly set access list values to match the config.
	// The RTX router stores filter numbers, not access list names.
	// We must set these values explicitly to ensure the state matches the config.
	if err := d.Set("access_list_ip_in", config.AccessListIPIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ip_out", config.AccessListIPOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_in", config.AccessListIPv6In); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_out", config.AccessListIPv6Out); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ip_dynamic_in", config.AccessListIPDynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ip_dynamic_out", config.AccessListIPDynamicOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_mac_in", config.AccessListMACIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_mac_out", config.AccessListMACOut); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface", d.Id())
	interfaceName := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface").Msgf("Resetting interface configuration: %s", interfaceName)

	err := apiClient.client.ResetInterface(ctx, interfaceName)
	if err != nil {
		// Check if it's already reset/clean
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to reset interface configuration: %v", err)
	}

	return nil
}

func resourceRTXInterfaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Validate interface name format
	if err := validateInterfaceConfigNameValue(importID); err != nil {
		return nil, fmt.Errorf("invalid import ID format: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface").Msgf("Importing interface configuration: %s", importID)

	// Verify interface exists and retrieve configuration
	config, err := apiClient.client.GetInterfaceConfig(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("failed to import interface %s: %v", importID, err)
	}

	// Set all attributes
	d.SetId(importID)
	d.Set("name", config.Name)
	d.Set("interface_name", config.Name)
	d.Set("description", config.Description)

	// Set IP address block
	if config.IPAddress != nil && (config.IPAddress.Address != "" || config.IPAddress.DHCP) {
		ipAddress := []map[string]interface{}{
			{
				"address": config.IPAddress.Address,
				"dhcp":    config.IPAddress.DHCP,
			},
		}
		d.Set("ip_address", ipAddress)
	}

	d.Set("access_list_ip_in", config.AccessListIPIn)
	d.Set("access_list_ip_out", config.AccessListIPOut)
	d.Set("access_list_ipv6_in", config.AccessListIPv6In)
	d.Set("access_list_ipv6_out", config.AccessListIPv6Out)
	d.Set("access_list_ip_dynamic_in", config.AccessListIPDynamicIn)
	d.Set("access_list_ip_dynamic_out", config.AccessListIPDynamicOut)
	d.Set("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn)
	d.Set("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut)
	d.Set("access_list_mac_in", config.AccessListMACIn)
	d.Set("access_list_mac_out", config.AccessListMACOut)
	d.Set("nat_descriptor", config.NATDescriptor)
	d.Set("proxyarp", config.ProxyARP)
	d.Set("mtu", config.MTU)

	return []*schema.ResourceData{d}, nil
}

// buildInterfaceConfigFromResourceData creates an InterfaceConfig from Terraform resource data
func buildInterfaceConfigFromResourceData(d *schema.ResourceData) client.InterfaceConfig {
	config := client.InterfaceConfig{
		Name:                     d.Get("name").(string),
		Description:              d.Get("description").(string),
		NATDescriptor:            d.Get("nat_descriptor").(int),
		ProxyARP:                 d.Get("proxyarp").(bool),
		MTU:                      d.Get("mtu").(int),
		AccessListIPIn:           d.Get("access_list_ip_in").(string),
		AccessListIPOut:          d.Get("access_list_ip_out").(string),
		AccessListIPv6In:         d.Get("access_list_ipv6_in").(string),
		AccessListIPv6Out:        d.Get("access_list_ipv6_out").(string),
		AccessListIPDynamicIn:    d.Get("access_list_ip_dynamic_in").(string),
		AccessListIPDynamicOut:   d.Get("access_list_ip_dynamic_out").(string),
		AccessListIPv6DynamicIn:  d.Get("access_list_ipv6_dynamic_in").(string),
		AccessListIPv6DynamicOut: d.Get("access_list_ipv6_dynamic_out").(string),
		AccessListMACIn:          d.Get("access_list_mac_in").(string),
		AccessListMACOut:         d.Get("access_list_mac_out").(string),
	}

	// Handle ip_address block
	if v, ok := d.GetOk("ip_address"); ok {
		ipList := v.([]interface{})
		if len(ipList) > 0 {
			ipMap := ipList[0].(map[string]interface{})
			config.IPAddress = &client.InterfaceIP{
				Address: ipMap["address"].(string),
				DHCP:    ipMap["dhcp"].(bool),
			}
		}
	}

	return config
}

// validateInterfaceConfigName validates the interface name format for Terraform schema
func validateInterfaceConfigName(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if err := validateInterfaceConfigNameValue(value); err != nil {
		return nil, []error{fmt.Errorf("%q %v", k, err)}
	}

	return nil, nil
}

// validateInterfaceConfigNameValue validates the interface name format
func validateInterfaceConfigNameValue(name string) error {
	if name == "" {
		return fmt.Errorf("interface name cannot be empty")
	}

	// Valid patterns: lan1, lan2, lan3, bridge1, bridge10, pp1, pp10, tunnel1, tunnel100
	pattern := regexp.MustCompile(`^(lan|bridge|pp|tunnel)\d+$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("must be a valid interface name (e.g., 'lan1', 'lan2', 'bridge1', 'pp1', 'tunnel1')")
	}

	return nil
}

// validateCIDROptional validates CIDR notation, allowing empty string
func validateCIDROptional(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	// Reuse the validateCIDR function for non-empty values
	return validateCIDR(v, k)
}

// convertParsedInterfaceConfig converts a parser InterfaceConfig to a client InterfaceConfig
func convertParsedInterfaceConfig(parsed *parsers.InterfaceConfig) *client.InterfaceConfig {
	config := &client.InterfaceConfig{
		Name:          parsed.Name,
		Description:   parsed.Description,
		NATDescriptor: parsed.NATDescriptor,
		ProxyARP:      parsed.ProxyARP,
		MTU:           parsed.MTU,
		// Access list fields are populated from separate ACL resources
		// and are not parsed from the interface config directly
	}

	if parsed.IPAddress != nil {
		config.IPAddress = &client.InterfaceIP{
			Address: parsed.IPAddress.Address,
			DHCP:    parsed.IPAddress.DHCP,
		}
	}

	return config
}

