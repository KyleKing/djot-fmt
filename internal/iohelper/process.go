package iohelper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/kyleking/djot-fmt/internal/formatter"
	"github.com/sivukhin/godjot/v2/djot_parser"
)

func ProcessFile(opts *Options) error {
	input, err := readInput(opts)
	if err != nil {
		return err
	}

	ast := djot_parser.BuildDjotAst(input)
	formatted := formatter.Format(ast)

	if opts.Check {
		return checkFormatted(input, formatted, opts.InputFile)
	}

	return writeOutput(formatted, opts)
}

func readInput(opts *Options) ([]byte, error) {
	if opts.InputFile == "" || opts.InputFile == "-" {
		return io.ReadAll(os.Stdin)
	}

	data, err := os.ReadFile(opts.InputFile)
	if err != nil {
		return nil, fmt.Errorf("reading input file: %w", err)
	}
	return data, nil
}

func writeOutput(formatted string, opts *Options) error {
	output := []byte(formatted)

	if opts.Write {
		if err := os.WriteFile(opts.InputFile, output, 0644); err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}
		return nil
	}

	var writer io.Writer = os.Stdout
	if opts.OutputFile != "" {
		f, err := os.Create(opts.OutputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		writer = f
	}

	_, err := writer.Write(output)
	return err
}

func checkFormatted(original []byte, formatted string, filename string) error {
	if bytes.Equal(original, []byte(formatted)) {
		return nil
	}

	if filename != "" {
		fmt.Fprintf(os.Stderr, "%s: not formatted\n", filename)
	} else {
		fmt.Fprintln(os.Stderr, "input: not formatted")
	}
	return errors.New("file not formatted")
}
