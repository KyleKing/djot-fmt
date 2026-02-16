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

	_, isSparse := state.Node.Attributes.TryGet(djot_parser.SparseListNodeKey)
	w.SetInSparseList(isSparse)
	next(nil)
	w.SetInSparseList(false)

	w.SetLastBlockType(BlockTypeList)
}

func formatOrderedList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
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

func formatTaskList(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
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
			// Preserve ordered list style (1., a., A., i., I.)
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

	// Push indent matching marker width for proper alignment
	indent := strings.Repeat(" ", len(marker))
	w.PushIndent(indent)
	w.SetInListItem(true)

	// Reset block type so that the first paragraph in the list item
	// doesn't add a blank line after the marker
	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)

	next(nil)

	// Add blank line after item in sparse lists
	if w.InSparseList() {
		w.WriteString("\n")
	}

	// Restore the block type for proper spacing after the list item
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
	// Check for math mode or raw inline format
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

	// Regular inline code - determine delimiter based on content
	content := extractTextContent(state.Node)

	delimiter := "`"
	needsSpaces := false

	// Check if content contains backticks and needs a different delimiter
	if containsSubstring(content, "`") {
		if containsSubstring(content, "``") {
			delimiter = "```"
		} else {
			delimiter = "``"
		}
		// Need spaces if content starts or ends with backticks
		needsSpaces = hasPrefix(content, "`") || hasSuffix(content, "`")
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

func formatDelete(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("{-")
	next(nil)
	state.Writer.WriteString("-}")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatInsert(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("{+")
	next(nil)
	state.Writer.WriteString("+}")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatHighlighted(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("{=")
	next(nil)
	state.Writer.WriteString("=}")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatSubscript(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("{~")
	next(nil)
	state.Writer.WriteString("~}")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
}

func formatSuperscript(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	state.Writer.WriteString("{^")
	next(nil)
	state.Writer.WriteString("^}")
	state.Writer.WriteString(formatAttributes(state.Node.Attributes))
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
	w.SetLastBlockType(BlockTypeParagraph) // Reuse existing block type
}

func formatCode(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	// Extract language from class attribute (format: "language-<lang>")
	class := state.Node.Attributes.Get("class")
	lang := ""

	if class != "" && len(class) > 9 && class[:9] == "language-" {
		lang = class[9:]
	}

	w.WriteString("```")

	if lang != "" {
		w.WriteString(lang)
	}

	w.WriteString("\n")

	// Process children (should be TextNode with code content)
	next(nil)

	w.WriteString("```\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatRaw(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	// Get raw format from attribute
	format := state.Node.Attributes.Get(djot_parser.RawBlockFormatKey)

	w.WriteString("```=")
	w.WriteString(format)
	w.WriteString("\n")

	// Process children (should be TextNode with raw content)
	next(nil)

	w.WriteString("```\n")
	w.SetLastBlockType(BlockTypeParagraph)
}

func formatQuote(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	// Push blockquote prefix - WriteString will apply it automatically
	w.PushLinePrefix("> ")

	// Process children (paragraphs, lists, etc.)
	previousBlockType := w.GetLastBlockType()
	w.SetLastBlockType(BlockTypeNone)
	next(nil)
	w.SetLastBlockType(previousBlockType)

	// Pop prefix
	w.PopLinePrefix()

	w.SetLastBlockType(BlockTypeParagraph)
}

func formatDiv(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	// Get class from attributes
	class := state.Node.Attributes.Get("class")

	w.WriteString(":::")

	if class != "" {
		w.WriteString(" ")
		w.WriteString(class)
	}

	w.WriteString("\n")

	// Process children
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

	// Terms are inline content
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

	// Get reference key and URL
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

	// Get footnote label
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

// Table rendering - simplified implementation
func formatTable(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n")
	}

	// Process children (rows and caption)
	next(nil)

	w.SetLastBlockType(BlockTypeParagraph)
}

func formatTableRow(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	// Check if this is a header row (first row with TableHeaderNode children)
	isHeader := false
	if len(state.Node.Children) > 0 {
		isHeader = state.Node.Children[0].Type == djot_parser.TableHeaderNode
	}

	// Start row
	w.WriteString("|")

	// Process cells
	next(nil)

	w.WriteString("\n")

	// Add separator row after header
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
	// Caption processing - simplified (would need special handling in real implementation)
	next(nil)
}

func formatHeading(state djot_parser.ConversionState[*Writer], next func(djot_parser.Children)) {
	w := state.Writer

	if w.NeedsBlankLine() {
		w.WriteString("\n\n")
	}

	// The level is stored in the $HeadingLevelKey attribute as "#" characters
	levelMarker := state.Node.Attributes.Get(djot_parser.HeadingLevelKey)
	w.WriteString(levelMarker)
	w.WriteString(" ")
	next(nil)
	w.WriteString(formatAttributes(state.Node.Attributes))
	w.WriteString("\n")
	w.SetLastBlockType(BlockTypeHeading)
}

// shouldSkipAttribute checks if an attribute should be skipped during formatting
func shouldSkipAttribute(key string) bool {
	// Skip internal attributes ($-prefixed)
	if len(key) > 0 && key[0] == '$' {
		return true
	}

	// Semantic attributes to skip (internal or structural)
	skipAttrs := map[string]bool{
		"href": true,
		"alt":  true,
		"src":  true,
	}

	return skipAttrs[key]
}

// formatAttributes formats user attributes (class, id, key=value) into djot syntax
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
			for _, cls := range splitSpaces(val) {
				classes = append(classes, "."+cls)
			}
		case "id":
			id = "#" + val
		default:
			escapedVal := replaceAll(val, `"`, `\"`)
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

	return "{" + joinStrings(parts, " ") + "}"
}

// splitSpaces splits a string on whitespace
func splitSpaces(s string) []string {
	if s == "" {
		return nil
	}

	var result []string

	var current string

	for _, char := range s {
		if char == ' ' || char == '\t' || char == '\n' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// replaceAll replaces all occurrences of old with replacement in s
func replaceAll(s, old, replacement string) string {
	var result string

	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += replacement
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}

	return result
}

// joinStrings joins strings with separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}

	return result
}

// extractTextContent recursively extracts all text content from a node
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

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring returns the index of the first occurrence of substr in s, or -1 if not found
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}

// hasPrefix checks if s starts with prefix
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// hasSuffix checks if s ends with suffix
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

var defaultRegistry = map[djot_parser.DjotNode]djot_parser.Conversion[*Writer]{
	// Supported node types
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
	// Phase 1: Inline formatting
	djot_parser.VerbatimNode:    formatVerbatim,
	djot_parser.DeleteNode:      formatDelete,
	djot_parser.InsertNode:      formatInsert,
	djot_parser.HighlightedNode: formatHighlighted,
	djot_parser.SubscriptNode:   formatSubscript,
	djot_parser.SuperscriptNode: formatSuperscript,
	djot_parser.LineBreakNode:   formatLineBreak,
	djot_parser.ImageNode:       formatImage,
	djot_parser.SpanNode:        formatSpan,
	djot_parser.SymbolsNode:     formatSymbols,

	// Phase 2: Block-level nodes
	djot_parser.ThematicBreakNode: formatThematicBreak,
	djot_parser.CodeNode:          formatCode,
	djot_parser.RawNode:           formatRaw,
	djot_parser.QuoteNode:         formatQuote,
	djot_parser.DivNode:           formatDiv,

	// Phase 4: Definition lists
	djot_parser.DefinitionListNode: formatDefinitionList,
	djot_parser.DefinitionTermNode: formatDefinitionTerm,
	djot_parser.DefinitionItemNode: formatDefinitionItem,

	// Phase 5: References and footnotes
	djot_parser.ReferenceDefNode: formatReferenceDef,
	djot_parser.FootnoteDefNode:  formatFootnoteDef,

	// Phase 3: Tables
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
