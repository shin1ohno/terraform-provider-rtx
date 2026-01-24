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

// ValidIPv6FilterProtocols defines valid protocols for IPv6 filters (includes icmp6)
var ValidIPv6FilterProtocols = []string{"tcp", "udp", "icmp6", "ip", "*", "gre", "esp", "ah"}

func resourceRTXAccessListIPv6() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages static IPv6 filters on RTX routers. Static IPv6 filters provide packet filtering based on source/destination addresses, protocols, and ports for IPv6 traffic.",
		CreateContext: resourceRTXAccessListIPv6Create,
		ReadContext:   resourceRTXAccessListIPv6Read,
		UpdateContext: resourceRTXAccessListIPv6Update,
		DeleteContext: resourceRTXAccessListIPv6Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListIPv6Import,
		},

		Schema: map[string]*schema.Schema{
			"sequence": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Sequence number determining evaluation order. Lower numbers are evaluated first (1-2147483647).",
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
				Description: "Source IPv6 address/prefix or '*' for any",
			},
			"destination": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Destination IPv6 address/prefix or '*' for any",
			},
			"protocol": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Protocol: tcp, udp, icmp6, ip, *, gre, esp, ah",
				ValidateFunc:     validation.StringInSlice(ValidIPv6FilterProtocols, true),
				DiffSuppressFunc: SuppressCaseDiff, // Protocol names are case-insensitive
			},
			"source_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Source port(s) or '*' for any. Only applicable for tcp/udp protocols.",
			},
			"dest_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Destination port(s) or '*' for any. Only applicable for tcp/udp protocols.",
			},
		},
	}
}

func resourceRTXAccessListIPv6Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", d.Id())
	filter := buildIPv6FilterFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Creating IPv6 filter: %+v", filter)

	err := apiClient.client.CreateIPv6Filter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to create IPv6 filter: %v", err)
	}

	// Use filter number as the ID
	d.SetId(strconv.Itoa(filter.Number))

	return resourceRTXAccessListIPv6Read(ctx, d, meta)
}

func resourceRTXAccessListIPv6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", d.Id())
	logger := logging.FromContext(ctx)

	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter ID: %v", err)
	}

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Reading IPv6 filter: %d", filterID)

	var filter *client.IPFilter

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract IPv6 filters from parsed config
			filters := parsedConfig.ExtractAccessListIPv6()
			for i := range filters {
				if filters[i].Number == filterID {
					filter = convertParsedIPv6Filter(&filters[i])
					logger.Debug().Str("resource", "rtx_access_list_ipv6").Msg("Found filter in SFTP cache")
					break
				}
			}
		}
		if filter == nil {
			// Filter not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_access_list_ipv6").Msg("Filter not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or filter not found in cache
	if filter == nil {
		filter, err = apiClient.client.GetIPv6Filter(ctx, filterID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("IPv6 filter %d not found, removing from state", filterID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read IPv6 filter: %v", err)
		}
	}

	// Set state from retrieved filter
	if err := d.Set("sequence", filter.Number); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("action", filter.Action); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("source", filter.SourceAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("destination", filter.DestAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("protocol", filter.Protocol); err != nil {
		return diag.FromErr(err)
	}
	if filter.SourcePort != "" {
		if err := d.Set("source_port", filter.SourcePort); err != nil {
			return diag.FromErr(err)
		}
	}
	if filter.DestPort != "" {
		if err := d.Set("dest_port", filter.DestPort); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// convertParsedIPv6Filter converts a parser IPFilter to a client IPFilter for IPv6
func convertParsedIPv6Filter(parsed *parsers.IPFilter) *client.IPFilter {
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

func resourceRTXAccessListIPv6Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", d.Id())
	filter := buildIPv6FilterFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Updating IPv6 filter: %+v", filter)

	err := apiClient.client.UpdateIPv6Filter(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to update IPv6 filter: %v", err)
	}

	return resourceRTXAccessListIPv6Read(ctx, d, meta)
}

func resourceRTXAccessListIPv6Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", d.Id())
	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Deleting IPv6 filter: %d", filterID)

	err = apiClient.client.DeleteIPv6Filter(ctx, filterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IPv6 filter: %v", err)
	}

	return nil
}

func resourceRTXAccessListIPv6Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Import by filter number
	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid filter ID for import: %v", err)
	}

	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Importing IPv6 filter: %d", filterID)

	filter, err := apiClient.client.GetIPv6Filter(ctx, filterID)
	if err != nil {
		return nil, fmt.Errorf("failed to import IPv6 filter %d: %v", filterID, err)
	}

	// Set all attributes
	if err := d.Set("sequence", filter.Number); err != nil {
		return nil, err
	}
	if err := d.Set("action", filter.Action); err != nil {
		return nil, err
	}
	if err := d.Set("source", filter.SourceAddress); err != nil {
		return nil, err
	}
	if err := d.Set("destination", filter.DestAddress); err != nil {
		return nil, err
	}
	if err := d.Set("protocol", filter.Protocol); err != nil {
		return nil, err
	}
	if filter.SourcePort != "" {
		if err := d.Set("source_port", filter.SourcePort); err != nil {
			return nil, err
		}
	}
	if filter.DestPort != "" {
		if err := d.Set("dest_port", filter.DestPort); err != nil {
			return nil, err
		}
	}

	return []*schema.ResourceData{d}, nil
}

func buildIPv6FilterFromResourceData(d *schema.ResourceData) client.IPFilter {
	filter := client.IPFilter{
		Number:        d.Get("sequence").(int),
		Action:        d.Get("action").(string),
		SourceAddress: d.Get("source").(string),
		DestAddress:   d.Get("destination").(string),
		Protocol:      d.Get("protocol").(string),
	}

	if v, ok := d.GetOk("source_port"); ok {
		filter.SourcePort = v.(string)
	}
	if v, ok := d.GetOk("dest_port"); ok {
		filter.DestPort = v.(string)
	}

	return filter
}
