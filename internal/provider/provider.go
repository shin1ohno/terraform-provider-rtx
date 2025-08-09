package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in documentation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("RTX_HOST", nil),
				Description: "The hostname or IP address of the RTX router. Can be set with RTX_HOST environment variable.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("RTX_USERNAME", nil),
				Description: "Username for RTX router authentication. Can be set with RTX_USERNAME environment variable.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("RTX_PASSWORD", nil),
				Description: "Password for RTX router authentication. Can be set with RTX_PASSWORD environment variable.",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     22,
				Description: "SSH port for RTX router connection. Defaults to 22.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "Connection timeout in seconds. Defaults to 30.",
			},
			"ssh_host_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RTX_SSH_HOST_KEY", nil),
				Description: "SSH host public key for verification (base64 encoded). If unset, uses known_hosts_file. Can be set with RTX_SSH_HOST_KEY environment variable.",
			},
			"known_hosts_file": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"ssh_host_key"},
				DefaultFunc:   schema.EnvDefaultFunc("RTX_KNOWN_HOSTS_FILE", "~/.ssh/known_hosts"),
				Description:   "Path to known_hosts file for SSH host key verification. Defaults to ~/.ssh/known_hosts. Can be set with RTX_KNOWN_HOSTS_FILE environment variable.",
			},
			"skip_host_key_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.EnvDefaultFunc("RTX_SKIP_HOST_KEY_CHECK", false),
				Description: "Skip SSH host key verification. WARNING: This is insecure and should only be used for testing. Can be set with RTX_SKIP_HOST_KEY_CHECK environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			// Resources will be added here
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rtx_system_info": dataSourceRTXSystemInfo(),
			"rtx_interfaces":  dataSourceRTXInterfaces(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type apiClient struct {
	client client.Client
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	port := d.Get("port").(int)
	timeout := d.Get("timeout").(int)
	sshHostKey := d.Get("ssh_host_key").(string)
	knownHostsFile := d.Get("known_hosts_file").(string)
	skipHostKeyCheck := d.Get("skip_host_key_check").(bool)

	var diags diag.Diagnostics

	// Expand ~ in known_hosts_file path
	if strings.HasPrefix(knownHostsFile, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			knownHostsFile = filepath.Join(home, knownHostsFile[2:])
		}
	}

	// Create client configuration
	config := &client.Config{
		Host:             host,
		Port:             port,
		Username:         username,
		Password:         password,
		Timeout:          timeout,
		HostKey:          sshHostKey,
		KnownHostsFile:   knownHostsFile,
		SkipHostKeyCheck: skipHostKeyCheck,
	}

	// Create SSH client with default options
	sshClient, err := client.NewClient(
		config,
		client.WithPromptDetector(client.NewDefaultPromptDetector()),
		client.WithRetryStrategy(client.NewExponentialBackoff()),
	)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create RTX client",
			Detail:   fmt.Sprintf("Failed to create SSH client: %v", err),
		})
		return nil, diags
	}

	// Test connection to RTX router
	if err := sshClient.Dial(ctx); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to connect to RTX router",
			Detail:   fmt.Sprintf("Failed to establish SSH connection to %s:%d: %v", host, port, err),
		})
		return nil, diags
	}

	// Test with a simple command
	testCmd := client.Command{
		Key:     "show environment",
		Payload: "show environment",
	}
	
	if _, err := sshClient.Run(ctx, testCmd); err != nil {
		// Close the connection if test fails
		sshClient.Close()
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "RTX router communication test failed",
			Detail:   fmt.Sprintf("Failed to execute test command: %v", err),
		})
		return nil, diags
	}

	c := &apiClient{
		client: sshClient,
	}

	// Success - no additional diagnostics needed

	return c, diags
}