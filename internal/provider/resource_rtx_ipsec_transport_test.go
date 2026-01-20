package provider

import (
	"testing"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func TestParseIPsecTransportConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parsers.IPsecTransport
	}{
		{
			name:  "single transport",
			input: "ipsec transport 1 1 udp 1701",
			expected: []parsers.IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
			},
		},
		{
			name:  "multiple transports",
			input: "ipsec transport 1 1 udp 1701\nipsec transport 2 2 tcp 443",
			expected: []parsers.IPsecTransport{
				{TransportID: 1, TunnelID: 1, Protocol: "udp", Port: 1701},
				{TransportID: 2, TunnelID: 2, Protocol: "tcp", Port: 443},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-matching lines",
			input:    "some other config line\nipsec tunnel 1",
			expected: nil,
		},
		{
			name:  "mixed content",
			input: "some other line\nipsec transport 5 3 udp 500\nanother line",
			expected: []parsers.IPsecTransport{
				{TransportID: 5, TunnelID: 3, Protocol: "udp", Port: 500},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsers.ParseIPsecTransportConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d transports, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if result[i].TransportID != expected.TransportID {
					t.Errorf("transport %d: expected TransportID %d, got %d", i, expected.TransportID, result[i].TransportID)
				}
				if result[i].TunnelID != expected.TunnelID {
					t.Errorf("transport %d: expected TunnelID %d, got %d", i, expected.TunnelID, result[i].TunnelID)
				}
				if result[i].Protocol != expected.Protocol {
					t.Errorf("transport %d: expected Protocol %s, got %s", i, expected.Protocol, result[i].Protocol)
				}
				if result[i].Port != expected.Port {
					t.Errorf("transport %d: expected Port %d, got %d", i, expected.Port, result[i].Port)
				}
			}
		})
	}
}

func TestBuildIPsecTransportCommand(t *testing.T) {
	tests := []struct {
		name      string
		transport parsers.IPsecTransport
		expected  string
	}{
		{
			name: "L2TP transport",
			transport: parsers.IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			expected: "ipsec transport 1 1 udp 1701",
		},
		{
			name: "TCP transport",
			transport: parsers.IPsecTransport{
				TransportID: 2,
				TunnelID:    5,
				Protocol:    "tcp",
				Port:        443,
			},
			expected: "ipsec transport 2 5 tcp 443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsers.BuildIPsecTransportCommand(tt.transport)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

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
			name:        "delete transport 100",
			transportID: 100,
			expected:    "no ipsec transport 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsers.BuildDeleteIPsecTransportCommand(tt.transportID)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateIPsecTransport(t *testing.T) {
	tests := []struct {
		name      string
		transport parsers.IPsecTransport
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid transport",
			transport: parsers.IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: false,
		},
		{
			name: "valid tcp transport",
			transport: parsers.IPsecTransport{
				TransportID: 2,
				TunnelID:    5,
				Protocol:    "tcp",
				Port:        443,
			},
			wantErr: false,
		},
		{
			name: "zero transport ID",
			transport: parsers.IPsecTransport{
				TransportID: 0,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "transport_id must be positive",
		},
		{
			name: "negative transport ID",
			transport: parsers.IPsecTransport{
				TransportID: -1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "transport_id must be positive",
		},
		{
			name: "zero tunnel ID",
			transport: parsers.IPsecTransport{
				TransportID: 1,
				TunnelID:    0,
				Protocol:    "udp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "tunnel_id must be positive",
		},
		{
			name: "empty protocol",
			transport: parsers.IPsecTransport{
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
			transport: parsers.IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "icmp",
				Port:        1701,
			},
			wantErr: true,
			errMsg:  "protocol must be one of",
		},
		{
			name: "zero port",
			transport: parsers.IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        0,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "port too high",
			transport: parsers.IPsecTransport{
				TransportID: 1,
				TunnelID:    1,
				Protocol:    "udp",
				Port:        70000,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parsers.ValidateIPsecTransport(tt.transport)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuildShowIPsecTransportCommand(t *testing.T) {
	expected := "show config | grep \"ipsec transport\""
	result := parsers.BuildShowIPsecTransportCommand()
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
