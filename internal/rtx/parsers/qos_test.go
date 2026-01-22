package parsers

import (
	"testing"
)

func TestParseQoSConfig(t *testing.T) {
	parser := NewQoSParser()

	tests := []struct {
		name        string
		raw         string
		iface       string
		wantType    string
		wantSpeed   int
		wantClasses int
		wantErr     bool
	}{
		{
			name: "basic priority queue",
			raw: `queue lan1 type priority
queue lan1 class filter 1 100
queue lan1 class priority 1 high
queue lan1 length 1 64`,
			iface:       "lan1",
			wantType:    "priority",
			wantClasses: 1,
			wantErr:     false,
		},
		{
			name: "multiple classes",
			raw: `queue lan1 type priority
queue lan1 class filter 1 100
queue lan1 class filter 2 200
queue lan1 class priority 1 high
queue lan1 class priority 2 normal
queue lan1 length 1 64
queue lan1 length 2 128`,
			iface:       "lan1",
			wantType:    "priority",
			wantClasses: 2,
			wantErr:     false,
		},
		{
			name: "with speed setting",
			raw: `queue lan1 type priority
queue lan1 class filter 1 100
speed lan1 1000000`,
			iface:       "lan1",
			wantType:    "priority",
			wantSpeed:   1000000,
			wantClasses: 1,
			wantErr:     false,
		},
		{
			name: "cbq queue type",
			raw: `queue lan2 type cbq
queue lan2 class filter 1 100`,
			iface:       "lan2",
			wantType:    "cbq",
			wantClasses: 1,
			wantErr:     false,
		},
		{
			name:     "empty config",
			raw:      "",
			iface:    "lan1",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "config for different interface",
			raw: `queue lan2 type priority
queue lan2 class filter 1 100`,
			iface:       "lan1",
			wantType:    "",
			wantClasses: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseQoSConfig(tt.raw, tt.iface)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQoSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if config.QueueType != tt.wantType {
				t.Errorf("QueueType = %q, want %q", config.QueueType, tt.wantType)
			}

			if config.Speed != tt.wantSpeed {
				t.Errorf("Speed = %d, want %d", config.Speed, tt.wantSpeed)
			}

			if len(config.Classes) != tt.wantClasses {
				t.Errorf("Classes count = %d, want %d", len(config.Classes), tt.wantClasses)
			}
		})
	}
}

func TestParseQoSConfigClasses(t *testing.T) {
	parser := NewQoSParser()

	raw := `queue lan1 type priority
queue lan1 class filter 1 100
queue lan1 class filter 2 200
queue lan1 class priority 1 high
queue lan1 class priority 2 low
queue lan1 length 1 64
queue lan1 length 2 32`

	config, err := parser.ParseQoSConfig(raw, "lan1")
	if err != nil {
		t.Fatalf("ParseQoSConfig() error = %v", err)
	}

	if len(config.Classes) != 2 {
		t.Fatalf("Expected 2 classes, got %d", len(config.Classes))
	}

	// Find class1 and class2
	var class1, class2 *QoSClass
	for i := range config.Classes {
		if config.Classes[i].Name == "class1" {
			class1 = &config.Classes[i]
		} else if config.Classes[i].Name == "class2" {
			class2 = &config.Classes[i]
		}
	}

	if class1 == nil {
		t.Fatal("class1 not found")
	}
	if class1.Filter != 100 {
		t.Errorf("class1.Filter = %d, want 100", class1.Filter)
	}
	if class1.Priority != "high" {
		t.Errorf("class1.Priority = %q, want 'high'", class1.Priority)
	}
	if class1.QueueLimit != 64 {
		t.Errorf("class1.QueueLimit = %d, want 64", class1.QueueLimit)
	}

	if class2 == nil {
		t.Fatal("class2 not found")
	}
	if class2.Filter != 200 {
		t.Errorf("class2.Filter = %d, want 200", class2.Filter)
	}
	if class2.Priority != "low" {
		t.Errorf("class2.Priority = %q, want 'low'", class2.Priority)
	}
	if class2.QueueLimit != 32 {
		t.Errorf("class2.QueueLimit = %d, want 32", class2.QueueLimit)
	}
}

func TestBuildQueueTypeCommand(t *testing.T) {
	tests := []struct {
		iface     string
		queueType string
		want      string
	}{
		{"lan1", "priority", "queue lan1 type priority"},
		{"lan2", "cbq", "queue lan2 type cbq"},
		{"wan1", "fifo", "queue wan1 type fifo"},
	}

	for _, tt := range tests {
		got := BuildQueueTypeCommand(tt.iface, tt.queueType)
		if got != tt.want {
			t.Errorf("BuildQueueTypeCommand(%q, %q) = %q, want %q", tt.iface, tt.queueType, got, tt.want)
		}
	}
}

func TestBuildQueueClassFilterCommand(t *testing.T) {
	tests := []struct {
		iface    string
		classNum int
		filter   int
		want     string
	}{
		{"lan1", 1, 100, "queue lan1 class filter 1 100"},
		{"lan2", 2, 200, "queue lan2 class filter 2 200"},
	}

	for _, tt := range tests {
		got := BuildQueueClassFilterCommand(tt.iface, tt.classNum, tt.filter)
		if got != tt.want {
			t.Errorf("BuildQueueClassFilterCommand(%q, %d, %d) = %q, want %q",
				tt.iface, tt.classNum, tt.filter, got, tt.want)
		}
	}
}

func TestBuildQueueClassPriorityCommand(t *testing.T) {
	tests := []struct {
		iface    string
		classNum int
		priority string
		want     string
	}{
		{"lan1", 1, "high", "queue lan1 class priority 1 high"},
		{"lan2", 2, "normal", "queue lan2 class priority 2 normal"},
		{"wan1", 3, "low", "queue wan1 class priority 3 low"},
	}

	for _, tt := range tests {
		got := BuildQueueClassPriorityCommand(tt.iface, tt.classNum, tt.priority)
		if got != tt.want {
			t.Errorf("BuildQueueClassPriorityCommand(%q, %d, %q) = %q, want %q",
				tt.iface, tt.classNum, tt.priority, got, tt.want)
		}
	}
}

func TestBuildSpeedCommand(t *testing.T) {
	tests := []struct {
		iface     string
		bandwidth int
		want      string
	}{
		{"lan1", 1000000, "speed lan1 1000000"},
		{"wan1", 100000000, "speed wan1 100000000"},
	}

	for _, tt := range tests {
		got := BuildSpeedCommand(tt.iface, tt.bandwidth)
		if got != tt.want {
			t.Errorf("BuildSpeedCommand(%q, %d) = %q, want %q", tt.iface, tt.bandwidth, got, tt.want)
		}
	}
}

func TestBuildQueueLengthCommand(t *testing.T) {
	tests := []struct {
		iface    string
		classNum int
		length   int
		want     string
	}{
		{"lan1", 1, 64, "queue lan1 length 1 64"},
		{"lan2", 2, 128, "queue lan2 length 2 128"},
	}

	for _, tt := range tests {
		got := BuildQueueLengthCommand(tt.iface, tt.classNum, tt.length)
		if got != tt.want {
			t.Errorf("BuildQueueLengthCommand(%q, %d, %d) = %q, want %q",
				tt.iface, tt.classNum, tt.length, got, tt.want)
		}
	}
}

func TestBuildQoSDeleteCommands(t *testing.T) {
	t.Run("DeleteQueueType", func(t *testing.T) {
		got := BuildDeleteQueueTypeCommand("lan1")
		want := "no queue lan1 type"
		if got != want {
			t.Errorf("BuildDeleteQueueTypeCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteQueueClassFilter", func(t *testing.T) {
		got := BuildDeleteQueueClassFilterCommand("lan1", 1)
		want := "no queue lan1 class filter 1"
		if got != want {
			t.Errorf("BuildDeleteQueueClassFilterCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteQueueClassPriority", func(t *testing.T) {
		got := BuildDeleteQueueClassPriorityCommand("lan1", 1)
		want := "no queue lan1 class priority 1"
		if got != want {
			t.Errorf("BuildDeleteQueueClassPriorityCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteSpeed", func(t *testing.T) {
		got := BuildDeleteSpeedCommand("lan1")
		want := "no speed lan1"
		if got != want {
			t.Errorf("BuildDeleteSpeedCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteQueueLength", func(t *testing.T) {
		got := BuildDeleteQueueLengthCommand("lan1", 1)
		want := "no queue lan1 length 1"
		if got != want {
			t.Errorf("BuildDeleteQueueLengthCommand() = %q, want %q", got, want)
		}
	})
}

func TestBuildDeleteQoSCommand(t *testing.T) {
	commands := BuildDeleteQoSCommand("lan1")
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
	if commands[0] != "no queue lan1 type" {
		t.Errorf("commands[0] = %q, want %q", commands[0], "no queue lan1 type")
	}
	if commands[1] != "no speed lan1" {
		t.Errorf("commands[1] = %q, want %q", commands[1], "no speed lan1")
	}
}

func TestBuildShowQoSCommand(t *testing.T) {
	got := BuildShowQoSCommand("lan1")
	want := `show config | grep "queue lan1\|speed lan1"`
	if got != want {
		t.Errorf("BuildShowQoSCommand() = %q, want %q", got, want)
	}
}

func TestValidateQoSConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  QoSConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: QoSConfig{
				Interface: "lan1",
				QueueType: "priority",
				Classes: []QoSClass{
					{Name: "class1", Priority: "high", BandwidthPercent: 30},
					{Name: "class2", Priority: "normal", BandwidthPercent: 70},
				},
			},
			wantErr: false,
		},
		{
			name: "missing interface",
			config: QoSConfig{
				QueueType: "priority",
			},
			wantErr: true,
			errMsg:  "interface is required",
		},
		{
			name: "invalid queue type",
			config: QoSConfig{
				Interface: "lan1",
				QueueType: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid queue type",
		},
		{
			name: "invalid priority",
			config: QoSConfig{
				Interface: "lan1",
				Classes: []QoSClass{
					{Name: "class1", Priority: "invalid"},
				},
			},
			wantErr: true,
			errMsg:  "invalid priority",
		},
		{
			name: "bandwidth exceeds 100%",
			config: QoSConfig{
				Interface: "lan1",
				Classes: []QoSClass{
					{Name: "class1", BandwidthPercent: 60},
					{Name: "class2", BandwidthPercent: 50},
				},
			},
			wantErr: true,
			errMsg:  "exceeds 100%",
		},
		{
			name: "negative speed",
			config: QoSConfig{
				Interface: "lan1",
				Speed:     -1,
			},
			wantErr: true,
			errMsg:  "non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQoSConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQoSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !qosContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateQoSConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateClassMap(t *testing.T) {
	tests := []struct {
		name    string
		cm      ClassMap
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid class map",
			cm: ClassMap{
				Name:                 "voip-traffic",
				MatchProtocol:        "sip",
				MatchDestinationPort: []int{5060},
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			cm:      ClassMap{},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "invalid name format",
			cm: ClassMap{
				Name: "123-invalid",
			},
			wantErr: true,
			errMsg:  "must start with a letter",
		},
		{
			name: "invalid port",
			cm: ClassMap{
				Name:                 "test",
				MatchDestinationPort: []int{70000},
			},
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClassMap(tt.cm)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClassMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !qosContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateClassMap() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidatePolicyMap(t *testing.T) {
	tests := []struct {
		name    string
		pm      PolicyMap
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid policy map",
			pm: PolicyMap{
				Name: "qos-policy",
				Classes: []PolicyMapClass{
					{Name: "voip", Priority: "high", BandwidthPercent: 30},
					{Name: "data", Priority: "normal", BandwidthPercent: 70},
				},
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			pm:      PolicyMap{},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "class missing name",
			pm: PolicyMap{
				Name: "test",
				Classes: []PolicyMapClass{
					{Priority: "high"},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "bandwidth exceeds 100%",
			pm: PolicyMap{
				Name: "test",
				Classes: []PolicyMapClass{
					{Name: "class1", BandwidthPercent: 60},
					{Name: "class2", BandwidthPercent: 50},
				},
			},
			wantErr: true,
			errMsg:  "exceeds 100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePolicyMap(tt.pm)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePolicyMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !qosContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePolicyMap() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateServicePolicy(t *testing.T) {
	tests := []struct {
		name    string
		sp      ServicePolicy
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid service policy output",
			sp: ServicePolicy{
				Interface: "lan1",
				Direction: "output",
				PolicyMap: "qos-policy",
			},
			wantErr: false,
		},
		{
			name: "valid service policy input",
			sp: ServicePolicy{
				Interface: "wan1",
				Direction: "input",
				PolicyMap: "ingress-policy",
			},
			wantErr: false,
		},
		{
			name: "missing interface",
			sp: ServicePolicy{
				Direction: "output",
				PolicyMap: "test",
			},
			wantErr: true,
			errMsg:  "interface is required",
		},
		{
			name: "invalid direction",
			sp: ServicePolicy{
				Interface: "lan1",
				Direction: "both",
				PolicyMap: "test",
			},
			wantErr: true,
			errMsg:  "must be 'input' or 'output'",
		},
		{
			name: "missing policy map",
			sp: ServicePolicy{
				Interface: "lan1",
				Direction: "output",
			},
			wantErr: true,
			errMsg:  "policy_map is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServicePolicy(tt.sp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServicePolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !qosContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateServicePolicy() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateShapeConfig(t *testing.T) {
	tests := []struct {
		name    string
		sc      ShapeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid shape config",
			sc: ShapeConfig{
				Interface:    "lan1",
				Direction:    "output",
				ShapeAverage: 1000000,
			},
			wantErr: false,
		},
		{
			name: "with burst",
			sc: ShapeConfig{
				Interface:    "wan1",
				Direction:    "output",
				ShapeAverage: 10000000,
				ShapeBurst:   1500,
			},
			wantErr: false,
		},
		{
			name: "missing interface",
			sc: ShapeConfig{
				Direction:    "output",
				ShapeAverage: 1000000,
			},
			wantErr: true,
			errMsg:  "interface is required",
		},
		{
			name: "invalid direction",
			sc: ShapeConfig{
				Interface:    "lan1",
				Direction:    "invalid",
				ShapeAverage: 1000000,
			},
			wantErr: true,
			errMsg:  "must be 'input' or 'output'",
		},
		{
			name: "zero shape average",
			sc: ShapeConfig{
				Interface:    "lan1",
				Direction:    "output",
				ShapeAverage: 0,
			},
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name: "negative burst",
			sc: ShapeConfig{
				Interface:    "lan1",
				Direction:    "output",
				ShapeAverage: 1000000,
				ShapeBurst:   -1,
			},
			wantErr: true,
			errMsg:  "must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShapeConfig(tt.sc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateShapeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !qosContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateShapeConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseServicePolicy(t *testing.T) {
	parser := NewQoSParser()

	tests := []struct {
		name       string
		raw        string
		iface      string
		wantPolicy string
		wantErr    bool
	}{
		{
			name:       "found policy",
			raw:        "queue lan1 type priority",
			iface:      "lan1",
			wantPolicy: "priority",
			wantErr:    false,
		},
		{
			name:       "not found",
			raw:        "queue lan2 type priority",
			iface:      "lan1",
			wantPolicy: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp, err := parser.ParseServicePolicy(tt.raw, tt.iface)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseServicePolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && sp.PolicyMap != tt.wantPolicy {
				t.Errorf("PolicyMap = %q, want %q", sp.PolicyMap, tt.wantPolicy)
			}
		})
	}
}

func TestParseShapeConfig(t *testing.T) {
	parser := NewQoSParser()

	tests := []struct {
		name      string
		raw       string
		iface     string
		wantSpeed int
		wantErr   bool
	}{
		{
			name:      "found speed",
			raw:       "speed lan1 1000000",
			iface:     "lan1",
			wantSpeed: 1000000,
			wantErr:   false,
		},
		{
			name:      "not found",
			raw:       "speed lan2 1000000",
			iface:     "lan1",
			wantSpeed: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc, err := parser.ParseShapeConfig(tt.raw, tt.iface)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseShapeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && sc.ShapeAverage != tt.wantSpeed {
				t.Errorf("ShapeAverage = %d, want %d", sc.ShapeAverage, tt.wantSpeed)
			}
		})
	}
}

// Helper function for string containment check
func qosContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
