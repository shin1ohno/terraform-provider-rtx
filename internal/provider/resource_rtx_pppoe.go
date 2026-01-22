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

func resourceRTXPPPoE() *schema.Resource {
	return &schema.Resource{
		Description: "Manages PPPoE connection configuration on RTX routers. This resource configures PPPoE (Point-to-Point Protocol over Ethernet) connections for WAN connectivity.",

		CreateContext: resourceRTXPPPoECreate,
		ReadContext:   resourceRTXPPPoERead,
		UpdateContext: resourceRTXPPPoEUpdate,
		DeleteContext: resourceRTXPPPoEDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXPPPoEImport,
		},

		Schema: map[string]*schema.Schema{
			"pp_number": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "PP interface number (1-based). This identifies the PPPoE connection.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Connection name or description.",
			},
			"bind_interface": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Physical interface to bind for PPPoE (e.g., 'lan2').",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PPPoE authentication username.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "PPPoE authentication password. This value is sensitive and will not be displayed in logs.",
			},
			"service_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "PPPoE service name (optional). Used to specify a particular service when multiple services are available.",
			},
			"ac_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "PPPoE Access Concentrator name (optional).",
			},
			"auth_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "chap",
				Description:  "Authentication method. Valid values: 'pap', 'chap', 'mschap', 'mschap-v2'. Defaults to 'chap'.",
				ValidateFunc: validation.StringInSlice([]string{"pap", "chap", "mschap", "mschap-v2"}, false),
			},
			"always_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Keep connection always active. Defaults to true.",
			},
			"disconnect_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				Description:  "Idle disconnect timeout in seconds. 0 means no automatic disconnect.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"reconnect_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Seconds between reconnect attempts (keepalive retry interval).",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"reconnect_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Maximum reconnect attempts (0 = unlimited).",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the PP interface is enabled. Defaults to true.",
			},
		},
	}
}

func resourceRTXPPPoECreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildPPPoEConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("Creating PPPoE configuration for PP %d", config.Number)

	err := apiClient.client.CreatePPPoE(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create PPPoE configuration: %v", err)
	}

	// Use PP number as the resource ID
	d.SetId(strconv.Itoa(config.Number))

	// Read back to ensure consistency
	return resourceRTXPPPoERead(ctx, d, meta)
}

func resourceRTXPPPoERead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	ppNum, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid PP number in resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("Reading PPPoE configuration for PP %d", ppNum)

	config, err := apiClient.client.GetPPPoE(ctx, ppNum)
	if err != nil {
		// Check if the configuration doesn't exist
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("PPPoE configuration for PP %d not found, removing from state", ppNum)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read PPPoE configuration: %v", err)
	}

	// Update the state
	if err := d.Set("pp_number", config.Number); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", config.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bind_interface", config.BindInterface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", config.ServiceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ac_name", config.ACName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("always_on", config.AlwaysOn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("disconnect_timeout", config.DisconnectTimeout); err != nil {
		return diag.FromErr(err)
	}
	if config.LCPReconnect != nil {
		d.Set("reconnect_interval", config.LCPReconnect.ReconnectInterval)
		d.Set("reconnect_attempts", config.LCPReconnect.ReconnectAttempts)
	} else {
		d.Set("reconnect_interval", nil)
		d.Set("reconnect_attempts", nil)
	}
	if err := d.Set("enabled", config.Enabled); err != nil {
		return diag.FromErr(err)
	}

	// Set authentication attributes if available
	if config.Authentication != nil {
		if err := d.Set("username", config.Authentication.Username); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("auth_method", config.Authentication.Method); err != nil {
			return diag.FromErr(err)
		}
		// Note: We don't set password from read as it may be encrypted in the router config
		// The password in state should be preserved from the user's configuration
	}

	return nil
}

func resourceRTXPPPoEUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildPPPoEConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("Updating PPPoE configuration for PP %d", config.Number)

	err := apiClient.client.UpdatePPPoE(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update PPPoE configuration: %v", err)
	}

	return resourceRTXPPPoERead(ctx, d, meta)
}

func resourceRTXPPPoEDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	ppNum, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid PP number in resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("Deleting PPPoE configuration for PP %d", ppNum)

	err = apiClient.client.DeletePPPoE(ctx, ppNum)
	if err != nil {
		// Check if already deleted
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete PPPoE configuration: %v", err)
	}

	return nil
}

func resourceRTXPPPoEImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pppoe").Msgf("Importing PPPoE configuration for PP %d", ppNum)

	// Verify the configuration exists and retrieve it
	config, err := apiClient.client.GetPPPoE(ctx, ppNum)
	if err != nil {
		return nil, fmt.Errorf("failed to import PPPoE configuration for PP %d: %v", ppNum, err)
	}

	// Set the resource ID
	d.SetId(strconv.Itoa(ppNum))

	// Set all attributes
	d.Set("pp_number", config.Number)
	d.Set("name", config.Name)
	d.Set("bind_interface", config.BindInterface)
	d.Set("service_name", config.ServiceName)
	d.Set("ac_name", config.ACName)
	d.Set("always_on", config.AlwaysOn)
	d.Set("disconnect_timeout", config.DisconnectTimeout)
	if config.LCPReconnect != nil {
		d.Set("reconnect_interval", config.LCPReconnect.ReconnectInterval)
		d.Set("reconnect_attempts", config.LCPReconnect.ReconnectAttempts)
	}
	d.Set("enabled", config.Enabled)

	// Set authentication attributes
	if config.Authentication != nil {
		d.Set("username", config.Authentication.Username)
		d.Set("auth_method", config.Authentication.Method)
		// Password is encrypted in router config, so we need to handle this specially
		// The user will need to specify the actual password in their Terraform configuration
		// We set an empty string to indicate it needs to be provided
		d.Set("password", "")
	}

	return []*schema.ResourceData{d}, nil
}

// buildPPPoEConfigFromResourceData creates a PPPoEConfig from Terraform resource data
func buildPPPoEConfigFromResourceData(d *schema.ResourceData) client.PPPoEConfig {
	config := client.PPPoEConfig{
		Number:            d.Get("pp_number").(int),
		Name:              d.Get("name").(string),
		BindInterface:     d.Get("bind_interface").(string),
		ServiceName:       d.Get("service_name").(string),
		ACName:            d.Get("ac_name").(string),
		AlwaysOn:          d.Get("always_on").(bool),
		Enabled:           d.Get("enabled").(bool),
		DisconnectTimeout: d.Get("disconnect_timeout").(int),
	}

	// Set authentication
	config.Authentication = &client.PPPAuth{
		Method:   d.Get("auth_method").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}

	// Reconnect/keepalive
	if intervalRaw, ok := d.GetOk("reconnect_interval"); ok {
		interval := intervalRaw.(int)
		attempts := 0
		if attemptsRaw, ok := d.GetOk("reconnect_attempts"); ok {
			attempts = attemptsRaw.(int)
		}
		config.LCPReconnect = &client.LCPReconnectConfig{
			ReconnectInterval: interval,
			ReconnectAttempts: attempts,
		}
	}

	return config
}
