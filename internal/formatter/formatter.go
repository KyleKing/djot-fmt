package formatter

import (
	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/sivukhin/godjot/v2/djot_parser"
)

func formatDocument(_ djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	next(nil)
}

func formatSection(_ djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	next(nil)
}

func formatText(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
	text := string(state.Node.Text)

	// Apply SLW wrapping if we're in a paragraph
	if state.Writer.InParagraph() && state.Writer.slwConfig != nil && state.Writer.slwConfig.Enabled {
		text = slw.WrapText(text, state.Writer.slwConfig)
	}

	state.Writer.WriteString(text)
}

func formatParagraph(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	w.SetInParagraph(true)
	next(nil)
	w.SetInParagraph(false)

	w.WriteString("\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatUnorderedList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeList)
}

func formatOrderedList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeList)
}

func formatTaskList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeList)
}

func formatListItem(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	marker := "- "

	if state.Parent != nil {
		switch state.Parent.Type {
		case djot_parser.OrderedListNode:
			marker = "1. "
		case djot_parser.TaskListNode:
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

func formatEmphasis(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("_")
	next(nil)
	state.Writer.WriteString("_")
}

func formatStrong(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("*")
	next(nil)
	state.Writer.WriteString("*")
}

func formatLink(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	url := state.Node.Attributes.Get("url")
	state.Writer.WriteString("[")
	next(nil)
	state.Writer.WriteString("](" + url + ")")
}

func formatHeading(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	// The level is stored in the $HeadingLevelKey attribute as "#" characters
	levelMarker := state.Node.Attributes.Get("$HeadingLevelKey")
	w.WriteString(levelMarker)
	w.WriteString(" ")
	next(nil)
	w.WriteString("\n")
	w.SetLastBlockType(BlockTypeHeading)
}

var defaultRegistry = map[djot_parser.DjotNode]djot_parser.Conversion[*Writer]{
	djot_parser.DocumentNode:      formatDocument,
	djot_parser.SectionNode:       formatSection,
	djot_parser.TextNode:          formatText,
	djot_parser.ParagraphNode:     formatParagraph,
	djot_parser.UnorderedListNode: formatUnorderedList,
	djot_parser.OrderedListNode:   formatOrderedList,
	djot_parser.TaskListNode:      formatTaskList,
	djot_parser.ListItemNode:      formatListItem,
	djot_parser.EmphasisNode:      formatEmphasis,
	djot_parser.StrongNode:        formatStrong,
	djot_parser.LinkNode:          formatLink,
	djot_parser.HeadingNode:       formatHeading,
}

func Format(ast []djot_parser.TreeNode[djot_parser.DjotNode]) string {
	writer := NewWriter()
	ctx := djot_parser.ConversionContext[*Writer]{
		Format:   "djot",
		Registry: defaultRegistry,
	}
	ctx.ConvertDjot(writer, ast...)

	return writer.String()
}

// FormatWithConfig formats the djot AST with custom SLW configuration
func FormatWithConfig(ast []djot_parser.TreeNode[djot_parser.DjotNode], slwConfig *slw.Config) string {
	writer := NewWriterWithConfig(slwConfig)
	ctx := djot_parser.ConversionContext[*Writer]{
		Format:   "djot",
		Registry: defaultRegistry,
	}
	ctx.ConvertDjot(writer, ast...)

	return writer.String()
}
