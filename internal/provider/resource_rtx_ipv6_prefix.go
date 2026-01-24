package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXIPv6Prefix() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPv6 prefix definitions on RTX routers. IPv6 prefixes can be configured as static, RA-derived, or DHCPv6-PD delegated.",
		CreateContext: resourceRTXIPv6PrefixCreate,
		ReadContext:   resourceRTXIPv6PrefixRead,
		UpdateContext: resourceRTXIPv6PrefixUpdate,
		DeleteContext: resourceRTXIPv6PrefixDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXIPv6PrefixImport,
		},

		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "The IPv6 prefix ID (1-255)",
				ValidateFunc: validation.IntBetween(1, 255),
			},
			"prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Static IPv6 prefix value (e.g., '2001:db8::') - required when source is 'static'",
				ValidateFunc: validateIPv6Prefix,
			},
			"prefix_length": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Prefix length in bits (1-128)",
				ValidateFunc: validation.IntBetween(1, 128),
			},
			"source": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Prefix source type: 'static', 'ra' (Router Advertisement), or 'dhcpv6-pd' (Prefix Delegation)",
				ValidateFunc: validation.StringInSlice([]string{"static", "ra", "dhcpv6-pd"}, false),
			},
			"interface": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Source interface name (required for 'ra' and 'dhcpv6-pd' sources, e.g., 'lan2', 'pp1')",
			},
		},

		CustomizeDiff: validateIPv6PrefixConfig,
	}
}

// validateIPv6Prefix validates that a string is a valid IPv6 prefix
func validateIPv6Prefix(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil // Allow empty for non-static sources
	}

	// Basic validation - should contain only valid IPv6 characters
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == ':') {
			return nil, []error{fmt.Errorf("%q contains invalid characters for an IPv6 prefix", k)}
		}
	}

	// Should contain at least one colon
	if !strings.Contains(value, ":") {
		return nil, []error{fmt.Errorf("%q must be a valid IPv6 prefix (e.g., '2001:db8::')", k)}
	}

	return nil, nil
}

// validateIPv6PrefixConfig performs custom validation on the resource configuration
func validateIPv6PrefixConfig(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	source := d.Get("source").(string)
	prefix := d.Get("prefix").(string)
	iface := d.Get("interface").(string)

	switch source {
	case "static":
		if prefix == "" {
			return fmt.Errorf("'prefix' is required when source is 'static'")
		}
		if iface != "" {
			return fmt.Errorf("'interface' should not be set when source is 'static'")
		}
	case "ra", "dhcpv6-pd":
		if iface == "" {
			return fmt.Errorf("'interface' is required when source is '%s'", source)
		}
		if prefix != "" {
			return fmt.Errorf("'prefix' should not be set when source is '%s' (it is derived dynamically)", source)
		}
	}

	return nil
}

func resourceRTXIPv6PrefixCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", d.Id())
	prefix := buildIPv6PrefixFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Creating IPv6 prefix: %+v", prefix)

	err := apiClient.client.CreateIPv6Prefix(ctx, prefix)
	if err != nil {
		return diag.Errorf("Failed to create IPv6 prefix: %v", err)
	}

	// Use prefix_id as the resource ID
	d.SetId(strconv.Itoa(prefix.ID))

	// Read back to ensure consistency
	return resourceRTXIPv6PrefixRead(ctx, d, meta)
}

func resourceRTXIPv6PrefixRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", d.Id())
	logger := logging.FromContext(ctx)

	prefixID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Reading IPv6 prefix: %d", prefixID)

	var prefix *client.IPv6Prefix

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, cacheErr := apiClient.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract IPv6 prefixes from parsed config
			prefixes := parsedConfig.ExtractIPv6Prefixes()
			for i := range prefixes {
				if prefixes[i].ID == prefixID {
					prefix = convertParsedIPv6Prefix(&prefixes[i])
					logger.Debug().Str("resource", "rtx_ipv6_prefix").Msg("Found IPv6 prefix in SFTP cache")
					break
				}
			}
		}
		if prefix == nil {
			// Prefix not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_ipv6_prefix").Msg("IPv6 prefix not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or prefix not found in cache
	if prefix == nil {
		prefix, err = apiClient.client.GetIPv6Prefix(ctx, prefixID)
		if err != nil {
			// Check if prefix doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_ipv6_prefix").Msgf("IPv6 prefix %d not found, removing from state", prefixID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read IPv6 prefix: %v", err)
		}
	}

	// Update the state
	if err := d.Set("prefix_id", prefix.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prefix", prefix.Prefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prefix_length", prefix.PrefixLength); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("source", prefix.Source); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interface", prefix.Interface); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPv6PrefixUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", d.Id())
	prefix := buildIPv6PrefixFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Updating IPv6 prefix: %+v", prefix)

	err := apiClient.client.UpdateIPv6Prefix(ctx, prefix)
	if err != nil {
		return diag.Errorf("Failed to update IPv6 prefix: %v", err)
	}

	return resourceRTXIPv6PrefixRead(ctx, d, meta)
}

func resourceRTXIPv6PrefixDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_ipv6_prefix", d.Id())
	prefixID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Deleting IPv6 prefix: %d", prefixID)

	err = apiClient.client.DeleteIPv6Prefix(ctx, prefixID)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IPv6 prefix: %v", err)
	}

	return nil
}

func resourceRTXIPv6PrefixImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as prefix_id
	prefixID, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected prefix_id (integer): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_ipv6_prefix").Msgf("Importing IPv6 prefix: %d", prefixID)

	// Verify prefix exists
	prefix, err := apiClient.client.GetIPv6Prefix(ctx, prefixID)
	if err != nil {
		return nil, fmt.Errorf("failed to import IPv6 prefix %d: %v", prefixID, err)
	}

	// Set all attributes
	d.SetId(strconv.Itoa(prefixID))
	d.Set("prefix_id", prefix.ID)
	d.Set("prefix", prefix.Prefix)
	d.Set("prefix_length", prefix.PrefixLength)
	d.Set("source", prefix.Source)
	d.Set("interface", prefix.Interface)

	return []*schema.ResourceData{d}, nil
}

// buildIPv6PrefixFromResourceData creates an IPv6Prefix from Terraform resource data
func buildIPv6PrefixFromResourceData(d *schema.ResourceData) client.IPv6Prefix {
	return client.IPv6Prefix{
		ID:           d.Get("prefix_id").(int),
		Prefix:       d.Get("prefix").(string),
		PrefixLength: d.Get("prefix_length").(int),
		Source:       d.Get("source").(string),
		Interface:    d.Get("interface").(string),
	}
}

// convertParsedIPv6Prefix converts a parser IPv6Prefix to a client IPv6Prefix
func convertParsedIPv6Prefix(parsed *parsers.IPv6Prefix) *client.IPv6Prefix {
	return &client.IPv6Prefix{
		ID:           parsed.ID,
		Prefix:       parsed.Prefix,
		PrefixLength: parsed.PrefixLength,
		Source:       parsed.Source,
		Interface:    parsed.Interface,
	}
}
