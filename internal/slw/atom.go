package slw

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const closingChars = `"')]}`

// atom is a run of text that must stay on one line. Spaces inside protected
// djot constructs (verbatim spans, math, link destinations, attribute blocks)
// are not break opportunities, so they end up inside an atom rather than
// between two.
type atom struct {
	text        string
	width       int
	sentenceEnd bool
}

func splitAtoms(line string, cfg *Config) []atom {
	var (
		atoms   []atom
		current strings.Builder
		guarded []bool
	)

	flush := func() {
		if current.Len() == 0 {
			return
		}

		text := current.String()
		atoms = append(atoms, atom{
			text:        text,
			width:       utf8.RuneCountInString(text),
			sentenceEnd: endsSentence(text, guarded, cfg),
		})
		current.Reset()

		guarded = nil
	}

	for i := 0; i < len(line); {
		if span := protectedSpan(line[i:]); span > 0 {
			appendGuarded(&current, &guarded, line[i:i+span], true)
			i += span

			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		if r == ' ' || r == '\t' {
			flush()

			i += size

			continue
		}

		appendGuarded(&current, &guarded, line[i:i+size], false)
		i += size
	}

	flush()

	return atoms
}

func appendGuarded(b *strings.Builder, guarded *[]bool, s string, protectedText bool) {
	b.WriteString(s)

	for range s {
		*guarded = append(*guarded, protectedText)
	}
}

// protectedSpan reports the byte length of the djot construct starting at s,
// or 0 when s does not start one.
func protectedSpan(s string) int {
	switch s[0] {
	case '\\':
		return escapeSpan(s)
	case '`':
		return fenceSpan(s, '`')
	case '$':
		return mathSpan(s)
	case '{':
		return balancedSpan(s, '{', '}')
	case '[', '!':
		return linkSpan(s)
	default:
		return 0
	}
}

func escapeSpan(s string) int {
	const escapePair = 2

	if len(s) < escapePair {
		return 0
	}

	_, size := utf8.DecodeRuneInString(s[1:])

	return 1 + size
}

func fenceSpan(s string, delim byte) int {
	open := 0
	for open < len(s) && s[open] == delim {
		open++
	}

	closer := strings.Repeat(string(delim), open)

	end := strings.Index(s[open:], closer)
	if end < 0 {
		return 0
	}

	return open + end + len(closer)
}

func mathSpan(s string) int {
	if strings.HasPrefix(s, "$$") {
		return fenceSpan(s, '$')
	}

	const delimiters = 2

	end := strings.IndexByte(s[1:], '$')
	if end < 0 {
		return 0
	}

	return end + delimiters
}

func balancedSpan(s string, openChar, closeChar byte) int {
	depth := 0

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case openChar:
			depth++
		case closeChar:
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}

	return 0
}

// linkSpan keeps a whole link, image, or span construct together, matching
// mdformat-slw, which never splits link text across lines.
func linkSpan(s string) int {
	start := 0
	if s[0] == '!' {
		if len(s) < 2 || s[1] != '[' {
			return 0
		}

		start = 1
	}

	body := balancedSpan(s[start:], '[', ']')
	if body == 0 {
		return 0
	}

	span := start + body

	for span < len(s) {
		var tail int

		switch s[span] {
		case '(':
			tail = balancedSpan(s[span:], '(', ')')
		case '[':
			tail = balancedSpan(s[span:], '[', ']')
		case '{':
			tail = balancedSpan(s[span:], '{', '}')
		}

		if tail == 0 {
			break
		}

		span += tail
	}

	return span
}

// endsSentence reports whether text closes a sentence, ignoring markers that
// sit inside a protected construct and words the abbreviation list suppresses.
func endsSentence(text string, guarded []bool, cfg *Config) bool {
	runes := []rune(text)
	guard := runeGuards(text, guarded)

	i := len(runes) - 1
	for i >= 0 && !guard[i] && strings.ContainsRune(closingChars, runes[i]) {
		i--
	}

	if i < 0 || guard[i] || !strings.ContainsRune(cfg.Markers, runes[i]) {
		return false
	}

	return !isAbbreviation(runes, i, cfg.Abbreviations)
}

func runeGuards(text string, guarded []bool) []bool {
	out := make([]bool, 0, len(text))

	for i := range text {
		if i < len(guarded) {
			out = append(out, guarded[i])
		} else {
			out = append(out, false)
		}
	}

	return out
}

func isAbbreviation(runes []rune, markerPos int, abbreviations map[string]bool) bool {
	start := markerPos - 1
	for start >= 0 && (unicode.IsLetter(runes[start]) || runes[start] == '.') {
		start--
	}

	start++

	return abbreviations[strings.ToLower(string(runes[start:markerPos]))]
}
