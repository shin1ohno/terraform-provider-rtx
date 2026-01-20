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

func resourceRTXIPsecTunnel() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages IPsec VPN tunnel configuration on RTX routers. Supports IKEv2 with pre-shared key authentication.",
		CreateContext: resourceRTXIPsecTunnelCreate,
		ReadContext:   resourceRTXIPsecTunnelRead,
		UpdateContext: resourceRTXIPsecTunnelUpdate,
		DeleteContext: resourceRTXIPsecTunnelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXIPsecTunnelImport,
		},

		Schema: map[string]*schema.Schema{
			"tunnel_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "Tunnel ID (1-65535).",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Tunnel description/name.",
			},
			"local_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Local endpoint IP address.",
				ValidateFunc: validateIPv4Address,
			},
			"remote_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Remote endpoint IP address or hostname (for dynamic DNS).",
			},
			"pre_shared_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Pre-shared key for IKE authentication.",
			},
			"local_network": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Local network in CIDR notation (e.g., '192.168.1.0/24').",
			},
			"remote_network": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Remote network in CIDR notation (e.g., '10.0.0.0/24').",
			},
			"ikev2_proposal": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "IKE Phase 1 proposal settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_aes256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Use AES-256 encryption.",
						},
						"encryption_aes128": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use AES-128 encryption.",
						},
						"encryption_3des": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use 3DES encryption.",
						},
						"integrity_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Use SHA-256 integrity.",
						},
						"integrity_sha1": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use SHA-1 integrity.",
						},
						"integrity_md5": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use MD5 integrity.",
						},
						"group_fourteen": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Use DH group 14 (2048-bit).",
						},
						"group_five": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use DH group 5 (1536-bit).",
						},
						"group_two": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use DH group 2 (1024-bit).",
						},
						"lifetime_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      86400,
							Description:  "IKE SA lifetime in seconds. Default is 86400 (24 hours).",
							ValidateFunc: validation.IntAtLeast(60),
						},
					},
				},
			},
			"ipsec_transform": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "IPsec Phase 2 transform settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "esp",
							Description:  "IPsec protocol: 'esp' or 'ah'.",
							ValidateFunc: validation.StringInSlice([]string{"esp", "ah"}, false),
						},
						"encryption_aes256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Use AES-256 encryption.",
						},
						"encryption_aes128": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use AES-128 encryption.",
						},
						"encryption_3des": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use 3DES encryption.",
						},
						"integrity_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Use SHA-256-HMAC integrity.",
						},
						"integrity_sha1": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use SHA-1-HMAC integrity.",
						},
						"integrity_md5": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use MD5-HMAC integrity.",
						},
						"pfs_group_fourteen": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Use PFS with DH group 14.",
						},
						"pfs_group_five": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use PFS with DH group 5.",
						},
						"pfs_group_two": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use PFS with DH group 2.",
						},
						"lifetime_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3600,
							Description:  "IPsec SA lifetime in seconds. Default is 3600 (1 hour).",
							ValidateFunc: validation.IntAtLeast(60),
						},
					},
				},
			},
			"dpd_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable Dead Peer Detection.",
			},
			"dpd_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				Description:  "DPD interval in seconds.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"dpd_retry": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				Description:  "DPD retry count before declaring peer dead (0 means disabled).",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable the IPsec tunnel.",
			},
		},
	}
}

func resourceRTXIPsecTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	tunnel := buildIPsecTunnelFromResourceData(d)

	log.Printf("[DEBUG] Creating IPsec tunnel: %+v", tunnel)

	err := apiClient.client.CreateIPsecTunnel(ctx, tunnel)
	if err != nil {
		return diag.Errorf("Failed to create IPsec tunnel: %v", err)
	}

	d.SetId(strconv.Itoa(tunnel.ID))

	return resourceRTXIPsecTunnelRead(ctx, d, meta)
}

func resourceRTXIPsecTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	tunnelID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid tunnel ID: %v", err)
	}

	log.Printf("[DEBUG] Reading IPsec tunnel: %d", tunnelID)

	tunnel, err := apiClient.client.GetIPsecTunnel(ctx, tunnelID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DEBUG] IPsec tunnel %d not found, removing from state", tunnelID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IPsec tunnel: %v", err)
	}

	// Update the state
	if err := d.Set("tunnel_id", tunnel.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", tunnel.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("local_address", tunnel.LocalAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("remote_address", tunnel.RemoteAddress); err != nil {
		return diag.FromErr(err)
	}
	// Note: pre_shared_key is not read back from router for security
	if err := d.Set("local_network", tunnel.LocalNetwork); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("remote_network", tunnel.RemoteNetwork); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dpd_enabled", tunnel.DPDEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dpd_interval", tunnel.DPDInterval); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dpd_retry", tunnel.DPDRetry); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", tunnel.Enabled); err != nil {
		return diag.FromErr(err)
	}

	// Set IKEv2 proposal
	ikev2Proposal := []map[string]interface{}{
		{
			"encryption_aes256": tunnel.IKEv2Proposal.EncryptionAES256,
			"encryption_aes128": tunnel.IKEv2Proposal.EncryptionAES128,
			"encryption_3des":   tunnel.IKEv2Proposal.Encryption3DES,
			"integrity_sha256":  tunnel.IKEv2Proposal.IntegritySHA256,
			"integrity_sha1":    tunnel.IKEv2Proposal.IntegritySHA1,
			"integrity_md5":     tunnel.IKEv2Proposal.IntegrityMD5,
			"group_fourteen":    tunnel.IKEv2Proposal.GroupFourteen,
			"group_five":        tunnel.IKEv2Proposal.GroupFive,
			"group_two":         tunnel.IKEv2Proposal.GroupTwo,
			"lifetime_seconds":  tunnel.IKEv2Proposal.LifetimeSeconds,
		},
	}
	if err := d.Set("ikev2_proposal", ikev2Proposal); err != nil {
		return diag.FromErr(err)
	}

	// Set IPsec transform
	ipsecTransform := []map[string]interface{}{
		{
			"protocol":           tunnel.IPsecTransform.Protocol,
			"encryption_aes256":  tunnel.IPsecTransform.EncryptionAES256,
			"encryption_aes128":  tunnel.IPsecTransform.EncryptionAES128,
			"encryption_3des":    tunnel.IPsecTransform.Encryption3DES,
			"integrity_sha256":   tunnel.IPsecTransform.IntegritySHA256,
			"integrity_sha1":     tunnel.IPsecTransform.IntegritySHA1,
			"integrity_md5":      tunnel.IPsecTransform.IntegrityMD5,
			"pfs_group_fourteen": tunnel.IPsecTransform.PFSGroupFourteen,
			"pfs_group_five":     tunnel.IPsecTransform.PFSGroupFive,
			"pfs_group_two":      tunnel.IPsecTransform.PFSGroupTwo,
			"lifetime_seconds":   tunnel.IPsecTransform.LifetimeSeconds,
		},
	}
	if err := d.Set("ipsec_transform", ipsecTransform); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXIPsecTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	tunnel := buildIPsecTunnelFromResourceData(d)

	log.Printf("[DEBUG] Updating IPsec tunnel: %+v", tunnel)

	err := apiClient.client.UpdateIPsecTunnel(ctx, tunnel)
	if err != nil {
		return diag.Errorf("Failed to update IPsec tunnel: %v", err)
	}

	return resourceRTXIPsecTunnelRead(ctx, d, meta)
}

func resourceRTXIPsecTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	tunnelID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid tunnel ID: %v", err)
	}

	log.Printf("[DEBUG] Deleting IPsec tunnel: %d", tunnelID)

	err = apiClient.client.DeleteIPsecTunnel(ctx, tunnelID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete IPsec tunnel: %v", err)
	}

	return nil
}

func resourceRTXIPsecTunnelImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	tunnelID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid import ID, expected tunnel ID as integer: %v", err)
	}

	log.Printf("[DEBUG] Importing IPsec tunnel: %d", tunnelID)

	tunnel, err := apiClient.client.GetIPsecTunnel(ctx, tunnelID)
	if err != nil {
		return nil, fmt.Errorf("failed to import IPsec tunnel %d: %v", tunnelID, err)
	}

	d.SetId(strconv.Itoa(tunnel.ID))
	d.Set("tunnel_id", tunnel.ID)
	d.Set("name", tunnel.Name)
	d.Set("local_address", tunnel.LocalAddress)
	d.Set("remote_address", tunnel.RemoteAddress)
	// pre_shared_key is not retrievable
	d.Set("local_network", tunnel.LocalNetwork)
	d.Set("remote_network", tunnel.RemoteNetwork)
	d.Set("dpd_enabled", tunnel.DPDEnabled)
	d.Set("dpd_interval", tunnel.DPDInterval)
	d.Set("dpd_retry", tunnel.DPDRetry)
	d.Set("enabled", tunnel.Enabled)

	ikev2Proposal := []map[string]interface{}{
		{
			"encryption_aes256": tunnel.IKEv2Proposal.EncryptionAES256,
			"encryption_aes128": tunnel.IKEv2Proposal.EncryptionAES128,
			"encryption_3des":   tunnel.IKEv2Proposal.Encryption3DES,
			"integrity_sha256":  tunnel.IKEv2Proposal.IntegritySHA256,
			"integrity_sha1":    tunnel.IKEv2Proposal.IntegritySHA1,
			"integrity_md5":     tunnel.IKEv2Proposal.IntegrityMD5,
			"group_fourteen":    tunnel.IKEv2Proposal.GroupFourteen,
			"group_five":        tunnel.IKEv2Proposal.GroupFive,
			"group_two":         tunnel.IKEv2Proposal.GroupTwo,
			"lifetime_seconds":  tunnel.IKEv2Proposal.LifetimeSeconds,
		},
	}
	d.Set("ikev2_proposal", ikev2Proposal)

	ipsecTransform := []map[string]interface{}{
		{
			"protocol":           tunnel.IPsecTransform.Protocol,
			"encryption_aes256":  tunnel.IPsecTransform.EncryptionAES256,
			"encryption_aes128":  tunnel.IPsecTransform.EncryptionAES128,
			"encryption_3des":    tunnel.IPsecTransform.Encryption3DES,
			"integrity_sha256":   tunnel.IPsecTransform.IntegritySHA256,
			"integrity_sha1":     tunnel.IPsecTransform.IntegritySHA1,
			"integrity_md5":      tunnel.IPsecTransform.IntegrityMD5,
			"pfs_group_fourteen": tunnel.IPsecTransform.PFSGroupFourteen,
			"pfs_group_five":     tunnel.IPsecTransform.PFSGroupFive,
			"pfs_group_two":      tunnel.IPsecTransform.PFSGroupTwo,
			"lifetime_seconds":   tunnel.IPsecTransform.LifetimeSeconds,
		},
	}
	d.Set("ipsec_transform", ipsecTransform)

	return []*schema.ResourceData{d}, nil
}

func buildIPsecTunnelFromResourceData(d *schema.ResourceData) client.IPsecTunnel {
	tunnel := client.IPsecTunnel{
		ID:            d.Get("tunnel_id").(int),
		Name:          d.Get("name").(string),
		LocalAddress:  d.Get("local_address").(string),
		RemoteAddress: d.Get("remote_address").(string),
		PreSharedKey:  d.Get("pre_shared_key").(string),
		LocalNetwork:  d.Get("local_network").(string),
		RemoteNetwork: d.Get("remote_network").(string),
		DPDEnabled:    d.Get("dpd_enabled").(bool),
		DPDInterval:   d.Get("dpd_interval").(int),
		DPDRetry:      d.Get("dpd_retry").(int),
		Enabled:       d.Get("enabled").(bool),
	}

	// Handle IKEv2 proposal
	if v, ok := d.GetOk("ikev2_proposal"); ok {
		proposalList := v.([]interface{})
		if len(proposalList) > 0 {
			pMap := proposalList[0].(map[string]interface{})
			tunnel.IKEv2Proposal = client.IKEv2Proposal{
				EncryptionAES256: pMap["encryption_aes256"].(bool),
				EncryptionAES128: pMap["encryption_aes128"].(bool),
				Encryption3DES:   pMap["encryption_3des"].(bool),
				IntegritySHA256:  pMap["integrity_sha256"].(bool),
				IntegritySHA1:    pMap["integrity_sha1"].(bool),
				IntegrityMD5:     pMap["integrity_md5"].(bool),
				GroupFourteen:    pMap["group_fourteen"].(bool),
				GroupFive:        pMap["group_five"].(bool),
				GroupTwo:         pMap["group_two"].(bool),
				LifetimeSeconds:  pMap["lifetime_seconds"].(int),
			}
		}
	}

	// Handle IPsec transform
	if v, ok := d.GetOk("ipsec_transform"); ok {
		transformList := v.([]interface{})
		if len(transformList) > 0 {
			tMap := transformList[0].(map[string]interface{})
			tunnel.IPsecTransform = client.IPsecTransform{
				Protocol:         tMap["protocol"].(string),
				EncryptionAES256: tMap["encryption_aes256"].(bool),
				EncryptionAES128: tMap["encryption_aes128"].(bool),
				Encryption3DES:   tMap["encryption_3des"].(bool),
				IntegritySHA256:  tMap["integrity_sha256"].(bool),
				IntegritySHA1:    tMap["integrity_sha1"].(bool),
				IntegrityMD5:     tMap["integrity_md5"].(bool),
				PFSGroupFourteen: tMap["pfs_group_fourteen"].(bool),
				PFSGroupFive:     tMap["pfs_group_five"].(bool),
				PFSGroupTwo:      tMap["pfs_group_two"].(bool),
				LifetimeSeconds:  tMap["lifetime_seconds"].(int),
			}
		}
	}

	return tunnel
}
