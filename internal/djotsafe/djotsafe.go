// Package djotsafe serializes godjot parsing, which is not safe to call
// concurrently.
package djotsafe

import (
	"sync"

	"github.com/sivukhin/godjot/v2/djot_parser"
)

// godjot's MatchInlineToken writes the package-level StartSymbols map on every
// match, guarded only by the RecordStartSymbol const, so two parses at once
// corrupt it. Remove this once that is fixed upstream.
var mu sync.Mutex

// BuildAst parses djot source. Safe to call from multiple goroutines.
func BuildAst(document []byte) []djot_parser.TreeNode[djot_parser.DjotNode] {
	mu.Lock()
	defer mu.Unlock()

	return djot_parser.BuildDjotAst(document)
}
