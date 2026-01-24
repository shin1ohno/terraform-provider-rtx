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

func resourceRTXAccessListIP() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a static IP filter (access list) on RTX routers. This creates an individual IP filter rule using the RTX native 'ip filter' command.",
		CreateContext: resourceRTXAccessListIPCreate,
		ReadContext:   resourceRTXAccessListIPRead,
		UpdateContext: resourceRTXAccessListIPUpdate,
		DeleteContext: resourceRTXAccessListIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListIPImport,
		},

		Schema: map[string]*schema.Schema{
			"filter_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Filter number (1-2147483647). This uniquely identifies the filter on the router.",
				ValidateFunc: validation.IntBetween(1, 2147483647),
			},
			"action": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Filter action: pass, reject, restrict, or restrict-log",
				ValidateFunc:     validation.StringInSlice([]string{"pass", "reject", "restrict", "restrict-log"}, true),
				DiffSuppressFunc: SuppressCaseDiff, // Filter actions are case-insensitive
			},
			"source": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source IP address/network in CIDR notation (e.g., '10.0.0.0/8') or '*' for any",
			},
			"destination": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Destination IP address/network in CIDR notation (e.g., '192.168.1.0/24') or '*' for any",
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Protocol: tcp, udp, icmp, ip, gre, esp, ah, or * for any",
				ValidateFunc:     validation.StringInSlice([]string{"tcp", "udp", "udp,tcp", "tcp,udp", "icmp", "ip", "gre", "esp", "ah", "tcpfin", "tcprst", "*"}, true),
				DiffSuppressFunc: SuppressCaseDiff, // Protocol names are case-insensitive
			},
			"source_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Source port number, range (e.g., '1024-65535'), or '*' for any. Only valid for TCP/UDP.",
			},
			"dest_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Destination port number, range (e.g., '80'), or '*' for any. Only valid for TCP/UDP.",
			},
			"established": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Match established TCP connections only. Only valid for TCP protocol.",
			},
		},
	}
}

func resourceRTXAccessListIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", d.Id())
	filter := buildIPFilterFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip").Msgf("Creating IP filter: %+v", filter)

	err := apiClient.client.CreateIPFilter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to create IP filter: %v", err)
	}

	// Set ID as the filter number
	d.SetId(strconv.Itoa(filter.Number))

	return resourceRTXAccessListIPRead(ctx, d, meta)
}

func resourceRTXAccessListIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", d.Id())
	logger := logging.FromContext(ctx)

	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter ID: %v", err)
	}

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Reading IP filter %d", filterID)

	var filter *client.IPFilter

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract IP filters from parsed config
			filters := parsedConfig.ExtractIPFilters()
			for i := range filters {
				if filters[i].Number == filterID {
					filter = convertParsedIPFilter(&filters[i])
					logger.Debug().Str("resource", "rtx_access_list_ip").Msg("Found filter in SFTP cache")
					break
				}
			}
		}
		if filter == nil {
			// Filter not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_access_list_ip").Msg("Filter not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or filter not found in cache
	if filter == nil {
		filter, err = apiClient.client.GetIPFilter(ctx, filterID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("IP filter %d not found, removing from state", filterID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read IP filter: %v", err)
		}
	}

	// Set the state from the retrieved filter
	if err := flattenIPFilterToResourceData(filter, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// convertParsedIPFilter converts a parser IPFilter to a client IPFilter
func convertParsedIPFilter(parsed *parsers.IPFilter) *client.IPFilter {
	return &client.IPFilter{
		Number:        parsed.Number,
		Action:        parsed.Action,
		SourceAddress: parsed.SourceAddress,
		SourceMask:    parsed.SourceMask,
		DestAddress:   parsed.DestAddress,
		DestMask:      parsed.DestMask,
		Protocol:      parsed.Protocol,
		SourcePort:    parsed.SourcePort,
		DestPort:      parsed.DestPort,
		Established:   parsed.Established,
	}
}

func resourceRTXAccessListIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", d.Id())
	filter := buildIPFilterFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip").Msgf("Updating IP filter: %+v", filter)

	err := apiClient.client.UpdateIPFilter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to update IP filter: %v", err)
	}

	return resourceRTXAccessListIPRead(ctx, d, meta)
}

func resourceRTXAccessListIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", d.Id())
	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip").Msgf("Deleting IP filter %d", filterID)

	err = apiClient.client.DeleteIPFilter(ctx, filterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Already deleted
			return nil
		}
		return diag.Errorf("Failed to delete IP filter: %v", err)
	}

	return nil
}

func resourceRTXAccessListIPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse the import ID as filter number
	filterID, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected filter number (integer): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip").Msgf("Importing IP filter %d", filterID)

	// Retrieve the filter to verify it exists
	filter, err := apiClient.client.GetIPFilter(ctx, filterID)
	if err != nil {
		return nil, fmt.Errorf("failed to import IP filter %d: %v", filterID, err)
	}

	// Set the resource ID
	d.SetId(strconv.Itoa(filterID))

	// Set the filter_id explicitly for import
	if err := d.Set("filter_id", filterID); err != nil {
		return nil, fmt.Errorf("failed to set filter_id: %v", err)
	}

	// Flatten the filter data to resource
	if err := flattenIPFilterToResourceData(filter, d); err != nil {
		return nil, fmt.Errorf("failed to import IP filter: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

// buildIPFilterFromResourceData creates an IPFilter from Terraform resource data
func buildIPFilterFromResourceData(d *schema.ResourceData) client.IPFilter {
	filter := client.IPFilter{
		Number:        d.Get("filter_id").(int),
		Action:        d.Get("action").(string),
		SourceAddress: d.Get("source").(string),
		DestAddress:   d.Get("destination").(string),
		Protocol:      d.Get("protocol").(string),
		SourcePort:    d.Get("source_port").(string),
		DestPort:      d.Get("dest_port").(string),
		Established:   d.Get("established").(bool),
	}

	return filter
}

// flattenIPFilterToResourceData sets Terraform resource data from an IPFilter
func flattenIPFilterToResourceData(filter *client.IPFilter, d *schema.ResourceData) error {
	if err := d.Set("filter_id", filter.Number); err != nil {
		return fmt.Errorf("failed to set filter_id: %w", err)
	}
	if err := d.Set("action", filter.Action); err != nil {
		return fmt.Errorf("failed to set action: %w", err)
	}
	if err := d.Set("source", filter.SourceAddress); err != nil {
		return fmt.Errorf("failed to set source: %w", err)
	}
	if err := d.Set("destination", filter.DestAddress); err != nil {
		return fmt.Errorf("failed to set destination: %w", err)
	}
	if err := d.Set("protocol", filter.Protocol); err != nil {
		return fmt.Errorf("failed to set protocol: %w", err)
	}

	// Set source_port - use empty string for "*" to avoid perpetual diff
	sourcePort := filter.SourcePort
	if sourcePort == "" {
		sourcePort = "*"
	}
	if err := d.Set("source_port", sourcePort); err != nil {
		return fmt.Errorf("failed to set source_port: %w", err)
	}

	// Set dest_port - use empty string for "*" to avoid perpetual diff
	destPort := filter.DestPort
	if destPort == "" {
		destPort = "*"
	}
	if err := d.Set("dest_port", destPort); err != nil {
		return fmt.Errorf("failed to set dest_port: %w", err)
	}

	if err := d.Set("established", filter.Established); err != nil {
		return fmt.Errorf("failed to set established: %w", err)
	}

	return nil
}
