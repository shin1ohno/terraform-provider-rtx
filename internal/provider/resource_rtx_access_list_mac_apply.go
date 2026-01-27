package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// resourceRTXAccessListMACApply returns the schema for the rtx_access_list_mac_apply resource
func resourceRTXAccessListMACApply() *schema.Resource {
	return &schema.Resource{
		Description: "Applies MAC access list filters to an interface. " +
			"This resource manages the binding of MAC ACL filter sequences to a specific interface and direction. " +
			"Use this resource when you need to apply filters from multiple ACLs to the same interface, " +
			"or when you want to manage interface filter bindings separately from the ACL definitions. " +
			"Note: MAC filters are only supported on Ethernet interfaces (lan, bridge). " +
			"PP and Tunnel interfaces are not supported.",

		CreateContext: resourceRTXAccessListMACApplyCreate,
		ReadContext:   resourceRTXAccessListMACApplyRead,
		UpdateContext: resourceRTXAccessListMACApplyUpdate,
		DeleteContext: resourceRTXAccessListMACApplyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListMACApplyImport,
		},

		CustomizeDiff: resourceRTXAccessListMACApplyCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"access_list": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the MAC access list to apply. This is used for tracking purposes.",
			},
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Interface name to apply the filters to (e.g., lan1, bridge1). PP and Tunnel interfaces are not supported for MAC filters.",
			},
			"direction": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Traffic direction: 'in' for incoming traffic, 'out' for outgoing traffic",
				ValidateFunc:     validation.StringInSlice([]string{"in", "out"}, true),
				DiffSuppressFunc: SuppressCaseDiff,
			},
			"filter_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of filter IDs to apply in order. If not specified, filters must be applied via the associated access_list resource.",
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},
		},
	}
}

// resourceRTXAccessListMACApplyCustomizeDiff checks for conflicts and validates interface type
func resourceRTXAccessListMACApplyCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Get interface and direction
	iface := d.Get("interface").(string)
	direction := strings.ToLower(d.Get("direction").(string))

	// Log for debugging
	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Running CustomizeDiff for MAC apply resource")

	// Validate interface type - MAC filters are NOT supported on PP and Tunnel interfaces
	if err := validateMACInterfaceType(iface); err != nil {
		return err
	}

	// TODO: Add conflict detection with inline applies when other ACL resources
	// are updated to support inline apply blocks. This would require checking
	// Terraform state for rtx_access_list_mac resources that have apply blocks
	// targeting the same interface:direction combination.

	return nil
}

// validateMACInterfaceType validates that the interface type supports MAC filters
// MAC filters are only supported on Ethernet interfaces (lan, bridge)
// PP and Tunnel interfaces are NOT supported
func validateMACInterfaceType(iface string) error {
	ifaceLower := strings.ToLower(iface)

	// Check for unsupported interface types
	if strings.HasPrefix(ifaceLower, "pp") {
		// Ensure it's actually a PP interface (pp followed by a number)
		rest := ifaceLower[2:]
		if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			return fmt.Errorf("MAC filters are not supported on PP interfaces: %s. Use lan or bridge interfaces instead", iface)
		}
	}

	if strings.HasPrefix(ifaceLower, "tunnel") {
		// Ensure it's actually a tunnel interface (tunnel followed by a number)
		rest := ifaceLower[6:]
		if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			return fmt.Errorf("MAC filters are not supported on Tunnel interfaces: %s. Use lan or bridge interfaces instead", iface)
		}
	}

	// Validate it's a supported interface type (lan or bridge)
	if !strings.HasPrefix(ifaceLower, "lan") && !strings.HasPrefix(ifaceLower, "bridge") {
		return fmt.Errorf("MAC filters are only supported on Ethernet interfaces (lan, bridge), not %s", iface)
	}

	return nil
}

func resourceRTXAccessListMACApplyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface := d.Get("interface").(string)
	direction := strings.ToLower(d.Get("direction").(string))
	aclName := d.Get("access_list").(string)

	// Build resource ID
	id := fmt.Sprintf("%s:%s", iface, direction)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Creating MAC access list apply")

	// Validate interface type before proceeding
	if err := validateMACInterfaceType(iface); err != nil {
		return diag.FromErr(err)
	}

	// Get filter IDs
	filterIDs := extractApplyFilterIDs(d)

	if len(filterIDs) == 0 {
		return diag.Errorf("filter_ids is required: at least one filter ID must be specified")
	}

	// Apply filters to interface
	err := apiClient.client.ApplyMACFiltersToInterface(ctx, iface, direction, filterIDs)
	if err != nil {
		return diag.Errorf("Failed to apply MAC filters to interface %s %s: %v", iface, direction, err)
	}

	d.SetId(id)

	return resourceRTXAccessListMACApplyRead(ctx, d, meta)
}

func resourceRTXAccessListMACApplyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse ID
	id := d.Id()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return diag.Errorf("Invalid resource ID format: %s, expected 'interface:direction'", id)
	}

	iface := parts[0]
	direction := parts[1]

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Reading MAC access list apply")

	// Get current filter IDs from router
	filterIDs, err := apiClient.client.GetMACInterfaceFilters(ctx, iface, direction)
	if err != nil {
		// If not found, the resource has been removed
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().
				Str("resource", "rtx_access_list_mac_apply").
				Str("interface", iface).
				Str("direction", direction).
				Msg("MAC filter apply not found, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read MAC filter apply for %s %s: %v", iface, direction, err)
	}

	// If no filters are applied, resource doesn't exist
	if len(filterIDs) == 0 {
		logger.Warn().
			Str("resource", "rtx_access_list_mac_apply").
			Str("interface", iface).
			Str("direction", direction).
			Msg("No MAC filters applied, removing from state")
		d.SetId("")
		return nil
	}

	// Update state
	d.Set("interface", iface)
	d.Set("direction", direction)
	d.Set("filter_ids", filterIDs)

	return nil
}

func resourceRTXAccessListMACApplyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse ID
	id := d.Id()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return diag.Errorf("Invalid resource ID format: %s, expected 'interface:direction'", id)
	}

	iface := parts[0]
	direction := parts[1]
	aclName := d.Get("access_list").(string)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Updating MAC access list apply")

	// Get filter IDs
	filterIDs := extractApplyFilterIDs(d)

	if len(filterIDs) == 0 {
		return diag.Errorf("filter_ids is required: at least one filter ID must be specified")
	}

	// Apply filters to interface (this will replace existing filters)
	err := apiClient.client.ApplyMACFiltersToInterface(ctx, iface, direction, filterIDs)
	if err != nil {
		return diag.Errorf("Failed to update MAC filters on interface %s %s: %v", iface, direction, err)
	}

	return resourceRTXAccessListMACApplyRead(ctx, d, meta)
}

func resourceRTXAccessListMACApplyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse ID
	id := d.Id()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return diag.Errorf("Invalid resource ID format: %s, expected 'interface:direction'", id)
	}

	iface := parts[0]
	direction := parts[1]

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_mac_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Deleting MAC access list apply")

	// Remove filters from interface
	err := apiClient.client.RemoveMACFiltersFromInterface(ctx, iface, direction)
	if err != nil {
		// Ignore "not found" errors
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to remove MAC filters from interface %s %s: %v", iface, direction, err)
	}

	return nil
}

func resourceRTXAccessListMACApplyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// ID format: interface:direction
	id := d.Id()
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format: %s, expected 'interface:direction' (e.g., 'lan1:in')", id)
	}

	iface := parts[0]
	direction := strings.ToLower(parts[1])

	// Validate direction
	if direction != "in" && direction != "out" {
		return nil, fmt.Errorf("invalid direction: %s, must be 'in' or 'out'", direction)
	}

	// Validate interface type
	if err := validateMACInterfaceType(iface); err != nil {
		return nil, err
	}

	// Set the ID and basic attributes
	d.SetId(fmt.Sprintf("%s:%s", iface, direction))
	d.Set("interface", iface)
	d.Set("direction", direction)
	// access_list will need to be set manually after import
	d.Set("access_list", "imported")

	return []*schema.ResourceData{d}, nil
}
