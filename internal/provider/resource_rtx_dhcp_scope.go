package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// dhcpScopeMutexes provides per-scope locking to prevent concurrent modifications
var dhcpScopeMutexes = struct {
	mutexes map[string]*sync.Mutex
	mutex   sync.RWMutex
}{
	mutexes: make(map[string]*sync.Mutex),
}

// getDHCPScopeMutex gets or creates a mutex for the specified DHCP scope
func getDHCPScopeMutex(scopeKey string) *sync.Mutex {
	dhcpScopeMutexes.mutex.RLock()
	if mutex, exists := dhcpScopeMutexes.mutexes[scopeKey]; exists {
		dhcpScopeMutexes.mutex.RUnlock()
		return mutex
	}
	dhcpScopeMutexes.mutex.RUnlock()

	// Need to create new mutex
	dhcpScopeMutexes.mutex.Lock()
	defer dhcpScopeMutexes.mutex.Unlock()

	// Double-check in case another goroutine created it
	if mutex, exists := dhcpScopeMutexes.mutexes[scopeKey]; exists {
		return mutex
	}

	// Create new mutex
	mutex := &sync.Mutex{}
	dhcpScopeMutexes.mutexes[scopeKey] = mutex
	return mutex
}

// lockDHCPScope locks operations for a specific DHCP scope
func lockDHCPScope(scopeKey string) func() {
	mutex := getDHCPScopeMutex(scopeKey)
	mutex.Lock()
	return func() {
		mutex.Unlock()
	}
}

func resourceRTXDHCPScope() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages DHCP scope configurations on RTX routers",
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
				ValidateFunc: validation.IntBetween(1, 255),
				Description:  "The DHCP scope ID (1-255)",
			},
			"range_start": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "The start IP address of the DHCP range",
				StateFunc:    normalizeIPAddress,
			},
			"range_end": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "The end IP address of the DHCP range",
				StateFunc:    normalizeIPAddress,
			},
			"prefix": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(8, 32),
				Description:  "The network prefix length (e.g., 24 for /24)",
			},
			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "The gateway IP address for this scope",
				StateFunc:    normalizeIPAddress,
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    4,
				Description: "List of DNS server IP addresses for this scope (max 4)",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
			},
			"lease_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      86400,
				ValidateFunc: validation.IntBetween(60, 31536000), // 1 minute to 1 year
				Description:  "The lease time in seconds for this scope (default: 86400 = 24 hours)",
			},
			"domain_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
				Description:  "The domain name for this scope",
			},
		},
	}
}

func resourceRTXDHCPScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scopeID := d.Get("scope_id").(int)
	scopeKey := fmt.Sprintf("dhcp-scope-%d", scopeID)
	
	// Lock this specific scope to prevent concurrent modifications
	unlock := lockDHCPScope(scopeKey)
	defer unlock()

	scope := client.DHCPScope{
		ID:         scopeID,
		RangeStart: d.Get("range_start").(string),
		RangeEnd:   d.Get("range_end").(string),
		Prefix:     d.Get("prefix").(int),
	}

	// Set optional fields
	if v, ok := d.GetOk("gateway"); ok {
		scope.Gateway = v.(string)
	}

	if v, ok := d.GetOk("dns_servers"); ok {
		dnsServers := make([]string, len(v.([]interface{})))
		for i, dns := range v.([]interface{}) {
			dnsServers[i] = dns.(string)
		}
		scope.DNSServers = dnsServers
	}

	if v, ok := d.GetOk("lease_time"); ok {
		// Client expects lease time in seconds
		scope.Lease = v.(int)
	}

	if v, ok := d.GetOk("domain_name"); ok {
		scope.DomainName = v.(string)
	}

	// Validate range consistency
	if err := validateIPRange(scope.RangeStart, scope.RangeEnd); err != nil {
		return diag.Errorf("Invalid IP range: %v", err)
	}

	err := apiClient.client.CreateDHCPScope(ctx, scope)
	if err != nil {
		return diag.Errorf("Failed to create DHCP scope: %v", err)
	}

	// Set the resource ID to scope_id
	d.SetId(strconv.Itoa(scope.ID))

	// Read back to ensure consistency
	return resourceRTXDHCPScopeRead(ctx, d, meta)
}

func resourceRTXDHCPScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scopeID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	// Get the specific scope using the new GetDHCPScope method
	found, err := apiClient.client.GetDHCPScope(ctx, scopeID)
	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, client.ErrNotFound) {
			// Resource no longer exists, remove from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to retrieve DHCP scope: %v", err)
	}

	// Update the state
	if err := d.Set("scope_id", found.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("range_start", found.RangeStart); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("range_end", found.RangeEnd); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prefix", found.Prefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gateway", found.Gateway); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dns_servers", found.DNSServers); err != nil {
		return diag.FromErr(err)
	}

	// Lease time is already in seconds
	if err := d.Set("lease_time", found.Lease); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("domain_name", found.DomainName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXDHCPScopeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scopeID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	scopeKey := fmt.Sprintf("dhcp-scope-%d", scopeID)
	
	// Lock this specific scope to prevent concurrent modifications
	unlock := lockDHCPScope(scopeKey)
	defer unlock()

	scope := client.DHCPScope{
		ID:         scopeID,
		RangeStart: d.Get("range_start").(string),
		RangeEnd:   d.Get("range_end").(string),
		Prefix:     d.Get("prefix").(int),
	}

	// Set optional fields
	if v, ok := d.GetOk("gateway"); ok {
		scope.Gateway = v.(string)
	}

	if v, ok := d.GetOk("dns_servers"); ok {
		dnsServers := make([]string, len(v.([]interface{})))
		for i, dns := range v.([]interface{}) {
			dnsServers[i] = dns.(string)
		}
		scope.DNSServers = dnsServers
	}

	if v, ok := d.GetOk("lease_time"); ok {
		// Client expects lease time in seconds
		scope.Lease = v.(int)
	}

	if v, ok := d.GetOk("domain_name"); ok {
		scope.DomainName = v.(string)
	}

	// Validate range consistency
	if err := validateIPRange(scope.RangeStart, scope.RangeEnd); err != nil {
		return diag.Errorf("Invalid IP range: %v", err)
	}

	err = apiClient.client.UpdateDHCPScope(ctx, scope)
	if err != nil {
		return diag.Errorf("Failed to update DHCP scope: %v", err)
	}

	// Read back to ensure consistency
	return resourceRTXDHCPScopeRead(ctx, d, meta)
}

func resourceRTXDHCPScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	scopeID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	scopeKey := fmt.Sprintf("dhcp-scope-%d", scopeID)
	
	// Lock this specific scope to prevent concurrent modifications
	unlock := lockDHCPScope(scopeKey)
	defer unlock()

	err = apiClient.client.DeleteDHCPScope(ctx, scopeID)
	if err != nil {
		return diag.Errorf("Failed to delete DHCP scope: %v", err)
	}

	return nil
}

func resourceRTXDHCPScopeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scopeID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected scope_id as integer: %v", err)
	}

	// Set the scope_id for the read operation
	if err := d.Set("scope_id", scopeID); err != nil {
		return nil, fmt.Errorf("error setting scope_id: %v", err)
	}
	d.SetId(d.Id()) // Keep the same ID

	// The Read function will populate the rest and validate existence
	diags := resourceRTXDHCPScopeRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to import DHCP scope: %v", diags[0].Summary)
	}

	// Check if the resource was found after read
	if d.Id() == "" {
		return nil, fmt.Errorf("DHCP scope with ID %d not found", scopeID)
	}

	return []*schema.ResourceData{d}, nil
}

// validateIPRange validates that start IP is less than or equal to end IP
func validateIPRange(start, end string) error {
	startIP := net.ParseIP(start)
	endIP := net.ParseIP(end)

	if startIP == nil {
		return fmt.Errorf("invalid start IP address: %s", start)
	}

	if endIP == nil {
		return fmt.Errorf("invalid end IP address: %s", end)
	}

	// Convert to 4-byte representation for comparison
	startIP = startIP.To4()
	endIP = endIP.To4()

	if startIP == nil || endIP == nil {
		return fmt.Errorf("only IPv4 addresses are supported")
	}

	// Compare IP addresses byte by byte
	for i := 0; i < 4; i++ {
		if startIP[i] > endIP[i] {
			return fmt.Errorf("start IP %s must be less than or equal to end IP %s", start, end)
		}
		if startIP[i] < endIP[i] {
			break // start < end, validation passed
		}
	}

	return nil
}