package parsers

import (
	"fmt"
	"sync"
)

// Parser is the interface that all model-specific parsers must implement
type Parser interface {
	// Parse takes raw command output and returns parsed data
	Parse(raw string) (interface{}, error)
	// CanHandle returns true if this parser can handle the given model
	CanHandle(model string) bool
}

// Registry manages parsers for different RTX models and commands
type Registry struct {
	mu      sync.RWMutex
	parsers map[string]Parser // key: "command:model" e.g., "interfaces:RTX1210"
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]Parser),
	}
}

// Register adds a parser to the registry
func (r *Registry) Register(command, model string, parser Parser) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s", command, model)
	r.parsers[key] = parser
}

// RegisterAlias creates an alias for an existing parser
func (r *Registry) RegisterAlias(command, model, aliasModel string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	sourceKey := fmt.Sprintf("%s:%s", command, model)
	parser, exists := r.parsers[sourceKey]
	if !exists {
		return fmt.Errorf("parser not found for %s:%s", command, model)
	}
	
	aliasKey := fmt.Sprintf("%s:%s", command, aliasModel)
	r.parsers[aliasKey] = parser
	return nil
}

// Get retrieves a parser for the given command and model
func (r *Registry) Get(command, model string) (Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Try exact match first
	key := fmt.Sprintf("%s:%s", command, model)
	if parser, exists := r.parsers[key]; exists {
		return parser, nil
	}
	
	// Try model family match (e.g., RTX12xx for RTX1210, RTX1220)
	if len(model) >= 5 && model[:3] == "RTX" {
		familyKey := fmt.Sprintf("%s:RTX%cxxx", command, model[3])
		if parser, exists := r.parsers[familyKey]; exists {
			return parser, nil
		}
	}
	
	return nil, fmt.Errorf("no parser found for command %q and model %q", command, model)
}

// GlobalRegistry is the default parser registry
var GlobalRegistry = NewRegistry()

// Register is a convenience function to register with the global registry
func Register(command, model string, parser Parser) {
	GlobalRegistry.Register(command, model, parser)
}

// RegisterAlias is a convenience function to create an alias in the global registry
func RegisterAlias(command, model, aliasModel string) error {
	return GlobalRegistry.RegisterAlias(command, model, aliasModel)
}

// Get is a convenience function to get from the global registry
func Get(command, model string) (Parser, error) {
	return GlobalRegistry.Get(command, model)
}