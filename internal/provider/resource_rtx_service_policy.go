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

func resourceRTXServicePolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Attaches a QoS policy-map to an interface on RTX routers. Service policies control traffic entering or leaving an interface.",
		CreateContext: resourceRTXServicePolicyCreate,
		ReadContext:   resourceRTXServicePolicyRead,
		UpdateContext: resourceRTXServicePolicyUpdate,
		DeleteContext: resourceRTXServicePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXServicePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The interface to attach the policy to (e.g., 'lan1', 'wan1')",
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Direction for the policy: 'input' or 'output'",
				ValidateFunc: validation.StringInSlice([]string{"input", "output"}, false),
			},
			"policy_map": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The policy-map name or queue type to apply (e.g., 'priority', 'cbq', or a policy-map name)",
			},
		},
	}
}

func resourceRTXServicePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	sp := buildServicePolicyFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_service_policy").Msgf("Creating service-policy: %+v", sp)

	err := apiClient.client.CreateServicePolicy(ctx, sp)
	if err != nil {
		return diag.Errorf("Failed to create service-policy: %v", err)
	}

	// Resource ID format: interface:direction
	d.SetId(fmt.Sprintf("%s:%s", sp.Interface, sp.Direction))

	return resourceRTXServicePolicyRead(ctx, d, meta)
}

func resourceRTXServicePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface, direction, err := parseServicePolicyID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_service_policy").Msgf("Reading service-policy: %s:%s", iface, direction)

	sp, err := apiClient.client.GetServicePolicy(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_service_policy").Msgf("Service-policy %s:%s not found, removing from state", iface, direction)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read service-policy: %v", err)
	}

	if err := d.Set("interface", sp.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("direction", sp.Direction); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("policy_map", sp.PolicyMap); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXServicePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	sp := buildServicePolicyFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_service_policy").Msgf("Updating service-policy: %+v", sp)

	err := apiClient.client.UpdateServicePolicy(ctx, sp)
	if err != nil {
		return diag.Errorf("Failed to update service-policy: %v", err)
	}

	return resourceRTXServicePolicyRead(ctx, d, meta)
}

func resourceRTXServicePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface, direction, err := parseServicePolicyID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_service_policy").Msgf("Deleting service-policy: %s:%s", iface, direction)

	err = apiClient.client.DeleteServicePolicy(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete service-policy: %v", err)
	}

	return nil
}

func resourceRTXServicePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as "interface:direction" format
	iface, direction, err := parseServicePolicyID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'interface:direction' (e.g., 'lan1:output'): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_service_policy").Msgf("Importing service-policy: %s:%s", iface, direction)

	sp, err := apiClient.client.GetServicePolicy(ctx, iface, direction)
	if err != nil {
		return nil, fmt.Errorf("failed to import service-policy %s:%s: %v", iface, direction, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", iface, direction))
	d.Set("interface", sp.Interface)
	d.Set("direction", sp.Direction)
	d.Set("policy_map", sp.PolicyMap)

	return []*schema.ResourceData{d}, nil
}

func buildServicePolicyFromResourceData(d *schema.ResourceData) client.ServicePolicy {
	return client.ServicePolicy{
		Interface: d.Get("interface").(string),
		Direction: d.Get("direction").(string),
		PolicyMap: d.Get("policy_map").(string),
	}
}

func parseServicePolicyID(id string) (string, string, error) {
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
