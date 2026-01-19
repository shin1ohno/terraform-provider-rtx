package client

import (
	"context"
	"strings"
	"testing"
)

// mockExecutor is a simple mock executor for testing AdminService
type mockExecutor struct {
	responses    map[string]string
	executedCmds []string
}

func (m *mockExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.executedCmds = append(m.executedCmds, cmd)

	for pattern, response := range m.responses {
		if strings.Contains(cmd, pattern) {
			return []byte(response), nil
		}
	}
	return []byte{}, nil
}

func TestAdminService_GetAdminConfig(t *testing.T) {
	executor := &mockExecutor{
		responses: map[string]string{},
	}

	service := NewAdminService(executor, nil)

	config, err := service.GetAdminConfig(context.Background())
	if err != nil {
		t.Fatalf("GetAdminConfig() error = %v", err)
	}

	// Passwords cannot be read back from router
	if config.LoginPassword != "" {
		t.Errorf("GetAdminConfig() LoginPassword should be empty, got %s", config.LoginPassword)
	}
	if config.AdminPassword != "" {
		t.Errorf("GetAdminConfig() AdminPassword should be empty, got %s", config.AdminPassword)
	}
}

func TestAdminService_ConfigureAdmin(t *testing.T) {
	tests := []struct {
		name    string
		config  AdminConfig
		wantErr bool
	}{
		{
			name: "set both passwords",
			config: AdminConfig{
				LoginPassword: "loginpass",
				AdminPassword: "adminpass",
			},
			wantErr: false,
		},
		{
			name: "set only login password",
			config: AdminConfig{
				LoginPassword: "loginonly",
			},
			wantErr: false,
		},
		{
			name: "set only admin password",
			config: AdminConfig{
				AdminPassword: "adminonly",
			},
			wantErr: false,
		},
		{
			name:    "empty config",
			config:  AdminConfig{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockExecutor{
				responses: map[string]string{
					"login password":         "",
					"administrator password": "",
					"save":                   "",
				},
			}

			service := NewAdminService(executor, nil)

			err := service.ConfigureAdmin(context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigureAdmin() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify commands were executed
			if tt.config.LoginPassword != "" {
				found := false
				for _, cmd := range executor.executedCmds {
					if strings.Contains(cmd, "login password") {
						found = true
						break
					}
				}
				if !found {
					t.Error("ConfigureAdmin() should have executed login password command")
				}
			}

			if tt.config.AdminPassword != "" {
				found := false
				for _, cmd := range executor.executedCmds {
					if strings.Contains(cmd, "administrator password") {
						found = true
						break
					}
				}
				if !found {
					t.Error("ConfigureAdmin() should have executed administrator password command")
				}
			}
		})
	}
}

func TestAdminService_GetAdminUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		response string
		wantErr  bool
	}{
		{
			name:     "existing user",
			username: "admin",
			response: `login user admin password123
user attribute admin administrator=on connection=ssh,telnet`,
			wantErr: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			response: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockExecutor{
				responses: map[string]string{
					"show config | grep": tt.response,
				},
			}

			service := NewAdminService(executor, nil)

			user, err := service.GetAdminUser(context.Background(), tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAdminUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && user != nil {
				if user.Username != tt.username {
					t.Errorf("GetAdminUser() username = %v, want %v", user.Username, tt.username)
				}
			}
		})
	}
}

func TestAdminService_CreateAdminUser(t *testing.T) {
	tests := []struct {
		name    string
		user    AdminUser
		wantErr bool
	}{
		{
			name: "create basic user",
			user: AdminUser{
				Username:  "testuser",
				Password:  "testpass",
				Encrypted: false,
				Attributes: AdminUserAttributes{
					Administrator: false,
					Connection:    []string{},
					GUIPages:      []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "create admin user with attributes",
			user: AdminUser{
				Username:  "admin",
				Password:  "adminpass",
				Encrypted: false,
				Attributes: AdminUserAttributes{
					Administrator: true,
					Connection:    []string{"ssh", "telnet"},
					GUIPages:      []string{"dashboard", "config"},
					LoginTimer:    300,
				},
			},
			wantErr: false,
		},
		{
			name: "create user with encrypted password",
			user: AdminUser{
				Username:  "secureuser",
				Password:  "$1$encrypted",
				Encrypted: true,
				Attributes: AdminUserAttributes{
					Administrator: false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid username",
			user: AdminUser{
				Username: "",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "invalid password",
			user: AdminUser{
				Username: "user",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockExecutor{
				responses: map[string]string{
					"login user":     "",
					"user attribute": "",
					"save":           "",
				},
			}

			service := NewAdminService(executor, nil)

			err := service.CreateAdminUser(context.Background(), tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAdminUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdminService_UpdateAdminUser(t *testing.T) {
	executor := &mockExecutor{
		responses: map[string]string{
			"login user":     "",
			"user attribute": "",
			"save":           "",
		},
	}

	service := NewAdminService(executor, nil)

	user := AdminUser{
		Username:  "admin",
		Password:  "newpassword",
		Encrypted: false,
		Attributes: AdminUserAttributes{
			Administrator: true,
			Connection:    []string{"ssh"},
		},
	}

	err := service.UpdateAdminUser(context.Background(), user)
	if err != nil {
		t.Errorf("UpdateAdminUser() error = %v", err)
	}

	// Verify login user command was executed
	found := false
	for _, cmd := range executor.executedCmds {
		if strings.Contains(cmd, "login user admin") {
			found = true
			break
		}
	}
	if !found {
		t.Error("UpdateAdminUser() should have executed login user command")
	}
}

func TestAdminService_DeleteAdminUser(t *testing.T) {
	executor := &mockExecutor{
		responses: map[string]string{
			"no login user":     "",
			"no user attribute": "",
			"save":              "",
		},
	}

	service := NewAdminService(executor, nil)

	err := service.DeleteAdminUser(context.Background(), "testuser")
	if err != nil {
		t.Errorf("DeleteAdminUser() error = %v", err)
	}

	// Verify delete commands were executed
	foundUserDelete := false
	foundAttrDelete := false
	for _, cmd := range executor.executedCmds {
		if strings.Contains(cmd, "no login user") {
			foundUserDelete = true
		}
		if strings.Contains(cmd, "no user attribute") {
			foundAttrDelete = true
		}
	}

	if !foundUserDelete {
		t.Error("DeleteAdminUser() should have executed no login user command")
	}
	if !foundAttrDelete {
		t.Error("DeleteAdminUser() should have executed no user attribute command")
	}
}

func TestAdminService_ListAdminUsers(t *testing.T) {
	executor := &mockExecutor{
		responses: map[string]string{
			"show config | grep": `login user admin password123
user attribute admin administrator=on connection=ssh
login user guest guestpass
user attribute guest administrator=off`,
		},
	}

	service := NewAdminService(executor, nil)

	users, err := service.ListAdminUsers(context.Background())
	if err != nil {
		t.Fatalf("ListAdminUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("ListAdminUsers() got %d users, want 2", len(users))
	}
}
