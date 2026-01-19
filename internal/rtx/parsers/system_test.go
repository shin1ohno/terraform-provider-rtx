package parsers

import (
	"testing"
)

func TestParseSystemConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SystemConfig
		wantErr  bool
	}{
		{
			name:  "timezone only",
			input: `timezone +09:00`,
			expected: &SystemConfig{
				Timezone:      "+09:00",
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name:  "negative timezone",
			input: `timezone -05:00`,
			expected: &SystemConfig{
				Timezone:      "-05:00",
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "console character only",
			input: `console character ja.utf8`,
			expected: &SystemConfig{
				Console: &ConsoleConfig{
					Character: "ja.utf8",
				},
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "console lines infinity",
			input: `console lines infinity`,
			expected: &SystemConfig{
				Console: &ConsoleConfig{
					Lines: "infinity",
				},
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "console lines number",
			input: `console lines 24`,
			expected: &SystemConfig{
				Console: &ConsoleConfig{
					Lines: "24",
				},
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "console prompt simple",
			input: `console prompt RTX1210`,
			expected: &SystemConfig{
				Console: &ConsoleConfig{
					Prompt: "RTX1210",
				},
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "console prompt with spaces",
			input: `console prompt "[RTX1210] "`,
			expected: &SystemConfig{
				Console: &ConsoleConfig{
					Prompt: "[RTX1210] ",
				},
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "full console configuration",
			input: `console character ja.utf8
console lines infinity
console prompt "[RTX1210] "`,
			expected: &SystemConfig{
				Console: &ConsoleConfig{
					Character: "ja.utf8",
					Lines:     "infinity",
					Prompt:    "[RTX1210] ",
				},
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name: "packet buffer small",
			input: `system packet-buffer small max-buffer=5000 max-free=1300`,
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 5000, MaxFree: 1300},
				},
			},
		},
		{
			name: "packet buffer all sizes",
			input: `system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=10000 max-free=4950
system packet-buffer large max-buffer=20000 max-free=5600`,
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 5000, MaxFree: 1300},
					{Size: "middle", MaxBuffer: 10000, MaxFree: 4950},
					{Size: "large", MaxBuffer: 20000, MaxFree: 5600},
				},
			},
		},
		{
			name: "statistics traffic on",
			input: `statistics traffic on`,
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{},
				Statistics: &StatisticsConfig{
					Traffic: true,
				},
			},
		},
		{
			name: "statistics nat off",
			input: `statistics nat off`,
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{},
				Statistics: &StatisticsConfig{
					NAT: false,
				},
			},
		},
		{
			name: "statistics both enabled",
			input: `statistics traffic on
statistics nat on`,
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{},
				Statistics: &StatisticsConfig{
					Traffic: true,
					NAT:     true,
				},
			},
		},
		{
			name: "full configuration",
			input: `timezone +09:00
console character ja.utf8
console lines infinity
console prompt "[RTX1210] "
system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=10000 max-free=4950
system packet-buffer large max-buffer=20000 max-free=5600
statistics traffic on
statistics nat on`,
			expected: &SystemConfig{
				Timezone: "+09:00",
				Console: &ConsoleConfig{
					Character: "ja.utf8",
					Lines:     "infinity",
					Prompt:    "[RTX1210] ",
				},
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 5000, MaxFree: 1300},
					{Size: "middle", MaxBuffer: 10000, MaxFree: 4950},
					{Size: "large", MaxBuffer: 20000, MaxFree: 5600},
				},
				Statistics: &StatisticsConfig{
					Traffic: true,
					NAT:     true,
				},
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{},
			},
		},
		{
			name:  "whitespace only",
			input: "   \n\n   \n",
			expected: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{},
			},
		},
	}

	parser := NewSystemParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseSystemConfig(tt.input)

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

			// Check timezone
			if result.Timezone != tt.expected.Timezone {
				t.Errorf("timezone = %q, want %q", result.Timezone, tt.expected.Timezone)
			}

			// Check console
			if tt.expected.Console == nil {
				if result.Console != nil {
					t.Errorf("console = %+v, want nil", result.Console)
				}
			} else {
				if result.Console == nil {
					t.Errorf("console = nil, want %+v", tt.expected.Console)
				} else {
					if result.Console.Character != tt.expected.Console.Character {
						t.Errorf("console.character = %q, want %q", result.Console.Character, tt.expected.Console.Character)
					}
					if result.Console.Lines != tt.expected.Console.Lines {
						t.Errorf("console.lines = %q, want %q", result.Console.Lines, tt.expected.Console.Lines)
					}
					if result.Console.Prompt != tt.expected.Console.Prompt {
						t.Errorf("console.prompt = %q, want %q", result.Console.Prompt, tt.expected.Console.Prompt)
					}
				}
			}

			// Check packet buffers
			if len(result.PacketBuffers) != len(tt.expected.PacketBuffers) {
				t.Errorf("packet_buffers count = %d, want %d", len(result.PacketBuffers), len(tt.expected.PacketBuffers))
			} else {
				for i, pb := range result.PacketBuffers {
					if pb.Size != tt.expected.PacketBuffers[i].Size {
						t.Errorf("packet_buffers[%d].size = %q, want %q", i, pb.Size, tt.expected.PacketBuffers[i].Size)
					}
					if pb.MaxBuffer != tt.expected.PacketBuffers[i].MaxBuffer {
						t.Errorf("packet_buffers[%d].max_buffer = %d, want %d", i, pb.MaxBuffer, tt.expected.PacketBuffers[i].MaxBuffer)
					}
					if pb.MaxFree != tt.expected.PacketBuffers[i].MaxFree {
						t.Errorf("packet_buffers[%d].max_free = %d, want %d", i, pb.MaxFree, tt.expected.PacketBuffers[i].MaxFree)
					}
				}
			}

			// Check statistics
			if tt.expected.Statistics == nil {
				if result.Statistics != nil {
					t.Errorf("statistics = %+v, want nil", result.Statistics)
				}
			} else {
				if result.Statistics == nil {
					t.Errorf("statistics = nil, want %+v", tt.expected.Statistics)
				} else {
					if result.Statistics.Traffic != tt.expected.Statistics.Traffic {
						t.Errorf("statistics.traffic = %v, want %v", result.Statistics.Traffic, tt.expected.Statistics.Traffic)
					}
					if result.Statistics.NAT != tt.expected.Statistics.NAT {
						t.Errorf("statistics.nat = %v, want %v", result.Statistics.NAT, tt.expected.Statistics.NAT)
					}
				}
			}
		})
	}
}

func TestBuildTimezoneCommand(t *testing.T) {
	tests := []struct {
		name     string
		tz       string
		expected string
	}{
		{
			name:     "positive timezone",
			tz:       "+09:00",
			expected: "timezone +09:00",
		},
		{
			name:     "negative timezone",
			tz:       "-05:00",
			expected: "timezone -05:00",
		},
		{
			name:     "UTC",
			tz:       "+00:00",
			expected: "timezone +00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildTimezoneCommand(tt.tz)
			if result != tt.expected {
				t.Errorf("BuildTimezoneCommand(%q) = %q, want %q", tt.tz, result, tt.expected)
			}
		})
	}
}

func TestBuildConsoleCommands(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "character ja.utf8",
			fn:       func() string { return BuildConsoleCharacterCommand("ja.utf8") },
			expected: "console character ja.utf8",
		},
		{
			name:     "lines infinity",
			fn:       func() string { return BuildConsoleLinesCommand("infinity") },
			expected: "console lines infinity",
		},
		{
			name:     "lines 24",
			fn:       func() string { return BuildConsoleLinesCommand("24") },
			expected: "console lines 24",
		},
		{
			name:     "prompt simple",
			fn:       func() string { return BuildConsolePromptCommand("RTX1210") },
			expected: "console prompt RTX1210",
		},
		{
			name:     "prompt with spaces",
			fn:       func() string { return BuildConsolePromptCommand("[RTX1210] ") },
			expected: `console prompt "[RTX1210] "`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildPacketBufferCommand(t *testing.T) {
	tests := []struct {
		name     string
		config   PacketBufferConfig
		expected string
	}{
		{
			name:     "small buffer",
			config:   PacketBufferConfig{Size: "small", MaxBuffer: 5000, MaxFree: 1300},
			expected: "system packet-buffer small max-buffer=5000 max-free=1300",
		},
		{
			name:     "middle buffer",
			config:   PacketBufferConfig{Size: "middle", MaxBuffer: 10000, MaxFree: 4950},
			expected: "system packet-buffer middle max-buffer=10000 max-free=4950",
		},
		{
			name:     "large buffer",
			config:   PacketBufferConfig{Size: "large", MaxBuffer: 20000, MaxFree: 5600},
			expected: "system packet-buffer large max-buffer=20000 max-free=5600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPacketBufferCommand(tt.config)
			if result != tt.expected {
				t.Errorf("BuildPacketBufferCommand(%+v) = %q, want %q", tt.config, result, tt.expected)
			}
		})
	}
}

func TestBuildStatisticsCommands(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "traffic on",
			fn:       func() string { return BuildStatisticsTrafficCommand(true) },
			expected: "statistics traffic on",
		},
		{
			name:     "traffic off",
			fn:       func() string { return BuildStatisticsTrafficCommand(false) },
			expected: "statistics traffic off",
		},
		{
			name:     "nat on",
			fn:       func() string { return BuildStatisticsNATCommand(true) },
			expected: "statistics nat on",
		},
		{
			name:     "nat off",
			fn:       func() string { return BuildStatisticsNATCommand(false) },
			expected: "statistics nat off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteCommands(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "delete timezone",
			fn:       BuildDeleteTimezoneCommand,
			expected: "no timezone",
		},
		{
			name:     "delete console character",
			fn:       BuildDeleteConsoleCharacterCommand,
			expected: "no console character",
		},
		{
			name:     "delete console lines",
			fn:       BuildDeleteConsoleLinesCommand,
			expected: "no console lines",
		},
		{
			name:     "delete console prompt",
			fn:       BuildDeleteConsolePromptCommand,
			expected: "no console prompt",
		},
		{
			name:     "delete packet buffer small",
			fn:       func() string { return BuildDeletePacketBufferCommand("small") },
			expected: "no system packet-buffer small",
		},
		{
			name:     "delete statistics traffic",
			fn:       BuildDeleteStatisticsTrafficCommand,
			expected: "no statistics traffic",
		},
		{
			name:     "delete statistics nat",
			fn:       BuildDeleteStatisticsNATCommand,
			expected: "no statistics nat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSystemCommands(t *testing.T) {
	config := &SystemConfig{
		Timezone: "+09:00",
		Console: &ConsoleConfig{
			Character: "ja.utf8",
			Lines:     "infinity",
			Prompt:    "RTX1210",
		},
		PacketBuffers: []PacketBufferConfig{
			{Size: "small", MaxBuffer: 5000, MaxFree: 1300},
			{Size: "large", MaxBuffer: 20000, MaxFree: 5600},
		},
		Statistics: &StatisticsConfig{
			Traffic: true,
			NAT:     true,
		},
	}

	commands := BuildDeleteSystemCommands(config)

	expected := []string{
		"no timezone",
		"no console character",
		"no console lines",
		"no console prompt",
		"no system packet-buffer small",
		"no system packet-buffer large",
		"no statistics traffic",
		"no statistics nat",
	}

	if len(commands) != len(expected) {
		t.Errorf("command count = %d, want %d", len(commands), len(expected))
		return
	}

	for i, cmd := range commands {
		if cmd != expected[i] {
			t.Errorf("commands[%d] = %q, want %q", i, cmd, expected[i])
		}
	}
}

func TestBuildShowSystemConfigCommand(t *testing.T) {
	result := BuildShowSystemConfigCommand()
	expected := `show config | grep -E "(timezone|console|packet-buffer|statistics)"`
	if result != expected {
		t.Errorf("BuildShowSystemConfigCommand() = %q, want %q", result, expected)
	}
}

func TestValidateSystemConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *SystemConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid full config",
			config: &SystemConfig{
				Timezone: "+09:00",
				Console: &ConsoleConfig{
					Character: "ja.utf8",
					Lines:     "infinity",
					Prompt:    "RTX1210",
				},
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 5000, MaxFree: 1300},
				},
				Statistics: &StatisticsConfig{
					Traffic: true,
					NAT:     true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid empty config",
			config: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{},
			},
			wantErr: false,
		},
		{
			name: "invalid timezone format - no sign",
			config: &SystemConfig{
				Timezone: "09:00",
			},
			wantErr: true,
			errMsg:  "invalid timezone format",
		},
		{
			name: "invalid timezone format - no colon",
			config: &SystemConfig{
				Timezone: "+0900",
			},
			wantErr: true,
			errMsg:  "invalid timezone format",
		},
		{
			name: "invalid character encoding",
			config: &SystemConfig{
				Console: &ConsoleConfig{
					Character: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid character encoding",
		},
		{
			name: "invalid console lines - negative",
			config: &SystemConfig{
				Console: &ConsoleConfig{
					Lines: "-1",
				},
			},
			wantErr: true,
			errMsg:  "invalid console lines",
		},
		{
			name: "invalid console lines - text",
			config: &SystemConfig{
				Console: &ConsoleConfig{
					Lines: "abc",
				},
			},
			wantErr: true,
			errMsg:  "invalid console lines",
		},
		{
			name: "invalid packet buffer size",
			config: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{
					{Size: "invalid", MaxBuffer: 5000, MaxFree: 1300},
				},
			},
			wantErr: true,
			errMsg:  "invalid packet buffer size",
		},
		{
			name: "invalid packet buffer max_buffer",
			config: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 0, MaxFree: 1300},
				},
			},
			wantErr: true,
			errMsg:  "max_buffer must be positive",
		},
		{
			name: "invalid packet buffer max_free",
			config: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 5000, MaxFree: 0},
				},
			},
			wantErr: true,
			errMsg:  "max_free must be positive",
		},
		{
			name: "max_free exceeds max_buffer",
			config: &SystemConfig{
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 1000, MaxFree: 2000},
				},
			},
			wantErr: true,
			errMsg:  "max_free cannot exceed max_buffer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSystemConfig(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
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

func TestIsValidTimezone(t *testing.T) {
	tests := []struct {
		tz    string
		valid bool
	}{
		{"+09:00", true},
		{"-05:00", true},
		{"+00:00", true},
		{"-12:00", true},
		{"+14:00", true},
		{"09:00", false},
		{"+0900", false},
		{"+9:00", false},
		{"", false},
		{"JST", false},
	}

	for _, tt := range tests {
		t.Run(tt.tz, func(t *testing.T) {
			result := isValidTimezone(tt.tz)
			if result != tt.valid {
				t.Errorf("isValidTimezone(%q) = %v, want %v", tt.tz, result, tt.valid)
			}
		})
	}
}

func TestIsValidCharacterEncoding(t *testing.T) {
	tests := []struct {
		encoding string
		valid    bool
	}{
		{"ja.utf8", true},
		{"ja.sjis", true},
		{"ascii", true},
		{"euc-jp", true},
		{"utf-8", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.encoding, func(t *testing.T) {
			result := isValidCharacterEncoding(tt.encoding)
			if result != tt.valid {
				t.Errorf("isValidCharacterEncoding(%q) = %v, want %v", tt.encoding, result, tt.valid)
			}
		})
	}
}

func TestIsValidConsoleLines(t *testing.T) {
	tests := []struct {
		lines string
		valid bool
	}{
		{"infinity", true},
		{"24", true},
		{"1", true},
		{"100", true},
		{"0", false},
		{"-1", false},
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.lines, func(t *testing.T) {
			result := isValidConsoleLines(tt.lines)
			if result != tt.valid {
				t.Errorf("isValidConsoleLines(%q) = %v, want %v", tt.lines, result, tt.valid)
			}
		})
	}
}

func TestIsValidPacketBufferSize(t *testing.T) {
	tests := []struct {
		size  string
		valid bool
	}{
		{"small", true},
		{"middle", true},
		{"large", true},
		{"medium", false},
		{"xlarge", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			result := isValidPacketBufferSize(tt.size)
			if result != tt.valid {
				t.Errorf("isValidPacketBufferSize(%q) = %v, want %v", tt.size, result, tt.valid)
			}
		})
	}
}

// containsStr is a helper function to check if s contains substr
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStrHelper(s, substr))
}

func containsStrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
