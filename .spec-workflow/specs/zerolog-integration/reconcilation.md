# Reconciliation

## Product principles
- Structured logging enhances monitoring/visibility while respecting TF_LOG levels; naming consistent with tech steering.

## Implementation alignment
- `github.com/rs/zerolog` dependency added; `internal/logging/logger.go` provides TF_LOG-based level parsing and console/json outputs; sanitizer hook exists.
- Some client components (SSH, parsers) use zerolog helpers, and unit tests cover log level parsing/sanitizer.
- Gaps: most provider/resources still use `log.Printf`, so CRUD paths lack structured fields; TF_LOG context propagation and per-operation loggers are not wired; host/service-level context fields are sparse; migration of test assertions/log capture incomplete.
