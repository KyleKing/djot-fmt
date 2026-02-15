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

### Basic Usage

```sh
# Format stdin to stdout
cat file.djot | djot-fmt

# Format file and write back
djot-fmt -w file.djot

# Format multiple files and write back
djot-fmt -w file1.djot file2.djot file3.djot

# Check if file is formatted (exit 1 if not)
djot-fmt -c file.djot

# Check multiple files
djot-fmt -c file1.djot file2.djot file3.djot

# Format to different file
djot-fmt -o output.djot input.djot
```

### Batch Processing

```sh
# Format all .djot files in current directory
fd -e djot -x djot-fmt -w

# Format all .djot files in specific directory
fd -e djot . content/ -x djot-fmt -w

# Check formatting for all .djot files
fd -e djot -x djot-fmt -c

# Using find (alternative to fd)
find . -name "*.djot" -exec djot-fmt -w {} \;
```

### Options

- `-w, --write` - Write result to source file(s) instead of stdout
- `-c, --check` - Check if file(s) are formatted (exit 1 if not)
- `-o, --output FILE` - Write output to FILE instead of stdout (single input file only)
- `-h, --help` - Show help message
- `-v, --version` - Show version information

### SLW (Semantic Line Wrap) Options

- `--no-wrap-sentences` - Disable semantic line wrapping
- `--slw-markers TEXT` - Characters that mark sentence endings (default: ".!?")
- `--slw-wrap INTEGER` - Maximum line width for wrapping (default: 88, set to 0 to disable)
- `--slw-min-line INTEGER` - Minimum line length before wrapping (default: 40, set to 0 for aggressive mode)

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
