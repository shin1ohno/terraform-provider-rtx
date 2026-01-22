package provider

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRTXRoutes() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get routing table information from an RTX router.",

		ReadContext: dataSourceRTXRoutesRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal identifier for this data source.",
			},
			"routes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of routes in the RTX router's routing table.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The destination network prefix (e.g., '192.168.1.0/24', '0.0.0.0/0').",
						},
						"gateway": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The next hop gateway IP address ('*' for directly connected routes).",
						},
						"interface": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The outgoing interface for this route.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The route protocol: S=static, C=connected, R=RIP, O=OSPF, B=BGP, D=DHCP.",
						},
						"metric": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The route metric (cost). May be 0 if not specified.",
						},
					},
				},
			},
		},
	}
}

func dataSourceRTXRoutesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiClient := meta.(*apiClient)

	// Get routes information from the router
	routes, err := apiClient.client.GetRoutes(ctx)
	if err != nil {
		return diag.Errorf("Failed to retrieve routes information: %v", err)
	}

	// Convert routes to schema format
	routesData := make([]interface{}, len(routes))
	for i, route := range routes {
		routeMap := map[string]interface{}{
			"destination": route.Destination,
			"gateway":     route.Gateway,
			"interface":   route.Interface,
			"protocol":    route.Protocol,
		}

		// Handle optional metric field
		if route.Metric != nil {
			routeMap["metric"] = *route.Metric
		} else {
			routeMap["metric"] = 0
		}

		routesData[i] = routeMap
	}

	// Set the resource data
	if err := d.Set("routes", routesData); err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID based on the routes information
	h := md5.New()
	for _, route := range routes {
		metric := 0
		if route.Metric != nil {
			metric = *route.Metric
		}
		h.Write([]byte(fmt.Sprintf("%s-%s-%s-%s-%d",
			route.Destination,
			route.Gateway,
			route.Interface,
			route.Protocol,
			metric,
		)))
	}
	id := fmt.Sprintf("%x", h.Sum(nil))
	d.SetId(id)

	return diags
}
