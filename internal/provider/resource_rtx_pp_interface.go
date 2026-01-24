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

func resourceRTXPPInterface() *schema.Resource {
	return &schema.Resource{
		Description: "Manages PP (Point-to-Point) interface IP configuration on RTX routers. This resource configures IP-level settings for PP interfaces used by PPPoE connections.",

		CreateContext: resourceRTXPPInterfaceCreate,
		ReadContext:   resourceRTXPPInterfaceRead,
		UpdateContext: resourceRTXPPInterfaceUpdate,
		DeleteContext: resourceRTXPPInterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXPPInterfaceImport,
		},

		Schema: map[string]*schema.Schema{
			"pp_number": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "PP interface number (1-based). This identifies the PP interface to configure.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ip_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "IP address for the PP interface. Use 'ipcp' for dynamic IP assignment from the ISP, or specify a static IP address in CIDR notation.",
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Maximum Transmission Unit size for the PP interface. Valid range: 64-1500. 0 means use default.",
				ValidateFunc: validation.IntBetween(0, 1500),
			},
			"tcp_mss": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "TCP Maximum Segment Size limit. Valid range: 1-1460. 0 means not set.",
				ValidateFunc: validation.IntBetween(0, 1460),
			},
			"nat_descriptor": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "NAT descriptor ID to bind to this PP interface. Use rtx_nat_masquerade or rtx_nat_static to define the descriptor.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"secure_filter_in": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Inbound security filter numbers. Order matters - first match wins.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},
			"secure_filter_out": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Outbound security filter numbers. Order matters - first match wins.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},
		},
	}
}

func resourceRTXPPInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_pp_interface", d.Id())
	ppNum := d.Get("pp_number").(int)
	config := buildPPIPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pp_interface").Msgf("Creating PP interface IP configuration for PP %d", ppNum)

	err := apiClient.client.ConfigurePPInterface(ctx, ppNum, config)
	if err != nil {
		return diag.Errorf("Failed to configure PP interface: %v", err)
	}

	// Use PP number as the resource ID
	d.SetId(strconv.Itoa(ppNum))

	// Read back to ensure consistency
	return resourceRTXPPInterfaceRead(ctx, d, meta)
}

func resourceRTXPPInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_pp_interface", d.Id())
	ppNum, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid PP number in resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pp_interface").Msgf("Reading PP interface IP configuration for PP %d", ppNum)

	config, err := apiClient.client.GetPPInterfaceConfig(ctx, ppNum)
	if err != nil {
		// Check if the configuration doesn't exist
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_pp_interface").Msgf("PP interface configuration for PP %d not found, removing from state", ppNum)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read PP interface configuration: %v", err)
	}

	// Update the state
	if err := d.Set("pp_number", ppNum); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip_address", config.Address); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mtu", config.MTU); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tcp_mss", config.TCPMSSLimit); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("nat_descriptor", config.NATDescriptor); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("secure_filter_in", config.SecureFilterIn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("secure_filter_out", config.SecureFilterOut); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXPPInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_pp_interface", d.Id())
	ppNum, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid PP number in resource ID: %v", err)
	}

	config := buildPPIPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pp_interface").Msgf("Updating PP interface IP configuration for PP %d", ppNum)

	err = apiClient.client.UpdatePPInterfaceConfig(ctx, ppNum, config)
	if err != nil {
		return diag.Errorf("Failed to update PP interface configuration: %v", err)
	}

	return resourceRTXPPInterfaceRead(ctx, d, meta)
}

func resourceRTXPPInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_pp_interface", d.Id())
	ppNum, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid PP number in resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pp_interface").Msgf("Resetting PP interface IP configuration for PP %d", ppNum)

	err = apiClient.client.ResetPPInterfaceConfig(ctx, ppNum)
	if err != nil {
		// Check if already reset
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to reset PP interface configuration: %v", err)
	}

	return nil
}

func resourceRTXPPInterfaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse PP number from import ID
	ppNum, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format: expected PP number (e.g., '1'), got '%s'", importID)
	}

	if ppNum < 1 {
		return nil, fmt.Errorf("invalid PP number: must be >= 1")
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pp_interface").Msgf("Importing PP interface configuration for PP %d", ppNum)

	// Verify the configuration exists and retrieve it
	config, err := apiClient.client.GetPPInterfaceConfig(ctx, ppNum)
	if err != nil {
		return nil, fmt.Errorf("failed to import PP interface configuration for PP %d: %v", ppNum, err)
	}

	// Set the resource ID
	d.SetId(strconv.Itoa(ppNum))

	// Set all attributes
	d.Set("pp_number", ppNum)
	d.Set("ip_address", config.Address)
	d.Set("mtu", config.MTU)
	d.Set("tcp_mss", config.TCPMSSLimit)
	d.Set("nat_descriptor", config.NATDescriptor)
	d.Set("secure_filter_in", config.SecureFilterIn)
	d.Set("secure_filter_out", config.SecureFilterOut)

	return []*schema.ResourceData{d}, nil
}

// buildPPIPConfigFromResourceData creates a PPIPConfig from Terraform resource data
func buildPPIPConfigFromResourceData(d *schema.ResourceData) client.PPIPConfig {
	config := client.PPIPConfig{
		Address:       GetStringValue(d, "ip_address"),
		MTU:           GetIntValue(d, "mtu"),
		TCPMSSLimit:   GetIntValue(d, "tcp_mss"),
		NATDescriptor: GetIntValue(d, "nat_descriptor"),
	}

	// Handle secure_filter_in
	if v, ok := d.GetOk("secure_filter_in"); ok {
		filtersList := v.([]interface{})
		filters := make([]int, len(filtersList))
		for i, f := range filtersList {
			filters[i] = f.(int)
		}
		config.SecureFilterIn = filters
	}

	// Handle secure_filter_out
	if v, ok := d.GetOk("secure_filter_out"); ok {
		filtersList := v.([]interface{})
		filters := make([]int, len(filtersList))
		for i, f := range filtersList {
			filters[i] = f.(int)
		}
		config.SecureFilterOut = filters
	}

	return config
}
