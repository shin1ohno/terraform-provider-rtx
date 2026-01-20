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

func resourceRTXDNSServer() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages DNS server configuration on RTX routers. This is a singleton resource - there is only one DNS server configuration per router.",
		CreateContext: resourceRTXDNSServerCreate,
		ReadContext:   resourceRTXDNSServerRead,
		UpdateContext: resourceRTXDNSServerUpdate,
		DeleteContext: resourceRTXDNSServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXDNSServerImport,
		},

		Schema: map[string]*schema.Schema{
			"domain_lookup": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable DNS domain lookup (dns domain lookup on/off)",
			},
			"domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default domain name for DNS queries (dns domain <name>)",
			},
			"name_servers": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    3,
				Description: "List of DNS server IP addresses (up to 3)",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIPAddress,
				},
			},
			"server_select": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Domain-based DNS server selection entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Selector ID (positive integer)",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"servers": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "DNS server IP addresses for this selector (IPv4 or IPv6)",
							MinItems:    1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateIPAddressAny,
							},
						},
						"edns": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable EDNS (Extension mechanisms for DNS)",
						},
						"record_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "DNS record type to match: a, aaaa, ptr, mx, ns, cname, any",
							ValidateFunc: validation.StringInSlice([]string{"a", "aaaa", "ptr", "mx", "ns", "cname", "any"}, false),
						},
						"query_pattern": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Domain pattern to match (e.g., '.', '*.example.com', 'internal.net')",
						},
						"original_sender": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source IP/CIDR restriction for DNS queries",
						},
						"restrict_pp": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "PP session restriction (0 = no restriction)",
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"hosts": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Static DNS host entries (dns static)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Hostname",
						},
						"address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IP address",
							ValidateFunc: validateIPAddress,
						},
					},
				},
			},
			"service_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable DNS service (dns service on/off)",
			},
			"private_address_spoof": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable DNS private address spoofing (dns private address spoof on/off)",
			},
		},
	}
}

func resourceRTXDNSServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildDNSConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dns_server").Msgf("Creating DNS server configuration: %+v", config)

	err := apiClient.client.ConfigureDNS(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure DNS server: %v", err)
	}

	// Use fixed ID "dns" for singleton resource
	d.SetId("dns")

	// Read back to ensure consistency
	return resourceRTXDNSServerRead(ctx, d, meta)
}

func resourceRTXDNSServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dns_server").Msg("Reading DNS server configuration")

	config, err := apiClient.client.GetDNS(ctx)
	if err != nil {
		return diag.Errorf("Failed to read DNS server configuration: %v", err)
	}

	// Update the state
	if err := d.Set("domain_lookup", config.DomainLookup); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain_name", config.DomainName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name_servers", config.NameServers); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_on", config.ServiceOn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_address_spoof", config.PrivateSpoof); err != nil {
		return diag.FromErr(err)
	}

	// Convert ServerSelect to list
	serverSelects := make([]map[string]interface{}, len(config.ServerSelect))
	for i, sel := range config.ServerSelect {
		serverSelects[i] = map[string]interface{}{
			"id":              sel.ID,
			"servers":         sel.Servers,
			"edns":            sel.EDNS,
			"record_type":     sel.RecordType,
			"query_pattern":   sel.QueryPattern,
			"original_sender": sel.OriginalSender,
			"restrict_pp":     sel.RestrictPP,
		}
	}
	if err := d.Set("server_select", serverSelects); err != nil {
		return diag.FromErr(err)
	}

	// Convert Hosts to list
	hosts := make([]map[string]interface{}, len(config.Hosts))
	for i, host := range config.Hosts {
		hosts[i] = map[string]interface{}{
			"name":    host.Name,
			"address": host.Address,
		}
	}
	if err := d.Set("hosts", hosts); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXDNSServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildDNSConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dns_server").Msgf("Updating DNS server configuration: %+v", config)

	err := apiClient.client.UpdateDNS(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update DNS server configuration: %v", err)
	}

	return resourceRTXDNSServerRead(ctx, d, meta)
}

func resourceRTXDNSServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dns_server").Msg("Deleting (resetting) DNS server configuration")

	err := apiClient.client.ResetDNS(ctx)
	if err != nil {
		return diag.Errorf("Failed to reset DNS server configuration: %v", err)
	}

	return nil
}

func resourceRTXDNSServerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Only accept "dns" as valid import ID (singleton resource)
	if importID != "dns" {
		return nil, fmt.Errorf("invalid import ID format, expected 'dns' for singleton resource, got: %s", importID)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dns_server").Msg("Importing DNS server configuration")

	// Verify configuration exists and retrieve it
	config, err := apiClient.client.GetDNS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import DNS server configuration: %v", err)
	}

	// Set the ID
	d.SetId("dns")

	// Set all attributes
	d.Set("domain_lookup", config.DomainLookup)
	d.Set("domain_name", config.DomainName)
	d.Set("name_servers", config.NameServers)
	d.Set("service_on", config.ServiceOn)
	d.Set("private_address_spoof", config.PrivateSpoof)

	// Set server_select
	serverSelects := make([]map[string]interface{}, len(config.ServerSelect))
	for i, sel := range config.ServerSelect {
		serverSelects[i] = map[string]interface{}{
			"id":              sel.ID,
			"servers":         sel.Servers,
			"edns":            sel.EDNS,
			"record_type":     sel.RecordType,
			"query_pattern":   sel.QueryPattern,
			"original_sender": sel.OriginalSender,
			"restrict_pp":     sel.RestrictPP,
		}
	}
	d.Set("server_select", serverSelects)

	// Set hosts
	hosts := make([]map[string]interface{}, len(config.Hosts))
	for i, host := range config.Hosts {
		hosts[i] = map[string]interface{}{
			"name":    host.Name,
			"address": host.Address,
		}
	}
	d.Set("hosts", hosts)

	return []*schema.ResourceData{d}, nil
}

// buildDNSConfigFromResourceData creates a DNSConfig from Terraform resource data
func buildDNSConfigFromResourceData(d *schema.ResourceData) client.DNSConfig {
	config := client.DNSConfig{
		DomainLookup: d.Get("domain_lookup").(bool),
		DomainName:   d.Get("domain_name").(string),
		ServiceOn:    d.Get("service_on").(bool),
		PrivateSpoof: d.Get("private_address_spoof").(bool),
		NameServers:  []string{},
		ServerSelect: []client.DNSServerSelect{},
		Hosts:        []client.DNSHost{},
	}

	// Handle name_servers list
	if v, ok := d.GetOk("name_servers"); ok {
		nameServersList := v.([]interface{})
		for _, ns := range nameServersList {
			config.NameServers = append(config.NameServers, ns.(string))
		}
	}

	// Handle server_select list
	if v, ok := d.GetOk("server_select"); ok {
		serverSelectList := v.([]interface{})
		for _, selRaw := range serverSelectList {
			selMap := selRaw.(map[string]interface{})

			// Extract servers
			servers := []string{}
			if serversRaw, ok := selMap["servers"].([]interface{}); ok {
				for _, s := range serversRaw {
					servers = append(servers, s.(string))
				}
			}

			// Extract new fields
			edns := false
			if v, ok := selMap["edns"].(bool); ok {
				edns = v
			}

			recordType := "a"
			if v, ok := selMap["record_type"].(string); ok && v != "" {
				recordType = v
			}

			queryPattern := ""
			if v, ok := selMap["query_pattern"].(string); ok {
				queryPattern = v
			}

			originalSender := ""
			if v, ok := selMap["original_sender"].(string); ok {
				originalSender = v
			}

			restrictPP := 0
			if v, ok := selMap["restrict_pp"].(int); ok {
				restrictPP = v
			}

			config.ServerSelect = append(config.ServerSelect, client.DNSServerSelect{
				ID:             selMap["id"].(int),
				Servers:        servers,
				EDNS:           edns,
				RecordType:     recordType,
				QueryPattern:   queryPattern,
				OriginalSender: originalSender,
				RestrictPP:     restrictPP,
			})
		}
	}

	// Handle hosts list
	if v, ok := d.GetOk("hosts"); ok {
		hostsList := v.([]interface{})
		for _, hostRaw := range hostsList {
			hostMap := hostRaw.(map[string]interface{})
			config.Hosts = append(config.Hosts, client.DNSHost{
				Name:    hostMap["name"].(string),
				Address: hostMap["address"].(string),
			})
		}
	}

	return config
}
