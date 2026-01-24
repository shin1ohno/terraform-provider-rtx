package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
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
				DiffSuppressFunc: suppressEquivalentEthernetFilterAction,
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

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ethernet_filter", d.Id())
	filter := buildEthernetFilterFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ethernet_filter").Msgf("Creating Ethernet filter: %d", filter.Number)

	err := apiClient.client.CreateEthernetFilter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to create Ethernet filter: %v", err)
	}

	d.SetId(strconv.Itoa(filter.Number))

	return resourceRTXEthernetFilterRead(ctx, d, meta)
}

func resourceRTXEthernetFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ethernet_filter", d.Id())
	logger := logging.FromContext(ctx)

	number, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter number: %v", err)
	}

	logger.Debug().Str("resource", "rtx_ethernet_filter").Msgf("Reading Ethernet filter: %d", number)

	var filter *client.EthernetFilter

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract Ethernet filters from parsed config
			filters := parsedConfig.ExtractEthernetFilters()
			for i := range filters {
				if filters[i].Number == number {
					filter = convertParsedEthernetFilter(&filters[i])
					logger.Debug().Str("resource", "rtx_ethernet_filter").Msg("Found filter in SFTP cache")
					break
				}
			}
		}
		if filter == nil {
			// Filter not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_ethernet_filter").Msg("Filter not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or filter not found in cache
	if filter == nil {
		filter, err = apiClient.client.GetEthernetFilter(ctx, number)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Warn().Str("resource", "rtx_ethernet_filter").Msgf("Ethernet filter %d not found, removing from state", number)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read Ethernet filter: %v", err)
		}
	}

	if err := flattenEthernetFilterToResourceData(filter, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// convertParsedEthernetFilter converts a parser EthernetFilter to a client EthernetFilter
func convertParsedEthernetFilter(parsed *parsers.EthernetFilter) *client.EthernetFilter {
	// Use DestinationMAC if available, otherwise fallback to DestMAC (deprecated field)
	destMAC := parsed.DestinationMAC
	if destMAC == "" {
		destMAC = parsed.DestMAC
	}
	return &client.EthernetFilter{
		Number:    parsed.Number,
		Action:    parsed.Action,
		SourceMAC: parsed.SourceMAC,
		DestMAC:   destMAC,
		EtherType: parsed.EtherType,
		VlanID:    parsed.VlanID,
	}
}

func resourceRTXEthernetFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ethernet_filter", d.Id())
	filter := buildEthernetFilterFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ethernet_filter").Msgf("Updating Ethernet filter: %d", filter.Number)

	err := apiClient.client.UpdateEthernetFilter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to update Ethernet filter: %v", err)
	}

	return resourceRTXEthernetFilterRead(ctx, d, meta)
}

func resourceRTXEthernetFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ethernet_filter", d.Id())
	number, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter number: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ethernet_filter").Msgf("Deleting Ethernet filter: %d", number)

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

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ethernet_filter").Msgf("Importing Ethernet filter: %d", number)

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

// suppressEquivalentEthernetFilterAction suppresses diff when the action values
// are functionally equivalent. RTX routers treat "pass" and "pass-nolog" as identical,
// as well as "reject" and "reject-nolog".
func suppressEquivalentEthernetFilterAction(k, old, new string, d *schema.ResourceData) bool {
	// Normalize both values to canonical form
	oldNorm := normalizeEthernetFilterAction(old)
	newNorm := normalizeEthernetFilterAction(new)
	return oldNorm == newNorm
}

// normalizeEthernetFilterAction normalizes action to canonical form.
// "pass-nolog" -> "pass", "reject-nolog" -> "reject"
func normalizeEthernetFilterAction(action string) string {
	switch action {
	case "pass-nolog":
		return "pass"
	case "reject-nolog":
		return "reject"
	default:
		return action
	}
}
