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

func resourceRTXIPFilterDynamic() *schema.Resource {
	return &schema.Resource{
		Description: `Manages an IPv4 dynamic (stateful) IP filter on RTX routers.

Dynamic filters provide stateful packet inspection for various protocols. Two forms are supported:

**Form 1 (Protocol-based):** Specify source, destination, and protocol for stateful inspection.
` + "```" + `hcl
resource "rtx_ip_filter_dynamic" "http" {
  filter_id   = 100
  source      = "*"
  destination = "*"
  protocol    = "www"
  syslog      = true
}
` + "```" + `

**Form 2 (Filter-reference):** Reference other static IP filters for complex rules.
` + "```" + `hcl
resource "rtx_ip_filter_dynamic" "custom" {
  filter_id       = 200
  source          = "*"
  destination     = "*"
  filter_list     = [1000, 1001]
  in_filter_list  = [2000]
  out_filter_list = [3000]
}
` + "```" + `
`,
		CreateContext: resourceRTXIPFilterDynamicCreate,
		ReadContext:   resourceRTXIPFilterDynamicRead,
		UpdateContext: resourceRTXIPFilterDynamicUpdate,
		DeleteContext: resourceRTXIPFilterDynamicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXIPFilterDynamicImport,
		},

		Schema: map[string]*schema.Schema{
			"filter_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Filter number (1-65535). This uniquely identifies the dynamic filter on the router.",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"source": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source address or '*' for any. Can be an IP address, network in CIDR notation, or '*'.",
			},
			"destination": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Destination address or '*' for any. Can be an IP address, network in CIDR notation, or '*'.",
			},
			// Form 1: Protocol-based dynamic filter
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Protocol for stateful inspection (Form 1). Valid values: ftp, www, smtp, pop3, dns, domain, " +
					"telnet, ssh, tcp, udp, *, tftp, submission, https, imap, imaps, pop3s, smtps, ldap, ldaps, bgp, sip, " +
					"ipsec-nat-t, ntp, snmp, rtsp, h323, pptp, l2tp, ike, esp. Cannot be used with filter_list.",
				ConflictsWith: []string{"filter_list", "in_filter_list", "out_filter_list"},
				ValidateFunc: validation.StringInSlice([]string{
					"ftp", "www", "smtp", "pop3", "dns", "domain", "telnet", "ssh",
					"tcp", "udp", "*",
					"tftp", "submission", "https", "imap", "imaps", "pop3s", "smtps",
					"ldap", "ldaps", "bgp", "sip", "ipsec-nat-t", "ntp", "snmp",
					"rtsp", "h323", "pptp", "l2tp", "ike", "esp",
				}, false),
			},
			// Form 2: Filter-reference based dynamic filter
			"filter_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of static filter numbers to reference (Form 2). Cannot be used with protocol.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 65535),
				},
				ConflictsWith: []string{"protocol"},
			},
			"in_filter_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of inbound filter numbers (Form 2). Used with filter_list.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 65535),
				},
				ConflictsWith: []string{"protocol"},
			},
			"out_filter_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of outbound filter numbers (Form 2). Used with filter_list.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 65535),
				},
				ConflictsWith: []string{"protocol"},
			},
			// Common options
			"syslog": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable syslog logging for this filter.",
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Timeout value in seconds. If not specified, uses system default.",
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func resourceRTXIPFilterDynamicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	filter, err := buildIPFilterDynamicFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Creating dynamic IP filter: %+v", filter)

	err = apiClient.client.CreateIPFilterDynamic(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to create dynamic IP filter: %v", err)
	}

	// Set ID as the filter number
	d.SetId(strconv.Itoa(filter.Number))

	return resourceRTXIPFilterDynamicRead(ctx, d, meta)
}

func resourceRTXIPFilterDynamicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter ID: %v", err)
	}

	log.Printf("[DEBUG] Reading dynamic IP filter %d", filterID)

	filter, err := apiClient.client.GetIPFilterDynamic(ctx, filterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] Dynamic IP filter %d not found, removing from state", filterID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read dynamic IP filter: %v", err)
	}

	// Set the state from the retrieved filter
	if err := flattenIPFilterDynamicToResourceData(filter, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPFilterDynamicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	filter, err := buildIPFilterDynamicFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Updating dynamic IP filter: %+v", filter)

	// For RTX routers, update is done by re-creating the filter with the same number
	err = apiClient.client.CreateIPFilterDynamic(ctx, filter)
	if err != nil {
		return diag.Errorf("Failed to update dynamic IP filter: %v", err)
	}

	return resourceRTXIPFilterDynamicRead(ctx, d, meta)
}

func resourceRTXIPFilterDynamicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	filterID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid filter ID: %v", err)
	}

	log.Printf("[DEBUG] Deleting dynamic IP filter %d", filterID)

	err = apiClient.client.DeleteIPFilterDynamic(ctx, filterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Already deleted
			return nil
		}
		return diag.Errorf("Failed to delete dynamic IP filter: %v", err)
	}

	return nil
}

func resourceRTXIPFilterDynamicImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse the import ID as filter number
	filterID, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected filter number (integer): %v", err)
	}

	log.Printf("[DEBUG] Importing dynamic IP filter %d", filterID)

	// Retrieve the filter to verify it exists
	filter, err := apiClient.client.GetIPFilterDynamic(ctx, filterID)
	if err != nil {
		return nil, fmt.Errorf("failed to import dynamic IP filter %d: %v", filterID, err)
	}

	// Set the resource ID
	d.SetId(strconv.Itoa(filterID))

	// Set the filter_id explicitly for import
	if err := d.Set("filter_id", filterID); err != nil {
		return nil, fmt.Errorf("failed to set filter_id: %v", err)
	}

	// Flatten the filter data to resource
	if err := flattenIPFilterDynamicToResourceData(filter, d); err != nil {
		return nil, fmt.Errorf("failed to import dynamic IP filter: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

// buildIPFilterDynamicFromResourceData creates an IPFilterDynamic from Terraform resource data
func buildIPFilterDynamicFromResourceData(d *schema.ResourceData) (client.IPFilterDynamic, error) {
	filter := client.IPFilterDynamic{
		Number:   d.Get("filter_id").(int),
		Source:   d.Get("source").(string),
		Dest:     d.Get("destination").(string),
		SyslogOn: d.Get("syslog").(bool),
	}

	// Check for Form 1 (protocol-based) or Form 2 (filter-reference)
	protocol := d.Get("protocol").(string)
	filterList := expandIntList(d.Get("filter_list").([]interface{}))

	if protocol != "" {
		// Form 1: Protocol-based
		filter.Protocol = protocol
	} else if len(filterList) > 0 {
		// Form 2: Filter-reference
		filter.FilterList = filterList
		filter.InFilterList = expandIntList(d.Get("in_filter_list").([]interface{}))
		filter.OutFilterList = expandIntList(d.Get("out_filter_list").([]interface{}))
	} else {
		return filter, fmt.Errorf("either 'protocol' or 'filter_list' must be specified")
	}

	// Handle optional timeout
	if v, ok := d.GetOk("timeout"); ok {
		timeout := v.(int)
		filter.Timeout = &timeout
	}

	return filter, nil
}

// flattenIPFilterDynamicToResourceData sets Terraform resource data from an IPFilterDynamic
func flattenIPFilterDynamicToResourceData(filter *client.IPFilterDynamic, d *schema.ResourceData) error {
	if err := d.Set("filter_id", filter.Number); err != nil {
		return fmt.Errorf("failed to set filter_id: %w", err)
	}
	if err := d.Set("source", filter.Source); err != nil {
		return fmt.Errorf("failed to set source: %w", err)
	}
	if err := d.Set("destination", filter.Dest); err != nil {
		return fmt.Errorf("failed to set destination: %w", err)
	}
	if err := d.Set("syslog", filter.SyslogOn); err != nil {
		return fmt.Errorf("failed to set syslog: %w", err)
	}

	// Determine if this is Form 1 or Form 2
	if len(filter.FilterList) > 0 {
		// Form 2: Filter-reference
		if err := d.Set("filter_list", filter.FilterList); err != nil {
			return fmt.Errorf("failed to set filter_list: %w", err)
		}
		if err := d.Set("in_filter_list", filter.InFilterList); err != nil {
			return fmt.Errorf("failed to set in_filter_list: %w", err)
		}
		if err := d.Set("out_filter_list", filter.OutFilterList); err != nil {
			return fmt.Errorf("failed to set out_filter_list: %w", err)
		}
		// Clear protocol for Form 2
		if err := d.Set("protocol", ""); err != nil {
			return fmt.Errorf("failed to clear protocol: %w", err)
		}
	} else {
		// Form 1: Protocol-based
		if err := d.Set("protocol", filter.Protocol); err != nil {
			return fmt.Errorf("failed to set protocol: %w", err)
		}
		// Clear filter lists for Form 1
		if err := d.Set("filter_list", nil); err != nil {
			return fmt.Errorf("failed to clear filter_list: %w", err)
		}
		if err := d.Set("in_filter_list", nil); err != nil {
			return fmt.Errorf("failed to clear in_filter_list: %w", err)
		}
		if err := d.Set("out_filter_list", nil); err != nil {
			return fmt.Errorf("failed to clear out_filter_list: %w", err)
		}
	}

	// Handle timeout
	if filter.Timeout != nil {
		if err := d.Set("timeout", *filter.Timeout); err != nil {
			return fmt.Errorf("failed to set timeout: %w", err)
		}
	}

	return nil
}

// expandIntList converts a []interface{} to []int
func expandIntList(input []interface{}) []int {
	result := make([]int, 0, len(input))
	for _, v := range input {
		result = append(result, v.(int))
	}
	return result
}
