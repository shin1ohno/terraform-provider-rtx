package provider

import (
	"context"
	"fmt"
	"crypto/md5"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRTXSystemInfo() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get system information from an RTX router.",

		ReadContext: dataSourceRTXSystemInfoRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal identifier for this data source.",
			},
			"model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The RTX router model number (e.g., RTX1210, RTX830).",
			},
			"firmware_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The firmware version running on the RTX router.",
			},
			"serial_number": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The serial number of the RTX router.",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The MAC address of the RTX router.",
			},
			"uptime": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The uptime of the RTX router.",
			},
		},
	}
}

func dataSourceRTXSystemInfoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiClient := meta.(*apiClient)
	
	// Get system information from the router
	systemInfo, err := apiClient.client.GetSystemInfo(ctx)
	if err != nil {
		return diag.Errorf("Failed to retrieve system information: %v", err)
	}

	// Set the resource data
	if err := d.Set("model", systemInfo.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("firmware_version", systemInfo.FirmwareVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("serial_number", systemInfo.SerialNumber); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mac_address", systemInfo.MACAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("uptime", systemInfo.Uptime); err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID based on the system information
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s-%s",
		systemInfo.Model,
		systemInfo.FirmwareVersion,
		systemInfo.SerialNumber,
		systemInfo.MACAddress,
	)))
	id := fmt.Sprintf("%x", h.Sum(nil))
	d.SetId(id)

	return diags
}