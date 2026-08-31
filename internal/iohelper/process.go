package iohelper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/KyleKing/djot-fmt/internal/djotsafe"
	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/KyleKing/djot-fmt/internal/validate"
)

const (
	formattedFileMode = 0o600
	diffContextLines  = 3
	issueTracker      = "https://github.com/KyleKing/djot-fmt/issues"
)

var (
	errNotFormatted     = errors.New("file not formatted")
	errValidationFailed = errors.New("formatting changed the document")
)

// ProcessFile formats inputFile and routes the result according to opts.
func ProcessFile(opts *Options, inputFile string) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			//nolint:err113 // Only the panic value carries a message, so wrapping a sentinel would change output.
			retErr = fmt.Errorf("%v", r)
		}
	}()

	input, err := readInput(inputFile)
	if err != nil {
		return err
	}

	ast := djotsafe.BuildAst(input)

	slwConfig := &slw.Config{
		Enabled:       !opts.NoWrapSentences,
		Markers:       opts.SlwMarkers,
		MinLineLength: opts.SlwMinLine,
		MaxLineWidth:  opts.SlwWrap,
		Abbreviations: slw.DefaultConfig().Abbreviations,
	}

	formatted := formatter.FormatWithConfig(ast, slwConfig)

	if err := validateOutput(opts, input, formatted, inputFile); err != nil {
		return err
	}

	if opts.Check {
		return checkFormatted(input, formatted, inputFile)
	}

	return writeOutput(formatted, opts, inputFile)
}

// validateOutput fails the run when formatting changed what the document means.
// A rejected difference is a djot-fmt bug rather than a problem with the input,
// so the message says so and points at the issue tracker.
func validateOutput(opts *Options, input []byte, formatted, inputFile string) error {
	if opts.NoValidate {
		return nil
	}

	result := validate.Compare(input, []byte(formatted), validate.Options{FromMD: opts.FromMD})

	displayName := inputFile
	if displayName == "" {
		displayName = "stdin"
	}

	for _, d := range result.Waived {
		fmt.Fprintf(os.Stderr, "%s: waived by --from-md: %s: %s\n", displayName, d.Category, d.Detail)
	}

	if result.OK() {
		return nil
	}

	fmt.Fprintf(os.Stderr, "%s: formatting changed the document, which is a bug in djot-fmt\n", displayName)

	for _, d := range result.Rejected {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", d.Category, d.Detail)
	}

	fmt.Fprintln(os.Stderr, "Please report this at "+issueTracker)
	fmt.Fprintln(os.Stderr, "Use --no-validate to skip this check.")

	return errValidationFailed
}

func readInput(inputFile string) ([]byte, error) {
	if inputFile == "" || inputFile == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading from stdin: %w", err)
		}

		return data, nil
	}

	//nolint:gosec // Reading a caller-named path is the purpose of this formatter.
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("reading input file: %w", err)
	}

	return data, nil
}

func writeOutput(formatted string, opts *Options, inputFile string) error {
	output := []byte(formatted)

	if opts.Write {
		if err := os.WriteFile(inputFile, output, formattedFileMode); err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}

		return nil
	}

	if opts.OutputFile != "" {
		return writeToFile(output, opts.OutputFile)
	}

	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func writeToFile(output []byte, path string) error {
	//nolint:gosec // Writing to the caller-named -o path is the documented behavior.
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}

	_, writeErr := f.Write(output)
	closeErr := f.Close()

	if writeErr != nil {
		return fmt.Errorf("writing output: %w", writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("closing output file: %w", closeErr)
	}

	return nil
}

func checkFormatted(original []byte, formatted, filename string) error {
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
		Context:  diffContextLines,
	})

	if err == nil && diff != "" {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(diff))
	}

	return errNotFormatted
}
