// Package generator provides code generation from YAML spec files.
package generator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Spec represents the root of a spec YAML file
type Spec struct {
	Command CommandSpec `yaml:"command"`
}

// CommandSpec represents a command specification
type CommandSpec struct {
	Name                 string                    `yaml:"name"`
	Description          string                    `yaml:"description"`
	Reference            string                    `yaml:"reference"`
	ApplicableModels     []string                  `yaml:"applicable_models"`
	Syntax               SyntaxSpec                `yaml:"syntax"`
	Terraform            TerraformSpec             `yaml:"terraform"`
	Parameters           map[string]Param          `yaml:"parameters"`
	SyntaxTests          []SyntaxTest              `yaml:"syntax_tests"`
	BoundaryTests        map[string][]BoundaryTest `yaml:"boundary_tests"`
	Pairwise             *PairwiseSpec             `yaml:"pairwise"`
	MultilineTests       []SyntaxTest              `yaml:"multiline_tests"`
	Notes                []string                  `yaml:"notes"`
	ImplementationStatus string                    `yaml:"implementation_status"`
}

// SyntaxSpec represents the command syntax
type SyntaxSpec struct {
	Set    interface{} `yaml:"set"`    // string or []string
	Delete interface{} `yaml:"delete"` // string or []string
}

// TerraformSpec represents the Terraform mapping
type TerraformSpec struct {
	ResourceName     string                   `yaml:"resource_name"`
	StructName       string                   `yaml:"struct_name"`
	Package          string                   `yaml:"package"`
	StructAdditions  map[string][]StructField `yaml:"struct_additions"`
	StructDefinition map[string][]StructField `yaml:"struct_definition"`
}

// StructField represents a struct field to add
type StructField struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	JSONTag     string   `yaml:"json_tag"`
	Description string   `yaml:"description"`
	Enum        []string `yaml:"enum"`
}

// Param represents a parameter definition
type Param struct {
	Description      string                 `yaml:"description"`
	Required         bool                   `yaml:"required"`
	Type             string                 `yaml:"type"`
	Range            []int                  `yaml:"range"`
	Default          interface{}            `yaml:"default"`
	Boundaries       []interface{}          `yaml:"boundaries"`
	EnumValues       []EnumValue            `yaml:"enum_values"`
	Variants         []ParamVariant         `yaml:"variants"`
	TerraformField   string                 `yaml:"terraform_field"`
	TerraformFields  map[string]string      `yaml:"terraform_fields"`
	ModelConstraints map[string]interface{} `yaml:"model_constraints"`
	Note             string                 `yaml:"note"`
}

// EnumValue represents an enum value with description
type EnumValue struct {
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
}

// ParamVariant represents a parameter variant
type ParamVariant struct {
	Type            string            `yaml:"type"`
	Value           string            `yaml:"value"`
	Pattern         string            `yaml:"pattern"`
	Description     string            `yaml:"description"`
	Range           []int             `yaml:"range"`
	TerraformField  string            `yaml:"terraform_field"`
	TerraformFields map[string]string `yaml:"terraform_fields"`
	TerraformValue  map[string]string `yaml:"terraform_value"`
}

// SyntaxTest represents a syntax test case
type SyntaxTest struct {
	Name             string            `yaml:"name"`
	RTX              string            `yaml:"rtx"`
	Terraform        interface{}       `yaml:"terraform"` // map or array of maps
	Bidirectional    *bool             `yaml:"bidirectional"`
	ParseOnly        bool              `yaml:"parse_only"`
	BuildOnly        bool              `yaml:"build_only"`
	Note             string            `yaml:"note"`
	Description      string            `yaml:"description"`
	ModelConstraints *ModelConstraints `yaml:"model_constraints"`
}

// ModelConstraints represents model-specific constraints
type ModelConstraints struct {
	ValidFor        []string `yaml:"valid_for"`
	InvalidFor      []string `yaml:"invalid_for"`
	RequiresLicense string   `yaml:"requires_license"`
	Unavailable     []string `yaml:"unavailable"`
}

// BoundaryTest represents a boundary value test
type BoundaryTest struct {
	Value         interface{} `yaml:"value"`
	Valid         bool        `yaml:"valid"`
	Description   string      `yaml:"description"`
	ErrorContains string      `yaml:"error_contains"`
	ValidFor      []string    `yaml:"valid_for"`
	InvalidFor    []string    `yaml:"invalid_for"`
}

// PairwiseSpec represents pairwise testing configuration
type PairwiseSpec struct {
	Enabled         bool                     `yaml:"enabled"`
	Parameters      []string                 `yaml:"parameters"`
	ParameterValues map[string][]interface{} `yaml:"parameter_values"`
	Constraints     []PairwiseConstraint     `yaml:"constraints"`
}

// PairwiseConstraint represents a constraint for pairwise testing
type PairwiseConstraint struct {
	Condition  string   `yaml:"condition"`
	Requires   string   `yaml:"requires"`
	InvalidFor []string `yaml:"invalid_for"`
}

// LoadSpec loads a spec from a YAML file
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Validate required fields
	if spec.Command.Name == "" {
		return nil, fmt.Errorf("command.name is required")
	}

	return &spec, nil
}
