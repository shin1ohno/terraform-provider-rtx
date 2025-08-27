package provider

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRTXDHCPScope() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get DHCP scope configurations from an RTX router.",

		ReadContext: dataSourceRTXDHCPScopeRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal identifier for this data source.",
			},
			"scopes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of DHCP scopes configured on the RTX router.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scope_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The DHCP scope ID.",
						},
						"range_start": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The start IP address of the DHCP range.",
						},
						"range_end": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The end IP address of the DHCP range.",
						},
						"prefix": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The network prefix length (e.g., 24 for /24).",
						},
						"gateway": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The gateway IP address for this scope.",
						},
						"dns_servers": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of DNS server IP addresses for this scope.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"lease": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The lease time in hours for this scope (0 if not specified).",
						},
						"domain_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The domain name for this scope.",
						},
					},
				},
			},
		},
	}
}

func dataSourceRTXDHCPScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiClient := meta.(*apiClient)

	// Get DHCP scope information from the router
	scopes, err := apiClient.client.GetDHCPScopes(ctx)
	if err != nil {
		return diag.Errorf("Failed to retrieve DHCP scopes information: %v", err)
	}

	// Convert scopes to schema format
	scopesData := make([]interface{}, len(scopes))
	for i, scope := range scopes {
		scopeMap := map[string]interface{}{
			"scope_id":    scope.ID,
			"range_start": scope.RangeStart,
			"range_end":   scope.RangeEnd,
			"prefix":      scope.Prefix,
			"gateway":     scope.Gateway,
			"dns_servers": scope.DNSServers,
			"lease":       scope.Lease,
			"domain_name": scope.DomainName,
		}

		scopesData[i] = scopeMap
	}

	// Set the resource data
	if err := d.Set("scopes", scopesData); err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID based on the scopes information
	h := md5.New()
	for _, scope := range scopes {
		dnsServers := ""
		for j, dns := range scope.DNSServers {
			if j > 0 {
				dnsServers += ","
			}
			dnsServers += dns
		}
		h.Write([]byte(fmt.Sprintf("%d-%s-%s-%d-%s-%s-%d-%s",
			scope.ID,
			scope.RangeStart,
			scope.RangeEnd,
			scope.Prefix,
			scope.Gateway,
			dnsServers,
			scope.Lease,
			scope.DomainName,
		)))
	}
	id := fmt.Sprintf("%x", h.Sum(nil))
	d.SetId(id)

	return diags
}