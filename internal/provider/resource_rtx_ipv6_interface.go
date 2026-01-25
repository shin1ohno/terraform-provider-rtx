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

func resourceRTXIPv6Interface() *schema.Resource {
	return &schema.Resource{
		Description: "Manages IPv6 interface configuration on RTX routers. This includes IPv6 addresses, Router Advertisement (RTADV), DHCPv6, MTU, and security filters.",

		CreateContext: resourceRTXIPv6InterfaceCreate,
		ReadContext:   resourceRTXIPv6InterfaceRead,
		UpdateContext: resourceRTXIPv6InterfaceUpdate,
		DeleteContext: resourceRTXIPv6InterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXIPv6InterfaceImport,
		},

		Schema: map[string]*schema.Schema{
			"interface": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Interface name (e.g., 'lan1', 'lan2', 'bridge1', 'pp1', 'tunnel1')",
				ValidateFunc: validateIPv6InterfaceName,
			},
			"address": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "IPv6 address configuration blocks. Multiple addresses can be configured on a single interface.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Static IPv6 address in CIDR notation (e.g., '2001:db8::1/64'). Either 'address' or 'prefix_ref' with 'interface_id' must be specified.",
							ValidateFunc: validateIPv6CIDROptional,
						},
						"prefix_ref": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Prefix reference for dynamic address (e.g., 'ra-prefix@lan2', 'dhcp-prefix@lan2'). Must be used with 'interface_id'.",
						},
						"interface_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Interface identifier with prefix length (e.g., '::1/64'). Used with 'prefix_ref'.",
						},
					},
				},
			},
			"rtadv": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Router Advertisement (RTADV) configuration for this interface.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable Router Advertisement on this interface.",
						},
						"prefix_id": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "IPv6 prefix ID to advertise. Must match an rtx_ipv6_prefix resource.",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"o_flag": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Other Configuration Flag (O flag). When set, clients should use DHCPv6 for other configuration (e.g., DNS).",
						},
						"m_flag": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Managed Address Configuration Flag (M flag). When set, clients should use DHCPv6 for address assignment.",
						},
						"lifetime": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Router lifetime in seconds. Set to 0 to use the default value.",
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"dhcpv6_service": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "DHCPv6 service mode: 'server', 'client', or '' (disabled).",
				ValidateFunc:     validateDHCPv6Service,
				DiffSuppressFunc: SuppressCaseDiff, // DHCPv6 service modes are case-insensitive
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "IPv6 MTU size (minimum 1280 for IPv6). Set to 0 to use the default MTU.",
				ValidateFunc: validation.IntBetween(0, 65535),
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
		},
	}
}

func resourceRTXIPv6InterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", d.Id())
	config := buildIPv6InterfaceConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_interface").Msgf("Creating IPv6 interface configuration: %+v", config)

	err := apiClient.client.ConfigureIPv6Interface(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure IPv6 interface: %v", err)
	}

	// Use interface name as the resource ID
	d.SetId(config.Interface)

	// Explicitly set access list values to match the config.
	// The RTX router stores filter numbers, not access list names.
	// We must set these values explicitly to ensure the state matches the config.
	if err := d.Set("access_list_ipv6_in", config.AccessListIPv6In); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_out", config.AccessListIPv6Out); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPv6InterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", d.Id())
	logger := logging.FromContext(ctx)

	interfaceName := d.Id()

	logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("Reading IPv6 interface configuration: %s", interfaceName)

	var config *client.IPv6InterfaceConfig
	var err error

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, cacheErr := apiClient.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract IPv6 interfaces from parsed config
			interfaces := parsedConfig.ExtractIPv6Interfaces()
			if parsed, ok := interfaces[interfaceName]; ok {
				config = convertParsedIPv6InterfaceConfig(parsed)
				logger.Debug().Str("resource", "rtx_ipv6_interface").Msg("Found IPv6 interface in SFTP cache")
			}
		}
		if config == nil {
			// Interface not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_ipv6_interface").Msg("IPv6 interface not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or interface not found in cache
	if config == nil {
		config, err = apiClient.client.GetIPv6InterfaceConfig(ctx, interfaceName)
		if err != nil {
			// Check if interface doesn't have any configuration
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_ipv6_interface").Msgf("IPv6 interface %s configuration not found, removing from state", interfaceName)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read IPv6 interface configuration: %v", err)
		}
	}

	// Update the state
	if err := d.Set("interface", config.Interface); err != nil {
		return diag.FromErr(err)
	}

	// Set addresses
	addresses := make([]map[string]interface{}, len(config.Addresses))
	for i, addr := range config.Addresses {
		addresses[i] = map[string]interface{}{
			"address":      addr.Address,
			"prefix_ref":   addr.PrefixRef,
			"interface_id": addr.InterfaceID,
		}
	}
	if err := d.Set("address", addresses); err != nil {
		return diag.FromErr(err)
	}

	// Set RTADV block
	if config.RTADV != nil && config.RTADV.Enabled {
		rtadv := []map[string]interface{}{
			{
				"enabled":   config.RTADV.Enabled,
				"prefix_id": config.RTADV.PrefixID,
				"o_flag":    config.RTADV.OFlag,
				"m_flag":    config.RTADV.MFlag,
				"lifetime":  config.RTADV.Lifetime,
			},
		}
		if err := d.Set("rtadv", rtadv); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("rtadv", []map[string]interface{}{}); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set other fields
	if err := d.Set("dhcpv6_service", config.DHCPv6Service); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mtu", config.MTU); err != nil {
		return diag.FromErr(err)
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

	if err := preserveOrSet("access_list_ipv6_in", config.AccessListIPv6In); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_out", config.AccessListIPv6Out); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := preserveOrSet("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPv6InterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", d.Id())
	config := buildIPv6InterfaceConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_interface").Msgf("Updating IPv6 interface configuration: %+v", config)

	err := apiClient.client.UpdateIPv6InterfaceConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update IPv6 interface configuration: %v", err)
	}

	// Explicitly set access list values to match the config.
	// The RTX router stores filter numbers, not access list names.
	// We must set these values explicitly to ensure the state matches the config.
	if err := d.Set("access_list_ipv6_in", config.AccessListIPv6In); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_out", config.AccessListIPv6Out); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPv6InterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_interface", d.Id())
	interfaceName := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_interface").Msgf("Resetting IPv6 interface configuration: %s", interfaceName)

	err := apiClient.client.ResetIPv6Interface(ctx, interfaceName)
	if err != nil {
		// Check if it's already reset/clean
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to reset IPv6 interface configuration: %v", err)
	}

	return nil
}

func resourceRTXIPv6InterfaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Validate interface name format
	if err := validateIPv6InterfaceNameValue(importID); err != nil {
		return nil, fmt.Errorf("invalid import ID format: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_interface").Msgf("Importing IPv6 interface configuration: %s", importID)

	// Verify interface exists and retrieve configuration
	config, err := apiClient.client.GetIPv6InterfaceConfig(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("failed to import IPv6 interface %s: %v", importID, err)
	}

	// Set all attributes
	d.SetId(importID)
	d.Set("interface", config.Interface)

	// Set addresses
	addresses := make([]map[string]interface{}, len(config.Addresses))
	for i, addr := range config.Addresses {
		addresses[i] = map[string]interface{}{
			"address":      addr.Address,
			"prefix_ref":   addr.PrefixRef,
			"interface_id": addr.InterfaceID,
		}
	}
	d.Set("address", addresses)

	// Set RTADV
	if config.RTADV != nil && config.RTADV.Enabled {
		rtadv := []map[string]interface{}{
			{
				"enabled":   config.RTADV.Enabled,
				"prefix_id": config.RTADV.PrefixID,
				"o_flag":    config.RTADV.OFlag,
				"m_flag":    config.RTADV.MFlag,
				"lifetime":  config.RTADV.Lifetime,
			},
		}
		d.Set("rtadv", rtadv)
	}

	d.Set("dhcpv6_service", config.DHCPv6Service)
	d.Set("mtu", config.MTU)
	d.Set("access_list_ipv6_in", config.AccessListIPv6In)
	d.Set("access_list_ipv6_out", config.AccessListIPv6Out)
	d.Set("access_list_ipv6_dynamic_in", config.AccessListIPv6DynamicIn)
	d.Set("access_list_ipv6_dynamic_out", config.AccessListIPv6DynamicOut)

	return []*schema.ResourceData{d}, nil
}

// buildIPv6InterfaceConfigFromResourceData creates an IPv6InterfaceConfig from Terraform resource data
func buildIPv6InterfaceConfigFromResourceData(d *schema.ResourceData) client.IPv6InterfaceConfig {
	config := client.IPv6InterfaceConfig{
		Interface:                d.Get("interface").(string),
		DHCPv6Service:            d.Get("dhcpv6_service").(string),
		MTU:                      d.Get("mtu").(int),
		AccessListIPv6In:         d.Get("access_list_ipv6_in").(string),
		AccessListIPv6Out:        d.Get("access_list_ipv6_out").(string),
		AccessListIPv6DynamicIn:  d.Get("access_list_ipv6_dynamic_in").(string),
		AccessListIPv6DynamicOut: d.Get("access_list_ipv6_dynamic_out").(string),
	}

	// Handle address blocks
	if v, ok := d.GetOk("address"); ok {
		addressList := v.([]interface{})
		for _, addrRaw := range addressList {
			addrMap := addrRaw.(map[string]interface{})
			addr := client.IPv6Address{
				Address:     addrMap["address"].(string),
				PrefixRef:   addrMap["prefix_ref"].(string),
				InterfaceID: addrMap["interface_id"].(string),
			}
			config.Addresses = append(config.Addresses, addr)
		}
	}

	// Handle rtadv block
	if v, ok := d.GetOk("rtadv"); ok {
		rtadvList := v.([]interface{})
		if len(rtadvList) > 0 {
			rtadvMap := rtadvList[0].(map[string]interface{})
			config.RTADV = &client.RTADVConfig{
				Enabled:  rtadvMap["enabled"].(bool),
				PrefixID: rtadvMap["prefix_id"].(int),
				OFlag:    rtadvMap["o_flag"].(bool),
				MFlag:    rtadvMap["m_flag"].(bool),
				Lifetime: rtadvMap["lifetime"].(int),
			}
		}
	}

	return config
}

// validateIPv6InterfaceName validates the interface name format for Terraform schema
func validateIPv6InterfaceName(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if err := validateIPv6InterfaceNameValue(value); err != nil {
		return nil, []error{fmt.Errorf("%q %v", k, err)}
	}

	return nil, nil
}

// validateIPv6InterfaceNameValue validates the interface name format
func validateIPv6InterfaceNameValue(name string) error {
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

// validateIPv6CIDROptional validates IPv6 CIDR notation, allowing empty string
func validateIPv6CIDROptional(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	// Check for basic IPv6 CIDR format
	if !strings.Contains(value, "/") {
		return nil, []error{fmt.Errorf("%q must be in CIDR notation (e.g., '2001:db8::1/64')", k)}
	}

	// Must contain colons (IPv6 address)
	if !strings.Contains(value, ":") {
		return nil, []error{fmt.Errorf("%q must be a valid IPv6 address in CIDR notation", k)}
	}

	return nil, nil
}

// validateDHCPv6Service validates the DHCPv6 service value
func validateDHCPv6Service(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	validValues := []string{"server", "client"}
	for _, valid := range validValues {
		if value == valid {
			return nil, nil
		}
	}

	return nil, []error{fmt.Errorf("%q must be one of 'server', 'client', or empty (disabled)", k)}
}

// convertParsedIPv6InterfaceConfig converts a parser IPv6InterfaceConfig to a client IPv6InterfaceConfig
func convertParsedIPv6InterfaceConfig(parsed *parsers.IPv6InterfaceConfig) *client.IPv6InterfaceConfig {
	config := &client.IPv6InterfaceConfig{
		Interface:     parsed.Interface,
		DHCPv6Service: parsed.DHCPv6Service,
		MTU:           parsed.MTU,
		// Access list fields are populated from separate ACL resources
		// and are not parsed from the interface config directly
	}

	// Convert addresses
	for _, addr := range parsed.Addresses {
		config.Addresses = append(config.Addresses, client.IPv6Address{
			Address:     addr.Address,
			PrefixRef:   addr.PrefixRef,
			InterfaceID: addr.InterfaceID,
		})
	}

	// Convert RTADV config
	if parsed.RTADV != nil {
		config.RTADV = &client.RTADVConfig{
			Enabled:  parsed.RTADV.Enabled,
			PrefixID: parsed.RTADV.PrefixID,
			OFlag:    parsed.RTADV.OFlag,
			MFlag:    parsed.RTADV.MFlag,
			Lifetime: parsed.RTADV.Lifetime,
		}
	}

	return config
}
