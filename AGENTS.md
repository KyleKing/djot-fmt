# Agent Instructions for djot-fmt

This file provides guidance to AI coding assistants working with this codebase.

## Project Overview

`djot-fmt` is a command-line tool for automatically formatting [djot](https://djot.net/) markup files. The primary focus is on correcting common list formatting issues:

- Missing newlines between list items
- Incorrect indentation for nested lists
- Proper blank line spacing before nested content

## Architecture

### Structure

```
├── main.go              # CLI entry point
├── internal/
│   ├── formatter/       # Core formatting logic
│   │   ├── formatter.go  # Node conversion functions
│   │   ├── writer.go     # Output writer with state tracking
│   │   └── formatter_test.go
│   └── iohelper/        # File I/O and CLI argument handling
│       ├── args.go       # Command-line argument parsing
│       ├── process.go    # File reading/writing logic
│       └── args_test.go
└── testdata/            # Test fixtures
```

### Key Design Patterns

**Conversion System** - Uses godjot's generic conversion pattern:
- Registry maps `DjotNode` types to conversion functions
- Each conversion function receives state and a callback for processing children
- Writer tracks formatting state (indentation, block types, context)

**State Tracking** - The `Writer` tracks:
- Current indentation level
- Last block type (for spacing decisions)
- Whether currently inside a list item (for nested list handling)

## Development Guidelines

### When Modifying Formatters

1. **Add new node type support**: Update `defaultRegistry` in `formatter.go` with a new conversion function
2. **Modify spacing**: Adjust `NeedsBlankLine()` logic or block type tracking in `writer.go`
3. **Handle edge cases**: Add test cases in `formatter_test.go` before implementing

### Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/formatter -v

# Test with coverage
go test -cover ./...
```

### Code Style

Follow Go best practices from `GO_BEST_PRACTICES.md`:
- Table-driven tests
- Small, single-responsibility functions
- Explicit error handling with context wrapping
- No dot imports (except in formatter.go for brevity)

### Linting Requirements

The project enforces strict linting rules. Ensure code passes all linters before committing:

**revive**:
- Never use built-in function names as parameter names (`new`, `make`, `len`, etc.)
- Use `switch` statements instead of if-else chains with 3+ branches

**gocritic**:
- Convert if-else chains to switch statements when appropriate
- Prefer switch for cleaner, more maintainable branching logic

**testifylint**:
- Use `require.Error(t, err)` and `require.NoError(t, err)` for error assertions in tests
- Never use `assert.Error` or `assert.NoError` - tests should stop on error assertion failures
- `assert.*` is acceptable for non-error value comparisons

**cyclop**:
- Maximum cyclomatic complexity is 10 per function
- Extract helper functions to reduce complexity when needed
- Break down complex conditionals and loops into smaller functions

## Common Tasks

### Adding Support for New Node Type

1. Add conversion function in `formatter.go`:
   ```go
   func formatNewNode(state ConversionState[*Writer], next func(Children)) {
       // Handle node formatting
       next(nil)  // Process children
   }
   ```

2. Register in `defaultRegistry`:
   ```go
   NewNodeType: formatNewNode,
   ```

3. Add test case in `formatter_test.go`

### Debugging Formatting Issues

Use the AST inspection pattern:
```go
ast := djot_parser.BuildDjotAst(input)
// Print AST structure to understand node hierarchy
```

### Local Development with godjot

The project uses a `replace` directive in `go.mod` to use the local godjot:
```
replace github.com/sivukhin/godjot/v2 => ../godjot
```

Build with:
```bash
GOWORK=off go build -o djot-fmt
```

## Future Enhancements

Potential areas for expansion:
- Support for all djot node types (ordered lists, tables, code blocks, etc.)
- Configurable formatting options (indentation width, line wrapping)
- Semantic line wrapping integration
- Code block formatting

## Resources

- [djot specification](https://htmlpreview.github.io/?https://github.com/jgm/djot/blob/master/doc/syntax.html)
- [godjot library](https://github.com/sivukhin/godjot)
- [Go best practices](GO_BEST_PRACTICES.md)
