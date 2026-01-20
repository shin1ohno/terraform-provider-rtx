package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Enable or disable the L2TP service. When disabled, all L2TP/L2TPv3 tunnels will be inactive.",
			},
			"protocols": {
				Type:        schema.TypeList,
				Optional:    true,
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

	enabled := d.Get("enabled").(bool)
	protocols := expandStringList(d.Get("protocols").([]interface{}))

	log.Printf("[DEBUG] Creating L2TP service configuration: enabled=%v, protocols=%v", enabled, protocols)

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

	log.Printf("[DEBUG] Reading L2TP service configuration")

	state, err := apiClient.client.GetL2TPServiceState(ctx)
	if err != nil {
		// Check if service is not configured
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			log.Printf("[DEBUG] L2TP service not configured, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read L2TP service configuration: %v", err)
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

	enabled := d.Get("enabled").(bool)
	protocols := expandStringList(d.Get("protocols").([]interface{}))

	log.Printf("[DEBUG] Updating L2TP service configuration: enabled=%v, protocols=%v", enabled, protocols)

	err := apiClient.client.SetL2TPServiceState(ctx, enabled, protocols)
	if err != nil {
		return diag.Errorf("Failed to update L2TP service configuration: %v", err)
	}

	return resourceRTXL2TPServiceRead(ctx, d, meta)
}

func resourceRTXL2TPServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Deleting L2TP service configuration (disabling service)")

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

	log.Printf("[DEBUG] Importing L2TP service configuration")

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
