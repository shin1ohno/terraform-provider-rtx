package provider

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXSNMPServer() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages SNMP configuration on RTX routers. This is a singleton resource - there is only one SNMP configuration per router.",
		CreateContext: resourceRTXSNMPServerCreate,
		ReadContext:   resourceRTXSNMPServerRead,
		UpdateContext: resourceRTXSNMPServerUpdate,
		DeleteContext: resourceRTXSNMPServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXSNMPServerImport,
		},

		Schema: map[string]*schema.Schema{
			"location": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "System location (SNMP sysLocation). Describes the physical location of the device.",
			},
			"contact": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "System contact (SNMP sysContact). Contact information for the device administrator.",
			},
			"chassis_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "System name (SNMP sysName). Unique identifier for the device.",
			},
			"community": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "SNMP community configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Community string name. This is sensitive as it acts as a password for SNMP access.",
						},
						"permission": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Access permission: 'ro' (read-only) or 'rw' (read-write)",
							ValidateFunc: validation.StringInSlice([]string{"ro", "rw"}, false),
						},
						"acl": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Access control list number to restrict which hosts can use this community",
						},
					},
				},
			},
			"host": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "SNMP trap host configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IP address of the SNMP trap receiver",
							ValidateFunc: validateIPAddress,
						},
						"community": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Community string to use when sending traps to this host",
						},
						"version": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "SNMP version to use: '1' or '2c'",
							ValidateFunc: validation.StringInSlice([]string{"1", "2c"}, false),
						},
					},
				},
			},
			"enable_traps": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of trap types to enable. Valid values: all, authentication, coldstart, warmstart, linkdown, linkup, enterprise",
			},
		},
	}
}

func resourceRTXSNMPServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSNMPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_snmp_server").Msgf("Creating SNMP configuration: %+v", config)

	err := apiClient.client.CreateSNMP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create SNMP configuration: %v", err)
	}

	// Use fixed ID "snmp" for singleton resource
	d.SetId("snmp")

	// Read back to ensure consistency
	return resourceRTXSNMPServerRead(ctx, d, meta)
}

func resourceRTXSNMPServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_snmp_server").Msg("Reading SNMP configuration")

	config, err := apiClient.client.GetSNMP(ctx)
	if err != nil {
		return diag.Errorf("Failed to read SNMP configuration: %v", err)
	}

	// Update the state
	if err := d.Set("location", config.SysLocation); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact", config.SysContact); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("chassis_id", config.SysName); err != nil {
		return diag.FromErr(err)
	}

	// Convert Communities to list
	communities := make([]map[string]interface{}, len(config.Communities))
	for i, c := range config.Communities {
		communities[i] = map[string]interface{}{
			"name":       c.Name,
			"permission": c.Permission,
			"acl":        c.ACL,
		}
	}
	if err := d.Set("community", communities); err != nil {
		return diag.FromErr(err)
	}

	// Convert Hosts to list
	hosts := make([]map[string]interface{}, len(config.Hosts))
	for i, h := range config.Hosts {
		hosts[i] = map[string]interface{}{
			"ip_address": h.Address,
			"community":  h.Community,
			"version":    h.Version,
		}
	}
	if err := d.Set("host", hosts); err != nil {
		return diag.FromErr(err)
	}

	// Set enable_traps
	if err := d.Set("enable_traps", config.TrapEnable); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXSNMPServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSNMPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_snmp_server").Msgf("Updating SNMP configuration: %+v", config)

	err := apiClient.client.UpdateSNMP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update SNMP configuration: %v", err)
	}

	return resourceRTXSNMPServerRead(ctx, d, meta)
}

func resourceRTXSNMPServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_snmp_server").Msg("Deleting SNMP configuration")

	err := apiClient.client.DeleteSNMP(ctx)
	if err != nil {
		return diag.Errorf("Failed to delete SNMP configuration: %v", err)
	}

	return nil
}

func resourceRTXSNMPServerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Only accept "snmp" as valid import ID (singleton resource)
	if importID != "snmp" {
		return nil, fmt.Errorf("invalid import ID format, expected 'snmp' for singleton resource, got: %s", importID)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_snmp_server").Msg("Importing SNMP configuration")

	// Verify configuration exists and retrieve it
	config, err := apiClient.client.GetSNMP(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import SNMP configuration: %v", err)
	}

	// Set the ID
	d.SetId("snmp")

	// Set all attributes
	d.Set("location", config.SysLocation)
	d.Set("contact", config.SysContact)
	d.Set("chassis_id", config.SysName)

	// Set communities
	communities := make([]map[string]interface{}, len(config.Communities))
	for i, c := range config.Communities {
		communities[i] = map[string]interface{}{
			"name":       c.Name,
			"permission": c.Permission,
			"acl":        c.ACL,
		}
	}
	d.Set("community", communities)

	// Set hosts
	hosts := make([]map[string]interface{}, len(config.Hosts))
	for i, h := range config.Hosts {
		hosts[i] = map[string]interface{}{
			"ip_address": h.Address,
			"community":  h.Community,
			"version":    h.Version,
		}
	}
	d.Set("host", hosts)

	// Set enable_traps
	d.Set("enable_traps", config.TrapEnable)

	return []*schema.ResourceData{d}, nil
}

// buildSNMPConfigFromResourceData creates an SNMPConfig from Terraform resource data
func buildSNMPConfigFromResourceData(d *schema.ResourceData) client.SNMPConfig {
	config := client.SNMPConfig{
		SysLocation: d.Get("location").(string),
		SysContact:  d.Get("contact").(string),
		SysName:     d.Get("chassis_id").(string),
		Communities: []client.SNMPCommunity{},
		Hosts:       []client.SNMPHost{},
		TrapEnable:  []string{},
	}

	// Handle community list
	if v, ok := d.GetOk("community"); ok {
		communityList := v.([]interface{})
		for _, cRaw := range communityList {
			cMap := cRaw.(map[string]interface{})
			config.Communities = append(config.Communities, client.SNMPCommunity{
				Name:       cMap["name"].(string),
				Permission: cMap["permission"].(string),
				ACL:        cMap["acl"].(string),
			})
		}
	}

	// Handle host list
	if v, ok := d.GetOk("host"); ok {
		hostList := v.([]interface{})
		for _, hRaw := range hostList {
			hMap := hRaw.(map[string]interface{})
			config.Hosts = append(config.Hosts, client.SNMPHost{
				Address:   hMap["ip_address"].(string),
				Community: hMap["community"].(string),
				Version:   hMap["version"].(string),
			})
		}
	}

	// Handle enable_traps list
	if v, ok := d.GetOk("enable_traps"); ok {
		trapList := v.([]interface{})
		for _, trap := range trapList {
			config.TrapEnable = append(config.TrapEnable, trap.(string))
		}
	}

	return config
}

// validateIPAddress is defined in resource_rtx_dhcp_scope.go
