package parsers

import (
	"testing"
)

func TestBGPParser_ParseBGPConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *BGPConfig
		wantErr  bool
	}{
		{
			name: "basic BGP configuration",
			input: `bgp use on
bgp autonomous-system 65001
bgp router id 10.0.0.1`,
			expected: &BGPConfig{
				Enabled:            true,
				ASN:                "65001",
				RouterID:           "10.0.0.1",
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors:          []BGPNeighbor{},
				Networks:           []BGPNetwork{},
			},
		},
		{
			name: "BGP with neighbor",
			input: `bgp use on
bgp autonomous-system 65001
bgp router id 10.0.0.1
bgp neighbor 1 address 10.0.0.2 as 65002
bgp neighbor 1 hold-time 90`,
			expected: &BGPConfig{
				Enabled:            true,
				ASN:                "65001",
				RouterID:           "10.0.0.1",
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors: []BGPNeighbor{
					{ID: 1, IP: "10.0.0.2", RemoteAS: "65002", HoldTime: 90},
				},
				Networks: []BGPNetwork{},
			},
		},
		{
			name: "BGP with multiple neighbors",
			input: `bgp use on
bgp autonomous-system 65001
bgp neighbor 1 address 10.0.0.2 as 65002
bgp neighbor 1 hold-time 90
bgp neighbor 1 password secret123
bgp neighbor 2 address 10.0.0.3 as 65003
bgp neighbor 2 multihop 2`,
			expected: &BGPConfig{
				Enabled:            true,
				ASN:                "65001",
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors: []BGPNeighbor{
					{ID: 1, IP: "10.0.0.2", RemoteAS: "65002", HoldTime: 90, Password: "secret123"},
					{ID: 2, IP: "10.0.0.3", RemoteAS: "65003", Multihop: 2},
				},
				Networks: []BGPNetwork{},
			},
		},
		{
			name: "BGP with network announcements",
			input: `bgp use on
bgp autonomous-system 65001
bgp import filter 1 include 192.168.1.0/255.255.255.0
bgp import filter 2 include 10.0.0.0/255.0.0.0`,
			expected: &BGPConfig{
				Enabled:            true,
				ASN:                "65001",
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors:          []BGPNeighbor{},
				Networks: []BGPNetwork{
					{Prefix: "192.168.1.0", Mask: "255.255.255.0"},
					{Prefix: "10.0.0.0", Mask: "255.0.0.0"},
				},
			},
		},
		{
			name: "BGP with redistribution",
			input: `bgp use on
bgp autonomous-system 65001
bgp import from static
bgp import from connected`,
			expected: &BGPConfig{
				Enabled:               true,
				ASN:                   "65001",
				DefaultIPv4Unicast:    true,
				LogNeighborChanges:    true,
				Neighbors:             []BGPNeighbor{},
				Networks:              []BGPNetwork{},
				RedistributeStatic:    true,
				RedistributeConnected: true,
			},
		},
		{
			name: "BGP disabled",
			input: `bgp use off
bgp autonomous-system 65001`,
			expected: &BGPConfig{
				Enabled:            false,
				ASN:                "65001",
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors:          []BGPNeighbor{},
				Networks:           []BGPNetwork{},
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: &BGPConfig{
				Enabled:            false,
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors:          []BGPNeighbor{},
				Networks:           []BGPNetwork{},
			},
		},
	}

	parser := NewBGPParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseBGPConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBGPConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", got.Enabled, tt.expected.Enabled)
			}
			if got.ASN != tt.expected.ASN {
				t.Errorf("ASN = %v, want %v", got.ASN, tt.expected.ASN)
			}
			if got.RouterID != tt.expected.RouterID {
				t.Errorf("RouterID = %v, want %v", got.RouterID, tt.expected.RouterID)
			}
			if len(got.Neighbors) != len(tt.expected.Neighbors) {
				t.Errorf("Neighbors count = %v, want %v", len(got.Neighbors), len(tt.expected.Neighbors))
			}
			if len(got.Networks) != len(tt.expected.Networks) {
				t.Errorf("Networks count = %v, want %v", len(got.Networks), len(tt.expected.Networks))
			}
			if got.RedistributeStatic != tt.expected.RedistributeStatic {
				t.Errorf("RedistributeStatic = %v, want %v", got.RedistributeStatic, tt.expected.RedistributeStatic)
			}
			if got.RedistributeConnected != tt.expected.RedistributeConnected {
				t.Errorf("RedistributeConnected = %v, want %v", got.RedistributeConnected, tt.expected.RedistributeConnected)
			}
		})
	}
}

func TestBuildBGPCommands(t *testing.T) {
	t.Run("BuildBGPUseCommand", func(t *testing.T) {
		if got := BuildBGPUseCommand(true); got != "bgp use on" {
			t.Errorf("BuildBGPUseCommand(true) = %v, want %v", got, "bgp use on")
		}
		if got := BuildBGPUseCommand(false); got != "bgp use off" {
			t.Errorf("BuildBGPUseCommand(false) = %v, want %v", got, "bgp use off")
		}
	})

	t.Run("BuildBGPASNCommand", func(t *testing.T) {
		if got := BuildBGPASNCommand("65001"); got != "bgp autonomous-system 65001" {
			t.Errorf("BuildBGPASNCommand() = %v, want %v", got, "bgp autonomous-system 65001")
		}
	})

	t.Run("BuildBGPRouterIDCommand", func(t *testing.T) {
		if got := BuildBGPRouterIDCommand("10.0.0.1"); got != "bgp router id 10.0.0.1" {
			t.Errorf("BuildBGPRouterIDCommand() = %v, want %v", got, "bgp router id 10.0.0.1")
		}
	})

	t.Run("BuildBGPNeighborCommand", func(t *testing.T) {
		neighbor := BGPNeighbor{ID: 1, IP: "10.0.0.2", RemoteAS: "65002"}
		expected := "bgp neighbor 1 address 10.0.0.2 as 65002"
		if got := BuildBGPNeighborCommand(neighbor); got != expected {
			t.Errorf("BuildBGPNeighborCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildBGPNeighborHoldTimeCommand", func(t *testing.T) {
		expected := "bgp neighbor 1 hold-time 90"
		if got := BuildBGPNeighborHoldTimeCommand(1, 90); got != expected {
			t.Errorf("BuildBGPNeighborHoldTimeCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildBGPNeighborPasswordCommand", func(t *testing.T) {
		expected := "bgp neighbor 1 password secret"
		if got := BuildBGPNeighborPasswordCommand(1, "secret"); got != expected {
			t.Errorf("BuildBGPNeighborPasswordCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildBGPNetworkCommand", func(t *testing.T) {
		network := BGPNetwork{Prefix: "192.168.1.0", Mask: "255.255.255.0"}
		expected := "bgp import filter 1 include 192.168.1.0/255.255.255.0"
		if got := BuildBGPNetworkCommand(1, network); got != expected {
			t.Errorf("BuildBGPNetworkCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildBGPRedistributeCommand", func(t *testing.T) {
		if got := BuildBGPRedistributeCommand("static"); got != "bgp import from static" {
			t.Errorf("BuildBGPRedistributeCommand(static) = %v, want %v", got, "bgp import from static")
		}
		if got := BuildBGPRedistributeCommand("connected"); got != "bgp import from connected" {
			t.Errorf("BuildBGPRedistributeCommand(connected) = %v, want %v", got, "bgp import from connected")
		}
	})
}

func TestValidateBGPConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  BGPConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: BGPConfig{
				ASN:      "65001",
				RouterID: "10.0.0.1",
				Neighbors: []BGPNeighbor{
					{ID: 1, IP: "10.0.0.2", RemoteAS: "65002"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing ASN",
			config: BGPConfig{
				RouterID: "10.0.0.1",
			},
			wantErr: true,
		},
		{
			name: "invalid ASN",
			config: BGPConfig{
				ASN: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid router ID",
			config: BGPConfig{
				ASN:      "65001",
				RouterID: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid neighbor IP",
			config: BGPConfig{
				ASN: "65001",
				Neighbors: []BGPNeighbor{
					{ID: 1, IP: "invalid", RemoteAS: "65002"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid neighbor hold-time",
			config: BGPConfig{
				ASN: "65001",
				Neighbors: []BGPNeighbor{
					{ID: 1, IP: "10.0.0.2", RemoteAS: "65002", HoldTime: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "4-byte ASN",
			config: BGPConfig{
				ASN: "4200000001",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBGPConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBGPConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
