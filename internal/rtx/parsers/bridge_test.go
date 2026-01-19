package parsers

import (
	"strings"
	"testing"
)

func TestParseBridgeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []BridgeConfig
		wantErr  bool
	}{
		{
			name:  "single bridge with one member",
			input: `bridge member bridge1 lan1`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1"},
				},
			},
		},
		{
			name:  "single bridge with multiple members",
			input: `bridge member bridge1 lan1 tunnel1`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1", "tunnel1"},
				},
			},
		},
		{
			name:  "single bridge with three members",
			input: `bridge member bridge1 lan1 tunnel1 tunnel2`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1", "tunnel1", "tunnel2"},
				},
			},
		},
		{
			name: "multiple bridges",
			input: `bridge member bridge1 lan1
bridge member bridge2 lan2 tunnel1`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1"},
				},
				{
					Name:    "bridge2",
					Members: []string{"lan2", "tunnel1"},
				},
			},
		},
		{
			name:  "bridge with VLAN interface",
			input: `bridge member bridge1 lan1/1 tunnel1`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1/1", "tunnel1"},
				},
			},
		},
		{
			name:  "bridge with pp interface",
			input: `bridge member bridge1 lan1 pp1`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1", "pp1"},
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []BridgeConfig{},
		},
		{
			name:     "no bridge config",
			input:    "some other config\nip lan1 address 192.168.1.1/24",
			expected: []BridgeConfig{},
		},
		{
			name:  "bridge with leading whitespace",
			input: `  bridge member bridge1 lan1 tunnel1`,
			expected: []BridgeConfig{
				{
					Name:    "bridge1",
					Members: []string{"lan1", "tunnel1"},
				},
			},
		},
	}

	parser := NewBridgeParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseBridgeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d bridges, got %d", len(tt.expected), len(result))
				return
			}

			// Create a map for easier comparison (order may vary)
			resultMap := make(map[string]BridgeConfig)
			for _, b := range result {
				resultMap[b.Name] = b
			}

			for _, expected := range tt.expected {
				got, ok := resultMap[expected.Name]
				if !ok {
					t.Errorf("bridge %s not found in result", expected.Name)
					continue
				}

				if len(got.Members) != len(expected.Members) {
					t.Errorf("%s: members count = %d, want %d", expected.Name, len(got.Members), len(expected.Members))
					continue
				}

				for i, member := range expected.Members {
					if got.Members[i] != member {
						t.Errorf("%s: member[%d] = %q, want %q", expected.Name, i, got.Members[i], member)
					}
				}
			}
		})
	}
}

func TestParseSingleBridge(t *testing.T) {
	parser := NewBridgeParser()

	input := `bridge member bridge1 lan1 tunnel1
bridge member bridge2 lan2`

	// Test finding existing bridge
	bridge, err := parser.ParseSingleBridge(input, "bridge1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if bridge == nil {
		t.Fatal("expected bridge, got nil")
	}
	if bridge.Name != "bridge1" {
		t.Errorf("Name = %q, want %q", bridge.Name, "bridge1")
	}
	if len(bridge.Members) != 2 {
		t.Errorf("Members count = %d, want 2", len(bridge.Members))
	}

	// Test not found
	_, err = parser.ParseSingleBridge(input, "bridge99")
	if err == nil {
		t.Error("expected error for non-existent bridge, got nil")
	}
}

func TestBuildBridgeMemberCommand(t *testing.T) {
	tests := []struct {
		name     string
		bridge   string
		members  []string
		expected string
	}{
		{
			name:     "single member",
			bridge:   "bridge1",
			members:  []string{"lan1"},
			expected: "bridge member bridge1 lan1",
		},
		{
			name:     "multiple members",
			bridge:   "bridge1",
			members:  []string{"lan1", "tunnel1"},
			expected: "bridge member bridge1 lan1 tunnel1",
		},
		{
			name:     "three members",
			bridge:   "bridge2",
			members:  []string{"lan1", "tunnel1", "tunnel2"},
			expected: "bridge member bridge2 lan1 tunnel1 tunnel2",
		},
		{
			name:     "empty members",
			bridge:   "bridge1",
			members:  []string{},
			expected: "bridge member bridge1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildBridgeMemberCommand(tt.bridge, tt.members)
			if result != tt.expected {
				t.Errorf("BuildBridgeMemberCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteBridgeCommand(t *testing.T) {
	result := BuildDeleteBridgeCommand("bridge1")
	expected := "no bridge member bridge1"
	if result != expected {
		t.Errorf("BuildDeleteBridgeCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowBridgeCommand(t *testing.T) {
	result := BuildShowBridgeCommand("bridge1")
	expected := `show config | grep "bridge member bridge1"`
	if result != expected {
		t.Errorf("BuildShowBridgeCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowAllBridgesCommand(t *testing.T) {
	result := BuildShowAllBridgesCommand()
	expected := `show config | grep "bridge member"`
	if result != expected {
		t.Errorf("BuildShowAllBridgesCommand() = %q, want %q", result, expected)
	}
}

func TestValidateBridgeName(t *testing.T) {
	tests := []struct {
		name    string
		bridge  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid bridge1",
			bridge:  "bridge1",
			wantErr: false,
		},
		{
			name:    "valid bridge10",
			bridge:  "bridge10",
			wantErr: false,
		},
		{
			name:    "valid bridge999",
			bridge:  "bridge999",
			wantErr: false,
		},
		{
			name:    "empty name",
			bridge:  "",
			wantErr: true,
			errMsg:  "bridge name is required",
		},
		{
			name:    "invalid format - no number",
			bridge:  "bridge",
			wantErr: true,
			errMsg:  "must be in format 'bridgeN'",
		},
		{
			name:    "invalid format - wrong prefix",
			bridge:  "br1",
			wantErr: true,
			errMsg:  "must be in format 'bridgeN'",
		},
		{
			name:    "invalid format - lan interface",
			bridge:  "lan1",
			wantErr: true,
			errMsg:  "must be in format 'bridgeN'",
		},
		{
			name:    "invalid format - tunnel interface",
			bridge:  "tunnel1",
			wantErr: true,
			errMsg:  "must be in format 'bridgeN'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBridgeName(tt.bridge)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateBridgeMember(t *testing.T) {
	tests := []struct {
		name    string
		member  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid lan1",
			member:  "lan1",
			wantErr: false,
		},
		{
			name:    "valid lan10",
			member:  "lan10",
			wantErr: false,
		},
		{
			name:    "valid VLAN interface",
			member:  "lan1/1",
			wantErr: false,
		},
		{
			name:    "valid VLAN interface 2",
			member:  "lan2/10",
			wantErr: false,
		},
		{
			name:    "valid tunnel1",
			member:  "tunnel1",
			wantErr: false,
		},
		{
			name:    "valid tunnel99",
			member:  "tunnel99",
			wantErr: false,
		},
		{
			name:    "valid pp1",
			member:  "pp1",
			wantErr: false,
		},
		{
			name:    "valid pp10",
			member:  "pp10",
			wantErr: false,
		},
		{
			name:    "valid loopback1",
			member:  "loopback1",
			wantErr: false,
		},
		{
			name:    "valid bridge interface",
			member:  "bridge2",
			wantErr: false,
		},
		{
			name:    "empty member",
			member:  "",
			wantErr: true,
			errMsg:  "member interface name is required",
		},
		{
			name:    "invalid format - eth0",
			member:  "eth0",
			wantErr: true,
			errMsg:  "invalid member interface name",
		},
		{
			name:    "invalid format - no number",
			member:  "lan",
			wantErr: true,
			errMsg:  "invalid member interface name",
		},
		{
			name:    "invalid format - arbitrary string",
			member:  "interface1",
			wantErr: true,
			errMsg:  "invalid member interface name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBridgeMember(tt.member)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateBridge(t *testing.T) {
	tests := []struct {
		name    string
		bridge  BridgeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bridge with single member",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1"},
			},
			wantErr: false,
		},
		{
			name: "valid bridge with multiple members",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1", "tunnel1"},
			},
			wantErr: false,
		},
		{
			name: "valid bridge with no members",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{},
			},
			wantErr: false,
		},
		{
			name: "invalid bridge name",
			bridge: BridgeConfig{
				Name:    "br1",
				Members: []string{"lan1"},
			},
			wantErr: true,
			errMsg:  "must be in format 'bridgeN'",
		},
		{
			name: "invalid member",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"eth0"},
			},
			wantErr: true,
			errMsg:  "invalid member interface name",
		},
		{
			name: "duplicate members",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1", "lan1"},
			},
			wantErr: true,
			errMsg:  "duplicate member interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBridge(tt.bridge)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
