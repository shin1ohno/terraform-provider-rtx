package parsers

import (
	"testing"
)

func TestParseDHCPBindings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		scopeID  int
		expected []DHCPBinding
		wantErr  bool
	}{
		{
			name:    "RTX830 format with multiple bindings",
			scopeID: 1,
			input: `Scope 1:
  192.168.1.100    00:11:22:33:44:55
  192.168.1.101    00:aa:bb:cc:dd:ee
  192.168.1.110    ethernet 00:12:34:56:78:90
`,
			expected: []DHCPBinding{
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.100",
					MACAddress:          "00:11:22:33:44:55",
					UseClientIdentifier: false,
				},
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.101",
					MACAddress:          "00:aa:bb:cc:dd:ee",
					UseClientIdentifier: false,
				},
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.110",
					MACAddress:          "00:12:34:56:78:90",
					UseClientIdentifier: true,
				},
			},
			wantErr: false,
		},
		{
			name:    "RTX1210 format",
			scopeID: 1,
			input: `DHCP Scope 1 Bindings:
IP Address       MAC Address         Type
192.168.1.50     00:a0:c5:12:34:56   MAC
192.168.1.51     00:a0:c5:12:34:57   ethernet
`,
			expected: []DHCPBinding{
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.50",
					MACAddress:          "00:a0:c5:12:34:56",
					UseClientIdentifier: false,
				},
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.51",
					MACAddress:          "00:a0:c5:12:34:57",
					UseClientIdentifier: true,
				},
			},
			wantErr: false,
		},
		{
			name:     "Empty bindings",
			scopeID:  1,
			input:    "Scope 1:\n",
			expected: []DHCPBinding{},
			wantErr:  false,
		},
		{
			name:     "No bindings message",
			scopeID:  1,
			input:    "No bindings found for scope 1",
			expected: []DHCPBinding{},
			wantErr:  false,
		},
		{
			name:    "MAC address without separators",
			scopeID: 1,
			input: `Scope 1:
  192.168.1.100    001122334455
`,
			expected: []DHCPBinding{
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.100",
					MACAddress:          "00:11:22:33:44:55",
					UseClientIdentifier: false,
				},
			},
			wantErr: false,
		},
		{
			name:    "MAC address with hyphens",
			scopeID: 1,
			input: `Scope 1:
  192.168.1.100    00-11-22-33-44-55
`,
			expected: []DHCPBinding{
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.100",
					MACAddress:          "00:11:22:33:44:55",
					UseClientIdentifier: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &dhcpBindingsParser{}
			result, err := parser.ParseBindings(tt.input, tt.scopeID)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBindings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("ParseBindings() returned %d bindings, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, binding := range result {
				if binding.ScopeID != tt.expected[i].ScopeID {
					t.Errorf("binding[%d].ScopeID = %d, want %d", i, binding.ScopeID, tt.expected[i].ScopeID)
				}
				if binding.IPAddress != tt.expected[i].IPAddress {
					t.Errorf("binding[%d].IPAddress = %s, want %s", i, binding.IPAddress, tt.expected[i].IPAddress)
				}
				if binding.MACAddress != tt.expected[i].MACAddress {
					t.Errorf("binding[%d].MACAddress = %s, want %s", i, binding.MACAddress, tt.expected[i].MACAddress)
				}
				if binding.UseClientIdentifier != tt.expected[i].UseClientIdentifier {
					t.Errorf("binding[%d].UseClientIdentifier = %v, want %v", i, binding.UseClientIdentifier, tt.expected[i].UseClientIdentifier)
				}
			}
		})
	}
}

func TestNormalizeMACAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Already normalized",
			input:    "00:11:22:33:44:55",
			expected: "00:11:22:33:44:55",
			wantErr:  false,
		},
		{
			name:     "No separators",
			input:    "001122334455",
			expected: "00:11:22:33:44:55",
			wantErr:  false,
		},
		{
			name:     "Hyphens",
			input:    "00-11-22-33-44-55",
			expected: "00:11:22:33:44:55",
			wantErr:  false,
		},
		{
			name:     "Mixed case",
			input:    "00:AA:bb:CC:dd:EE",
			expected: "00:aa:bb:cc:dd:ee",
			wantErr:  false,
		},
		{
			name:     "Invalid length",
			input:    "00:11:22",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid characters",
			input:    "00:11:22:GG:44:55",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeMACAddress(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeMACAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result != tt.expected {
				t.Errorf("normalizeMACAddress() = %s, want %s", result, tt.expected)
			}
		})
	}
}