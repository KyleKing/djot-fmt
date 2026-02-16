package formatter

import (
	"strings"

	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/sivukhin/godjot/v2/djot_parser"
	"github.com/sivukhin/godjot/v2/djot_tokenizer"
	"github.com/sivukhin/godjot/v2/tokenizer"
)

func formatDocument(_ djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	next(nil)
}

func formatSection(_ djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	next(nil)
}

func formatText(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
	text := string(state.Node.Text)

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

func formatList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.InListItem() {
		w.WriteString("\n")
	} else if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	_, isSparse := state.Node.Attributes.TryGet(djot_parser.SparseListNodeKey)
	w.SetInSparseList(isSparse)
	next(nil)
	w.SetInSparseList(false)

	w.SetLastBlockType(BlockTypeList)
}

func formatListItem(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	marker := "- "

	if state.Parent != nil {
		switch state.Parent.Type {
		case djot_parser.OrderedListNode:
			listStyle := state.Parent.Attributes.Get("type")
			if listStyle == "" {
				listStyle = "1"
			}

			marker = listStyle + ". "
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

	indent := strings.Repeat(" ", len(marker))
	w.PushIndent(indent)
	w.SetInListItem(true)

	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)

	next(nil)

	if w.InSparseList() {
		w.WriteString("\n")
	}

	w.SetLastBlockType(previousBlockType)
	w.PopIndent()
	w.SetInListItem(false)
}

func formatEmphasis(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("_")
	next(nil)
	state.Writer.WriteString("_")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatStrong(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("*")
	next(nil)
	state.Writer.WriteString("*")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatLink(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	url := state.Node.Attributes.Get(djot_parser.LinkHrefKey)
	state.Writer.WriteString("[")
	next(nil)
	state.Writer.WriteString("](" + url + ")")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatVerbatim(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	if _, ok := state.Node.Attributes.TryGet(djot_tokenizer.InlineMathKey); ok {
		state.Writer.WriteString("$")
		next(nil)
		state.Writer.WriteString("$")

		return
	}

	if _, ok := state.Node.Attributes.TryGet(djot_tokenizer.DisplayMathKey); ok {
		state.Writer.WriteString("$$")
		next(nil)
		state.Writer.WriteString("$$")

		return
	}

	content := extractTextContent(state.Node)

	delimiter := "`"
	needsSpaces := false

	if strings.Contains(content, "`") {
		if strings.Contains(content, "``") {
			delimiter = "```"
		} else {
			delimiter = "``"
		}

		needsSpaces = strings.HasPrefix(content, "`") || strings.HasSuffix(content, "`")
	}

	state.Writer.WriteString(delimiter)

	if needsSpaces {
		state.Writer.WriteString(" ")
	}

	next(nil)

	if needsSpaces {
		state.Writer.WriteString(" ")
	}

	state.Writer.WriteString(delimiter)
}

func makeInlineFormatter(openDelim, closeDelim string) djot_parser.Conversion[*Writer] {
	return func(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
		state.Writer.WriteString(openDelim)
		next(nil)
		state.Writer.WriteString(closeDelim)
		state.Writer.WriteString(formatAttributes(state.Node.Attributes))
	}
}

func formatLineBreak(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
	state.Writer.WriteString("\\\n")
}

func formatImage(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
	alt := state.Node.Attributes.Get(djot_parser.ImgAltKey)
	src := state.Node.Attributes.Get(djot_parser.ImgSrcKey)
	state.Writer.WriteString("![" + alt + "](" + src + ")")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatSpan(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("[")
	next(nil)
	state.Writer.WriteString("]")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatSymbols(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString(":")
	next(nil)
	state.Writer.WriteString(":")
}

func formatThematicBreak(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	w.WriteString("***\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatCode(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	class := state.Node.Attributes.Get("class")

	w.WriteString("```")

	if lang, ok := strings.CutPrefix(class, "language-"); ok {
		w.WriteString(lang)
	}

	w.WriteString("\n")

	next(nil)

	w.WriteString("```\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatRaw(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	format := state.Node.Attributes.Get(djot_parser.RawBlockFormatKey)

	w.WriteString("```=")
	w.WriteString(format)
	w.WriteString("\n")

	next(nil)

	w.WriteString("```\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatQuote(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	w.PushLinePrefix("> ")

	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)
	next(nil)
	w.SetLastBlockType(previousBlockType)

	w.PopLinePrefix()

	w.SetLastBlockType(BlockTypeParagraph)
}

func formatDiv(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	class := state.Node.Attributes.Get("class")

	w.WriteString(":::")

	if class != "" {
		w.WriteString(" ")
		w.WriteString(class)
	}

	w.WriteString("\n")

	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)
	next(nil)
	w.SetLastBlockType(previousBlockType)

	w.WriteString(":::\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatDefinitionList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	next(nil)
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatDefinitionTerm(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	next(nil)
	w.WriteString("\n")
}

func formatDefinitionItem(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	w.WriteString(": ")
	w.IncreaseIndent()

	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)
	next(nil)
	w.SetLastBlockType(previousBlockType)

	w.DecreaseIndent()
}

func formatReferenceDef(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	label := state.Node.Attributes.Get(djot_tokenizer.ReferenceKey)
	url := state.Node.Attributes.Get(djot_parser.LinkHrefKey)

	w.WriteString("[")
	w.WriteString(label)
	w.WriteString("]: ")
	w.WriteString(url)
	w.WriteString("\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatFootnoteDef(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	label := state.Node.Attributes.Get(djot_tokenizer.ReferenceKey)

	w.WriteString("[^")
	w.WriteString(label)
	w.WriteString("]: ")
	w.IncreaseIndent()

	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)
	next(nil)
	w.SetLastBlockType(previousBlockType)

	w.DecreaseIndent()
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatTable(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	next(nil)

	w.SetLastBlockType(BlockTypeParagraph)
}

func formatTableRow(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	isHeader := false
	if len(state.Node.Children) > 0 {
		isHeader = state.Node.Children[0].Type == djot_parser.TableHeaderNode
	}

	w.WriteString("|")
	next(nil)

	w.WriteString("\n")

	if isHeader && len(state.Node.Children) > 0 {
		w.WriteString("|")

		for range state.Node.Children {
			w.WriteString("---|")
		}

		w.WriteString("\n")
	}
}

func formatTableHeader(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	w.WriteString(" ")
	next(nil)
	w.WriteString(" |")
}

func formatTableCell(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	w.WriteString(" ")
	next(nil)
	w.WriteString(" |")
}

func formatTableCaption(_ djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	next(nil)
}

func formatHeading(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	levelMarker := state.Node.Attributes.Get(djot_parser.HeadingLevelKey)
	w.WriteString(levelMarker)
	w.WriteString(" ")
	next(nil)
	w.WriteString(formatAttributes(state.Node.Attributes))
	w.WriteString("\n")
	w.SetLastBlockType(BlockTypeHeading)
}

var skippedAttributes = map[string]bool{
	"href": true,
	"alt":  true,
	"src":  true,
}

func shouldSkipAttribute(key string) bool {
	if len(key) > 0 && key[0] == '$' {
		return true
	}

	return skippedAttributes[key]
}

func formatAttributes(attrs tokenizer.Attributes) string {
	var classes []string

	var id string

	var kvPairs []string

	for _, key := range attrs.Keys {
		if shouldSkipAttribute(key) {
			continue
		}

		val := attrs.Map[key]

		switch key {
		case "class":
			for _, cls := range strings.Fields(val) {
				classes = append(classes, "."+cls)
			}
		case "id":
			id = "#" + val
		default:
			escapedVal := strings.ReplaceAll(val, `"`, `\"`)
			kvPairs = append(kvPairs, key+`="`+escapedVal+`"`)
		}
	}

	var parts []string

	parts = append(parts, classes...)
	if id != "" {
		parts = append(parts, id)
	}

	parts = append(parts, kvPairs...)

	if len(parts) == 0 {
		return ""
	}

	return "{" + strings.Join(parts, " ") + "}"
}

func extractTextContent(node djot_parser.TreeNode[djot_parser.DjotNode]) string {
	if node.Type == djot_parser.TextNode {
		return string(node.Text)
	}

	var result string
	for _, child := range node.Children {
		result += extractTextContent(child)
	}

	return result
}

var defaultRegistry = map[djot_parser.DjotNode]djot_parser.Conversion[*Writer]{
	djot_parser.DocumentNode:  formatDocument,
	djot_parser.SectionNode:   formatSection,
	djot_parser.TextNode:      formatText,
	djot_parser.ParagraphNode: formatParagraph,
	djot_parser.HeadingNode:   formatHeading,

	djot_parser.UnorderedListNode: formatList,
	djot_parser.OrderedListNode:   formatList,
	djot_parser.TaskListNode:      formatList,
	djot_parser.ListItemNode:      formatListItem,

	djot_parser.EmphasisNode:    formatEmphasis,
	djot_parser.StrongNode:      formatStrong,
	djot_parser.LinkNode:        formatLink,
	djot_parser.VerbatimNode:    formatVerbatim,
	djot_parser.DeleteNode:      makeInlineFormatter("{-", "-}"),
	djot_parser.HighlightedNode: makeInlineFormatter("{=", "=}"),
	djot_parser.InsertNode:      makeInlineFormatter("{+", "+}"),
	djot_parser.SubscriptNode:   makeInlineFormatter("{~", "~}"),
	djot_parser.SuperscriptNode: makeInlineFormatter("{^", "^}"),
	djot_parser.LineBreakNode:   formatLineBreak,
	djot_parser.ImageNode:       formatImage,
	djot_parser.SpanNode:        formatSpan,
	djot_parser.SymbolsNode:     formatSymbols,

	djot_parser.ThematicBreakNode: formatThematicBreak,
	djot_parser.CodeNode:          formatCode,
	djot_parser.RawNode:           formatRaw,
	djot_parser.QuoteNode:         formatQuote,
	djot_parser.DivNode:           formatDiv,

	djot_parser.DefinitionListNode: formatDefinitionList,
	djot_parser.DefinitionTermNode: formatDefinitionTerm,
	djot_parser.DefinitionItemNode: formatDefinitionItem,

	djot_parser.ReferenceDefNode: formatReferenceDef,
	djot_parser.FootnoteDefNode:  formatFootnoteDef,

	djot_parser.TableNode:        formatTable,
	djot_parser.TableRowNode:     formatTableRow,
	djot_parser.TableHeaderNode:  formatTableHeader,
	djot_parser.TableCellNode:    formatTableCell,
	djot_parser.TableCaptionNode: formatTableCaption,
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

func FormatWithConfig(ast []djot_parser.TreeNode[djot_parser.DjotNode], slwConfig *slw.Config) string {
	writer := NewWriterWithConfig(slwConfig)
	ctx := djot_parser.ConversionContext[*Writer]{
		Format:   "djot",
		Registry: defaultRegistry,
	}
	ctx.ConvertDjot(writer, ast...)

	return writer.String()
}
