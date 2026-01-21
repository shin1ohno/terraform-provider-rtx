# Requirements Document: Security Hardening

## Introduction

This spec addresses security vulnerabilities identified during a comprehensive security audit of the terraform-provider-rtx codebase. The focus is on practical, low-risk fixes that improve the security posture without breaking existing functionality.

### Scope Decision

The following issues were **excluded** from this spec after evaluation:

| Issue | Reason for Exclusion |
|-------|---------------------|
| Command injection in schedule.go | RTX CLI is not a shell - no command chaining possible (`;`, `&&`, `\|\|` have no effect). Users create their own Terraform configs. |
| Test code hardcoded credentials | Test-only dummy data (`testpass`) with no production impact |
| Password field validation | Requires RTX router specification research - separate investigation |
| Race conditions | Mutex already implemented, no reported issues |

## Alignment with Product Vision

This security hardening aligns with providing a production-ready Terraform provider that enterprise users can confidently deploy in their infrastructure management workflows.

## Requirements

### REQ-1: Sanitize Sensitive Data in Debug Logs

**User Story:** As an infrastructure operator, I want sensitive data to be redacted from debug logs, so that authentication credentials are not exposed when troubleshooting.

#### Acceptance Criteria

1. WHEN a command containing `password` (case-insensitive) is logged THEN the system SHALL replace the command content with `[REDACTED - contains sensitive data]`
2. WHEN a command containing `pre-shared-key` (case-insensitive) is logged THEN the system SHALL replace the command content with `[REDACTED - contains sensitive data]`
3. WHEN a command containing `secret` (case-insensitive) is logged THEN the system SHALL replace the command content with `[REDACTED - contains sensitive data]`
4. WHEN debug logging outputs password prompt responses THEN the system SHALL log only success/failure status, not the actual response content
5. IF TF_LOG=DEBUG is enabled THEN sensitive credentials SHALL NOT appear in any log output

### REQ-2: Warn Users About Insecure SSH Host Key Verification

**User Story:** As a security-conscious operator, I want to be warned when using insecure SSH settings, so that I can make informed decisions about my security posture.

#### Acceptance Criteria

1. WHEN `skip_host_key_check` is set to `true` THEN the system SHALL log a WARNING message indicating the security risk
2. WHEN no host key verification method is configured THEN the system SHALL log a WARNING message recommending `known_hosts_file` or `host_key` configuration
3. IF `known_hosts_file` or `host_key` is configured THEN the system SHALL NOT log any host key warnings

### REQ-3: Enhance .gitignore for Security Best Practices

**User Story:** As a developer, I want sensitive file patterns to be excluded by default in .gitignore, so that credentials are not accidentally committed.

#### Acceptance Criteria

1. WHEN the repository is cloned THEN the .gitignore SHALL exclude `.env` and `.env.*` files
2. WHEN the repository is cloned THEN the .gitignore SHALL exclude private key files (`*.pem`, `*.key`, `*.p12`, `*.pfx`)
3. WHEN the repository is cloned THEN the .gitignore SHALL exclude common credential files (`credentials.json`, `secrets.json`, `*secret*`)

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Log sanitization logic should be in a dedicated utility function
- **Modular Design**: Sanitization patterns should be configurable/extensible
- **Dependency Management**: No new external dependencies required
- **Clear Interfaces**: Sanitization function should have a clear signature: `sanitizeForLog(input string) string`

### Performance

- Log sanitization SHOULD have O(n) complexity where n is the number of sensitive patterns
- Sanitization SHOULD NOT add more than 1ms overhead per log operation

### Security

- Sensitive patterns MUST be checked case-insensitively
- The sanitization function MUST be applied consistently across all command logging locations
- WARNING messages MUST clearly communicate security risks without exposing sensitive details

### Reliability

- Log sanitization MUST NOT cause panics or crashes
- Invalid input to sanitization function SHOULD return original input (fail-open for logging, fail-closed for security)

### Usability

- Warning messages SHOULD include actionable recommendations
- Log output SHOULD remain useful for debugging (redaction message should indicate why)

## Out of Scope

- Changing default SSH host key verification behavior (breaking change for existing users)
- Adding input validation for password fields (requires RTX specification research)
- Modifying test code credentials (no security impact)
