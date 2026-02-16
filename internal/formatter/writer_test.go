package formatter_test

import (
	"testing"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/stretchr/testify/assert"
)

func TestWriter_String_TrailingNewlineNormalization(t *testing.T) {
	tests := []struct {
		name     string
		writes   []string
		expected string
	}{
		{
			name:     "no trailing newline becomes single newline",
			writes:   []string{"Hello"},
			expected: "Hello\n",
		},
		{
			name:     "single trailing newline preserved",
			writes:   []string{"Hello\n"},
			expected: "Hello\n",
		},
		{
			name:     "multiple trailing newlines normalized to one",
			writes:   []string{"Hello\n\n\n"},
			expected: "Hello\n",
		},
		{
			name:     "empty content becomes single newline",
			writes:   []string{""},
			expected: "\n",
		},
		{
			name:     "only newlines become single newline",
			writes:   []string{"\n\n\n"},
			expected: "\n",
		},
		{
			name:     "multiple writes with trailing newlines",
			writes:   []string{"Line 1\n", "Line 2\n", "Line 3\n\n"},
			expected: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:     "content with internal newlines and no trailing",
			writes:   []string{"Line 1\nLine 2\nLine 3"},
			expected: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:     "content with internal and multiple trailing newlines",
			writes:   []string{"Line 1\nLine 2\n\n\n"},
			expected: "Line 1\nLine 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := formatter.NewWriter()

			for _, s := range tt.writes {
				w.WriteString(s)
			}

			result := w.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriter_String_PreservesInternalNewlines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "single internal newline",
			content:  "Line 1\nLine 2",
			expected: "Line 1\nLine 2\n",
		},
		{
			name:     "multiple internal newlines",
			content:  "Line 1\n\nLine 2\n\n\nLine 3",
			expected: "Line 1\n\nLine 2\n\n\nLine 3\n",
		},
		{
			name:     "blank lines preserved",
			content:  "Para 1\n\nPara 2\n\nPara 3",
			expected: "Para 1\n\nPara 2\n\nPara 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := formatter.NewWriter()
			w.WriteString(tt.content)

			result := w.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriter_String_Idempotency(t *testing.T) {
	tests := []string{
		"Hello",
		"Hello\n",
		"Hello\n\n\n",
		"Line 1\nLine 2\n",
		"Para 1\n\nPara 2\n\n\n",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			w1 := formatter.NewWriter()
			w1.WriteString(input)
			first := w1.String()

			w2 := formatter.NewWriter()
			w2.WriteString(first)
			second := w2.String()

			assert.Equal(t, first, second, "String() should be idempotent")
		})
	}
}
