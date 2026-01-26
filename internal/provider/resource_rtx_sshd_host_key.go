package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

func resourceRTXSSHDHostKey() *schema.Resource {
	return &schema.Resource{
		Description: "Manages SSH host key on RTX routers. " +
			"This is a singleton resource - only one instance should exist per router. " +
			"The host key is used for SSH server authentication. " +
			"If no host key exists, creating this resource will generate one.",
		CreateContext: resourceRTXSSHDHostKeyCreate,
		ReadContext:   resourceRTXSSHDHostKeyRead,
		DeleteContext: resourceRTXSSHDHostKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXSSHDHostKeyImport,
		},

		Schema: map[string]*schema.Schema{
			"fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SSH host key fingerprint",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Host key algorithm (e.g., ssh-rsa)",
			},
		},
	}
}

func resourceRTXSSHDHostKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", "sshd_host_key")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Creating SSHD host key resource")

	// Get current host key info
	keyInfo, err := apiClient.client.GetSSHDHostKey(ctx)
	if err != nil {
		return diag.Errorf("Failed to get SSHD host key info: %v", err)
	}

	// If no key exists (empty fingerprint), generate one
	if keyInfo.Fingerprint == "" {
		logger.Info().Str("resource", "rtx_sshd_host_key").Msg("No host key exists, generating new SSHD host key")

		err = apiClient.client.GenerateSSHDHostKey(ctx)
		if err != nil {
			return diag.Errorf("Failed to generate SSHD host key: %v", err)
		}

		// Get the new key info
		keyInfo, err = apiClient.client.GetSSHDHostKey(ctx)
		if err != nil {
			return diag.Errorf("Failed to get newly generated SSHD host key info: %v", err)
		}
	}

	// Use fixed ID for singleton resource
	d.SetId("sshd_host_key")

	// Set state
	if err := d.Set("fingerprint", keyInfo.Fingerprint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("algorithm", keyInfo.Algorithm); err != nil {
		return diag.FromErr(err)
	}

	logger.Debug().
		Str("resource", "rtx_sshd_host_key").
		Str("fingerprint", keyInfo.Fingerprint).
		Str("algorithm", keyInfo.Algorithm).
		Msg("SSHD host key resource created")

	return nil
}

func resourceRTXSSHDHostKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", d.Id())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Reading SSHD host key configuration")

	// Get current host key info
	keyInfo, err := apiClient.client.GetSSHDHostKey(ctx)
	if err != nil {
		return diag.Errorf("Failed to get SSHD host key info: %v", err)
	}

	// If fingerprint is empty, the key doesn't exist - remove from state
	if keyInfo.Fingerprint == "" {
		logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("SSHD host key not found, removing from state")
		d.SetId("")
		return nil
	}

	// Update state
	if err := d.Set("fingerprint", keyInfo.Fingerprint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("algorithm", keyInfo.Algorithm); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXSSHDHostKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", d.Id())
	logger := logging.FromContext(ctx)

	// No-op - host keys should persist on the router
	// Deleting from Terraform state doesn't delete the actual key
	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Removing SSHD host key from Terraform state (key persists on router)")

	return nil
}

func resourceRTXSSHDHostKeyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_sshd_host_key").Msg("Importing SSHD host key configuration")

	// Set the ID
	d.SetId("sshd_host_key")

	// Verify host key exists and get info
	keyInfo, err := apiClient.client.GetSSHDHostKey(ctx)
	if err != nil {
		return nil, err
	}

	// If no host key exists, return error
	if keyInfo.Fingerprint == "" {
		return nil, nil
	}

	// Set attributes
	d.Set("fingerprint", keyInfo.Fingerprint)
	d.Set("algorithm", keyInfo.Algorithm)

	return []*schema.ResourceData{d}, nil
}
