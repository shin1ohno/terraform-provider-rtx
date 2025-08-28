package provider

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		
		// Custom validation to ensure each gateway has exactly one of IP or interface
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			if gatewaysRaw := diff.Get("gateways"); gatewaysRaw != nil {
				gateways := gatewaysRaw.(*schema.Set)
				for i, gw := range gateways.List() {
					gateway := gw.(map[string]interface{})
					hasIP := gateway["ip"].(string) != ""
					hasInterface := gateway["interface"].(string) != ""
					
					if hasIP && hasInterface {
						return fmt.Errorf("gateway %d: cannot specify both ip and interface", i)
					}
					if !hasIP && !hasInterface {
						return fmt.Errorf("gateway %d: must specify either ip or interface", i)
					}
				}
			}
			return nil
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
			"gateways": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				MinItems:    1,
				Description: "List of gateways for this route. Each gateway can be an IP address or interface name.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Gateway IP address",
							ValidateFunc: validation.IsIPAddress,
							StateFunc:    normalizeIPAddress,
						},
						"interface": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Gateway interface name (wan1, lan1, pp1, dhcp, etc.)",
							ValidateFunc: validateInterfaceName,
						},
						"weight": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "ECMP weight for this gateway (1-255)",
							ValidateFunc:     validation.IntBetween(1, 255),
							DiffSuppressFunc: suppressDefault1,
						},
						"hide": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Hide this gateway route from 'show ip route'",
						},
					},
				},
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Route description",
			},
		},

	}
}

func resourceRTXStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	destination := d.Get("destination").(string)
	gatewaysSet := d.Get("gateways").(*schema.Set)
	gateways := gatewaysSet.List()
	
	// Convert gateways to client.Gateway structs
	var gatewayList []client.Gateway
	for _, gw := range gateways {
		gateway := gw.(map[string]interface{})
		clientGateway := client.Gateway{
			Weight: getIntWithDefault1(gateway, "weight"),
			Hide:   gateway["hide"].(bool),
		}
		
		if ip, ok := gateway["ip"]; ok && ip.(string) != "" {
			clientGateway.IP = ip.(string)
		}
		if iface, ok := gateway["interface"]; ok && iface.(string) != "" {
			clientGateway.Interface = iface.(string)
		}
		
		gatewayList = append(gatewayList, clientGateway)
	}

	route := client.StaticRoute{
		Destination: destination,
		Gateways:    gatewayList,
		Metric:      getIntWithDefault(d, "metric", 1),
		Description: d.Get("description").(string),
	}

	if v, ok := d.GetOk("interface"); ok {
		route.Interface = v.(string)
	}

	err := apiClient.client.CreateStaticRoute(ctx, route)
	if err != nil {
		return diag.Errorf("Failed to create static route: %v", err)
	}

	// Set the composite ID: destination with gateway summary
	id := buildStaticRouteIDWithGateways(route)
	d.SetId(id)

	// Read back to ensure consistency
	return resourceRTXStaticRouteRead(ctx, d, meta)
}

func resourceRTXStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the composite ID to get destination
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
	if err := d.Set("interface", found.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metric", found.Metric); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", found.Description); err != nil {
		return diag.FromErr(err)
	}

	// Convert gateways to schema format
	var gateways []map[string]interface{}
	
	// If new Gateways field is populated, use it
	if len(found.Gateways) > 0 {
		for _, gw := range found.Gateways {
			gateway := map[string]interface{}{
				"weight": gw.Weight,
				"hide":   gw.Hide,
			}
			if gw.IP != "" {
				gateway["ip"] = gw.IP
			}
			if gw.Interface != "" {
				gateway["interface"] = gw.Interface
			}
			gateways = append(gateways, gateway)
		}
	} else {
		// Fallback to legacy fields for backwards compatibility
		gateway := map[string]interface{}{
			"weight": found.Weight,
			"hide":   found.Hide,
		}
		if found.GatewayIP != "" {
			gateway["ip"] = found.GatewayIP
		}
		if found.GatewayInterface != "" {
			gateway["interface"] = found.GatewayInterface
		}
		gateways = append(gateways, gateway)
	}

	// Convert to []interface{} for schema.NewSet
	var gatewaysInterface []interface{}
	for _, gw := range gateways {
		gatewaysInterface = append(gatewaysInterface, gw)
	}
	
	gatewaysSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"interface": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"hide": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}), gatewaysInterface)
	if err := d.Set("gateways", gatewaysSet); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	destination := d.Get("destination").(string)
	gatewaysSet := d.Get("gateways").(*schema.Set)
	gateways := gatewaysSet.List()
	
	// Convert gateways to client.Gateway structs
	var gatewayList []client.Gateway
	for _, gw := range gateways {
		gateway := gw.(map[string]interface{})
		clientGateway := client.Gateway{
			Weight: getIntWithDefault1(gateway, "weight"),
			Hide:   gateway["hide"].(bool),
		}
		
		if ip, ok := gateway["ip"]; ok && ip.(string) != "" {
			clientGateway.IP = ip.(string)
		}
		if iface, ok := gateway["interface"]; ok && iface.(string) != "" {
			clientGateway.Interface = iface.(string)
		}
		
		gatewayList = append(gatewayList, clientGateway)
	}

	route := client.StaticRoute{
		Destination: destination,
		Gateways:    gatewayList,
		Metric:      getIntWithDefault(d, "metric", 1),
		Description: d.Get("description").(string),
	}

	if v, ok := d.GetOk("interface"); ok {
		route.Interface = v.(string)
	}

	// Update the route
	err := apiClient.client.UpdateStaticRoute(ctx, route)
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

	// Set destination
	if err := d.Set("destination", found.Destination); err != nil {
		return nil, fmt.Errorf("failed to set destination: %w", err)
	}
	
	// Convert gateways to schema format
	var gateways []map[string]interface{}
	
	// If new Gateways field is populated, use it
	if len(found.Gateways) > 0 {
		for _, gw := range found.Gateways {
			gateway := map[string]interface{}{
				"weight": gw.Weight,
				"hide":   gw.Hide,
			}
			if gw.IP != "" {
				gateway["ip"] = gw.IP
			}
			if gw.Interface != "" {
				gateway["interface"] = gw.Interface
			}
			gateways = append(gateways, gateway)
		}
	} else {
		// Fallback to legacy fields for backwards compatibility
		gateway := map[string]interface{}{
			"weight": found.Weight,
			"hide":   found.Hide,
		}
		if found.GatewayIP != "" {
			gateway["ip"] = found.GatewayIP
		}
		if found.GatewayInterface != "" {
			gateway["interface"] = found.GatewayInterface
		}
		gateways = append(gateways, gateway)
	}

	// Convert to []interface{} for schema.NewSet
	var gatewaysInterface []interface{}
	for _, gw := range gateways {
		gatewaysInterface = append(gatewaysInterface, gw)
	}
	
	gatewaysSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"interface": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"hide": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}), gatewaysInterface)
	if err := d.Set("gateways", gatewaysSet); err != nil {
		return nil, fmt.Errorf("failed to set gateways: %w", err)
	}
	
	if err := d.Set("interface", found.Interface); err != nil {
		return nil, fmt.Errorf("failed to set interface: %w", err)
	}
	if err := d.Set("metric", found.Metric); err != nil {
		return nil, fmt.Errorf("failed to set metric: %w", err)
	}
	if err := d.Set("description", found.Description); err != nil {
		return nil, fmt.Errorf("failed to set description: %w", err)
	}

	// Set the canonical ID
	id := buildStaticRouteIDWithGateways(*found)
	d.SetId(id)

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
	gateway := getGateway(route)
	return fmt.Sprintf("%s||%s||%s", route.Destination, gateway, route.Interface)
}

// buildMultiRouteID creates a composite ID from multiple route IDs
func buildMultiRouteID(routeIDs []string) string {
	// Use a deterministic hash of the route IDs
	return fmt.Sprintf("multi:%s", strings.Join(routeIDs, ":"))
}

// getGateway extracts gateway string from route
func getGateway(route client.StaticRoute) string {
	if route.GatewayIP != "" {
		return route.GatewayIP
	} else if route.GatewayInterface != "" {
		return route.GatewayInterface
	}
	return ""
}

// buildStaticRouteIDWithGateways creates ID for multi-gateway routes
func buildStaticRouteIDWithGateways(route client.StaticRoute) string {
	var gatewayStrs []string
	for _, gw := range route.Gateways {
		if gw.IP != "" {
			gatewayStrs = append(gatewayStrs, gw.IP)
		} else if gw.Interface != "" {
			gatewayStrs = append(gatewayStrs, gw.Interface)
		}
	}
	gatewayStr := strings.Join(gatewayStrs, ",")
	return fmt.Sprintf("%s||%s||%s", route.Destination, gatewayStr, route.Interface)
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

	return destination, gateway, iface, nil
}

// getIntWithDefault gets integer value from resource data with default
func getIntWithDefault(d *schema.ResourceData, key string, defaultValue int) int {
	if v, ok := d.GetOk(key); ok {
		return v.(int)
	}
	return defaultValue
}

// getIntWithDefault1 gets integer value from map with default 1
func getIntWithDefault1(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok && v != nil {
		if val, ok := v.(int); ok && val > 0 {
			return val
		}
	}
	return 1
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

	// RTX interface name patterns:
	// - wan1, lan1, pp1, tunnel1, etc.
	// - dhcp (for gateway dhcp)
	// - dhcp lan2 (for gateway dhcp lan2)
	validPatterns := []string{
		`^(wan|lan|pp|tunnel|loopback)\d+$`,  // Standard interfaces
		`^dhcp$`,                             // DHCP gateway
		`^dhcp\s+(wan|lan|pp|tunnel)\d+$`,   // DHCP with specific interface
	}
	
	for _, pattern := range validPatterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return warns, errs
		}
	}

	errs = append(errs, fmt.Errorf("%q must be a valid RTX interface name (wan1, lan1, pp1, tunnel1, etc.), 'dhcp', or 'dhcp <interface>'", k))
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

	if cidr == "" {
		return ""
	}

	// Parse and reformat to canonical form
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Return original value if parsing fails
		return cidr
	}

	// Get prefix length
	ones, _ := ipNet.Mask.Size()

	// Return canonical form using network address
	return fmt.Sprintf("%s/%d", ipNet.IP, ones)
}


// suppressDefault1 suppresses diff when default value is 1
func suppressDefault1(k, old, new string, d *schema.ResourceData) bool {
	return (old == "" && new == "1") || (old == "1" && new == "")
}

