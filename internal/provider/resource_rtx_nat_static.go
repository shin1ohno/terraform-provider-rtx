package provider

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXNATStatic() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages static NAT (Network Address Translation) on RTX routers. Static NAT provides one-to-one mapping between inside local and outside global addresses.",
		CreateContext: resourceRTXNATStaticCreate,
		ReadContext:   resourceRTXNATStaticRead,
		UpdateContext: resourceRTXNATStaticUpdate,
		DeleteContext: resourceRTXNATStaticDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXNATStaticImport,
		},

		Schema: map[string]*schema.Schema{
			"descriptor_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "The NAT descriptor ID (1-65535)",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"entry": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of static NAT mapping entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inside_local": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Inside local IP address (internal address)",
							ValidateFunc: validateNATIPAddress,
						},
						"inside_local_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Inside local port (1-65535, required if protocol is specified)",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"outside_global": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Outside global IP address (external address)",
							ValidateFunc: validateNATIPAddress,
						},
						"outside_global_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Outside global port (1-65535, required if protocol is specified)",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Protocol for port-based NAT: 'tcp' or 'udp' (required if ports are specified)",
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp"}, false),
						},
					},
				},
			},
		},

		CustomizeDiff: validateNATStaticEntries,
	}
}

// validateNATStaticEntries validates that entries have consistent port/protocol configuration
func validateNATStaticEntries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	entries := diff.Get("entry").([]interface{})

	for i, e := range entries {
		entry := e.(map[string]interface{})

		insideLocalPort := entry["inside_local_port"].(int)
		outsideGlobalPort := entry["outside_global_port"].(int)
		protocol := entry["protocol"].(string)

		// If any port is specified, protocol must be specified
		if (insideLocalPort > 0 || outsideGlobalPort > 0) && protocol == "" {
			return fmt.Errorf("entry[%d]: protocol is required when ports are specified", i)
		}

		// If protocol is specified, both ports should be specified
		if protocol != "" && (insideLocalPort == 0 || outsideGlobalPort == 0) {
			return fmt.Errorf("entry[%d]: both inside_local_port and outside_global_port are required when protocol is specified", i)
		}
	}

	return nil
}

func resourceRTXNATStaticCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	natStatic := buildNATStaticFromResourceData(d)

	log.Printf("[DEBUG] Creating NAT static: %+v", natStatic)

	err := apiClient.client.CreateNATStatic(ctx, natStatic)
	if err != nil {
		return diag.Errorf("Failed to create NAT static: %v", err)
	}

	// Set resource ID as the descriptor_id
	d.SetId(strconv.Itoa(natStatic.DescriptorID))

	// Read back to ensure consistency
	return resourceRTXNATStaticRead(ctx, d, meta)
}

func resourceRTXNATStaticRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	descriptorID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	log.Printf("[DEBUG] Reading NAT static: %d", descriptorID)

	natStatic, err := apiClient.client.GetNATStatic(ctx, descriptorID)
	if err != nil {
		// Check if NAT static doesn't exist
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] NAT static %d not found, removing from state", descriptorID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read NAT static: %v", err)
	}

	// Update the state
	if err := d.Set("descriptor_id", natStatic.DescriptorID); err != nil {
		return diag.FromErr(err)
	}

	entries := flattenNATStaticEntries(natStatic.Entries)
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXNATStaticUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	natStatic := buildNATStaticFromResourceData(d)

	log.Printf("[DEBUG] Updating NAT static: %+v", natStatic)

	err := apiClient.client.UpdateNATStatic(ctx, natStatic)
	if err != nil {
		return diag.Errorf("Failed to update NAT static: %v", err)
	}

	return resourceRTXNATStaticRead(ctx, d, meta)
}

func resourceRTXNATStaticDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	descriptorID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	log.Printf("[DEBUG] Deleting NAT static: %d", descriptorID)

	err = apiClient.client.DeleteNATStatic(ctx, descriptorID)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete NAT static: %v", err)
	}

	return nil
}

func resourceRTXNATStaticImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as descriptor_id
	descriptorID, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected descriptor_id (integer), got %q: %v", importID, err)
	}

	if descriptorID < 1 || descriptorID > 65535 {
		return nil, fmt.Errorf("descriptor_id must be between 1 and 65535, got %d", descriptorID)
	}

	log.Printf("[DEBUG] Importing NAT static: %d", descriptorID)

	// Verify NAT static exists
	natStatic, err := apiClient.client.GetNATStatic(ctx, descriptorID)
	if err != nil {
		return nil, fmt.Errorf("failed to import NAT static %d: %v", descriptorID, err)
	}

	// Set all attributes
	d.SetId(strconv.Itoa(descriptorID))
	d.Set("descriptor_id", natStatic.DescriptorID)

	entries := flattenNATStaticEntries(natStatic.Entries)
	d.Set("entry", entries)

	return []*schema.ResourceData{d}, nil
}

// buildNATStaticFromResourceData creates a NATStatic from Terraform resource data
func buildNATStaticFromResourceData(d *schema.ResourceData) client.NATStatic {
	natStatic := client.NATStatic{
		DescriptorID: d.Get("descriptor_id").(int),
		Entries:      expandNATStaticEntries(d.Get("entry").([]interface{})),
	}

	return natStatic
}

// expandNATStaticEntries converts Terraform list to []NATStaticEntry
func expandNATStaticEntries(entries []interface{}) []client.NATStaticEntry {
	result := make([]client.NATStaticEntry, 0, len(entries))

	for _, e := range entries {
		entry := e.(map[string]interface{})

		natEntry := client.NATStaticEntry{
			InsideLocal:   entry["inside_local"].(string),
			OutsideGlobal: entry["outside_global"].(string),
		}

		if v, ok := entry["inside_local_port"].(int); ok && v > 0 {
			natEntry.InsideLocalPort = v
		}

		if v, ok := entry["outside_global_port"].(int); ok && v > 0 {
			natEntry.OutsideGlobalPort = v
		}

		if v, ok := entry["protocol"].(string); ok && v != "" {
			natEntry.Protocol = v
		}

		result = append(result, natEntry)
	}

	return result
}

// flattenNATStaticEntries converts []NATStaticEntry to Terraform list
func flattenNATStaticEntries(entries []client.NATStaticEntry) []interface{} {
	result := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		e := map[string]interface{}{
			"inside_local":        entry.InsideLocal,
			"outside_global":      entry.OutsideGlobal,
			"inside_local_port":   entry.InsideLocalPort,
			"outside_global_port": entry.OutsideGlobalPort,
			"protocol":            entry.Protocol,
		}
		result = append(result, e)
	}

	return result
}

// validateNATIPAddress validates that the value is a valid IPv4 address
func validateNATIPAddress(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IP address, got %q", k, value)}
	}

	// Ensure it's IPv4
	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IPv4 address, got %q", k, value)}
	}

	return nil, nil
}
