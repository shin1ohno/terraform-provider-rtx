package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXIPsecTransport() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPsec transport mode configuration on RTX routers. Used for L2TP over IPsec and other transport mode VPN configurations.",
		CreateContext: resourceRTXIPsecTransportCreate,
		ReadContext:   resourceRTXIPsecTransportRead,
		UpdateContext: resourceRTXIPsecTransportUpdate,
		DeleteContext: resourceRTXIPsecTransportDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXIPsecTransportImport,
		},

		Schema: map[string]*schema.Schema{
			"transport_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Transport ID (1-65535).",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"tunnel_id": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Associated IPsec tunnel ID (1-65535).",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Transport protocol: 'udp' or 'tcp'.",
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp"}, true),
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Port number (1-65535). Common value is 1701 for L2TP.",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
		},
	}
}

func resourceRTXIPsecTransportCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	transport := buildIPsecTransportFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipsec_transport").Msgf("Creating IPsec transport: %+v", transport)

	err := apiClient.client.CreateIPsecTransport(ctx, transport)
	if err != nil {
		return diag.Errorf("Failed to create IPsec transport: %v", err)
	}

	d.SetId(strconv.Itoa(transport.TransportID))

	return resourceRTXIPsecTransportRead(ctx, d, meta)
}

func resourceRTXIPsecTransportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	transportID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid transport ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipsec_transport").Msgf("Reading IPsec transport: %d", transportID)

	transport, err := apiClient.client.GetIPsecTransport(ctx, transportID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_ipsec_transport").Msgf("IPsec transport %d not found, removing from state", transportID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IPsec transport: %v", err)
	}

	// Update the state
	if err := d.Set("transport_id", transport.TransportID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_id", transport.TunnelID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("protocol", transport.Protocol); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("port", transport.Port); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPsecTransportUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	transport := buildIPsecTransportFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipsec_transport").Msgf("Updating IPsec transport: %+v", transport)

	err := apiClient.client.UpdateIPsecTransport(ctx, transport)
	if err != nil {
		return diag.Errorf("Failed to update IPsec transport: %v", err)
	}

	return resourceRTXIPsecTransportRead(ctx, d, meta)
}

func resourceRTXIPsecTransportDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	transportID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid transport ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipsec_transport").Msgf("Deleting IPsec transport: %d", transportID)

	err = apiClient.client.DeleteIPsecTransport(ctx, transportID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IPsec transport: %v", err)
	}

	return nil
}

func resourceRTXIPsecTransportImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	transportID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid import ID, expected transport ID as integer: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipsec_transport").Msgf("Importing IPsec transport: %d", transportID)

	transport, err := apiClient.client.GetIPsecTransport(ctx, transportID)
	if err != nil {
		return nil, fmt.Errorf("failed to import IPsec transport %d: %v", transportID, err)
	}

	d.SetId(strconv.Itoa(transport.TransportID))
	d.Set("transport_id", transport.TransportID)
	d.Set("tunnel_id", transport.TunnelID)
	d.Set("protocol", transport.Protocol)
	d.Set("port", transport.Port)

	return []*schema.ResourceData{d}, nil
}

func buildIPsecTransportFromResourceData(d *schema.ResourceData) client.IPsecTransportConfig {
	return client.IPsecTransportConfig{
		TransportID: d.Get("transport_id").(int),
		TunnelID:    d.Get("tunnel_id").(int),
		Protocol:    d.Get("protocol").(string),
		Port:        d.Get("port").(int),
	}
}
