package provider

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXNATMasquerade() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages NAT masquerade (PAT/NAPT) configurations on RTX routers. NAT masquerade allows multiple internal hosts to share a single external IP address using port address translation.",
		CreateContext: resourceRTXNATMasqueradeCreate,
		ReadContext:   resourceRTXNATMasqueradeRead,
		UpdateContext: resourceRTXNATMasqueradeUpdate,
		DeleteContext: resourceRTXNATMasqueradeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXNATMasqueradeImport,
		},

		Schema: map[string]*schema.Schema{
			"descriptor_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "NAT descriptor ID (1-65535)",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"outer_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Outer (external) address: 'ipcp' for PPPoE-assigned address, interface name (e.g., 'pp1'), or specific IP address",
				ValidateFunc: func(v interface{}, k string) ([]string, []error) {
					value := v.(string)
					if value == "" {
						return nil, []error{fmt.Errorf("%q cannot be empty", k)}
					}
					// Valid values: "ipcp", interface names, or IP addresses
					// We allow any non-empty string here as the RTX router validates the actual value
					return nil, nil
				},
			},
			"inner_network": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Inner (internal) network range in format 'start_ip-end_ip' (e.g., '192.168.1.0-192.168.1.255')",
				ValidateFunc: validateIPRange,
			},
			"static_entry": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Static port mapping entries for port forwarding",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry_number": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Entry number for identification",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"inside_local": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Internal IP address",
							ValidateFunc: validateIPAddress,
						},
						"inside_local_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Internal port number (1-65535). Required for tcp/udp, omit for protocol-only entries (esp, ah, gre, icmp)",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"outside_global": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "ipcp",
							Description: "External IP address or 'ipcp' for PPPoE-assigned address",
						},
						"outside_global_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "External port number (1-65535). Required for tcp/udp, omit for protocol-only entries (esp, ah, gre, icmp)",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Protocol: 'tcp', 'udp' (require ports), or 'esp', 'ah', 'gre', 'icmp' (protocol-only, no ports)",
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "esp", "ah", "gre", "icmp", ""}, true),
						},
					},
				},
			},
		},
	}
}

func resourceRTXNATMasqueradeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	nat := buildNATMasqueradeFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_masquerade").Msgf("Creating NAT Masquerade: %+v", nat)

	err := apiClient.client.CreateNATMasquerade(ctx, nat)
	if err != nil {
		return diag.Errorf("Failed to create NAT masquerade: %v", err)
	}

	// Set resource ID as the descriptor_id
	d.SetId(strconv.Itoa(nat.DescriptorID))

	// Read back to ensure consistency
	return resourceRTXNATMasqueradeRead(ctx, d, meta)
}

func resourceRTXNATMasqueradeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the ID
	descriptorID, err := parseNATMasqueradeID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_masquerade").Msgf("Reading NAT Masquerade: %d", descriptorID)

	nat, err := apiClient.client.GetNATMasquerade(ctx, descriptorID)
	if err != nil {
		// Check if NAT masquerade doesn't exist
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_masquerade").Msgf("NAT Masquerade %d not found, removing from state", descriptorID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read NAT masquerade: %v", err)
	}

	// Update the state
	if err := d.Set("descriptor_id", nat.DescriptorID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("outer_address", nat.OuterAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("inner_network", nat.InnerNetwork); err != nil {
		return diag.FromErr(err)
	}

	// Convert static entries to schema format
	staticEntries := flattenStaticEntries(nat.StaticEntries)
	if err := d.Set("static_entry", staticEntries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXNATMasqueradeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	nat := buildNATMasqueradeFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_masquerade").Msgf("Updating NAT Masquerade: %+v", nat)

	err := apiClient.client.UpdateNATMasquerade(ctx, nat)
	if err != nil {
		return diag.Errorf("Failed to update NAT masquerade: %v", err)
	}

	return resourceRTXNATMasqueradeRead(ctx, d, meta)
}

func resourceRTXNATMasqueradeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the ID
	descriptorID, err := parseNATMasqueradeID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_masquerade").Msgf("Deleting NAT Masquerade: %d", descriptorID)

	err = apiClient.client.DeleteNATMasquerade(ctx, descriptorID)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete NAT masquerade: %v", err)
	}

	return nil
}

func resourceRTXNATMasqueradeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as descriptor_id
	descriptorID, err := parseNATMasqueradeID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected descriptor_id (e.g., '1'): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_masquerade").Msgf("Importing NAT Masquerade: %d", descriptorID)

	// Verify NAT masquerade exists
	nat, err := apiClient.client.GetNATMasquerade(ctx, descriptorID)
	if err != nil {
		return nil, fmt.Errorf("failed to import NAT masquerade %d: %v", descriptorID, err)
	}

	// Set all attributes
	d.SetId(strconv.Itoa(descriptorID))
	d.Set("descriptor_id", nat.DescriptorID)
	d.Set("outer_address", nat.OuterAddress)
	d.Set("inner_network", nat.InnerNetwork)
	d.Set("static_entry", flattenStaticEntries(nat.StaticEntries))

	return []*schema.ResourceData{d}, nil
}

// buildNATMasqueradeFromResourceData creates a NATMasquerade from Terraform resource data
func buildNATMasqueradeFromResourceData(d *schema.ResourceData) client.NATMasquerade {
	nat := client.NATMasquerade{
		DescriptorID: d.Get("descriptor_id").(int),
		OuterAddress: d.Get("outer_address").(string),
		InnerNetwork: d.Get("inner_network").(string),
	}

	// Process static entries
	if v, ok := d.GetOk("static_entry"); ok {
		entries := v.([]interface{})
		nat.StaticEntries = expandStaticEntries(entries)
	}

	return nat
}

// expandStaticEntries converts Terraform schema data to MasqueradeStaticEntry slice
func expandStaticEntries(entries []interface{}) []client.MasqueradeStaticEntry {
	result := make([]client.MasqueradeStaticEntry, 0, len(entries))

	for _, entry := range entries {
		e := entry.(map[string]interface{})
		staticEntry := client.MasqueradeStaticEntry{
			EntryNumber:   e["entry_number"].(int),
			InsideLocal:   e["inside_local"].(string),
			OutsideGlobal: e["outside_global"].(string),
		}

		if v, ok := e["inside_local_port"].(int); ok && v > 0 {
			staticEntry.InsideLocalPort = &v
		}

		if v, ok := e["outside_global_port"].(int); ok && v > 0 {
			staticEntry.OutsideGlobalPort = &v
		}

		if protocol, ok := e["protocol"].(string); ok {
			staticEntry.Protocol = protocol
		}

		result = append(result, staticEntry)
	}

	return result
}

// flattenStaticEntries converts MasqueradeStaticEntry slice to Terraform schema format
func flattenStaticEntries(entries []client.MasqueradeStaticEntry) []interface{} {
	result := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		e := map[string]interface{}{
			"entry_number":   entry.EntryNumber,
			"inside_local":   entry.InsideLocal,
			"outside_global": entry.OutsideGlobal,
			"protocol":       entry.Protocol,
		}

		// Handle optional port fields (nil for protocol-only entries like ESP/AH/GRE)
		if entry.InsideLocalPort != nil {
			e["inside_local_port"] = *entry.InsideLocalPort
		}
		if entry.OutsideGlobalPort != nil {
			e["outside_global_port"] = *entry.OutsideGlobalPort
		}

		result = append(result, e)
	}

	return result
}

// parseNATMasqueradeID parses the resource ID (descriptor_id) from string
func parseNATMasqueradeID(id string) (int, error) {
	descriptorID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid descriptor_id %q: %v", id, err)
	}

	if descriptorID < 1 || descriptorID > 65535 {
		return 0, fmt.Errorf("descriptor_id must be between 1 and 65535, got %d", descriptorID)
	}

	return descriptorID, nil
}

// validateIPRange validates an IP range in format "start_ip-end_ip"
func validateIPRange(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return nil, []error{fmt.Errorf("%q must be in format 'start_ip-end_ip' (e.g., '192.168.1.0-192.168.1.255'), got %q", k, value)}
	}

	// Validate start IP
	startIP := strings.TrimSpace(parts[0])
	if _, errs := validateIPAddress(startIP, k); len(errs) > 0 {
		return nil, []error{fmt.Errorf("%q has invalid start IP address: %v", k, errs[0])}
	}

	// Validate end IP
	endIP := strings.TrimSpace(parts[1])
	if _, errs := validateIPAddress(endIP, k); len(errs) > 0 {
		return nil, []error{fmt.Errorf("%q has invalid end IP address: %v", k, errs[0])}
	}

	return nil, nil
}
