package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

func resourceRTXSSHDAuthorizedKeys() *schema.Resource {
	return &schema.Resource{
		Description: "Manages SSH authorized keys for a user on RTX routers. " +
			"This resource allows you to configure SSH public key authentication for admin users. " +
			"Note: The router only returns fingerprints when reading keys, not the original key content. " +
			"After importing, you must provide the actual keys in your configuration.",
		CreateContext: resourceRTXSSHDAuthorizedKeysCreate,
		ReadContext:   resourceRTXSSHDAuthorizedKeysRead,
		UpdateContext: resourceRTXSSHDAuthorizedKeysUpdate,
		DeleteContext: resourceRTXSSHDAuthorizedKeysDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXSSHDAuthorizedKeysImport,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Username to manage authorized keys for. Must be an existing admin user. Changing this value forces a new resource to be created.",
				ValidateFunc: validateUsername,
			},
			"keys": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of SSH public keys in OpenSSH format (e.g., 'ssh-ed25519 AAAA... user@host'). Each key must be a valid SSH public key.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"key_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of authorized keys registered for this user.",
			},
		},
	}
}

func resourceRTXSSHDAuthorizedKeysCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	username := d.Get("username").(string)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	// Get keys from resource data
	keysRaw := d.Get("keys").([]interface{})
	keys := make([]string, len(keysRaw))
	for i, k := range keysRaw {
		keys[i] = k.(string)
	}

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("Creating SSHD authorized keys")

	// Set the authorized keys
	err := apiClient.client.SetSSHDAuthorizedKeys(ctx, username, keys)
	if err != nil {
		return diag.Errorf("Failed to set SSHD authorized keys for user %s: %v", username, err)
	}

	// Use username as the resource ID
	d.SetId(username)

	// Read back to ensure consistency
	return resourceRTXSSHDAuthorizedKeysRead(ctx, d, meta)
}

func resourceRTXSSHDAuthorizedKeysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	username := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Msg("Reading SSHD authorized keys")

	// Get the authorized keys from the router
	keys, err := apiClient.client.GetSSHDAuthorizedKeys(ctx, username)
	if err != nil {
		// Check if user or keys don't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no authorized keys") {
			logger.Debug().
				Str("resource", "rtx_sshd_authorized_keys").
				Str("username", username).
				Msg("No authorized keys found, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read SSHD authorized keys for user %s: %v", username, err)
	}

	// If no keys returned, remove from state
	if len(keys) == 0 {
		logger.Debug().
			Str("resource", "rtx_sshd_authorized_keys").
			Str("username", username).
			Msg("No authorized keys returned, removing from state")
		d.SetId("")
		return nil
	}

	// Set username
	if err := d.Set("username", username); err != nil {
		return diag.FromErr(err)
	}

	// Set key count (computed from router)
	if err := d.Set("key_count", len(keys)); err != nil {
		return diag.FromErr(err)
	}

	// Note: We can't read back the original key content, only fingerprints.
	// We keep the keys from state as-is (they were validated on create/update).
	// This is acceptable because:
	// 1. The keys in state were successfully applied during create/update
	// 2. The key_count matches, confirming the correct number of keys are registered
	// 3. If someone manually changes keys on the router, the count will differ

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("Read SSHD authorized keys successfully")

	return nil
}

func resourceRTXSSHDAuthorizedKeysUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	username := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	// Get new keys from resource data
	keysRaw := d.Get("keys").([]interface{})
	keys := make([]string, len(keysRaw))
	for i, k := range keysRaw {
		keys[i] = k.(string)
	}

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("Updating SSHD authorized keys")

	// SetSSHDAuthorizedKeys will delete all existing keys and re-register new ones
	err := apiClient.client.SetSSHDAuthorizedKeys(ctx, username, keys)
	if err != nil {
		return diag.Errorf("Failed to update SSHD authorized keys for user %s: %v", username, err)
	}

	return resourceRTXSSHDAuthorizedKeysRead(ctx, d, meta)
}

func resourceRTXSSHDAuthorizedKeysDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	username := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Msg("Deleting SSHD authorized keys")

	err := apiClient.client.DeleteSSHDAuthorizedKeys(ctx, username)
	if err != nil {
		// Check if already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete SSHD authorized keys for user %s: %v", username, err)
	}

	return nil
}

func resourceRTXSSHDAuthorizedKeysImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	username := d.Id()

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Msg("Importing SSHD authorized keys")

	// Verify keys exist for this user
	keys, err := apiClient.client.GetSSHDAuthorizedKeys(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to import SSHD authorized keys for user %s: %v", username, err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no authorized keys found for user %s", username)
	}

	// Set the ID and username
	d.SetId(username)
	d.Set("username", username)
	d.Set("key_count", len(keys))

	// Note: We can't recover the original key content, only fingerprints.
	// User will need to provide keys in config after import.
	// We set an empty list to avoid nil issues, but user MUST update config
	// with actual keys before applying.
	d.Set("keys", []string{})

	logger.Info().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("SSHD authorized keys imported. Note: You must provide the actual public keys in your configuration as they cannot be read from the router.")

	return []*schema.ResourceData{d}, nil
}
