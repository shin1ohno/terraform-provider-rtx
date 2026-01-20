package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXVLAN() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages VLAN interfaces on RTX routers. VLANs enable network segmentation using 802.1Q tagging on LAN interfaces.",
		CreateContext: resourceRTXVLANCreate,
		ReadContext:   resourceRTXVLANRead,
		UpdateContext: resourceRTXVLANUpdate,
		DeleteContext: resourceRTXVLANDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXVLANImport,
		},

		Schema: map[string]*schema.Schema{
			"vlan_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "The VLAN ID (1-4094)",
				ValidateFunc: validation.IntBetween(1, 4094),
			},
			"interface": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The parent interface (e.g., 'lan1', 'lan2')",
				ValidateFunc: validateVLANInterfaceName,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VLAN name/description",
			},
			"ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "IP address for the VLAN interface",
				ValidateFunc: validateIPAddress,
			},
			"ip_mask": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Subnet mask for the VLAN interface (required if ip_address is set)",
				ValidateFunc: validateSubnetMask,
			},
			"shutdown": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Administrative shutdown state (true = disabled, false = enabled)",
			},
			"vlan_interface": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The computed VLAN interface name (e.g., 'lan1/1')",
			},
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			// If ip_address is set, ip_mask must also be set
			ipAddress, ipAddressSet := diff.GetOk("ip_address")
			ipMask, ipMaskSet := diff.GetOk("ip_mask")

			if ipAddressSet && ipAddress.(string) != "" && (!ipMaskSet || ipMask.(string) == "") {
				return fmt.Errorf("ip_mask is required when ip_address is specified")
			}
			if ipMaskSet && ipMask.(string) != "" && (!ipAddressSet || ipAddress.(string) == "") {
				return fmt.Errorf("ip_address is required when ip_mask is specified")
			}

			return nil
		},
	}
}

func resourceRTXVLANCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	vlan := buildVLANFromResourceData(d)

	log.Printf("[DEBUG] Creating VLAN: %+v", vlan)

	err := apiClient.client.CreateVLAN(ctx, vlan)
	if err != nil {
		return diag.Errorf("Failed to create VLAN: %v", err)
	}

	// Set resource ID as interface/vlan_id format (e.g., "lan1/10")
	d.SetId(fmt.Sprintf("%s/%d", vlan.Interface, vlan.VlanID))

	// Read back to ensure consistency and get computed fields
	return resourceRTXVLANRead(ctx, d, meta)
}

func resourceRTXVLANRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the ID
	iface, vlanID, err := parseVLANID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	log.Printf("[DEBUG] Reading VLAN: %s/%d", iface, vlanID)

	vlan, err := apiClient.client.GetVLAN(ctx, iface, vlanID)
	if err != nil {
		// Check if VLAN doesn't exist
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] VLAN %s/%d not found, removing from state", iface, vlanID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read VLAN: %v", err)
	}

	// Update the state
	if err := d.Set("vlan_id", vlan.VlanID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interface", vlan.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", vlan.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip_address", vlan.IPAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip_mask", vlan.IPMask); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("shutdown", vlan.Shutdown); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vlan_interface", vlan.VlanInterface); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXVLANUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	vlan := buildVLANFromResourceData(d)

	// Get the current vlan_interface from state
	if v, ok := d.GetOk("vlan_interface"); ok {
		vlan.VlanInterface = v.(string)
	}

	log.Printf("[DEBUG] Updating VLAN: %+v", vlan)

	err := apiClient.client.UpdateVLAN(ctx, vlan)
	if err != nil {
		return diag.Errorf("Failed to update VLAN: %v", err)
	}

	return resourceRTXVLANRead(ctx, d, meta)
}

func resourceRTXVLANDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Parse the ID
	iface, vlanID, err := parseVLANID(d.Id())
	if err != nil {
		return diag.Errorf("Invalid resource ID: %v", err)
	}

	log.Printf("[DEBUG] Deleting VLAN: %s/%d", iface, vlanID)

	err = apiClient.client.DeleteVLAN(ctx, iface, vlanID)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete VLAN: %v", err)
	}

	return nil
}

func resourceRTXVLANImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as "interface/vlan_id" format (e.g., "lan1/10")
	iface, vlanID, err := parseVLANID(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected 'interface/vlan_id' (e.g., 'lan1/10'): %v", err)
	}

	log.Printf("[DEBUG] Importing VLAN: %s/%d", iface, vlanID)

	// Verify VLAN exists
	vlan, err := apiClient.client.GetVLAN(ctx, iface, vlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to import VLAN %s/%d: %v", iface, vlanID, err)
	}

	// Set all attributes
	d.SetId(fmt.Sprintf("%s/%d", iface, vlanID))
	d.Set("vlan_id", vlan.VlanID)
	d.Set("interface", vlan.Interface)
	d.Set("name", vlan.Name)
	d.Set("ip_address", vlan.IPAddress)
	d.Set("ip_mask", vlan.IPMask)
	d.Set("shutdown", vlan.Shutdown)
	d.Set("vlan_interface", vlan.VlanInterface)

	return []*schema.ResourceData{d}, nil
}

// buildVLANFromResourceData creates a VLAN from Terraform resource data
func buildVLANFromResourceData(d *schema.ResourceData) client.VLAN {
	vlan := client.VLAN{
		VlanID:    d.Get("vlan_id").(int),
		Interface: d.Get("interface").(string),
		Shutdown:  d.Get("shutdown").(bool),
	}

	if v, ok := d.GetOk("name"); ok {
		vlan.Name = v.(string)
	}

	if v, ok := d.GetOk("ip_address"); ok {
		vlan.IPAddress = v.(string)
	}

	if v, ok := d.GetOk("ip_mask"); ok {
		vlan.IPMask = v.(string)
	}

	if v, ok := d.GetOk("vlan_interface"); ok {
		vlan.VlanInterface = v.(string)
	}

	return vlan
}

// parseVLANID parses the resource ID in "interface/vlan_id" format
func parseVLANID(id string) (string, int, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("expected format 'interface/vlan_id', got %q", id)
	}

	iface := parts[0]
	vlanID, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid VLAN ID %q: %v", parts[1], err)
	}

	return iface, vlanID, nil
}

// validateVLANInterfaceName validates the VLAN parent interface name format
func validateVLANInterfaceName(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	// Interface must be in format "lanN" (e.g., lan1, lan2)
	if !strings.HasPrefix(value, "lan") {
		return nil, []error{fmt.Errorf("%q must be a LAN interface (e.g., 'lan1', 'lan2'), got %q", k, value)}
	}

	// Check that the rest is a number
	suffix := strings.TrimPrefix(value, "lan")
	if _, err := strconv.Atoi(suffix); err != nil {
		return nil, []error{fmt.Errorf("%q must be in format 'lanN' (e.g., 'lan1', 'lan2'), got %q", k, value)}
	}

	return nil, nil
}

// validateSubnetMask validates the subnet mask format
func validateSubnetMask(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	// Validate dotted decimal format
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return nil, []error{fmt.Errorf("%q must be a valid subnet mask (e.g., '255.255.255.0')", k)}
	}

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return nil, []error{fmt.Errorf("%q must be a valid subnet mask (e.g., '255.255.255.0')", k)}
		}
	}

	// Basic validation that it looks like a mask (first octets should be >= later octets)
	// A proper mask check would verify contiguous bits, but this is sufficient for basic validation
	var nums []int
	for _, part := range parts {
		num, _ := strconv.Atoi(part)
		nums = append(nums, num)
	}

	// Simple check: once we see a value less than 255, all following should be less than or equal
	foundPartial := false
	for i := 0; i < 4; i++ {
		if nums[i] < 255 {
			foundPartial = true
		} else if foundPartial {
			// Found 255 after a partial octet - invalid mask
			return nil, []error{fmt.Errorf("%q must be a valid subnet mask with contiguous bits", k)}
		}
	}

	return nil, nil
}
