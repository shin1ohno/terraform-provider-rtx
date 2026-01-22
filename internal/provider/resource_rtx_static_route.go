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
)

func resourceRTXStaticRoute() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages static routes on RTX routers. A static route defines a fixed path for network traffic to reach a specific destination network.",
		CreateContext: resourceRTXStaticRouteCreate,
		ReadContext:   resourceRTXStaticRouteRead,
		UpdateContext: resourceRTXStaticRouteUpdate,
		DeleteContext: resourceRTXStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXStaticRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"prefix": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The destination network prefix (e.g., '10.0.0.0' for a network, '0.0.0.0' for default route)",
				ValidateFunc: validateRoutePrefix,
			},
			"mask": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The subnet mask in dotted decimal notation (e.g., '255.255.255.0')",
				ValidateFunc: validateRouteMask,
			},
			"next_hop": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Next hop configuration for this route. Multiple next hops enable load balancing or failover.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Next hop gateway IP address (e.g., '192.168.1.1'). Either gateway or interface must be specified.",
							ValidateFunc: validateOptionalIPAddress,
						},
						"interface": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Outgoing interface (e.g., 'pp 1', 'tunnel 1'). Either gateway or interface must be specified.",
						},
						"distance": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Administrative distance (weight). Lower values are preferred. Range: 1-100. Defaults to router default if not specified.",
							ValidateFunc: validation.IntBetween(1, 100),
						},
						"permanent": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Keep route even when next hop is unreachable (keepalive). Defaults to router default if not specified.",
						},
						"filter": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "IP filter number to apply to this route. 0 means no filter. Defaults to router default if not specified.",
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
		},
	}
}

func resourceRTXStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	route := buildStaticRouteFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_static_route").Msgf("Creating static route: %+v", route)

	err := apiClient.client.CreateStaticRoute(ctx, route)
	if err != nil {
		return diag.Errorf("Failed to create static route: %v", err)
	}

	// Use prefix/mask as the resource ID
	d.SetId(fmt.Sprintf("%s/%s", route.Prefix, route.Mask))

	// Read back to ensure consistency
	return resourceRTXStaticRouteRead(ctx, d, meta)
}

func resourceRTXStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	prefix, mask, err := parseStaticRouteID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_static_route").Msgf("Reading static route: %s/%s", prefix, mask)

	route, err := apiClient.client.GetStaticRoute(ctx, prefix, mask)
	if err != nil {
		// Check if route doesn't exist
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_static_route").Msgf("Static route %s/%s not found, removing from state", prefix, mask)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read static route: %v", err)
	}

	// Update the state
	if err := d.Set("prefix", route.Prefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mask", route.Mask); err != nil {
		return diag.FromErr(err)
	}

	// Convert NextHops to list of maps
	nextHops := make([]map[string]interface{}, len(route.NextHops))
	for i, hop := range route.NextHops {
		nextHops[i] = map[string]interface{}{
			"gateway":   hop.NextHop,
			"interface": hop.Interface,
			"distance":  hop.Distance,
			"permanent": hop.Permanent,
			"filter":    hop.Filter,
		}
	}
	if err := d.Set("next_hop", nextHops); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	route := buildStaticRouteFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_static_route").Msgf("Updating static route: %+v", route)

	err := apiClient.client.UpdateStaticRoute(ctx, route)
	if err != nil {
		return diag.Errorf("Failed to update static route: %v", err)
	}

	return resourceRTXStaticRouteRead(ctx, d, meta)
}

func resourceRTXStaticRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	prefix, mask, err := parseStaticRouteID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_static_route").Msgf("Deleting static route: %s/%s", prefix, mask)

	err = apiClient.client.DeleteStaticRoute(ctx, prefix, mask)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete static route: %v", err)
	}

	return nil
}

func resourceRTXStaticRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as "prefix/mask" (e.g., "10.0.0.0/255.0.0.0" or "0.0.0.0/0.0.0.0")
	prefix, mask, err := parseStaticRouteID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'prefix/mask': %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_static_route").Msgf("Importing static route: %s/%s", prefix, mask)

	// Verify route exists
	route, err := apiClient.client.GetStaticRoute(ctx, prefix, mask)
	if err != nil {
		return nil, fmt.Errorf("failed to import static route %s/%s: %v", prefix, mask, err)
	}

	// Set all attributes
	d.SetId(fmt.Sprintf("%s/%s", prefix, mask))
	d.Set("prefix", route.Prefix)
	d.Set("mask", route.Mask)

	nextHops := make([]map[string]interface{}, len(route.NextHops))
	for i, hop := range route.NextHops {
		nextHops[i] = map[string]interface{}{
			"gateway":   hop.NextHop,
			"interface": hop.Interface,
			"distance":  hop.Distance,
			"permanent": hop.Permanent,
			"filter":    hop.Filter,
		}
	}
	d.Set("next_hop", nextHops)

	return []*schema.ResourceData{d}, nil
}

// buildStaticRouteFromResourceData creates a StaticRoute from Terraform resource data
func buildStaticRouteFromResourceData(d *schema.ResourceData) client.StaticRoute {
	route := client.StaticRoute{
		Prefix: d.Get("prefix").(string),
		Mask:   d.Get("mask").(string),
	}

	// Handle next_hop list
	if v, ok := d.GetOk("next_hop"); ok {
		nextHopsList := v.([]interface{})
		nextHops := make([]client.StaticRouteHop, len(nextHopsList))

		for i, h := range nextHopsList {
			hopMap := h.(map[string]interface{})
			nextHops[i] = client.StaticRouteHop{
				NextHop:   hopMap["gateway"].(string),
				Interface: hopMap["interface"].(string),
				Distance:  hopMap["distance"].(int),
				Permanent: hopMap["permanent"].(bool),
				Filter:    hopMap["filter"].(int),
			}
		}
		route.NextHops = nextHops
	}

	return route
}

// parseStaticRouteID parses the resource ID into prefix and mask
func parseStaticRouteID(id string) (prefix, mask string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format 'prefix/mask', got %q", id)
	}

	prefix = parts[0]
	mask = parts[1]

	// Validate both parts are valid IPs
	if net.ParseIP(prefix) == nil {
		return "", "", fmt.Errorf("invalid prefix IP address: %s", prefix)
	}
	if net.ParseIP(mask) == nil {
		return "", "", fmt.Errorf("invalid mask: %s", mask)
	}

	return prefix, mask, nil
}

// validateRoutePrefix validates a route prefix (destination network)
func validateRoutePrefix(v interface{}, k string) ([]string, []error) {
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
		return nil, []error{fmt.Errorf("%q must be an IPv4 address, got %q", k, value)}
	}

	return nil, nil
}

// validateRouteMask validates a subnet mask
func validateRouteMask(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, []error{fmt.Errorf("%q must be a valid subnet mask in dotted decimal notation (e.g., '255.255.255.0'), got %q", k, value)}
	}

	// Ensure it's IPv4
	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q must be an IPv4 subnet mask, got %q", k, value)}
	}

	// Validate that it's a valid mask (contiguous 1s followed by 0s)
	maskBytes := ip.To4()
	mask := net.IPv4Mask(maskBytes[0], maskBytes[1], maskBytes[2], maskBytes[3])
	ones, bits := mask.Size()
	if bits == 0 {
		return nil, []error{fmt.Errorf("%q is not a valid subnet mask: %s", k, value)}
	}

	// Verify the mask is contiguous
	maskInt := uint32(maskBytes[0])<<24 | uint32(maskBytes[1])<<16 | uint32(maskBytes[2])<<8 | uint32(maskBytes[3])
	expectedMask := uint32(0xFFFFFFFF) << (32 - ones)
	if maskInt != expectedMask {
		return nil, []error{fmt.Errorf("%q is not a valid contiguous subnet mask: %s", k, value)}
	}

	return nil, nil
}

// validateOptionalIPAddress validates an optional IP address field
func validateOptionalIPAddress(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil // Optional field, empty is OK
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IP address, got %q", k, value)}
	}

	// Ensure it's IPv4
	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q must be an IPv4 address, got %q", k, value)}
	}

	return nil, nil
}

// Helper function to convert CIDR prefix length to mask
func cidrPrefixToMask(prefixLen int) string {
	if prefixLen < 0 || prefixLen > 32 {
		return ""
	}

	mask := uint32(0xFFFFFFFF) << (32 - prefixLen)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}

// Helper function to convert mask to CIDR prefix length
func maskToCIDRPrefix(mask string) int {
	parts := strings.Split(mask, ".")
	if len(parts) != 4 {
		return -1
	}

	bits := 0
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return -1
		}
		for i := 7; i >= 0; i-- {
			if (num & (1 << i)) != 0 {
				bits++
			}
		}
	}
	return bits
}
