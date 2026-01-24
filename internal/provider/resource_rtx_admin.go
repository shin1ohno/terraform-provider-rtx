package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

func resourceRTXAdmin() *schema.Resource {
	return &schema.Resource{
		Description: "Manages admin password configuration on RTX routers. This is a singleton resource - only one instance can exist per router. " +
			"Note: Changing passwords requires the provider's admin_password to be set to the current password.",
		CreateContext: resourceRTXAdminCreate,
		ReadContext:   resourceRTXAdminRead,
		UpdateContext: resourceRTXAdminUpdate,
		DeleteContext: resourceRTXAdminDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAdminImport,
		},

		Schema: map[string]*schema.Schema{
			"login_password": WriteOnlyStringSchema("Login password for the RTX router. This password is used for initial authentication when connecting to the router"),
			"admin_password": WriteOnlyStringSchema("Administrator password for the RTX router. This password is required for entering administrator mode to make configuration changes"),
			"last_updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of the last password update performed by Terraform (RFC3339 format).",
			},
		},
	}
}

func resourceRTXAdminCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	config := buildAdminConfigFromResourceData(d)

	// Use fixed ID for singleton resource
	d.SetId("admin")

	// Only configure if passwords are provided and different from current
	if config.AdminPassword != "" || config.LoginPassword != "" {
		logger.Debug().Str("resource", "rtx_admin").Msg("Creating admin configuration")

		err := apiClient.client.ConfigureAdmin(ctx, config)
		if err != nil {
			return diag.Errorf("Failed to configure admin: %v", err)
		}

		// Record the timestamp of successful password update
		if err := d.Set("last_updated", time.Now().Format(time.RFC3339)); err != nil {
			return diag.Errorf("Failed to set last_updated: %v", err)
		}
	} else {
		logger.Debug().Str("resource", "rtx_admin").Msg("Creating admin configuration (no password changes)")
	}

	return nil
}

func resourceRTXAdminRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	// Passwords cannot be read back from the router for security reasons
	// The resource exists if it was created, and passwords remain in state
	// We simply verify the ID is still set
	if d.Id() == "" {
		return nil
	}

	logger.Debug().Str("resource", "rtx_admin").Msg("Reading admin configuration (passwords cannot be read from router)")

	// Try to use SFTP cache if enabled to verify admin config exists
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract admin config from parsed config
			parsedAdmin := parsedConfig.ExtractAdmin()
			if parsedAdmin != nil {
				// Admin config exists in cache - passwords cannot be read for security
				logger.Debug().Str("resource", "rtx_admin").Msg("Found admin config in SFTP cache (passwords not readable)")
			}
		}
	}

	// Resource still exists - passwords remain as stored in state
	return nil
}

func resourceRTXAdminUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	// Check if passwords have changed
	if d.HasChange("login_password") || d.HasChange("admin_password") {
		config := buildAdminConfigFromResourceData(d)

		logger.Debug().Str("resource", "rtx_admin").Msg("Updating admin configuration")

		err := apiClient.client.ConfigureAdmin(ctx, config)
		if err != nil {
			return diag.Errorf("Failed to update admin configuration: %v", err)
		}

		// Record the timestamp of successful password update
		if err := d.Set("last_updated", time.Now().Format(time.RFC3339)); err != nil {
			return diag.Errorf("Failed to set last_updated: %v", err)
		}
	}

	return nil
}

func resourceRTXAdminDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Note: RTX password removal also requires interactive commands.
	// To remove passwords on RTX, use the console:
	//   administrator password    (then press Enter without entering a password)
	//   login password            (then press Enter without entering a password)
	//
	// This delete only removes the state record, not the actual router passwords.

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin").Msg("Deleting admin configuration (state record only - RTX password commands are interactive)")

	return nil
}

func resourceRTXAdminImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// For import, we just set the ID
	// Passwords must be provided in the Terraform configuration after import
	d.SetId("admin")

	logging.FromContext(ctx).Info().Str("resource", "rtx_admin").Msg("Admin configuration imported. Note: Passwords must be set in configuration as they cannot be read from the router.")

	return []*schema.ResourceData{d}, nil
}

func buildAdminConfigFromResourceData(d *schema.ResourceData) client.AdminConfig {
	return client.AdminConfig{
		LoginPassword: d.Get("login_password").(string),
		AdminPassword: d.Get("admin_password").(string),
	}
}
