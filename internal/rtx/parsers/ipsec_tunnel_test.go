package parsers

import (
	"testing"
)

func TestIPsecTunnelParser_ParseIPsecTunnelConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected tunnel count
	}{
		{
			name: "basic IPsec tunnel",
			input: `tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 10.0.0.1
ipsec ike remote address 1 10.0.0.2
ipsec ike pre-shared-key 1 text secret123`,
			expected: 1,
		},
		{
			name: "IPsec tunnel with encryption settings",
			input: `tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 10.0.0.1
ipsec ike remote address 1 10.0.0.2
ipsec ike pre-shared-key 1 text secret123
ipsec ike encryption 1 aes-cbc-256
ipsec ike hash 1 sha256
ipsec ike group 1 modp2048`,
			expected: 1,
		},
		{
			name: "IPsec tunnel with DPD",
			input: `tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 10.0.0.1
ipsec ike remote address 1 10.0.0.2
ipsec ike keepalive use 1 on dpd 30`,
			expected: 1,
		},
		{
			name: "IPsec tunnel with DPD and retry",
			input: `tunnel select 1
ipsec tunnel 1
ipsec ike keepalive use 1 on dpd 30 5`,
			expected: 1,
		},
		{
			name: "multiple IPsec tunnels",
			input: `tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 10.0.0.1
ipsec ike remote address 1 10.0.0.2
tunnel select 2
ipsec tunnel 2
ipsec ike local address 2 10.0.0.1
ipsec ike remote address 2 10.0.0.3`,
			expected: 2,
		},
		{
			name: "IPsec SA policy",
			input: `tunnel select 1
ipsec tunnel 1
ipsec sa policy 101 1 esp aes-cbc sha-hmac`,
			expected: 1,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
	}

	parser := NewIPsecTunnelParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseIPsecTunnelConfig(tt.input)
			if err != nil {
				t.Errorf("ParseIPsecTunnelConfig() error = %v", err)
				return
			}
			if len(got) != tt.expected {
				t.Errorf("Tunnel count = %v, want %v", len(got), tt.expected)
			}
		})
	}
}

func TestIPsecTunnelParser_ParseDetails(t *testing.T) {
	input := `tunnel select 1
ipsec tunnel 1
ipsec ike local address 1 10.0.0.1
ipsec ike remote address 1 10.0.0.2
ipsec ike pre-shared-key 1 text mysecret
ipsec ike encryption 1 aes-cbc-256
ipsec ike hash 1 sha256
ipsec ike group 1 modp2048
ipsec ike keepalive use 1 on dpd 30 5
description Site-A VPN`

	parser := NewIPsecTunnelParser()
	tunnels, err := parser.ParseIPsecTunnelConfig(input)
	if err != nil {
		t.Fatalf("ParseIPsecTunnelConfig() error = %v", err)
	}

	if len(tunnels) != 1 {
		t.Fatalf("Expected 1 tunnel, got %d", len(tunnels))
	}

	tunnel := tunnels[0]

	if tunnel.ID != 1 {
		t.Errorf("ID = %v, want 1", tunnel.ID)
	}
	if tunnel.LocalAddress != "10.0.0.1" {
		t.Errorf("LocalAddress = %v, want 10.0.0.1", tunnel.LocalAddress)
	}
	if tunnel.RemoteAddress != "10.0.0.2" {
		t.Errorf("RemoteAddress = %v, want 10.0.0.2", tunnel.RemoteAddress)
	}
	if tunnel.PreSharedKey != "mysecret" {
		t.Errorf("PreSharedKey = %v, want mysecret", tunnel.PreSharedKey)
	}
	if !tunnel.IKEv2Proposal.EncryptionAES256 {
		t.Errorf("EncryptionAES256 = %v, want true", tunnel.IKEv2Proposal.EncryptionAES256)
	}
	if !tunnel.IKEv2Proposal.IntegritySHA256 {
		t.Errorf("IntegritySHA256 = %v, want true", tunnel.IKEv2Proposal.IntegritySHA256)
	}
	if !tunnel.IKEv2Proposal.GroupFourteen {
		t.Errorf("GroupFourteen = %v, want true", tunnel.IKEv2Proposal.GroupFourteen)
	}
	if !tunnel.DPDEnabled {
		t.Errorf("DPDEnabled = %v, want true", tunnel.DPDEnabled)
	}
	if tunnel.DPDInterval != 30 {
		t.Errorf("DPDInterval = %v, want 30", tunnel.DPDInterval)
	}
	if tunnel.DPDRetry != 5 {
		t.Errorf("DPDRetry = %v, want 5", tunnel.DPDRetry)
	}
	if tunnel.Name != "Site-A VPN" {
		t.Errorf("Name = %v, want 'Site-A VPN'", tunnel.Name)
	}
}

func TestBuildIPsecTunnelCommands(t *testing.T) {
	t.Run("BuildTunnelSelectCommand", func(t *testing.T) {
		expected := "tunnel select 1"
		if got := BuildTunnelSelectCommand(1); got != expected {
			t.Errorf("BuildTunnelSelectCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecTunnelCommand", func(t *testing.T) {
		expected := "ipsec tunnel 1"
		if got := BuildIPsecTunnelCommand(1); got != expected {
			t.Errorf("BuildIPsecTunnelCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKELocalAddressCommand", func(t *testing.T) {
		expected := "ipsec ike local address 1 10.0.0.1"
		if got := BuildIPsecIKELocalAddressCommand(1, "10.0.0.1"); got != expected {
			t.Errorf("BuildIPsecIKELocalAddressCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKERemoteAddressCommand", func(t *testing.T) {
		expected := "ipsec ike remote address 1 10.0.0.2"
		if got := BuildIPsecIKERemoteAddressCommand(1, "10.0.0.2"); got != expected {
			t.Errorf("BuildIPsecIKERemoteAddressCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEPreSharedKeyCommand", func(t *testing.T) {
		expected := "ipsec ike pre-shared-key 1 text secret123"
		if got := BuildIPsecIKEPreSharedKeyCommand(1, "secret123"); got != expected {
			t.Errorf("BuildIPsecIKEPreSharedKeyCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEEncryptionCommand AES256", func(t *testing.T) {
		proposal := IKEv2Proposal{EncryptionAES256: true}
		expected := "ipsec ike encryption 1 aes-cbc-256"
		if got := BuildIPsecIKEEncryptionCommand(1, proposal); got != expected {
			t.Errorf("BuildIPsecIKEEncryptionCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEEncryptionCommand 3DES", func(t *testing.T) {
		proposal := IKEv2Proposal{Encryption3DES: true}
		expected := "ipsec ike encryption 1 3des-cbc"
		if got := BuildIPsecIKEEncryptionCommand(1, proposal); got != expected {
			t.Errorf("BuildIPsecIKEEncryptionCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEHashCommand SHA256", func(t *testing.T) {
		proposal := IKEv2Proposal{IntegritySHA256: true}
		expected := "ipsec ike hash 1 sha256"
		if got := BuildIPsecIKEHashCommand(1, proposal); got != expected {
			t.Errorf("BuildIPsecIKEHashCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEGroupCommand modp2048", func(t *testing.T) {
		proposal := IKEv2Proposal{GroupFourteen: true}
		expected := "ipsec ike group 1 modp2048"
		if got := BuildIPsecIKEGroupCommand(1, proposal); got != expected {
			t.Errorf("BuildIPsecIKEGroupCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecSAPolicyCommand", func(t *testing.T) {
		transform := IPsecTransform{
			Protocol:         "esp",
			EncryptionAES256: true,
			IntegritySHA256:  true,
		}
		expected := "ipsec sa policy 101 1 esp aes-cbc-256 sha256-hmac"
		if got := BuildIPsecSAPolicyCommand(101, 1, transform); got != expected {
			t.Errorf("BuildIPsecSAPolicyCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEKeepaliveCommand", func(t *testing.T) {
		expected := "ipsec ike keepalive use 1 on dpd 30 5"
		if got := BuildIPsecIKEKeepaliveCommand(1, 30, 5); got != expected {
			t.Errorf("BuildIPsecIKEKeepaliveCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPsecIKEKeepaliveCommand without retry", func(t *testing.T) {
		expected := "ipsec ike keepalive use 1 on dpd 30"
		if got := BuildIPsecIKEKeepaliveCommand(1, 30, 0); got != expected {
			t.Errorf("BuildIPsecIKEKeepaliveCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildDeleteIPsecTunnelCommand", func(t *testing.T) {
		expected := "no ipsec tunnel 1"
		if got := BuildDeleteIPsecTunnelCommand(1); got != expected {
			t.Errorf("BuildDeleteIPsecTunnelCommand() = %v, want %v", got, expected)
		}
	})
}

func TestValidateIPsecTunnel(t *testing.T) {
	tests := []struct {
		name    string
		tunnel  IPsecTunnel
		wantErr bool
	}{
		{
			name: "valid tunnel",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "10.0.0.2",
				PreSharedKey:  "secret123",
			},
			wantErr: false,
		},
		{
			name: "invalid tunnel ID",
			tunnel: IPsecTunnel{
				ID:            0,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "10.0.0.2",
				PreSharedKey:  "secret123",
			},
			wantErr: true,
		},
		{
			name: "invalid local address",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "invalid",
				RemoteAddress: "10.0.0.2",
				PreSharedKey:  "secret123",
			},
			wantErr: true,
		},
		{
			name: "invalid remote address",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "invalid",
				PreSharedKey:  "secret123",
			},
			wantErr: true,
		},
		{
			name: "missing pre-shared key (optional)",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "10.0.0.2",
			},
			wantErr: false, // Pre-shared key is now optional (e.g., for L2TP transport mode)
		},
		{
			name: "valid tunnel with networks",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "10.0.0.2",
				PreSharedKey:  "secret123",
				LocalNetwork:  "192.168.1.0/24",
				RemoteNetwork: "192.168.2.0/24",
			},
			wantErr: false,
		},
		{
			name: "invalid local network",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "10.0.0.2",
				PreSharedKey:  "secret123",
				LocalNetwork:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid DPD interval",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "10.0.0.1",
				RemoteAddress: "10.0.0.2",
				PreSharedKey:  "secret123",
				DPDInterval:   -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPsecTunnel(tt.tunnel)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPsecTunnel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
