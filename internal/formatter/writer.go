package formatter

import (
	"strings"

	"github.com/KyleKing/djot-fmt/internal/slw"
)

type BlockType int

const (
	BlockTypeNone BlockType = iota
	BlockTypeParagraph
	BlockTypeList
	BlockTypeHeading
)

type Writer struct {
	output       strings.Builder
	indentStack  []string // Stack of indent strings for nested lists
	lastBlock    BlockType
	inListItem   bool
	lineStart    bool
	slwConfig    *slw.Config
	inParagraph  bool
	linePrefixes []string // Stack of line prefixes for blockquotes, etc.
	inSparseList bool
}

func NewWriter() *Writer {
	return &Writer{
		lineStart: true,
		slwConfig: slw.DefaultConfig(),
	}
}

func NewWriterWithConfig(slwConfig *slw.Config) *Writer {
	return &Writer{
		lineStart: true,
		slwConfig: slwConfig,
	}
}

func (w *Writer) WriteString(s string) *Writer {
	if len(w.linePrefixes) == 0 {
		w.writeStringDirect(s)
		return w
	}

	w.writeStringWithPrefixes(s)

	return w
}

func (w *Writer) writeStringDirect(s string) {
	w.output.WriteString(s)
	w.lineStart = len(s) > 0 && s[len(s)-1] == '\n'
}

func (w *Writer) writeStringWithPrefixes(s string) {
	prefix := strings.Join(w.linePrefixes, "")

	for i, char := range s {
		w.applyPrefixAtLineStart(i, char, prefix)
		w.output.WriteRune(char)
		w.applyPrefixAfterNewline(i, char, s, prefix)
	}

	w.lineStart = len(s) > 0 && s[len(s)-1] == '\n'
}

func (w *Writer) applyPrefixAtLineStart(index int, char rune, prefix string) {
	if index == 0 && w.lineStart {
		if char == '\n' {
			w.output.WriteString(strings.TrimRight(prefix, " "))
		} else {
			w.output.WriteString(prefix)
		}
	}
}

func (w *Writer) applyPrefixAfterNewline(index int, char rune, s string, prefix string) {
	if char == '\n' && index < len(s)-1 {
		nextIsNewline := index+1 < len(s) && s[index+1] == '\n'
		if nextIsNewline {
			w.output.WriteString(strings.TrimRight(prefix, " "))
		} else {
			w.output.WriteString(prefix)
		}
	}
}

func (w *Writer) WriteIndent() *Writer {
	for _, indent := range w.indentStack {
		w.output.WriteString(indent)
	}

	return w
}

func (w *Writer) PushIndent(indent string) *Writer {
	w.indentStack = append(w.indentStack, indent)
	return w
}

func (w *Writer) PopIndent() *Writer {
	if len(w.indentStack) > 0 {
		w.indentStack = w.indentStack[:len(w.indentStack)-1]
	}

	return w
}

// IncreaseIndent provides backward compatibility for fixed 2-space indentation
func (w *Writer) IncreaseIndent() *Writer {
	return w.PushIndent("  ")
}

// DecreaseIndent provides backward compatibility
func (w *Writer) DecreaseIndent() *Writer {
	return w.PopIndent()
}

func (w *Writer) SetLastBlockType(t BlockType) {
	w.lastBlock = t
}

func (w *Writer) GetLastBlockType() BlockType {
	return w.lastBlock
}

func (w *Writer) SetInListItem(inList bool) {
	w.inListItem = inList
}

func (w *Writer) NeedsBlankLine() bool {
	return w.lastBlock == BlockTypeParagraph || w.lastBlock == BlockTypeList || w.lastBlock == BlockTypeHeading
}

func (w *Writer) InListItem() bool {
	return w.inListItem
}

func (w *Writer) SetInParagraph(inPara bool) {
	w.inParagraph = inPara
}

func (w *Writer) InParagraph() bool {
	return w.inParagraph
}

func (w *Writer) PushLinePrefix(prefix string) {
	w.linePrefixes = append(w.linePrefixes, prefix)
}

func (w *Writer) PopLinePrefix() {
	if len(w.linePrefixes) > 0 {
		w.linePrefixes = w.linePrefixes[:len(w.linePrefixes)-1]
	}
}

func (w *Writer) SetInSparseList(sparse bool) {
	w.inSparseList = sparse
}

func (w *Writer) InSparseList() bool {
	return w.inSparseList
}

func (w *Writer) String() string {
	result := w.output.String()
	return strings.TrimRight(result, "\n") + "\n"
}
