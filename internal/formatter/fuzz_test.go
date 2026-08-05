package formatter_test

import (
	"testing"

	"github.com/sivukhin/godjot/v2/djot_parser"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/validate"
)

// FuzzFormatPreservesMeaning asserts the property the CLI enforces at runtime:
// formatting must not change what the document renders to. Seeds cover the
// constructs already known to round-trip; run with -fuzz to look for more.
func FuzzFormatPreservesMeaning(f *testing.F) {
	seeds := []string{
		"hello world\n",
		"# Heading\n\ntext\n",
		"- a\n- b\n",
		"- a\n\n- b\n",
		"> quoted\n",
		"``` go\nx := 1\n```\n",
		"_em_ and *strong* and `code`\n",
		"[link](https://example.com)\n",
		"| a | b |\n| --- | --- |\n| 1 | 2 |\n",
		"::: note\ninside\n:::\n",
		"One sentence. Another sentence. A third one.\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		ast := djot_parser.BuildDjotAst([]byte(input))
		output := formatter.Format(ast)

		result := validate.Compare([]byte(input), []byte(output), validate.Options{})
		for _, d := range result.Rejected {
			t.Errorf("formatting changed the document\ninput: %q\noutput: %q\n%s: %s",
				input, output, d.Category, d.Detail)
		}
	})
}
