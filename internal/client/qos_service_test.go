package client

import (
	"context"
	"testing"
)

// mockQoSExecutor is a mock executor for QoS tests
type mockQoSExecutor struct {
	commands []string
	outputs  map[string][]byte
	err      error
}

func (m *mockQoSExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.commands = append(m.commands, cmd)
	if m.err != nil {
		return nil, m.err
	}
	if output, ok := m.outputs[cmd]; ok {
		return output, nil
	}
	return []byte{}, nil
}

func (m *mockQoSExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	var allOutput []byte
	for _, cmd := range cmds {
		output, err := m.Run(ctx, cmd)
		if err != nil {
			return allOutput, err
		}
		allOutput = append(allOutput, output...)
	}
	return allOutput, nil
}

func (m *mockQoSExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return nil
}

func (m *mockQoSExecutor) SetLoginPassword(ctx context.Context, password string) error {
	return nil
}

func TestQoSService_CreateClassMap(t *testing.T) {
	tests := []struct {
		name    string
		cm      ClassMap
		wantErr bool
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
			name: "class map with filter",
			cm: ClassMap{
				Name:        "filtered-traffic",
				MatchFilter: 100,
			},
			wantErr: false,
		},
		{
			name: "invalid name",
			cm: ClassMap{
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockQoSExecutor{
				outputs: make(map[string][]byte),
			}
			service := NewQoSService(executor, nil)

			err := service.CreateClassMap(context.Background(), tt.cm)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClassMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQoSService_CreatePolicyMap(t *testing.T) {
	tests := []struct {
		name    string
		pm      PolicyMap
		wantErr bool
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
			name: "policy map with single class",
			pm: PolicyMap{
				Name: "simple-policy",
				Classes: []PolicyMapClass{
					{Name: "default", Priority: "normal"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - bandwidth exceeds 100%",
			pm: PolicyMap{
				Name: "invalid-policy",
				Classes: []PolicyMapClass{
					{Name: "class1", BandwidthPercent: 60},
					{Name: "class2", BandwidthPercent: 50},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - missing name",
			pm: PolicyMap{
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockQoSExecutor{
				outputs: make(map[string][]byte),
			}
			service := NewQoSService(executor, nil)

			err := service.CreatePolicyMap(context.Background(), tt.pm)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePolicyMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQoSService_CreateServicePolicy(t *testing.T) {
	tests := []struct {
		name    string
		sp      ServicePolicy
		wantErr bool
	}{
		{
			name: "valid service policy output",
			sp: ServicePolicy{
				Interface: "lan1",
				Direction: "output",
				PolicyMap: "priority",
			},
			wantErr: false,
		},
		{
			name: "valid service policy input",
			sp: ServicePolicy{
				Interface: "wan1",
				Direction: "input",
				PolicyMap: "cbq",
			},
			wantErr: false,
		},
		{
			name: "invalid direction",
			sp: ServicePolicy{
				Interface: "lan1",
				Direction: "both",
				PolicyMap: "priority",
			},
			wantErr: true,
		},
		{
			name: "missing interface",
			sp: ServicePolicy{
				Direction: "output",
				PolicyMap: "priority",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockQoSExecutor{
				outputs: make(map[string][]byte),
			}
			service := NewQoSService(executor, nil)

			err := service.CreateServicePolicy(context.Background(), tt.sp)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateServicePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQoSService_CreateShape(t *testing.T) {
	tests := []struct {
		name    string
		sc      ShapeConfig
		wantErr bool
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
			name: "shape config with burst",
			sc: ShapeConfig{
				Interface:    "wan1",
				Direction:    "output",
				ShapeAverage: 10000000,
				ShapeBurst:   1500,
			},
			wantErr: false,
		},
		{
			name: "invalid - zero shape average",
			sc: ShapeConfig{
				Interface:    "lan1",
				Direction:    "output",
				ShapeAverage: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid direction",
			sc: ShapeConfig{
				Interface:    "lan1",
				Direction:    "invalid",
				ShapeAverage: 1000000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockQoSExecutor{
				outputs: make(map[string][]byte),
			}
			service := NewQoSService(executor, nil)

			err := service.CreateShape(context.Background(), tt.sc)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShape() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQoSService_ListClassMaps(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: map[string][]byte{
			`show config | grep "queue\|speed"`: []byte(`queue lan1 type priority
queue lan1 class filter 1 100
queue lan1 class filter 2 200
queue lan1 class priority 1 high
queue lan1 class priority 2 normal`),
		},
	}
	service := NewQoSService(executor, nil)

	classMaps, err := service.ListClassMaps(context.Background())
	if err != nil {
		t.Fatalf("ListClassMaps() error = %v", err)
	}

	if len(classMaps) != 2 {
		t.Errorf("ListClassMaps() returned %d class maps, want 2", len(classMaps))
	}
}

func TestQoSService_ListServicePolicies(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: map[string][]byte{
			`show config | grep "queue\|speed"`: []byte(`queue lan1 type priority
queue lan2 type cbq`),
		},
	}
	service := NewQoSService(executor, nil)

	policies, err := service.ListServicePolicies(context.Background())
	if err != nil {
		t.Fatalf("ListServicePolicies() error = %v", err)
	}

	if len(policies) != 2 {
		t.Errorf("ListServicePolicies() returned %d policies, want 2", len(policies))
	}
}

func TestQoSService_ListShapes(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: map[string][]byte{
			`show config | grep "queue\|speed"`: []byte(`speed lan1 1000000
speed wan1 10000000`),
		},
	}
	service := NewQoSService(executor, nil)

	shapes, err := service.ListShapes(context.Background())
	if err != nil {
		t.Fatalf("ListShapes() error = %v", err)
	}

	if len(shapes) != 2 {
		t.Errorf("ListShapes() returned %d shapes, want 2", len(shapes))
	}

	// Verify shape values
	found := make(map[string]int)
	for _, s := range shapes {
		found[s.Interface] = s.ShapeAverage
	}

	if found["lan1"] != 1000000 {
		t.Errorf("lan1 shape = %d, want 1000000", found["lan1"])
	}
	if found["wan1"] != 10000000 {
		t.Errorf("wan1 shape = %d, want 10000000", found["wan1"])
	}
}

func TestQoSService_GetServicePolicy(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: map[string][]byte{
			`show config | grep "queue lan1\|speed lan1"`: []byte(`queue lan1 type priority`),
		},
	}
	service := NewQoSService(executor, nil)

	sp, err := service.GetServicePolicy(context.Background(), "lan1", "output")
	if err != nil {
		t.Fatalf("GetServicePolicy() error = %v", err)
	}

	if sp.Interface != "lan1" {
		t.Errorf("Interface = %q, want 'lan1'", sp.Interface)
	}
	if sp.PolicyMap != "priority" {
		t.Errorf("PolicyMap = %q, want 'priority'", sp.PolicyMap)
	}
}

func TestQoSService_GetShape(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: map[string][]byte{
			`show config | grep "queue lan1\|speed lan1"`: []byte(`speed lan1 5000000`),
		},
	}
	service := NewQoSService(executor, nil)

	sc, err := service.GetShape(context.Background(), "lan1", "output")
	if err != nil {
		t.Fatalf("GetShape() error = %v", err)
	}

	if sc.Interface != "lan1" {
		t.Errorf("Interface = %q, want 'lan1'", sc.Interface)
	}
	if sc.ShapeAverage != 5000000 {
		t.Errorf("ShapeAverage = %d, want 5000000", sc.ShapeAverage)
	}
}

func TestQoSService_DeleteServicePolicy(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: make(map[string][]byte),
	}
	service := NewQoSService(executor, nil)

	err := service.DeleteServicePolicy(context.Background(), "lan1", "output")
	if err != nil {
		t.Errorf("DeleteServicePolicy() error = %v", err)
	}

	// Verify command was sent
	if len(executor.commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(executor.commands))
	}
	if executor.commands[0] != "no queue lan1 type" {
		t.Errorf("Command = %q, want 'no queue lan1 type'", executor.commands[0])
	}
}

func TestQoSService_DeleteShape(t *testing.T) {
	executor := &mockQoSExecutor{
		outputs: make(map[string][]byte),
	}
	service := NewQoSService(executor, nil)

	err := service.DeleteShape(context.Background(), "lan1", "output")
	if err != nil {
		t.Errorf("DeleteShape() error = %v", err)
	}

	// Verify command was sent
	if len(executor.commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(executor.commands))
	}
	if executor.commands[0] != "no speed lan1" {
		t.Errorf("Command = %q, want 'no speed lan1'", executor.commands[0])
	}
}
