# Tasks Document: Security Hardening

## Summary

| Task | Description | Files | REQ |
|------|-------------|-------|-----|
| 1 | Create log sanitizer utility | `log_sanitizer.go` | REQ-1 |
| 2 | Add unit tests for log sanitizer | `log_sanitizer_test.go` | REQ-1 |
| 3 | Apply sanitizer to simple_executor.go | `simple_executor.go` | REQ-1 |
| 4 | Apply sanitizer to working_session.go | `working_session.go` | REQ-1 |
| 5 | Add SSH host key verification warnings | `ssh_dialer.go` | REQ-2 |
| 6 | Enhance .gitignore with security patterns | `.gitignore` | REQ-3 |

---

- [x] 1. Create log sanitizer utility
  - File: `internal/client/log_sanitizer.go`
  - Create `SanitizeCommandForLog(cmd string) string` function
  - Define sensitive patterns: `password`, `pre-shared-key`, `secret`, `community`
  - Use case-insensitive matching
  - Return `[REDACTED - contains sensitive data]` for matches
  - Purpose: Centralized utility for redacting sensitive data from logs
  - _Leverage: None (new file)_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-hardening, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Security Developer | Task: Create a log sanitizer utility in internal/client/log_sanitizer.go that provides SanitizeCommandForLog function to redact sensitive patterns (password, pre-shared-key, secret, community) from command strings using case-insensitive matching | Restrictions: Do not add external dependencies, keep the implementation simple and focused, return original string for non-matching input | _Leverage: Standard library only (strings, log) | _Requirements: REQ-1 - sanitize sensitive data in debug logs | Success: Function correctly identifies and redacts all sensitive patterns case-insensitively, returns [REDACTED - contains sensitive data] for matches, passes all unit tests | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [x] 2. Add unit tests for log sanitizer
  - File: `internal/client/log_sanitizer_test.go`
  - Test cases: password commands, pre-shared-key, case variations, safe commands, empty string
  - Verify redaction message format
  - Purpose: Ensure sanitizer reliability
  - _Leverage: internal/client/log_sanitizer.go_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-hardening, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for SanitizeCommandForLog in internal/client/log_sanitizer_test.go covering password commands, pre-shared-key, secret, community, case variations (PASSWORD, Password), safe commands, and empty strings | Restrictions: Follow Go testing conventions, use table-driven tests, do not test external dependencies | _Leverage: internal/client/log_sanitizer.go | _Requirements: REQ-1 | Success: All test cases pass, good coverage of edge cases, tests are clear and maintainable | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [x] 3. Apply sanitizer to simple_executor.go
  - File: `internal/client/simple_executor.go`
  - Modify line 33: wrap `cmd` with `SanitizeCommandForLog()`
  - Modify lines 128, 143: remove detailed response logging, keep only status
  - Purpose: Prevent sensitive data exposure in executor logs
  - _Leverage: internal/client/log_sanitizer.go_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-hardening, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify internal/client/simple_executor.go to use SanitizeCommandForLog for command logging at line 33, and simplify password response logging at lines 128 and 143 to only log success/failure status without response content | Restrictions: Do not change execution logic, only modify logging statements, maintain existing log format structure | _Leverage: internal/client/log_sanitizer.go | _Requirements: REQ-1 | Success: Commands containing sensitive patterns are redacted in logs, password responses no longer expose content, existing functionality unchanged | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [x] 4. Apply sanitizer to working_session.go
  - File: `internal/client/working_session.go`
  - Modify line 108: wrap `cmd` with `SanitizeCommandForLog()`
  - Purpose: Prevent sensitive data exposure in session logs
  - _Leverage: internal/client/log_sanitizer.go_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-hardening, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify internal/client/working_session.go to use SanitizeCommandForLog for command logging at line 108 | Restrictions: Do not change execution logic, only modify the logging statement, maintain existing log format | _Leverage: internal/client/log_sanitizer.go | _Requirements: REQ-1 | Success: Commands containing sensitive patterns are redacted in session logs, existing functionality unchanged | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [x] 5. Add SSH host key verification warnings
  - File: `internal/client/ssh_dialer.go`
  - Add warning log at line 92 before returning InsecureIgnoreHostKey
  - Warning message should explain risk and recommend alternatives
  - Purpose: Inform users about security risks of disabled host key verification
  - _Leverage: None_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec security-hardening, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Security Developer | Task: Add a warning log message in internal/client/ssh_dialer.go at line 92 before returning InsecureIgnoreHostKey, explaining the security risk and recommending known_hosts_file or host_key configuration | Restrictions: Use log.Printf with [WARN] prefix, do not change default behavior (backward compatibility), keep message concise but informative | _Leverage: None | _Requirements: REQ-2 | Success: Warning message appears when no host key verification is configured, message clearly explains risk and alternatives, no breaking changes | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [x] 6. Enhance .gitignore with security patterns
  - File: `.gitignore`
  - Add patterns: `.env`, `.env.*`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `credentials.json`, `secrets.json`, `*secret*`
  - Add comment section for security-related exclusions
  - Purpose: Prevent accidental commit of sensitive files
  - _Leverage: Existing .gitignore structure_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec security-hardening, first run spec-workflow-guide to get the workflow guide then implement the task: Role: DevOps Engineer | Task: Enhance .gitignore with security-related patterns including .env files, private keys (*.pem, *.key, *.p12, *.pfx), and credential files (credentials.json, secrets.json), with a clear comment section | Restrictions: Do not remove existing patterns, add new section at end of file, use clear comments | _Leverage: Existing .gitignore structure | _Requirements: REQ-3 | Success: New patterns are added with clear comments, existing patterns preserved, file remains well-organized | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_
