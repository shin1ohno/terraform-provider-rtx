package parsers

import (
	"testing"
)

// =============================================================================
// Task 8: TestParseIPsecTransportConfig - Transport mode parsing tests
// =============================================================================

// TestParseIPsecTransportConfig tests parsing of IPsec transport configurations
func TestParseIPsecTransportConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []IPsecTransport
	}{
		// Basic transport configuration
		{
			name:  "basic transport configuration",
			input: "ipsec transport 1 1 udp 1701",
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
			},
		},
		// L2TP over IPsec (most common use case)
		{
			name:  "L2TP over IPsec configuration",
			input: "ipsec transport 1 1 udp 1701",
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
			},
		},
		// TCP protocol
		{
			name:  "TCP transport configuration",
			input: "ipsec transport 2 1 tcp 443",
			expected: []IPsecTransport{
				{TransportID: 2, TunnelID: 1, Protocol: "tcp", Port: 443},
			},
		},
		// Multiple transport configurations
		{
			name: "multiple transport configurations",
			input: `ipsec transport 1 1 udp 1701
ipsec transport 2 2 udp 1701
ipsec transport 3 3 tcp 443`,
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
				{TransportID: 2, TunnelID: 2, Protocol: "udp", Port: 1701},
				{TransportID: 3, TunnelID: 3, Protocol: "tcp", Port: 443},
			},
		},
		// Transport with high IDs
		{
			name:  "transport with high IDs",
			input: "ipsec transport 100 50 udp 4500",
			expected: []IPsecTransport{
				{TransportID: 100, TunnelID: 50, Protocol: "udp", Port: 4500},
			},
		},
		// Transport with standard ports
		{
			name:  "IKE NAT-T port",
			input: "ipsec transport 1 1 udp 4500",
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 4500},
			},
		},
		// Empty input
		{
			name:     "empty input",
			input:    "",
			expected: []IPsecTransport{},
		},
		// Input with only whitespace
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []IPsecTransport{},
		},
		// Input with non-matching lines
		{
			name: "mixed content with non-matching lines",
			input: `# Comment line
ipsec transport 1 1 udp 1701
some other config
ipsec transport 2 2 tcp 443
tunnel select 1`,
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
				{TransportID: 2, TunnelID: 2, Protocol: "tcp", Port: 443},
			},
		},
		// Input with leading/trailing whitespace on lines
		{
			name:  "line with leading whitespace",
			input: "  ipsec transport 1 1 udp 1701  ",
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPsecTransportConfig(tt.input)
			if err != nil {
				t.Errorf("ParseIPsecTransportConfig() error = %v", err)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("ParseIPsecTransportConfig() returned %d transports, want %d", len(got), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if got[i].TransportID != expected.TransportID {
					t.Errorf("Transport[%d].TransportID = %v, want %v", i, got[i].TransportID, expected.TransportID)
				}
				if got[i].TunnelID != expected.TunnelID {
					t.Errorf("Transport[%d].TunnelID = %v, want %v", i, got[i].TunnelID, expected.TunnelID)
				}
				if got[i].Protocol != expected.Protocol {
					t.Errorf("Transport[%d].Protocol = %v, want %v", i, got[i].Protocol, expected.Protocol)
				}
				if got[i].Port != expected.Port {
					t.Errorf("Transport[%d].Port = %v, want %v", i, got[i].Port, expected.Port)
				}
			}
		})
	}
}

// TestParseIPsecTransportConfigEdgeCases tests edge cases for transport config parsing
func TestParseIPsecTransportConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectCount int
		description string
	}{
		// Minimum valid port
		{
			name:        "minimum port 1",
			input:       "ipsec transport 1 1 udp 1",
			expectCount: 1,
			description: "Port number 1 is valid",
		},
		// Maximum valid port
		{
			name:        "maximum port 65535",
			input:       "ipsec transport 1 1 udp 65535",
			expectCount: 1,
			description: "Port number 65535 is valid",
		},
		// Well-known ports
		{
			name:        "SSH port 22",
			input:       "ipsec transport 1 1 tcp 22",
			expectCount: 1,
			description: "SSH port configuration",
		},
		{
			name:        "HTTPS port 443",
			input:       "ipsec transport 1 1 tcp 443",
			expectCount: 1,
			description: "HTTPS port configuration",
		},
		{
			name:        "IKE port 500",
			input:       "ipsec transport 1 1 udp 500",
			expectCount: 1,
			description: "IKE port configuration",
		},
		{
			name:        "IKE NAT-T port 4500",
			input:       "ipsec transport 1 1 udp 4500",
			expectCount: 1,
			description: "IKE NAT-T port configuration",
		},
		// Different tunnel IDs mapping
		{
			name:        "transport ID differs from tunnel ID",
			input:       "ipsec transport 5 10 udp 1701",
			expectCount: 1,
			description: "Transport ID can be different from tunnel ID",
		},
		// Large IDs
		{
			name:        "large transport and tunnel IDs",
			input:       "ipsec transport 999 999 udp 1701",
			expectCount: 1,
			description: "Large ID values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPsecTransportConfig(tt.input)
			if err != nil {
				t.Errorf("ParseIPsecTransportConfig() error = %v", err)
				return
			}
			if len(got) != tt.expectCount {
				t.Errorf("ParseIPsecTransportConfig() returned %d transports, want %d", len(got), tt.expectCount)
			}
		})
	}
}

// TestParseIPsecTransportConfigNegative tests cases that should not match
func TestParseIPsecTransportConfigNegative(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectCount int
		description string
	}{
		// Invalid format - missing fields
		{
			name:        "missing port",
			input:       "ipsec transport 1 1 udp",
			expectCount: 0,
			description: "Command missing port should not match",
		},
		{
			name:        "missing protocol",
			input:       "ipsec transport 1 1",
			expectCount: 0,
			description: "Command missing protocol should not match",
		},
		{
			name:        "missing tunnel ID",
			input:       "ipsec transport 1",
			expectCount: 0,
			description: "Command missing tunnel ID should not match",
		},
		// Similar but different commands
		{
			name:        "ipsec tunnel command",
			input:       "ipsec tunnel 1",
			expectCount: 0,
			description: "ipsec tunnel should not match transport",
		},
		{
			name:        "tunnel select command",
			input:       "tunnel select 1",
			expectCount: 0,
			description: "tunnel select should not match transport",
		},
		// Malformed input
		{
			name:        "non-numeric transport ID",
			input:       "ipsec transport abc 1 udp 1701",
			expectCount: 0,
			description: "Non-numeric transport ID should not match",
		},
		{
			name:        "non-numeric tunnel ID",
			input:       "ipsec transport 1 abc udp 1701",
			expectCount: 0,
			description: "Non-numeric tunnel ID should not match",
		},
		{
			name:        "non-numeric port",
			input:       "ipsec transport 1 1 udp abc",
			expectCount: 0,
			description: "Non-numeric port should not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPsecTransportConfig(tt.input)
			if err != nil {
				t.Errorf("ParseIPsecTransportConfig() unexpected error = %v", err)
				return
			}
			if len(got) != tt.expectCount {
				t.Errorf("ParseIPsecTransportConfig() returned %d transports, want %d", len(got), tt.expectCount)
			}
		})
	}
}

// =============================================================================
// Task 9: TestParseIPsecTransportSA - SA configuration tests
// =============================================================================

// TestParseIPsecTransportSA tests parsing of transport configurations
// in the context of SA (Security Association) settings
func TestParseIPsecTransportSA(t *testing.T) {
	// Test various transport configurations that would be part of SA settings
	t.Run("transport mode for L2TP/IPsec", func(t *testing.T) {
		// L2TP/IPsec typically uses transport mode with UDP port 1701
		input := "ipsec transport 1 1 udp 1701"
		got, err := ParseIPsecTransportConfig(input)
		if err != nil {
			t.Fatalf("ParseIPsecTransportConfig() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Expected 1 transport, got %d", len(got))
		}
		if got[0].Protocol != "udp" {
			t.Errorf("Protocol = %v, want udp", got[0].Protocol)
		}
		if got[0].Port != 1701 {
			t.Errorf("Port = %v, want 1701", got[0].Port)
		}
	})

	t.Run("transport mode for IKE NAT-T", func(t *testing.T) {
		// NAT-T uses UDP port 4500
		input := "ipsec transport 2 1 udp 4500"
		got, err := ParseIPsecTransportConfig(input)
		if err != nil {
			t.Fatalf("ParseIPsecTransportConfig() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Expected 1 transport, got %d", len(got))
		}
		if got[0].Port != 4500 {
			t.Errorf("Port = %v, want 4500", got[0].Port)
		}
	})

	t.Run("multiple transports for different tunnels", func(t *testing.T) {
		// Multiple tunnels with their own transport configurations
		input := `ipsec transport 1 1 udp 1701
ipsec transport 2 2 udp 1701
ipsec transport 3 1 udp 4500`
		got, err := ParseIPsecTransportConfig(input)
		if err != nil {
			t.Fatalf("ParseIPsecTransportConfig() error = %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("Expected 3 transports, got %d", len(got))
		}

		// Verify each transport
		expectedPorts := []int{1701, 1701, 4500}
		for i, expected := range expectedPorts {
			if got[i].Port != expected {
				t.Errorf("Transport[%d].Port = %v, want %v", i, got[i].Port, expected)
			}
		}
	})
}

// TestIPsecTransportKeyLifetimeContext tests transport configurations
// in the context of key lifetime settings
func TestIPsecTransportKeyLifetimeContext(t *testing.T) {
	// These tests verify transport configurations that would be used
	// alongside key lifetime settings in a full IPsec configuration

	tests := []struct {
		name        string
		input       string
		expected    IPsecTransport
		description string
	}{
		{
			name:  "standard L2TP transport",
			input: "ipsec transport 1 1 udp 1701",
			expected: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			description: "Standard L2TP over IPsec transport",
		},
		{
			name:  "NAT-T transport",
			input: "ipsec transport 1 1 udp 4500",
			expected: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        4500,
			},
			description: "NAT-T transport for IPsec",
		},
		{
			name:  "IKE transport",
			input: "ipsec transport 1 1 udp 500",
			expected: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        500,
			},
			description: "IKE transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPsecTransportConfig(tt.input)
			if err != nil {
				t.Errorf("ParseIPsecTransportConfig() error = %v", err)
				return
			}
			if len(got) != 1 {
				t.Errorf("Expected 1 transport, got %d", len(got))
				return
			}

			if got[0] != tt.expected {
				t.Errorf("Transport = %+v, want %+v", got[0], tt.expected)
			}
		})
	}
}

// TestIPsecTransportPeerAddressContext tests transport configurations
// in the context of peer address settings
func TestIPsecTransportPeerAddressContext(t *testing.T) {
	// These tests simulate parsing transport configurations that would
	// appear alongside peer address configurations in show config output

	tests := []struct {
		name     string
		input    string
		expected []IPsecTransport
	}{
		{
			name: "transport with tunnel config context",
			input: `tunnel select 1
ipsec tunnel 1
ipsec transport 1 1 udp 1701
ipsec ike local address 1 192.168.1.1`,
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
			},
		},
		{
			name: "multiple transports in full config",
			input: `tunnel select 1
ipsec tunnel 1
ipsec transport 1 1 udp 1701
tunnel select 2
ipsec tunnel 2
ipsec transport 2 2 udp 1701`,
			expected: []IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
				{TransportID: 2, TunnelID: 2, Protocol: "udp", Port: 1701},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPsecTransportConfig(tt.input)
			if err != nil {
				t.Errorf("ParseIPsecTransportConfig() error = %v", err)
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("Expected %d transports, got %d", len(tt.expected), len(got))
				return
			}

			for i, expected := range tt.expected {
				if got[i] != expected {
					t.Errorf("Transport[%d] = %+v, want %+v", i, got[i], expected)
				}
			}
		})
	}
}

// =============================================================================
// Task 10: TestBuildIPsecTransportCommands - Command builder tests
// =============================================================================

// TestBuildIPsecTransportCommand tests building transport commands
func TestBuildIPsecTransportCommand(t *testing.T) {
	tests := []struct {
		name      string
		transport IPsecTransport
		expected  string
	}{
		// Basic L2TP configuration
		{
			name: "L2TP transport",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			expected: "ipsec transport 1 1 udp 1701",
		},
		// NAT-T configuration
		{
			name: "NAT-T transport",
			transport: IPsecTransport{
				TransportID: 2,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        4500,
			},
			expected: "ipsec transport 2 1 udp 4500",
		},
		// TCP transport
		{
			name: "TCP transport",
			transport: IPsecTransport{
				TransportID: 3,
				TunnelID:    2,
				Protocol:    "tcp",
				Port:        443,
			},
			expected: "ipsec transport 3 2 tcp 443",
		},
		// IKE transport
		{
			name: "IKE transport",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        500,
			},
			expected: "ipsec transport 1 1 udp 500",
		},
		// High IDs
		{
			name: "high IDs",
			transport: IPsecTransport{
				TransportID: 100,
				TunnelID:    50,
				Protocol:    "udp",
				Port:        1701,
			},
			expected: "ipsec transport 100 50 udp 1701",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIPsecTransportCommand(tt.transport)
			if got != tt.expected {
				t.Errorf("BuildIPsecTransportCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBuildDeleteIPsecTransportCommand tests building delete commands
func TestBuildDeleteIPsecTransportCommand(t *testing.T) {
	tests := []struct {
		name        string
		transportID int
		expected    string
	}{
		{
			name:        "delete transport 1",
			transportID: 1,
			expected:    "no ipsec transport 1",
		},
		{
			name:        "delete transport 10",
			transportID: 10,
			expected:    "no ipsec transport 10",
		},
		{
			name:        "delete transport 100",
			transportID: 100,
			expected:    "no ipsec transport 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDeleteIPsecTransportCommand(tt.transportID)
			if got != tt.expected {
				t.Errorf("BuildDeleteIPsecTransportCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBuildShowIPsecTransportCommand tests building show command
func TestBuildShowIPsecTransportCommand(t *testing.T) {
	expected := "show config | grep \"ipsec transport\""
	got := BuildShowIPsecTransportCommand()
	if got != expected {
		t.Errorf("BuildShowIPsecTransportCommand() = %q, want %q", got, expected)
	}
}

// TestIPsecTransportRoundTrip tests round-trip (parse -> build -> parse)
func TestIPsecTransportRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		transport IPsecTransport
	}{
		{
			name: "L2TP transport round-trip",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
		},
		{
			name: "NAT-T transport round-trip",
			transport: IPsecTransport{
				TransportID: 2,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        4500,
			},
		},
		{
			name: "TCP transport round-trip",
			transport: IPsecTransport{
				TransportID: 3,
				TunnelID:    2,
				Protocol:    "tcp",
				Port:        443,
			},
		},
		{
			name: "high IDs round-trip",
			transport: IPsecTransport{
				TransportID: 99,
				TunnelID:    88,
				Protocol:    "udp",
				Port:        12345,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build command from transport
			cmd := BuildIPsecTransportCommand(tt.transport)

			// Parse the command back
			parsed, err := ParseIPsecTransportConfig(cmd)
			if err != nil {
				t.Errorf("Round-trip parse error: %v", err)
				return
			}
			if len(parsed) != 1 {
				t.Errorf("Round-trip expected 1 transport, got %d", len(parsed))
				return
			}

			// Verify all fields match
			got := parsed[0]
			if got.TransportID != tt.transport.TransportID {
				t.Errorf("Round-trip TransportID = %v, want %v", got.TransportID, tt.transport.TransportID)
			}
			if got.TunnelID != tt.transport.TunnelID {
				t.Errorf("Round-trip TunnelID = %v, want %v", got.TunnelID, tt.transport.TunnelID)
			}
			if got.Protocol != tt.transport.Protocol {
				t.Errorf("Round-trip Protocol = %v, want %v", got.Protocol, tt.transport.Protocol)
			}
			if got.Port != tt.transport.Port {
				t.Errorf("Round-trip Port = %v, want %v", got.Port, tt.transport.Port)
			}
		})
	}
}

// TestIPsecTransportCommandGeneration tests that generated commands are correct
func TestIPsecTransportCommandGeneration(t *testing.T) {
	// Test various configurations and verify the generated commands
	t.Run("command format verification", func(t *testing.T) {
		transport := IPsecTransport{
			TransportID: 1,
			TunnelID:    2,
			Protocol:    "udp",
			Port:        1701,
		}

		cmd := BuildIPsecTransportCommand(transport)

		// Verify format: "ipsec transport <id> <tunnel_id> <proto> <port>"
		expected := "ipsec transport 1 2 udp 1701"
		if cmd != expected {
			t.Errorf("Command format incorrect: got %q, want %q", cmd, expected)
		}
	})

	t.Run("delete command format verification", func(t *testing.T) {
		cmd := BuildDeleteIPsecTransportCommand(5)

		// Verify format: "no ipsec transport <id>"
		expected := "no ipsec transport 5"
		if cmd != expected {
			t.Errorf("Delete command format incorrect: got %q, want %q", cmd, expected)
		}
	})
}

// =============================================================================
// Validation Tests
// =============================================================================

// TestValidateIPsecTransport tests the validation function
func TestValidateIPsecTransport(t *testing.T) {
	tests := []struct {
		name      string
		transport IPsecTransport
		wantErr   bool
		errMsg    string
	}{
		// Valid configurations
		{
			name: "valid L2TP transport",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: false,
		},
		{
			name: "valid TCP transport",
			transport: IPsecTransport{
				TransportID: 2,
				TunnelID:    1,
				Protocol:    "tcp",
				Port:        443,
			},
			wantErr: false,
		},
		{
			name: "valid UDP transport case insensitive",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "UDP",
				Port:        1701,
			},
			wantErr: false,
		},
		{
			name: "valid TCP transport case insensitive",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "TCP",
				Port:        443,
			},
			wantErr: false,
		},
		// Invalid transport ID
		{
			name: "invalid transport ID zero",
			transport: IPsecTransport{
				TransportID: 0,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "transport_id must be positive",
		},
		{
			name: "invalid transport ID negative",
			transport: IPsecTransport{
				TransportID: -1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "transport_id must be positive",
		},
		// Invalid tunnel ID
		{
			name: "invalid tunnel ID zero",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    0,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "tunnel_id must be positive",
		},
		{
			name: "invalid tunnel ID negative",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    -1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "tunnel_id must be positive",
		},
		// Invalid protocol
		{
			name: "empty protocol",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "protocol is required",
		},
		{
			name: "invalid protocol",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "icmp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "protocol must be one of",
		},
		{
			name: "invalid protocol unknown",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "sctp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "protocol must be one of",
		},
		// Invalid port
		{
			name: "invalid port zero",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        0,
			},
			wantErr: true,
			errMsg:  "port must be between",
		},
		{
			name: "invalid port negative",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        -1,
			},
			wantErr: true,
			errMsg:  "port must be between",
		},
		{
			name: "invalid port too large",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        65536,
			},
			wantErr: true,
			errMsg:  "port must be between",
		},
		// Valid edge cases for port
		{
			name: "valid port minimum",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1,
			},
			wantErr: false,
		},
		{
			name: "valid port maximum",
			transport: IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        65535,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPsecTransport(tt.transport)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPsecTransport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

// TestIPsecTransportIntegration tests transport parsing in realistic scenarios
func TestIPsecTransportIntegration(t *testing.T) {
	t.Run("L2TP/IPsec full configuration", func(t *testing.T) {
		// Simulates output from "show config | grep ipsec transport"
		input := `ipsec transport 1 1 udp 1701
ipsec transport 2 1 udp 4500`

		transports, err := ParseIPsecTransportConfig(input)
		if err != nil {
			t.Fatalf("ParseIPsecTransportConfig() error = %v", err)
		}
		if len(transports) != 2 {
			t.Fatalf("Expected 2 transports, got %d", len(transports))
		}

		// Verify L2TP transport
		if transports[0].Port != 1701 {
			t.Errorf("L2TP port = %v, want 1701", transports[0].Port)
		}

		// Verify NAT-T transport
		if transports[1].Port != 4500 {
			t.Errorf("NAT-T port = %v, want 4500", transports[1].Port)
		}
	})

	t.Run("validate and build workflow", func(t *testing.T) {
		transport := IPsecTransport{
			TransportID: 1,
			TunnelID:    1,
			Protocol:    "udp",
			Port:        1701,
		}

		// Validate first
		err := ValidateIPsecTransport(transport)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}

		// Build command
		cmd := BuildIPsecTransportCommand(transport)

		// Verify command can be parsed back
		parsed, err := ParseIPsecTransportConfig(cmd)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if len(parsed) != 1 {
			t.Fatalf("Expected 1 transport, got %d", len(parsed))
		}

		// Verify parsed matches original
		if parsed[0] != transport {
			t.Errorf("Parsed transport %+v does not match original %+v", parsed[0], transport)
		}
	})

	t.Run("create and delete workflow", func(t *testing.T) {
		transport := IPsecTransport{
			TransportID: 5,
			TunnelID:    3,
			Protocol:    "udp",
			Port:        1701,
		}

		// Build create command
		createCmd := BuildIPsecTransportCommand(transport)
		expectedCreate := "ipsec transport 5 3 udp 1701"
		if createCmd != expectedCreate {
			t.Errorf("Create command = %q, want %q", createCmd, expectedCreate)
		}

		// Build delete command
		deleteCmd := BuildDeleteIPsecTransportCommand(transport.TransportID)
		expectedDelete := "no ipsec transport 5"
		if deleteCmd != expectedDelete {
			t.Errorf("Delete command = %q, want %q", deleteCmd, expectedDelete)
		}
	})
}
