package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestDebugSFTPParsing is a debug test to see actual config parsing results
// Run with: go test -v -run TestDebugSFTPParsing ./internal/rtx/parsers/
func TestDebugSFTPParsing(t *testing.T) {
	// Read actual config from environment or file
	configPath := os.Getenv("RTX_CONFIG_PATH")
	if configPath == "" {
		configPath = "/tmp/rtx_config.txt"
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Skipf("Config file not found at %s: %v", configPath, err)
		return
	}

	rawConfig := string(configData)

	// Parse the config
	parser := NewConfigFileParser()
	parsedConfig, err := parser.Parse(rawConfig)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Println("=== CONFIG FILE PARSER RESULTS ===")
	fmt.Printf("Line count: %d\n", parsedConfig.LineCount)
	fmt.Printf("Command count: %d\n", parsedConfig.CommandCount)
	fmt.Printf("Context count: %d\n", len(parsedConfig.Contexts))

	for i, ctx := range parsedConfig.Contexts {
		fmt.Printf("  Context[%d]: Type=%v, ID=%d, Name=%q\n", i, ctx.Type, ctx.ID, ctx.Name)
	}

	// Extract and print L2TP tunnels
	fmt.Println("\n=== L2TP TUNNELS ===")
	l2tpTunnels := parsedConfig.ExtractL2TPTunnels()
	for _, tunnel := range l2tpTunnels {
		data, _ := json.MarshalIndent(tunnel, "", "  ")
		fmt.Printf("Tunnel %d:\n%s\n\n", tunnel.ID, string(data))
	}

	// Extract and print L2TP service
	fmt.Println("\n=== L2TP SERVICE ===")
	l2tpService := parsedConfig.ExtractL2TPService()
	if l2tpService != nil {
		data, _ := json.MarshalIndent(l2tpService, "", "  ")
		fmt.Printf("%s\n", string(data))
	} else {
		fmt.Println("L2TP service: nil")
	}

	// Extract and print system config
	fmt.Println("\n=== SYSTEM CONFIG ===")
	systemConfig := parsedConfig.ExtractSystem()
	if systemConfig != nil {
		data, _ := json.MarshalIndent(systemConfig, "", "  ")
		fmt.Printf("%s\n", string(data))
	} else {
		fmt.Println("System config: nil")
	}

	// Extract and print DHCP bindings
	fmt.Println("\n=== DHCP BINDINGS ===")
	dhcpBindings := parsedConfig.ExtractDHCPBindings()
	for i, binding := range dhcpBindings {
		data, _ := json.MarshalIndent(binding, "", "  ")
		fmt.Printf("Binding %d:\n%s\n\n", i, string(data))
	}

	// Extract and print DHCP scopes
	fmt.Println("\n=== DHCP SCOPES ===")
	dhcpScopes := parsedConfig.ExtractDHCPScopes()
	for _, scope := range dhcpScopes {
		data, _ := json.MarshalIndent(scope, "", "  ")
		fmt.Printf("Scope %d:\n%s\n\n", scope.ScopeID, string(data))
	}

	// Print relevant raw config lines for debugging
	fmt.Println("\n=== RELEVANT CONFIG LINES ===")
	fmt.Println("\n--- L2TP related ---")
	for _, cmd := range parsedConfig.Commands {
		line := cmd.Line
		if containsAny(line, []string{"l2tp", "tunnel select", "tunnel encapsulation", "tunnel endpoint", "tunnel enable"}) {
			ctx := "global"
			if cmd.Context != nil {
				ctx = fmt.Sprintf("%s:%d", cmd.Context.Type, cmd.Context.ID)
			}
			fmt.Printf("[%s] %s\n", ctx, line)
		}
	}

	fmt.Println("\n--- System related ---")
	for _, cmd := range parsedConfig.GetGlobalCommands() {
		line := cmd.Line
		if containsAny(line, []string{"timezone", "console", "system packet-buffer", "statistics"}) {
			fmt.Printf("%s\n", line)
		}
	}

	fmt.Println("\n--- DHCP related ---")
	for _, cmd := range parsedConfig.GetGlobalCommands() {
		line := cmd.Line
		if containsAny(line, []string{"dhcp scope"}) {
			fmt.Printf("%s\n", line)
		}
	}
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
