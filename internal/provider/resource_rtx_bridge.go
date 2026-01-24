package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXBridge() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages Ethernet bridge configurations on RTX routers. Bridges combine multiple interfaces into a single Layer 2 broadcast domain.",
		CreateContext: resourceRTXBridgeCreate,
		ReadContext:   resourceRTXBridgeRead,
		UpdateContext: resourceRTXBridgeUpdate,
		DeleteContext: resourceRTXBridgeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXBridgeImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The bridge name (e.g., 'bridge1', 'bridge2'). Must be in format 'bridgeN'.",
				ValidateFunc: validateBridgeName,
			},
			"members": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of member interfaces to include in the bridge (e.g., ['lan1', 'tunnel1']). Valid formats include 'lanN', 'lanN/N' (VLAN), 'tunnelN', 'ppN', 'loopbackN'.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateBridgeMember,
				},
			},
		},
	}
}

func resourceRTXBridgeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_bridge", d.Id())
	bridge := buildBridgeFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_bridge").Msgf("Creating bridge: %+v", bridge)

	err := apiClient.client.CreateBridge(ctx, bridge)
	if err != nil {
		return diag.Errorf("Failed to create bridge: %v", err)
	}

	// Set resource ID as bridge name
	d.SetId(bridge.Name)

	// Read back to ensure consistency
	return resourceRTXBridgeRead(ctx, d, meta)
}

func resourceRTXBridgeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_bridge", d.Id())
	logger := logging.FromContext(ctx)

	name := d.Id()

	logger.Debug().Str("resource", "rtx_bridge").Msgf("Reading bridge: %s", name)

	var bridge *client.BridgeConfig
	var err error

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, cacheErr := apiClient.client.GetCachedConfig(ctx)
		if cacheErr == nil && parsedConfig != nil {
			// Extract bridges from parsed config
			bridges := parsedConfig.ExtractBridges()
			for i := range bridges {
				if bridges[i].Name == name {
					bridge = convertParsedBridgeConfig(&bridges[i])
					logger.Debug().Str("resource", "rtx_bridge").Msg("Found bridge in SFTP cache")
					break
				}
			}
		}
		if bridge == nil {
			// Bridge not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_bridge").Msg("Bridge not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or bridge not found in cache
	if bridge == nil {
		bridge, err = apiClient.client.GetBridge(ctx, name)
		if err != nil {
			// Check if bridge doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_bridge").Msgf("Bridge %s not found, removing from state", name)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read bridge: %v", err)
		}
	}

	// Update the state
	if err := d.Set("name", bridge.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("members", bridge.Members); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXBridgeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_bridge", d.Id())
	bridge := buildBridgeFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_bridge").Msgf("Updating bridge: %+v", bridge)

	err := apiClient.client.UpdateBridge(ctx, bridge)
	if err != nil {
		return diag.Errorf("Failed to update bridge: %v", err)
	}

	return resourceRTXBridgeRead(ctx, d, meta)
}

func resourceRTXBridgeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_bridge", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_bridge").Msgf("Deleting bridge: %s", name)

	err := apiClient.client.DeleteBridge(ctx, name)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete bridge: %v", err)
	}

	return nil
}

func resourceRTXBridgeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Import ID should be the bridge name (e.g., "bridge1")
	if err := validateBridgeNameValue(importID); err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected bridge name (e.g., 'bridge1'): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_bridge").Msgf("Importing bridge: %s", importID)

	// Verify bridge exists
	bridge, err := apiClient.client.GetBridge(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("failed to import bridge %s: %v", importID, err)
	}

	// Set all attributes
	d.SetId(bridge.Name)
	d.Set("name", bridge.Name)
	d.Set("members", bridge.Members)

	return []*schema.ResourceData{d}, nil
}

// buildBridgeFromResourceData creates a BridgeConfig from Terraform resource data
func buildBridgeFromResourceData(d *schema.ResourceData) client.BridgeConfig {
	bridge := client.BridgeConfig{
		Name:    d.Get("name").(string),
		Members: []string{},
	}

	if v, ok := d.GetOk("members"); ok {
		memberList := v.([]interface{})
		for _, m := range memberList {
			bridge.Members = append(bridge.Members, m.(string))
		}
	}

	return bridge
}

// validateBridgeName validates the bridge name format for Terraform schema
func validateBridgeName(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if err := validateBridgeNameValue(value); err != nil {
		return nil, []error{fmt.Errorf("%q: %v", k, err)}
	}

	return nil, nil
}

// validateBridgeNameValue validates the bridge name format
func validateBridgeNameValue(value string) error {
	if value == "" {
		return fmt.Errorf("bridge name cannot be empty")
	}

	// Bridge name must be in format "bridgeN" (e.g., bridge1, bridge2)
	validNamePattern := regexp.MustCompile(`^bridge\d+$`)
	if !validNamePattern.MatchString(value) {
		return fmt.Errorf("bridge name must be in format 'bridgeN' (e.g., 'bridge1', 'bridge2'), got %q", value)
	}

	return nil
}

// validateBridgeMember validates a bridge member interface name for Terraform schema
func validateBridgeMember(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	// Valid member patterns
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^lan\d+$`),      // lan1, lan2, etc.
		regexp.MustCompile(`^lan\d+/\d+$`),  // lan1/1 (VLAN interfaces)
		regexp.MustCompile(`^tunnel\d+$`),   // tunnel1, tunnel2, etc.
		regexp.MustCompile(`^pp\d+$`),       // pp1, pp2, etc.
		regexp.MustCompile(`^loopback\d+$`), // loopback1, etc.
		regexp.MustCompile(`^bridge\d+$`),   // nested bridge (rare)
	}

	for _, pattern := range validPatterns {
		if pattern.MatchString(value) {
			return nil, nil
		}
	}

	return nil, []error{
		fmt.Errorf("%q must be a valid interface name (lan*, lan*/*, tunnel*, pp*, loopback*, bridge*), got %q", k, value),
	}
}

// convertParsedBridgeConfig converts a parser BridgeConfig to a client BridgeConfig
func convertParsedBridgeConfig(parsed *parsers.BridgeConfig) *client.BridgeConfig {
	return &client.BridgeConfig{
		Name:    parsed.Name,
		Members: parsed.Members,
	}
}
