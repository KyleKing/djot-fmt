package validate

import "strings"

type tokenKind int

const (
	kindTag tokenKind = iota
	kindWord
)

type token struct {
	name  string
	attrs string
	kind  tokenKind
}

func (t token) key() string {
	if t.kind == kindWord {
		return "w\x00" + t.name
	}

	return "t\x00" + t.name + "\x00" + t.attrs
}

func (t token) String() string {
	if t.kind == kindWord {
		return t.name
	}

	if t.attrs == "" {
		return "<" + t.name + ">"
	}

	return "<" + t.name + " " + t.attrs + ">"
}

// tokenize splits normalized HTML into tags and words. It assumes the input
// came from renderHTML, so tags are well formed and quoting is consistent.
func tokenize(html string) []token {
	var tokens []token

	for html != "" {
		start := strings.IndexByte(html, '<')
		if start < 0 {
			return append(tokens, words(html)...)
		}

		tokens = append(tokens, words(html[:start])...)

		end := findTagEnd(html[start:])
		if end < 0 {
			return append(tokens, words(html[start:])...)
		}

		tokens = append(tokens, parseTag(html[start+1:start+end]))
		html = html[start+end+1:]
	}

	return tokens
}

func findTagEnd(s string) int {
	inQuote := byte(0)

	for i := 1; i < len(s); i++ {
		switch {
		case inQuote != 0 && s[i] == inQuote:
			inQuote = 0
		case inQuote == 0 && (s[i] == '"' || s[i] == '\''):
			inQuote = s[i]
		case inQuote == 0 && s[i] == '>':
			return i
		}
	}

	return -1
}

func parseTag(body string) token {
	body = strings.TrimSuffix(strings.TrimSpace(body), "/")

	name, attrs, _ := strings.Cut(body, " ")

	return token{kind: kindTag, name: strings.ToLower(name), attrs: strings.TrimSpace(attrs)}
}

func words(text string) []token {
	fields := strings.Fields(text)
	tokens := make([]token, 0, len(fields))

	for _, f := range fields {
		tokens = append(tokens, token{kind: kindWord, name: f})
	}

	return tokens
}

func keys(tokens []token) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t.key()
	}

	return out
}

func onlyTags(tokens []token) []token {
	out := make([]token, 0, len(tokens))

	for _, t := range tokens {
		if t.kind == kindTag {
			out = append(out, t)
		}
	}

	return out
}
