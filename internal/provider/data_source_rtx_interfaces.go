package provider

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRTXInterfaces() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get network interface information from an RTX router.",

		ReadContext: dataSourceRTXInterfacesRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal identifier for this data source.",
			},
			"interfaces": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of network interfaces on the RTX router.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The interface name (e.g., LAN1, WAN1, PP1, VLAN1).",
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The interface type: lan, wan, pp, or vlan.",
						},
						"admin_up": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the interface is administratively up.",
						},
						"link_up": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the physical link is up.",
						},
						"mac": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The MAC address of the interface.",
						},
						"ipv4": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IPv4 address assigned to the interface.",
						},
						"ipv6": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IPv6 address assigned to the interface.",
						},
						"mtu": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The Maximum Transmission Unit (MTU) of the interface.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A description of the interface.",
						},
						"attributes": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Additional model-specific attributes.",
						},
					},
				},
			},
		},
	}
}

func dataSourceRTXInterfacesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	apiClient := meta.(*apiClient)

	// Get interfaces information from the router
	interfaces, err := apiClient.client.GetInterfaces(ctx)
	if err != nil {
		return diag.Errorf("Failed to retrieve interfaces information: %v", err)
	}

	// Convert interfaces to schema format
	interfacesData := make([]interface{}, len(interfaces))
	for i, iface := range interfaces {
		interfaceMap := map[string]interface{}{
			"name":     iface.Name,
			"kind":     iface.Kind,
			"admin_up": iface.AdminUp,
			"link_up":  iface.LinkUp,
		}

		// Always set all fields - Terraform will handle empty values appropriately
		interfaceMap["mac"] = iface.MAC
		interfaceMap["ipv4"] = iface.IPv4
		interfaceMap["ipv6"] = iface.IPv6
		interfaceMap["mtu"] = iface.MTU
		interfaceMap["description"] = iface.Description
		// Convert map[string]string to map[string]interface{} for Terraform
		attributes := make(map[string]interface{})
		for k, v := range iface.Attributes {
			attributes[k] = v
		}
		interfaceMap["attributes"] = attributes

		interfacesData[i] = interfaceMap
	}

	// Set the resource data
	if err := d.Set("interfaces", interfacesData); err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID based on the interfaces information
	h := md5.New()
	for _, iface := range interfaces {
		h.Write([]byte(fmt.Sprintf("%s-%s-%v-%v-%s-%s",
			iface.Name,
			iface.Kind,
			iface.AdminUp,
			iface.LinkUp,
			iface.MAC,
			iface.IPv4,
		)))
	}
	id := fmt.Sprintf("%x", h.Sum(nil))
	d.SetId(id)

	return diags
}
