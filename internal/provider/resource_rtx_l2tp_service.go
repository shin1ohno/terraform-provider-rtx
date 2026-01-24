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
)

func resourceRTXL2TPService() *schema.Resource {
	return &schema.Resource{
		Description: "Manages L2TP service configuration on RTX routers. " +
			"This is a singleton resource - only one instance can exist per router. " +
			"The L2TP service must be enabled for L2TP/L2TPv3 tunnels to function.",
		CreateContext: resourceRTXL2TPServiceCreate,
		ReadContext:   resourceRTXL2TPServiceRead,
		UpdateContext: resourceRTXL2TPServiceUpdate,
		DeleteContext: resourceRTXL2TPServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXL2TPServiceImport,
		},
		CustomizeDiff: customizeL2TPServiceDiff,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Enable or disable the L2TP service. When disabled, all L2TP/L2TPv3 tunnels will be inactive.",
			},
			"protocols": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of L2TP protocols to enable. Valid values are 'l2tp' (L2TPv2) and 'l2tpv3'. If not specified, defaults to all protocols when enabled.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"l2tp",
						"l2tpv3",
					}, false),
				},
			},
		},
	}
}

func resourceRTXL2TPServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp_service", d.Id())
	enabled := d.Get("enabled").(bool)
	protocols := expandStringList(d.Get("protocols").([]interface{}))

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp_service").Msgf("Creating L2TP service configuration: enabled=%v, protocols=%v", enabled, protocols)

	err := apiClient.client.SetL2TPServiceState(ctx, enabled, protocols)
	if err != nil {
		return diag.Errorf("Failed to configure L2TP service: %v", err)
	}

	// Use fixed ID for singleton resource
	d.SetId("default")

	// Read back to ensure consistency
	return resourceRTXL2TPServiceRead(ctx, d, meta)
}

func resourceRTXL2TPServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp_service", d.Id())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_l2tp_service").Msg("Reading L2TP service configuration")

	var state *client.L2TPServiceState

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract L2TP service from parsed config
			service := parsedConfig.ExtractL2TPService()
			if service != nil {
				state = &client.L2TPServiceState{
					Enabled:   service.Enabled,
					Protocols: service.Protocols,
				}
				logger.Debug().Str("resource", "rtx_l2tp_service").Msg("Found L2TP service in SFTP cache")
			}
		}
		if state == nil {
			// Service not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_l2tp_service").Msg("L2TP service not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or service not found in cache
	if state == nil {
		var err error
		state, err = apiClient.client.GetL2TPServiceState(ctx)
		if err != nil {
			// Check if service is not configured
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
				logger.Debug().Str("resource", "rtx_l2tp_service").Msg("L2TP service not configured, removing from state")
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read L2TP service configuration: %v", err)
		}
	}

	// Update the state
	if err := d.Set("enabled", state.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("protocols", state.Protocols); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXL2TPServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp_service", d.Id())
	enabled := d.Get("enabled").(bool)
	protocols := expandStringList(d.Get("protocols").([]interface{}))

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp_service").Msgf("Updating L2TP service configuration: enabled=%v, protocols=%v", enabled, protocols)

	err := apiClient.client.SetL2TPServiceState(ctx, enabled, protocols)
	if err != nil {
		return diag.Errorf("Failed to update L2TP service configuration: %v", err)
	}

	return resourceRTXL2TPServiceRead(ctx, d, meta)
}

func resourceRTXL2TPServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_l2tp_service", d.Id())
	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp_service").Msg("Deleting L2TP service configuration (disabling service)")

	// Disable L2TP service on delete
	err := apiClient.client.SetL2TPServiceState(ctx, false, nil)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to disable L2TP service: %v", err)
	}

	return nil
}

func resourceRTXL2TPServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Accept "default" as the import ID (singleton resource)
	if importID != "default" {
		return nil, fmt.Errorf("invalid import ID format, expected 'default', got %q", importID)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_l2tp_service").Msg("Importing L2TP service configuration")

	// Verify L2TP service state
	state, err := apiClient.client.GetL2TPServiceState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import L2TP service configuration: %v", err)
	}

	// Set the ID and attributes
	d.SetId("default")
	d.Set("enabled", state.Enabled)
	d.Set("protocols", state.Protocols)

	return []*schema.ResourceData{d}, nil
}

// expandStringList converts a []interface{} to []string
func expandStringList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

// customizeL2TPServiceDiff suppresses diff when protocols are functionally equivalent.
// RTX routers treat ["l2tpv3", "l2tp"] (or ["l2tp", "l2tpv3"]) as equivalent to []
// because both versions enabled is the default when no protocols are specified.
func customizeL2TPServiceDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	oldVal, newVal := d.GetChange("protocols")
	oldList := expandInterfaceList(oldVal.([]interface{}))
	newList := expandInterfaceList(newVal.([]interface{}))

	// Check if both are "all protocols" (either empty or both protocols specified)
	oldIsAllProtocols := isAllProtocols(oldList)
	newIsAllProtocols := isAllProtocols(newList)

	if oldIsAllProtocols && newIsAllProtocols {
		// Both represent "all protocols enabled", suppress the diff
		if err := d.Clear("protocols"); err != nil {
			return err
		}
	}

	return nil
}

// isAllProtocols returns true if the protocol list means "all protocols enabled"
// This is either an empty list or a list containing both "l2tp" and "l2tpv3"
func isAllProtocols(protocols []string) bool {
	if len(protocols) == 0 {
		return true
	}
	if len(protocols) == 2 {
		hasL2TP := false
		hasL2TPv3 := false
		for _, p := range protocols {
			if p == "l2tp" {
				hasL2TP = true
			}
			if p == "l2tpv3" {
				hasL2TPv3 = true
			}
		}
		return hasL2TP && hasL2TPv3
	}
	return false
}

// expandInterfaceList converts []interface{} to []string
func expandInterfaceList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		if s, ok := v.(string); ok {
			result[i] = s
		}
	}
	return result
}
