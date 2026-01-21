# RTX Test Data Directory

This directory contains test fixtures and pattern catalogs for RTX command parser testing.

## Directory Structure

```
testdata/
├── patterns/           # Command pattern catalog (YAML files)
│   ├── schema.yaml     # Pattern schema definition
│   ├── dns.yaml        # DNS command patterns
│   ├── static_route.yaml
│   ├── vlan.yaml
│   ├── interface.yaml
│   ├── ipsec.yaml
│   ├── ospf.yaml
│   ├── bgp.yaml
│   ├── nat.yaml
│   └── ...
├── fixtures/           # Generated test fixtures
│   ├── dns/           # DNS command test cases
│   ├── static_route/  # Static route test cases
│   ├── vlan/          # VLAN test cases
│   └── ...
├── import_fidelity/    # Complex import test cases
├── RTX830/            # RTX830 model-specific test data
└── RTX1210/           # RTX1210 model-specific test data
```

## Pattern Catalog Files

Pattern files (`patterns/*.yaml`) define:
- Command syntax patterns
- Parameter types and constraints
- Example commands from documentation
- Reference to parser implementation

See `patterns/schema.yaml` for the schema definition.

## Test Fixtures

Fixtures (`fixtures/*/*.txt`) contain:
- Input command strings
- Expected parsed output (JSON format)
- Source documentation reference

## Usage

1. Add patterns to appropriate YAML file in `patterns/`
2. Generate test fixtures from patterns
3. Run tests with `go test ./internal/rtx/parsers/...`
4. Fix any parser gaps identified by failing tests
