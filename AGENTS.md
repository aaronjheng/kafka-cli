# AGENTS.md

## CRITICAL RULES

- **NEVER commit or push changes unless the user EXPLICITLY asks you to.** Even if the user says "commit", do NOT also push unless they say "push". Do NOT assume the user wants to commit after making changes. Always wait for explicit instruction.

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

## Code Quality

### Formatting

- Use `gofumpt` for formatting.
- Use `gci` for import ordering (standard, default, localmodule).

### Linting

- All lint errors must be fixed before committing.
- Use `slog` for logging, not `fmt.Print`.
- Wrap errors with `fmt.Errorf("...: %w", err)`.
- Do not enable the `ireturn` linter; returning interfaces is common in this project.
- Do not enable deprecated linters.

## Git Workflow Rules

- NEVER run `git commit`, `git push`, or any other git mutations unless explicitly instructed
- Always ask for confirmation before committing changes
- Always ask for confirmation before pushing to remote

## Commit Message Guidelines

- Use a single sentence
- Do not use Conventional Commits (e.g., no `feat:`, `fix:` prefixes)
- Capitalize the first letter
- Example: "Add delete menu to connection list"
