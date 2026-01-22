package provider

import (
	"context"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXAdmin() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages admin password configuration on RTX routers. This is a singleton resource - only one instance can exist per router.",
		CreateContext: resourceRTXAdminCreate,
		ReadContext:   resourceRTXAdminRead,
		UpdateContext: resourceRTXAdminUpdate,
		DeleteContext: resourceRTXAdminDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAdminImport,
		},

		Schema: map[string]*schema.Schema{
			"login_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Login password for the RTX router. This password is used for initial authentication when connecting to the router.",
			},
			"admin_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Administrator password for the RTX router. This password is required for entering administrator mode to make configuration changes.",
			},
		},
	}
}

func resourceRTXAdminCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildAdminConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin").Msg("Creating admin configuration")

	err := apiClient.client.ConfigureAdmin(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure admin: %v", err)
	}

	// Use fixed ID for singleton resource
	d.SetId("admin")

	// Store the passwords in state (they cannot be read back from router)
	// Note: Read will not update these values since passwords are not returned
	return nil
}

func resourceRTXAdminRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Passwords cannot be read back from the router for security reasons
	// The resource exists if it was created, and passwords remain in state
	// We simply verify the ID is still set
	if d.Id() == "" {
		return nil
	}

	// Resource still exists - passwords remain as stored in state
	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin").Msg("Reading admin configuration (passwords cannot be read from router)")
	return nil
}

func resourceRTXAdminUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildAdminConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin").Msg("Updating admin configuration")

	err := apiClient.client.UpdateAdminConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update admin configuration: %v", err)
	}

	return nil
}

func resourceRTXAdminDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin").Msg("Deleting admin configuration")

	err := apiClient.client.ResetAdmin(ctx)
	if err != nil {
		return diag.Errorf("Failed to reset admin configuration: %v", err)
	}

	return nil
}

func resourceRTXAdminImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// For import, we just set the ID
	// Passwords must be provided in the Terraform configuration after import
	d.SetId("admin")

	logging.FromContext(ctx).Info().Str("resource", "rtx_admin").Msg("Admin configuration imported. Note: Passwords must be set in configuration as they cannot be read from the router.")

	return []*schema.ResourceData{d}, nil
}

// buildAdminConfigFromResourceData creates an AdminConfig from Terraform resource data
func buildAdminConfigFromResourceData(d *schema.ResourceData) client.AdminConfig {
	config := client.AdminConfig{}

	if v, ok := d.GetOk("login_password"); ok {
		config.LoginPassword = v.(string)
	}

	if v, ok := d.GetOk("admin_password"); ok {
		config.AdminPassword = v.(string)
	}

	return config
}
