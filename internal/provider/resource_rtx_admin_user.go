package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func resourceRTXAdminUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages admin user accounts on RTX routers. Each user can have different permissions and access methods.",
		CreateContext: resourceRTXAdminUserCreate,
		ReadContext:   resourceRTXAdminUserRead,
		UpdateContext: resourceRTXAdminUserUpdate,
		DeleteContext: resourceRTXAdminUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAdminUserImport,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Username for the admin user. Must start with a letter and contain only alphanumeric characters and underscores.",
				ValidateFunc: validateUsername,
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for the admin user. Required for create, optional for import.",
			},
			"encrypted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the password is already encrypted. If true, the password value will be used as-is (for encrypted passwords).",
			},
			"administrator": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the user has administrator privileges.",
			},
			"connection_methods": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Allowed connection methods for the user.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"serial", "telnet", "remote", "ssh", "sftp", "http",
					}, false),
				},
			},
			"gui_pages": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Allowed GUI pages for the user.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"dashboard", "lan-map", "config",
					}, false),
				},
			},
			"login_timer": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Login timeout in seconds. 0 means infinite (no timeout). If not specified, the router's default is used.",
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
	}
}

// validateUsername validates the username format
func validateUsername(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, []error{fmt.Errorf("%q cannot be empty", k)}
	}

	// Must start with a letter
	if len(value) > 0 && !((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) {
		return nil, []error{fmt.Errorf("%q must start with a letter", k)}
	}

	// Must contain only alphanumeric characters and underscores
	for _, c := range value {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return nil, []error{fmt.Errorf("%q must contain only alphanumeric characters and underscores", k)}
		}
	}

	return nil, nil
}

func resourceRTXAdminUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	user := buildAdminUserFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin_user").Msgf("Creating admin user: %s", user.Username)

	err := apiClient.client.CreateAdminUser(ctx, user)
	if err != nil {
		return diag.Errorf("Failed to create admin user: %v", err)
	}

	// Use username as the resource ID
	d.SetId(user.Username)

	return resourceRTXAdminUserRead(ctx, d, meta)
}

func resourceRTXAdminUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	logger := logging.FromContext(ctx)

	username := d.Id()

	logger.Debug().Str("resource", "rtx_admin_user").Msgf("Reading admin user: %s", username)

	var user *client.AdminUser

	// Try to use SFTP cache if enabled
	if apiClient.client.SFTPEnabled() {
		parsedConfig, err := apiClient.client.GetCachedConfig(ctx)
		if err == nil && parsedConfig != nil {
			// Extract admin users from parsed config
			users := parsedConfig.ExtractAdminUsers()
			for i := range users {
				if users[i].Username == username {
					user = convertParsedAdminUser(&users[i])
					logger.Debug().Str("resource", "rtx_admin_user").Msg("Found admin user in SFTP cache")
					break
				}
			}
		}
		if user == nil {
			// User not found in cache or cache error, fallback to SSH
			logger.Debug().Str("resource", "rtx_admin_user").Msg("Admin user not in cache, falling back to SSH")
		}
	}

	// Fallback to SSH if SFTP disabled or user not found in cache
	if user == nil {
		var err error
		user, err = apiClient.client.GetAdminUser(ctx, username)
		if err != nil {
			// Check if user doesn't exist
			if strings.Contains(err.Error(), "not found") {
				logger.Debug().Str("resource", "rtx_admin_user").Msgf("Admin user %s not found, removing from state", username)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to read admin user: %v", err)
		}
	}

	// Update the state with values from the router
	// Note: Password cannot be read back from the router
	if err := d.Set("username", user.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("encrypted", user.Encrypted); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("administrator", user.Attributes.Administrator); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("connection_methods", user.Attributes.Connection); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gui_pages", user.Attributes.GUIPages); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("login_timer", user.Attributes.LoginTimer); err != nil {
		return diag.FromErr(err)
	}

	// Password is not set from read - it remains as stored in state
	// This is because passwords are not displayed by the router for security

	return nil
}

func resourceRTXAdminUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	user := buildAdminUserFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin_user").Msgf("Updating admin user: %s", user.Username)

	err := apiClient.client.UpdateAdminUser(ctx, user)
	if err != nil {
		return diag.Errorf("Failed to update admin user: %v", err)
	}

	return resourceRTXAdminUserRead(ctx, d, meta)
}

func resourceRTXAdminUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	username := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin_user").Msgf("Deleting admin user: %s", username)

	err := apiClient.client.DeleteAdminUser(ctx, username)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete admin user: %v", err)
	}

	return nil
}

func resourceRTXAdminUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	username := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_admin_user").Msgf("Importing admin user: %s", username)

	// Verify user exists
	user, err := apiClient.client.GetAdminUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to import admin user %s: %v", username, err)
	}

	// Set all attributes
	d.SetId(username)
	d.Set("username", user.Username)
	d.Set("encrypted", user.Encrypted)
	d.Set("administrator", user.Attributes.Administrator)
	d.Set("connection_methods", user.Attributes.Connection)
	d.Set("gui_pages", user.Attributes.GUIPages)
	d.Set("login_timer", user.Attributes.LoginTimer)

	// Note: Password must be provided in the Terraform configuration after import
	logging.FromContext(ctx).Info().Str("resource", "rtx_admin_user").Msgf("Admin user %s imported. Note: Password must be set in configuration as it cannot be read from the router.", username)

	return []*schema.ResourceData{d}, nil
}

// buildAdminUserFromResourceData creates an AdminUser from Terraform resource data
func buildAdminUserFromResourceData(d *schema.ResourceData) client.AdminUser {
	user := client.AdminUser{
		Username:  d.Get("username").(string),
		Password:  d.Get("password").(string),
		Encrypted: d.Get("encrypted").(bool),
		Attributes: client.AdminUserAttributes{
			Administrator: d.Get("administrator").(bool),
			LoginTimer:    d.Get("login_timer").(int),
		},
	}

	// Handle connection set
	if v, ok := d.GetOk("connection_methods"); ok {
		connectionSet := v.(*schema.Set)
		connections := make([]string, connectionSet.Len())
		for i, conn := range connectionSet.List() {
			connections[i] = conn.(string)
		}
		user.Attributes.Connection = connections
	} else {
		user.Attributes.Connection = []string{}
	}

	// Handle gui_pages set
	if v, ok := d.GetOk("gui_pages"); ok {
		guiPagesSet := v.(*schema.Set)
		guiPages := make([]string, guiPagesSet.Len())
		for i, page := range guiPagesSet.List() {
			guiPages[i] = page.(string)
		}
		user.Attributes.GUIPages = guiPages
	} else {
		user.Attributes.GUIPages = []string{}
	}

	return user
}

// convertParsedAdminUser converts a parser UserConfig to a client AdminUser
func convertParsedAdminUser(parsed *parsers.UserConfig) *client.AdminUser {
	user := &client.AdminUser{
		Username:  parsed.Username,
		Password:  parsed.Password,
		Encrypted: parsed.Encrypted,
		Attributes: client.AdminUserAttributes{
			Administrator: parsed.Attributes.Administrator,
			Connection:    parsed.Attributes.Connection,
			GUIPages:      parsed.Attributes.GUIPages,
			LoginTimer:    parsed.Attributes.LoginTimer,
		},
	}

	// Ensure slices are not nil
	if user.Attributes.Connection == nil {
		user.Attributes.Connection = []string{}
	}
	if user.Attributes.GUIPages == nil {
		user.Attributes.GUIPages = []string{}
	}

	return user
}
