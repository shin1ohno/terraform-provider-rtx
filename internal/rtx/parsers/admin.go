package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// AdminConfig represents the admin configuration on an RTX router
type AdminConfig struct {
	LoginPassword string       `json:"login_password"`
	AdminPassword string       `json:"admin_password"`
	Users         []UserConfig `json:"users"`
}

// UserConfig represents a user account configuration
type UserConfig struct {
	Username   string         `json:"username"`
	Password   string         `json:"password"`
	Encrypted  bool           `json:"encrypted"`
	Attributes UserAttributes `json:"attributes"`
}

// UserAttributes represents user attribute configuration
type UserAttributes struct {
	Administrator *bool    `json:"administrator"`
	Connection    []string `json:"connection"`  // serial, telnet, remote, ssh, sftp, http
	GUIPages      []string `json:"gui_pages"`   // dashboard, lan-map, config
	LoginTimer    *int     `json:"login_timer"` // seconds (0 = infinite)
}

// AdminParser is the interface for parsing admin configuration
type AdminParser interface {
	ParseAdminConfig(raw string) (*AdminConfig, error)
	ParseUserConfig(raw string, username string) (*UserConfig, error)
}

// adminParser handles parsing of admin configuration output
type adminParser struct{}

// NewAdminParser creates a new admin parser
func NewAdminParser() AdminParser {
	return &adminParser{}
}

// ParseAdminConfig parses the output of "show config" to extract admin configuration
func (p *adminParser) ParseAdminConfig(raw string) (*AdminConfig, error) {
	config := &AdminConfig{
		Users: []UserConfig{},
	}

	lines := strings.Split(raw, "\n")

	// Patterns for matching admin configuration
	// Note: Passwords are not shown in show config output for security
	loginUserPattern := regexp.MustCompile(`^\s*login\s+user\s+(\S+)\s+(.+)$`)
	loginUserEncryptedPattern := regexp.MustCompile(`^\s*login\s+user\s+(\S+)\s+encrypted\s+(\S+)$`)
	userAttributePattern := regexp.MustCompile(`^\s*user\s+attribute\s+(\S+)\s+(.+)$`)

	// Track users for attribute merging
	userMap := make(map[string]*UserConfig)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse login user with encrypted password
		if matches := loginUserEncryptedPattern.FindStringSubmatch(line); len(matches) >= 3 {
			username := matches[1]
			password := matches[2]

			user := &UserConfig{
				Username:  username,
				Password:  password,
				Encrypted: true,
				Attributes: UserAttributes{
					Connection: []string{},
					GUIPages:   []string{},
				},
			}
			userMap[username] = user
			continue
		}

		// Parse login user with plaintext password
		if matches := loginUserPattern.FindStringSubmatch(line); len(matches) >= 3 {
			username := matches[1]
			password := matches[2]

			user := &UserConfig{
				Username:  username,
				Password:  password,
				Encrypted: false,
				Attributes: UserAttributes{
					Connection: []string{},
					GUIPages:   []string{},
				},
			}
			userMap[username] = user
			continue
		}

		// Parse user attributes
		if matches := userAttributePattern.FindStringSubmatch(line); len(matches) >= 3 {
			username := matches[1]
			attrStr := matches[2]

			// Parse attributes
			attrs := parseUserAttributeString(attrStr)

			if user, exists := userMap[username]; exists {
				user.Attributes = attrs
			} else {
				// Create user entry even if no login user line found
				userMap[username] = &UserConfig{
					Username:   username,
					Attributes: attrs,
				}
			}
			continue
		}
	}

	// Convert map to slice
	for _, user := range userMap {
		config.Users = append(config.Users, *user)
	}

	return config, nil
}

// ParseUserConfig parses the output of "show config" to extract a specific user configuration
func (p *adminParser) ParseUserConfig(raw string, username string) (*UserConfig, error) {
	config, err := p.ParseAdminConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, user := range config.Users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user %s not found", username)
}

// parseUserAttributeString parses the attribute string from "user attribute <username> <attrs>"
func parseUserAttributeString(attrStr string) UserAttributes {
	attrs := UserAttributes{
		Connection: []string{},
		GUIPages:   []string{},
	}

	// Track if administrator was explicitly set
	administratorFound := false

	// Parse key=value pairs
	parts := strings.Fields(attrStr)
	for _, part := range parts {
		if strings.HasPrefix(part, "administrator=") {
			administratorFound = true
			value := strings.TrimPrefix(part, "administrator=")
			// "on", "1", "2" all mean administrator is enabled
			// "off" means disabled
			// "2" = password-less elevation (some models)
			// "1" = password required for elevation (some models)
			isAdmin := value == "on" || value == "1" || value == "2"
			attrs.Administrator = &isAdmin
		} else if strings.HasPrefix(part, "connection=") {
			value := strings.TrimPrefix(part, "connection=")
			// Connection types are comma-separated
			if value != "" && value != "none" {
				attrs.Connection = strings.Split(value, ",")
			}
		} else if strings.HasPrefix(part, "gui-page=") {
			value := strings.TrimPrefix(part, "gui-page=")
			// GUI pages are comma-separated
			if value != "" && value != "none" {
				attrs.GUIPages = strings.Split(value, ",")
			}
		} else if strings.HasPrefix(part, "login-timer=") {
			value := strings.TrimPrefix(part, "login-timer=")
			if timer, err := strconv.Atoi(value); err == nil {
				attrs.LoginTimer = &timer
			}
		}
	}

	// If administrator was not explicitly set in the output, apply default value.
	// RTX routers don't output default values in "show config".
	// Default is "on" for RTX1210, RTX1220, RTX830, RTX5000, RTX3500, vRX series
	// Default is "1" for other models (which also means enabled)
	// So the default is always "enabled" (true)
	if !administratorFound {
		defaultAdmin := true
		attrs.Administrator = &defaultAdmin
	}

	return attrs
}

// BuildLoginPasswordCommand builds the command to set login password
func BuildLoginPasswordCommand(password string) string {
	return fmt.Sprintf("login password %s", password)
}

// BuildAdminPasswordCommand builds the command to set administrator password
func BuildAdminPasswordCommand(password string) string {
	return fmt.Sprintf("administrator password %s", password)
}

// BuildUserCommand builds the command to create/update a user
func BuildUserCommand(user UserConfig) string {
	if user.Encrypted {
		return fmt.Sprintf("login user %s encrypted %s", user.Username, user.Password)
	}
	return fmt.Sprintf("login user %s %s", user.Username, user.Password)
}

// BuildUserAttributeCommand builds the command to set user attributes
// Only includes attributes when the pointer is non-nil to preserve current router values
func BuildUserAttributeCommand(username string, attrs UserAttributes) string {
	var parts []string

	// Administrator flag - only include when explicitly set (non-nil)
	if attrs.Administrator != nil {
		if *attrs.Administrator {
			parts = append(parts, "administrator=on")
		} else {
			parts = append(parts, "administrator=off")
		}
	}
	// nil = don't include (preserve current router value)

	// Connection types
	if len(attrs.Connection) > 0 {
		parts = append(parts, fmt.Sprintf("connection=%s", strings.Join(attrs.Connection, ",")))
	}

	// GUI pages
	if len(attrs.GUIPages) > 0 {
		parts = append(parts, fmt.Sprintf("gui-page=%s", strings.Join(attrs.GUIPages, ",")))
	}

	// Login timer - only include when explicitly set (non-nil)
	if attrs.LoginTimer != nil {
		parts = append(parts, fmt.Sprintf("login-timer=%d", *attrs.LoginTimer))
	}
	// nil = don't include (preserve current router value)

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("user attribute %s %s", username, strings.Join(parts, " "))
}

// BuildDeleteUserCommand builds the command to delete a user
func BuildDeleteUserCommand(username string) string {
	return fmt.Sprintf("no login user %s", username)
}

// BuildDeleteUserAttributeCommand builds the command to delete user attributes
func BuildDeleteUserAttributeCommand(username string) string {
	return fmt.Sprintf("no user attribute %s", username)
}

// BuildShowLoginUserCommand builds the command to show login user configuration
// Note: RTX grep doesn't support \| (OR) operator, so we use a simpler pattern
func BuildShowLoginUserCommand(username string) string {
	return fmt.Sprintf("show config | grep \"%s\"", username)
}

// BuildShowAllUsersCommand builds the command to show all user configurations
// Note: RTX grep doesn't support \| (OR) operator, so we use "user" which matches both
func BuildShowAllUsersCommand() string {
	return "show config | grep \"user\""
}

// ValidateUserConfig validates the user configuration (requires password)
func ValidateUserConfig(user UserConfig) error {
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}

	if user.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Validate username format (alphanumeric and underscore only)
	validUsername := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validUsername.MatchString(user.Username) {
		return fmt.Errorf("username must start with a letter and contain only alphanumeric characters and underscores")
	}

	// Validate connection types
	validConnections := map[string]bool{
		"serial": true,
		"telnet": true,
		"remote": true,
		"ssh":    true,
		"sftp":   true,
		"http":   true,
	}
	for _, conn := range user.Attributes.Connection {
		if !validConnections[conn] {
			return fmt.Errorf("invalid connection type: %s", conn)
		}
	}

	// Validate GUI pages
	validGUIPages := map[string]bool{
		"dashboard": true,
		"lan-map":   true,
		"config":    true,
	}
	for _, page := range user.Attributes.GUIPages {
		if !validGUIPages[page] {
			return fmt.Errorf("invalid GUI page: %s", page)
		}
	}

	// Validate login timer
	if user.Attributes.LoginTimer != nil && *user.Attributes.LoginTimer < 0 {
		return fmt.Errorf("login timer cannot be negative")
	}

	return nil
}

// ValidateUserConfigForAttributeUpdate validates user configuration for attribute-only updates
// This is used when updating an imported user where the password is not known
func ValidateUserConfigForAttributeUpdate(user UserConfig) error {
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}

	// Validate username format (alphanumeric and underscore only)
	validUsername := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validUsername.MatchString(user.Username) {
		return fmt.Errorf("username must start with a letter and contain only alphanumeric characters and underscores")
	}

	// Validate connection types
	validConnections := map[string]bool{
		"serial": true,
		"telnet": true,
		"remote": true,
		"ssh":    true,
		"sftp":   true,
		"http":   true,
	}
	for _, conn := range user.Attributes.Connection {
		if !validConnections[conn] {
			return fmt.Errorf("invalid connection type: %s", conn)
		}
	}

	// Validate GUI pages
	validGUIPages := map[string]bool{
		"dashboard": true,
		"lan-map":   true,
		"config":    true,
	}
	for _, page := range user.Attributes.GUIPages {
		if !validGUIPages[page] {
			return fmt.Errorf("invalid GUI page: %s", page)
		}
	}

	// Validate login timer
	if user.Attributes.LoginTimer != nil && *user.Attributes.LoginTimer < 0 {
		return fmt.Errorf("login timer cannot be negative")
	}

	return nil
}
