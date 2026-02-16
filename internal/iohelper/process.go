package iohelper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sivukhin/godjot/v2/djot_parser"
)

func ProcessFile(opts *Options, inputFile string) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("%v", r)
		}
	}()

	input, err := readInput(inputFile)
	if err != nil {
		return err
	}

	ast := djot_parser.BuildDjotAst(input)

	slwConfig := &slw.Config{
		Enabled:       !opts.NoWrapSentences,
		Markers:       opts.SlwMarkers,
		MinLineLength: opts.SlwMinLine,
		MaxLineWidth:  opts.SlwWrap,
		Abbreviations: slw.DefaultConfig().Abbreviations,
	}

	formatted := formatter.FormatWithConfig(ast, slwConfig)

	if opts.Check {
		return checkFormatted(input, formatted, inputFile)
	}

	return writeOutput(formatted, opts, inputFile)
}

func readInput(inputFile string) ([]byte, error) {
	if inputFile == "" || inputFile == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading from stdin: %w", err)
		}

		return data, nil
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("reading input file: %w", err)
	}

	return data, nil
}

func writeOutput(formatted string, opts *Options, inputFile string) error {
	output := []byte(formatted)

	if opts.Write {
		if err := os.WriteFile(inputFile, output, 0600); err != nil {
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

	if _, err := writer.Write(output); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func checkFormatted(original []byte, formatted string, filename string) error {
	if bytes.Equal(original, []byte(formatted)) {
		return nil
	}

	displayName := filename
	if displayName == "" {
		displayName = "stdin"
	}

	fmt.Fprintf(os.Stderr, "%s: not formatted\n", displayName)

	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(original)),
		B:        difflib.SplitLines(formatted),
		FromFile: displayName,
		ToFile:   displayName + " (formatted)",
		Context:  3,
	})

	if err == nil && diff != "" {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(diff))
	}

	return errors.New("file not formatted")
}
