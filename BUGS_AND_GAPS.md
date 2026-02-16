# Bugs and Testing Gaps

Comprehensive plan to address every known bug, testing gap, and missing feature in djot-fmt.

---

## 1. Registry Not Wired to Phase 1 Handlers

**Status**: Bug (implemented handlers exist but aren't registered)

**Problem**: `formatVerbatim`, `formatDelete`, `formatInsert`, `formatHighlighted`, `formatSubscript`, `formatSuperscript`, `formatLineBreak`, `formatImage`, `formatSpan`, and `formatSymbols` are defined in `formatter.go` but the registry still maps all of them to `formatUnsupported`.

**Fix**:
- Update `defaultRegistry` to point each inline node type to its real handler
- Update `TestFormat_UnsupportedNodePanics` to remove "inline code" and "image" cases (they will no longer panic)
- Add fixture tests for each newly-wired handler

**Files**:
- `internal/formatter/formatter.go` -- update registry entries
- `internal/formatter/formatter_test.go` -- update panic test, add new cases
- `testdata/formatter/basic.txt` -- add fixture cases for inline code, delete, insert, highlight, subscript, superscript, line break, image, span, symbols

**Test fixtures to add** (in `basic.txt`):

```
inline code
.
Some `code` here.
.
Some `code` here.
.

strikethrough
.
{-deleted text-}
.
{-deleted text-}
.

inserted text
.
{+added text+}
.
{+added text+}
.

highlighted text
.
{=marked text=}
.
{=marked text=}
.

subscript
.
H{~2~}O
.
H{~2~}O
.

superscript
.
x{^2^}
.
x{^2^}
.

line break
.
line one\
line two
.
line one\
line two
.

image
.
![alt text](image.png)
.
![alt text](image.png)
.

symbol
.
:smile:
.
:smile:
.

span
.
[text]{.class}
.
[text]
.
```

---

## 2. No Idempotency Tests

**Status**: Testing gap (high priority)

**Problem**: Nothing verifies that `format(format(input)) == format(input)`. A formatter that isn't idempotent will cause `-c` to always fail after `-w`, or produce different results on repeated runs.

**Fix**: Add a generic test that runs every basic fixture through the formatter twice and asserts the outputs match.

**File**: `internal/formatter/formatter_test.go`

**Implementation**:

```go
func TestFormat_Idempotency(t *testing.T) {
    path := filepath.Join("../../testdata/formatter", "basic.txt")
    fixtures, err := readFixtures(path)
    require.NoError(t, err)

    for _, fixture := range fixtures {
        t.Run(fixture.Title, func(t *testing.T) {
            // First pass
            ast1 := djot_parser.BuildDjotAst([]byte(fixture.Input))
            first := formatter.Format(ast1)

            // Second pass
            ast2 := djot_parser.BuildDjotAst([]byte(first))
            second := formatter.Format(ast2)

            assert.Equal(t, first, second,
                "formatter is not idempotent for %q", fixture.Title)
        })
    }
}
```

**Note**: This test will likely expose existing bugs (e.g., the extra blank line between paragraph and list may compound on each pass). Those bugs should be tracked and fixed.

---

## 3. No Check Mode (`-c`) Tests

**Status**: Testing gap (high priority)

**Problem**: `process_test.go` only tests `-w` mode. The `-c` flag (which compares original to formatted and exits non-zero if different) has zero test coverage. The `checkFormatted` function compares raw bytes, which means trailing whitespace differences or newline normalization issues could cause false positives/negatives.

**Fix**: Add tests for check mode in `process_test.go`.

**File**: `internal/iohelper/process_test.go`

**Test cases**:

```go
func TestProcessFile_Check(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "already formatted",
            input:   "Hello, world.\n",
            wantErr: false,
        },
        {
            name:    "needs formatting",
            input:   "A paragraph.\n\n- item\n",  // missing extra blank line
            wantErr: true,
        },
        {
            name:    "empty file",
            input:   "",
            wantErr: false,  // verify behavior
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()
            inputFile := filepath.Join(tmpDir, "test.djot")
            err := os.WriteFile(inputFile, []byte(tt.input), 0600)
            require.NoError(t, err)

            opts := &iohelper.Options{
                Check:      true,
                InputFiles: []string{inputFile},
                SlwMarkers: ".!?",
                SlwWrap:    88,
                SlwMinLine: 40,
            }

            err = iohelper.ProcessFile(opts, inputFile)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }

            // Verify file is never modified in check mode
            result, readErr := os.ReadFile(inputFile)
            require.NoError(t, readErr)
            assert.Equal(t, tt.input, string(result))
        })
    }
}
```

---

## 4. No Output File (`-o`) Tests

**Status**: Testing gap (medium priority)

**Problem**: The `-o` flag creates a new output file but has no test coverage. Edge cases include: file creation, overwriting existing files, and ensuring the input file is not modified.

**Fix**: Add tests in `process_test.go`.

**File**: `internal/iohelper/process_test.go`

**Test cases**:

```go
func TestProcessFile_OutputFile(t *testing.T) {
    tmpDir := t.TempDir()
    inputFile := filepath.Join(tmpDir, "input.djot")
    outputFile := filepath.Join(tmpDir, "output.djot")

    input := "Hello, world.\n"
    err := os.WriteFile(inputFile, []byte(input), 0600)
    require.NoError(t, err)

    opts := &iohelper.Options{
        OutputFile: outputFile,
        InputFiles: []string{inputFile},
        SlwMarkers: ".!?",
        SlwWrap:    88,
        SlwMinLine: 40,
    }

    err = iohelper.ProcessFile(opts, inputFile)
    require.NoError(t, err)

    // Verify output file was created with formatted content
    result, err := os.ReadFile(outputFile)
    require.NoError(t, err)
    assert.Equal(t, "Hello, world.\n", string(result))

    // Verify input file was not modified
    original, err := os.ReadFile(inputFile)
    require.NoError(t, err)
    assert.Equal(t, input, string(original))
}
```

---

## 5. No Error Path Tests in process_test.go

**Status**: Testing gap (medium priority)

**Problem**: No tests for file-not-found, permission errors, or other I/O failures in `ProcessFile`.

**Fix**: Add error path tests.

**File**: `internal/iohelper/process_test.go`

**Test cases**:

```go
func TestProcessFile_FileNotFound(t *testing.T) {
    opts := &iohelper.Options{
        InputFiles: []string{"/nonexistent/file.djot"},
        SlwMarkers: ".!?",
        SlwWrap:    88,
        SlwMinLine: 40,
    }

    err := iohelper.ProcessFile(opts, "/nonexistent/file.djot")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "reading input file")
}

func TestProcessFile_WriteToReadOnlyDir(t *testing.T) {
    tmpDir := t.TempDir()
    inputFile := filepath.Join(tmpDir, "test.djot")
    err := os.WriteFile(inputFile, []byte("Hello.\n"), 0600)
    require.NoError(t, err)

    // Make directory read-only after creating the file
    err = os.Chmod(tmpDir, 0555)
    require.NoError(t, err)
    t.Cleanup(func() { os.Chmod(tmpDir, 0755) })

    opts := &iohelper.Options{
        Write:      true,
        InputFiles: []string{inputFile},
        SlwMarkers: ".!?",
        SlwWrap:    88,
        SlwMinLine: 40,
    }

    err = iohelper.ProcessFile(opts, inputFile)
    assert.Error(t, err)
}
```

---

## 6. Sparse (Loose) Lists Not Distinguished

**Status**: Bug (medium priority)

**Problem**: The godjot parser sets `$SparseListNodeKey = "true"` on lists that have blank lines between items (loose/sparse lists). The formatter ignores this attribute entirely, so tight and loose lists produce identical output. In djot, the distinction matters: loose lists render `<p>` tags inside `<li>`, tight lists don't.

**Current behavior**:
```
Input (loose):        Output (wrong):
- item 1              - item 1
                      - item 2
- item 2
```

**Expected behavior**: Loose lists should preserve blank lines between items.

**Fix**:
- In `formatUnorderedList`/`formatOrderedList`/`formatTaskList`: check `$SparseListNodeKey` attribute
- Pass sparse flag down to `formatListItem` (via Writer state or parent attribute check)
- In `formatListItem`: if list is sparse, add trailing blank line after each item

**Files**:
- `internal/formatter/formatter.go` -- check sparse attribute, adjust newline output
- `internal/formatter/writer.go` -- potentially add `inSparseList` state
- `testdata/formatter/basic.txt` -- add loose list fixture

**Test fixtures**:

```
loose unordered list
.
- item 1

- item 2

- item 3
.
- item 1

- item 2

- item 3
.

loose ordered list
.
1. first

2. second
.
1. first

1. second
.
```

---

## 7. Ordered List Indentation Mismatch

**Status**: Bug (low priority)

**Problem**: Nested ordered lists use 2-space indentation (matching unordered lists), but the `1. ` marker is 3 characters wide. Continuation lines within an ordered list item should align with the content after the marker, not with the marker itself.

**Current behavior**:
```
1. First item
  1. Nested item    <-- 2 spaces, but should be 3 to align
```

**Correct djot behavior**:
```
1. First item

   1. Nested item   <-- 3 spaces to match "1. " width
```

**Fix**: The indent width should be based on the marker width, not a fixed 2 spaces.

**Files**:
- `internal/formatter/writer.go` -- make indent width configurable per level or pass marker width
- `internal/formatter/formatter.go` -- pass marker width to indent system

**Approach options**:
- A: Change `IncreaseIndent()` to accept a width parameter
- B: Push indent strings onto a stack instead of counting levels
- Option B is more flexible for mixed list types

---

## 8. Ordered List Style Always `1.`

**Status**: Bug (low priority)

**Problem**: Djot supports multiple ordered list styles: `1.`, `a.`, `A.`, `i.`, `I.`, and `)` variants. The formatter always outputs `1.` regardless of input style.

**Current behavior**: `a. First` -> `1. First` (style lost)

**Fix**:
- Check the `type` attribute on OrderedListNode (values: `"1"`, `"a"`, `"A"`, etc.)
- Check the `start` attribute for non-1 starting numbers
- Pass list style to `formatListItem` via Writer state or parent attribute

**Files**:
- `internal/formatter/formatter.go` -- read `type` and `start` attributes from parent

**Test fixtures**:

```
alphabetical ordered list
.
a. Alpha
b. Bravo
.
a. Alpha
a. Bravo
.
```

---

## 9. User Attributes Silently Dropped

**Status**: Bug (medium priority)

**Problem**: Djot allows user-defined attributes on most elements: `{.class #id key="value"}`. These are stored in `Node.Attributes` but none of the formatters output them. The `formatSpan` handler has a TODO noting this.

**Example**: `*bold*{.highlight}` -> `*bold*` (attribute lost)

**Fix**: Implement an attribute serializer utility and call it from each handler that supports user attributes.

**Files**:
- `internal/formatter/formatter.go` -- add `formatAttributes` helper, call from relevant handlers
- New test fixtures for attributed elements

**Implementation**:

```go
func formatAttributes(attrs tokenizer.Attributes) string {
    var parts []string
    for _, key := range attrs.Keys {
        if strings.HasPrefix(key, "$") {
            continue  // skip internal attributes
        }
        val := attrs.Map[key]
        switch key {
        case "class":
            for _, cls := range strings.Fields(val) {
                parts = append(parts, "."+cls)
            }
        case "id":
            parts = append(parts, "#"+val)
        default:
            parts = append(parts, key+`="`+val+`"`)
        }
    }
    if len(parts) == 0 {
        return ""
    }
    return "{" + strings.Join(parts, " ") + "}"
}
```

**Affected handlers**: All inline (emphasis, strong, link, verbatim, etc.) and block (heading, paragraph, list, etc.) handlers that can have user attributes.

---

## 10. VerbatimNode Edge Cases

**Status**: Bug (exists in current Phase 1 implementation)

**Problem**: `formatVerbatim` always uses single backticks, but inline code containing backticks requires double (or more) backtick delimiters. Also, `$InlineMathKey` and `$DisplayMathKey` attribute checks use `!= ""` but these are marker attributes -- they exist with empty string value. Need to verify whether `TryGet` or `Get` returns empty string for present-but-empty vs absent attributes.

**Fix**:
- Check if child TextNode content contains backticks; if so, use more backticks for delimiter
- Verify math attribute detection works correctly with godjot's `Attributes.Get` behavior

**Files**:
- `internal/formatter/formatter.go` -- update `formatVerbatim`
- Add test fixtures with backtick-containing code

**Test fixtures**:

```
code with backtick
.
`` code`with`backticks ``
.
`` code`with`backticks ``
.
```

---

## 11. SLW Not Applied in All Text Contexts

**Status**: Gap (low priority)

**Problem**: SLW wrapping only applies when `state.Writer.InParagraph()` is true. Text inside blockquotes, definition items, table cells, and other containers should also be wrapped, but those contexts don't set `inParagraph`. Once those handlers are implemented, the SLW check should be broadened or each block handler should set paragraph context when appropriate.

**Fix**: When implementing QuoteNode, DefinitionItemNode, TableCellNode, etc., ensure their child paragraphs properly set `inParagraph = true` (which they will, since ParagraphNode sets it). Verify this works for nested paragraph contexts.

**Files**: No changes needed if ParagraphNode children are properly processed. Add integration tests to verify SLW works inside block containers.

---

## 12. `formatImage` Uses String Literals Instead of Constants

**Status**: Code quality (low priority)

**Problem**: `formatImage` uses `"alt"` and `"src"` string literals, but godjot defines `ImgAltKey` and `ImgSrcKey` constants.

**Fix**:

```go
func formatImage(state djot_parser.ConversionState[*Writer], _ func(djot_parser.Children)) {
    alt := state.Node.Attributes.Get(djot_parser.ImgAltKey)
    src := state.Node.Attributes.Get(djot_parser.ImgSrcKey)
    state.Writer.WriteString("![" + alt + "](" + src + ")")
}
```

**File**: `internal/formatter/formatter.go`

---

## 13. `formatHeading` Uses String Literal Instead of Constant

**Status**: Code quality (low priority)

**Problem**: `formatHeading` uses `"$HeadingLevelKey"` string literal, but godjot defines a `HeadingLevelKey` constant.

**Fix**:

```go
levelMarker := state.Node.Attributes.Get(djot_parser.HeadingLevelKey)
```

**File**: `internal/formatter/formatter.go`

---

## 14. No Multi-File Integration Test

**Status**: Testing gap (low priority)

**Problem**: `main.go` handles multiple files in a loop with error aggregation for check mode, but this flow has no test coverage. The logic for continuing after errors in check mode vs stopping immediately in write mode is untested.

**Fix**: Add a test that processes multiple files and verifies:
- All files are processed in check mode even when some fail
- Processing stops on first error in write mode
- Error messages include filenames

**Note**: Testing `main.go` directly requires either extracting `run()` to be testable or using `exec.Command` for subprocess testing. The `run()` function is already separated, but it reads from `os.Args`. Consider refactoring to accept args, or test via subprocess.

---

## 15. `Writer.String()` Trailing Newline Normalization

**Status**: Potential bug (low priority)

**Problem**: `Writer.String()` does `strings.TrimRight(result, "\n") + "\n"`, which collapses all trailing newlines to exactly one. This is generally correct, but could mask bugs where too many or too few newlines are emitted by handlers. It also means the formatter can never output a file that ends with multiple newlines (which djot allows).

**Fix**: Consider whether this normalization should be removed in favor of handlers being correct about newline output. If kept, add a test that verifies the behavior.

---

## Priority Order

| # | Issue | Priority | Effort |
|---|-------|----------|--------|
| 1 | Wire Phase 1 handlers into registry | Critical | Small |
| 2 | Idempotency tests | High | Small |
| 3 | Check mode tests | High | Small |
| 10 | VerbatimNode edge cases | High | Small |
| 6 | Sparse list distinction | Medium | Medium |
| 9 | User attributes dropped | Medium | Medium |
| 4 | Output file tests | Medium | Small |
| 5 | Error path tests | Medium | Small |
| 12 | Image attribute constants | Low | Trivial |
| 13 | Heading attribute constant | Low | Trivial |
| 7 | Ordered list indentation | Low | Medium |
| 8 | Ordered list style | Low | Small |
| 11 | SLW in all contexts | Low | Small |
| 14 | Multi-file integration test | Low | Medium |
| 15 | Trailing newline normalization | Low | Small |
