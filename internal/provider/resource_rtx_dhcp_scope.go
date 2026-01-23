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

func resourceRTXDHCPScope() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages DHCP scopes on RTX routers. A DHCP scope defines the IP address range and associated network parameters for DHCP address allocation.",
		CreateContext: resourceRTXDHCPScopeCreate,
		ReadContext:   resourceRTXDHCPScopeRead,
		UpdateContext: resourceRTXDHCPScopeUpdate,
		DeleteContext: resourceRTXDHCPScopeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXDHCPScopeImport,
		},

		Schema: map[string]*schema.Schema{
			"scope_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "The DHCP scope ID (positive integer)",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"network": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The network address in CIDR notation (e.g., '192.168.1.0/24')",
				ValidateFunc: validateCIDR,
			},
			"range_start": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Start IP address of the DHCP allocation range (parsed from IP range format)",
				ValidateFunc: validateIPAddress,
			},
			"range_end": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "End IP address of the DHCP allocation range (parsed from IP range format)",
				ValidateFunc: validateIPAddress,
			},
			"lease_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "DHCP lease duration in Go duration format (e.g., '72h', '30m') or 'infinite'",
				ValidateFunc: validateLeaseTime,
			},
			"exclude_ranges": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "IP address ranges to exclude from DHCP allocation",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Start IP address of the exclusion range",
							ValidateFunc: validateIPAddress,
						},
						"end": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "End IP address of the exclusion range",
							ValidateFunc: validateIPAddress,
						},
					},
				},
			},
			"options": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "DHCP options for client configuration (Cisco-compatible naming)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"routers": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    3,
							Description: "Default gateway addresses for DHCP clients (maximum 3)",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateIPAddress,
							},
						},
						"dns_servers": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    3,
							Description: "DNS server addresses for DHCP clients (maximum 3)",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateIPAddress,
							},
						},
						"domain_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Domain name for DHCP clients",
						},
					},
				},
			},
		},
	}
}

func resourceRTXDHCPScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scope := buildDHCPScopeFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dhcp_scope").Msgf("Creating DHCP scope: %+v", scope)

	err := apiClient.client.CreateDHCPScope(ctx, scope)
	if err != nil {
		return diag.Errorf("Failed to create DHCP scope: %v", err)
	}

	// Use scope_id as the resource ID
	d.SetId(strconv.Itoa(scope.ScopeID))

	// Read back to ensure consistency
	return resourceRTXDHCPScopeRead(ctx, d, meta)
}

func resourceRTXDHCPScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	scopeID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("Reading DHCP scope: %d", scopeID)

	var scope *client.DHCPScope

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract DHCP scopes from parsed config
			scopes := parsedConfig.ExtractDHCPScopes()
			for i := range scopes {
				if scopes[i].ScopeID == scopeID {
					scope = convertParsedDHCPScope(&scopes[i])
					logger.Debug().Str("resource", "rtx_dhcp_scope").Msg("Found scope in SFTP cache")
					break
				}
			}
		}
		if scope == nil {
			// Scope not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_dhcp_scope").Msg("Scope not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or scope not found in cache
	if scope == nil {
		scope, err = apiClient.client.GetDHCPScope(ctx, scopeID)
		if err != nil {
			// Check if scope doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_dhcp_scope").Msgf("DHCP scope %d not found, removing from state", scopeID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read DHCP scope: %v", err)
		}
	}

	// Update the state
	if err := d.Set("scope_id", scope.ScopeID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network", scope.Network); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("range_start", scope.RangeStart); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("range_end", scope.RangeEnd); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lease_time", scope.LeaseTime); err != nil {
		return diag.FromErr(err)
	}

	// Convert ExcludeRanges to list of maps
	excludeRanges := make([]map[string]interface{}, len(scope.ExcludeRanges))
	for i, r := range scope.ExcludeRanges {
		excludeRanges[i] = map[string]interface{}{
			"start": r.Start,
			"end":   r.End,
		}
	}
	if err := d.Set("exclude_ranges", excludeRanges); err != nil {
		return diag.FromErr(err)
	}

	// Convert Options to nested block
	options := []map[string]interface{}{}
	if len(scope.Options.Routers) > 0 || len(scope.Options.DNSServers) > 0 || scope.Options.DomainName != "" {
		optionsMap := map[string]interface{}{
			"routers":     scope.Options.Routers,
			"dns_servers": scope.Options.DNSServers,
			"domain_name": scope.Options.DomainName,
		}
		options = append(options, optionsMap)
	}
	if err := d.Set("options", options); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// convertParsedDHCPScope converts a parser DHCPScope to a client DHCPScope
func convertParsedDHCPScope(parsed *parsers.DHCPScope) *client.DHCPScope {
	scope := &client.DHCPScope{
		ScopeID:    parsed.ScopeID,
		Network:    parsed.Network,
		RangeStart: parsed.RangeStart,
		RangeEnd:   parsed.RangeEnd,
		LeaseTime:  parsed.LeaseTime,
		Options: client.DHCPScopeOptions{
			Routers:    parsed.Options.Routers,
			DNSServers: parsed.Options.DNSServers,
			DomainName: parsed.Options.DomainName,
		},
		ExcludeRanges: make([]client.ExcludeRange, len(parsed.ExcludeRanges)),
	}
	for i, r := range parsed.ExcludeRanges {
		scope.ExcludeRanges[i] = client.ExcludeRange{
			Start: r.Start,
			End:   r.End,
		}
	}
	return scope
}

func resourceRTXDHCPScopeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scope := buildDHCPScopeFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dhcp_scope").Msgf("Updating DHCP scope: %+v", scope)

	err := apiClient.client.UpdateDHCPScope(ctx, scope)
	if err != nil {
		return diag.Errorf("Failed to update DHCP scope: %v", err)
	}

	return resourceRTXDHCPScopeRead(ctx, d, meta)
}

func resourceRTXDHCPScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scopeID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dhcp_scope").Msgf("Deleting DHCP scope: %d", scopeID)

	err = apiClient.client.DeleteDHCPScope(ctx, scopeID)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete DHCP scope: %v", err)
	}

	return nil
}

func resourceRTXDHCPScopeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as scope_id
	scopeID, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected scope_id (integer): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_dhcp_scope").Msgf("Importing DHCP scope: %d", scopeID)

	// Verify scope exists
	scope, err := apiClient.client.GetDHCPScope(ctx, scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to import DHCP scope %d: %v", scopeID, err)
	}

	// Set all attributes
	d.SetId(strconv.Itoa(scopeID))
	d.Set("scope_id", scope.ScopeID)
	d.Set("network", scope.Network)
	d.Set("range_start", scope.RangeStart)
	d.Set("range_end", scope.RangeEnd)
	d.Set("lease_time", scope.LeaseTime)

	excludeRanges := make([]map[string]interface{}, len(scope.ExcludeRanges))
	for i, r := range scope.ExcludeRanges {
		excludeRanges[i] = map[string]interface{}{
			"start": r.Start,
			"end":   r.End,
		}
	}
	d.Set("exclude_ranges", excludeRanges)

	// Set options block
	options := []map[string]interface{}{}
	if len(scope.Options.Routers) > 0 || len(scope.Options.DNSServers) > 0 || scope.Options.DomainName != "" {
		optionsMap := map[string]interface{}{
			"routers":     scope.Options.Routers,
			"dns_servers": scope.Options.DNSServers,
			"domain_name": scope.Options.DomainName,
		}
		options = append(options, optionsMap)
	}
	d.Set("options", options)

	return []*schema.ResourceData{d}, nil
}

// buildDHCPScopeFromResourceData creates a DHCPScope from Terraform resource data
func buildDHCPScopeFromResourceData(d *schema.ResourceData) client.DHCPScope {
	scope := client.DHCPScope{
		ScopeID:   d.Get("scope_id").(int),
		Network:   d.Get("network").(string),
		LeaseTime: d.Get("lease_time").(string),
	}

	// Handle options block
	if v, ok := d.GetOk("options"); ok {
		optionsList := v.([]interface{})
		if len(optionsList) > 0 {
			optionsMap := optionsList[0].(map[string]interface{})

			// Parse routers
			if routersRaw, ok := optionsMap["routers"]; ok {
				routersList := routersRaw.([]interface{})
				routers := make([]string, len(routersList))
				for i, r := range routersList {
					routers[i] = r.(string)
				}
				scope.Options.Routers = routers
			}

			// Parse dns_servers
			if dnsRaw, ok := optionsMap["dns_servers"]; ok {
				dnsList := dnsRaw.([]interface{})
				dnsServers := make([]string, len(dnsList))
				for i, dns := range dnsList {
					dnsServers[i] = dns.(string)
				}
				scope.Options.DNSServers = dnsServers
			}

			// Parse domain_name
			if domainRaw, ok := optionsMap["domain_name"]; ok {
				scope.Options.DomainName = domainRaw.(string)
			}
		}
	}

	// Handle exclude_ranges
	if v, ok := d.GetOk("exclude_ranges"); ok {
		excludeRangesRaw := v.([]interface{})
		excludeRanges := make([]client.ExcludeRange, len(excludeRangesRaw))
		for i, r := range excludeRangesRaw {
			rangeMap := r.(map[string]interface{})
			excludeRanges[i] = client.ExcludeRange{
				Start: rangeMap["start"].(string),
				End:   rangeMap["end"].(string),
			}
		}
		scope.ExcludeRanges = excludeRanges
	}

	return scope
}

// validateCIDR validates that a string is a valid CIDR notation
func validateCIDR(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	_, _, err := net.ParseCIDR(value)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a valid CIDR notation (e.g., '192.168.1.0/24'): %v", k, err)}
	}

	return nil, nil
}

// validateIPAddress validates that a string is a valid IPv4 address
func validateIPAddress(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IP address", k)}
	}

	// Ensure it's IPv4
	if ip.To4() == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IPv4 address", k)}
	}

	return nil, nil
}

// validateIPAddressAny validates that a string is a valid IPv4 or IPv6 address
func validateIPAddressAny(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return nil, []error{fmt.Errorf("%q must be a valid IP address (IPv4 or IPv6)", k)}
	}

	return nil, nil
}

// validateLeaseTime validates the lease time format
func validateLeaseTime(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" || value == "infinite" {
		return nil, nil
	}

	// Check for Go duration format (e.g., "72h", "30m", "1h30m")
	value = strings.ToLower(value)

	// Simple validation: should contain at least one time unit
	hasUnit := false
	for _, unit := range []string{"h", "m", "s"} {
		if strings.Contains(value, unit) {
			hasUnit = true
			break
		}
	}

	if !hasUnit {
		return nil, []error{fmt.Errorf("%q must be a valid duration (e.g., '72h', '30m') or 'infinite'", k)}
	}

	return nil, nil
}
