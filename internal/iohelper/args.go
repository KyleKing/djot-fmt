// Package iohelper parses command line options and drives file formatting.
package iohelper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultSlwWrap    = 88
	defaultSlwMinLine = 40
)

var (
	errUnknownFlag     = errors.New("unknown flag")
	errFlagNeedsValue  = errors.New("requires a value")
	errWriteWithOutput = errors.New("cannot use both -w and -o")
	errWriteNeedsFile  = errors.New("-w requires at least one input file (cannot use with stdin)")
	errOutputOneFile   = errors.New("-o can only be used with a single input file")
	errCheckWithWrite  = errors.New("-c cannot be used with -w or -o")
)

// Options holds the parsed command line configuration.
type Options struct {
	OutputFile      string
	SlwMarkers      string
	InputFiles      []string
	SlwWrap         int
	SlwMinLine      int
	Write           bool
	Check           bool
	NoWrapSentences bool
}

// ParseArgs converts command line arguments into validated Options.
func ParseArgs(args []string) (*Options, error) {
	opts := &Options{
		SlwMarkers: ".!?",
		SlwWrap:    defaultSlwWrap,
		SlwMinLine: defaultSlwMinLine,
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
			opts.InputFiles = append(opts.InputFiles, arg)
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
		return i, fmt.Errorf("%w: %s", errUnknownFlag, flag)
	}

	return i, nil
}

func parseStringFlag(flag string, args []string, i int, target *string) (int, error) {
	if i+1 >= len(args) {
		return i, fmt.Errorf("%s %w", flag, errFlagNeedsValue)
	}

	*target = args[i+1]

	return i + 1, nil
}

func parseIntFlag(flag string, args []string, i int, target *int) (int, error) {
	if i+1 >= len(args) {
		return i, fmt.Errorf("%s %w", flag, errFlagNeedsValue)
	}

	val, err := strconv.Atoi(args[i+1])
	if err != nil {
		return i, fmt.Errorf("%s requires an integer: %w", flag, err)
	}

	*target = val

	return i + 1, nil
}

func validateOptions(opts *Options) error {
	if opts.Write && opts.OutputFile != "" {
		return errWriteWithOutput
	}

	if opts.Write && len(opts.InputFiles) == 0 {
		return errWriteNeedsFile
	}

	if opts.OutputFile != "" && len(opts.InputFiles) > 1 {
		return errOutputOneFile
	}

	if opts.Check && (opts.Write || opts.OutputFile != "") {
		return errCheckWithWrite
	}

	return nil
}
