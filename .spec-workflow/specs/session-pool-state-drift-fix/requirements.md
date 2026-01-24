# Requirements: SSH Session Pool for State Drift Fix

## Overview

SSH session initialization commands (`console character en.ascii`, `console lines infinity`) overwrite the router's configuration each time a new SSH session is created. This causes Terraform state drift when users have configured different values (e.g., `console character ja.utf8`).

## Problem Statement

### Current Behavior

1. Each Terraform operation creates a new SSH session to the RTX router
2. Session initialization executes:
   - `console character en.ascii` - sets character encoding
   - `console lines infinity` - disables paging
3. These commands modify the **running configuration** of the router
4. The router's configuration now differs from what was set via Terraform
5. On next `terraform plan`, state drift is detected

### Impact

- **rtx_system resource**: `console.character` shows diff on every plan
- **User confusion**: Repeated "in-place update" notifications for unchanged config
- **State inconsistency**: Terraform state doesn't match actual router state

### Example Scenario

```hcl
resource "rtx_system" "main" {
  console {
    character = "ja.utf8"  # User wants Japanese encoding
  }
}
```

After `terraform apply`:
1. Provider writes `console character ja.utf8` to router
2. Session closes
3. Next operation opens new session
4. Session init runs `console character en.ascii`
5. `terraform plan` shows: `character: "en.ascii" -> "ja.utf8"`

## Requirements

### Requirement 1: SSH Session Pool Implementation

**User Story:** As a Terraform user, I want my `terraform plan` to show no changes when my infrastructure hasn't changed, so that I can trust the plan output.

#### Acceptance Criteria

1. WHEN multiple Terraform operations occur THEN system SHALL reuse existing SSH sessions from the pool
2. WHEN an SSH session is reused THEN initialization commands SHALL NOT be re-executed
3. WHEN SSH session is idle for configurable timeout THEN system MAY close the session
4. WHEN SSH session becomes invalid (connection lost) THEN system SHALL create new session

### Requirement 2: Initialization Isolation

**User Story:** As a Terraform user managing rtx_system, I want session initialization to not modify my router configuration, so that my Terraform state remains consistent.

#### Acceptance Criteria

1. WHEN session is initialized THEN system SHALL preserve original console character setting
2. WHEN session is initialized THEN system SHALL preserve original console lines setting
3. IF initialization requires temporary settings THEN system SHALL restore original values on session close

### Requirement 3: Concurrent Operation Safety

**User Story:** As a user running parallel Terraform operations, I want my operations to execute safely without session conflicts.

#### Acceptance Criteria

1. WHEN multiple operations request sessions concurrently THEN system SHALL prevent race conditions
2. WHEN a session is in use THEN other operations SHALL wait or use different sessions
3. WHEN max sessions reached THEN system SHALL queue requests (not fail immediately)

### Requirement 4: Backward Compatibility

**User Story:** As an existing Terraform user, I want my current configurations to continue working without modification.

#### Acceptance Criteria

1. IF SSH session pool is disabled THEN system SHALL use current per-operation session behavior
2. WHEN upgrading provider THEN existing configurations SHALL work without changes
3. WHEN SSH session pool fails THEN system SHALL fallback to creating new SSH sessions

## Non-Functional Requirements

### Performance

- SSH session reuse should reduce overall operation time
- SSH pool management overhead should be minimal (<10ms per operation)
- SSH session health checks should not block operations

### Reliability

- Graceful handling of network disconnections
- Automatic SSH session recovery after transient failures
- No leaked SSH sessions after provider shutdown

### Resource Management

- Configurable maximum SSH sessions per router
- SSH sessions should timeout after inactivity
- Clean SSH session release on provider destruction

### Observability

- Log SSH session creation/reuse/close events
- Log SSH pool statistics (active sessions, waiting requests)
- Warning logs when falling back to new SSH sessions

## Out of Scope

- Multi-router SSH session management optimization
- SSH session migration between provider instances
- Persistent SSH session storage across Terraform runs

## Technical Constraints

- RTX routers have limited concurrent SSH session capacity (typically 4-8)
- Some operations (save config) may require exclusive SSH session access
- Administrator mode state must be tracked per SSH session

## References

- Affected file: `internal/client/working_session.go`
- Session initialization: lines 90-100
- Related resources: `rtx_system`, `rtx_l2tp_service`
