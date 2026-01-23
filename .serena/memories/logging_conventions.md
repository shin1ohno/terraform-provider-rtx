# Logging Conventions

## Library Choice
This project uses **Zerolog** (`github.com/rs/zerolog`) for all logging.

## Usage Guidelines

### Use the internal logging package for provider code
The main logging infrastructure is in `internal/logging/logger.go`:
- `logging.NewLogger()` - Creates a new zerolog logger with TF_LOG support
- `logging.Global()` - Returns the global logger instance
- `logging.WithFields()` - Adds fields to a logger

### Environment Variables
- `TF_LOG` - Controls log level (debug, info, warn, error)
- `TF_LOG_JSON` - Set to "1" for JSON output format
- CI environments automatically use JSON output

### Zerolog Best Practices
```go
// Structured logging with error
log.Error().Err(err).Str("file", path).Msg("Failed to read file")

// Fatal (exits program)
log.Fatal().Err(err).Msg("Critical error")

// Info with fields
log.Info().
    Str("resource", resourceName).
    Int("count", itemCount).
    Msg("Resource created")

// Debug
log.Debug().Interface("config", config).Msg("Configuration loaded")
```

### Do NOT use
- Standard library `log` package
- `fmt.Printf` for logging (only use for user-facing output)
- `tflog` from HashiCorp (use zerolog instead)

## File Locations
- `internal/logging/logger.go` - Main logging infrastructure
- `internal/logging/logger_test.go` - Logger tests
