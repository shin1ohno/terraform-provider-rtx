package parsers

import (
	"testing"
)

func TestPPTPParser_ParsePPTPConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *PPTPConfig
	}{
		{
			name: "basic PPTP config",
			input: `pptp service on
pptp tunnel disconnect time 300`,
			expected: &PPTPConfig{
				Enabled:        true,
				DisconnectTime: 300,
			},
		},
		{
			name: "PPTP with keepalive",
			input: `pptp service on
pptp keepalive use on`,
			expected: &PPTPConfig{
				Enabled:          true,
				KeepaliveEnabled: true,
			},
		},
		{
			name: "PPTP with authentication",
			input: `pptp service on
pp auth accept mschap-v2
pp auth myname vpnuser vpnpass123`,
			expected: &PPTPConfig{
				Enabled: true,
				Authentication: &PPTPAuth{
					Method:   "mschap-v2",
					Username: "vpnuser",
					Password: "vpnpass123",
				},
			},
		},
		{
			name: "PPTP with MPPE encryption",
			input: `pptp service on
ppp ccp type mppe-128`,
			expected: &PPTPConfig{
				Enabled: true,
				Encryption: &PPTPEncryption{
					MPPEBits: 128,
				},
			},
		},
		{
			name: "PPTP with MPPE require",
			input: `pptp service on
ppp ccp type mppe-128 require`,
			expected: &PPTPConfig{
				Enabled: true,
				Encryption: &PPTPEncryption{
					MPPEBits: 128,
					Required: true,
				},
			},
		},
		{
			name: "PPTP with IP pool",
			input: `pptp service on
ip pp remote address pool 192.168.1.100-192.168.1.200`,
			expected: &PPTPConfig{
				Enabled: true,
				IPPool: &PPTPIPPool{
					Start: "192.168.1.100",
					End:   "192.168.1.200",
				},
			},
		},
		{
			name:  "PPTP disabled",
			input: `pptp service off`,
			expected: &PPTPConfig{
				Enabled: false,
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: &PPTPConfig{
				Enabled: false,
			},
		},
	}

	parser := NewPPTPParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParsePPTPConfig(tt.input)
			if err != nil {
				t.Errorf("ParsePPTPConfig() error = %v", err)
				return
			}
			if got.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", got.Enabled, tt.expected.Enabled)
			}
			if got.DisconnectTime != tt.expected.DisconnectTime {
				t.Errorf("DisconnectTime = %v, want %v", got.DisconnectTime, tt.expected.DisconnectTime)
			}
			if got.KeepaliveEnabled != tt.expected.KeepaliveEnabled {
				t.Errorf("KeepaliveEnabled = %v, want %v", got.KeepaliveEnabled, tt.expected.KeepaliveEnabled)
			}
			if tt.expected.Authentication != nil {
				if got.Authentication == nil {
					t.Error("Authentication is nil, expected non-nil")
				} else {
					if got.Authentication.Method != tt.expected.Authentication.Method {
						t.Errorf("Authentication.Method = %v, want %v", got.Authentication.Method, tt.expected.Authentication.Method)
					}
				}
			}
			if tt.expected.Encryption != nil {
				if got.Encryption == nil {
					t.Error("Encryption is nil, expected non-nil")
				} else {
					if got.Encryption.MPPEBits != tt.expected.Encryption.MPPEBits {
						t.Errorf("Encryption.MPPEBits = %v, want %v", got.Encryption.MPPEBits, tt.expected.Encryption.MPPEBits)
					}
					if got.Encryption.Required != tt.expected.Encryption.Required {
						t.Errorf("Encryption.Required = %v, want %v", got.Encryption.Required, tt.expected.Encryption.Required)
					}
				}
			}
			if tt.expected.IPPool != nil {
				if got.IPPool == nil {
					t.Error("IPPool is nil, expected non-nil")
				} else {
					if got.IPPool.Start != tt.expected.IPPool.Start {
						t.Errorf("IPPool.Start = %v, want %v", got.IPPool.Start, tt.expected.IPPool.Start)
					}
					if got.IPPool.End != tt.expected.IPPool.End {
						t.Errorf("IPPool.End = %v, want %v", got.IPPool.End, tt.expected.IPPool.End)
					}
				}
			}
		})
	}
}

func TestPPTPParser_ParseFullConfig(t *testing.T) {
	input := `pptp service on
pptp tunnel disconnect time 300
pptp keepalive use on
pp auth accept mschap-v2
pp auth myname vpnuser vpnpass123
ppp ccp type mppe-128 require
ip pp remote address pool 192.168.1.100-192.168.1.200`

	parser := NewPPTPParser()
	config, err := parser.ParsePPTPConfig(input)
	if err != nil {
		t.Fatalf("ParsePPTPConfig() error = %v", err)
	}

	if !config.Enabled {
		t.Errorf("Enabled = %v, want true", config.Enabled)
	}
	if config.DisconnectTime != 300 {
		t.Errorf("DisconnectTime = %v, want 300", config.DisconnectTime)
	}
	if !config.KeepaliveEnabled {
		t.Errorf("KeepaliveEnabled = %v, want true", config.KeepaliveEnabled)
	}
	if config.Authentication == nil {
		t.Fatal("Authentication is nil")
	}
	if config.Authentication.Method != "mschap-v2" {
		t.Errorf("Authentication.Method = %v, want mschap-v2", config.Authentication.Method)
	}
	if config.Authentication.Username != "vpnuser" {
		t.Errorf("Authentication.Username = %v, want vpnuser", config.Authentication.Username)
	}
	if config.Encryption == nil {
		t.Fatal("Encryption is nil")
	}
	if config.Encryption.MPPEBits != 128 {
		t.Errorf("Encryption.MPPEBits = %v, want 128", config.Encryption.MPPEBits)
	}
	if !config.Encryption.Required {
		t.Errorf("Encryption.Required = %v, want true", config.Encryption.Required)
	}
	if config.IPPool == nil {
		t.Fatal("IPPool is nil")
	}
	if config.IPPool.Start != "192.168.1.100" {
		t.Errorf("IPPool.Start = %v, want 192.168.1.100", config.IPPool.Start)
	}
	if config.IPPool.End != "192.168.1.200" {
		t.Errorf("IPPool.End = %v, want 192.168.1.200", config.IPPool.End)
	}
}

func TestBuildPPTPCommands(t *testing.T) {
	t.Run("BuildPPTPServiceCommand", func(t *testing.T) {
		if got := BuildPPTPServiceCommand(true); got != "pptp service on" {
			t.Errorf("BuildPPTPServiceCommand(true) = %v, want 'pptp service on'", got)
		}
		if got := BuildPPTPServiceCommand(false); got != "pptp service off" {
			t.Errorf("BuildPPTPServiceCommand(false) = %v, want 'pptp service off'", got)
		}
	})

	t.Run("BuildPPTPTunnelDisconnectTimeCommand", func(t *testing.T) {
		expected := "pptp tunnel disconnect time 300"
		if got := BuildPPTPTunnelDisconnectTimeCommand(300); got != expected {
			t.Errorf("BuildPPTPTunnelDisconnectTimeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPTPKeepaliveCommand", func(t *testing.T) {
		if got := BuildPPTPKeepaliveCommand(true); got != "pptp keepalive use on" {
			t.Errorf("BuildPPTPKeepaliveCommand(true) = %v, want 'pptp keepalive use on'", got)
		}
		if got := BuildPPTPKeepaliveCommand(false); got != "pptp keepalive use off" {
			t.Errorf("BuildPPTPKeepaliveCommand(false) = %v, want 'pptp keepalive use off'", got)
		}
	})

	t.Run("BuildPPTPAuthAcceptCommand", func(t *testing.T) {
		expected := "pp auth accept mschap-v2"
		if got := BuildPPTPAuthAcceptCommand("mschap-v2"); got != expected {
			t.Errorf("BuildPPTPAuthAcceptCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPTPAuthMynameCommand", func(t *testing.T) {
		expected := "pp auth myname user pass123"
		if got := BuildPPTPAuthMynameCommand("user", "pass123"); got != expected {
			t.Errorf("BuildPPTPAuthMynameCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPPCCPTypeCommand MPPE-128", func(t *testing.T) {
		enc := PPTPEncryption{MPPEBits: 128}
		expected := "ppp ccp type mppe-128"
		if got := BuildPPPCCPTypeCommand(enc); got != expected {
			t.Errorf("BuildPPPCCPTypeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPPCCPTypeCommand MPPE-128 require", func(t *testing.T) {
		enc := PPTPEncryption{MPPEBits: 128, Required: true}
		expected := "ppp ccp type mppe-128 require"
		if got := BuildPPPCCPTypeCommand(enc); got != expected {
			t.Errorf("BuildPPPCCPTypeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPPCCPTypeCommand MPPE-40", func(t *testing.T) {
		enc := PPTPEncryption{MPPEBits: 40}
		expected := "ppp ccp type mppe-40"
		if got := BuildPPPCCPTypeCommand(enc); got != expected {
			t.Errorf("BuildPPPCCPTypeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPPCCPTypeCommand MPPE-56", func(t *testing.T) {
		enc := PPTPEncryption{MPPEBits: 56}
		expected := "ppp ccp type mppe-56"
		if got := BuildPPPCCPTypeCommand(enc); got != expected {
			t.Errorf("BuildPPPCCPTypeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPPCCPTypeCommand MPPE-any", func(t *testing.T) {
		enc := PPTPEncryption{MPPEBits: 0}
		expected := "ppp ccp type mppe-any"
		if got := BuildPPPCCPTypeCommand(enc); got != expected {
			t.Errorf("BuildPPPCCPTypeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPTPIPPoolCommand", func(t *testing.T) {
		expected := "ip pp remote address pool 192.168.1.100-192.168.1.200"
		if got := BuildPPTPIPPoolCommand("192.168.1.100", "192.168.1.200"); got != expected {
			t.Errorf("BuildPPTPIPPoolCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildDeletePPTPCommand", func(t *testing.T) {
		cmds := BuildDeletePPTPCommand()
		if len(cmds) != 3 {
			t.Errorf("BuildDeletePPTPCommand() returned %d commands, want 3", len(cmds))
		}
		if cmds[0] != "pptp service off" {
			t.Errorf("First command = %v, want 'pptp service off'", cmds[0])
		}
	})
}

func TestValidatePPTPConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  PPTPConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: PPTPConfig{
				Enabled: true,
				Authentication: &PPTPAuth{
					Method:   "mschap-v2",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid auth method",
			config: PPTPConfig{
				Authentication: &PPTPAuth{
					Method: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "valid auth method pap",
			config: PPTPConfig{
				Authentication: &PPTPAuth{
					Method: "pap",
				},
			},
			wantErr: false,
		},
		{
			name: "valid auth method chap",
			config: PPTPConfig{
				Authentication: &PPTPAuth{
					Method: "chap",
				},
			},
			wantErr: false,
		},
		{
			name: "valid auth method mschap",
			config: PPTPConfig{
				Authentication: &PPTPAuth{
					Method: "mschap",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid MPPE bits",
			config: PPTPConfig{
				Encryption: &PPTPEncryption{
					MPPEBits: 64, // Invalid
				},
			},
			wantErr: true,
		},
		{
			name: "valid MPPE 40 bits",
			config: PPTPConfig{
				Encryption: &PPTPEncryption{
					MPPEBits: 40,
				},
			},
			wantErr: false,
		},
		{
			name: "valid MPPE 56 bits",
			config: PPTPConfig{
				Encryption: &PPTPEncryption{
					MPPEBits: 56,
				},
			},
			wantErr: false,
		},
		{
			name: "valid MPPE 128 bits",
			config: PPTPConfig{
				Encryption: &PPTPEncryption{
					MPPEBits: 128,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid IP pool start",
			config: PPTPConfig{
				IPPool: &PPTPIPPool{
					Start: "invalid",
					End:   "192.168.1.200",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid IP pool end",
			config: PPTPConfig{
				IPPool: &PPTPIPPool{
					Start: "192.168.1.100",
					End:   "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid listen address",
			config: PPTPConfig{
				ListenAddress: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid listen address 0.0.0.0",
			config: PPTPConfig{
				ListenAddress: "0.0.0.0",
			},
			wantErr: false,
		},
		{
			name: "negative disconnect time",
			config: PPTPConfig{
				DisconnectTime: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePPTPConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePPTPConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
