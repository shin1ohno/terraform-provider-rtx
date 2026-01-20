package provider

import (
	"context"
	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceRTXDDNSStatus() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves DDNS registration status information from RTX routers. This includes status for both NetVolante DNS and custom DDNS providers.",
		ReadContext: dataSourceRTXDDNSStatusRead,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"netvolante", "custom", "all"}, false),
				Default:      "all",
				Description:  "Type of DDNS status to retrieve: 'netvolante', 'custom', or 'all' (default).",
			},
			"statuses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of DDNS status entries.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "DDNS type: 'netvolante' or 'custom'.",
						},
						"interface": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Interface (for NetVolante DNS).",
						},
						"server_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Server ID (for custom DDNS).",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Registered hostname.",
						},
						"current_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Currently registered IP address.",
						},
						"last_update": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last successful update timestamp.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status: 'success', 'error', or 'pending'.",
						},
						"error_message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Error message if status is 'error'.",
						},
					},
				},
			},
		},
	}
}

func dataSourceRTXDDNSStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	statusType := d.Get("type").(string)
	logging.FromContext(ctx).Debug().Str("resource", "rtx_ddns_status").Msgf("Reading DDNS status: type=%s", statusType)

	var statuses []map[string]interface{}

	// Fetch NetVolante status
	if statusType == "all" || statusType == "netvolante" {
		nvStatuses, err := apiClient.client.GetNetVolanteDNSStatus(ctx)
		if err != nil {
			return diag.Errorf("Failed to read NetVolante DNS status: %v", err)
		}
		for _, s := range nvStatuses {
			statuses = append(statuses, map[string]interface{}{
				"type":          s.Type,
				"interface":     s.Interface,
				"server_id":     0,
				"hostname":      s.Hostname,
				"current_ip":    s.CurrentIP,
				"last_update":   s.LastUpdate,
				"status":        s.Status,
				"error_message": s.ErrorMessage,
			})
		}
	}

	// Fetch custom DDNS status
	if statusType == "all" || statusType == "custom" {
		ddnsStatuses, err := apiClient.client.GetDDNSStatus(ctx)
		if err != nil {
			return diag.Errorf("Failed to read custom DDNS status: %v", err)
		}
		for _, s := range ddnsStatuses {
			statuses = append(statuses, map[string]interface{}{
				"type":          s.Type,
				"interface":     "",
				"server_id":     s.ServerID,
				"hostname":      s.Hostname,
				"current_ip":    s.CurrentIP,
				"last_update":   s.LastUpdate,
				"status":        s.Status,
				"error_message": s.ErrorMessage,
			})
		}
	}

	if err := d.Set("statuses", statuses); err != nil {
		return diag.FromErr(err)
	}

	// Use static ID for data source
	d.SetId("ddns_status")

	return nil
}
