// Command djot-fmt formats djot files.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/KyleKing/djot-fmt/internal/iohelper"
)

var errFilesFailed = errors.New("one or more files had errors")

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("djot-fmt %s (commit: %s, built: %s)\n", version, commit, date)
			os.Exit(0)
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		}
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := iohelper.ParseArgs(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	if len(opts.InputFiles) == 0 {
		if err := iohelper.ProcessFile(opts, ""); err != nil {
			return fmt.Errorf("processing stdin: %w", err)
		}

		return nil
	}

	var hasError bool

	for _, file := range opts.InputFiles {
		if err := iohelper.ProcessFile(opts, file); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)

			hasError = true

			if !opts.Check {
				return fmt.Errorf("processing %s: %w", file, err)
			}
		}
	}

	if hasError {
		return errFilesFailed
	}

	return nil
}

func printHelp() {
	fmt.Print(`djot-fmt - Automatically format djot files

Usage:
  djot-fmt [options] [files...]

Arguments:
  files              Files to format (default: stdin)

Options:
  -w, --write        Write result to source file instead of stdout
  -c, --check        Check if files are formatted (exit 1 if not)
  -o, --output FILE  Write output to FILE instead of stdout (single input file only)
  -h, --help         Show this help message
  -v, --version      Show version information

Validation Options:
  --no-validate      Skip the check that formatting preserved the document
  --from-md          Allow the list changes that formatting markdown as djot causes,
                     and report each one that was waived

SLW (Semantic Line Wrap) Options:
  --no-wrap-sentences      Disable semantic line wrapping
  --slw-markers TEXT       Characters that mark sentence endings (default: ".!?")
  --slw-wrap INTEGER       Maximum line width for wrapping (default: 88, set to 0 to disable)
  --slw-min-line INTEGER   Minimum line length before wrapping (default: 40, set to 0 for aggressive mode)

Examples:
  # Format stdin to stdout with SLW enabled (default)
  cat file.dj | djot-fmt

  # Format file and write back
  djot-fmt -w file.dj

  # Format multiple files and write back
  djot-fmt -w file1.dj file2.dj file3.dj

  # Check if files are formatted
  djot-fmt -c file1.dj file2.dj

  # Format to different file
  djot-fmt -o output.dj input.dj

  # Disable SLW wrapping
  djot-fmt --no-wrap-sentences file.dj

  # Aggressive SLW mode (always wrap after sentences)
  djot-fmt --slw-min-line 0 file.dj

  # Format all .dj files in a directory (using fd)
  fd -e dj . content/ -x djot-fmt -w {}

Focus:
  This tool formats djot files with the following features:
  - List formatting (indentation, spacing, etc.)
  - Semantic line wrapping (SLW) for cleaner diffs
  - Preserves inline formatting (emphasis, strong, links, etc.)
`)
}
