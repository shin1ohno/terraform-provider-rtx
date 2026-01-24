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

func resourceRTXOSPF() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages OSPF (Open Shortest Path First) configuration on RTX routers. OSPF is a singleton resource - only one OSPF configuration can exist per router.",
		CreateContext: resourceRTXOSPFCreate,
		ReadContext:   resourceRTXOSPFRead,
		UpdateContext: resourceRTXOSPFUpdate,
		DeleteContext: resourceRTXOSPFDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXOSPFImport,
		},

		Schema: map[string]*schema.Schema{
			"process_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "OSPF process ID.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"router_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "OSPF router ID in IPv4 address format.",
				ValidateFunc: validateIPv4Address,
			},
			"distance": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Administrative distance for OSPF routes.",
				ValidateFunc: validation.IntBetween(1, 255),
			},
			"default_information_originate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Originate a default route into the OSPF domain.",
			},
			"network": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Networks to include in OSPF.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Network IP address or interface name.",
							ValidateFunc: validateIPv4Address,
						},
						"wildcard": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Wildcard mask (inverse mask). For example, '0.0.0.255' for a /24 network.",
						},
						"area": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "OSPF area ID in decimal (e.g., '0') or dotted decimal (e.g., '0.0.0.0') format.",
						},
					},
				},
			},
			"area": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "OSPF area configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"area_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "OSPF Area ID in decimal (e.g., '0') or dotted decimal (e.g., '0.0.0.0') format.",
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "Area type: 'normal', 'stub', or 'nssa'.",
							ValidateFunc: validation.StringInSlice([]string{"normal", "stub", "nssa"}, false),
						},
						"no_summary": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "For stub/NSSA areas, suppress summary LSAs (totally stubby/NSSA).",
						},
					},
				},
			},
			"neighbor": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "OSPF neighbors for NBMA networks.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Neighbor IP address.",
							ValidateFunc: validateIPv4Address,
						},
						"priority": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Neighbor priority (0-255).",
							ValidateFunc: validation.IntBetween(0, 255),
						},
						"cost": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Cost to reach neighbor. 0 means default cost.",
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"redistribute_static": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Redistribute static routes into OSPF.",
			},
			"redistribute_connected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Redistribute connected routes into OSPF.",
			},
		},
	}
}

func resourceRTXOSPFCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ospf", d.Id())
	config := buildOSPFConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ospf").Msgf("Creating OSPF configuration: %+v", config)

	err := apiClient.client.CreateOSPF(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to create OSPF configuration: %v", err)
	}

	// OSPF is a singleton resource
	d.SetId("ospf")

	return resourceRTXOSPFRead(ctx, d, meta)
}

func resourceRTXOSPFRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ospf", d.Id())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ospf").Msg("Reading OSPF configuration")

	var config *client.OSPFConfig

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			parsed := parsedConfig.ExtractOSPF()
			if parsed != nil {
				config = convertParsedOSPFConfig(parsed)
				logger.Debug().Str("resource", "rtx_ospf").Msg("Found OSPF config in SFTP cache")
			}
		}
		if config == nil {
			logger.Debug().Str("resource", "rtx_ospf").Msg("OSPF config not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or config not found in cache
	if config == nil {
		var err error
		config, err = apiClient.client.GetOSPF(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_ospf").Msg("OSPF configuration not found, removing from state")
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read OSPF configuration: %v", err)
		}
	}

	if !config.Enabled {
		logger.Debug().Str("resource", "rtx_ospf").Msg("OSPF is disabled, removing from state")
		d.SetId("")
		return nil
	}

	// Update the state
	if err := d.Set("process_id", config.ProcessID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("router_id", config.RouterID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("distance", config.Distance); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_information_originate", config.DefaultOriginate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redistribute_static", config.RedistributeStatic); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redistribute_connected", config.RedistributeConnected); err != nil {
		return diag.FromErr(err)
	}

	// Convert networks
	networks := make([]map[string]interface{}, len(config.Networks))
	for i, net := range config.Networks {
		networks[i] = map[string]interface{}{
			"ip":       net.IP,
			"wildcard": net.Wildcard,
			"area":     net.Area,
		}
	}
	if err := d.Set("network", networks); err != nil {
		return diag.FromErr(err)
	}

	// Convert areas
	areas := make([]map[string]interface{}, len(config.Areas))
	for i, area := range config.Areas {
		areas[i] = map[string]interface{}{
			"area_id":    area.ID,
			"type":       area.Type,
			"no_summary": area.NoSummary,
		}
	}
	if err := d.Set("area", areas); err != nil {
		return diag.FromErr(err)
	}

	// Convert neighbors
	neighbors := make([]map[string]interface{}, len(config.Neighbors))
	for i, n := range config.Neighbors {
		neighbors[i] = map[string]interface{}{
			"ip":       n.IP,
			"priority": n.Priority,
			"cost":     n.Cost,
		}
	}
	if err := d.Set("neighbor", neighbors); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXOSPFUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ospf", d.Id())
	config := buildOSPFConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ospf").Msgf("Updating OSPF configuration: %+v", config)

	err := apiClient.client.UpdateOSPF(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update OSPF configuration: %v", err)
	}

	return resourceRTXOSPFRead(ctx, d, meta)
}

func resourceRTXOSPFDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ospf", d.Id())
	logging.FromContext(ctx).Debug().Str("resource", "rtx_ospf").Msg("Disabling OSPF configuration")

	err := apiClient.client.DeleteOSPF(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to disable OSPF: %v", err)
	}

	return nil
}

func resourceRTXOSPFImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	// Import ID should be "ospf" (singleton resource)
	if d.Id() != "ospf" {
		return nil, fmt.Errorf("import ID must be 'ospf' for this singleton resource")
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ospf").Msg("Importing OSPF configuration")

	config, err := apiClient.client.GetOSPF(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import OSPF configuration: %v", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("OSPF is not configured on this router")
	}

	d.SetId("ospf")
	d.Set("process_id", config.ProcessID)
	d.Set("router_id", config.RouterID)
	d.Set("distance", config.Distance)
	d.Set("default_information_originate", config.DefaultOriginate)
	d.Set("redistribute_static", config.RedistributeStatic)
	d.Set("redistribute_connected", config.RedistributeConnected)

	networks := make([]map[string]interface{}, len(config.Networks))
	for i, net := range config.Networks {
		networks[i] = map[string]interface{}{
			"ip":       net.IP,
			"wildcard": net.Wildcard,
			"area":     net.Area,
		}
	}
	d.Set("network", networks)

	areas := make([]map[string]interface{}, len(config.Areas))
	for i, area := range config.Areas {
		areas[i] = map[string]interface{}{
			"area_id":    area.ID,
			"type":       area.Type,
			"no_summary": area.NoSummary,
		}
	}
	d.Set("area", areas)

	neighbors := make([]map[string]interface{}, len(config.Neighbors))
	for i, n := range config.Neighbors {
		neighbors[i] = map[string]interface{}{
			"ip":       n.IP,
			"priority": n.Priority,
			"cost":     n.Cost,
		}
	}
	d.Set("neighbor", neighbors)

	return []*schema.ResourceData{d}, nil
}

func buildOSPFConfigFromResourceData(d *schema.ResourceData) client.OSPFConfig {
	config := client.OSPFConfig{
		Enabled:               true,
		ProcessID:             d.Get("process_id").(int),
		RouterID:              d.Get("router_id").(string),
		Distance:              d.Get("distance").(int),
		DefaultOriginate:      d.Get("default_information_originate").(bool),
		RedistributeStatic:    d.Get("redistribute_static").(bool),
		RedistributeConnected: d.Get("redistribute_connected").(bool),
	}

	// Handle networks
	if v, ok := d.GetOk("network"); ok {
		networkList := v.([]interface{})
		config.Networks = make([]client.OSPFNetwork, len(networkList))
		for i, n := range networkList {
			nMap := n.(map[string]interface{})
			config.Networks[i] = client.OSPFNetwork{
				IP:       nMap["ip"].(string),
				Wildcard: nMap["wildcard"].(string),
				Area:     nMap["area"].(string),
			}
		}
	}

	// Handle areas
	if v, ok := d.GetOk("area"); ok {
		areaList := v.([]interface{})
		config.Areas = make([]client.OSPFArea, len(areaList))
		for i, a := range areaList {
			aMap := a.(map[string]interface{})
			config.Areas[i] = client.OSPFArea{
				ID:        aMap["area_id"].(string),
				Type:      aMap["type"].(string),
				NoSummary: aMap["no_summary"].(bool),
			}
		}
	}

	// Handle neighbors
	if v, ok := d.GetOk("neighbor"); ok {
		neighborList := v.([]interface{})
		config.Neighbors = make([]client.OSPFNeighbor, len(neighborList))
		for i, n := range neighborList {
			nMap := n.(map[string]interface{})
			config.Neighbors[i] = client.OSPFNeighbor{
				IP:       nMap["ip"].(string),
				Priority: nMap["priority"].(int),
				Cost:     nMap["cost"].(int),
			}
		}
	}

	return config
}

// convertParsedOSPFConfig converts a parser OSPFConfig to a client OSPFConfig
func convertParsedOSPFConfig(parsed *parsers.OSPFConfig) *client.OSPFConfig {
	config := &client.OSPFConfig{
		Enabled:               parsed.Enabled,
		ProcessID:             parsed.ProcessID,
		RouterID:              parsed.RouterID,
		Distance:              parsed.Distance,
		DefaultOriginate:      parsed.DefaultOriginate,
		RedistributeStatic:    parsed.RedistributeStatic,
		RedistributeConnected: parsed.RedistributeConnected,
		Networks:              make([]client.OSPFNetwork, len(parsed.Networks)),
		Areas:                 make([]client.OSPFArea, len(parsed.Areas)),
		Neighbors:             make([]client.OSPFNeighbor, len(parsed.Neighbors)),
	}

	// Convert networks
	for i, n := range parsed.Networks {
		config.Networks[i] = client.OSPFNetwork{
			IP:       n.IP,
			Wildcard: n.Wildcard,
			Area:     n.Area,
		}
	}

	// Convert areas
	for i, a := range parsed.Areas {
		config.Areas[i] = client.OSPFArea{
			ID:        a.ID,
			Type:      a.Type,
			NoSummary: a.NoSummary,
		}
	}

	// Convert neighbors
	for i, n := range parsed.Neighbors {
		config.Neighbors[i] = client.OSPFNeighbor{
			IP:       n.IP,
			Priority: n.Priority,
			Cost:     n.Cost,
		}
	}

	return config
}
