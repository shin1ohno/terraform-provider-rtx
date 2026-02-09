package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// AdminService handles admin configuration operations
type AdminService struct {
	executor Executor
	client   *rtxClient
}

// NewAdminService creates a new admin service instance
func NewAdminService(executor Executor, client *rtxClient) *AdminService {
	return &AdminService{
		executor: executor,
		client:   client,
	}
}

// GetAdminConfig retrieves admin password configuration
// Note: Passwords cannot be read back from the router for security reasons
// This returns an empty config since passwords are not shown in "show config"
func (s *AdminService) GetAdminConfig(ctx context.Context) (*AdminConfig, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Note: RTX routers do not show passwords in "show config" output
	// We return an empty config to indicate the resource exists
	// The actual password values must be stored in Terraform state
	return &AdminConfig{
		LoginPassword: "", // Cannot be read from router
		AdminPassword: "", // Cannot be read from router
	}, nil
}

// ConfigureAdmin sets admin password configuration using interactive prompts
// The old password is taken from the provider configuration (client.config.AdminPassword)
// The new password is taken from the config parameter
func (s *AdminService) ConfigureAdmin(ctx context.Context, config AdminConfig) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set login password if provided
	if config.LoginPassword != "" {
		logging.FromContext(ctx).Debug().Str("service", "admin").Msg("Setting login password via interactive prompt")

		err := s.executor.SetLoginPassword(ctx, config.LoginPassword)
		if err != nil {
			return fmt.Errorf("failed to set login password: %w", err)
		}
	}

	// Set admin password if provided
	if config.AdminPassword != "" {
		// Get the current (old) password from the client config
		var oldPassword string
		if s.client != nil {
			oldPassword = s.client.config.AdminPassword
		}

		logging.FromContext(ctx).Debug().Str("service", "admin").Msg("Setting administrator password via interactive prompt")

		err := s.executor.SetAdministratorPassword(ctx, oldPassword, config.AdminPassword)
		if err != nil {
			return fmt.Errorf("failed to set administrator password: %w", err)
		}
	}

	return saveConfig(ctx, s.client, "admin config set")
}

// UpdateAdminConfig updates admin password configuration
func (s *AdminService) UpdateAdminConfig(ctx context.Context, config AdminConfig) error {
	// For passwords, update is the same as configure
	return s.ConfigureAdmin(ctx, config)
}

// ResetAdmin removes admin password configuration
func (s *AdminService) ResetAdmin(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Remove login password
	cmd := "no login password"
	logging.FromContext(ctx).Debug().Str("service", "admin").Msg("Removing login password")

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove login password: %w", err)
	}
	if err := checkOutputErrorIgnoringNotFound(output, "failed to remove login password"); err != nil {
		return err
	}

	// Remove admin password
	cmd = "no administrator password"
	logging.FromContext(ctx).Debug().Str("service", "admin").Msg("Removing administrator password")

	output, err = s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove administrator password: %w", err)
	}
	if err := checkOutputErrorIgnoringNotFound(output, "failed to remove administrator password"); err != nil {
		return err
	}

	return saveConfig(ctx, s.client, "admin config removed")
}

// GetAdminUser retrieves an admin user configuration
func (s *AdminService) GetAdminUser(ctx context.Context, username string) (*AdminUser, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowLoginUserCommand(username)
	logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(cmd)).Msg("Getting admin user")

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "admin").Str("output", string(output)).Msg("Admin user raw output")

	parser := parsers.NewAdminParser()
	userConfig, err := parser.ParseUserConfig(string(output), username)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin user: %w", err)
	}

	// Convert parsers.UserConfig to client.AdminUser
	user := s.fromParserUser(*userConfig)
	return &user, nil
}

// CreateAdminUser creates a new admin user
func (s *AdminService) CreateAdminUser(ctx context.Context, user AdminUser) error {
	// Convert client.AdminUser to parsers.UserConfig
	parserUser := s.toParserUser(user)

	// Validate input
	if err := parsers.ValidateUserConfig(parserUser); err != nil {
		return fmt.Errorf("invalid user config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute user creation command
	cmd := parsers.BuildUserCommand(parserUser)
	logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(cmd)).Msg("Creating admin user")

	if err := runCommand(ctx, s.executor, cmd); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Set user attributes if any are specified
	if (user.Attributes.Administrator != nil && *user.Attributes.Administrator) || len(user.Attributes.Connection) > 0 ||
		len(user.Attributes.GUIPages) > 0 || (user.Attributes.LoginTimer != nil && *user.Attributes.LoginTimer > 0) {
		attrCmd := parsers.BuildUserAttributeCommand(user.Username, parserUser.Attributes)
		if attrCmd != "" {
			logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(attrCmd)).Msg("Setting user attributes")

			if err := runCommand(ctx, s.executor, attrCmd); err != nil {
				return fmt.Errorf("failed to set user attributes: %w", err)
			}
		}
	}

	return saveConfig(ctx, s.client, "user created")
}

// UpdateAdminUser updates an existing admin user
// If password is empty, only attributes will be updated (useful for imported resources)
func (s *AdminService) UpdateAdminUser(ctx context.Context, user AdminUser) error {
	// Convert client.AdminUser to parsers.UserConfig
	parserUser := s.toParserUser(user)

	// Validate input - use different validation depending on whether password is set
	if user.Password != "" {
		if err := parsers.ValidateUserConfig(parserUser); err != nil {
			return fmt.Errorf("invalid user config: %w", err)
		}
	} else {
		// For attribute-only updates, only validate username
		if err := parsers.ValidateUserConfigForAttributeUpdate(parserUser); err != nil {
			return fmt.Errorf("invalid user config: %w", err)
		}
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Only update user password if provided (otherwise just update attributes)
	if user.Password != "" {
		cmd := parsers.BuildUserCommand(parserUser)
		logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(cmd)).Msg("Updating admin user")

		if err := runCommand(ctx, s.executor, cmd); err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}
	}

	// Update user attributes
	attrCmd := parsers.BuildUserAttributeCommand(user.Username, parserUser.Attributes)
	if attrCmd != "" {
		logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(attrCmd)).Msg("Updating user attributes")

		if err := runCommand(ctx, s.executor, attrCmd); err != nil {
			return fmt.Errorf("failed to update user attributes: %w", err)
		}
	}

	return saveConfig(ctx, s.client, "user updated")
}

// DeleteAdminUser removes an admin user
func (s *AdminService) DeleteAdminUser(ctx context.Context, username string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Delete user attributes first
	attrCmd := parsers.BuildDeleteUserAttributeCommand(username)
	logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(attrCmd)).Msg("Deleting user attributes")

	_, _ = s.executor.Run(ctx, attrCmd) // Ignore errors for cleanup

	// Delete user
	cmd := parsers.BuildDeleteUserCommand(username)
	logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(cmd)).Msg("Deleting admin user")

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete admin user: %w", err)
	}

	if err := checkOutputErrorIgnoringNotFound(output, "failed to delete admin user"); err != nil {
		return err
	}

	return saveConfig(ctx, s.client, "user deleted")
}

// ListAdminUsers retrieves all admin users
func (s *AdminService) ListAdminUsers(ctx context.Context) ([]AdminUser, error) {
	cmd := parsers.BuildShowAllUsersCommand()
	logging.FromContext(ctx).Debug().Str("service", "admin").Str("command", SanitizeCommandForLog(cmd)).Msg("Listing admin users")

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list admin users: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "admin").Str("output", string(output)).Msg("Admin users raw output")

	parser := parsers.NewAdminParser()
	config, err := parser.ParseAdminConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin users: %w", err)
	}

	// Convert parsers.UserConfig to client.AdminUser
	users := make([]AdminUser, len(config.Users))
	for i, pu := range config.Users {
		users[i] = s.fromParserUser(pu)
	}

	return users, nil
}

// toParserUser converts client.AdminUser to parsers.UserConfig
func (s *AdminService) toParserUser(user AdminUser) parsers.UserConfig {
	return parsers.UserConfig{
		Username:  user.Username,
		Password:  user.Password,
		Encrypted: user.Encrypted,
		Attributes: parsers.UserAttributes{
			Administrator: user.Attributes.Administrator,
			Connection:    user.Attributes.Connection,
			GUIPages:      user.Attributes.GUIPages,
			LoginTimer:    user.Attributes.LoginTimer,
		},
	}
}

// fromParserUser converts parsers.UserConfig to client.AdminUser
func (s *AdminService) fromParserUser(pu parsers.UserConfig) AdminUser {
	connection := pu.Attributes.Connection
	if connection == nil {
		connection = []string{}
	}

	guiPages := pu.Attributes.GUIPages
	if guiPages == nil {
		guiPages = []string{}
	}

	return AdminUser{
		Username:  pu.Username,
		Password:  pu.Password,
		Encrypted: pu.Encrypted,
		Attributes: AdminUserAttributes{
			Administrator: pu.Attributes.Administrator,
			Connection:    connection,
			GUIPages:      guiPages,
			LoginTimer:    pu.Attributes.LoginTimer,
		},
	}
}
