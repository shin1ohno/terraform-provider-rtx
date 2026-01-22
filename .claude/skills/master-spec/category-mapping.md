# Master Spec Category Mapping

This document defines how incremental specs map to Master Spec categories.

## Mapping Rules

### 1. Direct Resource Specs
Specs named `rtx-{resource}` map directly to the `{resource}` category.

| Spec Pattern | Master Spec Category |
|--------------|---------------------|
| `rtx-dns-server` | `dns-server` |
| `rtx-dhcp-scope` | `dhcp-scope` |
| `rtx-static-route` | `static-route` |
| `rtx-interface` | `interface` |
| `rtx-vlan` | `vlan` |
| `rtx-bgp` | `bgp` |
| `rtx-ospf` | `ospf` |
| `rtx-nat-static` | `nat-static` |
| `rtx-nat-masquerade` | `nat-masquerade` |
| `rtx-ip-filter` | `ip-filter` |
| `rtx-ethernet-filter` | `ethernet-filter` |

### 2. Enhancement/Refactor Specs
Specs that enhance or refactor existing resources map to the base resource category.

| Spec Pattern | Master Spec Category | Extraction Rule |
|--------------|---------------------|-----------------|
| `dns-server-select-refactor` | `dns-server` | Extract resource name from prefix |
| `filter-nat-enhancements` | `filter-nat` | Extract base component |
| `import-fidelity-fix` | `core` | Infrastructure/cross-cutting |

### 3. Infrastructure Specs
Cross-cutting concerns map to special categories.

| Spec Pattern | Master Spec Category |
|--------------|---------------------|
| `rtx-command-parser` | `core/parser` |
| `parser-*` | `core/parser` |
| `security-*` | `core/security` |
| `zerolog-integration` | `core/logging` |

### 4. Custom Mappings
Project-specific mappings can be defined here:

```yaml
# Custom category mappings
mappings:
  # spec-name: category
  "my-custom-spec": "custom-category"
```

## Category Structure

```
.spec-workflow/master-specs/
├── dns-server/
│   ├── requirements.md
│   └── design.md
├── dhcp-scope/
│   ├── requirements.md
│   └── design.md
├── core/
│   ├── parser/
│   │   ├── requirements.md
│   │   └── design.md
│   ├── security/
│   │   ├── requirements.md
│   │   └── design.md
│   └── logging/
│       ├── requirements.md
│       └── design.md
└── ...
```

## Auto-Detection Algorithm

When category is not explicitly mapped:

1. **Check for `rtx-` prefix**: Extract resource name after prefix
2. **Check for `-refactor` or `-enhancement` suffix**: Extract base name
3. **Check for `parser-` or `security-` prefix**: Map to `core/{prefix}`
4. **Default**: Use the full spec name as category
