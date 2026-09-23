# AGENTS.md

## Commands

```bash
# Build
go build ./...

# Lint
just lint

# Lint with auto-fix
just lint-with-fix

# Update dependencies
just bump-deps

# Install
just install
```

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

## Commits & Pull Requests

- Commit message: no Conventional Commit prefixes, capitalize the first letter (e.g. "Add delete menu to connection list").
