package provider

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXAccessListExtended() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPv4 extended access lists (ACLs) on RTX routers. Extended ACLs provide granular control over packet filtering based on source/destination addresses, protocols, and ports.",
		CreateContext: resourceRTXAccessListExtendedCreate,
		ReadContext:   resourceRTXAccessListExtendedRead,
		UpdateContext: resourceRTXAccessListExtendedUpdate,
		DeleteContext: resourceRTXAccessListExtendedDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListExtendedImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the access list (used as identifier)",
			},
			"entry": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of ACL entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sequence": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Sequence number (determines order, typically 10, 20, 30...)",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"ace_rule_action": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Action: 'permit' or 'deny'",
							ValidateFunc: validation.StringInSlice([]string{"permit", "deny"}, false),
						},
						"ace_rule_protocol": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Protocol: tcp, udp, icmp, ip, gre, esp, ah, or *",
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "icmp", "ip", "gre", "esp", "ah", "*"}, false),
						},
						"source_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any source address",
						},
						"source_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source IP address (e.g., '192.168.1.0')",
						},
						"source_prefix_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source wildcard mask (e.g., '0.0.0.255')",
						},
						"source_port_equal": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port equals (e.g., '80', '443')",
						},
						"source_port_range": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port range (e.g., '1024-65535')",
						},
						"destination_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any destination address",
						},
						"destination_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination IP address (e.g., '10.0.0.0')",
						},
						"destination_prefix_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination wildcard mask (e.g., '0.0.0.255')",
						},
						"destination_port_equal": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port equals (e.g., '80', '443')",
						},
						"destination_port_range": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port range (e.g., '1024-65535')",
						},
						"established": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match established TCP connections (ACK or RST flag set)",
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

		CustomizeDiff: validateAccessListExtendedEntries,
	}
}

func validateAccessListExtendedEntries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	entries := diff.Get("entry").([]interface{})

	for i, e := range entries {
		entry := e.(map[string]interface{})

		sourceAny := entry["source_any"].(bool)
		sourcePrefix := entry["source_prefix"].(string)
		destAny := entry["destination_any"].(bool)
		destPrefix := entry["destination_prefix"].(string)
		protocol := entry["ace_rule_protocol"].(string)
		established := entry["established"].(bool)

		// Either source_any or source_prefix must be specified
		if !sourceAny && sourcePrefix == "" {
			return fmt.Errorf("entry[%d]: either source_any must be true or source_prefix must be specified", i)
		}

		// Either destination_any or destination_prefix must be specified
		if !destAny && destPrefix == "" {
			return fmt.Errorf("entry[%d]: either destination_any must be true or destination_prefix must be specified", i)
		}

		// Established is only valid for TCP
		if established && protocol != "tcp" {
			return fmt.Errorf("entry[%d]: established can only be set to true for tcp protocol", i)
		}
	}

	return nil
}

func resourceRTXAccessListExtendedCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildAccessListExtendedFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Creating access list extended: %+v", acl)

	err := apiClient.client.CreateAccessListExtended(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create access list extended: %v", err)
	}

	d.SetId(acl.Name)

	return resourceRTXAccessListExtendedRead(ctx, d, meta)
}

func resourceRTXAccessListExtendedRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Reading access list extended: %s", name)

	acl, err := apiClient.client.GetAccessListExtended(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Access list extended %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read access list extended: %v", err)
	}

	if err := d.Set("name", acl.Name); err != nil {
		return diag.FromErr(err)
	}

	entries := flattenAccessListExtendedEntries(acl.Entries)
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXAccessListExtendedUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	acl := buildAccessListExtendedFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Updating access list extended: %+v", acl)

	err := apiClient.client.UpdateAccessListExtended(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update access list extended: %v", err)
	}

	return resourceRTXAccessListExtendedRead(ctx, d, meta)
}

func resourceRTXAccessListExtendedDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Deleting access list extended: %s", name)

	err := apiClient.client.DeleteAccessListExtended(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete access list extended: %v", err)
	}

	return nil
}

func resourceRTXAccessListExtendedImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Importing access list extended: %s", name)

	acl, err := apiClient.client.GetAccessListExtended(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import access list extended %s: %v", name, err)
	}

	d.SetId(name)
	d.Set("name", acl.Name)

	entries := flattenAccessListExtendedEntries(acl.Entries)
	d.Set("entry", entries)

	return []*schema.ResourceData{d}, nil
}

func buildAccessListExtendedFromResourceData(d *schema.ResourceData) client.AccessListExtended {
	acl := client.AccessListExtended{
		Name:    d.Get("name").(string),
		Entries: expandAccessListExtendedEntries(d.Get("entry").([]interface{})),
	}
	return acl
}

func expandAccessListExtendedEntries(entries []interface{}) []client.AccessListExtendedEntry {
	result := make([]client.AccessListExtendedEntry, 0, len(entries))

	for _, e := range entries {
		entry := e.(map[string]interface{})

		aclEntry := client.AccessListExtendedEntry{
			Sequence:        entry["sequence"].(int),
			AceRuleAction:   entry["ace_rule_action"].(string),
			AceRuleProtocol: entry["ace_rule_protocol"].(string),
			SourceAny:       entry["source_any"].(bool),
			DestinationAny:  entry["destination_any"].(bool),
			Established:     entry["established"].(bool),
			Log:             entry["log"].(bool),
		}

		if v, ok := entry["source_prefix"].(string); ok && v != "" {
			aclEntry.SourcePrefix = v
		}
		if v, ok := entry["source_prefix_mask"].(string); ok && v != "" {
			aclEntry.SourcePrefixMask = v
		}
		if v, ok := entry["source_port_equal"].(string); ok && v != "" {
			aclEntry.SourcePortEqual = v
		}
		if v, ok := entry["source_port_range"].(string); ok && v != "" {
			aclEntry.SourcePortRange = v
		}
		if v, ok := entry["destination_prefix"].(string); ok && v != "" {
			aclEntry.DestinationPrefix = v
		}
		if v, ok := entry["destination_prefix_mask"].(string); ok && v != "" {
			aclEntry.DestinationPrefixMask = v
		}
		if v, ok := entry["destination_port_equal"].(string); ok && v != "" {
			aclEntry.DestinationPortEqual = v
		}
		if v, ok := entry["destination_port_range"].(string); ok && v != "" {
			aclEntry.DestinationPortRange = v
		}

		result = append(result, aclEntry)
	}

	return result
}

func flattenAccessListExtendedEntries(entries []client.AccessListExtendedEntry) []interface{} {
	result := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		e := map[string]interface{}{
			"sequence":                 entry.Sequence,
			"ace_rule_action":          entry.AceRuleAction,
			"ace_rule_protocol":        entry.AceRuleProtocol,
			"source_any":               entry.SourceAny,
			"source_prefix":            entry.SourcePrefix,
			"source_prefix_mask":       entry.SourcePrefixMask,
			"source_port_equal":        entry.SourcePortEqual,
			"source_port_range":        entry.SourcePortRange,
			"destination_any":          entry.DestinationAny,
			"destination_prefix":       entry.DestinationPrefix,
			"destination_prefix_mask":  entry.DestinationPrefixMask,
			"destination_port_equal":   entry.DestinationPortEqual,
			"destination_port_range":   entry.DestinationPortRange,
			"established":              entry.Established,
			"log":                      entry.Log,
		}
		result = append(result, e)
	}

	return result
}
