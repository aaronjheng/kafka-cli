# AGENTS.md

## Commands

```bash
# Build
go build ./...

# Lint
just lint

# Check go mod tidy
go mod tidy -diff

# Update dependencies
just bump-deps
```

## Project Structure

- `cmd/kafka/` - CLI entry point and command definitions
- `internal/admin/` - Kafka admin client operations (topics, groups)
- `internal/config/` - Configuration loading and management
- `internal/kafka/` - Kafka client, dialer, and connection utilities
- `internal/ssh/` - SSH proxy support

## Code Style

- Use `gofumpt` for formatting
- Use `gci` for import ordering (standard, default, localmodule)
- All lint errors must be fixed before committing
- Use `slog` for logging, not `fmt.Print`
- Wrap errors with `fmt.Errorf("...: %w", err)`

## Code Quality

- Do not enable the `ireturn` linter; returning interfaces is common in this project.
- Do not enable deprecated linters.
