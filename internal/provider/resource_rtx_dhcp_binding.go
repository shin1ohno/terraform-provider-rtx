package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The MAC address of the device",
				StateFunc:   normalizeMACAddress,
			},
			"use_client_identifier": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Use ethernet client identifier format instead of MAC address",
			},
		},
	}
}

func resourceRTXDHCPBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	binding := client.DHCPBinding{
		ScopeID:             d.Get("scope_id").(int),
		IPAddress:           d.Get("ip_address").(string),
		MACAddress:          d.Get("mac_address").(string),
		UseClientIdentifier: d.Get("use_client_identifier").(bool),
	}

	err := apiClient.client.CreateDHCPBinding(ctx, binding)
	if err != nil {
		return diag.Errorf("Failed to create DHCP binding: %v", err)
	}

	// Set the ID as composite of scope_id and ip_address
	d.SetId(fmt.Sprintf("%d:%s", binding.ScopeID, binding.IPAddress))

	// Read back to ensure consistency
	return resourceRTXDHCPBindingRead(ctx, d, meta)
}

func resourceRTXDHCPBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Starting with ID=%s", d.Id())

	// Parse the composite ID
	scopeID, ipAddress, err := parseDHCPBindingID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Parsed scopeID=%d, ipAddress=%s", scopeID, ipAddress)

	// Get all bindings for the scope
	bindings, err := apiClient.client.GetDHCPBindings(ctx, scopeID)
	if err != nil {
		return diag.Errorf("Failed to retrieve DHCP bindings: %v", err)
	}
	
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Retrieved %d bindings", len(bindings))

	// Find our specific binding
	var found *client.DHCPBinding
	for _, binding := range bindings {
		log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Checking binding IP=%s against target=%s", binding.IPAddress, ipAddress)
		if binding.IPAddress == ipAddress {
			found = &binding
			break
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
	if err := d.Set("use_client_identifier", found.UseClientIdentifier); err != nil {
		return diag.FromErr(err)
	}
	
	// IMPORTANT: Always set the ID at the end of Read function
	d.SetId(fmt.Sprintf("%d:%s", found.ScopeID, found.IPAddress))
	log.Printf("[DEBUG] resourceRTXDHCPBindingRead: Set ID to %s", d.Id())

	return nil
}

func resourceRTXDHCPBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the composite ID
	scopeID, ipAddress, err := parseDHCPBindingID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	err = apiClient.client.DeleteDHCPBinding(ctx, scopeID, ipAddress)
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
	// Expected format: "scope_id:ip_address"
	importID := d.Id()
	scopeID, ipAddress, err := parseDHCPBindingID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'scope_id:ip_address': %v", err)
	}

	// Set the parsed values
	d.Set("scope_id", scopeID)
	d.Set("ip_address", ipAddress)
	
	// IMPORTANT: Keep the original ID for the Read function to use
	// The Read function expects the ID to be in the format "scope_id:ip_address"
	d.SetId(importID)

	// The Read function will populate the rest
	diags := resourceRTXDHCPBindingRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to import DHCP binding: %v", diags[0].Summary)
	}

	// Check if the resource was found
	if d.Id() == "" {
		return nil, fmt.Errorf("DHCP binding with scope_id=%d and ip_address=%s not found", scopeID, ipAddress)
	}

	return []*schema.ResourceData{d}, nil
}

// parseDHCPBindingID parses the composite ID into scope_id and ip_address
func parseDHCPBindingID(id string) (int, string, error) {
	// Use LastIndex to handle potential colons in IPv6 addresses
	lastColon := strings.LastIndex(id, ":")
	if lastColon == -1 {
		return 0, "", fmt.Errorf("expected format 'scope_id:ip_address', got %s", id)
	}
	
	scopeIDStr := id[:lastColon]
	ipAddress := id[lastColon+1:]
	
	scopeID, err := strconv.Atoi(scopeIDStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid scope_id: %v", err)
	}
	
	return scopeID, ipAddress, nil
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