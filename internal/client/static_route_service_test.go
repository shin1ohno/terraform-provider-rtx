package client

import (
	"context"
	"testing"
)

// mockExecutor implements Executor interface for testing
type mockStaticRouteExecutor struct {
	responses map[string][]byte
}

func (m *mockStaticRouteExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	if response, ok := m.responses[cmd]; ok {
		return response, nil
	}
	return []byte{}, nil
}

func (m *mockStaticRouteExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
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

func (m *mockStaticRouteExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return nil
}

func (m *mockStaticRouteExecutor) SetLoginPassword(ctx context.Context, password string) error {
	return nil
}

func TestStaticRouteService_GetRoute_MultiGateway(t *testing.T) {
	// This test verifies that GetRoute correctly captures multiple gateways
	// for the same prefix/mask when the grep command returns multiple lines.
	//
	// Bug scenario (REQ-3): Routes with gateways 192.168.1.20 and 192.168.1.21
	// only import first gateway.

	mockExec := &mockStaticRouteExecutor{
		responses: map[string][]byte{
			// Simulate RTX output with multiple gateway lines for same prefix
			`show config | grep "ip route 10.33.128.0/21"`: []byte(
				"ip route 10.33.128.0/21 gateway 192.168.1.20\n" +
					"ip route 10.33.128.0/21 gateway 192.168.1.21\n",
			),
			"save": []byte(""),
		},
	}

	service := NewStaticRouteService(mockExec, nil)

	route, err := service.GetRoute(context.Background(), "10.33.128.0", "255.255.248.0")
	if err != nil {
		t.Fatalf("GetRoute failed: %v", err)
	}

	// Verify both gateways are captured
	if len(route.NextHops) != 2 {
		t.Fatalf("expected 2 next_hops, got %d", len(route.NextHops))
	}

	expectedGateways := []string{"192.168.1.20", "192.168.1.21"}
	for i, expected := range expectedGateways {
		if route.NextHops[i].NextHop != expected {
			t.Errorf("next_hops[%d].NextHop = %q, want %q", i, route.NextHops[i].NextHop, expected)
		}
	}
}

func TestStaticRouteService_GetRoute_MultiGatewayWithWeight(t *testing.T) {
	// Test multi-gateway with different weight values
	mockExec := &mockStaticRouteExecutor{
		responses: map[string][]byte{
			`show config | grep "ip route 192.168.100.0/24"`: []byte(
				"ip route 192.168.100.0/24 gateway 10.0.0.1 weight 1\n" +
					"ip route 192.168.100.0/24 gateway 10.0.0.2 weight 2\n" +
					"ip route 192.168.100.0/24 gateway 10.0.0.3 weight 3\n",
			),
			"save": []byte(""),
		},
	}

	service := NewStaticRouteService(mockExec, nil)

	route, err := service.GetRoute(context.Background(), "192.168.100.0", "255.255.255.0")
	if err != nil {
		t.Fatalf("GetRoute failed: %v", err)
	}

	// Verify all 3 gateways are captured
	if len(route.NextHops) != 3 {
		t.Fatalf("expected 3 next_hops, got %d", len(route.NextHops))
	}

	expectedHops := []struct {
		gateway  string
		distance int
	}{
		{"10.0.0.1", 1},
		{"10.0.0.2", 2},
		{"10.0.0.3", 3},
	}

	for i, expected := range expectedHops {
		if route.NextHops[i].NextHop != expected.gateway {
			t.Errorf("next_hops[%d].NextHop = %q, want %q", i, route.NextHops[i].NextHop, expected.gateway)
		}
		if route.NextHops[i].Distance != expected.distance {
			t.Errorf("next_hops[%d].Distance = %d, want %d", i, route.NextHops[i].Distance, expected.distance)
		}
	}
}

func TestStaticRouteService_GetRoute_MultiGatewayWithFilter(t *testing.T) {
	// Test multi-gateway with filter attribute
	mockExec := &mockStaticRouteExecutor{
		responses: map[string][]byte{
			`show config | grep "ip route 10.0.0.0/8"`: []byte(
				"ip route 10.0.0.0/8 gateway 192.168.1.1 filter 100\n" +
					"ip route 10.0.0.0/8 gateway 192.168.1.2 filter 200\n",
			),
			"save": []byte(""),
		},
	}

	service := NewStaticRouteService(mockExec, nil)

	route, err := service.GetRoute(context.Background(), "10.0.0.0", "255.0.0.0")
	if err != nil {
		t.Fatalf("GetRoute failed: %v", err)
	}

	// Verify both gateways with filters are captured
	if len(route.NextHops) != 2 {
		t.Fatalf("expected 2 next_hops, got %d", len(route.NextHops))
	}

	expectedHops := []struct {
		gateway string
		filter  int
	}{
		{"192.168.1.1", 100},
		{"192.168.1.2", 200},
	}

	for i, expected := range expectedHops {
		if route.NextHops[i].NextHop != expected.gateway {
			t.Errorf("next_hops[%d].NextHop = %q, want %q", i, route.NextHops[i].NextHop, expected.gateway)
		}
		if route.NextHops[i].Filter != expected.filter {
			t.Errorf("next_hops[%d].Filter = %d, want %d", i, route.NextHops[i].Filter, expected.filter)
		}
	}
}

func TestStaticRouteService_GetRoute_SingleGateway(t *testing.T) {
	// Test single gateway (regression test)
	mockExec := &mockStaticRouteExecutor{
		responses: map[string][]byte{
			`show config | grep "ip route default"`: []byte(
				"ip route default gateway 192.168.0.1\n",
			),
			"save": []byte(""),
		},
	}

	service := NewStaticRouteService(mockExec, nil)

	route, err := service.GetRoute(context.Background(), "0.0.0.0", "0.0.0.0")
	if err != nil {
		t.Fatalf("GetRoute failed: %v", err)
	}

	// Verify single gateway is captured
	if len(route.NextHops) != 1 {
		t.Fatalf("expected 1 next_hop, got %d", len(route.NextHops))
	}

	if route.NextHops[0].NextHop != "192.168.0.1" {
		t.Errorf("next_hops[0].NextHop = %q, want %q", route.NextHops[0].NextHop, "192.168.0.1")
	}
}

func TestStaticRouteService_GetRoute_MixedInterfaceAndGateway(t *testing.T) {
	// Test mixed gateway types (IP gateway and interface gateway)
	mockExec := &mockStaticRouteExecutor{
		responses: map[string][]byte{
			`show config | grep "ip route 10.0.0.0/8"`: []byte(
				"ip route 10.0.0.0/8 gateway 192.168.1.1\n" +
					"ip route 10.0.0.0/8 gateway pp 1 weight 2\n",
			),
			"save": []byte(""),
		},
	}

	service := NewStaticRouteService(mockExec, nil)

	route, err := service.GetRoute(context.Background(), "10.0.0.0", "255.0.0.0")
	if err != nil {
		t.Fatalf("GetRoute failed: %v", err)
	}

	// Verify both gateways are captured
	if len(route.NextHops) != 2 {
		t.Fatalf("expected 2 next_hops, got %d", len(route.NextHops))
	}

	// First hop: IP gateway
	if route.NextHops[0].NextHop != "192.168.1.1" {
		t.Errorf("next_hops[0].NextHop = %q, want %q", route.NextHops[0].NextHop, "192.168.1.1")
	}
	if route.NextHops[0].Distance != 1 {
		t.Errorf("next_hops[0].Distance = %d, want %d", route.NextHops[0].Distance, 1)
	}

	// Second hop: PP interface with weight
	if route.NextHops[1].Interface != "pp 1" {
		t.Errorf("next_hops[1].Interface = %q, want %q", route.NextHops[1].Interface, "pp 1")
	}
	if route.NextHops[1].Distance != 2 {
		t.Errorf("next_hops[1].Distance = %d, want %d", route.NextHops[1].Distance, 2)
	}
}

func TestStaticRouteService_GetRoute_ECMP_SamePrefix(t *testing.T) {
	// Test ECMP (Equal Cost Multi-Path) routing - exact case from REQ-3
	// Routes with gateways 192.168.1.20 and 192.168.1.21 should both be captured
	mockExec := &mockStaticRouteExecutor{
		responses: map[string][]byte{
			`show config | grep "ip route 10.33.128.0/21"`: []byte(
				"ip route 10.33.128.0/21 gateway 192.168.1.20\n" +
					"ip route 10.33.128.0/21 gateway 192.168.1.21\n",
			),
			"save": []byte(""),
		},
	}

	service := NewStaticRouteService(mockExec, nil)

	route, err := service.GetRoute(context.Background(), "10.33.128.0", "255.255.248.0")
	if err != nil {
		t.Fatalf("GetRoute failed: %v", err)
	}

	// Verify both ECMP gateways are captured
	if len(route.NextHops) != 2 {
		t.Fatalf("expected 2 next_hops for ECMP, got %d", len(route.NextHops))
	}

	// Verify exact gateways from REQ-3 bug report
	expectedGateways := []string{"192.168.1.20", "192.168.1.21"}
	for i, expected := range expectedGateways {
		if route.NextHops[i].NextHop != expected {
			t.Errorf("ECMP next_hops[%d].NextHop = %q, want %q", i, route.NextHops[i].NextHop, expected)
		}
	}
}
