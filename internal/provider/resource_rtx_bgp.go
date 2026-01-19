package provider

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXBGP() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages BGP (Border Gateway Protocol) configuration on RTX routers. BGP is a singleton resource - only one BGP configuration can exist per router.",
		CreateContext: resourceRTXBGPCreate,
		ReadContext:   resourceRTXBGPRead,
		UpdateContext: resourceRTXBGPUpdate,
		DeleteContext: resourceRTXBGPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXBGPImport,
		},

		Schema: map[string]*schema.Schema{
			"asn": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Autonomous System Number (1-4294967295). Supports 4-byte ASN.",
				ValidateFunc: validateASN,
			},
			"router_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "BGP router ID in IPv4 address format. If not set, uses the highest loopback or interface IP.",
				ValidateFunc: validateIPv4Address,
			},
			"default_ipv4_unicast": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable IPv4 unicast address family by default for new neighbors.",
			},
			"log_neighbor_changes": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Log neighbor up/down changes.",
			},
			"neighbor": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "BGP neighbor configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Neighbor ID (1-based index for RTX router).",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Neighbor IP address.",
							ValidateFunc: validateIPv4Address,
						},
						"remote_as": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Remote AS number (1-4294967295).",
							ValidateFunc: validateASN,
						},
						"hold_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Hold time in seconds (3-65535). Default is 90.",
							ValidateFunc: validation.IntBetween(3, 65535),
						},
						"keepalive": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Keepalive interval in seconds (1-21845). Default is 30.",
							ValidateFunc: validation.IntBetween(1, 21845),
						},
						"multihop": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "eBGP multihop TTL (1-255). Required for non-directly connected eBGP peers.",
							ValidateFunc: validation.IntBetween(1, 255),
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "MD5 authentication password for the BGP session.",
						},
						"local_address": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Local IP address for the BGP session.",
							ValidateFunc: validateIPv4Address,
						},
					},
				},
			},
			"network": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Networks to announce via BGP.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Network prefix to announce.",
							ValidateFunc: validateIPv4Address,
						},
						"mask": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Network mask in dotted decimal notation.",
							ValidateFunc: validateRouteMask,
						},
					},
				},
			},
			"redistribute_static": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Redistribute static routes into BGP.",
			},
			"redistribute_connected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Redistribute connected routes into BGP.",
			},
		},
	}
}

func resourceRTXBGPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildBGPConfigFromResourceData(d)

	log.Printf("[DEBUG] Creating BGP configuration: %+v", config)

	err := apiClient.client.ConfigureBGP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create BGP configuration: %v", err)
	}

	// BGP is a singleton resource
	d.SetId("bgp")

	return resourceRTXBGPRead(ctx, d, meta)
}

func resourceRTXBGPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Reading BGP configuration")

	config, err := apiClient.client.GetBGPConfig(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			log.Printf("[DEBUG] BGP configuration not found, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read BGP configuration: %v", err)
	}

	if !config.Enabled {
		log.Printf("[DEBUG] BGP is disabled, removing from state")
		d.SetId("")
		return nil
	}

	// Update the state
	if err := d.Set("asn", config.ASN); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("router_id", config.RouterID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_ipv4_unicast", config.DefaultIPv4Unicast); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("log_neighbor_changes", config.LogNeighborChanges); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redistribute_static", config.RedistributeStatic); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redistribute_connected", config.RedistributeConnected); err != nil {
		return diag.FromErr(err)
	}

	// Convert neighbors
	neighbors := make([]map[string]interface{}, len(config.Neighbors))
	for i, n := range config.Neighbors {
		neighbors[i] = map[string]interface{}{
			"id":            n.ID,
			"ip":            n.IP,
			"remote_as":     n.RemoteAS,
			"hold_time":     n.HoldTime,
			"keepalive":     n.Keepalive,
			"multihop":      n.Multihop,
			"password":      n.Password,
			"local_address": n.LocalAddress,
		}
	}
	if err := d.Set("neighbor", neighbors); err != nil {
		return diag.FromErr(err)
	}

	// Convert networks
	networks := make([]map[string]interface{}, len(config.Networks))
	for i, net := range config.Networks {
		networks[i] = map[string]interface{}{
			"prefix": net.Prefix,
			"mask":   net.Mask,
		}
	}
	if err := d.Set("network", networks); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXBGPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildBGPConfigFromResourceData(d)

	log.Printf("[DEBUG] Updating BGP configuration: %+v", config)

	err := apiClient.client.UpdateBGPConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update BGP configuration: %v", err)
	}

	return resourceRTXBGPRead(ctx, d, meta)
}

func resourceRTXBGPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Disabling BGP configuration")

	err := apiClient.client.ResetBGP(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to disable BGP: %v", err)
	}

	return nil
}

func resourceRTXBGPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	// Import ID should be "bgp" (singleton resource)
	if d.Id() != "bgp" {
		return nil, fmt.Errorf("import ID must be 'bgp' for this singleton resource")
	}

	log.Printf("[DEBUG] Importing BGP configuration")

	config, err := apiClient.client.GetBGPConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import BGP configuration: %v", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("BGP is not configured on this router")
	}

	d.SetId("bgp")
	d.Set("asn", config.ASN)
	d.Set("router_id", config.RouterID)
	d.Set("default_ipv4_unicast", config.DefaultIPv4Unicast)
	d.Set("log_neighbor_changes", config.LogNeighborChanges)
	d.Set("redistribute_static", config.RedistributeStatic)
	d.Set("redistribute_connected", config.RedistributeConnected)

	neighbors := make([]map[string]interface{}, len(config.Neighbors))
	for i, n := range config.Neighbors {
		neighbors[i] = map[string]interface{}{
			"id":            n.ID,
			"ip":            n.IP,
			"remote_as":     n.RemoteAS,
			"hold_time":     n.HoldTime,
			"keepalive":     n.Keepalive,
			"multihop":      n.Multihop,
			"password":      n.Password,
			"local_address": n.LocalAddress,
		}
	}
	d.Set("neighbor", neighbors)

	networks := make([]map[string]interface{}, len(config.Networks))
	for i, net := range config.Networks {
		networks[i] = map[string]interface{}{
			"prefix": net.Prefix,
			"mask":   net.Mask,
		}
	}
	d.Set("network", networks)

	return []*schema.ResourceData{d}, nil
}

func buildBGPConfigFromResourceData(d *schema.ResourceData) client.BGPConfig {
	config := client.BGPConfig{
		Enabled:               true,
		ASN:                   d.Get("asn").(string),
		RouterID:              d.Get("router_id").(string),
		DefaultIPv4Unicast:    d.Get("default_ipv4_unicast").(bool),
		LogNeighborChanges:    d.Get("log_neighbor_changes").(bool),
		RedistributeStatic:    d.Get("redistribute_static").(bool),
		RedistributeConnected: d.Get("redistribute_connected").(bool),
	}

	// Handle neighbors
	if v, ok := d.GetOk("neighbor"); ok {
		neighborList := v.([]interface{})
		config.Neighbors = make([]client.BGPNeighbor, len(neighborList))
		for i, n := range neighborList {
			nMap := n.(map[string]interface{})
			config.Neighbors[i] = client.BGPNeighbor{
				ID:           nMap["id"].(int),
				IP:           nMap["ip"].(string),
				RemoteAS:     nMap["remote_as"].(string),
				HoldTime:     nMap["hold_time"].(int),
				Keepalive:    nMap["keepalive"].(int),
				Multihop:     nMap["multihop"].(int),
				Password:     nMap["password"].(string),
				LocalAddress: nMap["local_address"].(string),
			}
		}
	}

	// Handle networks
	if v, ok := d.GetOk("network"); ok {
		networkList := v.([]interface{})
		config.Networks = make([]client.BGPNetwork, len(networkList))
		for i, n := range networkList {
			nMap := n.(map[string]interface{})
			config.Networks[i] = client.BGPNetwork{
				Prefix: nMap["prefix"].(string),
				Mask:   nMap["mask"].(string),
			}
		}
	}

	return config
}

// validateASN validates an Autonomous System Number
func validateASN(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	asn, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a valid AS number (1-4294967295), got %q", k, value)}
	}

	if asn == 0 {
		return nil, []error{fmt.Errorf("%q must be between 1 and 4294967295, got %q", k, value)}
	}

	return nil, nil
}

// validateIPv4Address validates an IPv4 address
func validateIPv4Address(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil // Optional field
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IPv4 address, got %q", k, value)}
	}

	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q must be an IPv4 address, got %q", k, value)}
	}

	return nil, nil
}
