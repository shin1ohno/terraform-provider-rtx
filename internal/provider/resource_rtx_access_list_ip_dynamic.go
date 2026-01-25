package provider

import (
	"context"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// resourceRTXAccessListIPDynamic returns the schema for the rtx_access_list_ip_dynamic resource
func resourceRTXAccessListIPDynamic() *schema.Resource {
	return &schema.Resource{
		Description: `Manages a named collection of IPv4 dynamic (stateful) IP filters on RTX routers.

Dynamic filters provide stateful packet inspection for various protocols. This resource groups
multiple dynamic filter entries under a single name for easier management and reference.

` + "```" + `hcl
resource "rtx_access_list_ip_dynamic" "outbound_stateful" {
  name = "outbound-stateful"

  entry {
    sequence    = 100
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = true
  }

  entry {
    sequence    = 101
    source      = "*"
    destination = "*"
    protocol    = "ftp"
  }

  entry {
    sequence    = 102
    source      = "*"
    destination = "*"
    protocol    = "dns"
    timeout     = 60
  }
}
` + "```" + `
`,
		CreateContext: resourceRTXAccessListIPDynamicCreate,
		ReadContext:   resourceRTXAccessListIPDynamicRead,
		UpdateContext: resourceRTXAccessListIPDynamicUpdate,
		DeleteContext: resourceRTXAccessListIPDynamicDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListIPDynamicImport,
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
				Description: "List of dynamic filter entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sequence": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Sequence number (determines order and filter number)",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"source": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Source address or '*' for any. Can be an IP address, network in CIDR notation, or '*'.",
						},
						"destination": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Destination address or '*' for any. Can be an IP address, network in CIDR notation, or '*'.",
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Protocol for stateful inspection. Valid values: ftp, www, smtp, pop3, dns, domain, " +
								"telnet, ssh, tcp, udp, *, tftp, submission, https, imap, imaps, pop3s, smtps, ldap, ldaps, bgp, sip, " +
								"ipsec-nat-t, ntp, snmp, rtsp, h323, pptp, l2tp, ike, esp.",
							ValidateFunc: validation.StringInSlice([]string{
								"ftp", "www", "smtp", "pop3", "dns", "domain", "telnet", "ssh",
								"tcp", "udp", "*",
								"tftp", "submission", "https", "imap", "imaps", "pop3s", "smtps",
								"ldap", "ldaps", "bgp", "sip", "ipsec-nat-t", "ntp", "snmp",
								"rtsp", "h323", "pptp", "l2tp", "ike", "esp",
							}, false),
						},
						"syslog": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable syslog logging for this filter.",
						},
						"timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Timeout value in seconds. If not specified, uses system default.",
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
		},
	}
}

func resourceRTXAccessListIPDynamicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", d.Id())
	acl := buildAccessListIPDynamicFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Creating dynamic IP access list: %s", acl.Name)

	err := apiClient.client.CreateAccessListIPDynamic(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create dynamic IP access list: %v", err)
	}

	d.SetId(acl.Name)

	return resourceRTXAccessListIPDynamicRead(ctx, d, meta)
}

func resourceRTXAccessListIPDynamicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Reading dynamic IP access list: %s", name)

	// Get current sequences from state to filter results
	// This prevents other access lists' filters from leaking into this resource's state
	currentSeqs := make(map[int]bool)
	if currentEntries, ok := d.GetOk("entry"); ok {
		for _, e := range currentEntries.([]interface{}) {
			entry := e.(map[string]interface{})
			if seq, ok := entry["sequence"].(int); ok && seq > 0 {
				currentSeqs[seq] = true
			}
		}
	}

	acl, err := apiClient.client.GetAccessListIPDynamic(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Warn().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Dynamic IP access list %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read dynamic IP access list: %v", err)
	}

	d.Set("name", acl.Name)

	entries := make([]map[string]interface{}, 0, len(acl.Entries))
	for _, entry := range acl.Entries {
		// Only include sequences that are already in state
		// This prevents filters from other access lists from appearing here
		// After import (no entries in state), this returns empty - terraform config defines entries
		if !currentSeqs[entry.Sequence] {
			continue
		}

		e := map[string]interface{}{
			"sequence":    entry.Sequence,
			"source":      entry.Source,
			"destination": entry.Destination,
			"protocol":    entry.Protocol,
			"syslog":      entry.Syslog,
		}
		if entry.Timeout != nil {
			e["timeout"] = *entry.Timeout
		}
		entries = append(entries, e)
	}
	d.Set("entry", entries)

	return nil
}

func resourceRTXAccessListIPDynamicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", d.Id())
	acl := buildAccessListIPDynamicFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Updating dynamic IP access list: %s", acl.Name)

	err := apiClient.client.UpdateAccessListIPDynamic(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update dynamic IP access list: %v", err)
	}

	return resourceRTXAccessListIPDynamicRead(ctx, d, meta)
}

func resourceRTXAccessListIPDynamicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip_dynamic", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Deleting dynamic IP access list: %s", name)

	// Collect filter numbers to delete
	var filterNums []int
	entries := d.Get("entry").([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		num := entry["sequence"].(int)
		if num > 0 {
			filterNums = append(filterNums, num)
		}
	}

	err := apiClient.client.DeleteAccessListIPDynamic(ctx, name, filterNums)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete dynamic IP access list: %v", err)
	}

	return nil
}

func resourceRTXAccessListIPDynamicImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip_dynamic").Msgf("Importing dynamic IP access list: %s", name)

	// Import only sets the name - entries are intentionally NOT imported.
	// This is because RTX doesn't track which filters belong to which "named list".
	// The Terraform configuration defines which entries belong to this access list.
	// After import, run `terraform apply` to bind the configured entries to this resource.
	d.SetId(name)
	d.Set("name", name)

	// Don't set entries - let Terraform config define them
	// This prevents filters from other access lists from being incorrectly imported

	return []*schema.ResourceData{d}, nil
}

func buildAccessListIPDynamicFromResourceData(d *schema.ResourceData) client.AccessListIPDynamic {
	acl := client.AccessListIPDynamic{
		Name:    d.Get("name").(string),
		Entries: make([]client.AccessListIPDynamicEntry, 0),
	}

	entries := d.Get("entry").([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		aclEntry := client.AccessListIPDynamicEntry{
			Sequence:    entry["sequence"].(int),
			Source:      entry["source"].(string),
			Destination: entry["destination"].(string),
			Protocol:    entry["protocol"].(string),
			Syslog:      entry["syslog"].(bool),
		}

		if v, ok := entry["timeout"]; ok && v.(int) > 0 {
			timeout := v.(int)
			aclEntry.Timeout = &timeout
		}

		acl.Entries = append(acl.Entries, aclEntry)
	}

	return acl
}
