package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXDDNS() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages custom DDNS provider configuration on RTX routers. Use this resource to configure third-party DDNS services like No-IP, DynDNS, or other compatible providers.",
		CreateContext: resourceRTXDDNSCreate,
		ReadContext:   resourceRTXDDNSRead,
		UpdateContext: resourceRTXDDNSUpdate,
		DeleteContext: resourceRTXDDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXDDNSImport,
		},

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4),
				Description:  "DDNS server ID (1-4). Each ID can be configured with a different DDNS provider.",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "DDNS update URL (must start with http:// or https://). This is the provider's update endpoint.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "DDNS hostname to update (e.g., 'example.no-ip.org').",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DDNS account username for authentication.",
			},
			"password": WriteOnlyStringSchema("DDNS account password for authentication"),
		},
	}
}

func resourceRTXDDNSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildDDNSServerConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ddns").Msgf("Creating DDNS configuration: server_id=%d, hostname=%s", config.ID, config.Hostname)

	err := apiClient.client.ConfigureDDNS(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure DDNS: %v", err)
	}

	d.SetId(strconv.Itoa(config.ID))

	return resourceRTXDDNSRead(ctx, d, meta)
}

func resourceRTXDDNSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid DDNS server ID: %s", d.Id())
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ddns").Msgf("Reading DDNS configuration for server_id: %d", id)

	config, err := apiClient.client.GetDDNSByID(ctx, id)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Failed to read DDNS configuration: %v", err)
	}
	if config == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("server_id", config.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", config.URL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hostname", config.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", config.Username); err != nil {
		return diag.FromErr(err)
	}
	// Note: Password is not read back from router for security reasons

	return nil
}

func resourceRTXDDNSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildDDNSServerConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ddns").Msgf("Updating DDNS configuration: server_id=%d, hostname=%s", config.ID, config.Hostname)

	err := apiClient.client.UpdateDDNS(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update DDNS configuration: %v", err)
	}

	return resourceRTXDDNSRead(ctx, d, meta)
}

func resourceRTXDDNSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid DDNS server ID: %s", d.Id())
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ddns").Msgf("Deleting DDNS configuration for server_id: %d", id)

	err = apiClient.client.DeleteDDNS(ctx, id)
	if err != nil {
		return diag.Errorf("Failed to delete DDNS configuration: %v", err)
	}

	return nil
}

func resourceRTXDDNSImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid DDNS server ID: %s (must be integer 1-4)", d.Id())
	}
	if id < 1 || id > 4 {
		return nil, fmt.Errorf("invalid DDNS server ID: %d (must be 1-4)", id)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ddns").Msgf("Importing DDNS configuration for server_id: %d", id)

	config, err := apiClient.client.GetDDNSByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to import DDNS configuration: %v", err)
	}
	if config == nil {
		return nil, fmt.Errorf("DDNS configuration not found for server_id: %d", id)
	}

	d.SetId(strconv.Itoa(config.ID))
	d.Set("server_id", config.ID)
	d.Set("url", config.URL)
	d.Set("hostname", config.Hostname)
	d.Set("username", config.Username)

	return []*schema.ResourceData{d}, nil
}

func buildDDNSServerConfigFromResourceData(d *schema.ResourceData) client.DDNSServerConfig {
	return client.DDNSServerConfig{
		ID:       d.Get("server_id").(int),
		URL:      d.Get("url").(string),
		Hostname: d.Get("hostname").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}
}
