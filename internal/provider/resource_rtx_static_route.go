package provider

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXStaticRoute() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages static routes on RTX routers",
		CreateContext: resourceRTXStaticRouteCreate,
		ReadContext:   resourceRTXStaticRouteRead,
		UpdateContext: resourceRTXStaticRouteUpdate,
		DeleteContext: resourceRTXStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXStaticRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"destination": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "CIDR destination network (e.g., 192.168.0.0/24). Use 0.0.0.0/0 for default route.",
				StateFunc:    normalizeCIDR,
				ValidateFunc: validateCIDR,
			},
			"gateway_ip": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"gateway_ip", "gateway_interface"},
				Description:   "Next-hop IP address",
				ValidateFunc:  validation.IsIPAddress,
				StateFunc:     normalizeIPAddress,
			},
			"gateway_interface": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"gateway_ip", "gateway_interface"},
				Description:   "Next-hop interface name (wan1, lan1, pp1, etc.)",
				ValidateFunc:  validateInterfaceName,
			},
			"interface": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Outgoing interface used to send the packet",
				ValidateFunc: validateInterfaceName,
			},
			"metric": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "Route metric (1-65535). RTX default is 1",
				ValidateFunc:     validation.IntBetween(1, 65535),
				DiffSuppressFunc: suppressDefault1,
			},
			"weight": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "ECMP weight (1-255). Ignored unless multiple routes share the same destination",
				ValidateFunc:     validation.IntBetween(1, 255),
				DiffSuppressFunc: suppressDefault1,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Route description",
			},
			"hide": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, the route is hidden from 'show ip route'",
			},
		},

		// Custom validation
		CustomizeDiff: customdiff.All(
			// Ensure at least one gateway method is specified
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				return validateGatewaySpecification(ctx, d, meta)
			},
		),
	}
}

func resourceRTXStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	route := client.StaticRoute{
		Destination: d.Get("destination").(string),
		Metric:      getIntWithDefault(d, "metric", 1),
		Weight:      getIntWithDefault(d, "weight", 0),
		Description: d.Get("description").(string),
		Hide:        d.Get("hide").(bool),
	}

	// Set gateway method
	if v, ok := d.GetOk("gateway_ip"); ok {
		route.GatewayIP = v.(string)
	}
	if v, ok := d.GetOk("gateway_interface"); ok {
		route.GatewayInterface = v.(string)
	}
	if v, ok := d.GetOk("interface"); ok {
		route.Interface = v.(string)
	}

	err := apiClient.client.CreateStaticRoute(ctx, route)
	if err != nil {
		return diag.Errorf("Failed to create static route: %v", err)
	}

	// Set the composite ID: destination||gateway||interface
	id := buildStaticRouteID(route)
	d.SetId(id)

	// Read back to ensure consistency
	return resourceRTXStaticRouteRead(ctx, d, meta)
}

func resourceRTXStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the composite ID
	destination, gateway, iface, err := parseStaticRouteID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	// Get the specific route
	found, err := apiClient.client.GetStaticRoute(ctx, destination, gateway, iface)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			// Resource no longer exists, remove from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to retrieve static route: %v", err)
	}

	// Update the state
	if err := d.Set("destination", found.Destination); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gateway_ip", found.GatewayIP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gateway_interface", found.GatewayInterface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interface", found.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metric", found.Metric); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("weight", found.Weight); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", found.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hide", found.Hide); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the composite ID
	destination, gateway, iface, err := parseStaticRouteID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	route := client.StaticRoute{
		Destination:      destination,
		GatewayIP:        gateway,
		GatewayInterface: iface,
		Interface:        iface,
		Metric:           getIntWithDefault(d, "metric", 1),
		Weight:           getIntWithDefault(d, "weight", 0),
		Description:      d.Get("description").(string),
		Hide:             d.Get("hide").(bool),
	}

	// Update the route
	err = apiClient.client.UpdateStaticRoute(ctx, route)
	if err != nil {
		return diag.Errorf("Failed to update static route: %v", err)
	}

	// Read back to ensure consistency
	return resourceRTXStaticRouteRead(ctx, d, meta)
}

func resourceRTXStaticRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the composite ID
	destination, gateway, iface, err := parseStaticRouteID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	err = apiClient.client.DeleteStaticRoute(ctx, destination, gateway, iface)
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

	// Parse the import ID: destination||gateway||interface
	destination, gateway, iface, err := parseStaticRouteID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'destination||gateway||interface': %v", err)
	}

	// Get the specific route
	found, err := apiClient.client.GetStaticRoute(ctx, destination, gateway, iface)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve static route: %v", err)
	}

	// Set the parsed values
	if err := d.Set("destination", found.Destination); err != nil {
		return nil, fmt.Errorf("failed to set destination: %w", err)
	}
	if err := d.Set("gateway_ip", found.GatewayIP); err != nil {
		return nil, fmt.Errorf("failed to set gateway_ip: %w", err)
	}
	if err := d.Set("gateway_interface", found.GatewayInterface); err != nil {
		return nil, fmt.Errorf("failed to set gateway_interface: %w", err)
	}
	if err := d.Set("interface", found.Interface); err != nil {
		return nil, fmt.Errorf("failed to set interface: %w", err)
	}
	if err := d.Set("metric", found.Metric); err != nil {
		return nil, fmt.Errorf("failed to set metric: %w", err)
	}
	if err := d.Set("weight", found.Weight); err != nil {
		return nil, fmt.Errorf("failed to set weight: %w", err)
	}
	if err := d.Set("description", found.Description); err != nil {
		return nil, fmt.Errorf("failed to set description: %w", err)
	}
	if err := d.Set("hide", found.Hide); err != nil {
		return nil, fmt.Errorf("failed to set hide: %w", err)
	}

	// Set the canonical ID
	d.SetId(buildStaticRouteID(*found))

	// The Read function will populate the rest and validate consistency
	diags := resourceRTXStaticRouteRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to import static route: %v", diags[0].Summary)
	}

	// Check if the resource was found after read
	if d.Id() == "" {
		return nil, fmt.Errorf("static route validation failed after import")
	}

	return []*schema.ResourceData{d}, nil
}

// Helper functions

// buildStaticRouteID creates a composite ID from route components
func buildStaticRouteID(route client.StaticRoute) string {
	gateway := route.GatewayIP
	if gateway == "" && route.GatewayInterface != "" {
		gateway = "if:" + route.GatewayInterface
	}
	if gateway == "" && route.GatewayIP != "" {
		gateway = "ip:" + route.GatewayIP
	}
	
	return fmt.Sprintf("%s||%s||%s", route.Destination, gateway, route.Interface)
}

// parseStaticRouteID parses the composite ID into components
func parseStaticRouteID(id string) (destination, gateway, iface string, err error) {
	parts := strings.SplitN(id, "||", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected format 'destination||gateway||interface', got %s", id)
	}

	destination = parts[0]
	gateway = parts[1]
	iface = parts[2]

	// Parse gateway type prefix if present
	if strings.HasPrefix(gateway, "ip:") {
		gateway = strings.TrimPrefix(gateway, "ip:")
	} else if strings.HasPrefix(gateway, "if:") {
		gateway = strings.TrimPrefix(gateway, "if:")
	}

	return destination, gateway, iface, nil
}

// getIntWithDefault gets integer value from resource data with default
func getIntWithDefault(d *schema.ResourceData, key string, defaultValue int) int {
	if v, ok := d.GetOk(key); ok {
		return v.(int)
	}
	return defaultValue
}

// Validation functions

// validateCIDR validates CIDR notation
func validateCIDR(v interface{}, k string) (warns []string, errs []error) {
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("expected type of %q to be string", k))
		return warns, errs
	}

	if value == "" {
		errs = append(errs, fmt.Errorf("%q cannot be empty", k))
		return warns, errs
	}

	// Parse CIDR
	if _, _, err := net.ParseCIDR(value); err != nil {
		errs = append(errs, fmt.Errorf("%q is not a valid CIDR notation: %v", k, err))
	}

	return warns, errs
}

// validateInterfaceName validates RTX interface names
func validateInterfaceName(v interface{}, k string) (warns []string, errs []error) {
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("expected type of %q to be string", k))
		return warns, errs
	}

	if value == "" {
		return warns, errs // Empty is allowed for optional fields
	}

	// RTX interface name pattern: wan1, lan1, pp1, tunnel1, etc.
	validPattern := regexp.MustCompile(`^(wan|lan|pp|tunnel|loopback)\d+$`)
	if !validPattern.MatchString(value) {
		errs = append(errs, fmt.Errorf("%q must be a valid RTX interface name (wan1, lan1, pp1, tunnel1, etc.)", k))
	}

	return warns, errs
}

// normalizeCIDR normalizes CIDR notation
func normalizeCIDR(val interface{}) string {
	if val == nil {
		return ""
	}

	cidr, ok := val.(string)
	if !ok {
		return ""
	}

	// Parse and reformat to canonical form
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Return original value if parsing fails
		return cidr
	}

	// Get prefix length
	ones, _ := ipNet.Mask.Size()

	// Return canonical form
	return fmt.Sprintf("%s/%d", ip, ones)
}

// suppressDefault1 suppresses diff when default value is 1
func suppressDefault1(k, old, new string, d *schema.ResourceData) bool {
	return (old == "" && new == "1") || (old == "1" && new == "")
}

// validateGatewaySpecification ensures at least one gateway method is specified
func validateGatewaySpecification(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	gatewayIP := d.Get("gateway_ip").(string)
	gatewayInterface := d.Get("gateway_interface").(string)

	// Check that at least one gateway method is specified
	if gatewayIP == "" && gatewayInterface == "" {
		return fmt.Errorf("exactly one of 'gateway_ip' or 'gateway_interface' must be specified")
	}

	return nil
}