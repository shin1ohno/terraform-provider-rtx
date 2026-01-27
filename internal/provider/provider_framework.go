package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/provider/resources/admin_user"
	"github.com/sh1/terraform-provider-rtx/internal/provider/resources/ipsec_tunnel"
	"github.com/sh1/terraform-provider-rtx/internal/provider/resources/l2tp"
)

// Ensure RTXFrameworkProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &RTXFrameworkProvider{}
)

// RTXFrameworkProvider defines the provider implementation using Plugin Framework.
type RTXFrameworkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// RTXProviderModel describes the provider data model.
type RTXProviderModel struct {
	Host                 types.String `tfsdk:"host"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	PrivateKey           types.String `tfsdk:"private_key"`
	PrivateKeyFile       types.String `tfsdk:"private_key_file"`
	PrivateKeyPassphrase types.String `tfsdk:"private_key_passphrase"`
	AdminPassword        types.String `tfsdk:"admin_password"`
	Port                 types.Int64  `tfsdk:"port"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	SSHHostKey           types.String `tfsdk:"ssh_host_key"`
	KnownHostsFile       types.String `tfsdk:"known_hosts_file"`
	SkipHostKeyCheck     types.Bool   `tfsdk:"skip_host_key_check"`
	MaxParallelism       types.Int64  `tfsdk:"max_parallelism"`
	UseSFTP              types.Bool   `tfsdk:"use_sftp"`
	SFTPConfigPath       types.String `tfsdk:"sftp_config_path"`
	SSHSessionPool       types.List   `tfsdk:"ssh_session_pool"`
}

// SSHSessionPoolModel describes the SSH session pool configuration.
type SSHSessionPoolModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	MaxSessions types.Int64  `tfsdk:"max_sessions"`
	IdleTimeout types.String `tfsdk:"idle_timeout"`
}


// NewFramework creates a new Framework provider factory function.
func NewFramework(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RTXFrameworkProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *RTXFrameworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rtx"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *RTXFrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The RTX provider allows management of Yamaha RTX series routers.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The hostname or IP address of the RTX router. Can be set with RTX_HOST environment variable.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for RTX router authentication. Can be set with RTX_USERNAME environment variable.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for RTX router authentication. Can be set with RTX_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"private_key": schema.StringAttribute{
				Description: "SSH private key content (PEM format) for authentication. Can be set with RTX_PRIVATE_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"private_key_file": schema.StringAttribute{
				Description: "Path to SSH private key file for authentication. Can be set with RTX_PRIVATE_KEY_FILE environment variable.",
				Optional:    true,
			},
			"private_key_passphrase": schema.StringAttribute{
				Description: "Passphrase for encrypted private key. Can be set with RTX_PRIVATE_KEY_PASSPHRASE environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"admin_password": schema.StringAttribute{
				Description: "Administrator password for RTX router configuration changes. If not set, uses the same as password. Can be set with RTX_ADMIN_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"port": schema.Int64Attribute{
				Description: "SSH port for RTX router connection. Defaults to 22.",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Connection timeout in seconds. Defaults to 30.",
				Optional:    true,
			},
			"ssh_host_key": schema.StringAttribute{
				Description: "SSH host public key for verification (base64 encoded). If unset, uses known_hosts_file. Can be set with RTX_SSH_HOST_KEY environment variable.",
				Optional:    true,
			},
			"known_hosts_file": schema.StringAttribute{
				Description: "Path to known_hosts file for SSH host key verification. Defaults to ~/.ssh/known_hosts. Can be set with RTX_KNOWN_HOSTS_FILE environment variable.",
				Optional:    true,
			},
			"skip_host_key_check": schema.BoolAttribute{
				Description: "Skip SSH host key verification. WARNING: This is insecure and should only be used for testing. Can be set with RTX_SKIP_HOST_KEY_CHECK environment variable.",
				Optional:    true,
			},
			"max_parallelism": schema.Int64Attribute{
				Description: "Maximum number of concurrent operations. RTX routers support up to 8 simultaneous SSH connections, but lower values are more stable. Defaults to 4. Can be set with RTX_MAX_PARALLELISM environment variable.",
				Optional:    true,
			},
			"use_sftp": schema.BoolAttribute{
				Description: "Use SFTP-based configuration reading for faster bulk operations. Defaults to false. Can be set with RTX_USE_SFTP environment variable.",
				Optional:    true,
			},
			"sftp_config_path": schema.StringAttribute{
				Description: "SFTP path to the configuration file (e.g., /system/config0). If empty, the path will be auto-detected. Can be set with RTX_SFTP_CONFIG_PATH environment variable.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"ssh_session_pool": schema.ListNestedBlock{
				Description: "SSH session pool configuration for improved performance and state consistency.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Description: "Enable SSH session pooling. When enabled, SSH sessions are reused across operations, improving performance and preventing state drift. Defaults to true.",
							Optional:    true,
						},
						"max_sessions": schema.Int64Attribute{
							Description: "Maximum number of concurrent SSH sessions in the pool. RTX routers typically support up to 8 SSH connections. Defaults to 2.",
							Optional:    true,
						},
						"idle_timeout": schema.StringAttribute{
							Description: "Duration after which idle sessions are closed. Uses Go duration format (e.g., '5m', '30s', '1h'). Defaults to '5m'.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares an RTX API client for data sources and resources.
func (p *RTXFrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config RTXProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize logger and add to context
	logger := logging.NewLogger()
	ctx = logging.WithContext(ctx, logger)

	// Get values with environment variable fallbacks
	host := getStringValue(config.Host, "RTX_HOST", "")
	username := getStringValue(config.Username, "RTX_USERNAME", "")
	password := getStringValue(config.Password, "RTX_PASSWORD", "")
	privateKey := getStringValue(config.PrivateKey, "RTX_PRIVATE_KEY", "")
	privateKeyFile := getStringValue(config.PrivateKeyFile, "RTX_PRIVATE_KEY_FILE", "")
	privateKeyPassphrase := getStringValue(config.PrivateKeyPassphrase, "RTX_PRIVATE_KEY_PASSPHRASE", "")
	adminPassword := getStringValue(config.AdminPassword, "RTX_ADMIN_PASSWORD", "")
	sshHostKey := getStringValue(config.SSHHostKey, "RTX_SSH_HOST_KEY", "")
	knownHostsFile := getStringValue(config.KnownHostsFile, "RTX_KNOWN_HOSTS_FILE", "~/.ssh/known_hosts")
	sftpConfigPath := getStringValue(config.SFTPConfigPath, "RTX_SFTP_CONFIG_PATH", "")

	port := getInt64Value(config.Port, "RTX_PORT", 22)
	timeout := getInt64Value(config.Timeout, "RTX_TIMEOUT", 30)
	maxParallelism := getInt64Value(config.MaxParallelism, "RTX_MAX_PARALLELISM", 4)

	skipHostKeyCheck := getBoolValue(config.SkipHostKeyCheck, "RTX_SKIP_HOST_KEY_CHECK", false)
	useSFTP := getBoolValue(config.UseSFTP, "RTX_USE_SFTP", false)

	// Validate required fields
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing RTX Host",
			"The provider cannot create the RTX client as there is a missing or empty value for the RTX host. "+
				"Set the host value in the configuration or use the RTX_HOST environment variable.",
		)
	}
	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing RTX Username",
			"The provider cannot create the RTX client as there is a missing or empty value for the RTX username. "+
				"Set the username value in the configuration or use the RTX_USERNAME environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// SSH session pool configuration (defaults)
	sshPoolEnabled := true
	sshPoolMaxSessions := 2
	sshPoolIdleTimeout := "5m"

	// Read ssh_session_pool block if provided
	if !config.SSHSessionPool.IsNull() && !config.SSHSessionPool.IsUnknown() {
		var poolConfigs []SSHSessionPoolModel
		resp.Diagnostics.Append(config.SSHSessionPool.ElementsAs(ctx, &poolConfigs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(poolConfigs) > 0 {
			poolConfig := poolConfigs[0]
			if !poolConfig.Enabled.IsNull() && !poolConfig.Enabled.IsUnknown() {
				sshPoolEnabled = poolConfig.Enabled.ValueBool()
			}
			if !poolConfig.MaxSessions.IsNull() && !poolConfig.MaxSessions.IsUnknown() {
				sshPoolMaxSessions = int(poolConfig.MaxSessions.ValueInt64())
			}
			if !poolConfig.IdleTimeout.IsNull() && !poolConfig.IdleTimeout.IsUnknown() {
				sshPoolIdleTimeout = poolConfig.IdleTimeout.ValueString()
			}
		}
	}

	// If admin_password is not set, use the same as password
	if adminPassword == "" {
		adminPassword = password
	}

	// Expand ~ in known_hosts_file path
	if strings.HasPrefix(knownHostsFile, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			knownHostsFile = filepath.Join(home, knownHostsFile[2:])
		}
	}

	// Create client configuration
	clientConfig := &client.Config{
		Host:                 host,
		Port:                 int(port),
		Username:             username,
		Password:             password,
		PrivateKey:           privateKey,
		PrivateKeyFile:       privateKeyFile,
		PrivateKeyPassphrase: privateKeyPassphrase,
		AdminPassword:        adminPassword,
		Timeout:              int(timeout),
		HostKey:              sshHostKey,
		KnownHostsFile:       knownHostsFile,
		SkipHostKeyCheck:     skipHostKeyCheck,
		MaxParallelism:       int(maxParallelism),
		SFTPEnabled:          useSFTP,
		SFTPConfigPath:       sftpConfigPath,
		SSHPoolEnabled:       sshPoolEnabled,
		SSHPoolMaxSessions:   sshPoolMaxSessions,
		SSHPoolIdleTimeout:   sshPoolIdleTimeout,
	}

	// Create SSH client with default options
	sshClient, err := client.NewClient(
		clientConfig,
		client.WithPromptDetector(client.NewDefaultPromptDetector()),
		client.WithRetryStrategy(client.NewExponentialBackoff()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create RTX Client",
			fmt.Sprintf("Failed to create SSH client: %v", err),
		)
		return
	}

	// Test connection to RTX router
	if err := sshClient.Dial(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to RTX Router",
			fmt.Sprintf("Failed to establish SSH connection to %s:%d: %v", host, port, err),
		)
		return
	}

	// Test with a simple command
	testCmd := client.Command{
		Key:     "show environment",
		Payload: "show environment",
	}

	logger.Debug().Msg("Provider: Running test command")
	if _, err := sshClient.Run(ctx, testCmd); err != nil {
		// Close the connection if test fails
		sshClient.Close()
		resp.Diagnostics.AddError(
			"RTX Router Communication Test Failed",
			fmt.Sprintf("Failed to execute test command: %v", err),
		)
		return
	}
	logger.Debug().Msg("Provider: Test command successful")

	// Store provider data for resources and data sources
	providerData := &fwhelpers.ProviderData{
		Client: sshClient,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// Resources defines the resources implemented in the provider.
func (p *RTXFrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Phase 2: High-priority sensitive resources
		admin_user.NewAdminUserResource,
		ipsec_tunnel.NewIPsecTunnelResource,
		l2tp.NewL2TPResource,
		// Phase 3: Normal resources (will be added)
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *RTXFrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	// Data sources will be added as they are migrated
	return []func() datasource.DataSource{
		// Data sources (will be added)
	}
}

// Helper functions to get values with environment variable fallbacks

func getStringValue(attr types.String, envVar, defaultValue string) string {
	if !attr.IsNull() && !attr.IsUnknown() {
		return attr.ValueString()
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return defaultValue
}

func getInt64Value(attr types.Int64, envVar string, defaultValue int64) int64 {
	if !attr.IsNull() && !attr.IsUnknown() {
		return attr.ValueInt64()
	}
	if v := os.Getenv(envVar); v != "" {
		var i int64
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return defaultValue
}

func getBoolValue(attr types.Bool, envVar string, defaultValue bool) bool {
	if !attr.IsNull() && !attr.IsUnknown() {
		return attr.ValueBool()
	}
	if v := os.Getenv(envVar); v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return defaultValue
}
