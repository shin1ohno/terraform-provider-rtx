package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// resourceRTXEthernetFilter returns the schema for the rtx_ethernet_filter resource
func resourceRTXEthernetFilter() *schema.Resource {
	return &schema.Resource{
		Description: "Manages Ethernet (Layer 2) filters on RTX routers. " +
			"Ethernet filters can match traffic based on MAC addresses or DHCP binding status.",

		CreateContext: resourceRTXEthernetFilterCreate,
		ReadContext:   resourceRTXEthernetFilterRead,
		UpdateContext: resourceRTXEthernetFilterUpdate,
		DeleteContext: resourceRTXEthernetFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXEthernetFilterImport,
		},

		Schema: map[string]*schema.Schema{
			"number": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Filter number (1-512)",
				ValidateFunc: validation.IntBetween(1, 512),
			},
			"action": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Action to take: pass-log, pass-nolog, reject-log, reject-nolog, pass, reject",
				ValidateFunc: validation.StringInSlice([]string{
					"pass-log", "pass-nolog", "reject-log", "reject-nolog", "pass", "reject",
				}, false),
			},
			// MAC-based filter fields
			"source_mac": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Source MAC address (e.g., 00:11:22:33:44:55 or * for any)",
				ConflictsWith: []string{"dhcp_type"},
			},
			"destination_mac": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Destination MAC address (e.g., 00:11:22:33:44:55 or * for any)",
				ConflictsWith: []string{"dhcp_type"},
			},
			"ether_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Ethernet type (e.g., 0x0800 for IPv4, 0x0806 for ARP)",
				ConflictsWith: []string{"dhcp_type"},
			},
			"vlan_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "VLAN ID to match (1-4094)",
				ValidateFunc:  validation.IntBetween(1, 4094),
				ConflictsWith: []string{"dhcp_type"},
			},
			// DHCP-based filter fields
			"dhcp_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DHCP filter type: dhcp-bind or dhcp-not-bind",
				ValidateFunc: validation.StringInSlice([]string{
					"dhcp-bind", "dhcp-not-bind",
				}, false),
				ConflictsWith: []string{"source_mac", "destination_mac", "ether_type", "vlan_id", "offset", "byte_list"},
			},
			"dhcp_scope": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "DHCP scope number (for DHCP-based filters)",
				ValidateFunc: validation.IntAtLeast(1),
			},
			// Advanced byte-match filter fields
			"offset": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Byte offset for byte-match filtering",
				ConflictsWith: []string{"dhcp_type"},
			},
			"byte_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Byte patterns for byte-match filtering",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"dhcp_type"},
			},
		},
	}
}

func resourceRTXEthernetFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	filter := buildEthernetFilterFromResourceData(d)

	log.Printf("[DEBUG] Creating Ethernet filter: %d", filter.Number)

	err := apiClient.client.CreateEthernetFilter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to create Ethernet filter: %v", err)
	}

	d.SetId(strconv.Itoa(filter.Number))

	return resourceRTXEthernetFilterRead(ctx, d, meta)
}

func resourceRTXEthernetFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	number, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter number: %v", err)
	}

	log.Printf("[DEBUG] Reading Ethernet filter: %d", number)

	filter, err := apiClient.client.GetEthernetFilter(ctx, number)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			log.Printf("[WARN] Ethernet filter %d not found, removing from state", number)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read Ethernet filter: %v", err)
	}

	if err := flattenEthernetFilterToResourceData(filter, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXEthernetFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	filter := buildEthernetFilterFromResourceData(d)

	log.Printf("[DEBUG] Updating Ethernet filter: %d", filter.Number)

	err := apiClient.client.UpdateEthernetFilter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to update Ethernet filter: %v", err)
	}

	return resourceRTXEthernetFilterRead(ctx, d, meta)
}

func resourceRTXEthernetFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	number, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter number: %v", err)
	}

	log.Printf("[DEBUG] Deleting Ethernet filter: %d", number)

	err = apiClient.client.DeleteEthernetFilter(ctx, number)
	if err != nil {
		return diag.Errorf("Failed to delete Ethernet filter: %v", err)
	}

	d.SetId("")

	return nil
}

func resourceRTXEthernetFilterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	number, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid filter number: %v", err)
	}

	if number < 1 || number > 512 {
		return nil, fmt.Errorf("filter number must be between 1 and 512, got: %d", number)
	}

	log.Printf("[DEBUG] Importing Ethernet filter: %d", number)

	d.Set("number", number)

	return []*schema.ResourceData{d}, nil
}

func buildEthernetFilterFromResourceData(d *schema.ResourceData) client.EthernetFilter {
	filter := client.EthernetFilter{
		Number:    d.Get("number").(int),
		Action:    d.Get("action").(string),
		SourceMAC: d.Get("source_mac").(string),
		DestMAC:   d.Get("destination_mac").(string),
	}

	if v, ok := d.GetOk("ether_type"); ok {
		filter.EtherType = v.(string)
	}

	if v, ok := d.GetOk("vlan_id"); ok {
		filter.VlanID = v.(int)
	}

	return filter
}

func flattenEthernetFilterToResourceData(filter *client.EthernetFilter, d *schema.ResourceData) error {
	d.Set("number", filter.Number)
	d.Set("action", filter.Action)

	if filter.SourceMAC != "" {
		d.Set("source_mac", filter.SourceMAC)
	}
	if filter.DestMAC != "" {
		d.Set("destination_mac", filter.DestMAC)
	}
	if filter.EtherType != "" {
		d.Set("ether_type", filter.EtherType)
	}
	if filter.VlanID > 0 {
		d.Set("vlan_id", filter.VlanID)
	}

	return nil
}
