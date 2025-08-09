package parsers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInterfacesParsers(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		dataFile  string
		wantCount int
		validate  func(t *testing.T, interfaces []Interface)
	}{
		{
			name:      "RTX830 interfaces",
			model:     "RTX830",
			dataFile:  "../testdata/RTX830/show_interface.txt",
			wantCount: 5,
			validate: func(t *testing.T, interfaces []Interface) {
				// Validate LAN1
				lan1 := findInterface(interfaces, "LAN1")
				if lan1 == nil {
					t.Fatal("LAN1 not found")
				}
				if lan1.Kind != "lan" {
					t.Errorf("LAN1 kind = %s, want lan", lan1.Kind)
				}
				if !lan1.LinkUp {
					t.Error("LAN1 should be up")
				}
				if lan1.IPv4 != "192.168.1.254/24" {
					t.Errorf("LAN1 IPv4 = %s, want 192.168.1.254/24", lan1.IPv4)
				}
				if lan1.MAC != "00:A0:DE:12:34:56" {
					t.Errorf("LAN1 MAC = %s, want 00:A0:DE:12:34:56", lan1.MAC)
				}
				
				// Validate LAN2
				lan2 := findInterface(interfaces, "LAN2")
				if lan2 == nil {
					t.Fatal("LAN2 not found")
				}
				if lan2.LinkUp {
					t.Error("LAN2 should be down")
				}
				
				// Validate PP1
				pp1 := findInterface(interfaces, "PP1")
				if pp1 == nil {
					t.Fatal("PP1 not found")
				}
				if pp1.Kind != "pp" {
					t.Errorf("PP1 kind = %s, want pp", pp1.Kind)
				}
			},
		},
		{
			name:      "RTX1210 interfaces",
			model:     "RTX1210",
			dataFile:  "../testdata/RTX1210/show_interface.txt",
			wantCount: 5,
			validate: func(t *testing.T, interfaces []Interface) {
				// Validate LAN1
				lan1 := findInterface(interfaces, "LAN1")
				if lan1 == nil {
					t.Fatal("LAN1 not found")
				}
				if !lan1.LinkUp {
					t.Error("LAN1 should be up")
				}
				if lan1.IPv4 != "192.168.1.254/24" {
					t.Errorf("LAN1 IPv4 = %s, want 192.168.1.254/24", lan1.IPv4)
				}
				if lan1.MAC != "00:A0:DE:AB:CD:01" {
					t.Errorf("LAN1 MAC = %s, want 00:A0:DE:AB:CD:01", lan1.MAC)
				}
				if lan1.MTU != 1500 {
					t.Errorf("LAN1 MTU = %d, want 1500", lan1.MTU)
				}
				
				// Validate WAN1 with IPv6
				wan1 := findInterface(interfaces, "WAN1")
				if wan1 == nil {
					t.Fatal("WAN1 not found")
				}
				if wan1.IPv6 != "2001:db8::1/64" {
					t.Errorf("WAN1 IPv6 = %s, want 2001:db8::1/64", wan1.IPv6)
				}
				
				// Validate PP1
				pp1 := findInterface(interfaces, "PP1")
				if pp1 == nil {
					t.Fatal("PP1 not found")
				}
				if pp1.MTU != 1454 {
					t.Errorf("PP1 MTU = %d, want 1454", pp1.MTU)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test data
			data, err := os.ReadFile(tt.dataFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}
			
			// Get parser
			parser, err := Get("interfaces", tt.model)
			if err != nil {
				t.Fatalf("failed to get parser: %v", err)
			}
			
			// Parse
			result, err := parser.Parse(string(data))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			
			// Type assert
			interfaces, ok := result.([]Interface)
			if !ok {
				t.Fatalf("expected []Interface, got %T", result)
			}
			
			// Check count
			if len(interfaces) != tt.wantCount {
				t.Errorf("got %d interfaces, want %d", len(interfaces), tt.wantCount)
			}
			
			// Validate
			if tt.validate != nil {
				tt.validate(t, interfaces)
			}
		})
	}
}

func TestInterfaceKind(t *testing.T) {
	tests := []struct {
		name     string
		wantKind string
	}{
		{"LAN1", "lan"},
		{"LAN2", "lan"},
		{"WAN1", "wan"},
		{"WAN2", "wan"},
		{"PP1", "pp"},
		{"PP10", "pp"},
		{"VLAN10", "vlan"},
		{"VLAN10.20", "vlan"},
		{"unknown", "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInterfaceKind(tt.name)
			if got != tt.wantKind {
				t.Errorf("getInterfaceKind(%s) = %s, want %s", tt.name, got, tt.wantKind)
			}
		})
	}
}

func TestGoldenFiles(t *testing.T) {
	models := []string{"RTX830", "RTX1210"}
	
	for _, model := range models {
		t.Run(model+" golden", func(t *testing.T) {
			// Read input
			inputPath := filepath.Join("..", "testdata", model, "show_interface.txt")
			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("failed to read input: %v", err)
			}
			
			// Get parser
			parser, err := Get("interfaces", model)
			if err != nil {
				t.Fatalf("failed to get parser: %v", err)
			}
			
			// Parse
			result, err := parser.Parse(string(input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			
			// Convert to JSON for comparison
			got, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal result: %v", err)
			}
			
			// Check golden file
			goldenPath := filepath.Join("..", "testdata", model, "show_interface.golden.json")
			
			if update := os.Getenv("UPDATE_GOLDEN"); update == "true" {
				// Update golden file
				err = os.WriteFile(goldenPath, got, 0644)
				if err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Log("Updated golden file")
			} else {
				// Compare with golden file
				want, err := os.ReadFile(goldenPath)
				if err != nil {
					// Create initial golden file
					err = os.WriteFile(goldenPath, got, 0644)
					if err != nil {
						t.Fatalf("failed to create golden file: %v", err)
					}
					t.Log("Created golden file")
					return
				}
				
				if string(got) != string(want) {
					t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", string(got), string(want))
				}
			}
		})
	}
}

func findInterface(interfaces []Interface, name string) *Interface {
	for i := range interfaces {
		if interfaces[i].Name == name {
			return &interfaces[i]
		}
	}
	return nil
}

func TestParserCanHandle(t *testing.T) {
	tests := []struct {
		parser    InterfacesParser
		model     string
		canHandle bool
	}{
		{&rtx830InterfacesParser{}, "RTX830", true},
		{&rtx830InterfacesParser{}, "RTX1210", false},
		{&rtx12xxInterfacesParser{}, "RTX1210", true},
		{&rtx12xxInterfacesParser{}, "RTX1220", true},
		{&rtx12xxInterfacesParser{}, "RTX830", false},
	}
	
	for _, tt := range tests {
		name := reflect.TypeOf(tt.parser).Elem().Name() + "/" + tt.model
		t.Run(name, func(t *testing.T) {
			got := tt.parser.CanHandle(tt.model)
			if got != tt.canHandle {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.model, got, tt.canHandle)
			}
		})
	}
}