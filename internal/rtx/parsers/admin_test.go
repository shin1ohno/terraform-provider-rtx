package parsers

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseAdminConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *AdminConfig
		wantErr  bool
	}{
		{
			name: "single user with plaintext password",
			input: `
login user admin password123
user attribute admin administrator=on connection=ssh,telnet
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "admin",
						Password:  "password123",
						Encrypted: false,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{"ssh", "telnet"},
							GUIPages:      []string{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "user with encrypted password",
			input: `
login user admin encrypted $1$abcdef123456
user attribute admin administrator=on
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "admin",
						Password:  "$1$abcdef123456",
						Encrypted: true,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{},
							GUIPages:      []string{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple users",
			input: `
login user admin password123
user attribute admin administrator=on connection=ssh,telnet,http gui-page=dashboard,config
login user guest guestpass
user attribute guest administrator=off connection=http gui-page=dashboard login-timer=300
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "admin",
						Password:  "password123",
						Encrypted: false,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{"ssh", "telnet", "http"},
							GUIPages:      []string{"dashboard", "config"},
						},
					},
					{
						Username:  "guest",
						Password:  "guestpass",
						Encrypted: false,
						Attributes: UserAttributes{
							Administrator: false,
							Connection:    []string{"http"},
							GUIPages:      []string{"dashboard"},
							LoginTimer:    300,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "user with all connection types",
			input: `
login user operator operpass
user attribute operator administrator=off connection=serial,telnet,remote,ssh,sftp,http
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "operator",
						Password:  "operpass",
						Encrypted: false,
						Attributes: UserAttributes{
							Administrator: false,
							Connection:    []string{"serial", "telnet", "remote", "ssh", "sftp", "http"},
							GUIPages:      []string{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "empty config",
			input: "",
			expected: &AdminConfig{
				Users: []UserConfig{},
			},
			wantErr: false,
		},
		{
			name: "user with explicit login-timer=3600",
			input: `
login user netadmin encrypted $1$secure123
user attribute netadmin administrator=on connection=ssh,telnet login-timer=3600
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "netadmin",
						Password:  "$1$secure123",
						Encrypted: true,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{"ssh", "telnet"},
							GUIPages:      []string{},
							LoginTimer:    3600,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "user with login-timer at different positions",
			input: `
login user testuser testpass
user attribute testuser login-timer=7200 administrator=on connection=http
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "testuser",
						Password:  "testpass",
						Encrypted: false,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{"http"},
							GUIPages:      []string{},
							LoginTimer:    7200,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "user with all gui-pages",
			input: `
login user guiuser guipass
user attribute guiuser administrator=on gui-page=dashboard,lan-map,config
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "guiuser",
						Password:  "guipass",
						Encrypted: false,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{},
							GUIPages:      []string{"dashboard", "lan-map", "config"},
							LoginTimer:    0,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "user with full attributes for REQ-5 import fidelity",
			input: `
login user admin encrypted $1$hashpass
user attribute admin administrator=on connection=ssh,telnet,http gui-page=dashboard,lan-map,config login-timer=3600
`,
			expected: &AdminConfig{
				Users: []UserConfig{
					{
						Username:  "admin",
						Password:  "$1$hashpass",
						Encrypted: true,
						Attributes: UserAttributes{
							Administrator: true,
							Connection:    []string{"ssh", "telnet", "http"},
							GUIPages:      []string{"dashboard", "lan-map", "config"},
							LoginTimer:    3600,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	parser := NewAdminParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseAdminConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAdminConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check number of users
				if len(result.Users) != len(tt.expected.Users) {
					t.Errorf("ParseAdminConfig() got %d users, want %d", len(result.Users), len(tt.expected.Users))
					return
				}

				// Create map for easier comparison
				expectedMap := make(map[string]UserConfig)
				for _, u := range tt.expected.Users {
					expectedMap[u.Username] = u
				}

				for _, user := range result.Users {
					expected, exists := expectedMap[user.Username]
					if !exists {
						t.Errorf("ParseAdminConfig() unexpected user %s", user.Username)
						continue
					}

					if user.Username != expected.Username {
						t.Errorf("ParseAdminConfig() username = %v, want %v", user.Username, expected.Username)
					}
					if user.Password != expected.Password {
						t.Errorf("ParseAdminConfig() password = %v, want %v", user.Password, expected.Password)
					}
					if user.Encrypted != expected.Encrypted {
						t.Errorf("ParseAdminConfig() encrypted = %v, want %v", user.Encrypted, expected.Encrypted)
					}
					if user.Attributes.Administrator != expected.Attributes.Administrator {
						t.Errorf("ParseAdminConfig() administrator = %v, want %v", user.Attributes.Administrator, expected.Attributes.Administrator)
					}
					if !reflect.DeepEqual(user.Attributes.Connection, expected.Attributes.Connection) {
						t.Errorf("ParseAdminConfig() connection = %v, want %v", user.Attributes.Connection, expected.Attributes.Connection)
					}
					if !reflect.DeepEqual(user.Attributes.GUIPages, expected.Attributes.GUIPages) {
						t.Errorf("ParseAdminConfig() guiPages = %v, want %v", user.Attributes.GUIPages, expected.Attributes.GUIPages)
					}
					if user.Attributes.LoginTimer != expected.Attributes.LoginTimer {
						t.Errorf("ParseAdminConfig() loginTimer = %v, want %v", user.Attributes.LoginTimer, expected.Attributes.LoginTimer)
					}
				}
			}
		})
	}
}

func TestParseUserConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		username string
		expected *UserConfig
		wantErr  bool
	}{
		{
			name: "find existing user",
			input: `
login user admin password123
user attribute admin administrator=on connection=ssh
login user guest guestpass
user attribute guest administrator=off
`,
			username: "admin",
			expected: &UserConfig{
				Username:  "admin",
				Password:  "password123",
				Encrypted: false,
				Attributes: UserAttributes{
					Administrator: true,
					Connection:    []string{"ssh"},
					GUIPages:      []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "user not found",
			input: `
login user admin password123
`,
			username: "nonexistent",
			expected: nil,
			wantErr:  true,
		},
	}

	parser := NewAdminParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseUserConfig(tt.input, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUserConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				if result.Username != tt.expected.Username {
					t.Errorf("ParseUserConfig() username = %v, want %v", result.Username, tt.expected.Username)
				}
				if result.Password != tt.expected.Password {
					t.Errorf("ParseUserConfig() password = %v, want %v", result.Password, tt.expected.Password)
				}
			}
		})
	}
}

func TestBuildLoginPasswordCommand(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{
			name:     "simple password",
			password: "mypassword",
			expected: "login password mypassword",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!123",
			expected: "login password P@ssw0rd!123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildLoginPasswordCommand(tt.password)
			if result != tt.expected {
				t.Errorf("BuildLoginPasswordCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildAdminPasswordCommand(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{
			name:     "simple password",
			password: "adminpass",
			expected: "administrator password adminpass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAdminPasswordCommand(tt.password)
			if result != tt.expected {
				t.Errorf("BuildAdminPasswordCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildUserCommand(t *testing.T) {
	tests := []struct {
		name     string
		user     UserConfig
		expected string
	}{
		{
			name: "plaintext password",
			user: UserConfig{
				Username:  "admin",
				Password:  "password123",
				Encrypted: false,
			},
			expected: "login user admin password123",
		},
		{
			name: "encrypted password",
			user: UserConfig{
				Username:  "admin",
				Password:  "$1$abcdef123456",
				Encrypted: true,
			},
			expected: "login user admin encrypted $1$abcdef123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUserCommand(tt.user)
			if result != tt.expected {
				t.Errorf("BuildUserCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildUserAttributeCommand(t *testing.T) {
	tests := []struct {
		name     string
		username string
		attrs    UserAttributes
		expected string
	}{
		{
			name:     "administrator on",
			username: "admin",
			attrs: UserAttributes{
				Administrator: true,
			},
			expected: "user attribute admin administrator=on",
		},
		{
			name:     "administrator off with connections",
			username: "guest",
			attrs: UserAttributes{
				Administrator: false,
				Connection:    []string{"ssh", "telnet"},
			},
			expected: "user attribute guest administrator=off connection=ssh,telnet",
		},
		{
			name:     "full attributes",
			username: "operator",
			attrs: UserAttributes{
				Administrator: true,
				Connection:    []string{"ssh", "http"},
				GUIPages:      []string{"dashboard", "config"},
				LoginTimer:    600,
			},
			expected: "user attribute operator administrator=on connection=ssh,http gui-page=dashboard,config login-timer=600",
		},
		{
			name:     "empty attributes",
			username: "user",
			attrs: UserAttributes{
				Administrator: false,
			},
			expected: "user attribute user administrator=off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUserAttributeCommand(tt.username, tt.attrs)
			if result != tt.expected {
				t.Errorf("BuildUserAttributeCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteUserCommand(t *testing.T) {
	expected := "no login user testuser"
	result := BuildDeleteUserCommand("testuser")
	if result != expected {
		t.Errorf("BuildDeleteUserCommand() = %v, want %v", result, expected)
	}
}

func TestBuildDeleteUserAttributeCommand(t *testing.T) {
	expected := "no user attribute testuser"
	result := BuildDeleteUserAttributeCommand("testuser")
	if result != expected {
		t.Errorf("BuildDeleteUserAttributeCommand() = %v, want %v", result, expected)
	}
}

func TestValidateUserConfig(t *testing.T) {
	tests := []struct {
		name    string
		user    UserConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: UserConfig{
				Username: "admin",
				Password: "password123",
				Attributes: UserAttributes{
					Administrator: true,
					Connection:    []string{"ssh", "telnet"},
					GUIPages:      []string{"dashboard"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty username",
			user: UserConfig{
				Username: "",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username is required",
		},
		{
			name: "empty password",
			user: UserConfig{
				Username: "admin",
				Password: "",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "invalid username format - starts with number",
			user: UserConfig{
				Username: "1admin",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username must start with a letter",
		},
		{
			name: "invalid username format - special characters",
			user: UserConfig{
				Username: "admin@user",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username must start with a letter",
		},
		{
			name: "invalid connection type",
			user: UserConfig{
				Username: "admin",
				Password: "password123",
				Attributes: UserAttributes{
					Connection: []string{"invalid"},
				},
			},
			wantErr: true,
			errMsg:  "invalid connection type",
		},
		{
			name: "invalid GUI page",
			user: UserConfig{
				Username: "admin",
				Password: "password123",
				Attributes: UserAttributes{
					GUIPages: []string{"invalid-page"},
				},
			},
			wantErr: true,
			errMsg:  "invalid GUI page",
		},
		{
			name: "negative login timer",
			user: UserConfig{
				Username: "admin",
				Password: "password123",
				Attributes: UserAttributes{
					LoginTimer: -1,
				},
			},
			wantErr: true,
			errMsg:  "login timer cannot be negative",
		},
		{
			name: "valid username with underscore",
			user: UserConfig{
				Username: "admin_user",
				Password: "password123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserConfig(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateUserConfig() error = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestParseUserAttributeString(t *testing.T) {
	tests := []struct {
		name     string
		attrStr  string
		expected UserAttributes
	}{
		{
			name:    "login-timer only",
			attrStr: "login-timer=3600",
			expected: UserAttributes{
				Administrator: false,
				Connection:    []string{},
				GUIPages:      []string{},
				LoginTimer:    3600,
			},
		},
		{
			name:    "gui-page only",
			attrStr: "gui-page=dashboard,lan-map,config",
			expected: UserAttributes{
				Administrator: false,
				Connection:    []string{},
				GUIPages:      []string{"dashboard", "lan-map", "config"},
				LoginTimer:    0,
			},
		},
		{
			name:    "all attributes",
			attrStr: "administrator=on connection=ssh,telnet,http gui-page=dashboard,lan-map,config login-timer=3600",
			expected: UserAttributes{
				Administrator: true,
				Connection:    []string{"ssh", "telnet", "http"},
				GUIPages:      []string{"dashboard", "lan-map", "config"},
				LoginTimer:    3600,
			},
		},
		{
			name:    "attributes in different order",
			attrStr: "login-timer=7200 gui-page=config administrator=off connection=http",
			expected: UserAttributes{
				Administrator: false,
				Connection:    []string{"http"},
				GUIPages:      []string{"config"},
				LoginTimer:    7200,
			},
		},
		{
			name:    "connection none",
			attrStr: "administrator=on connection=none",
			expected: UserAttributes{
				Administrator: true,
				Connection:    []string{},
				GUIPages:      []string{},
				LoginTimer:    0,
			},
		},
		{
			name:    "gui-page none",
			attrStr: "administrator=on gui-page=none",
			expected: UserAttributes{
				Administrator: true,
				Connection:    []string{},
				GUIPages:      []string{},
				LoginTimer:    0,
			},
		},
		{
			name:    "login-timer zero",
			attrStr: "administrator=on login-timer=0",
			expected: UserAttributes{
				Administrator: true,
				Connection:    []string{},
				GUIPages:      []string{},
				LoginTimer:    0,
			},
		},
		// REQ-5: Verify import fidelity for admin user attributes
		{
			name:    "REQ-5 import fidelity - shin1ohno example",
			attrStr: "on administrator=off gui-page=dashboard,lan-map,config login-timer=3600",
			expected: UserAttributes{
				Administrator: false,
				Connection:    []string{},
				GUIPages:      []string{"dashboard", "lan-map", "config"},
				LoginTimer:    3600,
			},
		},
		{
			name:    "REQ-5 hyphen-separated keys only",
			attrStr: "administrator=on login-timer=1800 gui-page=dashboard",
			expected: UserAttributes{
				Administrator: true,
				Connection:    []string{},
				GUIPages:      []string{"dashboard"},
				LoginTimer:    1800,
			},
		},
		{
			name:    "REQ-5 large login-timer value",
			attrStr: "administrator=off login-timer=86400 connection=ssh",
			expected: UserAttributes{
				Administrator: false,
				Connection:    []string{"ssh"},
				GUIPages:      []string{},
				LoginTimer:    86400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseUserAttributeString(tt.attrStr)
			if result.Administrator != tt.expected.Administrator {
				t.Errorf("parseUserAttributeString() Administrator = %v, want %v", result.Administrator, tt.expected.Administrator)
			}
			if !reflect.DeepEqual(result.Connection, tt.expected.Connection) {
				t.Errorf("parseUserAttributeString() Connection = %v, want %v", result.Connection, tt.expected.Connection)
			}
			if !reflect.DeepEqual(result.GUIPages, tt.expected.GUIPages) {
				t.Errorf("parseUserAttributeString() GUIPages = %v, want %v", result.GUIPages, tt.expected.GUIPages)
			}
			if result.LoginTimer != tt.expected.LoginTimer {
				t.Errorf("parseUserAttributeString() LoginTimer = %v, want %v", result.LoginTimer, tt.expected.LoginTimer)
			}
		})
	}
}
