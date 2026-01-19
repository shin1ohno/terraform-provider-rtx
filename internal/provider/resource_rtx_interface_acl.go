package provider

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

var interfaceACLNameRegex = regexp.MustCompile(`^(lan[0-9]+|pp[0-9]+|tunnel[0-9]+|bridge[0-9]+|vlan[0-9]+)$`)

// resourceRTXInterfaceACL returns the schema for the rtx_interface_acl resource
func resourceRTXInterfaceACL() *schema.Resource {
	return &schema.Resource{
		Description: "Manages ACL bindings to an interface on RTX routers. " +
			"This resource applies access control lists (ACLs) to interfaces for traffic filtering.",

		CreateContext: resourceRTXInterfaceACLCreate,
		ReadContext:   resourceRTXInterfaceACLRead,
		UpdateContext: resourceRTXInterfaceACLUpdate,
		DeleteContext: resourceRTXInterfaceACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXInterfaceACLImport,
		},

		Schema: map[string]*schema.Schema{
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Interface name (lan1, lan2, pp1, tunnel1, etc.)",
				ValidateFunc: validation.StringMatch(
					interfaceACLNameRegex,
					"must be a valid interface name (lan1, lan2, pp1, tunnel1, etc.)",
				),
			},
			"ip_access_group_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound IPv4 access list name",
			},
			"ip_access_group_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound IPv4 access list name",
			},
			"ipv6_access_group_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound IPv6 access list name",
			},
			"ipv6_access_group_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound IPv6 access list name",
			},
			"dynamic_filters_in": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Inbound dynamic filter numbers (for stateful inspection)",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"dynamic_filters_out": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Outbound dynamic filter numbers (for stateful inspection)",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"ipv6_dynamic_filters_in": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Inbound IPv6 dynamic filter numbers",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"ipv6_dynamic_filters_out": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Outbound IPv6 dynamic filter numbers",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func resourceRTXInterfaceACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildInterfaceACLFromResourceData(d)

	log.Printf("[DEBUG] Creating interface ACL for: %s", acl.Interface)

	err := apiClient.client.CreateInterfaceACL(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create interface ACL: %v", err)
	}

	d.SetId(acl.Interface)

	return resourceRTXInterfaceACLRead(ctx, d, meta)
}

func resourceRTXInterfaceACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface := d.Id()

	log.Printf("[DEBUG] Reading interface ACL: %s", iface)

	acl, err := apiClient.client.GetInterfaceACL(ctx, iface)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Interface ACL %s not found, removing from state", iface)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read interface ACL: %v", err)
	}

	d.Set("interface", acl.Interface)
	d.Set("ip_access_group_in", acl.IPAccessGroupIn)
	d.Set("ip_access_group_out", acl.IPAccessGroupOut)
	d.Set("ipv6_access_group_in", acl.IPv6AccessGroupIn)
	d.Set("ipv6_access_group_out", acl.IPv6AccessGroupOut)
	d.Set("dynamic_filters_in", acl.DynamicFiltersIn)
	d.Set("dynamic_filters_out", acl.DynamicFiltersOut)
	d.Set("ipv6_dynamic_filters_in", acl.IPv6DynamicFiltersIn)
	d.Set("ipv6_dynamic_filters_out", acl.IPv6DynamicFiltersOut)

	return nil
}

func resourceRTXInterfaceACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildInterfaceACLFromResourceData(d)

	log.Printf("[DEBUG] Updating interface ACL for: %s", acl.Interface)

	err := apiClient.client.UpdateInterfaceACL(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update interface ACL: %v", err)
	}

	return resourceRTXInterfaceACLRead(ctx, d, meta)
}

func resourceRTXInterfaceACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface := d.Id()

	log.Printf("[DEBUG] Deleting interface ACL for: %s", iface)

	err := apiClient.client.DeleteInterfaceACL(ctx, iface)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete interface ACL: %v", err)
	}

	return nil
}

func resourceRTXInterfaceACLImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	iface := d.Id()

	log.Printf("[DEBUG] Importing interface ACL: %s", iface)

	acl, err := apiClient.client.GetInterfaceACL(ctx, iface)
	if err != nil {
		return nil, fmt.Errorf("failed to import interface ACL %s: %v", iface, err)
	}

	d.SetId(iface)
	d.Set("interface", acl.Interface)
	d.Set("ip_access_group_in", acl.IPAccessGroupIn)
	d.Set("ip_access_group_out", acl.IPAccessGroupOut)
	d.Set("ipv6_access_group_in", acl.IPv6AccessGroupIn)
	d.Set("ipv6_access_group_out", acl.IPv6AccessGroupOut)
	d.Set("dynamic_filters_in", acl.DynamicFiltersIn)
	d.Set("dynamic_filters_out", acl.DynamicFiltersOut)
	d.Set("ipv6_dynamic_filters_in", acl.IPv6DynamicFiltersIn)
	d.Set("ipv6_dynamic_filters_out", acl.IPv6DynamicFiltersOut)

	return []*schema.ResourceData{d}, nil
}

func buildInterfaceACLFromResourceData(d *schema.ResourceData) client.InterfaceACL {
	acl := client.InterfaceACL{
		Interface:          d.Get("interface").(string),
		IPAccessGroupIn:    d.Get("ip_access_group_in").(string),
		IPAccessGroupOut:   d.Get("ip_access_group_out").(string),
		IPv6AccessGroupIn:  d.Get("ipv6_access_group_in").(string),
		IPv6AccessGroupOut: d.Get("ipv6_access_group_out").(string),
	}

	if v, ok := d.GetOk("dynamic_filters_in"); ok {
		for _, num := range v.([]interface{}) {
			acl.DynamicFiltersIn = append(acl.DynamicFiltersIn, num.(int))
		}
	}
	if v, ok := d.GetOk("dynamic_filters_out"); ok {
		for _, num := range v.([]interface{}) {
			acl.DynamicFiltersOut = append(acl.DynamicFiltersOut, num.(int))
		}
	}
	if v, ok := d.GetOk("ipv6_dynamic_filters_in"); ok {
		for _, num := range v.([]interface{}) {
			acl.IPv6DynamicFiltersIn = append(acl.IPv6DynamicFiltersIn, num.(int))
		}
	}
	if v, ok := d.GetOk("ipv6_dynamic_filters_out"); ok {
		for _, num := range v.([]interface{}) {
			acl.IPv6DynamicFiltersOut = append(acl.IPv6DynamicFiltersOut, num.(int))
		}
	}

	return acl
}
