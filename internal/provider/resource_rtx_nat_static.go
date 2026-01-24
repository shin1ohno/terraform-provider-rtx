package provider

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
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

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_nat_static", d.Id())
	natStatic := buildNATStaticFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_static").Msgf("Creating NAT static: %+v", natStatic)

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

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_nat_static", d.Id())
	logger := logging.FromContext(ctx)

	descriptorID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logger.Debug().Str("resource", "rtx_nat_static").Msgf("Reading NAT static: %d", descriptorID)

	var natStatic *client.NATStatic

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract NAT static configs from parsed config
			nats := parsedConfig.ExtractNATStatic()
			for i := range nats {
				if nats[i].DescriptorID == descriptorID {
					natStatic = convertParsedNATStatic(&nats[i])
					logger.Debug().Str("resource", "rtx_nat_static").Msg("Found NAT static in SFTP cache")
					break
				}
			}
		}
		if natStatic == nil {
			// NAT static not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_nat_static").Msg("NAT static not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or NAT static not found in cache
	if natStatic == nil {
		natStatic, err = apiClient.client.GetNATStatic(ctx, descriptorID)
		if err != nil {
			// Check if NAT static doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_nat_static").Msgf("NAT static %d not found, removing from state", descriptorID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read NAT static: %v", err)
		}
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

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_nat_static", d.Id())
	natStatic := buildNATStaticFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_static").Msgf("Updating NAT static: %+v", natStatic)

	err := apiClient.client.UpdateNATStatic(ctx, natStatic)
	if err != nil {
		return diag.Errorf("Failed to update NAT static: %v", err)
	}

	return resourceRTXNATStaticRead(ctx, d, meta)
}

func resourceRTXNATStaticDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_nat_static", d.Id())
	descriptorID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_static").Msgf("Deleting NAT static: %d", descriptorID)

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

	logging.FromContext(ctx).Debug().Str("resource", "rtx_nat_static").Msgf("Importing NAT static: %d", descriptorID)

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
			natEntry.InsideLocalPort = &v
		}

		if v, ok := entry["outside_global_port"].(int); ok && v > 0 {
			natEntry.OutsideGlobalPort = &v
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
			"inside_local":   entry.InsideLocal,
			"outside_global": entry.OutsideGlobal,
			"protocol":       entry.Protocol,
		}
		if entry.InsideLocalPort != nil {
			e["inside_local_port"] = *entry.InsideLocalPort
		} else {
			e["inside_local_port"] = 0
		}
		if entry.OutsideGlobalPort != nil {
			e["outside_global_port"] = *entry.OutsideGlobalPort
		} else {
			e["outside_global_port"] = 0
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

// convertParsedNATStatic converts a parser NATStatic to a client NATStatic
func convertParsedNATStatic(parsed *parsers.NATStatic) *client.NATStatic {
	nat := &client.NATStatic{
		DescriptorID: parsed.DescriptorID,
		Entries:      make([]client.NATStaticEntry, len(parsed.Entries)),
	}
	for i, entry := range parsed.Entries {
		nat.Entries[i] = client.NATStaticEntry{
			InsideLocal:   entry.InsideLocal,
			OutsideGlobal: entry.OutsideGlobal,
			Protocol:      entry.Protocol,
		}
		if entry.InsideLocalPort > 0 {
			port := entry.InsideLocalPort
			nat.Entries[i].InsideLocalPort = &port
		}
		if entry.OutsideGlobalPort > 0 {
			port := entry.OutsideGlobalPort
			nat.Entries[i].OutsideGlobalPort = &port
		}
	}
	return nat
}
