package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

var interfaceMACACLNameRegex = regexp.MustCompile(`^(lan[0-9]+|bridge[0-9]+|vlan[0-9]+)$`)

// resourceRTXInterfaceMACACL returns the schema for the rtx_interface_mac_acl resource
func resourceRTXInterfaceMACACL() *schema.Resource {
	return &schema.Resource{
		Description: "Manages MAC ACL bindings to an interface on RTX routers. " +
			"This resource applies MAC address access lists to interfaces for Layer 2 filtering.",

		CreateContext: resourceRTXInterfaceMACACLCreate,
		ReadContext:   resourceRTXInterfaceMACACLRead,
		UpdateContext: resourceRTXInterfaceMACACLUpdate,
		DeleteContext: resourceRTXInterfaceMACACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXInterfaceMACACLImport,
		},

		Schema: map[string]*schema.Schema{
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Interface name (lan1, lan2, bridge1, vlan1, etc.)",
				ValidateFunc: validation.StringMatch(
					interfaceMACACLNameRegex,
					"must be a valid interface name (lan1, lan2, bridge1, vlan1, etc.)",
				),
			},
			"mac_access_group_in": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Inbound MAC access list name",
			},
			"mac_access_group_out": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Outbound MAC access list name",
			},
		},
	}
}

func resourceRTXInterfaceMACACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface_mac_acl", d.Id())
	acl := buildInterfaceMACACLFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface_mac_acl").Msgf("Creating interface MAC ACL for: %s", acl.Interface)

	err := apiClient.client.CreateInterfaceMACACL(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create interface MAC ACL: %v", err)
	}

	d.SetId(acl.Interface)

	return resourceRTXInterfaceMACACLRead(ctx, d, meta)
}

func resourceRTXInterfaceMACACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface_mac_acl", d.Id())
	iface := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface_mac_acl").Msgf("Reading interface MAC ACL: %s", iface)

	acl, err := apiClient.client.GetInterfaceMACACL(ctx, iface)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Warn().Str("resource", "rtx_interface_mac_acl").Msgf("Interface MAC ACL %s not found, removing from state", iface)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read interface MAC ACL: %v", err)
	}

	d.Set("interface", acl.Interface)
	d.Set("mac_access_group_in", acl.MACAccessGroupIn)
	d.Set("mac_access_group_out", acl.MACAccessGroupOut)

	return nil
}

func resourceRTXInterfaceMACACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface_mac_acl", d.Id())
	acl := buildInterfaceMACACLFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface_mac_acl").Msgf("Updating interface MAC ACL for: %s", acl.Interface)

	err := apiClient.client.UpdateInterfaceMACACL(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update interface MAC ACL: %v", err)
	}

	return resourceRTXInterfaceMACACLRead(ctx, d, meta)
}

func resourceRTXInterfaceMACACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_interface_mac_acl", d.Id())
	iface := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface_mac_acl").Msgf("Deleting interface MAC ACL for: %s", iface)

	err := apiClient.client.DeleteInterfaceMACACL(ctx, iface)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete interface MAC ACL: %v", err)
	}

	return nil
}

func resourceRTXInterfaceMACACLImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	iface := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_interface_mac_acl").Msgf("Importing interface MAC ACL: %s", iface)

	acl, err := apiClient.client.GetInterfaceMACACL(ctx, iface)
	if err != nil {
		return nil, fmt.Errorf("failed to import interface MAC ACL %s: %v", iface, err)
	}

	d.SetId(iface)
	d.Set("interface", acl.Interface)
	d.Set("mac_access_group_in", acl.MACAccessGroupIn)
	d.Set("mac_access_group_out", acl.MACAccessGroupOut)

	return []*schema.ResourceData{d}, nil
}

func buildInterfaceMACACLFromResourceData(d *schema.ResourceData) client.InterfaceMACACL {
	return client.InterfaceMACACL{
		Interface:         d.Get("interface").(string),
		MACAccessGroupIn:  d.Get("mac_access_group_in").(string),
		MACAccessGroupOut: d.Get("mac_access_group_out").(string),
	}
}
