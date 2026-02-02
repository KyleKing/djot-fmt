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
	output      strings.Builder
	indentLevel int
	lastBlock   BlockType
	inListItem  bool
	lineStart   bool
	slwConfig   *slw.Config
	inParagraph bool
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
	w.output.WriteString(s)
	w.lineStart = len(s) > 0 && s[len(s)-1] == '\n'

	return w
}

func (w *Writer) WriteIndent() *Writer {
	for range w.indentLevel {
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

func (w *Writer) SetInParagraph(inPara bool) {
	w.inParagraph = inPara
}

func (w *Writer) InParagraph() bool {
	return w.inParagraph
}

func (w *Writer) String() string {
	result := w.output.String()
	return strings.TrimRight(result, "\n") + "\n"
}
