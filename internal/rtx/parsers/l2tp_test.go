package parsers

import (
	"testing"
)

func TestL2TPParser_ParseL2TPConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected tunnel count
	}{
		{
			name: "L2TPv3 tunnel",
			input: `tunnel select 1
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 10.0.0.2
l2tp local router-id 1.1.1.1
l2tp remote router-id 2.2.2.2`,
			expected: 1,
		},
		{
			name: "L2TPv3 with keepalive",
			input: `tunnel select 1
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 10.0.0.2
l2tp keepalive use on 30 5`,
			expected: 1,
		},
		{
			name: "L2TPv3 with always-on",
			input: `tunnel select 1
tunnel encapsulation l2tpv3
l2tp always-on on`,
			expected: 1,
		},
		{
			name: "L2TPv3 with disconnect time",
			input: `tunnel select 1
tunnel encapsulation l2tpv3
l2tp tunnel disconnect time 300`,
			expected: 1,
		},
		{
			name: "multiple L2TPv3 tunnels",
			input: `tunnel select 1
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 10.0.0.2
tunnel select 2
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 10.0.0.3`,
			expected: 2,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
	}

	parser := NewL2TPParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseL2TPConfig(tt.input)
			if err != nil {
				t.Errorf("ParseL2TPConfig() error = %v", err)
				return
			}
			if len(got) != tt.expected {
				t.Errorf("Tunnel count = %v, want %v", len(got), tt.expected)
			}
		})
	}
}

func TestL2TPParser_ParseL2TPv3Details(t *testing.T) {
	input := `tunnel select 1
tunnel encapsulation l2tpv3
tunnel endpoint address 10.0.0.1 10.0.0.2
l2tp local router-id 1.1.1.1
l2tp remote router-id 2.2.2.2
l2tp remote end-id branch-office
l2tp always-on on
l2tp keepalive use on 30 5
l2tp tunnel disconnect time 300
description Site-B L2VPN`

	parser := NewL2TPParser()
	tunnels, err := parser.ParseL2TPConfig(input)
	if err != nil {
		t.Fatalf("ParseL2TPConfig() error = %v", err)
	}

	if len(tunnels) != 1 {
		t.Fatalf("Expected 1 tunnel, got %d", len(tunnels))
	}

	tunnel := tunnels[0]

	if tunnel.ID != 1 {
		t.Errorf("ID = %v, want 1", tunnel.ID)
	}
	if tunnel.Version != "l2tpv3" {
		t.Errorf("Version = %v, want l2tpv3", tunnel.Version)
	}
	if tunnel.Mode != "l2vpn" {
		t.Errorf("Mode = %v, want l2vpn", tunnel.Mode)
	}
	if tunnel.TunnelSource != "10.0.0.1" {
		t.Errorf("TunnelSource = %v, want 10.0.0.1", tunnel.TunnelSource)
	}
	if tunnel.TunnelDest != "10.0.0.2" {
		t.Errorf("TunnelDest = %v, want 10.0.0.2", tunnel.TunnelDest)
	}
	if tunnel.L2TPv3Config == nil {
		t.Fatal("L2TPv3Config is nil")
	}
	if tunnel.L2TPv3Config.LocalRouterID != "1.1.1.1" {
		t.Errorf("LocalRouterID = %v, want 1.1.1.1", tunnel.L2TPv3Config.LocalRouterID)
	}
	if tunnel.L2TPv3Config.RemoteRouterID != "2.2.2.2" {
		t.Errorf("RemoteRouterID = %v, want 2.2.2.2", tunnel.L2TPv3Config.RemoteRouterID)
	}
	if tunnel.L2TPv3Config.RemoteEndID != "branch-office" {
		t.Errorf("RemoteEndID = %v, want branch-office", tunnel.L2TPv3Config.RemoteEndID)
	}
	if !tunnel.AlwaysOn {
		t.Errorf("AlwaysOn = %v, want true", tunnel.AlwaysOn)
	}
	if !tunnel.KeepaliveEnabled {
		t.Errorf("KeepaliveEnabled = %v, want true", tunnel.KeepaliveEnabled)
	}
	if tunnel.KeepaliveConfig == nil {
		t.Fatal("KeepaliveConfig is nil")
	}
	if tunnel.KeepaliveConfig.Interval != 30 {
		t.Errorf("Keepalive Interval = %v, want 30", tunnel.KeepaliveConfig.Interval)
	}
	if tunnel.KeepaliveConfig.Retry != 5 {
		t.Errorf("Keepalive Retry = %v, want 5", tunnel.KeepaliveConfig.Retry)
	}
	if tunnel.DisconnectTime != 300 {
		t.Errorf("DisconnectTime = %v, want 300", tunnel.DisconnectTime)
	}
	if tunnel.Name != "Site-B L2VPN" {
		t.Errorf("Name = %v, want 'Site-B L2VPN'", tunnel.Name)
	}
}

func TestL2TPParser_TunnelAuth(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedEnabled  bool
		expectedPassword string
	}{
		{
			name: "tunnel auth on with password",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint address 10.0.0.1 10.0.0.2
 l2tp local router-id 1.1.1.1
 l2tp remote router-id 2.2.2.2
 l2tp tunnel auth on secret123
tunnel enable 1`,
			expectedEnabled:  true,
			expectedPassword: "secret123",
		},
		{
			name: "tunnel auth on without password",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
 l2tp tunnel auth on
tunnel enable 1`,
			expectedEnabled:  true,
			expectedPassword: "",
		},
		{
			name: "tunnel auth off",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
 l2tp tunnel auth off
tunnel enable 1`,
			expectedEnabled:  false,
			expectedPassword: "",
		},
		{
			name: "no tunnel auth line",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
 l2tp local router-id 1.1.1.1
tunnel enable 1`,
			expectedEnabled:  false,
			expectedPassword: "",
		},
	}

	parser := NewL2TPParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tunnels, err := parser.ParseL2TPConfig(tt.input)
			if err != nil {
				t.Fatalf("ParseL2TPConfig() error = %v", err)
			}
			if len(tunnels) != 1 {
				t.Fatalf("Expected 1 tunnel, got %d", len(tunnels))
			}

			tunnel := tunnels[0]
			if tunnel.L2TPv3Config == nil {
				if tt.expectedEnabled {
					t.Errorf("L2TPv3Config is nil, but expected tunnel auth enabled")
				}
				return
			}

			if tunnel.L2TPv3Config.TunnelAuth == nil {
				if tt.expectedEnabled {
					t.Errorf("TunnelAuth is nil, but expected enabled=%v", tt.expectedEnabled)
				}
				return
			}

			if tunnel.L2TPv3Config.TunnelAuth.Enabled != tt.expectedEnabled {
				t.Errorf("TunnelAuth.Enabled = %v, want %v", tunnel.L2TPv3Config.TunnelAuth.Enabled, tt.expectedEnabled)
			}
			if tunnel.L2TPv3Config.TunnelAuth.Password != tt.expectedPassword {
				t.Errorf("TunnelAuth.Password = %v, want %v", tunnel.L2TPv3Config.TunnelAuth.Password, tt.expectedPassword)
			}
		})
	}
}

func TestBuildL2TPCommands(t *testing.T) {
	t.Run("BuildL2TPServiceCommand", func(t *testing.T) {
		if got := BuildL2TPServiceCommand(true); got != "l2tp service on" {
			t.Errorf("BuildL2TPServiceCommand(true) = %v, want 'l2tp service on'", got)
		}
		if got := BuildL2TPServiceCommand(false); got != "l2tp service off" {
			t.Errorf("BuildL2TPServiceCommand(false) = %v, want 'l2tp service off'", got)
		}
	})

	t.Run("BuildPPSelectAnonymousCommand", func(t *testing.T) {
		expected := "pp select anonymous"
		if got := BuildPPSelectAnonymousCommand(); got != expected {
			t.Errorf("BuildPPSelectAnonymousCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPBindTunnelCommand", func(t *testing.T) {
		expected := "pp bind tunnel1"
		if got := BuildPPBindTunnelCommand(1); got != expected {
			t.Errorf("BuildPPBindTunnelCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPAuthAcceptCommand", func(t *testing.T) {
		expected := "pp auth accept chap"
		if got := BuildPPAuthAcceptCommand("chap"); got != expected {
			t.Errorf("BuildPPAuthAcceptCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildPPAuthMynameCommand", func(t *testing.T) {
		expected := "pp auth myname user pass123"
		if got := BuildPPAuthMynameCommand("user", "pass123"); got != expected {
			t.Errorf("BuildPPAuthMynameCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPPPRemotePoolCommand", func(t *testing.T) {
		expected := "ip pp remote address pool 192.168.1.100-192.168.1.200"
		if got := BuildIPPPRemotePoolCommand("192.168.1.100", "192.168.1.200"); got != expected {
			t.Errorf("BuildIPPPRemotePoolCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildTunnelEncapsulationCommand", func(t *testing.T) {
		expected := "tunnel encapsulation l2tpv3"
		if got := BuildTunnelEncapsulationCommand(1, "l2tpv3"); got != expected {
			t.Errorf("BuildTunnelEncapsulationCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildTunnelEndpointCommand", func(t *testing.T) {
		expected := "tunnel endpoint address 10.0.0.1 10.0.0.2"
		if got := BuildTunnelEndpointCommand("10.0.0.1", "10.0.0.2"); got != expected {
			t.Errorf("BuildTunnelEndpointCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildL2TPLocalRouterIDCommand", func(t *testing.T) {
		expected := "l2tp local router-id 1.1.1.1"
		if got := BuildL2TPLocalRouterIDCommand("1.1.1.1"); got != expected {
			t.Errorf("BuildL2TPLocalRouterIDCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildL2TPRemoteRouterIDCommand", func(t *testing.T) {
		expected := "l2tp remote router-id 2.2.2.2"
		if got := BuildL2TPRemoteRouterIDCommand("2.2.2.2"); got != expected {
			t.Errorf("BuildL2TPRemoteRouterIDCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildL2TPRemoteEndIDCommand", func(t *testing.T) {
		expected := "l2tp remote end-id branch"
		if got := BuildL2TPRemoteEndIDCommand("branch"); got != expected {
			t.Errorf("BuildL2TPRemoteEndIDCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildL2TPAlwaysOnCommand", func(t *testing.T) {
		if got := BuildL2TPAlwaysOnCommand(true); got != "l2tp always-on on" {
			t.Errorf("BuildL2TPAlwaysOnCommand(true) = %v, want 'l2tp always-on on'", got)
		}
		if got := BuildL2TPAlwaysOnCommand(false); got != "l2tp always-on off" {
			t.Errorf("BuildL2TPAlwaysOnCommand(false) = %v, want 'l2tp always-on off'", got)
		}
	})

	t.Run("BuildL2TPKeepaliveCommand", func(t *testing.T) {
		expected := "l2tp keepalive use on 30 5"
		if got := BuildL2TPKeepaliveCommand(30, 5); got != expected {
			t.Errorf("BuildL2TPKeepaliveCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildL2TPDisconnectTimeCommand", func(t *testing.T) {
		expected := "l2tp tunnel disconnect time 300"
		if got := BuildL2TPDisconnectTimeCommand(300); got != expected {
			t.Errorf("BuildL2TPDisconnectTimeCommand() = %v, want %v", got, expected)
		}
	})
}

func TestParseL2TPServiceConfig(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedEnabled   bool
		expectedProtocols []string
	}{
		{
			name:              "service on without protocols",
			input:             "l2tp service on",
			expectedEnabled:   true,
			expectedProtocols: []string{},
		},
		{
			name:              "service on with l2tpv3",
			input:             "l2tp service on l2tpv3",
			expectedEnabled:   true,
			expectedProtocols: []string{"l2tpv3"},
		},
		{
			name:              "service on with l2tp",
			input:             "l2tp service on l2tp",
			expectedEnabled:   true,
			expectedProtocols: []string{"l2tp"},
		},
		{
			name:              "service on with both protocols",
			input:             "l2tp service on l2tpv3 l2tp",
			expectedEnabled:   true,
			expectedProtocols: []string{"l2tpv3", "l2tp"},
		},
		{
			name:              "service off",
			input:             "l2tp service off",
			expectedEnabled:   false,
			expectedProtocols: []string{},
		},
		{
			name:              "empty input (default off)",
			input:             "",
			expectedEnabled:   false,
			expectedProtocols: []string{},
		},
		{
			name:              "no l2tp service line (default off)",
			input:             "tunnel select 1\ntunnel encapsulation l2tpv3",
			expectedEnabled:   false,
			expectedProtocols: []string{},
		},
		{
			name: "service on in multiline config",
			input: `# other config
l2tp service on l2tpv3 l2tp
tunnel select 1`,
			expectedEnabled:   true,
			expectedProtocols: []string{"l2tpv3", "l2tp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseL2TPServiceConfig(tt.input)
			if err != nil {
				t.Errorf("ParseL2TPServiceConfig() error = %v", err)
				return
			}
			if got.Enabled != tt.expectedEnabled {
				t.Errorf("Enabled = %v, want %v", got.Enabled, tt.expectedEnabled)
			}
			if len(got.Protocols) != len(tt.expectedProtocols) {
				t.Errorf("Protocols count = %v, want %v", len(got.Protocols), len(tt.expectedProtocols))
				return
			}
			for i, p := range got.Protocols {
				if p != tt.expectedProtocols[i] {
					t.Errorf("Protocol[%d] = %v, want %v", i, p, tt.expectedProtocols[i])
				}
			}
		})
	}
}

func TestBuildL2TPServiceCommandWithProtocols(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		protocols []string
		expected  string
	}{
		{
			name:      "service on without protocols",
			enabled:   true,
			protocols: nil,
			expected:  "l2tp service on",
		},
		{
			name:      "service on with empty protocols",
			enabled:   true,
			protocols: []string{},
			expected:  "l2tp service on",
		},
		{
			name:      "service on with l2tpv3",
			enabled:   true,
			protocols: []string{"l2tpv3"},
			expected:  "l2tp service on l2tpv3",
		},
		{
			name:      "service on with both protocols",
			enabled:   true,
			protocols: []string{"l2tpv3", "l2tp"},
			expected:  "l2tp service on l2tpv3 l2tp",
		},
		{
			name:      "service off (protocols ignored)",
			enabled:   false,
			protocols: []string{"l2tpv3", "l2tp"},
			expected:  "l2tp service off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildL2TPServiceCommandWithProtocols(tt.enabled, tt.protocols)
			if got != tt.expected {
				t.Errorf("BuildL2TPServiceCommandWithProtocols() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateL2TPConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  L2TPConfig
		wantErr bool
	}{
		{
			name: "valid L2TPv3 config",
			config: L2TPConfig{
				ID:           1,
				Version:      "l2tpv3",
				Mode:         "l2vpn",
				TunnelSource: "10.0.0.1",
				TunnelDest:   "10.0.0.2",
				L2TPv3Config: &L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
				},
			},
			wantErr: false,
		},
		{
			name: "valid L2TPv2 config",
			config: L2TPConfig{
				ID:      1,
				Version: "l2tp",
				Mode:    "lns",
				Authentication: &L2TPAuth{
					Method: "chap",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			config: L2TPConfig{
				ID:      1,
				Version: "invalid",
				Mode:    "lns",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			config: L2TPConfig{
				ID:      1,
				Version: "l2tp",
				Mode:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid auth method",
			config: L2TPConfig{
				ID:      1,
				Version: "l2tp",
				Mode:    "lns",
				Authentication: &L2TPAuth{
					Method: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid IP pool",
			config: L2TPConfig{
				ID:      1,
				Version: "l2tp",
				Mode:    "lns",
				IPPool: &L2TPIPPool{
					Start: "invalid",
					End:   "192.168.1.200",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tunnel source",
			config: L2TPConfig{
				ID:           1,
				Version:      "l2tpv3",
				Mode:         "l2vpn",
				TunnelSource: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid L2TPv3 router ID",
			config: L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				L2TPv3Config: &L2TPv3Config{
					LocalRouterID: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateL2TPConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateL2TPConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
