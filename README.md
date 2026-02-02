# djot-fmt

Automatic formatter for [djot](https://djot.net/) markup files.

## Overview

`djot-fmt` is a command-line tool that automatically formats djot files, focusing primarily on fixing common list formatting issues:

- Missing newlines between list items
- Incorrect indentation for nested lists
- Blank line spacing before nested content

## Installation

```sh
go install github.com/KyleKing/djot-fmt@latest
```

## Usage

```sh
# Format stdin to stdout
cat file.djot | djot-fmt

# Format file and write back
djot-fmt -w file.djot

# Check if file is formatted (exit 1 if not)
djot-fmt -c file.djot

# Format to different file
djot-fmt -o output.djot input.djot
```

### Options

- `-w, --write` - Write result to source file instead of stdout
- `-c, --check` - Check if file is formatted (exit 1 if not)
- `-o, --output FILE` - Write output to FILE instead of stdout
- `-h, --help` - Show help message
- `-v, --version` - Show version information

## Development

This project uses [mise](https://mise.jdx.dev/) for tool management and [hk](https://github.com/jdx/hk) for git hooks.

### Setup

```sh
# Install mise (if not already installed)
# See: https://mise.jdx.dev/getting-started.html

# Install project tools
mise install

# Install git hooks
hk install --mise
```

### Common Commands

```sh
# Run all checks (linting, tests)
mise run ci

# Format code
mise run fmt

# Run tests
mise run test

# Build binary
mise run build

# Install locally
mise run install
```

## Project Structure

```
.
├── main.go              # CLI entry point
├── internal/
│   ├── formatter/       # Core formatting logic
│   │   ├── formatter.go
│   │   ├── writer.go
│   │   └── formatter_test.go
│   └── iohelper/        # File I/O and argument parsing
│       ├── args.go
│       ├── process.go
│       └── args_test.go
└── testdata/            # Test fixtures
```

## Roadmap

Future enhancements under consideration:

- Support for all djot node types (ordered lists, tables, code blocks, etc.)
- Configurable formatting options (indentation width, line wrapping)
- Semantic line wrapping (similar to [mdformat-slw](https://github.com/KyleKing/mdformat-slw))
- Code block formatting integration

## License

MIT

## Credits

Built using [godjot](https://github.com/sivukhin/godjot) for djot parsing and AST manipulation.
