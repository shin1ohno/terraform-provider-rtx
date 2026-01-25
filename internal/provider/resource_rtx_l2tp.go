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
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXL2TP() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages L2TP/L2TPv3 tunnel configuration on RTX routers. Supports both L2TPv2 (LNS for remote access VPN) and L2TPv3 (L2VPN for site-to-site).",
		CreateContext: resourceRTXL2TPCreate,
		ReadContext:   resourceRTXL2TPRead,
		UpdateContext: resourceRTXL2TPUpdate,
		DeleteContext: resourceRTXL2TPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXL2TPImport,
		},

		Schema: map[string]*schema.Schema{
			"tunnel_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Tunnel ID (1-65535).",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"tunnel_interface": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tunnel interface name (e.g., 'tunnel1'). Computed from tunnel_id.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Tunnel description.",
			},
			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "L2TP version: 'l2tp' (v2) or 'l2tpv3' (v3).",
				ValidateFunc: validation.StringInSlice([]string{"l2tp", "l2tpv3"}, false),
			},
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Operating mode: 'lns' (L2TPv2 server) or 'l2vpn' (L2TPv3 site-to-site).",
				ValidateFunc: validation.StringInSlice([]string{"lns", "l2vpn"}, false),
			},
			"shutdown": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Administratively shut down the tunnel.",
			},
			"tunnel_source": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Source IP address or interface.",
			},
			"tunnel_destination": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Destination IP address or FQDN.",
			},
			"tunnel_dest_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Destination type: 'ip' or 'fqdn'.",
				ValidateFunc: validation.StringInSlice([]string{"ip", "fqdn"}, false),
			},
			"authentication": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "L2TPv2 authentication settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Authentication method: 'pap', 'chap', 'mschap', or 'mschap-v2'.",
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
			"ip_pool": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "IP pool for L2TPv2 LNS clients.",
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
			"ipsec_profile": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "IPsec encryption settings for L2TP.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable IPsec encryption.",
						},
						"pre_shared_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "IPsec pre-shared key.",
						},
						"tunnel_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Associated IPsec tunnel ID.",
						},
					},
				},
			},
			"l2tpv3_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "L2TPv3-specific configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"local_router_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "Local router ID. Required for L2TPv3, but optional for import compatibility.",
							ValidateFunc: validateIPv4Address,
						},
						"remote_router_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "Remote router ID. Required for L2TPv3, but optional for import compatibility.",
							ValidateFunc: validateIPv4Address,
						},
						"remote_end_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Remote end ID (hostname).",
						},
						"session_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Session ID.",
						},
						"cookie_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Cookie size: 0, 4, or 8 bytes.",
							ValidateFunc: validation.IntInSlice([]int{0, 4, 8}),
						},
						"bridge_interface": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Bridge interface for L2VPN.",
						},
						"tunnel_auth_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable tunnel authentication.",
						},
						"tunnel_auth_password": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Preserve imported value when not specified in config
							Sensitive:   true,
							Description: "Tunnel authentication password.",
						},
					},
				},
			},
			"keepalive_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable keepalive.",
			},
			"keepalive_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Keepalive interval in seconds.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"keepalive_retry": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Keepalive retry count.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"disconnect_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Idle disconnect time in seconds. 0 means no timeout.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"always_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable always-on mode.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable the L2TP tunnel.",
			},
		},
	}
}

func resourceRTXL2TPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp", d.Id())
	config := buildL2TPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp").Msgf("Creating L2TP tunnel: %+v", config)

	err := apiClient.client.CreateL2TP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create L2TP tunnel: %v", err)
	}

	d.SetId(strconv.Itoa(config.ID))

	return resourceRTXL2TPRead(ctx, d, meta)
}

func resourceRTXL2TPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp", d.Id())
	logger := logging.FromContext(ctx)

	tunnelID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid tunnel ID: %v", err)
	}

	logger.Debug().Str("resource", "rtx_l2tp").Msgf("Reading L2TP tunnel: %d", tunnelID)

	var config *client.L2TPConfig

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract L2TP tunnels from parsed config
			tunnels := parsedConfig.ExtractL2TPTunnels()
			for i := range tunnels {
				if tunnels[i].ID == tunnelID {
					config = convertParsedL2TPConfig(&tunnels[i])
					logger.Debug().Str("resource", "rtx_l2tp").Msg("Found L2TP tunnel in SFTP cache")
					break
				}
			}
		}
		if config == nil {
			// Tunnel not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_l2tp").Msg("L2TP tunnel not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or tunnel not found in cache
	if config == nil {
		config, err = apiClient.client.GetL2TP(ctx, tunnelID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_l2tp").Msgf("L2TP tunnel %d not found, removing from state", tunnelID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read L2TP tunnel: %v", err)
		}
	}

	// Update the state
	if err := d.Set("tunnel_id", config.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_interface", fmt.Sprintf("tunnel%d", config.ID)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", config.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", config.Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mode", config.Mode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("shutdown", config.Shutdown); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_source", config.TunnelSource); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_destination", config.TunnelDest); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_dest_type", config.TunnelDestType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("keepalive_enabled", config.KeepaliveEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("disconnect_time", config.DisconnectTime); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("always_on", config.AlwaysOn); err != nil {
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

	// Set IPsec profile
	if config.IPsecProfile != nil {
		ipsec := []map[string]interface{}{
			{
				"enabled":        config.IPsecProfile.Enabled,
				"pre_shared_key": config.IPsecProfile.PreSharedKey,
				"tunnel_id":      config.IPsecProfile.TunnelID,
			},
		}
		if err := d.Set("ipsec_profile", ipsec); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set L2TPv3 config
	if config.L2TPv3Config != nil {
		l2tpv3 := []map[string]interface{}{
			{
				"local_router_id":      config.L2TPv3Config.LocalRouterID,
				"remote_router_id":     config.L2TPv3Config.RemoteRouterID,
				"remote_end_id":        config.L2TPv3Config.RemoteEndID,
				"session_id":           config.L2TPv3Config.SessionID,
				"cookie_size":          config.L2TPv3Config.CookieSize,
				"bridge_interface":     config.L2TPv3Config.BridgeInterface,
				"tunnel_auth_enabled":  config.L2TPv3Config.TunnelAuth != nil && config.L2TPv3Config.TunnelAuth.Enabled,
				"tunnel_auth_password": "",
			},
		}
		if config.L2TPv3Config.TunnelAuth != nil {
			l2tpv3[0]["tunnel_auth_password"] = config.L2TPv3Config.TunnelAuth.Password
		}
		if err := d.Set("l2tpv3_config", l2tpv3); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set keepalive config
	if config.KeepaliveConfig != nil {
		if err := d.Set("keepalive_interval", config.KeepaliveConfig.Interval); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("keepalive_retry", config.KeepaliveConfig.Retry); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRTXL2TPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp", d.Id())
	config := buildL2TPConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp").Msgf("Updating L2TP tunnel: %+v", config)

	err := apiClient.client.UpdateL2TP(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update L2TP tunnel: %v", err)
	}

	return resourceRTXL2TPRead(ctx, d, meta)
}

func resourceRTXL2TPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp", d.Id())
	tunnelID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid tunnel ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp").Msgf("Deleting L2TP tunnel: %d", tunnelID)

	err = apiClient.client.DeleteL2TP(ctx, tunnelID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete L2TP tunnel: %v", err)
	}

	return nil
}

func resourceRTXL2TPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	tunnelID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid import ID, expected tunnel ID as integer: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp").Msgf("Importing L2TP tunnel: %d", tunnelID)

	config, err := apiClient.client.GetL2TP(ctx, tunnelID)
	if err != nil {
		return nil, fmt.Errorf("failed to import L2TP tunnel %d: %v", tunnelID, err)
	}

	d.SetId(strconv.Itoa(config.ID))
	d.Set("tunnel_id", config.ID)
	d.Set("tunnel_interface", fmt.Sprintf("tunnel%d", config.ID))
	d.Set("name", config.Name)
	d.Set("version", config.Version)
	d.Set("mode", config.Mode)
	d.Set("shutdown", config.Shutdown)
	d.Set("tunnel_source", config.TunnelSource)
	d.Set("tunnel_destination", config.TunnelDest)
	d.Set("tunnel_dest_type", config.TunnelDestType)
	d.Set("keepalive_enabled", config.KeepaliveEnabled)
	d.Set("disconnect_time", config.DisconnectTime)
	d.Set("always_on", config.AlwaysOn)
	d.Set("enabled", config.Enabled)

	// Set authentication
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

	// Set IP pool
	if config.IPPool != nil {
		ipPool := []map[string]interface{}{
			{
				"start": config.IPPool.Start,
				"end":   config.IPPool.End,
			},
		}
		d.Set("ip_pool", ipPool)
	}

	// Set IPsec profile
	if config.IPsecProfile != nil {
		ipsec := []map[string]interface{}{
			{
				"enabled":        config.IPsecProfile.Enabled,
				"pre_shared_key": config.IPsecProfile.PreSharedKey,
				"tunnel_id":      config.IPsecProfile.TunnelID,
			},
		}
		d.Set("ipsec_profile", ipsec)
	}

	// Set L2TPv3 config (including tunnel_auth)
	if config.L2TPv3Config != nil {
		l2tpv3 := []map[string]interface{}{
			{
				"local_router_id":      config.L2TPv3Config.LocalRouterID,
				"remote_router_id":     config.L2TPv3Config.RemoteRouterID,
				"remote_end_id":        config.L2TPv3Config.RemoteEndID,
				"session_id":           config.L2TPv3Config.SessionID,
				"cookie_size":          config.L2TPv3Config.CookieSize,
				"bridge_interface":     config.L2TPv3Config.BridgeInterface,
				"tunnel_auth_enabled":  config.L2TPv3Config.TunnelAuth != nil && config.L2TPv3Config.TunnelAuth.Enabled,
				"tunnel_auth_password": "",
			},
		}
		if config.L2TPv3Config.TunnelAuth != nil {
			l2tpv3[0]["tunnel_auth_password"] = config.L2TPv3Config.TunnelAuth.Password
		}
		d.Set("l2tpv3_config", l2tpv3)
	}

	// Set keepalive config
	if config.KeepaliveConfig != nil {
		d.Set("keepalive_interval", config.KeepaliveConfig.Interval)
		d.Set("keepalive_retry", config.KeepaliveConfig.Retry)
	}

	return []*schema.ResourceData{d}, nil
}

func buildL2TPConfigFromResourceData(d *schema.ResourceData) client.L2TPConfig {
	config := client.L2TPConfig{
		ID:               d.Get("tunnel_id").(int),
		Name:             d.Get("name").(string),
		Version:          d.Get("version").(string),
		Mode:             d.Get("mode").(string),
		Shutdown:         d.Get("shutdown").(bool),
		TunnelSource:     d.Get("tunnel_source").(string),
		TunnelDest:       d.Get("tunnel_destination").(string),
		TunnelDestType:   d.Get("tunnel_dest_type").(string),
		KeepaliveEnabled: d.Get("keepalive_enabled").(bool),
		DisconnectTime:   d.Get("disconnect_time").(int),
		AlwaysOn:         d.Get("always_on").(bool),
		Enabled:          d.Get("enabled").(bool),
	}

	// Handle authentication
	if v, ok := d.GetOk("authentication"); ok {
		authList := v.([]interface{})
		if len(authList) > 0 {
			aMap := authList[0].(map[string]interface{})
			config.Authentication = &client.L2TPAuth{
				Method:   aMap["method"].(string),
				Username: aMap["username"].(string),
				Password: aMap["password"].(string),
			}
		}
	}

	// Handle IP pool
	if v, ok := d.GetOk("ip_pool"); ok {
		poolList := v.([]interface{})
		if len(poolList) > 0 {
			pMap := poolList[0].(map[string]interface{})
			config.IPPool = &client.L2TPIPPool{
				Start: pMap["start"].(string),
				End:   pMap["end"].(string),
			}
		}
	}

	// Handle IPsec profile
	if v, ok := d.GetOk("ipsec_profile"); ok {
		ipsecList := v.([]interface{})
		if len(ipsecList) > 0 {
			iMap := ipsecList[0].(map[string]interface{})
			config.IPsecProfile = &client.L2TPIPsec{
				Enabled:      iMap["enabled"].(bool),
				PreSharedKey: iMap["pre_shared_key"].(string),
				TunnelID:     iMap["tunnel_id"].(int),
			}
		}
	}

	// Handle L2TPv3 config
	if v, ok := d.GetOk("l2tpv3_config"); ok {
		l2tpv3List := v.([]interface{})
		if len(l2tpv3List) > 0 {
			lMap := l2tpv3List[0].(map[string]interface{})
			config.L2TPv3Config = &client.L2TPv3Config{
				LocalRouterID:   lMap["local_router_id"].(string),
				RemoteRouterID:  lMap["remote_router_id"].(string),
				RemoteEndID:     lMap["remote_end_id"].(string),
				SessionID:       lMap["session_id"].(int),
				CookieSize:      lMap["cookie_size"].(int),
				BridgeInterface: lMap["bridge_interface"].(string),
			}
			if lMap["tunnel_auth_enabled"].(bool) {
				config.L2TPv3Config.TunnelAuth = &client.L2TPTunnelAuth{
					Enabled:  true,
					Password: lMap["tunnel_auth_password"].(string),
				}
			}
		}
	}

	// Handle keepalive config
	if d.Get("keepalive_enabled").(bool) {
		config.KeepaliveConfig = &client.L2TPKeepalive{
			Interval: d.Get("keepalive_interval").(int),
			Retry:    d.Get("keepalive_retry").(int),
		}
	}

	return config
}

// convertParsedL2TPConfig converts a parser L2TPConfig to a client L2TPConfig
func convertParsedL2TPConfig(parsed *parsers.L2TPConfig) *client.L2TPConfig {
	config := &client.L2TPConfig{
		ID:               parsed.ID,
		Name:             parsed.Name,
		Version:          parsed.Version,
		Mode:             parsed.Mode,
		Shutdown:         parsed.Shutdown,
		TunnelSource:     parsed.TunnelSource,
		TunnelDest:       parsed.TunnelDest,
		TunnelDestType:   parsed.TunnelDestType,
		KeepaliveEnabled: parsed.KeepaliveEnabled,
		DisconnectTime:   parsed.DisconnectTime,
		AlwaysOn:         parsed.AlwaysOn,
		Enabled:          parsed.Enabled,
	}

	if parsed.Authentication != nil {
		config.Authentication = &client.L2TPAuth{
			Method:   parsed.Authentication.Method,
			Username: parsed.Authentication.Username,
			Password: parsed.Authentication.Password,
		}
	}

	if parsed.IPPool != nil {
		config.IPPool = &client.L2TPIPPool{
			Start: parsed.IPPool.Start,
			End:   parsed.IPPool.End,
		}
	}

	if parsed.IPsecProfile != nil {
		config.IPsecProfile = &client.L2TPIPsec{
			Enabled:      parsed.IPsecProfile.Enabled,
			PreSharedKey: parsed.IPsecProfile.PreSharedKey,
			TunnelID:     parsed.IPsecProfile.TunnelID,
		}
	}

	if parsed.L2TPv3Config != nil {
		config.L2TPv3Config = &client.L2TPv3Config{
			LocalRouterID:   parsed.L2TPv3Config.LocalRouterID,
			RemoteRouterID:  parsed.L2TPv3Config.RemoteRouterID,
			RemoteEndID:     parsed.L2TPv3Config.RemoteEndID,
			SessionID:       parsed.L2TPv3Config.SessionID,
			CookieSize:      parsed.L2TPv3Config.CookieSize,
			BridgeInterface: parsed.L2TPv3Config.BridgeInterface,
		}
		if parsed.L2TPv3Config.TunnelAuth != nil {
			config.L2TPv3Config.TunnelAuth = &client.L2TPTunnelAuth{
				Enabled:  parsed.L2TPv3Config.TunnelAuth.Enabled,
				Password: parsed.L2TPv3Config.TunnelAuth.Password,
			}
		}
	}

	if parsed.KeepaliveConfig != nil {
		config.KeepaliveConfig = &client.L2TPKeepalive{
			Interval: parsed.KeepaliveConfig.Interval,
			Retry:    parsed.KeepaliveConfig.Retry,
		}
	}

	return config
}
