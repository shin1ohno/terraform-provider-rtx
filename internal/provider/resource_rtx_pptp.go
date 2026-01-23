package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXPPTP() *schema.Resource {
	return &schema.Resource{
		Description: "Manages PPTP VPN server configuration on RTX routers. PPTP is a singleton resource.\n\n" +
			"**Security Warning:** PPTP is considered insecure due to known vulnerabilities in its authentication and encryption protocols. " +
			"Consider using L2TP/IPsec or IKEv2 instead for better security.",
		CreateContext: resourceRTXPPTPCreate,
		ReadContext:   resourceRTXPPTPRead,
		UpdateContext: resourceRTXPPTPUpdate,
		DeleteContext: resourceRTXPPTPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXPPTPImport,
		},

		Schema: map[string]*schema.Schema{
			"shutdown": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Administratively shut down PPTP service.",
			},
			"listen_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "IP address to listen on.",
				ValidateFunc: validateIPv4Address,
			},
			"max_connections": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Maximum concurrent connections. 0 means no limit.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"authentication": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "PPTP authentication settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Authentication method: 'pap', 'chap', 'mschap', or 'mschap-v2'. Note: mschap-v2 is required for MPPE encryption.",
							ValidateFunc: validation.StringInSlice([]string{"pap", "chap", "mschap", "mschap-v2"}, false),
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Username for authentication.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Password for authentication.",
						},
					},
				},
			},
			"encryption": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "MPPE encryption settings. Requires mschap or mschap-v2 authentication.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mppe_bits": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "MPPE encryption strength: 40, 56, or 128 bits.",
							ValidateFunc: validation.IntInSlice([]int{40, 56, 128}),
						},
						"required": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Require encryption for all connections.",
						},
					},
				},
			},
			"ip_pool": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "IP pool for PPTP clients.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Start IP address of the pool.",
							ValidateFunc: validateIPv4Address,
						},
						"end": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "End IP address of the pool.",
							ValidateFunc: validateIPv4Address,
						},
					},
				},
			},
			"disconnect_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Idle disconnect time in seconds. 0 means no timeout.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"keepalive_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable keepalive for PPTP connections.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable PPTP service.",
			},
		},
	}
}

func resourceRTXPPTPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildPPTPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pptp").Msgf("Creating PPTP configuration: %+v", config)

	err := apiClient.client.CreatePPTP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create PPTP configuration: %v", err)
	}

	// PPTP is a singleton resource
	d.SetId("pptp")

	return resourceRTXPPTPRead(ctx, d, meta)
}

func resourceRTXPPTPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_pptp").Msg("Reading PPTP configuration")

	var config *client.PPTPConfig

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractPPTP()
			if parsed != nil {
				config = convertParsedPPTPConfig(parsed)
				logger.Debug().Str("resource", "rtx_pptp").Msg("Found PPTP config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_pptp").Msg("PPTP config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = apiClient.client.GetPPTP(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_pptp").Msg("PPTP configuration not found, removing from state")
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read PPTP configuration: %v", err)
		}
	}

	if !config.Enabled {
		logger.Debug().Str("resource", "rtx_pptp").Msg("PPTP is disabled, removing from state")
		d.SetId("")
		return nil
	}

	// Update the state
	if err := d.Set("shutdown", config.Shutdown); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("listen_address", config.ListenAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("max_connections", config.MaxConnections); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("disconnect_time", config.DisconnectTime); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("keepalive_enabled", config.KeepaliveEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", config.Enabled); err != nil {
		return diag.FromErr(err)
	}

	// Set authentication
	if config.Authentication != nil {
		auth := []map[string]interface{}{
			{
				"method":   config.Authentication.Method,
				"username": config.Authentication.Username,
				"password": config.Authentication.Password,
			},
		}
		if err := d.Set("authentication", auth); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set encryption
	if config.Encryption != nil {
		encryption := []map[string]interface{}{
			{
				"mppe_bits": config.Encryption.MPPEBits,
				"required":  config.Encryption.Required,
			},
		}
		if err := d.Set("encryption", encryption); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set IP pool
	if config.IPPool != nil {
		ipPool := []map[string]interface{}{
			{
				"start": config.IPPool.Start,
				"end":   config.IPPool.End,
			},
		}
		if err := d.Set("ip_pool", ipPool); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRTXPPTPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildPPTPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pptp").Msgf("Updating PPTP configuration: %+v", config)

	err := apiClient.client.UpdatePPTP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update PPTP configuration: %v", err)
	}

	return resourceRTXPPTPRead(ctx, d, meta)
}

func resourceRTXPPTPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pptp").Msg("Disabling PPTP configuration")

	err := apiClient.client.DeletePPTP(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to disable PPTP: %v", err)
	}

	return nil
}

func resourceRTXPPTPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	// Import ID should be "pptp" (singleton resource)
	if d.Id() != "pptp" {
		return nil, fmt.Errorf("import ID must be 'pptp' for this singleton resource")
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_pptp").Msg("Importing PPTP configuration")

	config, err := apiClient.client.GetPPTP(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import PPTP configuration: %v", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("PPTP is not configured on this router")
	}

	d.SetId("pptp")
	d.Set("shutdown", config.Shutdown)
	d.Set("listen_address", config.ListenAddress)
	d.Set("max_connections", config.MaxConnections)
	d.Set("disconnect_time", config.DisconnectTime)
	d.Set("keepalive_enabled", config.KeepaliveEnabled)
	d.Set("enabled", config.Enabled)

	if config.Authentication != nil {
		auth := []map[string]interface{}{
			{
				"method":   config.Authentication.Method,
				"username": config.Authentication.Username,
				"password": config.Authentication.Password,
			},
		}
		d.Set("authentication", auth)
	}

	if config.Encryption != nil {
		encryption := []map[string]interface{}{
			{
				"mppe_bits": config.Encryption.MPPEBits,
				"required":  config.Encryption.Required,
			},
		}
		d.Set("encryption", encryption)
	}

	if config.IPPool != nil {
		ipPool := []map[string]interface{}{
			{
				"start": config.IPPool.Start,
				"end":   config.IPPool.End,
			},
		}
		d.Set("ip_pool", ipPool)
	}

	return []*schema.ResourceData{d}, nil
}

func buildPPTPConfigFromResourceData(d *schema.ResourceData) client.PPTPConfig {
	config := client.PPTPConfig{
		Shutdown:         d.Get("shutdown").(bool),
		ListenAddress:    d.Get("listen_address").(string),
		MaxConnections:   d.Get("max_connections").(int),
		DisconnectTime:   d.Get("disconnect_time").(int),
		KeepaliveEnabled: d.Get("keepalive_enabled").(bool),
		Enabled:          d.Get("enabled").(bool),
	}

	// Handle authentication
	if v, ok := d.GetOk("authentication"); ok {
		authList := v.([]interface{})
		if len(authList) > 0 {
			aMap := authList[0].(map[string]interface{})
			config.Authentication = &client.PPTPAuth{
				Method:   aMap["method"].(string),
				Username: aMap["username"].(string),
				Password: aMap["password"].(string),
			}
		}
	}

	// Handle encryption
	if v, ok := d.GetOk("encryption"); ok {
		encList := v.([]interface{})
		if len(encList) > 0 {
			eMap := encList[0].(map[string]interface{})
			config.Encryption = &client.PPTPEncryption{
				MPPEBits: eMap["mppe_bits"].(int),
				Required: eMap["required"].(bool),
			}
		}
	}

	// Handle IP pool
	if v, ok := d.GetOk("ip_pool"); ok {
		poolList := v.([]interface{})
		if len(poolList) > 0 {
			pMap := poolList[0].(map[string]interface{})
			config.IPPool = &client.PPTPIPPool{
				Start: pMap["start"].(string),
				End:   pMap["end"].(string),
			}
		}
	}

	return config
}

// convertParsedPPTPConfig converts a parser PPTPConfig to a client PPTPConfig
func convertParsedPPTPConfig(parsed *parsers.PPTPConfig) *client.PPTPConfig {
	config := &client.PPTPConfig{
		Shutdown:         parsed.Shutdown,
		ListenAddress:    parsed.ListenAddress,
		MaxConnections:   parsed.MaxConnections,
		DisconnectTime:   parsed.DisconnectTime,
		KeepaliveEnabled: parsed.KeepaliveEnabled,
		Enabled:          parsed.Enabled,
	}

	// Convert authentication
	if parsed.Authentication != nil {
		config.Authentication = &client.PPTPAuth{
			Method:   parsed.Authentication.Method,
			Username: parsed.Authentication.Username,
			Password: parsed.Authentication.Password,
		}
	}

	// Convert encryption
	if parsed.Encryption != nil {
		config.Encryption = &client.PPTPEncryption{
			MPPEBits: parsed.Encryption.MPPEBits,
			Required: parsed.Encryption.Required,
		}
	}

	// Convert IP pool
	if parsed.IPPool != nil {
		config.IPPool = &client.PPTPIPPool{
			Start: parsed.IPPool.Start,
			End:   parsed.IPPool.End,
		}
	}

	return config
}
