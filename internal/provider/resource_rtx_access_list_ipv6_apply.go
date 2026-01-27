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

// resourceRTXAccessListIPv6Apply returns the schema for the rtx_access_list_ipv6_apply resource
func resourceRTXAccessListIPv6Apply() *schema.Resource {
	return &schema.Resource{
		Description: "Applies IPv6 access list filters to an interface. " +
			"This resource manages the binding of IPv6 ACL filter sequences to a specific interface and direction. " +
			"Use this resource when you need to apply filters from multiple ACLs to the same interface, " +
			"or when you want to manage interface filter bindings separately from the ACL definitions.",

		CreateContext: resourceRTXAccessListIPv6ApplyCreate,
		ReadContext:   resourceRTXAccessListIPv6ApplyRead,
		UpdateContext: resourceRTXAccessListIPv6ApplyUpdate,
		DeleteContext: resourceRTXAccessListIPv6ApplyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListIPv6ApplyImport,
		},

		CustomizeDiff: resourceRTXAccessListIPv6ApplyCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"access_list": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the IPv6 access list to apply. This is used for tracking purposes.",
			},
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Interface name to apply the filters to (e.g., lan1, pp1, tunnel1)",
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

// resourceRTXAccessListIPv6ApplyCustomizeDiff checks for conflicts with inline applies
func resourceRTXAccessListIPv6ApplyCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Get interface and direction
	iface := d.Get("interface").(string)
	direction := strings.ToLower(d.Get("direction").(string))

	// Log for debugging
	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Running CustomizeDiff for IPv6 apply resource")

	// TODO: Add conflict detection with inline applies when other ACL resources
	// are updated to support inline apply blocks. This would require checking
	// Terraform state for rtx_access_list_ipv6 resources that have apply blocks
	// targeting the same interface:direction combination.

	return nil
}

func resourceRTXAccessListIPv6ApplyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	iface := d.Get("interface").(string)
	direction := strings.ToLower(d.Get("direction").(string))
	aclName := d.Get("access_list").(string)

	// Build resource ID
	id := fmt.Sprintf("%s:%s", iface, direction)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Creating IPv6 access list apply")

	// Get filter IDs
	filterIDs := extractApplyFilterIDs(d)

	if len(filterIDs) == 0 {
		return diag.Errorf("filter_ids is required: at least one filter ID must be specified")
	}

	// Apply filters to interface
	err := apiClient.client.ApplyIPv6FiltersToInterface(ctx, iface, direction, filterIDs)
	if err != nil {
		return diag.Errorf("Failed to apply IPv6 filters to interface %s %s: %v", iface, direction, err)
	}

	d.SetId(id)

	return resourceRTXAccessListIPv6ApplyRead(ctx, d, meta)
}

func resourceRTXAccessListIPv6ApplyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Reading IPv6 access list apply")

	// Get current filter IDs from router
	filterIDs, err := apiClient.client.GetIPv6InterfaceFilters(ctx, iface, direction)
	if err != nil {
		// If not found, the resource has been removed
		if strings.Contains(err.Error(), "not found") {
			logger.Warn().
				Str("resource", "rtx_access_list_ipv6_apply").
				Str("interface", iface).
				Str("direction", direction).
				Msg("IPv6 filter apply not found, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IPv6 filter apply for %s %s: %v", iface, direction, err)
	}

	// If no filters are applied, resource doesn't exist
	if len(filterIDs) == 0 {
		logger.Warn().
			Str("resource", "rtx_access_list_ipv6_apply").
			Str("interface", iface).
			Str("direction", direction).
			Msg("No IPv6 filters applied, removing from state")
		d.SetId("")
		return nil
	}

	// Update state
	d.Set("interface", iface)
	d.Set("direction", direction)
	d.Set("filter_ids", filterIDs)

	return nil
}

func resourceRTXAccessListIPv6ApplyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Str("access_list", aclName).
		Msg("Updating IPv6 access list apply")

	// Get filter IDs
	filterIDs := extractApplyFilterIDs(d)

	if len(filterIDs) == 0 {
		return diag.Errorf("filter_ids is required: at least one filter ID must be specified")
	}

	// Apply filters to interface (this will replace existing filters)
	err := apiClient.client.ApplyIPv6FiltersToInterface(ctx, iface, direction, filterIDs)
	if err != nil {
		return diag.Errorf("Failed to update IPv6 filters on interface %s %s: %v", iface, direction, err)
	}

	return resourceRTXAccessListIPv6ApplyRead(ctx, d, meta)
}

func resourceRTXAccessListIPv6ApplyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6_apply", id)

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_access_list_ipv6_apply").
		Str("interface", iface).
		Str("direction", direction).
		Msg("Deleting IPv6 access list apply")

	// Remove filters from interface
	err := apiClient.client.RemoveIPv6FiltersFromInterface(ctx, iface, direction)
	if err != nil {
		// Ignore "not found" errors
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to remove IPv6 filters from interface %s %s: %v", iface, direction, err)
	}

	return nil
}

func resourceRTXAccessListIPv6ApplyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	// Set the ID and basic attributes
	d.SetId(fmt.Sprintf("%s:%s", iface, direction))
	d.Set("interface", iface)
	d.Set("direction", direction)
	// access_list will need to be set manually after import
	d.Set("access_list", "imported")

	return []*schema.ResourceData{d}, nil
}
