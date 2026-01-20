package iohelper

import (
	"errors"
	"fmt"
)

type Options struct {
	InputFile  string
	OutputFile string
	Write      bool
	Check      bool
}

func ParseArgs(args []string) (*Options, error) {
	opts := &Options{}

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
		default:
			if arg[0] == '-' {
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
