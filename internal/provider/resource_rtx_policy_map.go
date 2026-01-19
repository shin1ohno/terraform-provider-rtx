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

func resourceRTXPolicyMap() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages QoS policy-map configurations on RTX routers. Policy-maps define actions to take on classified traffic.",
		CreateContext: resourceRTXPolicyMapCreate,
		ReadContext:   resourceRTXPolicyMapRead,
		UpdateContext: resourceRTXPolicyMapUpdate,
		DeleteContext: resourceRTXPolicyMapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXPolicyMapImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The policy-map name. Must start with a letter and contain only alphanumeric characters, underscores, and hyphens.",
				ValidateFunc: validatePolicyMapName,
			},
			"class": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of class definitions within this policy-map",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Class name (references a class-map)",
						},
						"priority": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Priority level: 'high', 'normal', or 'low'",
							ValidateFunc: validation.StringInSlice([]string{"high", "normal", "low"}, false),
						},
						"bandwidth_percent": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Bandwidth percentage allocation (1-100)",
							ValidateFunc: validation.IntBetween(1, 100),
						},
						"police_cir": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Committed Information Rate in bps for policing",
						},
						"queue_limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Queue depth limit",
						},
					},
				},
			},
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			// Validate that total bandwidth_percent doesn't exceed 100%
			classesRaw := diff.Get("class").([]interface{})
			totalBandwidth := 0

			for _, classRaw := range classesRaw {
				classMap := classRaw.(map[string]interface{})
				if bw, ok := classMap["bandwidth_percent"].(int); ok {
					totalBandwidth += bw
				}
			}

			if totalBandwidth > 100 {
				return fmt.Errorf("total bandwidth_percent across all classes (%d%%) exceeds 100%%", totalBandwidth)
			}

			return nil
		},
	}
}

func resourceRTXPolicyMapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	pm := buildPolicyMapFromResourceData(d)

	log.Printf("[DEBUG] Creating policy-map: %+v", pm)

	err := apiClient.client.CreatePolicyMap(ctx, pm)
	if err != nil {
		return diag.Errorf("Failed to create policy-map: %v", err)
	}

	d.SetId(pm.Name)

	return resourceRTXPolicyMapRead(ctx, d, meta)
}

func resourceRTXPolicyMapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	log.Printf("[DEBUG] Reading policy-map: %s", name)

	pm, err := apiClient.client.GetPolicyMap(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] Policy-map %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read policy-map: %v", err)
	}

	if err := d.Set("name", pm.Name); err != nil {
		return diag.FromErr(err)
	}

	classes := make([]interface{}, len(pm.Classes))
	for i, class := range pm.Classes {
		classMap := map[string]interface{}{
			"name":              class.Name,
			"priority":          class.Priority,
			"bandwidth_percent": class.BandwidthPercent,
			"police_cir":        class.PoliceCIR,
			"queue_limit":       class.QueueLimit,
		}
		classes[i] = classMap
	}

	if err := d.Set("class", classes); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXPolicyMapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	pm := buildPolicyMapFromResourceData(d)

	log.Printf("[DEBUG] Updating policy-map: %+v", pm)

	err := apiClient.client.UpdatePolicyMap(ctx, pm)
	if err != nil {
		return diag.Errorf("Failed to update policy-map: %v", err)
	}

	return resourceRTXPolicyMapRead(ctx, d, meta)
}

func resourceRTXPolicyMapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	log.Printf("[DEBUG] Deleting policy-map: %s", name)

	err := apiClient.client.DeletePolicyMap(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete policy-map: %v", err)
	}

	return nil
}

func resourceRTXPolicyMapImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	log.Printf("[DEBUG] Importing policy-map: %s", name)

	pm, err := apiClient.client.GetPolicyMap(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import policy-map %s: %v", name, err)
	}

	d.SetId(pm.Name)
	d.Set("name", pm.Name)

	classes := make([]interface{}, len(pm.Classes))
	for i, class := range pm.Classes {
		classMap := map[string]interface{}{
			"name":              class.Name,
			"priority":          class.Priority,
			"bandwidth_percent": class.BandwidthPercent,
			"police_cir":        class.PoliceCIR,
			"queue_limit":       class.QueueLimit,
		}
		classes[i] = classMap
	}
	d.Set("class", classes)

	return []*schema.ResourceData{d}, nil
}

func buildPolicyMapFromResourceData(d *schema.ResourceData) client.PolicyMap {
	pm := client.PolicyMap{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("class"); ok {
		classesRaw := v.([]interface{})
		pm.Classes = make([]client.PolicyMapClass, len(classesRaw))

		for i, classRaw := range classesRaw {
			classMap := classRaw.(map[string]interface{})
			pm.Classes[i] = client.PolicyMapClass{
				Name: classMap["name"].(string),
			}

			if priority, ok := classMap["priority"].(string); ok {
				pm.Classes[i].Priority = priority
			}
			if bw, ok := classMap["bandwidth_percent"].(int); ok {
				pm.Classes[i].BandwidthPercent = bw
			}
			if cir, ok := classMap["police_cir"].(int); ok {
				pm.Classes[i].PoliceCIR = cir
			}
			if ql, ok := classMap["queue_limit"].(int); ok {
				pm.Classes[i].QueueLimit = ql
			}
		}
	}

	return pm
}

func validatePolicyMapName(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	// Must start with a letter
	if len(value) > 0 && !((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) {
		return nil, []error{fmt.Errorf("%q must start with a letter, got %q", k, value)}
	}

	// Must contain only alphanumeric, underscore, and hyphen
	for _, c := range value {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return nil, []error{fmt.Errorf("%q must contain only letters, numbers, underscores, and hyphens, got %q", k, value)}
		}
	}

	return nil, nil
}
