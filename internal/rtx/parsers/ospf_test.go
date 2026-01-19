package parsers

import (
	"testing"
)

func TestOSPFParser_ParseOSPFConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *OSPFConfig
	}{
		{
			name: "basic OSPF configuration",
			input: `ospf use on
ospf router id 10.0.0.1`,
			expected: &OSPFConfig{
				Enabled:   true,
				RouterID:  "10.0.0.1",
				ProcessID: 1,
				Distance:  110,
				Networks:  []OSPFNetwork{},
				Areas:     []OSPFArea{},
				Neighbors: []OSPFNeighbor{},
			},
		},
		{
			name: "OSPF with areas",
			input: `ospf use on
ospf router id 10.0.0.1
ospf area 0
ospf area 1 stub`,
			expected: &OSPFConfig{
				Enabled:   true,
				RouterID:  "10.0.0.1",
				ProcessID: 1,
				Distance:  110,
				Networks:  []OSPFNetwork{},
				Areas: []OSPFArea{
					{ID: "0", Type: "normal"},
					{ID: "1", Type: "stub"},
				},
				Neighbors: []OSPFNeighbor{},
			},
		},
		{
			name: "OSPF with stub no-summary",
			input: `ospf use on
ospf router id 10.0.0.1
ospf area 1 stub no-summary`,
			expected: &OSPFConfig{
				Enabled:   true,
				RouterID:  "10.0.0.1",
				ProcessID: 1,
				Distance:  110,
				Networks:  []OSPFNetwork{},
				Areas: []OSPFArea{
					{ID: "1", Type: "stub", NoSummary: true},
				},
				Neighbors: []OSPFNeighbor{},
			},
		},
		{
			name: "OSPF with NSSA",
			input: `ospf use on
ospf router id 10.0.0.1
ospf area 2 nssa
ospf area 3 nssa no-summary`,
			expected: &OSPFConfig{
				Enabled:   true,
				RouterID:  "10.0.0.1",
				ProcessID: 1,
				Distance:  110,
				Networks:  []OSPFNetwork{},
				Areas: []OSPFArea{
					{ID: "2", Type: "nssa"},
					{ID: "3", Type: "nssa", NoSummary: true},
				},
				Neighbors: []OSPFNeighbor{},
			},
		},
		{
			name: "OSPF with interface in area",
			input: `ospf use on
ospf router id 10.0.0.1
ip lan1 ospf area 0
ip lan2 ospf area 1`,
			expected: &OSPFConfig{
				Enabled:   true,
				RouterID:  "10.0.0.1",
				ProcessID: 1,
				Distance:  110,
				Networks: []OSPFNetwork{
					{IP: "lan1", Area: "0"},
					{IP: "lan2", Area: "1"},
				},
				Areas:     []OSPFArea{},
				Neighbors: []OSPFNeighbor{},
			},
		},
		{
			name: "OSPF with redistribution",
			input: `ospf use on
ospf router id 10.0.0.1
ospf import from static
ospf import from connected`,
			expected: &OSPFConfig{
				Enabled:               true,
				RouterID:              "10.0.0.1",
				ProcessID:             1,
				Distance:              110,
				Networks:              []OSPFNetwork{},
				Areas:                 []OSPFArea{},
				Neighbors:             []OSPFNeighbor{},
				RedistributeStatic:    true,
				RedistributeConnected: true,
			},
		},
		{
			name: "OSPF disabled",
			input: `ospf use off
ospf router id 10.0.0.1`,
			expected: &OSPFConfig{
				Enabled:   false,
				RouterID:  "10.0.0.1",
				ProcessID: 1,
				Distance:  110,
				Networks:  []OSPFNetwork{},
				Areas:     []OSPFArea{},
				Neighbors: []OSPFNeighbor{},
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: &OSPFConfig{
				Enabled:   false,
				ProcessID: 1,
				Distance:  110,
				Networks:  []OSPFNetwork{},
				Areas:     []OSPFArea{},
				Neighbors: []OSPFNeighbor{},
			},
		},
	}

	parser := NewOSPFParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseOSPFConfig(tt.input)
			if err != nil {
				t.Errorf("ParseOSPFConfig() error = %v", err)
				return
			}
			if got.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", got.Enabled, tt.expected.Enabled)
			}
			if got.RouterID != tt.expected.RouterID {
				t.Errorf("RouterID = %v, want %v", got.RouterID, tt.expected.RouterID)
			}
			if len(got.Areas) != len(tt.expected.Areas) {
				t.Errorf("Areas count = %v, want %v", len(got.Areas), len(tt.expected.Areas))
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

func TestBuildOSPFCommands(t *testing.T) {
	t.Run("BuildOSPFEnableCommand", func(t *testing.T) {
		if got := BuildOSPFEnableCommand(); got != "ospf use on" {
			t.Errorf("BuildOSPFEnableCommand() = %v, want %v", got, "ospf use on")
		}
	})

	t.Run("BuildOSPFDisableCommand", func(t *testing.T) {
		if got := BuildOSPFDisableCommand(); got != "ospf use off" {
			t.Errorf("BuildOSPFDisableCommand() = %v, want %v", got, "ospf use off")
		}
	})

	t.Run("BuildOSPFRouterIDCommand", func(t *testing.T) {
		expected := "ospf router id 10.0.0.1"
		if got := BuildOSPFRouterIDCommand("10.0.0.1"); got != expected {
			t.Errorf("BuildOSPFRouterIDCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildOSPFAreaCommand normal", func(t *testing.T) {
		area := OSPFArea{ID: "0", Type: "normal"}
		expected := "ospf area 0"
		if got := BuildOSPFAreaCommand(area); got != expected {
			t.Errorf("BuildOSPFAreaCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildOSPFAreaCommand stub", func(t *testing.T) {
		area := OSPFArea{ID: "1", Type: "stub"}
		expected := "ospf area 1 stub"
		if got := BuildOSPFAreaCommand(area); got != expected {
			t.Errorf("BuildOSPFAreaCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildOSPFAreaCommand stub no-summary", func(t *testing.T) {
		area := OSPFArea{ID: "1", Type: "stub", NoSummary: true}
		expected := "ospf area 1 stub no-summary"
		if got := BuildOSPFAreaCommand(area); got != expected {
			t.Errorf("BuildOSPFAreaCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildOSPFAreaCommand nssa", func(t *testing.T) {
		area := OSPFArea{ID: "2", Type: "nssa"}
		expected := "ospf area 2 nssa"
		if got := BuildOSPFAreaCommand(area); got != expected {
			t.Errorf("BuildOSPFAreaCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildIPOSPFAreaCommand", func(t *testing.T) {
		expected := "ip lan1 ospf area 0"
		if got := BuildIPOSPFAreaCommand("lan1", "0"); got != expected {
			t.Errorf("BuildIPOSPFAreaCommand() = %v, want %v", got, expected)
		}
	})

	t.Run("BuildOSPFImportCommand", func(t *testing.T) {
		expected := "ospf import from static"
		if got := BuildOSPFImportCommand("static"); got != expected {
			t.Errorf("BuildOSPFImportCommand() = %v, want %v", got, expected)
		}
	})
}

func TestValidateOSPFConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  OSPFConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: OSPFConfig{
				RouterID: "10.0.0.1",
				Areas: []OSPFArea{
					{ID: "0", Type: "normal"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing router ID",
			config: OSPFConfig{
				Areas: []OSPFArea{
					{ID: "0"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid router ID",
			config: OSPFConfig{
				RouterID: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid area with dotted decimal",
			config: OSPFConfig{
				RouterID: "10.0.0.1",
				Areas: []OSPFArea{
					{ID: "0.0.0.0"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid area type",
			config: OSPFConfig{
				RouterID: "10.0.0.1",
				Areas: []OSPFArea{
					{ID: "0", Type: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid interface network",
			config: OSPFConfig{
				RouterID: "10.0.0.1",
				Networks: []OSPFNetwork{
					{IP: "lan1", Area: "0"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOSPFConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOSPFConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidAreaID(t *testing.T) {
	tests := []struct {
		areaID string
		want   bool
	}{
		{"0", true},
		{"1", true},
		{"4294967295", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"10.0.0.1", true},
		{"invalid", false},
		{"-1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.areaID, func(t *testing.T) {
			if got := isValidAreaID(tt.areaID); got != tt.want {
				t.Errorf("isValidAreaID(%q) = %v, want %v", tt.areaID, got, tt.want)
			}
		})
	}
}
