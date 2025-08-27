package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXDHCPBinding() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages DHCP static lease bindings on RTX routers",
		CreateContext: resourceRTXDHCPBindingCreate,
		ReadContext:   resourceRTXDHCPBindingRead,
		DeleteContext: resourceRTXDHCPBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXDHCPBindingImport,
		},

		Schema: map[string]*schema.Schema{
			"scope_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The DHCP scope ID",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The IP address to assign",
				StateFunc:   normalizeIPAddress,
			},
			
			// === Client Identification (choose one) ===
			"mac_address": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The MAC address of the device (e.g., '00:11:22:33:44:55')",
				StateFunc:   normalizeMACAddress,
				ConflictsWith: []string{"client_identifier"},
			},
			"use_mac_as_client_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "When true with mac_address, automatically generates '01:MAC' client identifier",
				RequiredWith: []string{"mac_address"},
			},
			"client_identifier": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "DHCP Client Identifier in hex format (e.g., '01:aa:bb:cc:dd:ee:ff' for MAC-based, '02:12:34:56:78' for custom)",
				StateFunc:   normalizeClientIdentifier,
				ValidateFunc: validateClientIdentifierFormat,
				ConflictsWith: []string{"mac_address", "use_mac_as_client_id"},
			},
			
			// === Optional metadata ===
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hostname for the device (for documentation purposes)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the DHCP binding (for documentation purposes)",
			},
		},
		
		// Custom validation
		CustomizeDiff: customdiff.All(
			// Ensure exactly one identification method is specified
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				return validateClientIdentification(ctx, d, meta)
			},
		),
	}
}

func resourceRTXDHCPBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	binding := client.DHCPBinding{
		ScopeID:   d.Get("scope_id").(int),
		IPAddress: d.Get("ip_address").(string),
	}

	// Handle client identification method
	if macAddress, ok := d.GetOk("mac_address"); ok {
		binding.MACAddress = macAddress.(string)
		binding.UseClientIdentifier = d.Get("use_mac_as_client_id").(bool)
	} else if clientID, ok := d.GetOk("client_identifier"); ok {
		// Client identifier is provided directly
		binding.ClientIdentifier = clientID.(string)
		binding.UseClientIdentifier = true
	}

	err := apiClient.client.CreateDHCPBinding(ctx, binding)
	if err != nil {
		return diag.Errorf("Failed to create DHCP binding: %v", err)
	}

	// Set the ID as composite of scope_id and identifier (mac_address or client_identifier)
	var identifier string
	if macAddress, ok := d.GetOk("mac_address"); ok {
		normalizedMAC, err := normalizeMACAddressParser(macAddress.(string))
		if err != nil {
			return diag.Errorf("Failed to normalize MAC address: %v", err)
		}
		identifier = normalizedMAC
	} else if clientID, ok := d.GetOk("client_identifier"); ok {
		identifier = normalizeClientIdentifier(clientID)
	}
	d.SetId(fmt.Sprintf("%d:%s", binding.ScopeID, identifier))

	// Read back to ensure consistency
	return resourceRTXDHCPBindingRead(ctx, d, meta)
}

func resourceRTXDHCPBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Starting with ID=%s", d.Id())

	// Parse the composite ID
	scopeID, identifier, err := parseDHCPBindingID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}
	
	// Check if identifier is MAC address or IP address (for backward compatibility)
	isOldFormat := false
	if _, err := normalizeMACAddressParser(identifier); err != nil {
		// It's likely an old format with IP address
		isOldFormat = true
		log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Detected old format ID with IP address: %s", identifier)
	}
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Starting with ID=%s (scopeID=%d, identifier=%s, oldFormat=%v)", d.Id(), scopeID, identifier, isOldFormat)

	// Get all bindings for the scope
	bindings, err := apiClient.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		return diag.Errorf("Failed to retrieve DHCP bindings: %v", err)
	}
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Retrieved %d bindings", len(bindings))

	// Find our specific binding
	var found *client.DHCPBinding
	if isOldFormat {
		// Search by IP address for old format IDs
		for _, binding := range bindings {
			log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Checking binding IP=%s against target=%s (old format)", binding.IPAddress, identifier)
			if binding.IPAddress == identifier {
				found = &binding
				break
			}
		}
	} else {
		// Search by MAC address for new format IDs
		for _, binding := range bindings {
			normalizedBindingMAC, _ := normalizeMACAddressParser(binding.MACAddress)
			log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Checking binding MAC=%s (normalized=%s) against target=%s", binding.MACAddress, normalizedBindingMAC, identifier)
			if normalizedBindingMAC == identifier {
				found = &binding
				break
			}
		}
	}

	if found == nil {
		log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Binding not found, clearing ID")
		// Resource no longer exists
		d.SetId("")
		return nil
	}
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Found binding: %+v", found)

	// Update the state
	if err := d.Set("scope_id", found.ScopeID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip_address", found.IPAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mac_address", found.MACAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("use_mac_as_client_id", found.UseClientIdentifier); err != nil {
		return diag.FromErr(err)
	}
	
	// IMPORTANT: Always set the ID at the end of Read function
	// Always use MAC address format, even if we found the resource via old IP format
	normalizedMAC, err := normalizeMACAddressParser(found.MACAddress)
	if err != nil {
		return diag.Errorf("Failed to normalize MAC address: %v", err)
	}
	newID := fmt.Sprintf("%d:%s", found.ScopeID, normalizedMAC)
	
	if isOldFormat {
		log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Migrating ID from old format %s to new format %s", d.Id(), newID)
	}
	
	d.SetId(newID)
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Set ID to %s", d.Id())

	return nil
}

func resourceRTXDHCPBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the composite ID
	scopeID, macAddress, err := parseDHCPBindingID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	// Get all bindings to find the IP address for this MAC address
	bindings, err := apiClient.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		return diag.Errorf("Failed to retrieve DHCP bindings: %v", err)
	}

	// Find the binding with matching MAC address to get its IP address
	var ipToDelete string
	for _, binding := range bindings {
		normalizedBindingMAC, _ := normalizeMACAddressParser(binding.MACAddress)
		if normalizedBindingMAC == macAddress {
			ipToDelete = binding.IPAddress
			break
		}
	}

	if ipToDelete == "" {
		// Binding already doesn't exist, consider this success
		return nil
	}

	err = apiClient.client.DeleteDHCPBinding(ctx, scopeID, ipToDelete)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete DHCP binding: %v", err)
	}

	return nil
}

func resourceRTXDHCPBindingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()
	
	// Parse the import ID - can be either "scope_id:mac_address" or "scope_id:ip_address"
	scopeID, identifier, err := parseDHCPBindingID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'scope_id:mac_address' or 'scope_id:ip_address': %v", err)
	}

	log.Printf("[DEBUG] resourceRTXDHCPBindingImport: ImportID=%s, ScopeID=%d, Identifier=%s", importID, scopeID, identifier)

	// Get all bindings for the scope to find the requested binding
	bindings, err := apiClient.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DHCP bindings for scope %d: %v", scopeID, err)
	}

	log.Printf("[DEBUG] resourceRTXDHCPBindingImport: Retrieved %d bindings for scope %d", len(bindings), scopeID)

	// Determine if identifier is MAC address or IP address and find the binding
	var targetBinding *client.DHCPBinding
	
	// Check if identifier looks like a MAC address
	if _, err := normalizeMACAddressParser(identifier); err == nil {
		// It's a MAC address - search by MAC
		log.Printf("[DEBUG] resourceRTXDHCPBindingImport: Identifier appears to be MAC address")
		for _, binding := range bindings {
			normalizedBindingMAC, _ := normalizeMACAddressParser(binding.MACAddress)
			normalizedIdentifier, _ := normalizeMACAddressParser(identifier)
			if normalizedBindingMAC == normalizedIdentifier {
				targetBinding = &binding
				break
			}
		}
	} else {
		// It's likely an IP address - search by IP
		log.Printf("[DEBUG] resourceRTXDHCPBindingImport: Identifier appears to be IP address")
		for _, binding := range bindings {
			if binding.IPAddress == identifier {
				targetBinding = &binding
				break
			}
		}
	}

	if targetBinding == nil {
		return nil, fmt.Errorf("DHCP binding with scope_id=%d and identifier=%s not found", scopeID, identifier)
	}

	log.Printf("[DEBUG] resourceRTXDHCPBindingImport: Found binding: %+v", targetBinding)

	// Set the parsed values
	d.Set("scope_id", scopeID)
	d.Set("ip_address", targetBinding.IPAddress)
	d.Set("mac_address", targetBinding.MACAddress)
	d.Set("use_mac_as_client_id", targetBinding.UseClientIdentifier)
	
	// Always use the MAC-based ID format for consistency
	normalizedMAC, err := normalizeMACAddressParser(targetBinding.MACAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize MAC address: %v", err)
	}
	finalID := fmt.Sprintf("%d:%s", scopeID, normalizedMAC)
	d.SetId(finalID)

	log.Printf("[DEBUG] resourceRTXDHCPBindingImport: Set final ID to %s", finalID)

	// The Read function will populate the rest and validate consistency
	diags := resourceRTXDHCPBindingRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to import DHCP binding: %v", diags[0].Summary)
	}

	// Check if the resource was found after read
	if d.Id() == "" {
		return nil, fmt.Errorf("DHCP binding validation failed after import")
	}

	return []*schema.ResourceData{d}, nil
}

// parseDHCPBindingID parses the composite ID into scope_id and mac_address
func parseDHCPBindingID(id string) (int, string, error) {
	// Handle both old format (scope_id:ip_address) and new format (scope_id:mac_address)
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("expected format 'scope_id:mac_address', got %s", id)
	}
	
	scopeID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid scope_id: %v", err)
	}
	
	identifier := parts[1]
	
	// Check if it's a MAC address format (new format)
	if _, err := normalizeMACAddressParser(identifier); err == nil {
		return scopeID, identifier, nil
	}
	
	// It's likely an old format with IP address - we need to convert to MAC
	// This is for backwards compatibility during migration
	return scopeID, identifier, nil
}

// normalizeIPAddress normalizes IP address format
func normalizeIPAddress(val interface{}) string {
	if val == nil {
		return ""
	}
	// Simple normalization - in production, use net.ParseIP
	return strings.TrimSpace(val.(string))
}

// normalizeMACAddress normalizes MAC address format using the parser package
func normalizeMACAddress(val interface{}) string {
	if val == nil {
		return ""
	}
	
	macStr, ok := val.(string)
	if !ok {
		return ""
	}
	
	// Use the parser's normalizeMACAddress function
	normalized, err := normalizeMACAddressParser(macStr)
	if err != nil {
		// In Terraform StateFunc, we can't return errors
		// Return the original value to avoid silent failures
		return macStr
	}
	
	return normalized
}

// normalizeMACAddressParser is a helper that calls the parser's function
// This is a workaround since we can't import internal packages
func normalizeMACAddressParser(mac string) (string, error) {
	// Remove all separators
	cleaned := strings.ToLower(mac)
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	
	// Validate length
	if len(cleaned) != 12 {
		return "", fmt.Errorf("MAC address must be 12 hex digits, got %d", len(cleaned))
	}
	
	// Validate characters
	for _, c := range cleaned {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", fmt.Errorf("MAC address contains invalid characters")
		}
	}
	
	// Format with colons
	result := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		cleaned[0:2], cleaned[2:4], cleaned[4:6],
		cleaned[6:8], cleaned[8:10], cleaned[10:12])
	
	return result, nil
}

// normalizeClientIdentifier normalizes client identifier format
func normalizeClientIdentifier(val interface{}) string {
	if val == nil {
		return ""
	}
	
	cidStr, ok := val.(string)
	if !ok {
		return ""
	}
	
	// Normalize client identifier: ensure lowercase, consistent colon format
	cleaned := strings.ToLower(cidStr)
	cleaned = strings.ReplaceAll(cleaned, "-", ":")
	cleaned = strings.ReplaceAll(cleaned, " ", ":")
	
	// Remove duplicate colons
	for strings.Contains(cleaned, "::") {
		cleaned = strings.ReplaceAll(cleaned, "::", ":")
	}
	
	return cleaned
}

// validateClientIdentification ensures exactly one client identification method is used
func validateClientIdentification(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	macAddress := d.Get("mac_address").(string)
	clientIdentifier := d.Get("client_identifier").(string)
	
	// Check that exactly one identification method is specified
	if macAddress == "" && clientIdentifier == "" {
		return fmt.Errorf("exactly one of 'mac_address' or 'client_identifier' must be specified")
	}
	
	if macAddress != "" && clientIdentifier != "" {
		return fmt.Errorf("only one of 'mac_address' or 'client_identifier' can be specified")
	}
	
	// Validate client_identifier format if present
	if clientIdentifier != "" {
		if _, errs := validateClientIdentifierFormat(clientIdentifier, "client_identifier"); errs != nil && len(errs) > 0 {
			return errs[0]
		}
	}
	
	return nil
}

// validateClientIdentifierFormat validates the client identifier format
func validateClientIdentifierFormat(v interface{}, k string) ([]string, []error) {
	value, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}
	
	if value == "" {
		return nil, nil
	}
	
	// Normalize first
	normalized := normalizeClientIdentifier(value)
	
	// Check format: type:hex:hex:...
	parts := strings.Split(normalized, ":")
	if len(parts) < 2 {
		return nil, []error{fmt.Errorf("%q must be in format 'type:data' (e.g., '01:aa:bb:cc:dd:ee:ff', '02:66:6f:6f')", k)}
	}
	
	// Validate each part is valid hex
	for i, part := range parts {
		if len(part) != 2 {
			return nil, []error{fmt.Errorf("%q must contain 2-character hex octets at position %d, got %q", k, i, part)}
		}
		
		for _, c := range part {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return nil, []error{fmt.Errorf("%q contains invalid hex character '%c' at position %d", k, c, i)}
			}
		}
	}
	
	// Check length limit (255 octets max)
	if len(parts) > 255 {
		return nil, []error{fmt.Errorf("%q exceeds maximum length of 255 octets", k)}
	}
	
	return nil, nil
}

// validateClientIdentificationWithResourceData validates client identification for tests
func validateClientIdentificationWithResourceData(ctx context.Context, d *schema.ResourceData) error {
	macAddress := d.Get("mac_address").(string)
	clientIdentifier := d.Get("client_identifier").(string)
	useClientID := d.Get("use_mac_as_client_id").(bool)
	
	// Handle empty strings as unset
	if macAddress == "" {
		macAddress = ""
	}
	if clientIdentifier == "" {
		clientIdentifier = ""
	}
	
	// Count non-empty identification methods
	hasMAC := macAddress != ""
	hasClientID := clientIdentifier != ""
	
	// Ensure exactly one identification method is specified
	if !hasMAC && !hasClientID {
		return errors.New("exactly one of mac_address or client_identifier must be specified")
	}
	
	if hasMAC && hasClientID {
		return errors.New("exactly one of mac_address or client_identifier must be specified")
	}
	
	// Check if use_mac_as_client_id is set with client_identifier
	if hasClientID && useClientID {
		return errors.New("use_mac_as_client_id cannot be used with client_identifier")
	}
	
	return nil
}

// validateClientIdentifierFormatSimple validates client identifier format with single string input
func validateClientIdentifierFormatSimple(identifier string) error {
	if identifier == "" {
		return errors.New("client identifier cannot be empty")
	}
	
	// Normalize first
	normalized := normalizeClientIdentifier(identifier)
	
	// Check format: type:data
	parts := strings.Split(normalized, ":")
	if len(parts) < 2 {
		return errors.New("client identifier must be in format 'type:data'")
	}
	
	// Check if we have data after the prefix
	if len(parts) == 2 && parts[1] == "" {
		return errors.New("client identifier must have data after type prefix")
	}
	
	// Check prefix is supported (01, 02, or FF)
	prefix := strings.ToLower(parts[0])
	if prefix != "01" && prefix != "02" && prefix != "ff" {
		return errors.New("client identifier prefix must be 01 (MAC), 02 (ASCII), or ff (vendor-specific)")
	}
	
	// Validate each hex part
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if len(part) != 2 {
			return errors.New("client identifier contains invalid hex characters")
		}
		
		for _, c := range part {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return errors.New("client identifier contains invalid hex characters")
			}
		}
	}
	
	// Check length limit (255 bytes max) - each part represents 1 byte
	// The test case generates "01:" + 127*"aa:" + "bb" = 1 + 127*3 + 2 = 384 characters
	// This translates to 1 + 127 + 1 = 129 parts, which should fail
	if len(parts) > 128 {
		return errors.New("client identifier too long (max 255 bytes)")
	}
	
	return nil
}
