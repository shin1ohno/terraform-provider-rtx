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

func resourceRTXClassMap() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages QoS class-map configurations on RTX routers. Class-maps classify traffic based on various match criteria.",
		CreateContext: resourceRTXClassMapCreate,
		ReadContext:   resourceRTXClassMapRead,
		UpdateContext: resourceRTXClassMapUpdate,
		DeleteContext: resourceRTXClassMapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXClassMapImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The class-map name. Must start with a letter and contain only alphanumeric characters, underscores, and hyphens.",
				ValidateFunc: validateClassMapName,
			},
			"match_protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Protocol to match (e.g., 'sip', 'http', 'ftp')",
			},
			"match_destination_port": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of destination ports to match",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 65535),
				},
			},
			"match_source_port": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of source ports to match",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 65535),
				},
			},
			"match_dscp": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DSCP value to match (e.g., 'ef', 'af11', '46')",
			},
			"match_filter": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "IP filter number to reference for matching (1-65535)",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
		},
	}
}

func resourceRTXClassMapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	cm := buildClassMapFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_class_map").Msgf("Creating class-map: %+v", cm)

	err := apiClient.client.CreateClassMap(ctx, cm)
	if err != nil {
		return diag.Errorf("Failed to create class-map: %v", err)
	}

	d.SetId(cm.Name)

	return resourceRTXClassMapRead(ctx, d, meta)
}

func resourceRTXClassMapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_class_map").Msgf("Reading class-map: %s", name)

	cm, err := apiClient.client.GetClassMap(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_class_map").Msgf("Class-map %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read class-map: %v", err)
	}

	if err := d.Set("name", cm.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_protocol", cm.MatchProtocol); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_destination_port", cm.MatchDestinationPort); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_source_port", cm.MatchSourcePort); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_dscp", cm.MatchDSCP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_filter", cm.MatchFilter); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXClassMapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	cm := buildClassMapFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_class_map").Msgf("Updating class-map: %+v", cm)

	err := apiClient.client.UpdateClassMap(ctx, cm)
	if err != nil {
		return diag.Errorf("Failed to update class-map: %v", err)
	}

	return resourceRTXClassMapRead(ctx, d, meta)
}

func resourceRTXClassMapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_class_map").Msgf("Deleting class-map: %s", name)

	err := apiClient.client.DeleteClassMap(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete class-map: %v", err)
	}

	return nil
}

func resourceRTXClassMapImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_class_map").Msgf("Importing class-map: %s", name)

	cm, err := apiClient.client.GetClassMap(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import class-map %s: %v", name, err)
	}

	d.SetId(cm.Name)
	d.Set("name", cm.Name)
	d.Set("match_protocol", cm.MatchProtocol)
	d.Set("match_destination_port", cm.MatchDestinationPort)
	d.Set("match_source_port", cm.MatchSourcePort)
	d.Set("match_dscp", cm.MatchDSCP)
	d.Set("match_filter", cm.MatchFilter)

	return []*schema.ResourceData{d}, nil
}

func buildClassMapFromResourceData(d *schema.ResourceData) client.ClassMap {
	cm := client.ClassMap{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("match_protocol"); ok {
		cm.MatchProtocol = v.(string)
	}

	if v, ok := d.GetOk("match_destination_port"); ok {
		ports := v.([]interface{})
		cm.MatchDestinationPort = make([]int, len(ports))
		for i, p := range ports {
			cm.MatchDestinationPort[i] = p.(int)
		}
	}

	if v, ok := d.GetOk("match_source_port"); ok {
		ports := v.([]interface{})
		cm.MatchSourcePort = make([]int, len(ports))
		for i, p := range ports {
			cm.MatchSourcePort[i] = p.(int)
		}
	}

	if v, ok := d.GetOk("match_dscp"); ok {
		cm.MatchDSCP = v.(string)
	}

	if v, ok := d.GetOk("match_filter"); ok {
		cm.MatchFilter = v.(int)
	}

	return cm
}

func validateClassMapName(v interface{}, k string) ([]string, []error) {
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
