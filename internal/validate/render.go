// Package validate compares the HTML rendering of a document before and after
// formatting so that a formatting bug cannot silently change what a file means.
package validate

import (
	"regexp"
	"strings"

	"github.com/sivukhin/godjot/v2/djot_html"
	"github.com/sivukhin/godjot/v2/djot_parser"
)

var (
	whitespaceRun  = regexp.MustCompile(`\s+`)
	spaceBeforeTag = regexp.MustCompile(` (<[a-zA-Z][-a-zA-Z0-9]*>)`)
	spaceAfterTag  = regexp.MustCompile(`(</[a-zA-Z][-a-zA-Z0-9]*>) `)
)

// renderHTML converts djot source to HTML normalized so that only meaningful
// differences survive. Whitespace outside code is insignificant in HTML, and
// empty paragraphs are ignored by user agents, so both are removed.
func renderHTML(source []byte, langs []string) string {
	ast := djot_parser.BuildDjotAst(source)
	html := djot_html.New().ConvertDjot(&djot_html.HtmlWriter{}, ast...).String()

	return normalizeHTML(html, langs)
}

func normalizeHTML(html string, langs []string) string {
	html = stripFormattedCode(html, langs)
	html = whitespaceRun.ReplaceAllString(html, " ")
	html = strings.ReplaceAll(html, "<p> ", "<p>")
	html = strings.ReplaceAll(html, " </p>", "</p>")
	html = spaceAfterTag.ReplaceAllString(html, "$1")
	html = spaceBeforeTag.ReplaceAllString(html, "$1")
	html = strings.ReplaceAll(html, "<p></p>", "")

	return strings.TrimSpace(html)
}

// stripFormattedCode removes the body of code blocks whose language has a
// configured formatter, since those bodies are rewritten on purpose.
func stripFormattedCode(html string, langs []string) string {
	for _, lang := range langs {
		html = removeBetween(html, `<code class="language-`+lang+`">`, "</code>")
	}

	return html
}

func removeBetween(html, openTag, closeTag string) string {
	var b strings.Builder

	for {
		start := strings.Index(html, openTag)
		if start < 0 {
			break
		}

		rest := html[start+len(openTag):]

		end := strings.Index(rest, closeTag)
		if end < 0 {
			break
		}

		b.WriteString(html[:start])
		b.WriteString(openTag)
		b.WriteString(closeTag)

		html = rest[end+len(closeTag):]
	}

	b.WriteString(html)

	return b.String()
}
