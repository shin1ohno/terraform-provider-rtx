package provider

import (
	"context"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXSFTPD() *schema.Resource {
	return &schema.Resource{
		Description: "Manages SFTP daemon (sftpd) configuration on RTX routers. " +
			"This is a singleton resource - only one instance should exist per router. " +
			"SFTPD requires SSHD to be enabled for the service to work.",
		CreateContext: resourceRTXSFTPDCreate,
		ReadContext:   resourceRTXSFTPDRead,
		UpdateContext: resourceRTXSFTPDUpdate,
		DeleteContext: resourceRTXSFTPDDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXSFTPDImport,
		},

		Schema: map[string]*schema.Schema{
			"hosts": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of interfaces to listen on. At least one interface must be specified.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringMatch(
						regexp.MustCompile(`^(lan\d+|pp\d+|bridge\d+|tunnel\d+)$`),
						"must be a valid interface name (e.g., lan1, pp1, bridge1, tunnel1)",
					),
				},
			},
		},
	}
}

func resourceRTXSFTPDCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSFTPDConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msgf("Creating SFTPD configuration: hosts=%v", config.Hosts)

	err := apiClient.client.ConfigureSFTPD(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure SFTPD: %v", err)
	}

	// Use fixed ID for singleton resource
	d.SetId("sftpd")

	// Read back to ensure consistency
	return resourceRTXSFTPDRead(ctx, d, meta)
}

func resourceRTXSFTPDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msg("Reading SFTPD configuration")

	config, err := apiClient.client.GetSFTPD(ctx)
	if err != nil {
		// Check if not configured
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not configured") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msg("SFTPD not configured, removing from state")
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read SFTPD configuration: %v", err)
	}

	// If no hosts are configured, the resource doesn't exist
	if len(config.Hosts) == 0 {
		logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msg("SFTPD hosts not configured, removing from state")
		d.SetId("")
		return nil
	}

	// Update the state
	if err := d.Set("hosts", config.Hosts); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRTXSFTPDUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSFTPDConfigFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msgf("Updating SFTPD configuration: hosts=%v", config.Hosts)

	err := apiClient.client.UpdateSFTPD(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update SFTPD configuration: %v", err)
	}

	return resourceRTXSFTPDRead(ctx, d, meta)
}

func resourceRTXSFTPDDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msg("Deleting SFTPD configuration")

	err := apiClient.client.ResetSFTPD(ctx)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to remove SFTPD configuration: %v", err)
	}

	return nil
}

func resourceRTXSFTPDImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_sftpd").Msg("Importing SFTPD configuration")

	// Verify SFTPD is configured
	config, err := apiClient.client.GetSFTPD(ctx)
	if err != nil {
		return nil, err
	}

	if len(config.Hosts) == 0 {
		return nil, nil // Not configured, nothing to import
	}

	// Set the ID
	d.SetId("sftpd")

	// Set attributes
	d.Set("hosts", config.Hosts)

	return []*schema.ResourceData{d}, nil
}

// buildSFTPDConfigFromResourceData creates an SFTPDConfig from Terraform resource data
func buildSFTPDConfigFromResourceData(d *schema.ResourceData) client.SFTPDConfig {
	config := client.SFTPDConfig{
		Hosts: []string{},
	}

	// Parse hosts list
	if v, ok := d.GetOk("hosts"); ok {
		hostsList := v.([]interface{})
		hosts := make([]string, len(hostsList))
		for i, h := range hostsList {
			hosts[i] = h.(string)
		}
		config.Hosts = hosts
	}

	return config
}
