package formatter

import (
	"testing"

	"github.com/kyleking/djot-fmt/internal/slw"
	"github.com/sivukhin/godjot/v2/djot_parser"
	"github.com/stretchr/testify/assert"
)

func TestFormat_SimpleParagraph(t *testing.T) {
	ast := []djot_parser.TreeNode[djot_parser.DjotNode]{
		{
			Type: djot_parser.ParagraphNode,
			Children: []djot_parser.TreeNode[djot_parser.DjotNode]{
				{Type: djot_parser.TextNode, Text: []byte("Hello, world!")},
			},
		},
	}

	result := Format(ast)
	expected := "Hello, world!\n"
	assert.Equal(t, expected, result)
}

func TestFormat_SimpleList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tight list",
			input:    "- one\n- two\n- three\n",
			expected: "- one\n- two\n- three\n",
		},
		{
			name:     "list with paragraph after",
			input:    "- one\n- two\n\nParagraph after.\n",
			expected: "- one\n- two\n\nParagraph after.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))
			result := Format(ast)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormat_NestedList(t *testing.T) {
	input := "- one\n- two\n\n  - sub\n  - sub\n"
	expected := "- one\n- two\n\n  - sub\n  - sub\n"

	ast := djot_parser.BuildDjotAst([]byte(input))
	result := Format(ast)
	assert.Equal(t, expected, result)
}

func TestFormat_InlineFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "emphasis",
			input:    "_italic_\n",
			expected: "_italic_\n",
		},
		{
			name:     "strong",
			input:    "*bold*\n",
			expected: "*bold*\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))
			result := Format(ast)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormat_OrderedList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple numbered list - normalized to 1.",
			input:    "1. First\n2. Second\n",
			expected: "1. First\n1. Second\n",
		},
		{
			name:     "nested ordered list - all items normalized",
			input:    "1. First\n\n   1. Nested\n   2. Items\n",
			expected: "1. First\n\n  1. Nested\n  1. Items\n",
		},
		{
			name:     "non-standard numbering gets normalized",
			input:    "5. Fifth\n6. Sixth\n",
			expected: "1. Fifth\n1. Sixth\n",
		},
		{
			name:     "tight ordered list",
			input:    "1. one\n2. two\n3. three\n",
			expected: "1. one\n1. two\n1. three\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))
			result := Format(ast)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormat_TaskList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unchecked task",
			input:    "- [ ] Task\n",
			expected: "- [ ] Task\n",
		},
		{
			name:     "checked task",
			input:    "- [x] Done\n",
			expected: "- [x] Done\n",
		},
		{
			name:     "mixed tasks",
			input:    "- [ ] Todo\n- [x] Done\n- [ ] Pending\n",
			expected: "- [ ] Todo\n- [x] Done\n- [ ] Pending\n",
		},
		{
			name:     "task with inline formatting",
			input:    "- [ ] Inline _formatting_ *bold*\n",
			expected: "- [ ] Inline _formatting_ *bold*\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))
			result := Format(ast)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormat_SLWWrapping(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		slwConfig *slw.Config
		expected  string
	}{
		{
			name:  "basic SLW wrapping in paragraph",
			input: "This is a long sentence that exceeds the minimum length. It should be wrapped! Does it work? Yes it does.",
			slwConfig: &slw.Config{
				Enabled:       true,
				Markers:       ".!?",
				MinLineLength: 40,
				MaxLineWidth:  88,
				Abbreviations: slw.DefaultConfig().Abbreviations,
			},
			expected: "This is a long sentence that exceeds the minimum length.\nIt should be wrapped!\nDoes it work?\nYes it does.\n",
		},
		{
			name:  "SLW disabled",
			input: "This is a long sentence. It should not be wrapped! Even though it's long?",
			slwConfig: &slw.Config{
				Enabled:       false,
				Markers:       ".!?",
				MinLineLength: 40,
				MaxLineWidth:  88,
				Abbreviations: slw.DefaultConfig().Abbreviations,
			},
			expected: "This is a long sentence. It should not be wrapped! Even though it’s long?\n",
		},
		{
			name:  "SLW with abbreviations",
			input: "Dr. Smith met with Prof. Johnson yesterday morning. They discussed important topics that were quite significant.",
			slwConfig: &slw.Config{
				Enabled:       true,
				Markers:       ".!?",
				MinLineLength: 40,
				MaxLineWidth:  88,
				Abbreviations: slw.DefaultConfig().Abbreviations,
			},
			expected: "Dr. Smith met with Prof. Johnson yesterday morning.\nThey discussed important topics that were quite significant.\n",
		},
		{
			name:  "aggressive mode (min 0)",
			input: "Short sentence. Another one.",
			slwConfig: &slw.Config{
				Enabled:       true,
				Markers:       ".!?",
				MinLineLength: 0,
				MaxLineWidth:  88,
				Abbreviations: slw.DefaultConfig().Abbreviations,
			},
			expected: "Short sentence.\nAnother one.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))
			result := FormatWithConfig(ast, tt.slwConfig)
			// Handle smart quote conversion by djot parser
			if !assert.Equal(t, tt.expected, result) {
				t.Logf("Result length: %d, Expected length: %d", len(result), len(tt.expected))
			}
		})
	}
}
