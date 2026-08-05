// The godjot tokenizer records matched symbols in a package-level map, so parsing cannot run concurrently.
//
//nolint:paralleltest // Tests parse djot through godjot, whose tokenizer is not goroutine safe.
package formatter_test

import (
	"path/filepath"
	"testing"

	"github.com/sivukhin/godjot/v2/djot_parser"
	"github.com/stretchr/testify/assert"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/testutil"
)

func fixturePath(filename string) string {
	return filepath.Join("..", "..", "testdata", "formatter", filename)
}

func formatDefault(fixture testutil.Fixture) string {
	return formatter.Format(djot_parser.BuildDjotAst([]byte(fixture.Input)))
}

func formatWithFixtureOptions(fixture testutil.Fixture) string {
	config := testutil.ConfigFromOptions(fixture.Options)

	return formatter.FormatWithConfig(djot_parser.BuildDjotAst([]byte(fixture.Input)), config)
}

func runFixtures(t *testing.T, filename string, format func(testutil.Fixture) string) {
	t.Helper()

	fixtures, err := testutil.ReadFixtures(fixturePath(filename))
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			if !assert.Equal(t, fixture.Expected, format(fixture)) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
			}
		})
	}
}

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
	runFixtures(t, "basic.txt", formatDefault)
}

func TestFormat_SLWFixtures(t *testing.T) {
	runFixtures(t, "slw.txt", formatWithFixtureOptions)
}

func TestFormat_InlineFixtures(t *testing.T) {
	runFixtures(t, "inline.txt", formatDefault)
}

func TestFormat_Idempotency(t *testing.T) {
	fixtureFiles := []string{
		"basic.txt",
		"inline.txt",
		"slw.txt",
	}

	for _, filename := range fixtureFiles {
		t.Run(filename, func(t *testing.T) {
			fixtures, err := testutil.ReadFixtures(fixturePath(filename))
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
