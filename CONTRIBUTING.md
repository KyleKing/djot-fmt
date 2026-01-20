# Contributing to djot-fmt

## Development Setup

1. Install [mise](https://mise.jdx.dev/)
2. Install project tools: `mise install`
3. Install git hooks: `hk install --mise`

## Before Committing

Pre-commit hooks will automatically run formatters and linters. To run manually:

```sh
# Run all checks
mise run ci

# Or run individual commands
mise run fmt   # Format code
mise run test  # Run tests
mise run lint  # Run linters
```

## Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test changes
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

## Testing

```sh
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestName ./internal/formatter
```

## Code Style

Follow standard Go conventions:

- Use `gofmt` for formatting (handled by pre-commit hooks)
- Follow effective Go guidelines
- Write table-driven tests
- Add godoc comments for exported functions
