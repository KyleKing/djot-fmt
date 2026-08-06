package formatter

import (
	"strings"
	"unicode/utf8"

	"github.com/KyleKing/djot-fmt/internal/slw"
)

// BlockType identifies the kind of block most recently written.
type BlockType int

// Block kinds tracked to decide where blank lines belong.
const (
	BlockTypeNone BlockType = iota
	BlockTypeParagraph
	BlockTypeList
	BlockTypeHeading
)

// Writer accumulates formatted djot output while tracking indentation and block state.
type Writer struct {
	slwConfig    *slw.Config
	output       strings.Builder
	inline       *strings.Builder
	indentStack  []string
	linePrefixes []string
	lastBlock    BlockType
	lineWidth    int
	inListItem   bool
	lineStart    bool
	inSparseList bool
}

// NewWriter returns a Writer using the default semantic line wrapping config.
func NewWriter() *Writer {
	return &Writer{
		lineStart: true,
		slwConfig: slw.DefaultConfig(),
	}
}

// NewWriterWithConfig returns a Writer using the given semantic line wrapping config.
func NewWriterWithConfig(slwConfig *slw.Config) *Writer {
	return &Writer{
		lineStart: true,
		slwConfig: slwConfig,
	}
}

// WriteString appends s, applying any active line prefixes.
//
//nolint:unparam // Chainable writer API keeps a uniform signature across write helpers.
func (w *Writer) WriteString(s string) *Writer {
	if w.inline != nil {
		w.inline.WriteString(s)
		return w
	}

	if len(w.linePrefixes) == 0 {
		w.writeStringDirect(s)
		return w
	}

	w.writeStringWithPrefixes(s)

	return w
}

func (w *Writer) writeStringDirect(s string) {
	w.emit(s)
	w.lineStart = s != "" && s[len(s)-1] == '\n'
}

func (w *Writer) emit(s string) {
	for _, r := range s {
		if r == '\n' {
			w.lineWidth = 0
		} else {
			w.lineWidth++
		}
	}

	w.output.WriteString(s)
}

func (w *Writer) writeStringWithPrefixes(s string) {
	prefix := strings.Join(w.linePrefixes, "")

	for i, char := range s {
		w.applyPrefixAtLineStart(i, char, prefix)
		w.emit(string(char))
		w.applyPrefixAfterNewline(i, char, s, prefix)
	}

	w.lineStart = s != "" && s[len(s)-1] == '\n'
}

func (w *Writer) applyPrefixAtLineStart(index int, char rune, prefix string) {
	if index == 0 && w.lineStart {
		if char == '\n' {
			w.emit(strings.TrimRight(prefix, " "))
		} else {
			w.emit(prefix)
		}
	}
}

func (w *Writer) applyPrefixAfterNewline(index int, char rune, s, prefix string) {
	if char == '\n' && index < len(s)-1 {
		nextIsNewline := index+1 < len(s) && s[index+1] == '\n'
		if nextIsNewline {
			w.emit(strings.TrimRight(prefix, " "))
		} else {
			w.emit(prefix)
		}
	}
}

// WriteIndent writes the current indentation stack.
func (w *Writer) WriteIndent() *Writer {
	for _, indent := range w.indentStack {
		w.emit(indent)
	}

	return w
}

// PushIndent adds an indentation level.
func (w *Writer) PushIndent(indent string) *Writer {
	w.indentStack = append(w.indentStack, indent)
	return w
}

// PopIndent removes the innermost indentation level.
func (w *Writer) PopIndent() *Writer {
	if len(w.indentStack) > 0 {
		w.indentStack = w.indentStack[:len(w.indentStack)-1]
	}

	return w
}

// IncreaseIndent pushes one two-space indentation level.
//
//nolint:unparam // Chainable writer API keeps a uniform signature across write helpers.
func (w *Writer) IncreaseIndent() *Writer {
	return w.PushIndent("  ")
}

// DecreaseIndent removes the innermost indentation level.
//
//nolint:unparam // Chainable writer API keeps a uniform signature across write helpers.
func (w *Writer) DecreaseIndent() *Writer {
	return w.PopIndent()
}

// SetLastBlockType records the kind of block just written.
func (w *Writer) SetLastBlockType(t BlockType) {
	w.lastBlock = t
}

// GetLastBlockType reports the kind of block just written.
func (w *Writer) GetLastBlockType() BlockType {
	return w.lastBlock
}

// SetInListItem marks whether output is inside a list item.
func (w *Writer) SetInListItem(inList bool) {
	w.inListItem = inList
}

// NeedsBlankLine reports whether a blank line must separate the previous block from the next.
func (w *Writer) NeedsBlankLine() bool {
	return w.lastBlock == BlockTypeParagraph || w.lastBlock == BlockTypeList || w.lastBlock == BlockTypeHeading
}

// InListItem reports whether output is inside a list item.
func (w *Writer) InListItem() bool {
	return w.inListItem
}

// BeginInline redirects writes into a buffer so a block can wrap its finished
// inline text before any indent or line prefix is applied. Blocks do not nest
// inline content, so there is only ever one buffer open.
func (w *Writer) BeginInline() {
	w.inline = &strings.Builder{}
}

// EndInline returns the buffered inline text and resumes normal writing.
func (w *Writer) EndInline() string {
	text := w.inline.String()
	w.inline = nil

	return text
}

// WrapInline applies semantic line wrapping to a block's finished inline text,
// measured from the column where that text will land.
func (w *Writer) WrapInline(text string) string {
	if w.slwConfig == nil || !w.slwConfig.Enabled {
		return text
	}

	prefix := utf8.RuneCountInString(strings.Join(w.linePrefixes, ""))

	first := w.lineWidth
	if w.lineStart {
		first = prefix
	}

	return slw.Wrap(text, slw.Layout{FirstColumn: first, ContinuationColumn: prefix}, w.slwConfig)
}

// PushLinePrefix adds a prefix written at the start of every subsequent line.
func (w *Writer) PushLinePrefix(prefix string) {
	w.linePrefixes = append(w.linePrefixes, prefix)
}

// PopLinePrefix removes the innermost line prefix.
func (w *Writer) PopLinePrefix() {
	if len(w.linePrefixes) > 0 {
		w.linePrefixes = w.linePrefixes[:len(w.linePrefixes)-1]
	}
}

// SetInSparseList marks whether output is inside a list whose items are blank-line separated.
func (w *Writer) SetInSparseList(sparse bool) {
	w.inSparseList = sparse
}

// InSparseList reports whether output is inside a list whose items are blank-line separated.
func (w *Writer) InSparseList() bool {
	return w.inSparseList
}

func (w *Writer) String() string {
	result := w.output.String()
	return strings.TrimRight(result, "\n") + "\n"
}
