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

func resourceRTXNetVolanteDNS() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages NetVolante DNS (Yamaha's free DDNS service) configuration on RTX routers. Use this resource to register your router's IP address with a *.netvolante.jp hostname.",
		CreateContext: resourceRTXNetVolanteDNSCreate,
		ReadContext:   resourceRTXNetVolanteDNSRead,
		UpdateContext: resourceRTXNetVolanteDNSUpdate,
		DeleteContext: resourceRTXNetVolanteDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXNetVolanteDNSImport,
		},

		Schema: map[string]*schema.Schema{
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Interface to use for DDNS updates (e.g., 'pp 1', 'lan1'). This determines which IP address is registered.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NetVolante DNS hostname (e.g., 'example.aa0.netvolante.jp'). Must end with .netvolante.jp.",
			},
			"server": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 2),
				Description:  "NetVolante DNS server number (1 or 2). Default is 1.",
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(1, 3600),
				Description:  "Update timeout in seconds (1-3600). Default is 60.",
			},
			"ipv6_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable IPv6 address registration with NetVolante DNS. Default is false.",
			},
			"auto_hostname": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable automatic hostname generation. When enabled, the router generates a unique hostname.",
			},
		},
	}
}

func resourceRTXNetVolanteDNSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildNetVolanteConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_netvolante_dns").Msgf("Creating NetVolante DNS configuration: %+v", config)

	err := apiClient.client.ConfigureNetVolanteDNS(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure NetVolante DNS: %v", err)
	}

	// Use interface as ID since NetVolante config is per-interface
	d.SetId(config.Interface)

	return resourceRTXNetVolanteDNSRead(ctx, d, meta)
}

func resourceRTXNetVolanteDNSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface := d.Id()
	logging.FromContext(ctx).Debug().Str("resource", "rtx_netvolante_dns").Msgf("Reading NetVolante DNS configuration for interface: %s", iface)

	config, err := apiClient.client.GetNetVolanteDNSByInterface(ctx, iface)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Failed to read NetVolante DNS configuration: %v", err)
	}
	if config == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("interface", config.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hostname", config.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server", config.Server); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout", config.Timeout); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ipv6_enabled", config.IPv6); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auto_hostname", config.AutoHostname); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXNetVolanteDNSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildNetVolanteConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_netvolante_dns").Msgf("Updating NetVolante DNS configuration: %+v", config)

	err := apiClient.client.UpdateNetVolanteDNS(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update NetVolante DNS configuration: %v", err)
	}

	return resourceRTXNetVolanteDNSRead(ctx, d, meta)
}

func resourceRTXNetVolanteDNSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface := d.Id()
	logging.FromContext(ctx).Debug().Str("resource", "rtx_netvolante_dns").Msgf("Deleting NetVolante DNS configuration for interface: %s", iface)

	err := apiClient.client.DeleteNetVolanteDNS(ctx, iface)
	if err != nil {
		return diag.Errorf("Failed to delete NetVolante DNS configuration: %v", err)
	}

	return nil
}

func resourceRTXNetVolanteDNSImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_netvolante_dns").Msgf("Importing NetVolante DNS configuration for interface: %s", importID)

	config, err := apiClient.client.GetNetVolanteDNSByInterface(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("failed to import NetVolante DNS configuration: %v", err)
	}
	if config == nil {
		return nil, fmt.Errorf("NetVolante DNS configuration not found for interface: %s", importID)
	}

	d.SetId(config.Interface)
	d.Set("interface", config.Interface)
	d.Set("hostname", config.Hostname)
	d.Set("server", config.Server)
	d.Set("timeout", config.Timeout)
	d.Set("ipv6_enabled", config.IPv6)
	d.Set("auto_hostname", config.AutoHostname)

	return []*schema.ResourceData{d}, nil
}

func buildNetVolanteConfigFromResourceData(d *schema.ResourceData) client.NetVolanteConfig {
	return client.NetVolanteConfig{
		Interface:    d.Get("interface").(string),
		Hostname:     d.Get("hostname").(string),
		Server:       d.Get("server").(int),
		Timeout:      d.Get("timeout").(int),
		IPv6:         d.Get("ipv6_enabled").(bool),
		AutoHostname: d.Get("auto_hostname").(bool),
		Use:          true,
	}
}
