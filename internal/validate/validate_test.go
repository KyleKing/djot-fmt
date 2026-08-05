package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KyleKing/djot-fmt/internal/validate"
)

func categories(diffs []validate.Difference) []validate.Category {
	out := make([]validate.Category, 0, len(diffs))
	for _, d := range diffs {
		out = append(out, d.Category)
	}

	return out
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		output   string
		rejected []validate.Category
		waived   []validate.Category
	}{
		{
			name:   "identical",
			input:  "hello world\n",
			output: "hello world\n",
		},
		{
			name:   "whitespace only change is invisible",
			input:  "one sentence. two sentence.\n",
			output: "one sentence.\ntwo sentence.\n",
		},
		{
			name:     "frontmatter turned into thematic breaks",
			input:    "---\ntitle: Hi\n---\n\nbody\n",
			output:   "***\n\ntitle: Hi\n\n***\n\nbody\n",
			rejected: []validate.Category{validate.CategoryBlockKind, validate.CategoryBlockKind},
		},
		{
			name:     "dropped content",
			input:    "keep this\n\ndrop this\n",
			output:   "keep this\n",
			rejected: []validate.Category{validate.CategoryText, validate.CategoryText, validate.CategoryLooseTight},
		},
		{
			name:     "merged paragraphs",
			input:    "one\n\ntwo\n",
			output:   "one\ntwo\n",
			rejected: []validate.Category{validate.CategoryLooseTight},
		},
		{
			name:     "start attribute lost",
			input:    "5. five\n6. six\n",
			output:   "1. five\n1. six\n",
			rejected: []validate.Category{validate.CategoryListAttrs},
		},
		{
			name:     "list type lost",
			input:    "i. one\nii. two\n",
			output:   "a. one\na. two\n",
			rejected: []validate.Category{validate.CategoryListAttrs},
		},
		{
			name:     "heading level changed",
			input:    "# Title\n",
			output:   "## Title\n",
			rejected: []validate.Category{validate.CategoryBlockKind, validate.CategoryBlockKind},
		},
		{
			name:   "emphasis dropped",
			input:  "_word_\n",
			output: "word\n",
			rejected: []validate.Category{
				validate.CategoryUnclassified, validate.CategoryUnclassified,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := validate.Compare([]byte(tt.input), []byte(tt.output), validate.Options{})

			assert.Equal(t, tt.rejected, nilIfEmpty(categories(result.Rejected)))
			assert.Equal(t, tt.waived, nilIfEmpty(categories(result.Waived)))
			assert.Equal(t, len(tt.rejected) == 0, result.OK())
		})
	}
}

func TestCompareFromMD(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		output   string
		rejected []validate.Category
		waived   []validate.Category
	}{
		{
			name:   "nesting a markdown list is waived",
			input:  "- a\n  - b\n",
			output: "- a\n\n  - b\n",
			waived: []validate.Category{validate.CategoryListNesting, validate.CategoryListNesting},
		},
		{
			name:   "tight list becoming loose is waived",
			input:  "- a\n- b\n",
			output: "- a\n\n- b\n",
			waived: []validate.Category{
				validate.CategoryLooseTight, validate.CategoryLooseTight,
				validate.CategoryLooseTight, validate.CategoryLooseTight,
			},
		},
		{
			name:     "content escaping a list item is never waived",
			input:    "- a\n\n- b\n\n  nested para\n",
			output:   "- a\n\n- b\n\nnested para\n",
			rejected: []validate.Category{validate.CategoryListNesting},
			waived:   []validate.Category{validate.CategoryListNesting},
		},
		{
			name:     "merging paragraphs is never waived",
			input:    "one\n\ntwo\n",
			output:   "one\ntwo\n",
			rejected: []validate.Category{validate.CategoryLooseTight},
		},
		{
			name:     "frontmatter loss is never waived",
			input:    "---\ntitle: Hi\n---\n\nbody\n",
			output:   "***\n\ntitle: Hi\n\n***\n\nbody\n",
			rejected: []validate.Category{validate.CategoryBlockKind, validate.CategoryBlockKind},
		},
		{
			name:     "list attributes are never waived",
			input:    "5. five\n6. six\n",
			output:   "1. five\n1. six\n",
			rejected: []validate.Category{validate.CategoryListAttrs},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := validate.Compare([]byte(tt.input), []byte(tt.output), validate.Options{FromMD: true})

			assert.Equal(t, tt.rejected, nilIfEmpty(categories(result.Rejected)))
			assert.Equal(t, tt.waived, nilIfEmpty(categories(result.Waived)))
		})
	}
}

func TestCompareIgnoresFormattedCode(t *testing.T) {
	t.Parallel()

	input := "``` python\nx=1\n```\n"
	output := "``` python\nx = 1\n```\n"

	assert.False(t, validate.Compare([]byte(input), []byte(output), validate.Options{}).OK())
	assert.True(t, validate.Compare([]byte(input), []byte(output), validate.Options{
		CodeLanguages: []string{"python"},
	}).OK())
}

func TestCompareKeepsUnformattedCode(t *testing.T) {
	t.Parallel()

	input := "``` rust\nlet x=1;\n```\n"
	output := "``` rust\nlet x = 1;\n```\n"

	assert.False(t, validate.Compare([]byte(input), []byte(output), validate.Options{
		CodeLanguages: []string{"python"},
	}).OK())
}

func nilIfEmpty(c []validate.Category) []validate.Category {
	if len(c) == 0 {
		return nil
	}

	return c
}
