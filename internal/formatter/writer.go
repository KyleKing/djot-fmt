package formatter

import "strings"

type BlockType int

const (
	BlockTypeNone BlockType = iota
	BlockTypeParagraph
	BlockTypeList
	BlockTypeHeading
)

type Writer struct {
	output       strings.Builder
	indentLevel  int
	lastBlock    BlockType
	inListItem   bool
	lineStart    bool
}

func NewWriter() *Writer {
	return &Writer{
		lineStart: true,
	}
}

func (w *Writer) WriteString(s string) *Writer {
	w.output.WriteString(s)
	w.lineStart = len(s) > 0 && s[len(s)-1] == '\n'
	return w
}

func (w *Writer) WriteIndent() *Writer {
	for i := 0; i < w.indentLevel; i++ {
		w.output.WriteString("  ")
	}
	return w
}

func (w *Writer) IncreaseIndent() *Writer {
	w.indentLevel++
	return w
}

func (w *Writer) DecreaseIndent() *Writer {
	if w.indentLevel > 0 {
		w.indentLevel--
	}
	return w
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
	return w.lastBlock == BlockTypeParagraph || w.lastBlock == BlockTypeList
}

func (w *Writer) InListItem() bool {
	return w.inListItem
}

func (w *Writer) String() string {
	result := w.output.String()
	return strings.TrimRight(result, "\n") + "\n"
}
