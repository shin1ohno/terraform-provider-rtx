# Master Requirements: {resource-name}

## Overview

{Brief description of the resource and its purpose in the overall system}

## Alignment with Product Vision

{How this resource supports the goals outlined in product.md}

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `{terraform_resource_name}` |
| Type | {singleton/collection} |
| Import Support | {yes/no} |
| Last Updated | {date} |
| Source Specs | {list of contributing specs} |

## Functional Requirements

### Core Operations

#### Create
{Description of create operation and requirements}

#### Read
{Description of read operation and requirements}

#### Update
{Description of update operation and requirements}

#### Delete
{Description of delete operation and requirements}

### Feature Requirements

{Numbered list of feature requirements with user stories and acceptance criteria}

### Requirement 1: {Feature Name}

**User Story:** As a {role}, I want {feature}, so that {benefit}

#### Acceptance Criteria

1. WHEN {event} THEN {system} SHALL {response}
2. IF {precondition} THEN {system} SHALL {response}

{Repeat for each requirement}

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each file should have a single, well-defined purpose
- **Modular Design**: Components, utilities, and services should be isolated and reusable
- **Dependency Management**: Minimize interdependencies between modules
- **Clear Interfaces**: Define clean contracts between components and layers

### Performance
{Performance requirements specific to this resource}

### Security
{Security requirements specific to this resource}

### Reliability
{Reliability requirements specific to this resource}

### Validation
{Input validation requirements}

## RTX Commands Reference

```
{List of RTX router commands used by this resource}
```

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | {description} |
| `terraform apply` | ✅ Required | {description} |
| `terraform destroy` | ✅ Required | {description} |
| `terraform import` | ✅ Required | {description} |
| `terraform refresh` | ✅ Required | {description} |
| `terraform state` | ✅ Required | {description} |

### Import Specification
- **Import ID Format**: {format}
- **Import Command**: `terraform import {resource_type}.{name} {id}`
- **Post-Import**: {requirements}

## Example Usage

```hcl
{Complete Terraform configuration example}
```

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status must not be stored in state
- {Additional state handling notes}

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| {date} | {spec-name} | {summary of changes} |
