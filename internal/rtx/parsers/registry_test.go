package parsers

import (
	"testing"
)

// mockParser is a test parser implementation
type mockParser struct {
	canHandle bool
	result    interface{}
	err       error
}

func (m *mockParser) Parse(raw string) (interface{}, error) {
	return m.result, m.err
}

func (m *mockParser) CanHandle(model string) bool {
	return m.canHandle
}

func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewRegistry()
		parser := &mockParser{canHandle: true, result: "test"}

		// Register parser
		registry.Register("test", "RTX1210", parser)

		// Get parser
		got, err := registry.Get("test", "RTX1210")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != parser {
			t.Errorf("expected parser %v, got %v", parser, got)
		}
	})

	t.Run("Get non-existent parser", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.Get("test", "RTX1210")
		if err == nil {
			t.Fatal("expected error for non-existent parser")
		}
	})

	t.Run("Register alias", func(t *testing.T) {
		registry := NewRegistry()
		parser := &mockParser{canHandle: true, result: "test"}

		// Register original parser
		registry.Register("test", "RTX1210", parser)

		// Create alias
		err := registry.RegisterAlias("test", "RTX1210", "RTX1220")
		if err != nil {
			t.Fatalf("unexpected error creating alias: %v", err)
		}

		// Get via alias
		got, err := registry.Get("test", "RTX1220")
		if err != nil {
			t.Fatalf("unexpected error getting via alias: %v", err)
		}

		if got != parser {
			t.Errorf("expected parser %v, got %v", parser, got)
		}
	})

	t.Run("Register alias for non-existent parser", func(t *testing.T) {
		registry := NewRegistry()

		err := registry.RegisterAlias("test", "RTX1210", "RTX1220")
		if err == nil {
			t.Fatal("expected error for non-existent source parser")
		}
	})

	t.Run("Model family match", func(t *testing.T) {
		registry := NewRegistry()
		parser := &mockParser{canHandle: true, result: "test"}

		// Register with family pattern
		registry.Register("test", "RTX1xxx", parser)

		// Get with specific model
		got, err := registry.Get("test", "RTX1210")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != parser {
			t.Errorf("expected parser %v, got %v", parser, got)
		}
	})
}
