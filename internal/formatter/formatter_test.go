package formatter_test

import (
	"path/filepath"
	"testing"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/testutil"
	"github.com/sivukhin/godjot/v2/djot_parser"
	"github.com/stretchr/testify/assert"
)

func TestFormat_AllNodeTypesSupported(t *testing.T) {
	supportedInputs := []struct {
		name  string
		input string
	}{
		{"inline code", "`code`\n"},
		{"code block", "```\ncode\n```\n"},
		{"table", "| header |\n|---|\n| cell |\n"},
		{"definition list", "term\n: definition\n"},
		{"blockquote", "> quote\n"},
		{"thematic break", "***\n"},
		{"reference", "[ref]: https://example.com\n"},
	}

	for _, tt := range supportedInputs {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))
			result := formatter.Format(ast)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormat_SimpleParagraphAST(t *testing.T) {
	ast := []djot_parser.TreeNode[djot_parser.DjotNode]{
		{
			Type: djot_parser.ParagraphNode,
			Children: []djot_parser.TreeNode[djot_parser.DjotNode]{
				{Type: djot_parser.TextNode, Text: []byte("Hello, world!")},
			},
		},
	}

	result := formatter.Format(ast)
	expected := "Hello, world!\n"
	assert.Equal(t, expected, result)
}

func TestFormat_BasicFixtures(t *testing.T) {
	path := filepath.Join("../../testdata/formatter", "basic.txt")

	fixtures, err := testutil.ReadFixtures(path)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(fixture.Input))
			result := formatter.Format(ast)

			if !assert.Equal(t, fixture.Expected, result) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
			}
		})
	}
}

func TestFormat_SLWFixtures(t *testing.T) {
	path := filepath.Join("../../testdata/formatter", "slw.txt")

	fixtures, err := testutil.ReadFixtures(path)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			config := testutil.ConfigFromOptions(fixture.Options)
			ast := djot_parser.BuildDjotAst([]byte(fixture.Input))
			result := formatter.FormatWithConfig(ast, config)

			if !assert.Equal(t, fixture.Expected, result) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
			}
		})
	}
}

func TestFormat_InlineFixtures(t *testing.T) {
	path := filepath.Join("../../testdata/formatter", "inline.txt")

	fixtures, err := testutil.ReadFixtures(path)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(fixture.Input))
			result := formatter.Format(ast)

			if !assert.Equal(t, fixture.Expected, result) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
			}
		})
	}
}

func TestFormat_Idempotency(t *testing.T) {
	fixtureFiles := []string{
		"basic.txt",
		"inline.txt",
		"slw.txt",
	}

	for _, filename := range fixtureFiles {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join("../../testdata/formatter", filename)

			fixtures, err := testutil.ReadFixtures(path)
			if err != nil {
				t.Fatalf("Failed to read fixtures from %s: %v", filename, err)
			}

			for _, fixture := range fixtures {
				t.Run(fixture.Title, func(t *testing.T) {
					ast1 := djot_parser.BuildDjotAst([]byte(fixture.Input))
					first := formatter.Format(ast1)

					ast2 := djot_parser.BuildDjotAst([]byte(first))
					second := formatter.Format(ast2)

					if !assert.Equal(t, first, second) {
						t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
						t.Logf("Input: %q", fixture.Input)
					}
				})
			}
		})
	}
}
