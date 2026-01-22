# Master Spec Update Skill

---
name: master-spec
description: Update the master specification when a new spec is completed. Consolidates incremental specs into a complete, authoritative specification document.
disable-model-invocation: true
allowed-tools: Read, Write, Edit, Glob, Grep, Bash(git:*)
---

## Overview

This skill updates the **Master Specification** - a complete, authoritative specification that represents the current state of a feature or component. Unlike incremental specs (which track changes), Master Specs are **overwritten** with each update to maintain a single source of truth.

## Critical: Source of Truth Hierarchy

**IMPORTANT**: When creating or updating Master Specs, follow this source of truth hierarchy:

1. **Implementation Code** (Primary Source)
   - Actual Go source files (`*.go`)
   - Service implementations, parsers, resource definitions
   - This is the **single source of truth** for what the system actually does

2. **Test Code** (Secondary Source)
   - Unit tests and integration tests (`*_test.go`)
   - Acceptance tests
   - Tests document expected behavior and edge cases

3. **Existing Specs** (Supplementary Information)
   - `.spec-workflow/specs/` and `.spec-workflow/archive/specs/`
   - Use as **reference only** - specs may be outdated or incomplete
   - Not all implemented features have corresponding specs

**Workflow implication**: Always read and analyze the actual implementation before referencing specs. Specs help explain intent and context, but the code is authoritative.

## Arguments

- `$ARGUMENTS`: The spec name or path that was just completed (e.g., `rtx-dns-server` or `dns-server-select-refactor`)

## Workflow

### 1. Locate the Completed Spec

Find the completed spec in one of these locations:
- `.spec-workflow/specs/{spec-name}/` (active specs)
- `.spec-workflow/archive/specs/{spec-name}/` (archived specs)

Read both `requirements.md` and `design.md` from the completed spec.

### 2. Identify the Master Spec Category

Determine which Master Spec this update belongs to:
- Extract the primary resource/component name from the spec
- For example: `rtx-dns-server` and `dns-server-select-refactor` both belong to `dns-server` Master Spec

### 3. Load or Create Master Spec

Check for existing Master Spec at the configured location (default: `.spec-workflow/master-specs/{category}/`):
- If exists: Load current `requirements.md` and `design.md`
- If not exists: Create new from steering templates

### 4. Merge Strategy

Apply the following merge strategy:

**Requirements (`requirements.md`):**
- **Overwrite** the entire requirements document with the new spec's requirements
- Preserve the steering template structure
- Ensure all sections from the template are present

**Design (`design.md`):**
- **Overwrite** the entire design document with the new spec's design
- Preserve the steering template structure
- Update architecture diagrams if present
- Update component interfaces and data models

### 5. Validate Against Templates

Ensure the updated Master Spec conforms to steering templates:
- `.spec-workflow/templates/requirements-template.md`
- `.spec-workflow/templates/design-template.md`

### 6. Save and Report

- Save the updated Master Spec files
- Report what was updated:
  - Spec source
  - Master Spec category
  - Sections updated
  - Any validation warnings

## Configuration

Master Spec location can be configured per-project. Default structure:

```
.spec-workflow/
├── master-specs/
│   ├── {category}/
│   │   ├── requirements.md
│   │   └── design.md
│   └── ...
├── specs/           # Active incremental specs
├── archive/         # Completed specs
└── templates/       # Steering templates
```

## Example Usage

After completing `dns-server-select-refactor` spec:

```
/master-spec dns-server-select-refactor
```

This will:
1. Read `.spec-workflow/archive/specs/dns-server-select-refactor/requirements.md`
2. Read `.spec-workflow/archive/specs/dns-server-select-refactor/design.md`
3. Identify category as `dns-server`
4. Update `.spec-workflow/master-specs/dns-server/requirements.md`
5. Update `.spec-workflow/master-specs/dns-server/design.md`

## Special Cases

### New Resource/Component
If this is the first spec for a resource/component:
- Create the Master Spec directory
- Initialize from templates
- Populate with the spec content

### Refactoring Spec
If the spec is a refactor (e.g., `dns-server-select-refactor`):
- Extract the base component name (`dns-server`)
- Merge the refactoring changes into the existing Master Spec
- The refactored design becomes the new authoritative design

### Multiple Related Specs
If multiple specs complete the same feature:
- Process each in chronological order
- Later specs overwrite earlier ones
- Maintain change history in commit messages

## Template Compliance

The Master Spec must maintain these sections (per steering templates):

**requirements.md:**
- Introduction
- Alignment with Product Vision
- Requirements (with User Stories and Acceptance Criteria)
- Non-Functional Requirements

**design.md:**
- Overview
- Steering Document Alignment
- Code Reuse Analysis
- Architecture
- Components and Interfaces
- Data Models
- Error Handling
- Testing Strategy
