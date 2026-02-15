package iohelper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	InputFiles []string
	OutputFile string
	Write      bool
	Check      bool
	// SLW options
	NoWrapSentences bool   // Disable SLW
	SlwMarkers      string // Sentence markers (default ".!?")
	SlwWrap         int    // Max line width (default 88, 0 to disable)
	SlwMinLine      int    // Min line length before wrapping (default 40, 0 for aggressive)
}

func ParseArgs(args []string) (*Options, error) {
	opts := &Options{
		// Set SLW defaults
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	var err error

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			i, err = parseFlag(arg, args, i, opts)
			if err != nil {
				return nil, err
			}
		} else {
			if err := setInputFile(arg, opts); err != nil {
				return nil, err
			}
		}
	}

	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

func parseFlag(flag string, args []string, i int, opts *Options) (int, error) {
	switch flag {
	case "-w", "--write":
		opts.Write = true
	case "-c", "--check":
		opts.Check = true
	case "-o", "--output":
		return parseStringFlag(flag, args, i, &opts.OutputFile)
	case "--no-wrap-sentences":
		opts.NoWrapSentences = true
	case "--slw-markers":
		return parseStringFlag(flag, args, i, &opts.SlwMarkers)
	case "--slw-wrap":
		return parseIntFlag(flag, args, i, &opts.SlwWrap)
	case "--slw-min-line":
		return parseIntFlag(flag, args, i, &opts.SlwMinLine)
	default:
		return i, fmt.Errorf("unknown flag: %s", flag)
	}

	return i, nil
}

func parseStringFlag(flag string, args []string, i int, target *string) (int, error) {
	if i+1 >= len(args) {
		return i, fmt.Errorf("%s requires a value", flag)
	}

	*target = args[i+1]

	return i + 1, nil
}

func parseIntFlag(flag string, args []string, i int, target *int) (int, error) {
	if i+1 >= len(args) {
		return i, fmt.Errorf("%s requires a value", flag)
	}

	val, err := strconv.Atoi(args[i+1])
	if err != nil {
		return i, fmt.Errorf("%s requires an integer: %w", flag, err)
	}

	*target = val

	return i + 1, nil
}

func setInputFile(file string, opts *Options) error {
	opts.InputFiles = append(opts.InputFiles, file)
	return nil
}

func validateOptions(opts *Options) error {
	if opts.Write && opts.OutputFile != "" {
		return errors.New("cannot use both -w and -o")
	}

	if opts.Write && len(opts.InputFiles) == 0 {
		return errors.New("-w requires at least one input file (cannot use with stdin)")
	}

	if opts.OutputFile != "" && len(opts.InputFiles) > 1 {
		return errors.New("-o can only be used with a single input file")
	}

	if opts.Check && (opts.Write || opts.OutputFile != "") {
		return errors.New("-c cannot be used with -w or -o")
	}

	return nil
}
