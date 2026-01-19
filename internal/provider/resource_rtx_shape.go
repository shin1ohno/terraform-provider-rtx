package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXShape() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages QoS traffic shaping configurations on RTX routers. Shaping limits the rate of outgoing traffic on an interface.",
		CreateContext: resourceRTXShapeCreate,
		ReadContext:   resourceRTXShapeRead,
		UpdateContext: resourceRTXShapeUpdate,
		DeleteContext: resourceRTXShapeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXShapeImport,
		},

		Schema: map[string]*schema.Schema{
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The interface to apply traffic shaping to (e.g., 'lan1', 'wan1')",
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Direction for shaping: 'input' or 'output'",
				ValidateFunc: validation.StringInSlice([]string{"input", "output"}, false),
			},
			"shape_average": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Average rate limit in bits per second (bps)",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"shape_burst": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Burst size in bytes (optional)",
			},
		},
	}
}

func resourceRTXShapeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	sc := buildShapeConfigFromResourceData(d)

	log.Printf("[DEBUG] Creating shape: %+v", sc)

	err := apiClient.client.CreateShape(ctx, sc)
	if err != nil {
		return diag.Errorf("Failed to create shape: %v", err)
	}

	// Resource ID format: interface:direction
	d.SetId(fmt.Sprintf("%s:%s", sc.Interface, sc.Direction))

	return resourceRTXShapeRead(ctx, d, meta)
}

func resourceRTXShapeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface, direction, err := parseShapeID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	log.Printf("[DEBUG] Reading shape: %s:%s", iface, direction)

	sc, err := apiClient.client.GetShape(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] Shape %s:%s not found, removing from state", iface, direction)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read shape: %v", err)
	}

	if err := d.Set("interface", sc.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("direction", sc.Direction); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("shape_average", sc.ShapeAverage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("shape_burst", sc.ShapeBurst); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXShapeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	sc := buildShapeConfigFromResourceData(d)

	log.Printf("[DEBUG] Updating shape: %+v", sc)

	err := apiClient.client.UpdateShape(ctx, sc)
	if err != nil {
		return diag.Errorf("Failed to update shape: %v", err)
	}

	return resourceRTXShapeRead(ctx, d, meta)
}

func resourceRTXShapeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface, direction, err := parseShapeID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	log.Printf("[DEBUG] Deleting shape: %s:%s", iface, direction)

	err = apiClient.client.DeleteShape(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete shape: %v", err)
	}

	return nil
}

func resourceRTXShapeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as "interface:direction" format
	iface, direction, err := parseShapeID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'interface:direction' (e.g., 'lan1:output'): %v", err)
	}

	log.Printf("[DEBUG] Importing shape: %s:%s", iface, direction)

	sc, err := apiClient.client.GetShape(ctx, iface, direction)
	if err != nil {
		return nil, fmt.Errorf("failed to import shape %s:%s: %v", iface, direction, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", iface, direction))
	d.Set("interface", sc.Interface)
	d.Set("direction", sc.Direction)
	d.Set("shape_average", sc.ShapeAverage)
	d.Set("shape_burst", sc.ShapeBurst)

	return []*schema.ResourceData{d}, nil
}

func buildShapeConfigFromResourceData(d *schema.ResourceData) client.ShapeConfig {
	sc := client.ShapeConfig{
		Interface:    d.Get("interface").(string),
		Direction:    d.Get("direction").(string),
		ShapeAverage: d.Get("shape_average").(int),
	}

	if v, ok := d.GetOk("shape_burst"); ok {
		sc.ShapeBurst = v.(int)
	}

	return sc
}

func parseShapeID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format 'interface:direction', got %q", id)
	}

	iface := parts[0]
	direction := parts[1]

	if iface == "" {
		return "", "", fmt.Errorf("interface cannot be empty")
	}
	if direction != "input" && direction != "output" {
		return "", "", fmt.Errorf("direction must be 'input' or 'output', got %q", direction)
	}

	return iface, direction, nil
}
