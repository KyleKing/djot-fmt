package formatter

import (
	. "github.com/sivukhin/godjot/v2/djot_parser"
)

func formatDocument(state ConversionState[*Writer], next func(Children)) {
	next(nil)
}

func formatText(state ConversionState[*Writer], next func(Children)) {
	state.Writer.WriteString(string(state.Node.Text))
}

func formatParagraph(state ConversionState[*Writer], next func(Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	next(nil)
	w.WriteString("\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatUnorderedList(state ConversionState[*Writer], next func(Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeList)
}

func formatOrderedList(state ConversionState[*Writer], next func(Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeList)
}

func formatTaskList(state ConversionState[*Writer], next func(Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeList)
}

func formatListItem(state ConversionState[*Writer], next func(Children)) {
	w := state.Writer

	marker := "- "
	if state.Parent != nil {
		switch state.Parent.Type {
		case OrderedListNode:
			marker = "1. "
		case TaskListNode:
			class := state.Node.Attributes.Get("class")
			if class == "checked" {
				marker = "- [x] "
			} else {
				marker = "- [ ] "
			}
		}
	}

	w.WriteIndent().WriteString(marker)
	w.IncreaseIndent()
	w.SetInListItem(true)

	// Reset block type so that the first paragraph in the list item
	// doesn't add a blank line after the marker
	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)

	next(nil)

	// Restore the block type for proper spacing after the list item
	w.SetLastBlockType(previousBlockType)
	w.DecreaseIndent()
	w.SetInListItem(false)
}

func formatEmphasis(state ConversionState[*Writer], next func(Children)) {
	state.Writer.WriteString("_")
	next(nil)
	state.Writer.WriteString("_")
}

func formatStrong(state ConversionState[*Writer], next func(Children)) {
	state.Writer.WriteString("*")
	next(nil)
	state.Writer.WriteString("*")
}

func formatLink(state ConversionState[*Writer], next func(Children)) {
	url := state.Node.Attributes.Get("url")
	state.Writer.WriteString("[")
	next(nil)
	state.Writer.WriteString("](" + url + ")")
}

func formatHeading(state ConversionState[*Writer], next func(Children)) {
	level := int(state.Node.Attributes.Get("level")[0] - '0')
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	for i := 0; i < level; i++ {
		w.WriteString("#")
	}
	w.WriteString(" " + string(state.Node.Text) + "\n")
	w.SetLastBlockType(BlockTypeHeading)
}

var defaultRegistry = map[DjotNode]Conversion[*Writer]{
	DocumentNode:      formatDocument,
	TextNode:          formatText,
	ParagraphNode:     formatParagraph,
	UnorderedListNode: formatUnorderedList,
	OrderedListNode:   formatOrderedList,
	TaskListNode:      formatTaskList,
	ListItemNode:      formatListItem,
	EmphasisNode:      formatEmphasis,
	StrongNode:        formatStrong,
	LinkNode:          formatLink,
	HeadingNode:       formatHeading,
}

func Format(ast []TreeNode[DjotNode]) string {
	writer := NewWriter()
	ctx := ConversionContext[*Writer]{
		Format:   "djot",
		Registry: defaultRegistry,
	}
	ctx.ConvertDjot(writer, ast...)
	return writer.String()
}
