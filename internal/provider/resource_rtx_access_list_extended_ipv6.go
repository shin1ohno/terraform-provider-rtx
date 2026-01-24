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

func resourceRTXAccessListExtendedIPv6() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPv6 extended access lists (ACLs) on RTX routers. Extended ACLs provide granular control over IPv6 packet filtering based on source/destination addresses, protocols, and ports.",
		CreateContext: resourceRTXAccessListExtendedIPv6Create,
		ReadContext:   resourceRTXAccessListExtendedIPv6Read,
		UpdateContext: resourceRTXAccessListExtendedIPv6Update,
		DeleteContext: resourceRTXAccessListExtendedIPv6Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListExtendedIPv6Import,
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
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Action: 'permit' or 'deny'",
							ValidateFunc:     validation.StringInSlice([]string{"permit", "deny"}, true),
							DiffSuppressFunc: SuppressCaseDiff, // ACL actions are case-insensitive
						},
						"ace_rule_protocol": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Protocol: tcp, udp, icmpv6, ipv6, or *",
							ValidateFunc:     validation.StringInSlice([]string{"tcp", "udp", "icmpv6", "ipv6", "ip", "*"}, true),
							DiffSuppressFunc: SuppressCaseDiff, // Protocol names are case-insensitive
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
							Description: "Source IPv6 address (e.g., '2001:db8::')",
						},
						"source_prefix_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Source prefix length (e.g., 64)",
							ValidateFunc: validation.IntBetween(0, 128),
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
							Description: "Destination IPv6 address (e.g., '2001:db8:1::')",
						},
						"destination_prefix_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Destination prefix length (e.g., 64)",
							ValidateFunc: validation.IntBetween(0, 128),
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

		CustomizeDiff: validateAccessListExtendedIPv6Entries,
	}
}

func validateAccessListExtendedIPv6Entries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

func resourceRTXAccessListExtendedIPv6Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", d.Id())
	acl := buildAccessListExtendedIPv6FromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Creating IPv6 access list extended: %+v", acl)

	err := apiClient.client.CreateAccessListExtendedIPv6(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create IPv6 access list extended: %v", err)
	}

	d.SetId(acl.Name)

	return resourceRTXAccessListExtendedIPv6Read(ctx, d, meta)
}

func resourceRTXAccessListExtendedIPv6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Reading IPv6 access list extended: %s", name)

	acl, err := apiClient.client.GetAccessListExtendedIPv6(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("IPv6 access list extended %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IPv6 access list extended: %v", err)
	}

	if err := d.Set("name", acl.Name); err != nil {
		return diag.FromErr(err)
	}

	entries := flattenAccessListExtendedIPv6Entries(acl.Entries)
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXAccessListExtendedIPv6Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", d.Id())
	acl := buildAccessListExtendedIPv6FromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Updating IPv6 access list extended: %+v", acl)

	err := apiClient.client.UpdateAccessListExtendedIPv6(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update IPv6 access list extended: %v", err)
	}

	return resourceRTXAccessListExtendedIPv6Read(ctx, d, meta)
}

func resourceRTXAccessListExtendedIPv6Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended_ipv6", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Deleting IPv6 access list extended: %s", name)

	err := apiClient.client.DeleteAccessListExtendedIPv6(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IPv6 access list extended: %v", err)
	}

	return nil
}

func resourceRTXAccessListExtendedIPv6Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended_ipv6").Msgf("Importing IPv6 access list extended: %s", name)

	acl, err := apiClient.client.GetAccessListExtendedIPv6(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import IPv6 access list extended %s: %v", name, err)
	}

	d.SetId(name)
	d.Set("name", acl.Name)

	entries := flattenAccessListExtendedIPv6Entries(acl.Entries)
	d.Set("entry", entries)

	return []*schema.ResourceData{d}, nil
}

func buildAccessListExtendedIPv6FromResourceData(d *schema.ResourceData) client.AccessListExtendedIPv6 {
	acl := client.AccessListExtendedIPv6{
		Name:    d.Get("name").(string),
		Entries: expandAccessListExtendedIPv6Entries(d.Get("entry").([]interface{})),
	}
	return acl
}

func expandAccessListExtendedIPv6Entries(entries []interface{}) []client.AccessListExtendedIPv6Entry {
	result := make([]client.AccessListExtendedIPv6Entry, 0, len(entries))

	for _, e := range entries {
		entry := e.(map[string]interface{})

		aclEntry := client.AccessListExtendedIPv6Entry{
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
		if v, ok := entry["source_prefix_length"].(int); ok && v > 0 {
			aclEntry.SourcePrefixLength = v
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
		if v, ok := entry["destination_prefix_length"].(int); ok && v > 0 {
			aclEntry.DestinationPrefixLength = v
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

func flattenAccessListExtendedIPv6Entries(entries []client.AccessListExtendedIPv6Entry) []interface{} {
	result := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		e := map[string]interface{}{
			"sequence":                  entry.Sequence,
			"ace_rule_action":           entry.AceRuleAction,
			"ace_rule_protocol":         entry.AceRuleProtocol,
			"source_any":                entry.SourceAny,
			"source_prefix":             entry.SourcePrefix,
			"source_prefix_length":      entry.SourcePrefixLength,
			"source_port_equal":         entry.SourcePortEqual,
			"source_port_range":         entry.SourcePortRange,
			"destination_any":           entry.DestinationAny,
			"destination_prefix":        entry.DestinationPrefix,
			"destination_prefix_length": entry.DestinationPrefixLength,
			"destination_port_equal":    entry.DestinationPortEqual,
			"destination_port_range":    entry.DestinationPortRange,
			"established":               entry.Established,
			"log":                       entry.Log,
		}
		result = append(result, e)
	}

	return result
}
