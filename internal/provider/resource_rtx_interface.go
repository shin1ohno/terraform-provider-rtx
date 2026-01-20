package provider

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
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
			"secure_filter_in": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Inbound security filter numbers. Order matters - first match wins.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},
			"secure_filter_out": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Outbound security filter numbers. Order matters - first match wins.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},
			"dynamic_filter_out": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Dynamic filter numbers for outbound stateful inspection.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
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
		},
	}
}

func resourceRTXInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildInterfaceConfigFromResourceData(d)

	log.Printf("[DEBUG] Creating interface configuration: %+v", config)

	err := apiClient.client.ConfigureInterface(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure interface: %v", err)
	}

	// Use interface name as the resource ID
	d.SetId(config.Name)

	// Read back to ensure consistency
	return resourceRTXInterfaceRead(ctx, d, meta)
}

func resourceRTXInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	interfaceName := d.Id()

	log.Printf("[DEBUG] Reading interface configuration: %s", interfaceName)

	config, err := apiClient.client.GetInterfaceConfig(ctx, interfaceName)
	if err != nil {
		// Check if interface doesn't have any configuration
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] Interface %s configuration not found, removing from state", interfaceName)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read interface configuration: %v", err)
	}

	// Update the state
	if err := d.Set("name", config.Name); err != nil {
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

	// Set security filters
	if err := d.Set("secure_filter_in", config.SecureFilterIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("secure_filter_out", config.SecureFilterOut); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dynamic_filter_out", config.DynamicFilterOut); err != nil {
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

	config := buildInterfaceConfigFromResourceData(d)

	log.Printf("[DEBUG] Updating interface configuration: %+v", config)

	err := apiClient.client.UpdateInterfaceConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update interface configuration: %v", err)
	}

	return resourceRTXInterfaceRead(ctx, d, meta)
}

func resourceRTXInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	interfaceName := d.Id()

	log.Printf("[DEBUG] Resetting interface configuration: %s", interfaceName)

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

	log.Printf("[DEBUG] Importing interface configuration: %s", importID)

	// Verify interface exists and retrieve configuration
	config, err := apiClient.client.GetInterfaceConfig(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("failed to import interface %s: %v", importID, err)
	}

	// Set all attributes
	d.SetId(importID)
	d.Set("name", config.Name)
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

	d.Set("secure_filter_in", config.SecureFilterIn)
	d.Set("secure_filter_out", config.SecureFilterOut)
	d.Set("dynamic_filter_out", config.DynamicFilterOut)
	d.Set("nat_descriptor", config.NATDescriptor)
	d.Set("proxyarp", config.ProxyARP)
	d.Set("mtu", config.MTU)

	return []*schema.ResourceData{d}, nil
}

// buildInterfaceConfigFromResourceData creates an InterfaceConfig from Terraform resource data
func buildInterfaceConfigFromResourceData(d *schema.ResourceData) client.InterfaceConfig {
	config := client.InterfaceConfig{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		NATDescriptor: d.Get("nat_descriptor").(int),
		ProxyARP:      d.Get("proxyarp").(bool),
		MTU:           d.Get("mtu").(int),
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

	// Handle secure_filter_in
	if v, ok := d.GetOk("secure_filter_in"); ok {
		filtersList := v.([]interface{})
		filters := make([]int, len(filtersList))
		for i, f := range filtersList {
			filters[i] = f.(int)
		}
		config.SecureFilterIn = filters
	}

	// Handle secure_filter_out
	if v, ok := d.GetOk("secure_filter_out"); ok {
		filtersList := v.([]interface{})
		filters := make([]int, len(filtersList))
		for i, f := range filtersList {
			filters[i] = f.(int)
		}
		config.SecureFilterOut = filters
	}

	// Handle dynamic_filter_out
	if v, ok := d.GetOk("dynamic_filter_out"); ok {
		filtersList := v.([]interface{})
		filters := make([]int, len(filtersList))
		for i, f := range filtersList {
			filters[i] = f.(int)
		}
		config.DynamicFilterOut = filters
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
