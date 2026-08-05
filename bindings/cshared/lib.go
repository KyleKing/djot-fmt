// Package main exposes the djot formatter as a C shared library.
//
// Every exported function recovers from panics because a panic crossing the cgo
// boundary terminates the host process, which for a Python caller means killing
// the interpreter rather than raising.
package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/sivukhin/godjot/v2/djot_parser"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/slw"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// formatMu serializes parsing because the godjot tokenizer mutates a package-level
// map on every inline token match, so concurrent parses abort the process, and a
// ctypes call releases the GIL, which lets two Python threads arrive here at once.
var formatMu sync.Mutex

// DjotFormat formats djot source and returns a malloc'd C string the caller owns.
// On failure it returns NULL and writes a malloc'd message to errOut. Both
// pointers must be released with DjotFree.
//
//export DjotFormat
func DjotFormat(
	input *C.char,
	wrapSentences C.int,
	markers *C.char,
	maxLineWidth C.int,
	minLineLength C.int,
	errOut **C.char,
) (result *C.char) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			setError(errOut, panicMessage(r))
		}
	}()

	formatMu.Lock()
	defer formatMu.Unlock()

	*errOut = nil

	config := &slw.Config{
		Enabled:       wrapSentences != 0,
		Markers:       C.GoString(markers),
		MinLineLength: int(minLineLength),
		MaxLineWidth:  int(maxLineWidth),
		Abbreviations: slw.DefaultConfig().Abbreviations,
	}

	ast := djot_parser.BuildDjotAst([]byte(C.GoString(input)))

	return C.CString(formatter.FormatWithConfig(ast, config))
}

// DjotVersion returns a malloc'd "version commit date" triple for the caller to free.
//
//export DjotVersion
func DjotVersion() *C.char {
	return C.CString(version + " " + commit + " " + date)
}

// DjotFree releases a pointer returned by DjotFormat or DjotVersion.
//
//export DjotFree
func DjotFree(p *C.char) {
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}

func setError(errOut **C.char, message string) {
	if errOut != nil {
		*errOut = C.CString(message)
	}
}

func panicMessage(r any) string {
	message := "unknown panic in djot formatter"

	switch v := r.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	}

	return message
}

func main() {}
