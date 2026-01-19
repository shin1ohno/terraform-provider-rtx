package provider

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXIPFilterDynamic() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPv4 dynamic (stateful) IP filters on RTX routers. Dynamic filters provide stateful packet inspection for protocols like FTP, HTTP, and SMTP.",
		CreateContext: resourceRTXIPFilterDynamicCreate,
		ReadContext:   resourceRTXIPFilterDynamicRead,
		UpdateContext: resourceRTXIPFilterDynamicUpdate,
		DeleteContext: resourceRTXIPFilterDynamicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"entry": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of dynamic filter entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Filter number (unique identifier, 1-65535)",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"source": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Source address or '*' for any",
						},
						"destination": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Destination address or '*' for any",
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Protocol: ftp, www, smtp, pop3, dns, telnet, ssh, tcp, udp, or *",
							ValidateFunc: validation.StringInSlice([]string{"ftp", "www", "smtp", "pop3", "dns", "domain", "telnet", "ssh", "tcp", "udp", "*"}, false),
						},
						"syslog": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable syslog for this filter",
						},
					},
				},
			},
		},
	}
}

func resourceRTXIPFilterDynamicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildIPFilterDynamicConfigFromResourceData(d)

	log.Printf("[DEBUG] Creating IP filter dynamic: %+v", config)

	err := apiClient.client.CreateIPFilterDynamicConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create IP filter dynamic: %v", err)
	}

	// Use "ip_filter_dynamic" as the ID since this is a singleton-like resource
	d.SetId("ip_filter_dynamic")

	return resourceRTXIPFilterDynamicRead(ctx, d, meta)
}

func resourceRTXIPFilterDynamicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Reading IP filter dynamic")

	config, err := apiClient.client.GetIPFilterDynamicConfig(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] IP filter dynamic not found, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IP filter dynamic: %v", err)
	}

	entries := flattenIPFilterDynamicEntries(config.Entries)
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPFilterDynamicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildIPFilterDynamicConfigFromResourceData(d)

	log.Printf("[DEBUG] Updating IP filter dynamic: %+v", config)

	err := apiClient.client.UpdateIPFilterDynamicConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update IP filter dynamic: %v", err)
	}

	return resourceRTXIPFilterDynamicRead(ctx, d, meta)
}

func resourceRTXIPFilterDynamicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Deleting IP filter dynamic configuration")

	err := apiClient.client.DeleteIPFilterDynamicConfig(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IP filter dynamic: %v", err)
	}

	return nil
}

func buildIPFilterDynamicConfigFromResourceData(d *schema.ResourceData) client.IPFilterDynamicConfig {
	entries := d.Get("entry").([]interface{})
	config := client.IPFilterDynamicConfig{
		Entries: make([]client.IPFilterDynamicEntry, 0, len(entries)),
	}

	for _, e := range entries {
		entry := e.(map[string]interface{})
		config.Entries = append(config.Entries, client.IPFilterDynamicEntry{
			Number:   entry["number"].(int),
			Source:   entry["source"].(string),
			Dest:     entry["destination"].(string),
			Protocol: entry["protocol"].(string),
			Syslog:   entry["syslog"].(bool),
		})
	}

	return config
}

func flattenIPFilterDynamicEntries(entries []client.IPFilterDynamicEntry) []interface{} {
	result := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		e := map[string]interface{}{
			"number":      entry.Number,
			"source":      entry.Source,
			"destination": entry.Dest,
			"protocol":    entry.Protocol,
			"syslog":      entry.Syslog,
		}
		result = append(result, e)
	}

	return result
}
