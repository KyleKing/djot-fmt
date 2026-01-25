package iohelper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	InputFile  string
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

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-w", "--write":
			opts.Write = true
		case "-c", "--check":
			opts.Check = true
		case "-o", "--output":
			if i+1 >= len(args) {
				return nil, errors.New("-o/--output requires a file argument")
			}
			i++
			opts.OutputFile = args[i]
		case "--no-wrap-sentences":
			opts.NoWrapSentences = true
		case "--slw-markers":
			if i+1 >= len(args) {
				return nil, errors.New("--slw-markers requires a value")
			}
			i++
			opts.SlwMarkers = args[i]
		case "--slw-wrap":
			if i+1 >= len(args) {
				return nil, errors.New("--slw-wrap requires a value")
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, fmt.Errorf("--slw-wrap requires an integer: %w", err)
			}
			opts.SlwWrap = val
		case "--slw-min-line":
			if i+1 >= len(args) {
				return nil, errors.New("--slw-min-line requires a value")
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, fmt.Errorf("--slw-min-line requires an integer: %w", err)
			}
			opts.SlwMinLine = val
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			if opts.InputFile != "" {
				return nil, errors.New("multiple input files not supported")
			}
			opts.InputFile = arg
		}
	}

	if opts.Write && opts.OutputFile != "" {
		return nil, errors.New("cannot use both -w and -o")
	}

	if opts.Write && opts.InputFile == "" {
		return nil, errors.New("-w requires an input file (cannot use with stdin)")
	}

	if opts.Check && (opts.Write || opts.OutputFile != "") {
		return nil, errors.New("-c cannot be used with -w or -o")
	}

	return opts, nil
}
