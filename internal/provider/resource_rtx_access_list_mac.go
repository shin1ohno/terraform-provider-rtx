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

// resourceRTXAccessListMAC returns the schema for the rtx_access_list_mac resource
func resourceRTXAccessListMAC() *schema.Resource {
	return &schema.Resource{
		Description: "Manages MAC address access lists on RTX routers. " +
			"MAC ACLs filter traffic based on source and destination MAC addresses.",

		CreateContext: resourceRTXAccessListMACCreate,
		ReadContext:   resourceRTXAccessListMACRead,
		UpdateContext: resourceRTXAccessListMACUpdate,
		DeleteContext: resourceRTXAccessListMACDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListMACImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Access list name (identifier)",
			},
			"filter_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Optional RTX filter ID to enable numeric ethernet filter mode",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"apply": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Optional application of ethernet filters to an interface",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interface": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Interface to apply filters (e.g., lan1)",
						},
						"direction": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Direction to apply filters (in or out)",
							ValidateFunc: validation.StringInSlice([]string{"in", "out"}, false),
						},
						"filter_ids": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "List of filter IDs to apply in order",
							Elem: &schema.Schema{
								Type:         schema.TypeInt,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
				},
			},
			"entry": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of MAC ACL entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sequence": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Sequence number (determines order of evaluation)",
							ValidateFunc: validation.IntBetween(1, 99999),
						},
						"ace_action": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Action to take (permit/deny or RTX pass/reject with log/nolog)",
							ValidateFunc: validation.StringInSlice([]string{"permit", "deny", "pass-log", "pass-nolog", "reject-log", "reject-nolog", "pass", "reject"}, false),
						},
						"source_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any source MAC address",
						},
						"source_address": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source MAC address (e.g., 00:00:00:00:00:00)",
						},
						"source_address_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source MAC wildcard mask",
						},
						"destination_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any destination MAC address",
						},
						"destination_address": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination MAC address (e.g., 00:00:00:00:00:00)",
						},
						"destination_address_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination MAC wildcard mask",
						},
						"ether_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Ethernet type (e.g., 0x0800 for IPv4, 0x0806 for ARP)",
						},
						"vlan_id": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "VLAN ID to match",
							ValidateFunc: validation.IntBetween(1, 4094),
						},
						"log": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable logging for this entry",
						},
						"filter_id": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Explicit filter number for this entry (overrides sequence)",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"dhcp_match": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "DHCP-based match settings",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "DHCP match type (dhcp-bind or dhcp-not-bind)",
										ValidateFunc: validation.StringInSlice([]string{"dhcp-bind", "dhcp-not-bind"}, false),
									},
									"scope": {
										Type:         schema.TypeInt,
										Optional:     true,
										Description:  "DHCP scope number",
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"offset": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Offset for byte matching",
							ValidateFunc: validation.IntAtLeast(0),
						},
						"byte_list": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Byte list (hex) for offset matching",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceRTXAccessListMACCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildAccessListMACFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Creating MAC access list: %s", acl.Name)

	err := apiClient.client.CreateAccessListMAC(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create MAC access list: %v", err)
	}

	d.SetId(acl.Name)

	return resourceRTXAccessListMACRead(ctx, d, meta)
}

func resourceRTXAccessListMACRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Reading MAC access list: %s", name)

	acl, err := apiClient.client.GetAccessListMAC(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Warn().Str("resource", "rtx_access_list_mac").Msgf("MAC access list %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read MAC access list: %v", err)
	}

	d.Set("name", acl.Name)
	d.Set("filter_id", acl.FilterID)

	entries := make([]map[string]interface{}, 0, len(acl.Entries))
	for _, entry := range acl.Entries {
		e := map[string]interface{}{
			"sequence":                 entry.Sequence,
			"ace_action":               entry.AceAction,
			"source_any":               entry.SourceAny,
			"source_address":           entry.SourceAddress,
			"source_address_mask":      entry.SourceAddressMask,
			"destination_any":          entry.DestinationAny,
			"destination_address":      entry.DestinationAddress,
			"destination_address_mask": entry.DestinationAddressMask,
			"ether_type":               entry.EtherType,
			"vlan_id":                  entry.VlanID,
			"log":                      entry.Log,
			"filter_id":                entry.FilterID,
			"offset":                   entry.Offset,
			"byte_list":                entry.ByteList,
		}

		if entry.DHCPType != "" {
			e["dhcp_match"] = []map[string]interface{}{
				{
					"type":  entry.DHCPType,
					"scope": entry.DHCPScope,
				},
			}
		}

		entries = append(entries, e)
	}
	d.Set("entry", entries)

	if acl.Apply != nil {
		d.Set("apply", []map[string]interface{}{
			{
				"interface":  acl.Apply.Interface,
				"direction":  acl.Apply.Direction,
				"filter_ids": acl.Apply.FilterIDs,
			},
		})
	} else {
		d.Set("apply", nil)
	}

	return nil
}

func resourceRTXAccessListMACUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildAccessListMACFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Updating MAC access list: %s", acl.Name)

	err := apiClient.client.UpdateAccessListMAC(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update MAC access list: %v", err)
	}

	return resourceRTXAccessListMACRead(ctx, d, meta)
}

func resourceRTXAccessListMACDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Deleting MAC access list: %s", name)

	// Collect filter numbers to delete (filter_id overrides sequence)
	var filterNums []int
	entries := d.Get("entry").([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		num := entry["filter_id"].(int)
		if num == 0 {
			num = entry["sequence"].(int)
		}
		if num > 0 {
			filterNums = append(filterNums, num)
		}
	}

	err := apiClient.client.DeleteAccessListMAC(ctx, name, filterNums)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete MAC access list: %v", err)
	}

	return nil
}

func resourceRTXAccessListMACImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Importing MAC access list: %s", name)

	acl, err := apiClient.client.GetAccessListMAC(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import MAC access list %s: %v", name, err)
	}

	d.SetId(name)
	d.Set("name", acl.Name)
	d.Set("filter_id", acl.FilterID)

	entries := make([]map[string]interface{}, 0, len(acl.Entries))
	for _, entry := range acl.Entries {
		e := map[string]interface{}{
			"sequence":                 entry.Sequence,
			"ace_action":               entry.AceAction,
			"source_any":               entry.SourceAny,
			"source_address":           entry.SourceAddress,
			"source_address_mask":      entry.SourceAddressMask,
			"destination_any":          entry.DestinationAny,
			"destination_address":      entry.DestinationAddress,
			"destination_address_mask": entry.DestinationAddressMask,
			"ether_type":               entry.EtherType,
			"vlan_id":                  entry.VlanID,
			"log":                      entry.Log,
			"filter_id":                entry.FilterID,
			"offset":                   entry.Offset,
			"byte_list":                entry.ByteList,
		}
		if entry.DHCPType != "" {
			e["dhcp_match"] = []map[string]interface{}{
				{
					"type":  entry.DHCPType,
					"scope": entry.DHCPScope,
				},
			}
		}
		entries = append(entries, e)
	}
	d.Set("entry", entries)

	if acl.Apply != nil {
		d.Set("apply", []map[string]interface{}{
			{
				"interface":  acl.Apply.Interface,
				"direction":  acl.Apply.Direction,
				"filter_ids": acl.Apply.FilterIDs,
			},
		})
	} else {
		d.Set("apply", nil)
	}

	return []*schema.ResourceData{d}, nil
}

func buildAccessListMACFromResourceData(d *schema.ResourceData) client.AccessListMAC {
	acl := client.AccessListMAC{
		Name:     d.Get("name").(string),
		FilterID: d.Get("filter_id").(int),
		Entries:  make([]client.AccessListMACEntry, 0),
	}

	entries := d.Get("entry").([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		aclEntry := client.AccessListMACEntry{
			Sequence:               entry["sequence"].(int),
			AceAction:              entry["ace_action"].(string),
			SourceAny:              entry["source_any"].(bool),
			SourceAddress:          entry["source_address"].(string),
			SourceAddressMask:      entry["source_address_mask"].(string),
			DestinationAny:         entry["destination_any"].(bool),
			DestinationAddress:     entry["destination_address"].(string),
			DestinationAddressMask: entry["destination_address_mask"].(string),
			EtherType:              entry["ether_type"].(string),
			VlanID:                 entry["vlan_id"].(int),
			Log:                    entry["log"].(bool),
			FilterID:               entry["filter_id"].(int),
			Offset:                 entry["offset"].(int),
		}

		if aclEntry.FilterID == 0 && acl.FilterID > 0 {
			aclEntry.FilterID = acl.FilterID
		}

		if v, ok := entry["byte_list"].([]interface{}); ok {
			for _, b := range v {
				aclEntry.ByteList = append(aclEntry.ByteList, b.(string))
			}
		}

		if v, ok := entry["dhcp_match"].([]interface{}); ok && len(v) > 0 {
			m := v[0].(map[string]interface{})
			aclEntry.DHCPType = m["type"].(string)
			aclEntry.DHCPScope = m["scope"].(int)
		}

		acl.Entries = append(acl.Entries, aclEntry)
	}

	if v, ok := d.GetOk("apply"); ok {
		applyList := v.([]interface{})
		if len(applyList) > 0 {
			m := applyList[0].(map[string]interface{})
			var ids []int
			if rawIDs, ok := m["filter_ids"].([]interface{}); ok {
				for _, id := range rawIDs {
					ids = append(ids, id.(int))
				}
			}
			acl.Apply = &client.MACApply{
				Interface: m["interface"].(string),
				Direction: m["direction"].(string),
				FilterIDs: ids,
			}
		}
	}

	return acl
}
