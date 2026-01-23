package provider

import (
	"context"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXIPv6FilterDynamic() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPv6 dynamic (stateful) IP filters on RTX routers. Dynamic filters provide stateful packet inspection for protocols like FTP, HTTP, and SMTP.",
		CreateContext: resourceRTXIPv6FilterDynamicCreate,
		ReadContext:   resourceRTXIPv6FilterDynamicRead,
		UpdateContext: resourceRTXIPv6FilterDynamicUpdate,
		DeleteContext: resourceRTXIPv6FilterDynamicDelete,
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
							Description:  "Filter number (unique identifier)",
							ValidateFunc: validation.IntAtLeast(1),
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
							Description:  "Protocol: ftp, www, smtp, submission, pop3, dns, telnet, ssh, tcp, udp, or *",
							ValidateFunc: validation.StringInSlice([]string{"ftp", "www", "smtp", "submission", "pop3", "dns", "domain", "telnet", "ssh", "tcp", "udp", "*"}, false),
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

func resourceRTXIPv6FilterDynamicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildIPv6FilterDynamicConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msgf("Creating IPv6 filter dynamic: %+v", config)

	err := apiClient.client.CreateIPv6FilterDynamicConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create IPv6 filter dynamic: %v", err)
	}

	// Use "ipv6_filter_dynamic" as the ID since this is a singleton-like resource
	d.SetId("ipv6_filter_dynamic")

	return resourceRTXIPv6FilterDynamicRead(ctx, d, meta)
}

func resourceRTXIPv6FilterDynamicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msg("Reading IPv6 filter dynamic")

	var config *client.IPv6FilterDynamicConfig

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsedFilters := parsedConfig.ExtractIPv6FiltersDynamic()
			if parsedFilters != nil {
				config = convertParsedIPv6FiltersDynamic(parsedFilters)
				logger.Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msgf("Found %d IPv6 dynamic filters in SFTP cache", len(config.Entries))
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msg("IPv6 dynamic filters not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = apiClient.client.GetIPv6FilterDynamicConfig(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msg("IPv6 filter dynamic not found, removing from state")
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read IPv6 filter dynamic: %v", err)
		}
	}

	entries := flattenIPv6FilterDynamicEntries(config.Entries)
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// convertParsedIPv6FiltersDynamic converts parser IPFilterDynamic slice to client IPv6FilterDynamicConfig
func convertParsedIPv6FiltersDynamic(parsed []parsers.IPFilterDynamic) *client.IPv6FilterDynamicConfig {
	config := &client.IPv6FilterDynamicConfig{
		Entries: make([]client.IPv6FilterDynamicEntry, len(parsed)),
	}
	for i, p := range parsed {
		config.Entries[i] = client.IPv6FilterDynamicEntry{
			Number:   p.Number,
			Source:   p.Source,
			Dest:     p.Dest,
			Protocol: p.Protocol,
			Syslog:   p.SyslogOn,
		}
	}
	return config
}

func resourceRTXIPv6FilterDynamicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildIPv6FilterDynamicConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msgf("Updating IPv6 filter dynamic: %+v", config)

	err := apiClient.client.UpdateIPv6FilterDynamicConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update IPv6 filter dynamic: %v", err)
	}

	return resourceRTXIPv6FilterDynamicRead(ctx, d, meta)
}

func resourceRTXIPv6FilterDynamicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_filter_dynamic").Msg("Deleting IPv6 filter dynamic configuration")

	err := apiClient.client.DeleteIPv6FilterDynamicConfig(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IPv6 filter dynamic: %v", err)
	}

	return nil
}

func buildIPv6FilterDynamicConfigFromResourceData(d *schema.ResourceData) client.IPv6FilterDynamicConfig {
	entries := d.Get("entry").([]interface{})
	config := client.IPv6FilterDynamicConfig{
		Entries: make([]client.IPv6FilterDynamicEntry, 0, len(entries)),
	}

	for _, e := range entries {
		entry := e.(map[string]interface{})
		config.Entries = append(config.Entries, client.IPv6FilterDynamicEntry{
			Number:   entry["number"].(int),
			Source:   entry["source"].(string),
			Dest:     entry["destination"].(string),
			Protocol: entry["protocol"].(string),
			Syslog:   entry["syslog"].(bool),
		})
	}

	return config
}

func flattenIPv6FilterDynamicEntries(entries []client.IPv6FilterDynamicEntry) []interface{} {
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
