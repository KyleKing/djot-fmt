package main

import (
	"fmt"
	"os"

	"github.com/KyleKing/djot-fmt/internal/iohelper"
)

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

	return iohelper.ProcessFile(opts)
}

func printHelp() {
	fmt.Print(`djot-fmt - Automatically format djot files

Usage:
  djot-fmt [options] [file]

Arguments:
  file               File to format (default: stdin)

Options:
  -w, --write        Write result to source file instead of stdout
  -c, --check        Check if file is formatted (exit 1 if not)
  -o, --output FILE  Write output to FILE instead of stdout
  -h, --help         Show this help message
  -v, --version      Show version information

SLW (Semantic Line Wrap) Options:
  --no-wrap-sentences      Disable semantic line wrapping
  --slw-markers TEXT       Characters that mark sentence endings (default: ".!?")
  --slw-wrap INTEGER       Maximum line width for wrapping (default: 88, set to 0 to disable)
  --slw-min-line INTEGER   Minimum line length before wrapping (default: 40, set to 0 for aggressive mode)

Examples:
  # Format stdin to stdout with SLW enabled (default)
  cat file.djot | djot-fmt

  # Format file and write back
  djot-fmt -w file.djot

  # Check if file is formatted
  djot-fmt -c file.djot

  # Format to different file
  djot-fmt -o output.djot input.djot

  # Disable SLW wrapping
  djot-fmt --no-wrap-sentences file.djot

  # Aggressive SLW mode (always wrap after sentences)
  djot-fmt --slw-min-line 0 file.djot

Focus:
  This tool formats djot files with the following features:
  - List formatting (indentation, spacing, etc.)
  - Semantic line wrapping (SLW) for cleaner diffs
  - Preserves inline formatting (emphasis, strong, links, etc.)
`)
}
