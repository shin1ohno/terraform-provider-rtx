# Tasks Document

## Phase 1: Infrastructure

- [x] 1. Add zerolog dependency to go.mod
  - File: go.mod
  - Add `github.com/rs/zerolog` dependency
  - Run `go mod tidy` to resolve dependencies
  - Purpose: Make zerolog available for use in the project
  - _Leverage: existing go.mod structure_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add zerolog dependency to go.mod by running `go get github.com/rs/zerolog` and verify with `go mod tidy` | Restrictions: Do not modify any other dependencies, do not upgrade existing packages | Success: zerolog appears in go.mod, `go mod tidy` runs without errors | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [x] 2. Create logging package with NewLogger function
  - File: internal/logging/logger.go
  - Create `NewLogger()` function that returns configured zerolog.Logger
  - Support TF_LOG environment variable for log level
  - Use ConsoleWriter when TF_LOG is set, JSON otherwise
  - Purpose: Provide centralized logger configuration
  - _Leverage: TF_LOG environment variable patterns from Terraform ecosystem_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in logging infrastructure | Task: Create internal/logging/logger.go with NewLogger() that configures zerolog based on TF_LOG env var (DEBUG/INFO/WARN/ERROR), uses ConsoleWriter for development (when TF_LOG set), JSON for production | Restrictions: Do not import standard log package, follow Go naming conventions, keep file focused on logger creation | Success: NewLogger() returns properly configured logger, TF_LOG=DEBUG shows debug output with ConsoleWriter, unset TF_LOG outputs JSON at warn level | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [x] 3. Add context helper functions
  - File: internal/logging/context.go
  - Create `WithContext(ctx, logger)` to store logger in context
  - Create `FromContext(ctx)` to retrieve logger from context
  - Provide fallback to global logger when context has no logger
  - Purpose: Enable logger propagation through call stack
  - _Leverage: context.Context patterns_
  - _Requirements: REQ-5_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with context propagation expertise | Task: Create internal/logging/context.go with WithContext() and FromContext() functions using context.WithValue/Value pattern, provide global logger fallback | Restrictions: Use unexported context key type to avoid collisions, do not panic on missing logger | Success: Logger can be stored and retrieved from context, missing logger returns global fallback | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [x] 4. Create sanitizing hook for sensitive data
  - File: internal/logging/sanitizer.go
  - Create zerolog.Hook that sanitizes sensitive fields
  - Integrate with existing SanitizeCommandForLog patterns
  - Purpose: Prevent sensitive data from appearing in logs
  - _Leverage: internal/client/log_sanitizer.go patterns_
  - _Requirements: REQ-3 (sensitive data protection)_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Security-focused Go Developer | Task: Create internal/logging/sanitizer.go with zerolog.Hook implementation that checks message and fields for sensitive patterns (password, secret, key, community), redacts them before output | Restrictions: Reuse patterns from internal/client/log_sanitizer.go, do not modify existing sanitizer, hook must not panic | Success: Hook redacts sensitive data in log messages and fields, existing SanitizeCommandForLog patterns are respected | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [x] 5. Write unit tests for logging package
  - File: internal/logging/logger_test.go
  - Test NewLogger with various TF_LOG values
  - Test context functions
  - Test sanitizing hook
  - Purpose: Ensure logging infrastructure works correctly
  - _Leverage: stretchr/testify_
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Engineer | Task: Create internal/logging/logger_test.go with tests for NewLogger() (test TF_LOG=DEBUG/INFO/WARN/ERROR/unset), context functions (store/retrieve/fallback), sanitizing hook (password/secret redaction) | Restrictions: Use testify assertions, capture log output with bytes.Buffer, clean up env vars after tests | Success: All tests pass, coverage for main code paths, tests are isolated and repeatable | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

## Phase 2: Client Layer Migration

- [ ] 6. Initialize logger in provider ConfigureContextFunc
  - File: internal/provider/provider.go
  - Create logger in ConfigureContextFunc
  - Store logger in context for downstream use
  - Purpose: Bootstrap logging for all provider operations
  - _Leverage: internal/logging package, existing ConfigureContextFunc_
  - _Requirements: REQ-2, REQ-5_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Modify internal/provider/provider.go ConfigureContextFunc to create logger with logging.NewLogger(), add to context with logging.WithContext(), pass context to client creation | Restrictions: Do not change provider schema, maintain backward compatibility, keep existing error handling | Success: Logger is created and stored in context during provider configuration, downstream functions can retrieve logger | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [x] 7. Migrate client/executor logs to zerolog
  - Files: internal/client/simple_executor.go, internal/client/executor.go
  - Replace log.Printf with zerolog structured logging
  - Add fields: host, command (sanitized), duration
  - Purpose: Structured logging for command execution
  - _Leverage: internal/logging, internal/client/log_sanitizer.go_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Replace log.Printf calls in simple_executor.go and executor.go with zerolog using logging.FromContext(ctx), add Str("host", addr), Str("command", SanitizeCommandForLog(cmd)), Dur("duration", elapsed) fields | Restrictions: Preserve existing log messages as Msg(), maintain log levels (DEBUG for normal ops, ERROR for failures), use existing SanitizeCommandForLog | Success: All log.Printf replaced, logs include structured fields, sensitive commands are redacted | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [x] 8. Migrate client/session logs to zerolog
  - Files: internal/client/working_session.go, internal/client/ssh_dialer.go, internal/client/rtx_session.go
  - Replace log.Printf with zerolog structured logging
  - Add fields: host, port, user, session_id
  - Purpose: Structured logging for SSH session management
  - _Leverage: internal/logging_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Replace log.Printf in working_session.go, ssh_dialer.go, rtx_session.go with zerolog, add Str("host"), Int("port"), Str("user") fields for connection context | Restrictions: Do not log passwords or keys, maintain existing log levels, preserve error context | Success: All session-related logs use zerolog with structured fields, no sensitive data in logs | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [ ] 9. Migrate client/service logs to zerolog
  - Files: internal/client/*_service.go (all service files)
  - Replace log.Printf with zerolog structured logging
  - Add fields: service_name, operation
  - Purpose: Structured logging for service layer operations
  - _Leverage: internal/logging_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Replace log.Printf in all *_service.go files with zerolog, add Str("service", serviceName), Str("operation", opName) fields | Restrictions: Batch similar changes, maintain existing log levels, keep consistent field naming across services | Success: All service files use zerolog, consistent structured fields across services | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

## Phase 3: Provider Layer Migration

- [ ] 10. Migrate provider/resource logs to zerolog
  - Files: internal/provider/resource_*.go (all resource files)
  - Replace log.Printf with zerolog structured logging
  - Add fields: resource_type, resource_id, operation (create/read/update/delete)
  - Purpose: Structured logging for Terraform resource operations
  - _Leverage: internal/logging_
  - _Requirements: REQ-4_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Replace log.Printf in all resource_*.go files with zerolog, add Str("resource_type", "rtx_*"), Str("resource_id", id), Str("operation", "create/read/update/delete") fields in CRUD functions | Restrictions: Extract logger from context in each CRUD function, maintain existing error handling, keep consistent field naming | Success: All resource files use zerolog with CRUD operation context, logs are traceable by resource | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

## Phase 4: Validation

- [ ] 11. Remove standard log package imports
  - Files: All .go files that import "log"
  - Remove `import "log"` statements
  - Ensure no remaining log.Printf calls
  - Purpose: Complete migration to zerolog
  - _Leverage: grep/sed for bulk changes_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Search for remaining `import "log"` and `log.Printf` in all .go files, remove unused imports, fix any missed conversions | Restrictions: Do not remove log imports from test files if needed for test output, verify each file compiles after changes | Success: No production code imports standard log package, `go build ./...` succeeds | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [ ] 12. Run full test suite and fix failures
  - Files: All *_test.go files as needed
  - Run `go test ./...`
  - Fix any test failures related to logging changes
  - Update test assertions if needed
  - Purpose: Ensure migration doesn't break existing functionality
  - _Leverage: stretchr/testify_
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Engineer | Task: Run `go test ./...`, identify failures related to logging changes, fix tests by updating log capture methods or assertions, ensure all tests pass | Restrictions: Do not change test logic unrelated to logging, preserve test coverage, document any test changes | Success: `go test ./...` passes with no failures, test coverage maintained | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_

- [ ] 13. Manual verification with Terraform operations
  - Test provider with TF_LOG=DEBUG
  - Verify structured log output
  - Verify ConsoleWriter formatting
  - Test with TF_LOG unset for JSON output
  - Purpose: End-to-end validation of logging system
  - _Leverage: examples/ directory_
  - _Requirements: All_
  - _Prompt: Implement the task for spec zerolog-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Build provider, run `TF_LOG=DEBUG terraform plan` with example config, verify logs show structured fields (host, command, resource_type, operation), verify ConsoleWriter format is readable, test without TF_LOG for JSON output | Restrictions: Do not modify example configs, document any issues found | Success: Logs are structured and readable in debug mode, JSON output in production mode, all fields present as designed | After completion: Mark task as [-] in tasks.md before starting, use log-implementation tool to record what was done, then mark as [x]_
