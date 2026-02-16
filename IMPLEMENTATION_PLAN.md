# Implementation Plan: Full Djot Node Support

## Context

The djot-fmt formatter currently handles 12 of 37 node types defined by the godjot parser. The remaining 25 node types were registered with a `formatUnsupported` panic handler to prevent silent content loss. This document plans the implementation of all missing handlers.

### Why were these not implemented before?

The project started with a narrow scope: "correcting common list formatting issues" (from AGENTS.md). The godjot library's conversion system silently skips unregistered node types via `continue` in the conversion loop -- this made the gap invisible during development. There were no tests for any content beyond paragraphs, lists, headings, emphasis, strong, and links, so the silent data loss was never caught. The `PROJECT_STATUS.md` incorrectly claimed unsupported types "will pass through" -- they were silently dropped.

### Bugs found during this audit

1. **Link URL lost**: `formatLink` read attribute `"url"` but godjot stores it as `"href"` (`LinkHrefKey`). Links rendered as `[text]()`. **Fixed.**
2. **Silent content loss**: All 25 unhandled node types had their content (and all children) silently deleted. **Fixed with panic + recovery to return clear error.**

---

## Implementation Phases

### Phase 1: Inline Formatting Nodes (Low Complexity)

These follow the same pattern as the existing `formatEmphasis` and `formatStrong` -- wrap children with delimiter characters.

#### 1.1 VerbatimNode (inline code)

- **Djot syntax**: `` `code` ``
- **AST structure**: VerbatimNode > TextNode (containing code text)
- **Attributes**: `$RawInlineFormatKey` (for raw inline), `$InlineMathKey`, `$DisplayMathKey`
- **Implementation**: Write backtick(s), process children, write backtick(s)
- **Edge cases**:
  - Content containing backticks needs more backtick delimiters
  - Math mode (`$...$` / `$$...$$`) uses different delimiters
  - Raw inline format (`` `code`{=html} ``) needs attribute output
- **Test fixtures**:
  - Simple inline code: `` `hello` `` -> `` `hello` ``
  - Code with backticks inside (needs double backticks)
  - Math inline: `$x^2$`
  - Display math: `$$E = mc^2$$`

#### 1.2 DeleteNode (strikethrough)

- **Djot syntax**: `{-deleted-}`
- **AST structure**: DeleteNode > [inline children]
- **Implementation**: Write `{-`, process children, write `-}`
- **Test fixtures**: `{-struck-}` -> `{-struck-}`

#### 1.3 InsertNode

- **Djot syntax**: `{+inserted+}`
- **AST structure**: InsertNode > [inline children]
- **Implementation**: Write `{+`, process children, write `+}`
- **Test fixtures**: `{+added+}` -> `{+added+}`

#### 1.4 HighlightedNode

- **Djot syntax**: `{=highlighted=}`
- **AST structure**: HighlightedNode > [inline children]
- **Implementation**: Write `{=`, process children, write `=}`
- **Test fixtures**: `{=marked=}` -> `{=marked=}`

#### 1.5 SubscriptNode

- **Djot syntax**: `{~sub~}` or `~sub~`
- **AST structure**: SubscriptNode > [inline children]
- **Implementation**: Write `{~`, process children, write `~}`
- **Normalization choice**: Always use `{~...~}` form for consistency
- **Test fixtures**: `H~2~O` -> `H{~2~}O` (or preserve original form)

#### 1.6 SuperscriptNode

- **Djot syntax**: `{^super^}` or `^super^`
- **AST structure**: SuperscriptNode > [inline children]
- **Implementation**: Write `{^`, process children, write `^}`
- **Normalization choice**: Always use `{^...^}` form for consistency
- **Test fixtures**: `x^2^` -> `x{^2^}`

#### 1.7 LineBreakNode (hard break)

- **Djot syntax**: `\` at end of line
- **AST structure**: LineBreakNode (no children, no text)
- **Implementation**: Write `\` + newline
- **Test fixtures**: `line one\` + newline + `line two`

#### 1.8 ImageNode

- **Djot syntax**: `![alt text](url)`
- **AST structure**: ImageNode (no children; alt in `"alt"` attr, src in `"src"` attr)
- **Implementation**: Write `![`, write alt attr, write `](`, write src attr, write `)`
- **Test fixtures**:
  - Simple image: `![photo](img.png)` -> `![photo](img.png)`
  - Image with empty alt: `![](img.png)`

#### 1.9 SpanNode

- **Djot syntax**: `[text]{.class #id key="val"}`
- **AST structure**: SpanNode > [inline children], with user attributes
- **Implementation**: Write `[`, process children, write `]`, write attributes block
- **Attribute output**: Need utility to serialize attributes as `{.class #id key="val"}`
- **Test fixtures**: `[text]{.highlight}` -> `[text]{.highlight}`

#### 1.10 SymbolsNode

- **Djot syntax**: `:emoji_name:`
- **AST structure**: SymbolsNode > TextNode (containing symbol name)
- **Implementation**: Write `:`, process children, write `:`
- **Test fixtures**: `:smile:` -> `:smile:`

---

### Phase 2: Block-Level Nodes (Medium Complexity)

These require managing blank line spacing, indentation, and writer state.

#### 2.1 CodeNode (fenced code block)

- **Djot syntax**: ` ``` ` ... ` ``` ` with optional language
- **AST structure**: CodeNode > TextNode (containing code), `class` attr = `language-<lang>`
- **Implementation**:
  - Write blank line if needed
  - Write ` ``` ` + language (extracted from class attr by stripping `language-` prefix)
  - Write newline
  - Write code content from child TextNode
  - Write ` ``` ` + newline
  - Set block type (new `BlockTypeCode` or reuse `BlockTypeParagraph`)
- **Edge cases**:
  - Code containing triple backticks (need more backticks for fence)
  - Empty code blocks
  - Code with trailing newlines
- **Test fixtures**:
  - Simple code block with language
  - Code block without language
  - Code block after paragraph (blank line spacing)

#### 2.2 RawNode (raw output block)

- **Djot syntax**: ` ```=html ` ... ` ``` `
- **AST structure**: RawNode > TextNode, `$RawBlockLevelKey` attr = format string
- **Implementation**: Similar to CodeNode but write ` ```=<format> ` as opener
- **Test fixtures**: Raw HTML block, raw LaTeX block

#### 2.3 QuoteNode (blockquote)

- **Djot syntax**: `> text`
- **AST structure**: QuoteNode > [block children (paragraphs, lists, etc.)]
- **Implementation**:
  - Write blank line if needed
  - Set a "in blockquote" state on Writer
  - Prefix each line with `> `
  - Process children
  - Clear blockquote state
  - Set block type
- **Complexity**: Line prefixing requires intercepting WriteString or post-processing
- **Design options**:
  - A: Buffer children output, then prepend `> ` to each line
  - B: Track blockquote depth in Writer, apply prefix in WriteString
  - Option B is more composable for nested blockquotes
- **Test fixtures**:
  - Simple blockquote
  - Blockquote with multiple paragraphs
  - Nested blockquotes (`> > text`)
  - Blockquote containing a list

#### 2.4 ThematicBreakNode (horizontal rule)

- **Djot syntax**: `***` or `---` (three or more)
- **AST structure**: ThematicBreakNode (no children, no text)
- **Implementation**: Write blank line if needed, write `***\n`, set block type
- **Normalization**: Always output `***` regardless of input chars/count
- **Test fixtures**: `***`, `---`, `*****` all -> `***`

#### 2.5 DivNode (generic container)

- **Djot syntax**: `::: classname` ... `:::`
- **AST structure**: DivNode > [block children], `class` attribute
- **Implementation**:
  - Write blank line if needed
  - Write `::: ` + class + newline
  - Process children
  - Write `:::` + newline
  - Set block type
- **Test fixtures**:
  - Div with class
  - Div with nested content
  - Div without class

---

### Phase 3: Table Nodes (High Complexity)

Tables require coordinating multiple node types and handling column alignment.

#### 3.1 TableNode, TableRowNode, TableHeaderNode, TableCellNode, TableCaptionNode

- **Djot syntax**: Pipe tables
  ```
  | Header 1 | Header 2 |
  |----------|----------|
  | Cell 1   | Cell 2   |
  ```
- **AST structure**:
  - TableNode > [TableCaptionNode?] + [TableRowNode...]
  - TableRowNode > [TableHeaderNode | TableCellNode...]
  - TableHeaderNode/TableCellNode > [inline children]
  - Alignment in `style` attr: `text-align: left|center|right`
- **Implementation approach**:
  1. Collect all cell contents first (two-pass)
  2. Calculate column widths for alignment
  3. Render header row with `|` separators
  4. Render separator row with `-` and `:` for alignment
  5. Render data rows
  6. Optionally render caption with `^ ` prefix
- **Complexity**: Highest in the project -- requires buffering, width calculation, alignment
- **Test fixtures**:
  - Simple 2x2 table
  - Table with alignment (left, center, right)
  - Table with caption
  - Table with inline formatting in cells
  - Table with varying column widths
  - Single-column table

---

### Phase 4: Definition Lists (Medium Complexity)

#### 4.1 DefinitionListNode, DefinitionTermNode, DefinitionItemNode

- **Djot syntax**:
  ```
  Term 1
  : Definition 1

  Term 2
  : Definition 2
  ```
- **AST structure**:
  - DefinitionListNode > [DefinitionTermNode, DefinitionItemNode...]
  - DefinitionTermNode > [inline children]
  - DefinitionItemNode > [block children]
- **Implementation**:
  - DefinitionListNode: Handle spacing, process children
  - DefinitionTermNode: Process inline children, write newline
  - DefinitionItemNode: Write `: `, increase indent, process children, decrease indent
- **Test fixtures**:
  - Simple definition list
  - Multi-paragraph definition
  - Definition with inline formatting
  - Multiple definitions for one term

---

### Phase 5: Reference and Footnote Nodes (Medium-High Complexity)

#### 5.1 ReferenceDefNode

- **Djot syntax**: `[label]: url`
- **AST structure**: ReferenceDefNode with `$ReferenceKey` attribute
- **Note**: These may be consumed by the parser and not appear in the final AST. Need to verify whether they survive into the tree or are fully resolved during parsing.
- **Implementation**: If present in AST, write `[label]: url` format
- **Test fixtures**: Reference-style links with definitions

#### 5.2 FootnoteDefNode

- **Djot syntax**: `[^label]: content`
- **AST structure**: FootnoteDefNode > [block children], `$ReferenceKey` attr
- **Implementation**:
  - Write `[^label]:` + space
  - Increase indent
  - Process children
  - Decrease indent
- **Test fixtures**:
  - Simple footnote
  - Multi-paragraph footnote
  - Footnote with inline formatting

---

## Shared Infrastructure Needed

### Attribute Serialization

Multiple node types (SpanNode, DivNode, CodeNode with attrs, etc.) need to serialize attributes back to djot syntax: `{.class #id key="value"}`.

```go
func formatAttributes(attrs tokenizer.Attributes) string {
    // Skip internal $ attributes
    // Format class as .class
    // Format id as #id
    // Format others as key="value"
}
```

### Blockquote Line Prefixing

The Writer needs enhancement to support line prefixing for blockquotes. Options:
- Add `linePrefix string` field to Writer
- Apply prefix in `WriteString` after each newline
- Support nesting (multiple prefixes)

### New BlockTypes

```go
const (
    BlockTypeCode         // for code/raw blocks
    BlockTypeThematicBreak // for horizontal rules
    BlockTypeDiv          // for div containers
    BlockTypeTable        // for tables
    BlockTypeDefinition   // for definition lists
)
```

---

## Test Strategy

### Unit Tests (fixture-based)

Add new fixture files or extend `testdata/formatter/basic.txt`:

- `testdata/formatter/inline.txt` -- VerbatimNode, DeleteNode, InsertNode, etc.
- `testdata/formatter/blocks.txt` -- CodeNode, QuoteNode, ThematicBreakNode, DivNode
- `testdata/formatter/tables.txt` -- all table variations
- `testdata/formatter/definitions.txt` -- definition lists
- `testdata/formatter/footnotes.txt` -- footnotes and references

### Integration Tests

Extend `process_test.go` with:
- Documents combining multiple node types
- Real-world djot document samples
- Round-trip formatting (format, then format again, output should be identical)
- Check mode (`-c`) with each new node type

### Idempotency Tests

Every formatter must produce idempotent output: `format(format(input)) == format(input)`. Add a generic test that runs every fixture through the formatter twice and asserts the outputs match.

### Edge Case Tests

- Empty documents
- Documents with only whitespace
- Very deeply nested structures
- Mixed node types in sequence
- Attributes on every node type that supports them

---

## Existing Bugs and Gaps (Beyond Missing Nodes)

### Bugs

1. **Link URL attribute key** -- `formatLink` used `"url"` instead of `"href"`. **Fixed in this session.**
2. **PROJECT_STATUS.md inaccuracy** -- Claims unsupported types "will pass through." They are silently dropped. Should be corrected.

### Testing Gaps

1. **No idempotency test** -- No test verifies that formatting is stable across multiple passes
2. **No check mode (`-c`) tests** -- `process_test.go` only tests write mode
3. **No stdin/stdout integration tests** -- Only file-based tests exist
4. **No error path tests in process_test.go** -- Missing tests for: file not found, permission denied, invalid arguments
5. **No test for `-o` (output file) flag** -- Only `-w` is tested
6. **Ordered list indentation** -- Nested ordered lists use 2-space indent but `1. ` is 3 chars wide, so continuation lines may misalign
7. **Sparse (loose) lists** -- The `$SparseListNodeKey` attribute is not checked; tight and loose lists may format identically when they shouldn't
8. **SLW interaction with non-paragraph blocks** -- SLW wrapping is only applied inside paragraphs, but text in definition items, blockquote paragraphs, and table cells should also be wrapped

### Missing Features

1. **Attribute preservation** -- User-defined attributes (`{.class #id}`) on any node are silently dropped
2. **Heading ID preservation** -- Auto-generated section IDs are not output
3. **Ordered list style** -- Always outputs `1.` but djot supports `a.`, `A.`, etc.
4. **Ordered list start number** -- Always starts at 1 but input may start at other numbers

---

## Suggested Implementation Order

1. **Phase 1** (inline nodes) -- Lowest risk, highest coverage gain, most are trivial
2. **Phase 2.4** (ThematicBreakNode) -- Trivial, no children
3. **Phase 2.1-2.2** (CodeNode, RawNode) -- Common in real docs, moderate complexity
4. **Phase 2.3** (QuoteNode) -- Requires Writer changes for line prefixing
5. **Phase 2.5** (DivNode) -- Similar to QuoteNode
6. **Phase 4** (definition lists) -- Moderate complexity
7. **Phase 5** (references, footnotes) -- Need to verify AST behavior first
8. **Phase 3** (tables) -- Most complex, save for last
9. **Idempotency and integration tests** -- After each phase
10. **Attribute serialization** -- Cross-cutting, implement when SpanNode/DivNode are tackled

---

## Estimated Scope

- **Phase 1**: ~10 small functions, ~15 test fixtures
- **Phase 2**: ~5 functions + Writer enhancement, ~20 test fixtures
- **Phase 3**: ~5 functions + table rendering logic, ~10 test fixtures
- **Phase 4**: ~3 functions, ~8 test fixtures
- **Phase 5**: ~2 functions, ~6 test fixtures
- **Infrastructure**: Attribute serializer, Writer prefix support, new BlockTypes
