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

Or from PyPI, which ships the same formatter as a shared library bound with `ctypes`:

```sh
uv tool install djot-fmt   # or: pipx install djot-fmt
uvx djot-fmt file.dj       # run without installing
```

The Python wheels carry a compiled library rather than a reimplementation, so output
is byte-identical to the Go binary. Formatting runs in-process, which costs about
9 microseconds for a small document against roughly 3 milliseconds to spawn the CLI.

## Python API

```python
import djot_fmt

djot_fmt.format('-  a\n-  b\n')  # '- a\n- b\n'
djot_fmt.format(source, wrap_sentences=False)
```

`format` raises `djot_fmt.DjotFormatError` on input the formatter rejects. Calls are
serialized internally because the underlying djot parser is not goroutine-safe.

The shared library is opened on the first call rather than at import, so a process
that imports `djot_fmt` and then forks stays safe. Calling into it before forking
does not, because the Go runtime does not survive `fork()` without `exec()`. Use the
`spawn` or `forkserver` start method with `multiprocessing`.

## Usage

### Basic Usage

```sh
# Format stdin to stdout
cat file.dj | djot-fmt

# Format file and write back
djot-fmt -w file.dj

# Format multiple files and write back
djot-fmt -w file1.dj file2.dj file3.dj

# Check if file is formatted (exit 1 if not)
djot-fmt -c file.dj

# Check multiple files
djot-fmt -c file1.dj file2.dj file3.dj

# Format to different file
djot-fmt -o output.dj input.dj
```

### Batch Processing

```sh
# Format all .dj files in current directory
fd -e dj -x djot-fmt -w

# Format all .dj files in specific directory
fd -e dj . content/ -x djot-fmt -w

# Check formatting for all .dj files
fd -e dj -x djot-fmt -c

# Using find (alternative to fd)
find . -name "*.dj" -exec djot-fmt -w {} \;
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

### Python bindings

```sh
# Build the wheel (compiles bindings/cshared with -buildmode=c-shared)
uv build --wheel

# Run the Python tests against the built wheel
uv run --isolated --with dist/*.whl --with pytest pytest bindings/python/tests
```

`bindings/cshared/lib.go` is the cgo boundary. Every exported function there recovers
from panics, because a panic crossing into C terminates the host process, which for a
Python caller means killing the interpreter. Strings allocated on the Go side are freed
through `DjotFree`.

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
