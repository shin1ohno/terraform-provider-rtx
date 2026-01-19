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
							Description:  "Action to take (permit or deny)",
							ValidateFunc: validation.StringInSlice([]string{"permit", "deny"}, false),
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
					},
				},
			},
		},
	}
}

func resourceRTXAccessListMACCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildAccessListMACFromResourceData(d)

	log.Printf("[DEBUG] Creating MAC access list: %s", acl.Name)

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

	log.Printf("[DEBUG] Reading MAC access list: %s", name)

	acl, err := apiClient.client.GetAccessListMAC(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] MAC access list %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read MAC access list: %v", err)
	}

	d.Set("name", acl.Name)

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
		}
		entries = append(entries, e)
	}
	d.Set("entry", entries)

	return nil
}

func resourceRTXAccessListMACUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildAccessListMACFromResourceData(d)

	log.Printf("[DEBUG] Updating MAC access list: %s", acl.Name)

	err := apiClient.client.UpdateAccessListMAC(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update MAC access list: %v", err)
	}

	return resourceRTXAccessListMACRead(ctx, d, meta)
}

func resourceRTXAccessListMACDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	log.Printf("[DEBUG] Deleting MAC access list: %s", name)

	err := apiClient.client.DeleteAccessListMAC(ctx, name)
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

	log.Printf("[DEBUG] Importing MAC access list: %s", name)

	acl, err := apiClient.client.GetAccessListMAC(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import MAC access list %s: %v", name, err)
	}

	d.SetId(name)
	d.Set("name", acl.Name)

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
		}
		entries = append(entries, e)
	}
	d.Set("entry", entries)

	return []*schema.ResourceData{d}, nil
}

func buildAccessListMACFromResourceData(d *schema.ResourceData) client.AccessListMAC {
	acl := client.AccessListMAC{
		Name:    d.Get("name").(string),
		Entries: make([]client.AccessListMACEntry, 0),
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
		}
		acl.Entries = append(acl.Entries, aclEntry)
	}

	return acl
}
