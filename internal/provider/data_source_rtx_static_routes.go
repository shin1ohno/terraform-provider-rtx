package provider

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRTXStaticRoutes() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get static route configurations from an RTX router.",

		ReadContext: dataSourceRTXStaticRoutesRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal identifier for this data source.",
			},
			"routes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of static routes configured on the RTX router.",
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
							Description: "The next hop gateway IP address (optional).",
						},
						"interface": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The outgoing interface for this route (optional).",
						},
						"metric": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The route metric (cost).",
						},
						"weight": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The route weight for load balancing.",
						},
						"hide": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this route is hidden from advertisements.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A description of the static route (optional).",
						},
					},
				},
			},
		},
	}
}

func dataSourceRTXStaticRoutesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiClient := meta.(*apiClient)

	// Get static routes information from the router
	staticRoutes, err := apiClient.client.GetStaticRoutes(ctx)
	if err != nil {
		return diag.Errorf("Failed to retrieve static routes information: %v", err)
	}

	// Convert static routes to schema format
	routesData := make([]interface{}, len(staticRoutes))
	for i, route := range staticRoutes {
		// Determine gateway field from GatewayIP or GatewayInterface
		gateway := route.GatewayIP
		if gateway == "" && route.GatewayInterface != "" {
			gateway = route.GatewayInterface
		}

		routeMap := map[string]interface{}{
			"destination": route.Destination,
			"gateway":     gateway,
			"interface":   route.Interface,
			"metric":      route.Metric,
			"weight":      route.Weight,
			"hide":        route.Hide,
			"description": route.Description,
		}

		routesData[i] = routeMap
	}

	// Set the resource data
	if err := d.Set("routes", routesData); err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID based on the static routes information
	h := md5.New()
	for _, route := range staticRoutes {
		// Determine gateway field from GatewayIP or GatewayInterface
		gateway := route.GatewayIP
		if gateway == "" && route.GatewayInterface != "" {
			gateway = route.GatewayInterface
		}

		h.Write([]byte(fmt.Sprintf("%s-%s-%s-%d-%d-%v-%s",
			route.Destination,
			gateway,
			route.Interface,
			route.Metric,
			route.Weight,
			route.Hide,
			route.Description,
		)))
	}
	id := fmt.Sprintf("%x", h.Sum(nil))
	d.SetId(id)

	return diags
}